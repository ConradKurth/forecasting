package manager

import (
	"context"
	"strconv"
	"strings"
	"time"

	"github.com/ConradKurth/forecasting/backend/internal/db"
	"github.com/ConradKurth/forecasting/backend/internal/interfaces"
	"github.com/ConradKurth/forecasting/backend/internal/repository/core"
	"github.com/ConradKurth/forecasting/backend/internal/repository/shopify"
	shopifyapi "github.com/ConradKurth/forecasting/backend/internal/shopify"
	"github.com/ConradKurth/forecasting/backend/pkg/id"
	"github.com/ConradKurth/forecasting/backend/pkg/logger"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgtype"
	"github.com/pkg/errors"
)

// SyncStatus represents the status of a synchronization operation
type SyncStatus string

const (
	// Standard sync statuses
	SyncStatusPending    SyncStatus = "pending"
	SyncStatusInProgress SyncStatus = "in_progress"
	SyncStatusCompleted  SyncStatus = "completed"
	SyncStatusFailed     SyncStatus = "failed"

	// Special result statuses
	SyncStatusSyncStarted     SyncStatus = "sync_started"
	SyncStatusNeverSynced     SyncStatus = "never_synced"
	SyncStatusPartialSyncOnly SyncStatus = "partial_sync_only"
)

// EntityType represents the type of entity being synchronized
type EntityType string

const (
	EntityTypeFullSync       EntityType = "full_sync"
	EntityTypeProduct        EntityType = "product"
	EntityTypeInventoryItem  EntityType = "inventory_item"
	EntityTypeInventoryLevel EntityType = "inventory_level"
	EntityTypeOrder          EntityType = "order"
	EntityTypeLocation       EntityType = "location"
)

// Type conversion helpers to keep conversions DRY and centralized

// ConvertSyncStatus provides bidirectional conversion between core and manager SyncStatus types

func FromCoreStatus(coreStatus core.SyncStatus) SyncStatus {
	switch coreStatus {
	case core.SyncStatusPending:
		return SyncStatusPending
	case core.SyncStatusInProgress:
		return SyncStatusInProgress
	case core.SyncStatusCompleted:
		return SyncStatusCompleted
	case core.SyncStatusFailed:
		return SyncStatusFailed
	default:
		return SyncStatusPending
	}
}

func ToCoreStatus(managerStatus SyncStatus) core.SyncStatus {
	switch managerStatus {
	case SyncStatusPending:
		return core.SyncStatusPending
	case SyncStatusInProgress:
		return core.SyncStatusInProgress
	case SyncStatusCompleted:
		return core.SyncStatusCompleted
	case SyncStatusFailed:
		return core.SyncStatusFailed
	default:
		return core.SyncStatusPending
	}
}

// ConvertEntityType provides bidirectional conversion between core and manager EntityType types

func FromCoreEntity(coreEntity core.EntityType) EntityType {
	switch coreEntity {
	case core.EntityTypeFullSync:
		return EntityTypeFullSync
	case core.EntityTypeProducts:
		return EntityTypeProduct
	case core.EntityTypeInventory:
		return EntityTypeInventoryItem
	case core.EntityTypeOrders:
		return EntityTypeOrder
	case core.EntityTypeLocations:
		return EntityTypeLocation
	default:
		return EntityTypeFullSync
	}
}

func ToCoreEntity(managerEntity EntityType) core.EntityType {
	switch managerEntity {
	case EntityTypeFullSync:
		return core.EntityTypeFullSync
	case EntityTypeProduct:
		return core.EntityTypeProducts
	case EntityTypeInventoryItem:
		return core.EntityTypeInventory
	case EntityTypeOrder:
		return core.EntityTypeOrders
	case EntityTypeLocation:
		return core.EntityTypeLocations
	default:
		return core.EntityTypeFullSync
	}
}

// InventorySyncManager orchestrates inventory synchronization operations
// It ensures data consistency by wrapping operations in database transactions
// and coordinates between multiple repositories for complex sync workflows
type InventorySyncManager struct {
	database db.Database
	queue    interfaces.Queue
}

// NewInventorySyncManager creates a new InventorySyncManager instance
func NewInventorySyncManager(database db.Database, queue interfaces.Queue) *InventorySyncManager {
	return &InventorySyncManager{
		database: database,
		queue:    queue,
	}
}

// SyncRequest represents a synchronization request
type SyncRequest struct {
	UserID     id.ID[id.User] `json:"user_id"`
	ShopDomain string         `json:"shop_domain"`
	Force      bool           `json:"force,omitempty"`
}

