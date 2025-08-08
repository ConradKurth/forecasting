package shopify

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"reflect"
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

// ResponseWithPagination wraps response data with pagination info
type ResponseWithPagination struct {
	Data       []byte
	Pagination PaginationInfo
}

// makeRequest makes a request to the Shopify API with rate limiting
func (c *Client) makeRequest(ctx context.Context, method, path string, params url.Values) ([]byte, error) {
	resp, err := c.makeRequestWithPagination(ctx, method, path, params)
	if err != nil {
		return nil, err
	}
	return resp.Data, nil
}

// makeRequestWithPagination makes a request to the Shopify API with rate limiting and returns pagination info
func (c *Client) makeRequestWithPagination(ctx context.Context, method, path string, params url.Values) (*ResponseWithPagination, error) {
	// Wait for rate limiter
	if err := c.rateLimiter.Wait(ctx); err != nil {
		return nil, errors.Wrap(err, "rate limiter error")
	}

	requestURL := fmt.Sprintf("https://%s/admin/api/2023-10%s", c.shopDomain, path)
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

	// Extract pagination info from Link header
	pagination := PaginationInfo{}
	if linkHeader := resp.Header.Get("Link"); linkHeader != "" {
		pagination = c.parseLinkHeader(linkHeader)
	}

	return &ResponseWithPagination{
		Data:       body,
		Pagination: pagination,
	}, nil
}

// parseLinkHeader parses the Link header to extract pagination information
func (c *Client) parseLinkHeader(linkHeader string) PaginationInfo {
	pagination := PaginationInfo{}

	// Split by comma to get individual links
	links := strings.Split(linkHeader, ",")

	for _, link := range links {
		link = strings.TrimSpace(link)

		// Extract URL and relation
		parts := strings.Split(link, ";")
		if len(parts) < 2 {
			continue
		}

		urlPart := strings.TrimSpace(parts[0])
		relPart := strings.TrimSpace(parts[1])

		// Remove angle brackets from URL
		if strings.HasPrefix(urlPart, "<") && strings.HasSuffix(urlPart, ">") {
			urlPart = urlPart[1 : len(urlPart)-1]
		}

		// Parse the page_info from URL
		if parsedURL, err := url.Parse(urlPart); err == nil {
			pageInfo := parsedURL.Query().Get("page_info")

			if strings.Contains(relPart, `rel="next"`) {
				pagination.NextPageInfo = pageInfo
			} else if strings.Contains(relPart, `rel="previous"`) {
				pagination.PreviousPageInfo = pageInfo
			}
		}
	}

	return pagination
}

// addPaginationParams adds common pagination parameters to url.Values
func addPaginationParams(params url.Values, limit int, pageInfo string) {
	if limit > 0 {
		params.Set("limit", strconv.Itoa(limit))
	}
	if pageInfo != "" {
		params.Set("page_info", pageInfo)
	}
}

// convertIDsToString converts a slice of int64 IDs to a comma-separated string
func convertIDsToString(ids []int64) string {
	idStrings := make([]string, len(ids))
	for i, id := range ids {
		idStrings[i] = strconv.FormatInt(id, 10)
	}
	return strings.Join(idStrings, ",")
}

// makePaginatedRequest is a generic helper for making paginated requests and unmarshaling responses
func (c *Client) makePaginatedRequest(ctx context.Context, endpoint string, params url.Values, target interface{}) error {
	resp, err := c.makeRequestWithPagination(ctx, "GET", endpoint, params)
	if err != nil {
		return err
	}

	if err := json.Unmarshal(resp.Data, target); err != nil {
		return errors.Wrap(err, "failed to unmarshal response")
	}

	// Set pagination info using reflection
	// This assumes the target has a Pagination field of type PaginationInfo
	v := reflect.ValueOf(target).Elem()
	if paginationField := v.FieldByName("Pagination"); paginationField.IsValid() && paginationField.CanSet() {
		paginationField.Set(reflect.ValueOf(resp.Pagination))
	}

	return nil
}

// GetProducts retrieves products from Shopify
func (c *Client) GetProducts(ctx context.Context, limit int, pageInfo string) (*ProductsResponse, error) {
	params := url.Values{}
	addPaginationParams(params, limit, pageInfo)

	var response ProductsResponse
	if err := c.makePaginatedRequest(ctx, "/products.json", params, &response); err != nil {
		return nil, errors.Wrap(err, "failed to get products")
	}

	return &response, nil
}

// GetLocations retrieves locations from Shopify with pagination support
func (c *Client) GetLocations(ctx context.Context, limit int, pageInfo string) (*LocationsResponse, error) {
	params := url.Values{}
	addPaginationParams(params, limit, pageInfo)

	var response LocationsResponse
	if err := c.makePaginatedRequest(ctx, "/locations.json", params, &response); err != nil {
		return nil, errors.Wrap(err, "failed to get locations")
	}

	return &response, nil
}

// GetInventoryLevels retrieves inventory levels for given inventory item IDs with pagination support
func (c *Client) GetInventoryLevels(ctx context.Context, inventoryItemIDs []int64, limit int, pageInfo string) (*InventoryLevelsResponse, error) {
	if len(inventoryItemIDs) == 0 {
		return &InventoryLevelsResponse{}, nil
	}

	params := url.Values{}
	params.Set("inventory_item_ids", convertIDsToString(inventoryItemIDs))
	addPaginationParams(params, limit, pageInfo)

	var response InventoryLevelsResponse
	if err := c.makePaginatedRequest(ctx, "/inventory_levels.json", params, &response); err != nil {
		return nil, errors.Wrap(err, "failed to get inventory levels")
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

// GetInventoryItems retrieves multiple inventory items with pagination support
func (c *Client) GetInventoryItems(ctx context.Context, inventoryItemIDs []int64, limit int, pageInfo string) (*InventoryItemsResponse, error) {
	if len(inventoryItemIDs) == 0 {
		return &InventoryItemsResponse{}, nil
	}

	params := url.Values{}
	params.Set("ids", convertIDsToString(inventoryItemIDs))
	addPaginationParams(params, limit, pageInfo)

	var response InventoryItemsResponse
	if err := c.makePaginatedRequest(ctx, "/inventory_items.json", params, &response); err != nil {
		return nil, errors.Wrap(err, "failed to get inventory items")
	}

	return &response, nil
}

// GetOrders retrieves orders from Shopify with date filtering
func (c *Client) GetOrders(ctx context.Context, createdAtMin time.Time, limit int, pageInfo string) (*OrdersResponse, error) {
	params := url.Values{}
	params.Set("status", "any")
	params.Set("created_at_min", createdAtMin.Format(time.RFC3339))
	params.Set("fields", "id,name,created_at,updated_at,financial_status,fulfillment_status,total_price,currency,line_items")
	addPaginationParams(params, limit, pageInfo)

	var response OrdersResponse
	if err := c.makePaginatedRequest(ctx, "/orders.json", params, &response); err != nil {
		return nil, errors.Wrap(err, "failed to get orders")
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
