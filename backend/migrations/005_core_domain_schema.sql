-- +goose Up
-- +goose StatementBegin

-- Core domain schema for multi-platform inventory forecasting
-- This schema is platform-agnostic and can work with Shopify, WooCommerce, etc.

-- Enum types
CREATE TYPE platform_type AS ENUM (
    'shopify',
    'woocommerce', 
    'bigcommerce',
    'square',
    'magento'
);

CREATE TYPE product_status AS ENUM (
    'active',
    'archived',
    'draft'
);

CREATE TYPE financial_status AS ENUM (
    'pending',
    'authorized',
    'partially_paid',
    'paid',
    'partially_refunded',
    'refunded',
    'voided'
);

CREATE TYPE fulfillment_status AS ENUM (
    'fulfilled',
    'null',
    'partial',
    'restocked'
);

CREATE TYPE sync_status AS ENUM (
    'pending',
    'in_progress',
    'completed',
    'failed'
);

CREATE TYPE entity_type AS ENUM (
    'products',
    'orders',
    'inventory',
    'locations',
    'full_sync'
);

-- Platform integrations table - tracks which platforms are connected to which shops
CREATE TABLE platform_integrations (
    id TEXT PRIMARY KEY,
    shop_id TEXT NOT NULL REFERENCES shopify_store(id),
    platform_type platform_type NOT NULL,
    platform_shop_id TEXT NOT NULL, -- External platform's shop identifier
    is_active BOOLEAN DEFAULT true,
    created_at TIMESTAMP DEFAULT NOW(),
    updated_at TIMESTAMP DEFAULT NOW(),
    UNIQUE(platform_shop_id, platform_type)
);

-- Locations - physical or virtual locations where inventory is stored
CREATE TABLE locations (
    id TEXT PRIMARY KEY,
    integration_id TEXT NOT NULL REFERENCES platform_integrations(id),
    external_id TEXT, -- Platform-specific location ID (e.g., Shopify location ID)
    name TEXT NOT NULL,
    address TEXT,
    country TEXT,
    province TEXT,
    is_active BOOLEAN DEFAULT true,
    created_at TIMESTAMP DEFAULT NOW(),
    updated_at TIMESTAMP DEFAULT NOW(),
    UNIQUE(integration_id, external_id)
);

-- Products - core product catalog
CREATE TABLE products (
    id TEXT PRIMARY KEY,
    integration_id TEXT NOT NULL REFERENCES platform_integrations(id),
    external_id TEXT, -- Platform-specific product ID
    title TEXT NOT NULL,
    handle TEXT NOT NULL,
    product_type TEXT,
    status product_status NOT NULL,
    created_at TIMESTAMP DEFAULT NOW(),
    updated_at TIMESTAMP DEFAULT NOW(),
    UNIQUE(integration_id, handle)
);

-- Product variants - different variations of products (size, color, etc.)
CREATE TABLE product_variants (
    id TEXT PRIMARY KEY,
    product_id TEXT NOT NULL REFERENCES products(id) ON DELETE CASCADE,
    external_id TEXT, -- Platform-specific variant ID
    sku TEXT,
    price DECIMAL(10,2),
    inventory_item_id TEXT,
    created_at TIMESTAMP DEFAULT NOW(),
    updated_at TIMESTAMP DEFAULT NOW(),
    UNIQUE(product_id, external_id)
);

-- Inventory items - represents items that can be tracked in inventory
CREATE TABLE inventory_items (
    id TEXT PRIMARY KEY,
    integration_id TEXT NOT NULL REFERENCES platform_integrations(id),
    external_id TEXT, -- Platform-specific inventory item ID
    sku TEXT,
    tracked BOOLEAN DEFAULT false,
    cost DECIMAL(10,2),
    created_at TIMESTAMP DEFAULT NOW(),
    updated_at TIMESTAMP DEFAULT NOW(),
    UNIQUE(integration_id, external_id)
);

-- Inventory levels - current stock levels at each location
CREATE TABLE inventory_levels (
    id TEXT PRIMARY KEY,
    inventory_item_id TEXT NOT NULL REFERENCES inventory_items(id),
    location_id TEXT NOT NULL REFERENCES locations(id),
    available INTEGER DEFAULT 0,
    updated_at TIMESTAMP DEFAULT NOW(),
    UNIQUE(inventory_item_id, location_id)
);

