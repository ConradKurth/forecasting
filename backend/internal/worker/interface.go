package worker

import "context"

// Queue represents a job queue interface for enqueueing background tasks
type Queue interface {
	// EnqueueShopifyStoreSync enqueues a Shopify store sync task
	EnqueueShopifyStoreSync(ctx context.Context, userID, shopName, token string) error

	// Close closes the queue client connection
	Close() error
}