// SyncResult represents the result of a synchronization operation
type SyncResult struct {
	IntegrationID string     `json:"integration_id"`
	Status        SyncStatus `json:"status"`
	LastSynced    *time.Time `json:"last_synced,omitempty"`
	Error         string     `json:"error,omitempty"`
}

// ShopifySyncData holds all normalized data fetched from Shopify API
type ShopifySyncData struct {
	Locations       []core.InsertLocationsBatchParams       `json:"locations"`
	Products        []core.InsertProductsBatchParams        `json:"products"`
	ProductVariants []core.InsertProductVariantsBatchParams `json:"product_variants"`
	InventoryItems  []core.InsertInventoryItemsBatchParams  `json:"inventory_items"`
	Orders          []core.InsertOrdersBatchParams          `json:"orders"`
}

// Stats for tracking sync progress
type SyncStats struct {
	LocationsCount       int `json:"locations_count"`
	ProductsCount        int `json:"products_count"`
	ProductVariantsCount int `json:"product_variants_count"`
	InventoryItemsCount  int `json:"inventory_items_count"`
	OrdersCount          int `json:"orders_count"`
}

// TriggerShopifySync orchestrates a complete Shopify synchronization process
// This includes validating access, creating integrations, and triggering async sync
func (m *InventorySyncManager) TriggerShopifySync(ctx context.Context, req SyncRequest) (*SyncResult, error) {
	// Validate user exists
	_, err := m.database.GetUsers().GetUserByID(ctx, req.UserID)
	if err != nil {
		return nil, errors.Wrap(err, "failed to get user")
	}

	// Get shop and validate it exists
	shop, err := m.database.GetShopify().GetShopifyStoreByDomain(ctx, req.ShopDomain)
	if err != nil {
		return nil, errors.Wrap(err, "shop not found")
	}

	// Get shopify user to get access token
	shopifyUser, err := m.database.GetShopify().GetShopifyUserByUserAndStore(ctx, shopify.GetShopifyUserByUserAndStoreParams{
		UserID:         req.UserID,
		ShopifyStoreID: shop.ID,
	})
	if err != nil {
		return nil, errors.Wrap(err, "failed to get access token")
	}

	accessToken := shopifyUser.AccessToken.String()
	if accessToken == "" {
		return nil, errors.New("no access token found for user and shop")
	}

	// Create or get platform integration
	integration, err := m.getOrCreateShopifyIntegration(ctx, shop.ID, req.ShopDomain)
	if err != nil {
		return nil, errors.Wrap(err, "failed to get/create integration")
	}

	// Check if sync should proceed
	if !req.Force {
		shouldSkip, skipReason, err := m.shouldSkipSync(ctx, integration.ID)
		if err != nil {
			return nil, errors.Wrap(err, "failed to check sync status")
		}
		if shouldSkip {
			return &SyncResult{
				IntegrationID: integration.ID.String(),
				Status:        skipReason,
			}, nil
		}
	}

	// Set sync state to in_progress before enqueuing the task
	err = m.database.WithTx(ctx, func(tx *db.TxDB) error {
		return m.updateSyncState(ctx, tx, integration.ID, EntityTypeFullSync, SyncStatusInProgress, "")
	})
	if err != nil {
		return nil, errors.Wrap(err, "failed to set sync state to in_progress")
	}

	// Enqueue async sync task
	err = m.queue.EnqueueShopifyInventorySync(ctx, integration.ID.String(), req.ShopDomain, accessToken)
	if err != nil {
		// If enqueueing fails, set status back to failed
		_ = m.database.WithTx(ctx, func(tx *db.TxDB) error {
			return m.updateSyncState(ctx, tx, integration.ID, EntityTypeFullSync, SyncStatusFailed, "failed to enqueue sync task")
		})
		return nil, errors.Wrap(err, "failed to enqueue sync task")
	}

	return &SyncResult{
		IntegrationID: integration.ID.String(),
		Status:        SyncStatusInProgress,
	}, nil
}

