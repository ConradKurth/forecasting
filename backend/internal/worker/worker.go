package worker

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/ConradKurth/forecasting/backend/internal/factory"
	"github.com/ConradKurth/forecasting/backend/internal/shopify"
	"github.com/ConradKurth/forecasting/backend/pkg/id"
	"github.com/ConradKurth/forecasting/backend/pkg/logger"
	"github.com/hibiken/asynq"
)

// Worker handles background job processing
type Worker struct {
	serviceFactory *factory.ServiceFactory
}

// New creates a new worker instance
func New(serviceFactory *factory.ServiceFactory) *Worker {
	return &Worker{
		serviceFactory: serviceFactory,
	}
}

// RegisterHandlers registers all task handlers with the given mux
func (w *Worker) RegisterHandlers(mux *asynq.ServeMux) {
	mux.HandleFunc(TypeShopifyStoreSync, w.HandleShopifyStoreSync)
	mux.HandleFunc(TypeShopifyInventorySync, w.HandleShopifyInventorySync)
	mux.HandleFunc(TypeShopifyLocationsSync, w.HandleShopifyLocationsSync)
	mux.HandleFunc(TypeShopifyProductsSync, w.HandleShopifyProductsSync)
	mux.HandleFunc(TypeShopifyOrdersSync, w.HandleShopifyOrdersSync)
}

// HandleShopifyStoreSync processes Shopify store synchronization tasks
func (w *Worker) HandleShopifyStoreSync(ctx context.Context, t *asynq.Task) error {
	var payload ShopifyStoreSyncPayload
	if err := json.Unmarshal(t.Payload(), &payload); err != nil {
		return fmt.Errorf("failed to unmarshal shopify store sync payload: %w", err)
	}

	// Fetch shop info using Shopify API client
	shopifyClient := shopify.NewClient(payload.ShopName, payload.Token)

	var shopInfo struct {
		Name     string
		Currency string
		Timezone string
	}

	shop, err := shopifyClient.GetShop(ctx)
	if err != nil {
		logger.Error("Failed to fetch shop info", "shop", payload.ShopName, "error", err)
		// Use fallback shop info
		shopInfo = struct {
			Name     string
			Currency string
			Timezone string
		}{
			Name:     payload.ShopName,
			Currency: "USD",
			Timezone: "UTC",
		}
		logger.Info("Using fallback shop info", "shop", payload.ShopName)
	} else {
		// Use fetched shop info from Shopify
		shopInfo = struct {
			Name     string
			Currency string
			Timezone string
		}{
			Name:     shop.Name,
			Currency: shop.Currency,
			Timezone: shop.IanaTimezone,
		}
	}

	// Get Shopify store service
	shopifyStoreService := w.serviceFactory.CreateShopifyStoreService()

	// Update store with fetched information
	_, err = shopifyStoreService.CreateOrUpdateStore(
		ctx,
		payload.ShopName,
		&shopInfo.Name,
		&shopInfo.Timezone,
		&shopInfo.Currency,
	)
	if err != nil {
		return fmt.Errorf("failed to update store in database: %w", err)
	}

	return nil
}

// HandleShopifyInventorySync processes full Shopify inventory synchronization tasks
func (w *Worker) HandleShopifyInventorySync(ctx context.Context, t *asynq.Task) error {
	var payload ShopifyInventorySyncPayload
	if err := json.Unmarshal(t.Payload(), &payload); err != nil {
		return fmt.Errorf("failed to unmarshal shopify inventory sync payload: %w", err)
	}

	integrationID, err := id.ParseTyped[id.PlatformIntegration](payload.IntegrationID)
	if err != nil {
		return fmt.Errorf("failed to parse integration ID: %w", err)
	}

	coreInventoryService := w.serviceFactory.CreateCoreInventoryService()
	return coreInventoryService.SyncShopifyData(ctx, integrationID, payload.ShopDomain, payload.AccessToken)
}

// HandleShopifyLocationsSync processes Shopify locations synchronization tasks
func (w *Worker) HandleShopifyLocationsSync(ctx context.Context, t *asynq.Task) error {
	var payload ShopifyInventorySyncPayload
	if err := json.Unmarshal(t.Payload(), &payload); err != nil {
		return fmt.Errorf("failed to unmarshal shopify locations sync payload: %w", err)
	}

	integrationID, err := id.ParseTyped[id.PlatformIntegration](payload.IntegrationID)
	if err != nil {
		return fmt.Errorf("failed to parse integration ID: %w", err)
	}

	coreInventoryService := w.serviceFactory.CreateCoreInventoryService()
	coreInventoryService.SetShopifyClient(shopify.NewClient(payload.ShopDomain, payload.AccessToken))
	return coreInventoryService.SyncLocations(ctx, integrationID)
}

// HandleShopifyProductsSync processes Shopify products synchronization tasks
func (w *Worker) HandleShopifyProductsSync(ctx context.Context, t *asynq.Task) error {
	var payload ShopifyInventorySyncPayload
	if err := json.Unmarshal(t.Payload(), &payload); err != nil {
		return fmt.Errorf("failed to unmarshal shopify products sync payload: %w", err)
	}

	integrationID, err := id.ParseTyped[id.PlatformIntegration](payload.IntegrationID)
	if err != nil {
		return fmt.Errorf("failed to parse integration ID: %w", err)
	}

	coreInventoryService := w.serviceFactory.CreateCoreInventoryService()
	coreInventoryService.SetShopifyClient(shopify.NewClient(payload.ShopDomain, payload.AccessToken))
	return coreInventoryService.SyncProducts(ctx, integrationID)
}

// HandleShopifyOrdersSync processes Shopify orders synchronization tasks
func (w *Worker) HandleShopifyOrdersSync(ctx context.Context, t *asynq.Task) error {
	var payload ShopifyInventorySyncPayload
	if err := json.Unmarshal(t.Payload(), &payload); err != nil {
		return fmt.Errorf("failed to unmarshal shopify orders sync payload: %w", err)
	}

	integrationID, err := id.ParseTyped[id.PlatformIntegration](payload.IntegrationID)
	if err != nil {
		return fmt.Errorf("failed to parse integration ID: %w", err)
	}

	coreInventoryService := w.serviceFactory.CreateCoreInventoryService()
	coreInventoryService.SetShopifyClient(shopify.NewClient(payload.ShopDomain, payload.AccessToken))
	return coreInventoryService.SyncOrders(ctx, integrationID)
}
