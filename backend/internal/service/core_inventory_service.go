package service

import (
	"context"
	"encoding/json"
	"strconv"
	"time"

	"github.com/ConradKurth/forecasting/backend/internal/repository/core"
	"github.com/ConradKurth/forecasting/backend/internal/shopify"
	"github.com/ConradKurth/forecasting/backend/pkg/id"
	"github.com/ConradKurth/forecasting/backend/pkg/logger"
	"github.com/hibiken/asynq"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgtype"
	"github.com/pkg/errors"
)

// CoreInventoryService provides business logic for core inventory operations
type CoreInventoryService struct {
	coreRepo      core.Querier
	shopifyClient *shopify.Client
	queueClient   *asynq.Client
}

// NewCoreInventoryService creates a new CoreInventoryService instance
func NewCoreInventoryService(coreRepo core.Querier, shopClient *shopify.Client, queue *asynq.Client) *CoreInventoryService {
	return &CoreInventoryService{
		coreRepo:      coreRepo,
		shopifyClient: shopClient,
		queueClient:   queue,
	}
}

// SyncShopifyData syncs all data from Shopify for a given integration
func (s *CoreInventoryService) SyncShopifyData(ctx context.Context, integrationID id.ID[id.PlatformIntegration], shopDomain, accessToken string) error {
	if s.shopifyClient == nil {
		s.shopifyClient = shopify.NewClient(shopDomain, accessToken)
	}

	// Check if sync was run recently
	syncState, err := s.coreRepo.GetSyncState(ctx, core.GetSyncStateParams{
		IntegrationID: integrationID,
		EntityType:    core.EntityTypeFullSync,
	})
	if err != nil && !errors.Is(err, pgx.ErrNoRows) {
		return errors.Wrap(err, "failed to get sync state")
	}

	if err == nil {
		if syncState.SyncStatus == core.SyncStatusInProgress {
			logger.Info("Sync already in progress", "integration_id", integrationID)
			return nil
		}
		if time.Since(syncState.LastSyncedAt.Time) < 4*time.Hour && syncState.SyncStatus == core.SyncStatusCompleted {
			logger.Info("Skipping sync - completed recently",
				"integration_id", integrationID,
				"last_synced", syncState.LastSyncedAt.Time,
			)
			return nil
		}
	}

	logger.Info("Starting Shopify sync", "integration_id", integrationID)

	// Mark sync as in progress
	_, err = s.coreRepo.UpsertSyncState(ctx, core.UpsertSyncStateParams{
		ID:            id.NewGeneration[id.SyncState](),
		IntegrationID: integrationID,
		EntityType:    core.EntityTypeFullSync,
		LastSyncedAt:  pgtype.Timestamp{Time: time.Now(), Valid: true},
		SyncStatus:    core.SyncStatusInProgress,
		ErrorMessage:  pgtype.Text{},
	})
	if err != nil {
		return errors.Wrap(err, "failed to mark sync in progress")
	}

	// Sync in logical order
	if err := s.SyncLocations(ctx, integrationID); err != nil {
		s.markSyncFailed(ctx, integrationID, "locations", err)
		return errors.Wrap(err, "failed to sync locations")
	}

	if err := s.SyncProducts(ctx, integrationID); err != nil {
		s.markSyncFailed(ctx, integrationID, "products", err)
		return errors.Wrap(err, "failed to sync products")
	}

	if err := s.SyncOrders(ctx, integrationID); err != nil {
		s.markSyncFailed(ctx, integrationID, "orders", err)
		return errors.Wrap(err, "failed to sync orders")
	}

	// Mark sync as completed
	_, err = s.coreRepo.UpsertSyncState(ctx, core.UpsertSyncStateParams{
		ID:            id.NewGeneration[id.SyncState](),
		IntegrationID: integrationID,
		EntityType:    core.EntityTypeFullSync,
		LastSyncedAt:  pgtype.Timestamp{Time: time.Now(), Valid: true},
		SyncStatus:    core.SyncStatusCompleted,
		ErrorMessage:  pgtype.Text{},
	})
	if err != nil {
		return errors.Wrap(err, "failed to mark sync completed")
	}

	logger.Info("Completed Shopify sync", "integration_id", integrationID)
	return nil
}