// GetSyncStatus retrieves the current synchronization status for a shop
func (m *InventorySyncManager) GetSyncStatus(ctx context.Context, userID id.ID[id.User], shopDomain string) (*SyncResult, error) {
	// Get shop
	shop, err := m.database.GetShopify().GetShopifyStoreByDomain(ctx, shopDomain)
	if err != nil {
		return nil, errors.Wrap(err, "shop not found")
	}

	// Get integration - if it doesn't exist, return never synced status
	integration, err := m.database.GetCore().GetPlatformIntegrationByShopAndType(ctx, core.GetPlatformIntegrationByShopAndTypeParams{
		ShopID:       shop.ID,
		PlatformType: core.PlatformTypeShopify,
	})
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			// No integration found - return never synced status
			return &SyncResult{
				IntegrationID: "",
				Status:        SyncStatusNeverSynced,
			}, nil
		}
		return nil, errors.Wrap(err, "failed to get integration")
	}

	// Get sync states
	syncStates, err := m.database.GetCore().GetSyncStatesByIntegrationID(ctx, integration.ID)
	if err != nil {
		return nil, errors.Wrap(err, "failed to get sync states")
	}

	result := &SyncResult{
		IntegrationID: integration.ID.String(),
	}

	if len(syncStates) == 0 {
		result.Status = SyncStatusNeverSynced
	} else {
		// Find the most recent full_sync
		for _, state := range syncStates {
			if state.EntityType == core.EntityTypeFullSync {
				result.Status = FromCoreStatus(state.SyncStatus)
				if state.ErrorMessage.Valid {
					result.Error = state.ErrorMessage.String
				}
				if state.LastSyncedAt.Valid {
					result.LastSynced = &state.LastSyncedAt.Time
				}
				break
			}
		}
		if result.Status == "" {
			result.Status = SyncStatusPartialSyncOnly
		}
	}

	return result, nil
}

// getOrCreateShopifyIntegration gets or creates a Shopify platform integration
func (m *InventorySyncManager) getOrCreateShopifyIntegration(ctx context.Context, shopID id.ID[id.ShopifyStore], shopDomain string) (core.PlatformIntegration, error) {
	// Try to get existing integration
	integration, err := m.database.GetCore().GetPlatformIntegrationByShopAndType(ctx, core.GetPlatformIntegrationByShopAndTypeParams{
		ShopID:       shopID,
		PlatformType: core.PlatformTypeShopify,
	})
	if err == nil {
		return integration, nil
	}

	// Create new integration
	integration, err = m.database.GetCore().CreatePlatformIntegration(ctx, core.CreatePlatformIntegrationParams{
		ID:             id.NewGeneration[id.PlatformIntegration](),
		ShopID:         shopID,
		PlatformType:   core.PlatformTypeShopify,
		PlatformShopID: shopDomain,
		IsActive:       pgtype.Bool{Bool: true, Valid: true},
	})
	if err != nil {
		return core.PlatformIntegration{}, errors.Wrap(err, "failed to create platform integration")
	}

	return integration, nil
}

// shouldSkipSync determines if a sync should be skipped based on current state
func (m *InventorySyncManager) shouldSkipSync(ctx context.Context, integrationID id.ID[id.PlatformIntegration]) (bool, SyncStatus, error) {
	syncState, err := m.database.GetCore().GetSyncState(ctx, core.GetSyncStateParams{
		IntegrationID: integrationID,
		EntityType:    core.EntityTypeFullSync,
	})
	if err != nil {
		// No sync state exists, proceed with sync
		return false, "", nil
	}

	if syncState.SyncStatus == core.SyncStatusInProgress {
		return true, SyncStatusInProgress, nil
	}

	// Check if recently synced (within 30 minutes)
	if syncState.SyncStatus == core.SyncStatusCompleted &&
		syncState.LastSyncedAt.Valid &&
		time.Since(syncState.LastSyncedAt.Time) < 30*time.Minute { // 30 minutes
		return true, SyncStatusCompleted, nil
	}

	return false, "", nil
}

