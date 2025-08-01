package worker

import (
	"context"

	"github.com/ConradKurth/forecasting/backend/internal/config"
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
func (c *Client) EnqueueShopifyStoreSync(ctx context.Context, userID, shopName, token string) error {
	task, err := NewShopifyStoreSyncTask(userID, shopName, token)
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