-- Orders - customer orders
CREATE TABLE orders (
    id TEXT PRIMARY KEY,
    integration_id TEXT NOT NULL REFERENCES platform_integrations(id),
    external_id TEXT, -- Platform-specific order ID
    created_at TIMESTAMP NOT NULL,
    financial_status financial_status NOT NULL DEFAULT 'pending',
    fulfillment_status fulfillment_status NOT NULL DEFAULT 'null',
    total_price DECIMAL(10,2),
    cancelled_at TIMESTAMP,
    UNIQUE(integration_id, external_id)
);

-- Order line items - individual items within orders
CREATE TABLE order_line_items (
    id TEXT PRIMARY KEY,
    order_id TEXT NOT NULL REFERENCES orders(id) ON DELETE CASCADE,
    external_id TEXT, -- Platform-specific line item ID
    product_id TEXT REFERENCES products(id),
    variant_id TEXT REFERENCES product_variants(id),
    inventory_item_id TEXT REFERENCES inventory_items(id),
    quantity INTEGER NOT NULL,
    price DECIMAL(10,2) NOT NULL,
    UNIQUE(order_id, external_id)
);

-- Sync state tracking - tracks when each integration was last synced
CREATE TABLE sync_states (
    id TEXT PRIMARY KEY,
    integration_id TEXT NOT NULL REFERENCES platform_integrations(id),
    entity_type entity_type NOT NULL,
    last_synced_at TIMESTAMP DEFAULT NOW(),
    sync_status sync_status NOT NULL DEFAULT 'pending',
    error_message TEXT,
    UNIQUE(integration_id, entity_type)
);

-- Indexes for performance
CREATE INDEX idx_platform_integrations_shop_id ON platform_integrations(shop_id);
CREATE INDEX idx_platform_integrations_platform_type ON platform_integrations(platform_type);
CREATE INDEX idx_locations_integration_id ON locations(integration_id);
CREATE INDEX idx_products_integration_id ON products(integration_id);
CREATE INDEX idx_products_handle ON products(handle);
CREATE INDEX idx_product_variants_product_id ON product_variants(product_id);
CREATE INDEX idx_product_variants_inventory_item_id ON product_variants(inventory_item_id);
CREATE INDEX idx_inventory_items_integration_id ON inventory_items(integration_id);
CREATE INDEX idx_inventory_levels_inventory_item_id ON inventory_levels(inventory_item_id);
CREATE INDEX idx_inventory_levels_location_id ON inventory_levels(location_id);
CREATE INDEX idx_orders_integration_id ON orders(integration_id);
CREATE INDEX idx_orders_created_at ON orders(created_at);
CREATE INDEX idx_order_line_items_order_id ON order_line_items(order_id);
CREATE INDEX idx_order_line_items_product_id ON order_line_items(product_id);
CREATE INDEX idx_order_line_items_variant_id ON order_line_items(variant_id);
CREATE INDEX idx_sync_states_integration_id ON sync_states(integration_id);
CREATE INDEX idx_sync_states_entity_type ON sync_states(entity_type);

-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin

-- Drop tables in reverse dependency order
DROP TABLE IF EXISTS sync_states;
DROP TABLE IF EXISTS order_line_items;
DROP TABLE IF EXISTS orders;
DROP TABLE IF EXISTS inventory_levels;
DROP TABLE IF EXISTS inventory_items;
DROP TABLE IF EXISTS product_variants;
DROP TABLE IF EXISTS products;
DROP TABLE IF EXISTS locations;
DROP TABLE IF EXISTS platform_integrations;

-- Drop enum types
DROP TYPE IF EXISTS entity_type;
DROP TYPE IF EXISTS sync_status;
DROP TYPE IF EXISTS fulfillment_status;
DROP TYPE IF EXISTS financial_status;
DROP TYPE IF EXISTS product_status;
DROP TYPE IF EXISTS platform_type;

-- +goose StatementEnd