// SyncInventory performs comprehensive inventory synchronization for a platform integration
// with optimized batch processing - fetch all data first, normalize, then batch insert
func (m *InventorySyncManager) SyncInventory(ctx context.Context, integrationID id.ID[id.PlatformIntegration]) error {
	logger.Info("Starting optimized inventory sync", "integration_id", integrationID)

	// Start a database transaction for consistency
	return m.database.WithTx(ctx, func(tx *db.TxDB) error {
		// Status should already be in_progress from TriggerShopifySync

		// Get platform integration details
		integration, err := tx.GetCore().GetPlatformIntegrationByID(ctx, integrationID)
		if err != nil {
			return errors.Wrap(err, "failed to get platform integration")
		}

		// Get shopify store and access token
		shopifyUser, err := m.getShopifyUser(ctx, tx, integration.ShopID)
		if err != nil {
			return errors.Wrap(err, "failed to get shopify user")
		}

		accessToken := shopifyUser.AccessToken.String()
		if accessToken == "" {
			return errors.New("no access token found")
		}

		// Create shopify client directly (without ShopifyManager dependency)
		client := shopifyapi.NewClient(integration.PlatformShopID, accessToken)

		// Phase 1: Fetch all data from API
		logger.Info("Phase 1: Fetching all data from Shopify API", "integration_id", integrationID)

		syncData, err := m.fetchAllShopifyData(ctx, client, integrationID)
		if err != nil {
			return m.handleSyncError(ctx, tx, integrationID, "failed to fetch data from API", err)
		}

		// Phase 2: Normalize and batch insert
		logger.Info("Phase 2: Normalizing and batch inserting data", "integration_id", integrationID)

		if err := m.batchSyncAllData(ctx, tx, syncData); err != nil {
			return m.handleSyncError(ctx, tx, integrationID, "failed to batch sync data", err)
		}

		// Mark sync as completed
		if err := m.updateSyncState(ctx, tx, integrationID, EntityTypeFullSync, SyncStatusCompleted, ""); err != nil {
			return errors.Wrap(err, "failed to update sync state to completed")
		}

		logger.Info("Optimized inventory sync completed successfully", "integration_id", integrationID)
		return nil
	})
}

// updateSyncState updates the sync state for a given integration and entity type
func (m *InventorySyncManager) updateSyncState(ctx context.Context, tx *db.TxDB, integrationID id.ID[id.PlatformIntegration], entityType EntityType, status SyncStatus, errorMessage string) error {
	var errorMsg pgtype.Text
	if errorMessage != "" {
		errorMsg = pgtype.Text{String: errorMessage, Valid: true}
	}

	var lastSyncedAt pgtype.Timestamp
	if status == SyncStatusCompleted {
		lastSyncedAt = pgtype.Timestamp{Time: time.Now().UTC(), Valid: true}
	}

	_, err := tx.GetCore().UpsertSyncState(ctx, core.UpsertSyncStateParams{
		ID:            id.NewGeneration[id.SyncState](),
		IntegrationID: integrationID,
		EntityType:    ToCoreEntity(entityType),
		SyncStatus:    ToCoreStatus(status),
		ErrorMessage:  errorMsg,
		LastSyncedAt:  lastSyncedAt,
	})
	return err
}

// getShopifyUser retrieves the shopify user for a given shop ID
func (m *InventorySyncManager) getShopifyUser(ctx context.Context, tx *db.TxDB, shopID id.ID[id.ShopifyStore]) (shopify.ShopifyUser, error) {
	// Get any shopify user for this store (there might be multiple, we just need one with valid access token)
	shopifyUsers, err := tx.GetShopify().GetShopifyUsersByStore(ctx, shopID)
	if err != nil {
		return shopify.ShopifyUser{}, errors.Wrap(err, "failed to get shopify users")
	}

	if len(shopifyUsers) == 0 {
		return shopify.ShopifyUser{}, errors.New("no shopify users found for store")
	}

	// Return the first user with a valid access token
	for _, user := range shopifyUsers {
		if user.AccessToken.String() != "" {
			return user, nil
		}
	}

	return shopify.ShopifyUser{}, errors.New("no shopify user with valid access token found")
}

// handleSyncError handles sync errors by updating the sync state and logging
func (m *InventorySyncManager) handleSyncError(ctx context.Context, tx *db.TxDB, integrationID id.ID[id.PlatformIntegration], message string, err error) error {
	fullError := errors.Wrap(err, message)
	logger.Error("Sync error", "integration_id", integrationID, "error", fullError)

	if updateErr := m.updateSyncState(ctx, tx, integrationID, EntityTypeFullSync, SyncStatusFailed, fullError.Error()); updateErr != nil {
		logger.Error("Failed to update sync state to failed", "integration_id", integrationID, "error", updateErr)
	}

	return fullError
}