// EnqueueShopifySync enqueues Shopify sync tasks using the task queue instead of running synchronously
func (s *CoreInventoryService) EnqueueShopifySync(ctx context.Context, integrationID id.ID[id.PlatformIntegration], shopDomain, accessToken string) error {
	if s.queueClient == nil {
		return errors.New("queue client not set")
	}

	// Check if sync was run recently
	syncState, err := s.coreRepo.GetSyncState(ctx, core.GetSyncStateParams{
		IntegrationID: integrationID,
		EntityType:    core.EntityTypeFullSync,
	})
	if err == nil {
		if syncState.SyncStatus == core.SyncStatusInProgress {
			logger.Info("Sync already in progress", "integration_id", integrationID)
			return nil
		}
		if time.Since(syncState.LastSyncedAt.Time) < 4*time.Hour && syncState.SyncStatus == core.SyncStatusCompleted {
			logger.Info("Skipping sync - completed recently",
				"integration_id", integrationID,
				"last_synced", syncState.LastSyncedAt.Time,
			)
			return nil
		}
	}

	logger.Info("Enqueueing Shopify sync tasks", "integration_id", integrationID)

	// Create task payloads
	integrationIDStr := integrationID.String()

	// Enqueue individual sync tasks
	tasks := []struct {
		taskType string
		creator  func(string, string, string) (*asynq.Task, error)
	}{
		{"locations", newShopifyLocationsSyncTask},
		{"products", newShopifyProductsSyncTask},
		{"orders", newShopifyOrdersSyncTask},
	}

	for _, taskInfo := range tasks {
		task, err := taskInfo.creator(integrationIDStr, shopDomain, accessToken)
		if err != nil {
			return errors.Wrapf(err, "failed to create %s sync task", taskInfo.taskType)
		}

		info, err := s.queueClient.Enqueue(task)
		if err != nil {
			return errors.Wrapf(err, "failed to enqueue %s sync task", taskInfo.taskType)
		}

		logger.Info("Enqueued sync task",
			"task_type", taskInfo.taskType,
			"task_id", info.ID,
			"integration_id", integrationID,
		)
	}

	return nil
}

// SyncLocations syncs locations from Shopify
func (s *CoreInventoryService) SyncLocations(ctx context.Context, integrationID id.ID[id.PlatformIntegration]) error {
	response, err := s.shopifyClient.GetLocations(ctx)
	if err != nil {
		return errors.Wrap(err, "failed to get locations from Shopify")
	}

	for _, loc := range response.Locations {
		address := s.buildAddress(loc.Address1, loc.Address2, loc.City)

		_, err := s.coreRepo.UpsertLocation(ctx, core.UpsertLocationParams{
			ID:            id.NewGeneration[id.Location](),
			IntegrationID: integrationID,
			ExternalID:    pgtype.Text{String: strconv.FormatInt(loc.ID, 10), Valid: true},
			Name:          loc.Name,
			Address:       pgtype.Text{String: address, Valid: address != ""},
			Country:       pgtype.Text{String: loc.Country, Valid: loc.Country != ""},
			Province:      pgtype.Text{String: loc.Province, Valid: loc.Province != ""},
			IsActive:      pgtype.Bool{Bool: true, Valid: true},
		})
		if err != nil {
			return errors.Wrapf(err, "failed to upsert location %d", loc.ID)
		}
	}

	logger.Info("Synced locations",
		"integration_id", integrationID,
		"count", len(response.Locations),
	)
	return nil
}

