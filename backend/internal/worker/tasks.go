package worker

import (
	"encoding/json"

	"github.com/hibiken/asynq"
)

// Task types
const (
	TypeShopifyStoreSync = "shopify:store_sync"
)

// ShopifyStoreSyncPayload contains data needed for Shopify store sync
type ShopifyStoreSyncPayload struct {
	UserID   string `json:"user_id"`
	ShopName string `json:"shop_name"`
	Token    string `json:"token"`
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
