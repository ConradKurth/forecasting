package worker

import (
	"context"

	"github.com/ConradKurth/forecasting/backend/internal/config"
	"github.com/ConradKurth/forecasting/backend/pkg/id"
	"github.com/hibiken/asynq"
)

// Client wraps asynq.Client for enqueueing tasks
type Client struct {
	client *asynq.Client
}

// NewClient creates a new worker client for enqueueing tasks
func NewClient() *Client {
	redisOpt := asynq.RedisClientOpt{
		Addr: config.Values.Redis.URL,
	}

	return &Client{
		client: asynq.NewClient(redisOpt),
	}
}

// EnqueueShopifyStoreSync enqueues a Shopify store sync task
func (c *Client) EnqueueShopifyStoreSync(ctx context.Context, userID, shopID, token string) error {
	// Parse userID back to the proper type
	userIDParsed, err := id.ParseTyped[id.User](userID)
	if err != nil {
		return err
	}
	
	// Parse shopID back to the proper type  
	shopIDParsed, err := id.ParseTyped[id.ShopifyStore](shopID)
	if err != nil {
		return err
	}
	
	task, err := NewShopifyStoreSyncTask(userIDParsed, shopIDParsed)
	if err != nil {
		return err
	}

	_, err = c.client.EnqueueContext(ctx, task)
	return err
}

// EnqueueShopifyInventorySync enqueues a Shopify inventory sync task
func (c *Client) EnqueueShopifyInventorySync(ctx context.Context, integrationID, shopDomain, accessToken string) error {
	// Parse integrationID back to the proper type
	integrationIDParsed, err := id.ParseTyped[id.PlatformIntegration](integrationID)
	if err != nil {
		return err
	}
	
	task, err := NewShopifyInventorySyncTask(integrationIDParsed)
	if err != nil {
		return err
	}

	_, err = c.client.EnqueueContext(ctx, task)
	return err
}

func (c *Client) EnqueueShopifyLocationsSync(ctx context.Context, integrationID id.ID[id.PlatformIntegration]) error {
	task, err := NewShopifyLocationsSyncTask(integrationID)
	if err != nil {
		return err
	}

	_, err = c.client.EnqueueContext(ctx, task)
	return err
}

func (c *Client) EnqueueShopifyProductsSync(ctx context.Context, integrationID id.ID[id.PlatformIntegration]) error {
	task, err := NewShopifyProductsSyncTask(integrationID)
	if err != nil {
		return err
	}

	_, err = c.client.EnqueueContext(ctx, task)
	return err
}

// Close closes the worker client connection
func (c *Client) Close() error {
	return c.client.Close()
}
