package worker

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/ConradKurth/forecasting/backend/internal/interfaces"
	"github.com/ConradKurth/forecasting/backend/pkg/logger"
	"github.com/hibiken/asynq"
)

// Worker handles background job processing
type Worker struct {
	shopifyManager interfaces.ShopifyManager
	syncManager    interfaces.InventorySyncManager
}

// New creates a new worker instance
func New(shopifyManager interfaces.ShopifyManager, syncManager interfaces.InventorySyncManager) *Worker {
	return &Worker{
		shopifyManager: shopifyManager,
		syncManager:    syncManager,
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

	// Use the ShopifyManager to sync store information
	err := w.shopifyManager.SyncStoreInfo(ctx, payload.UserID, payload.ShopID)
	if err != nil {
		return fmt.Errorf("failed to sync store info: %w", err)
	}

	logger.Info("Successfully synced store info", "user_id", payload.UserID, "shop_id", payload.ShopID)
	return nil
}

// HandleShopifyInventorySync processes full Shopify inventory synchronization tasks
func (w *Worker) HandleShopifyInventorySync(ctx context.Context, t *asynq.Task) error {
	var payload ShopifyInventorySyncPayload
	if err := json.Unmarshal(t.Payload(), &payload); err != nil {
		return fmt.Errorf("failed to unmarshal shopify inventory sync payload: %w", err)
	}

	logger.Info("Inventory sync requested", "integration_id", payload.IntegrationID)

	// Use the injected sync manager to perform the inventory sync
	return w.syncManager.SyncInventory(ctx, payload.IntegrationID)
}

// HandleShopifyLocationsSync processes Shopify locations synchronization tasks
func (w *Worker) HandleShopifyLocationsSync(ctx context.Context, t *asynq.Task) error {
	var payload ShopifyInventorySyncPayload
	if err := json.Unmarshal(t.Payload(), &payload); err != nil {
		return fmt.Errorf("failed to unmarshal shopify locations sync payload: %w", err)
	}

	logger.Info("Locations sync requested", "integration_id", payload.IntegrationID)
	return nil
}

// HandleShopifyProductsSync processes Shopify products synchronization tasks
func (w *Worker) HandleShopifyProductsSync(ctx context.Context, t *asynq.Task) error {
	var payload ShopifyInventorySyncPayload
	if err := json.Unmarshal(t.Payload(), &payload); err != nil {
		return fmt.Errorf("failed to unmarshal shopify products sync payload: %w", err)
	}

	logger.Info("Products sync requested", "integration_id", payload.IntegrationID)
	return nil
}

// HandleShopifyOrdersSync processes Shopify orders synchronization tasks
func (w *Worker) HandleShopifyOrdersSync(ctx context.Context, t *asynq.Task) error {
	var payload ShopifyInventorySyncPayload
	if err := json.Unmarshal(t.Payload(), &payload); err != nil {
		return fmt.Errorf("failed to unmarshal shopify orders sync payload: %w", err)
	}

	logger.Info("Orders sync requested", "integration_id", payload.IntegrationID)
	return nil
}
