package shopify

import (
	"time"
)

// ShopifyProduct represents a product from Shopify API
type ShopifyProduct struct {
	ID          int64                  `json:"id"`
	Title       string                 `json:"title"`
	Handle      string                 `json:"handle"`
	ProductType string                 `json:"product_type"`
	Status      string                 `json:"status"`
	Variants    []ShopifyProductVariant `json:"variants"`
	CreatedAt   time.Time              `json:"created_at"`
	UpdatedAt   time.Time              `json:"updated_at"`
}

// ShopifyProductVariant represents a product variant from Shopify API
type ShopifyProductVariant struct {
	ID                int64   `json:"id"`
	ProductID         int64   `json:"product_id"`
	SKU               string  `json:"sku"`
	Price             string  `json:"price"`
	InventoryItemID   int64   `json:"inventory_item_id"`
	InventoryQuantity int     `json:"inventory_quantity"`
	CreatedAt         time.Time `json:"created_at"`
	UpdatedAt         time.Time `json:"updated_at"`
}

// ShopifyInventoryItem represents an inventory item from Shopify API
type ShopifyInventoryItem struct {
	ID       int64   `json:"id"`
	SKU      string  `json:"sku"`
	Tracked  bool    `json:"tracked"`
	Cost     string  `json:"cost"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}

// ShopifyInventoryLevel represents an inventory level from Shopify API
type ShopifyInventoryLevel struct {
	InventoryItemID int64 `json:"inventory_item_id"`
	LocationID      int64 `json:"location_id"`
	Available       int   `json:"available"`
	UpdatedAt       time.Time `json:"updated_at"`
}

// ShopifyLocation represents a location from Shopify API
type ShopifyLocation struct {
	ID        int64  `json:"id"`
	Name      string `json:"name"`
	Address1  string `json:"address1"`
	Address2  string `json:"address2"`
	City      string `json:"city"`
	Country   string `json:"country"`
	Province  string `json:"province"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}

// ShopifyOrder represents an order from Shopify API
type ShopifyOrder struct {
	ID                int64                     `json:"id"`
	CreatedAt         time.Time                 `json:"created_at"`
	FinancialStatus   string                    `json:"financial_status"`
	FulfillmentStatus *string                   `json:"fulfillment_status"`
	TotalPrice        string                    `json:"total_price"`
	CancelledAt       *time.Time                `json:"cancelled_at"`
	LineItems         []ShopifyOrderLineItem    `json:"line_items"`
}

// ShopifyOrderLineItem represents an order line item from Shopify API
type ShopifyOrderLineItem struct {
	ID        int64  `json:"id"`
	ProductID *int64 `json:"product_id"`
	VariantID *int64 `json:"variant_id"`
	Quantity  int    `json:"quantity"`
	Price     string `json:"price"`
}

// API Response wrappers
type ProductsResponse struct {
	Products []ShopifyProduct `json:"products"`
}

type ProductResponse struct {
	Product ShopifyProduct `json:"product"`
}

type InventoryLevelsResponse struct {
	InventoryLevels []ShopifyInventoryLevel `json:"inventory_levels"`
}

type InventoryItemResponse struct {
	InventoryItem ShopifyInventoryItem `json:"inventory_item"`
}

type LocationsResponse struct {
	Locations []ShopifyLocation `json:"locations"`
}

type OrdersResponse struct {
	Orders []ShopifyOrder `json:"orders"`
}

type OrderResponse struct {
	Order ShopifyOrder `json:"order"`
}