// fetchAllShopifyData fetches all data from Shopify API and normalizes it for batch insertion
func (m *InventorySyncManager) fetchAllShopifyData(ctx context.Context, client *shopifyapi.Client, integrationID id.ID[id.PlatformIntegration]) (*ShopifySyncData, error) {
	logger.Info("Fetching all Shopify data")

	var locations []shopifyapi.ShopifyLocation
	var products []shopifyapi.ShopifyProduct
	var orders []shopifyapi.ShopifyOrder

	// Fetch locations
	logger.Info("Fetching locations from API")
	pageInfo := ""
	for {
		response, err := client.GetLocations(ctx, 250, pageInfo)
		if err != nil {
			return nil, errors.Wrap(err, "failed to fetch locations")
		}
		locations = append(locations, response.Locations...)
		if response.Pagination.NextPageInfo == "" {
			break
		}
		pageInfo = response.Pagination.NextPageInfo
	}
	logger.Info("Locations fetched", "count", len(locations))

	// Fetch products
	logger.Info("Fetching products from API")
	pageInfo = ""
	for {
		response, err := client.GetProducts(ctx, 250, pageInfo)
		if err != nil {
			return nil, errors.Wrap(err, "failed to fetch products")
		}
		products = append(products, response.Products...)
		if response.Pagination.NextPageInfo == "" {
			break
		}
		pageInfo = response.Pagination.NextPageInfo
	}
	logger.Info("Products fetched", "count", len(products))

	// TODO: Fetch orders from last 30 days - disabled due to Shopify protected customer data restrictions
	// Requires Shopify app approval for protected customer data access
	// logger.Info("Fetching orders from API")
	// createdAtMin := time.Now().AddDate(0, 0, -30)
	// pageInfo = ""
	// for {
	// 	response, err := client.GetOrders(ctx, createdAtMin, 250, pageInfo)
	// 	if err != nil {
	// 		return nil, errors.Wrap(err, "failed to fetch orders")
	// 	}
	// 	orders = append(orders, response.Orders...)
	// 	if response.Pagination.NextPageInfo == "" {
	// 		break
	// 	}
	// 	pageInfo = response.Pagination.NextPageInfo
	// }
	// logger.Info("Orders fetched", "count", len(orders))

	logger.Info("All data fetched successfully",
		"locations", len(locations),
		"products", len(products),
		"orders", len(orders))

	// Now normalize all data for batch insertion
	return m.normalizeShopifyData(ctx, client, integrationID, locations, products, orders)
}

