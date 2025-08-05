package worker

import (
	"encoding/json"

	"github.com/ConradKurth/forecasting/backend/pkg/id"
	"github.com/hibiken/asynq"
)

// Task types
const (
	TypeShopifyStoreSync     = "shopify:store_sync"
	TypeShopifyInventorySync = "shopify:inventory_sync"
	TypeShopifyLocationsSync = "shopify:locations_sync"
	TypeShopifyProductsSync  = "shopify:products_sync"
	TypeShopifyOrdersSync    = "shopify:orders_sync"
)

// ShopifyStoreSyncPayload contains data needed for Shopify store sync
type ShopifyStoreSyncPayload struct {
	UserID id.ID[id.User]         `json:"user_id"`
	ShopID id.ID[id.ShopifyStore] `json:"shop_id"`
}

// ShopifyInventorySyncPayload contains data needed for Shopify inventory sync
type ShopifyInventorySyncPayload struct {
	IntegrationID id.ID[id.PlatformIntegration] `json:"integration_id"`
}

// NewShopifyStoreSyncTask creates a new task for syncing Shopify store data
func NewShopifyStoreSyncTask(userID id.ID[id.User], shopID id.ID[id.ShopifyStore]) (*asynq.Task, error) {
	payload := ShopifyStoreSyncPayload{
		UserID: userID,
		ShopID: shopID,
	}

	data, err := json.Marshal(payload)
	if err != nil {
		return nil, err
	}

	return asynq.NewTask(TypeShopifyStoreSync, data), nil
}

// NewShopifyInventorySyncTask creates a new task for syncing Shopify inventory data
func NewShopifyInventorySyncTask(integrationID id.ID[id.PlatformIntegration]) (*asynq.Task, error) {
	payload := ShopifyInventorySyncPayload{
		IntegrationID: integrationID,
	}

	data, err := json.Marshal(payload)
	if err != nil {
		return nil, err
	}

	return asynq.NewTask(TypeShopifyInventorySync, data), nil
}

// NewShopifyLocationsSyncTask creates a new task for syncing Shopify locations
func NewShopifyLocationsSyncTask(integrationID id.ID[id.PlatformIntegration]) (*asynq.Task, error) {
	payload := ShopifyInventorySyncPayload{
		IntegrationID: integrationID,
	}

	data, err := json.Marshal(payload)
	if err != nil {
		return nil, err
	}

	return asynq.NewTask(TypeShopifyLocationsSync, data), nil
}

// NewShopifyProductsSyncTask creates a new task for syncing Shopify products
func NewShopifyProductsSyncTask(integrationID id.ID[id.PlatformIntegration]) (*asynq.Task, error) {
	payload := ShopifyInventorySyncPayload{
		IntegrationID: integrationID,
	}

	data, err := json.Marshal(payload)
	if err != nil {
		return nil, err
	}

	return asynq.NewTask(TypeShopifyProductsSync, data), nil
}

// NewShopifyOrdersSyncTask creates a new task for syncing Shopify orders
func NewShopifyOrdersSyncTask(integrationID id.ID[id.PlatformIntegration]) (*asynq.Task, error) {
	payload := ShopifyInventorySyncPayload{
		IntegrationID: integrationID,
	}

	data, err := json.Marshal(payload)
	if err != nil {
		return nil, err
	}

	return asynq.NewTask(TypeShopifyOrdersSync, data), nil
}
