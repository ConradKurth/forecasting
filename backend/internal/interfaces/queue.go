package interfaces

import (
	"context"

	"github.com/ConradKurth/forecasting/backend/pkg/id"
)

// Queue represents a job queue interface for enqueueing background tasks
type Queue interface {
	// EnqueueShopifyStoreSync enqueues a Shopify store sync task
	EnqueueShopifyStoreSync(ctx context.Context, userID, shopID, token string) error

	// EnqueueShopifyInventorySync enqueues a Shopify inventory sync task
	EnqueueShopifyInventorySync(ctx context.Context, integrationID, shopDomain, accessToken string) error

	// Close closes the queue client connection
	Close() error
}

// ShopifyManager interface defines what the worker needs from a Shopify manager
type ShopifyManager interface {
	// SyncStoreInfo updates store information by fetching from Shopify API
	SyncStoreInfo(ctx context.Context, userID id.ID[id.User], shopID id.ID[id.ShopifyStore]) error
}

// InventorySyncManager interface defines what the worker needs from an inventory sync manager
type InventorySyncManager interface {
	// SyncInventory performs inventory synchronization
	SyncInventory(ctx context.Context, integrationID id.ID[id.PlatformIntegration]) error
}