// normalizeShopifyData converts raw Shopify API data into normalized database structures
func (m *InventorySyncManager) normalizeShopifyData(ctx context.Context, client *shopifyapi.Client, integrationID id.ID[id.PlatformIntegration],
	locations []shopifyapi.ShopifyLocation, products []shopifyapi.ShopifyProduct, orders []shopifyapi.ShopifyOrder) (*ShopifySyncData, error) {

	logger.Info("Normalizing Shopify data for batch insertion")
	now := pgtype.Timestamp{Time: time.Now().UTC(), Valid: true}

	syncData := &ShopifySyncData{}

	// Track external ID to internal ID mappings for products
	productIDMap := make(map[string]id.ID[id.Product])

	// 1. Normalize locations
	for _, location := range locations {
		var addressParts []string
		if location.Address1 != "" {
			addressParts = append(addressParts, location.Address1)
		}
		if location.Address2 != "" {
			addressParts = append(addressParts, location.Address2)
		}
		if location.City != "" {
			addressParts = append(addressParts, location.City)
		}
		address := strings.Join(addressParts, ", ")

		syncData.Locations = append(syncData.Locations, core.InsertLocationsBatchParams{
			ID:            id.NewGeneration[id.Location](),
			IntegrationID: integrationID,
			ExternalID:    pgtype.Text{String: strconv.FormatInt(location.ID, 10), Valid: true},
			Name:          location.Name,
			Address:       pgtype.Text{String: address, Valid: address != ""},
			Country:       pgtype.Text{String: location.Country, Valid: location.Country != ""},
			Province:      pgtype.Text{String: location.Province, Valid: location.Province != ""},
			IsActive:      pgtype.Bool{Bool: true, Valid: true},
			CreatedAt:     now,
			UpdatedAt:     now,
		})
	}

	// 2. Normalize products
	for _, product := range products {
		var status core.ProductStatus
		switch product.Status {
		case "active":
			status = core.ProductStatusActive
		case "archived":
			status = core.ProductStatusArchived
		case "draft":
			status = core.ProductStatusDraft
		default:
			status = core.ProductStatusDraft
		}

		productID := id.NewGeneration[id.Product]()
		productIDMap[strconv.FormatInt(product.ID, 10)] = productID

		syncData.Products = append(syncData.Products, core.InsertProductsBatchParams{
			ID:            productID,
			IntegrationID: integrationID,
			ExternalID:    pgtype.Text{String: strconv.FormatInt(product.ID, 10), Valid: true},
			Title:         product.Title,
			Handle:        product.Handle,
			ProductType:   pgtype.Text{String: product.ProductType, Valid: product.ProductType != ""},
			Status:        status,
			CreatedAt:     now,
			UpdatedAt:     now,
		})

		// 3. Normalize product variants for this product
		for _, variant := range product.Variants {
			var price pgtype.Numeric
			if err := price.Scan(variant.Price); err != nil {
				logger.Warn("Failed to parse variant price", "variant_id", variant.ID, "price", variant.Price, "error", err)
				price = pgtype.Numeric{}
			}

			syncData.ProductVariants = append(syncData.ProductVariants, core.InsertProductVariantsBatchParams{
				ID:              id.NewGeneration[id.ProductVariant](),
				ProductID:       productID,
				ExternalID:      pgtype.Text{String: strconv.FormatInt(variant.ID, 10), Valid: true},
				Sku:             pgtype.Text{String: variant.SKU, Valid: variant.SKU != ""},
				Price:           price,
				InventoryItemID: pgtype.Text{String: strconv.FormatInt(variant.InventoryItemID, 10), Valid: variant.InventoryItemID != 0},
				CreatedAt:       now,
				UpdatedAt:       now,
			})
		}
	}

	// 4. Get inventory items for all variants
	inventoryItemIDs := make(map[int64]bool)
	var inventoryItemIDList []int64

	for _, product := range products {
		for _, variant := range product.Variants {
			if variant.InventoryItemID != 0 && !inventoryItemIDs[variant.InventoryItemID] {
				inventoryItemIDs[variant.InventoryItemID] = true
				inventoryItemIDList = append(inventoryItemIDList, variant.InventoryItemID)
			}
		}
	}

	// Fetch inventory items in batches
	if len(inventoryItemIDList) > 0 {
		logger.Info("Fetching inventory items", "count", len(inventoryItemIDList))
		batchSize := 100
		for i := 0; i < len(inventoryItemIDList); i += batchSize {
			end := i + batchSize
			if end > len(inventoryItemIDList) {
				end = len(inventoryItemIDList)
			}

			batch := inventoryItemIDList[i:end]
			response, err := client.GetInventoryItems(ctx, batch, 100, "")
			if err != nil {
				return nil, errors.Wrapf(err, "failed to get inventory items batch %d-%d", i, end)
			}

			for _, item := range response.InventoryItems {
				var cost pgtype.Numeric
				if item.Cost != "" {
					if err := cost.Scan(item.Cost); err != nil {
						logger.Warn("Failed to parse inventory item cost", "item_id", item.ID, "cost", item.Cost, "error", err)
						cost = pgtype.Numeric{}
					}
				}

				syncData.InventoryItems = append(syncData.InventoryItems, core.InsertInventoryItemsBatchParams{
					ID:            id.NewGeneration[id.InventoryItem](),
					IntegrationID: integrationID,
					ExternalID:    pgtype.Text{String: strconv.FormatInt(item.ID, 10), Valid: true},
					Sku:           pgtype.Text{String: item.SKU, Valid: item.SKU != ""},
					Tracked:       pgtype.Bool{Bool: item.Tracked, Valid: true},
					Cost:          cost,
					CreatedAt:     now,
					UpdatedAt:     now,
				})
			}
		}
	}

	// 5. Normalize orders
	for _, order := range orders {
		var cancelledAt pgtype.Timestamp
		if order.CancelledAt != nil {
			cancelledAt = pgtype.Timestamp{Time: *order.CancelledAt, Valid: true}
		}

		var financialStatus core.FinancialStatus
		switch order.FinancialStatus {
		case "pending":
			financialStatus = core.FinancialStatusPending
		case "authorized":
			financialStatus = core.FinancialStatusAuthorized
		case "partially_paid":
			financialStatus = core.FinancialStatusPartiallyPaid
		case "paid":
			financialStatus = core.FinancialStatusPaid
		case "partially_refunded":
			financialStatus = core.FinancialStatusPartiallyRefunded
		case "refunded":
			financialStatus = core.FinancialStatusRefunded
		case "voided":
			financialStatus = core.FinancialStatusVoided
		default:
			financialStatus = core.FinancialStatusPending
		}

		var fulfillmentStatus core.FulfillmentStatus
		if order.FulfillmentStatus != nil {
			switch *order.FulfillmentStatus {
			case "fulfilled":
				fulfillmentStatus = core.FulfillmentStatusFulfilled
			case "partial":
				fulfillmentStatus = core.FulfillmentStatusPartial
			case "restocked":
				fulfillmentStatus = core.FulfillmentStatusRestocked
			default:
				fulfillmentStatus = core.FulfillmentStatusNull
			}
		} else {
			fulfillmentStatus = core.FulfillmentStatusNull
		}

		var totalPrice pgtype.Numeric
		if err := totalPrice.Scan(order.TotalPrice); err != nil {
			logger.Warn("Failed to parse order total price", "order_id", order.ID, "price", order.TotalPrice, "error", err)
			totalPrice = pgtype.Numeric{}
		}

		syncData.Orders = append(syncData.Orders, core.InsertOrdersBatchParams{
			ID:                id.NewGeneration[id.Order](),
			IntegrationID:     integrationID,
			ExternalID:        pgtype.Text{String: strconv.FormatInt(order.ID, 10), Valid: true},
			CreatedAt:         pgtype.Timestamp{Time: order.CreatedAt, Valid: true},
			FinancialStatus:   financialStatus,
			FulfillmentStatus: fulfillmentStatus,
			TotalPrice:        totalPrice,
			CancelledAt:       cancelledAt,
		})
	}

	logger.Info("Data normalization completed",
		"locations", len(syncData.Locations),
		"products", len(syncData.Products),
		"variants", len(syncData.ProductVariants),
		"inventory_items", len(syncData.InventoryItems),
		"orders", len(syncData.Orders))

	return syncData, nil
}

