package shopify

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strconv"
	"strings"
	"time"

	goshopify "github.com/bold-commerce/go-shopify/v4"
	"github.com/pkg/errors"
	"golang.org/x/time/rate"
)

// Client represents a Shopify API client
type Client struct {
	shopDomain  string
	accessToken string
	httpClient  *http.Client
	rateLimiter *rate.Limiter
	goShopify   *goshopify.Client
}

// NewClient creates a new Shopify API client
func NewClient(shopDomain, accessToken string) *Client {
	// Create go-shopify client for known methods
	goShopifyClient, err := goshopify.NewClient(goshopify.App{}, shopDomain, accessToken)
	if err != nil {
		goShopifyClient = nil // Fallback to custom implementation
	}

	return &Client{
		shopDomain:  shopDomain,
		accessToken: accessToken,
		httpClient: &http.Client{
			Timeout: 30 * time.Second,
		},
		// Shopify rate limit: 4 requests per second
		rateLimiter: rate.NewLimiter(4, 1),
		goShopify:   goShopifyClient,
	}
}

// makeRequest makes a request to the Shopify API with rate limiting
func (c *Client) makeRequest(ctx context.Context, method, path string, params url.Values) ([]byte, error) {
	// Wait for rate limiter
	if err := c.rateLimiter.Wait(ctx); err != nil {
		return nil, errors.Wrap(err, "rate limiter error")
	}

	requestURL := fmt.Sprintf("https://%s.myshopify.com/admin/api/2023-10%s", c.shopDomain, path)
	if len(params) > 0 {
		requestURL += "?" + params.Encode()
	}

	req, err := http.NewRequestWithContext(ctx, method, requestURL, nil)
	if err != nil {
		return nil, errors.Wrap(err, "failed to create request")
	}

	req.Header.Set("X-Shopify-Access-Token", c.accessToken)
	req.Header.Set("Content-Type", "application/json")

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, errors.Wrap(err, "failed to execute request")
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, errors.Wrap(err, "failed to read response body")
	}

	if resp.StatusCode >= 400 {
		return nil, fmt.Errorf("shopify API error: status %d, body: %s", resp.StatusCode, string(body))
	}

	return body, nil
}

// GetProducts retrieves products from Shopify
func (c *Client) GetProducts(ctx context.Context, limit int, pageInfo string) (*ProductsResponse, error) {
	params := url.Values{}
	if limit > 0 {
		params.Set("limit", strconv.Itoa(limit))
	}
	if pageInfo != "" {
		params.Set("page_info", pageInfo)
	}

	body, err := c.makeRequest(ctx, "GET", "/products.json", params)
	if err != nil {
		return nil, errors.Wrap(err, "failed to get products")
	}

	var response ProductsResponse
	if err := json.Unmarshal(body, &response); err != nil {
		return nil, errors.Wrap(err, "failed to unmarshal products response")
	}

	return &response, nil
}

// GetLocations retrieves locations from Shopify
func (c *Client) GetLocations(ctx context.Context) (*LocationsResponse, error) {
	body, err := c.makeRequest(ctx, "GET", "/locations.json", nil)
	if err != nil {
		return nil, errors.Wrap(err, "failed to get locations")
	}

	var response LocationsResponse
	if err := json.Unmarshal(body, &response); err != nil {
		return nil, errors.Wrap(err, "failed to unmarshal locations response")
	}

	return &response, nil
}

// GetInventoryLevels retrieves inventory levels for given inventory item IDs
func (c *Client) GetInventoryLevels(ctx context.Context, inventoryItemIDs []int64) (*InventoryLevelsResponse, error) {
	if len(inventoryItemIDs) == 0 {
		return &InventoryLevelsResponse{}, nil
	}

	params := url.Values{}

	// Convert IDs to comma-separated string
	idStrings := make([]string, len(inventoryItemIDs))
	for i, id := range inventoryItemIDs {
		idStrings[i] = strconv.FormatInt(id, 10)
	}
	params.Set("inventory_item_ids", strings.Join(idStrings, ","))

	body, err := c.makeRequest(ctx, "GET", "/inventory_levels.json", params)
	if err != nil {
		return nil, errors.Wrap(err, "failed to get inventory levels")
	}

	var response InventoryLevelsResponse
	if err := json.Unmarshal(body, &response); err != nil {
		return nil, errors.Wrap(err, "failed to unmarshal inventory levels response")
	}

	return &response, nil
}

// GetInventoryItem retrieves a single inventory item
func (c *Client) GetInventoryItem(ctx context.Context, inventoryItemID int64) (*InventoryItemResponse, error) {
	path := fmt.Sprintf("/inventory_items/%d.json", inventoryItemID)

	body, err := c.makeRequest(ctx, "GET", path, nil)
	if err != nil {
		return nil, errors.Wrap(err, "failed to get inventory item")
	}

	var response InventoryItemResponse
	if err := json.Unmarshal(body, &response); err != nil {
		return nil, errors.Wrap(err, "failed to unmarshal inventory item response")
	}

	return &response, nil
}

// GetOrders retrieves orders from Shopify with date filtering
func (c *Client) GetOrders(ctx context.Context, createdAtMin time.Time, limit int, pageInfo string) (*OrdersResponse, error) {
	params := url.Values{}
	params.Set("status", "any")
	params.Set("created_at_min", createdAtMin.Format(time.RFC3339))

	if limit > 0 {
		params.Set("limit", strconv.Itoa(limit))
	}
	if pageInfo != "" {
		params.Set("page_info", pageInfo)
	}

	body, err := c.makeRequest(ctx, "GET", "/orders.json", params)
	if err != nil {
		return nil, errors.Wrap(err, "failed to get orders")
	}

	var response OrdersResponse
	if err := json.Unmarshal(body, &response); err != nil {
		return nil, errors.Wrap(err, "failed to unmarshal orders response")
	}

	return &response, nil
}

// GetShop retrieves shop information from Shopify
func (c *Client) GetShop(ctx context.Context) (*goshopify.Shop, error) {

	// Use the go-shopify client for the GetShop method
	shop, err := c.goShopify.Shop.Get(ctx, nil)
	if err != nil {
		return nil, errors.Wrap(err, "failed to get shop using go-shopify client")
	}
	return shop, nil
}