// SyncProducts syncs products and variants from Shopify
func (s *CoreInventoryService) SyncProducts(ctx context.Context, integrationID id.ID[id.PlatformIntegration]) error {
	var allProducts []shopify.ShopifyProduct
	pageInfo := ""

	// Simple pagination implementation
	for {
		response, err := s.shopifyClient.GetProducts(ctx, 250, pageInfo)
		if err != nil {
			return errors.Wrap(err, "failed to get products from Shopify")
		}

		allProducts = append(allProducts, response.Products...)

		if len(response.Products) < 250 {
			break
		}
		break // Simple implementation without real pagination
	}

	// Process products
	for _, prod := range allProducts {
		productID := id.NewGeneration[id.Product]()

		// Upsert product
		_, err := s.coreRepo.UpsertProduct(ctx, core.UpsertProductParams{
			ID:            productID,
			IntegrationID: integrationID,
			ExternalID:    pgtype.Text{String: strconv.FormatInt(prod.ID, 10), Valid: true},
			Title:         prod.Title,
			Handle:        prod.Handle,
			ProductType:   pgtype.Text{String: prod.ProductType, Valid: prod.ProductType != ""},
			Status:        core.ProductStatus(prod.Status),
		})
		if err != nil {
			return errors.Wrapf(err, "failed to upsert product %d", prod.ID)
		}

		// Process variants
		for _, variant := range prod.Variants {
			variantID := id.NewGeneration[id.ProductVariant]()

			_, err := s.coreRepo.UpsertProductVariant(ctx, core.UpsertProductVariantParams{
				ID:              variantID,
				ProductID:       productID,
				ExternalID:      pgtype.Text{String: strconv.FormatInt(variant.ID, 10), Valid: true},
				Sku:             pgtype.Text{String: variant.SKU, Valid: variant.SKU != ""},
				Price:           pgtype.Numeric{}, // TODO: Convert variant.Price string to pgtype.Numeric
				InventoryItemID: pgtype.Text{String: strconv.FormatInt(variant.InventoryItemID, 10), Valid: variant.InventoryItemID > 0},
			})
			if err != nil {
				return errors.Wrapf(err, "failed to upsert variant %d", variant.ID)
			}
		}
	}

	logger.Info("Synced products",
		"integration_id", integrationID,
		"count", len(allProducts),
	)
	return nil
}

// SyncOrders syncs orders from the last 90 days
func (s *CoreInventoryService) SyncOrders(ctx context.Context, integrationID id.ID[id.PlatformIntegration]) error {
	since := time.Now().AddDate(0, 0, -90)
	var allOrders []shopify.ShopifyOrder
	pageInfo := ""

	for {
		response, err := s.shopifyClient.GetOrders(ctx, since, 250, pageInfo)
		if err != nil {
			return errors.Wrap(err, "failed to get orders from Shopify")
		}

		allOrders = append(allOrders, response.Orders...)

		if len(response.Orders) < 250 {
			break
		}
	}

	for _, order := range allOrders {
		orderID := id.NewGeneration[id.Order]()

		_, err := s.coreRepo.UpsertOrder(ctx, core.UpsertOrderParams{
			ID:                orderID,
			IntegrationID:     integrationID,
			ExternalID:        pgtype.Text{String: strconv.FormatInt(order.ID, 10), Valid: true},
			CreatedAt:         pgtype.Timestamp{Time: order.CreatedAt, Valid: true},
			FinancialStatus:   core.FinancialStatus(order.FinancialStatus),
			FulfillmentStatus: core.FulfillmentStatus(getStringValue(order.FulfillmentStatus)),
			TotalPrice:        pgtype.Numeric{}, // TODO: Convert order.TotalPrice to pgtype.Numeric
			CancelledAt:       pgtype.Timestamp{Time: getTimeValue(order.CancelledAt), Valid: order.CancelledAt != nil},
		})
		if err != nil {
			return errors.Wrapf(err, "failed to upsert order %d", order.ID)
		}

		// Process line items - skip for now since we need to lookup product/variant IDs
		// TODO: Implement proper lookup of internal product/variant IDs from external IDs
		logger.Info("Skipping line items sync for now - needs product/variant ID lookup",
			"order_id", order.ID,
			"line_items_count", len(order.LineItems),
		)
	}

	logger.Info("Synced orders",
		"integration_id", integrationID,
		"count", len(allOrders),
	)
	return nil
}

