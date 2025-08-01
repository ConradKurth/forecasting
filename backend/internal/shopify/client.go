package shopify

import (
	"context"
	"fmt"

	goshopify "github.com/bold-commerce/go-shopify/v4"
)

// Client is a wrapper around the go-shopify SDK
type Client struct {
	client *goshopify.Client
}

// NewClient creates a new Shopify API client using the go-shopify SDK
func NewClient(shopDomain, accessToken string) (*Client, error) {
	// Create app configuration (minimal for API access)
	app := goshopify.App{}

	// Create client with access token
	client, err := goshopify.NewClient(app, shopDomain, accessToken,
		goshopify.WithVersion("2023-10"), // Use a stable API version
		goshopify.WithRetry(3),           // Retry on rate limits
	)
	if err != nil {
		return nil, fmt.Errorf("failed to create shopify client: %w", err)
	}

	return &Client{client: client}, nil
}

// ShopInfo contains the essential shop information we want to sync
type ShopInfo struct {
	Name     string
	Currency string
	Timezone string
	Domain   string
}

// GetShopInfo retrieves and formats the essential shop information for syncing
func (c *Client) GetShopInfo(ctx context.Context) (*ShopInfo, error) {
	shop, err := c.client.Shop.Get(ctx, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to get shop info: %w", err)
	}

	return &ShopInfo{
		Name:     shop.Name,
		Currency: shop.Currency,
		Timezone: shop.IanaTimezone,
		Domain:   shop.Domain,
	}, nil
}
