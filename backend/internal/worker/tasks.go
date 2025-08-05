package worker

import (
	"encoding/json"

	"github.com/hibiken/asynq"
)

// Task types
const (
	TypeShopifyStoreSync      = "shopify:store_sync"
	TypeShopifyInventorySync  = "shopify:inventory_sync"
	TypeShopifyLocationsSync  = "shopify:locations_sync"
	TypeShopifyProductsSync   = "shopify:products_sync"
	TypeShopifyOrdersSync     = "shopify:orders_sync"
)

// ShopifyStoreSyncPayload contains data needed for Shopify store sync
type ShopifyStoreSyncPayload struct {
	UserID   string `json:"user_id"`
	ShopName string `json:"shop_name"`
	Token    string `json:"token"`
}

// ShopifyInventorySyncPayload contains data needed for Shopify inventory sync
type ShopifyInventorySyncPayload struct {
	IntegrationID string `json:"integration_id"`
	ShopDomain    string `json:"shop_domain"`
	AccessToken   string `json:"access_token"`
}

// NewShopifyStoreSyncTask creates a new task for syncing Shopify store data
func NewShopifyStoreSyncTask(userID, shopName, token string) (*asynq.Task, error) {
	payload := ShopifyStoreSyncPayload{
		UserID:   userID,
		ShopName: shopName,
		Token:    token,
	}

	data, err := json.Marshal(payload)
	if err != nil {
		return nil, err
	}

	return asynq.NewTask(TypeShopifyStoreSync, data), nil
}

// NewShopifyInventorySyncTask creates a new task for syncing Shopify inventory data
func NewShopifyInventorySyncTask(integrationID, shopDomain, accessToken string) (*asynq.Task, error) {
	payload := ShopifyInventorySyncPayload{
		IntegrationID: integrationID,
		ShopDomain:    shopDomain,
		AccessToken:   accessToken,
	}

	data, err := json.Marshal(payload)
	if err != nil {
		return nil, err
	}

	return asynq.NewTask(TypeShopifyInventorySync, data), nil
}

// NewShopifyLocationsSyncTask creates a new task for syncing Shopify locations
func NewShopifyLocationsSyncTask(integrationID, shopDomain, accessToken string) (*asynq.Task, error) {
	payload := ShopifyInventorySyncPayload{
		IntegrationID: integrationID,
		ShopDomain:    shopDomain,
		AccessToken:   accessToken,
	}

	data, err := json.Marshal(payload)
	if err != nil {
		return nil, err
	}

	return asynq.NewTask(TypeShopifyLocationsSync, data), nil
}

// NewShopifyProductsSyncTask creates a new task for syncing Shopify products
func NewShopifyProductsSyncTask(integrationID, shopDomain, accessToken string) (*asynq.Task, error) {
	payload := ShopifyInventorySyncPayload{
		IntegrationID: integrationID,
		ShopDomain:    shopDomain,
		AccessToken:   accessToken,
	}

	data, err := json.Marshal(payload)
	if err != nil {
		return nil, err
	}

	return asynq.NewTask(TypeShopifyProductsSync, data), nil
}

// NewShopifyOrdersSyncTask creates a new task for syncing Shopify orders
func NewShopifyOrdersSyncTask(integrationID, shopDomain, accessToken string) (*asynq.Task, error) {
	payload := ShopifyInventorySyncPayload{
		IntegrationID: integrationID,
		ShopDomain:    shopDomain,
		AccessToken:   accessToken,
	}

	data, err := json.Marshal(payload)
	if err != nil {
		return nil, err
	}

	return asynq.NewTask(TypeShopifyOrdersSync, data), nil
}