// batchSyncAllData performs batch insertion of all normalized data
func (m *InventorySyncManager) batchSyncAllData(ctx context.Context, tx *db.TxDB, syncData *ShopifySyncData) error {
	const batchSize = 250

	// 1. Batch insert locations
	if len(syncData.Locations) > 0 {
		logger.Info("Batch inserting locations", "count", len(syncData.Locations))
		if err := m.batchInsertLocations(ctx, tx, syncData.Locations, batchSize); err != nil {
			return errors.Wrap(err, "failed to batch insert locations")
		}
	}

	// 2. Batch insert products
	if len(syncData.Products) > 0 {
		logger.Info("Batch inserting products", "count", len(syncData.Products))
		if err := m.batchInsertProducts(ctx, tx, syncData.Products, batchSize); err != nil {
			return errors.Wrap(err, "failed to batch insert products")
		}
	}

	// 3. Batch insert product variants
	if len(syncData.ProductVariants) > 0 {
		logger.Info("Batch inserting product variants", "count", len(syncData.ProductVariants))
		if err := m.batchInsertProductVariants(ctx, tx, syncData.ProductVariants, batchSize); err != nil {
			return errors.Wrap(err, "failed to batch insert product variants")
		}
	}

	// 4. Batch insert inventory items
	if len(syncData.InventoryItems) > 0 {
		logger.Info("Batch inserting inventory items", "count", len(syncData.InventoryItems))
		if err := m.batchInsertInventoryItems(ctx, tx, syncData.InventoryItems, batchSize); err != nil {
			return errors.Wrap(err, "failed to batch insert inventory items")
		}
	}

	// 5. Batch insert orders
	if len(syncData.Orders) > 0 {
		logger.Info("Batch inserting orders", "count", len(syncData.Orders))
		if err := m.batchInsertOrders(ctx, tx, syncData.Orders, batchSize); err != nil {
			return errors.Wrap(err, "failed to batch insert orders")
		}
	}

	logger.Info("All batch insertions completed successfully")
	return nil
}

// batchInsertLocations inserts locations using upsert to handle conflicts
func (m *InventorySyncManager) batchInsertLocations(ctx context.Context, tx *db.TxDB, locations []core.InsertLocationsBatchParams, batchSize int) error {
	for i := 0; i < len(locations); i += batchSize {
		end := i + batchSize
		if end > len(locations) {
			end = len(locations)
		}

		batch := locations[i:end]
		
		// Use individual upserts instead of batch insert to handle ON CONFLICT
		for _, location := range batch {
			_, err := tx.GetCore().UpsertLocation(ctx, core.UpsertLocationParams{
				ID:            location.ID,
				IntegrationID: location.IntegrationID,
				ExternalID:    location.ExternalID,
				Name:          location.Name,
				Address:       location.Address,
				Country:       location.Country,
				Province:      location.Province,
				IsActive:      location.IsActive,
			})
			if err != nil {
				return errors.Wrapf(err, "failed to upsert location %v", location.ExternalID)
			}
		}

		logger.Info("Upserted locations batch", "start", i, "end", end, "count", len(batch))
	}
	return nil
}