// Helper methods
func (s *CoreInventoryService) buildAddress(address1, address2, city string) string {
	var parts []string
	if address1 != "" {
		parts = append(parts, address1)
	}
	if address2 != "" {
		parts = append(parts, address2)
	}
	if city != "" {
		parts = append(parts, city)
	}

	address := ""
	for i, part := range parts {
		if i > 0 {
			address += ", "
		}
		address += part
	}
	return address
}

func (s *CoreInventoryService) markSyncFailed(ctx context.Context, integrationID id.ID[id.PlatformIntegration], entityType core.EntityType, syncErr error) {
	_, err := s.coreRepo.UpsertSyncState(ctx, core.UpsertSyncStateParams{
		ID:            id.NewGeneration[id.SyncState](),
		IntegrationID: integrationID,
		EntityType:    entityType,
		LastSyncedAt:  pgtype.Timestamp{Time: time.Now(), Valid: true},
		SyncStatus:    core.SyncStatusFailed,
		ErrorMessage:  pgtype.Text{String: syncErr.Error(), Valid: true},
	})
	if err != nil {
		logger.Error("Failed to mark sync as failed", "error", err)
	}
}

func getStringValue(ptr *string) string {
	if ptr == nil {
		return ""
	}
	return *ptr
}

func getTimeValue(ptr *time.Time) time.Time {
	if ptr == nil {
		return time.Time{}
	}
	return *ptr
}

// getInt64Value is unused but kept for future line item processing
// func getInt64Value(ptr *int64) int64 {
// 	if ptr == nil {
// 		return 0
// 	}
// 	return *ptr
// }

// Task creation functions for the queue (avoid importing worker package due to circular dependency)

// Task type constants (must match those in worker package)
const (
	TypeShopifyLocationsSyncTask = "shopify:locations_sync"
	TypeShopifyProductsSyncTask  = "shopify:products_sync"
	TypeShopifyOrdersSyncTask    = "shopify:orders_sync"
)

// Task payload structure (matches ShopifyInventorySyncPayload in worker package)
type shopifyInventorySyncPayload struct {
	IntegrationID string `json:"integration_id"`
	ShopDomain    string `json:"shop_domain"`
	AccessToken   string `json:"access_token"`
}

func newShopifyLocationsSyncTask(integrationID, shopDomain, accessToken string) (*asynq.Task, error) {
	payload := shopifyInventorySyncPayload{
		IntegrationID: integrationID,
		ShopDomain:    shopDomain,
		AccessToken:   accessToken,
	}

	data, err := json.Marshal(payload)
	if err != nil {
		return nil, err
	}

	return asynq.NewTask(TypeShopifyLocationsSyncTask, data), nil
}

func newShopifyProductsSyncTask(integrationID, shopDomain, accessToken string) (*asynq.Task, error) {
	payload := shopifyInventorySyncPayload{
		IntegrationID: integrationID,
		ShopDomain:    shopDomain,
		AccessToken:   accessToken,
	}

	data, err := json.Marshal(payload)
	if err != nil {
		return nil, err
	}

	return asynq.NewTask(TypeShopifyProductsSyncTask, data), nil
}

func newShopifyOrdersSyncTask(integrationID, shopDomain, accessToken string) (*asynq.Task, error) {
	payload := shopifyInventorySyncPayload{
		IntegrationID: integrationID,
		ShopDomain:    shopDomain,
		AccessToken:   accessToken,
	}

	data, err := json.Marshal(payload)
	if err != nil {
		return nil, err
	}

	return asynq.NewTask(TypeShopifyOrdersSyncTask, data), nil
}