// batchInsertProducts inserts products using upsert to handle conflicts
func (m *InventorySyncManager) batchInsertProducts(ctx context.Context, tx *db.TxDB, products []core.InsertProductsBatchParams, batchSize int) error {
	for i := 0; i < len(products); i += batchSize {
		end := i + batchSize
		if end > len(products) {
			end = len(products)
		}

		batch := products[i:end]
		
		// Use individual upserts instead of batch insert to handle ON CONFLICT
		for _, product := range batch {
			_, err := tx.GetCore().UpsertProduct(ctx, core.UpsertProductParams{
				ID:            product.ID,
				IntegrationID: product.IntegrationID,
				ExternalID:    product.ExternalID,
				Title:         product.Title,
				Handle:        product.Handle,
				ProductType:   product.ProductType,
				Status:        product.Status,
			})
			if err != nil {
				return errors.Wrapf(err, "failed to upsert product %s", product.Handle)
			}
		}

		logger.Info("Upserted products batch", "start", i, "end", end, "count", len(batch))
	}
	return nil
}

// batchInsertProductVariants inserts product variants using upsert to handle conflicts
func (m *InventorySyncManager) batchInsertProductVariants(ctx context.Context, tx *db.TxDB, variants []core.InsertProductVariantsBatchParams, batchSize int) error {
	for i := 0; i < len(variants); i += batchSize {
		end := i + batchSize
		if end > len(variants) {
			end = len(variants)
		}

		batch := variants[i:end]
		
		// Use individual upserts instead of batch insert to handle ON CONFLICT
		for _, variant := range batch {
			_, err := tx.GetCore().UpsertProductVariant(ctx, core.UpsertProductVariantParams{
				ID:              variant.ID,
				ProductID:       variant.ProductID,
				ExternalID:      variant.ExternalID,
				Sku:             variant.Sku,
				Price:           variant.Price,
				InventoryItemID: variant.InventoryItemID,
			})
			if err != nil {
				return errors.Wrapf(err, "failed to upsert product variant %v", variant.ExternalID)
			}
		}

		logger.Info("Upserted product variants batch", "start", i, "end", end, "count", len(batch))
	}
	return nil
}

// batchInsertInventoryItems inserts inventory items using upsert to handle conflicts
func (m *InventorySyncManager) batchInsertInventoryItems(ctx context.Context, tx *db.TxDB, items []core.InsertInventoryItemsBatchParams, batchSize int) error {
	for i := 0; i < len(items); i += batchSize {
		end := i + batchSize
		if end > len(items) {
			end = len(items)
		}

		batch := items[i:end]
		
		// Use individual upserts instead of batch insert to handle ON CONFLICT
		for _, item := range batch {
			_, err := tx.GetCore().UpsertInventoryItem(ctx, core.UpsertInventoryItemParams{
				ID:            item.ID,
				IntegrationID: item.IntegrationID,
				ExternalID:    item.ExternalID,
				Sku:           item.Sku,
				Tracked:       item.Tracked,
				Cost:          item.Cost,
			})
			if err != nil {
				return errors.Wrapf(err, "failed to upsert inventory item %v", item.ExternalID)
			}
		}

		logger.Info("Upserted inventory items batch", "start", i, "end", end, "count", len(batch))
	}
	return nil
}

// batchInsertOrders inserts orders using upsert to handle conflicts
func (m *InventorySyncManager) batchInsertOrders(ctx context.Context, tx *db.TxDB, orders []core.InsertOrdersBatchParams, batchSize int) error {
	for i := 0; i < len(orders); i += batchSize {
		end := i + batchSize
		if end > len(orders) {
			end = len(orders)
		}

		batch := orders[i:end]
		
		// Use individual upserts instead of batch insert to handle ON CONFLICT
		for _, order := range batch {
			_, err := tx.GetCore().UpsertOrder(ctx, core.UpsertOrderParams{
				ID:                order.ID,
				IntegrationID:     order.IntegrationID,
				ExternalID:        order.ExternalID,
				CreatedAt:         order.CreatedAt,
				FinancialStatus:   order.FinancialStatus,
				FulfillmentStatus: order.FulfillmentStatus,
				TotalPrice:        order.TotalPrice,
				CancelledAt:       order.CancelledAt,
			})
			if err != nil {
				return errors.Wrapf(err, "failed to upsert order %v", order.ExternalID)
			}
		}

		logger.Info("Upserted orders batch", "start", i, "end", end, "count", len(batch))
	}
	return nil
}
