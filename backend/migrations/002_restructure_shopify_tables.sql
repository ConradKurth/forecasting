-- +goose Up
-- +goose StatementBegin

-- Drop existing shopify_store table since we're restructuring
DROP TABLE IF EXISTS shopify_store;

-- Create new shopify_store table for store metadata
CREATE TABLE shopify_store (
    id TEXT PRIMARY KEY,
    shop_domain TEXT UNIQUE NOT NULL,
    shop_name TEXT,
    timezone TEXT,
    currency TEXT,
    created_at TIMESTAMP DEFAULT NOW(),
    updated_at TIMESTAMP DEFAULT NOW()
);

-- Create shopify_users table for auth data
CREATE TABLE shopify_users (
    id TEXT PRIMARY KEY,
    user_id TEXT NOT NULL REFERENCES users(id),
    shopify_store_id TEXT NOT NULL REFERENCES shopify_store(id),
    access_token TEXT NOT NULL,
    scope TEXT NOT NULL,
    expires_at TIMESTAMP,
    created_at TIMESTAMP DEFAULT NOW(),
    updated_at TIMESTAMP DEFAULT NOW(),
    UNIQUE(user_id, shopify_store_id)
);

-- Remove access_token and shop_domain from users table since they'll be in shopify tables
ALTER TABLE users DROP COLUMN IF EXISTS access_token;
ALTER TABLE users DROP COLUMN IF EXISTS shop_domain;

-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin

-- Add back columns to users table
ALTER TABLE users ADD COLUMN shop_domain TEXT;
ALTER TABLE users ADD COLUMN access_token TEXT;

-- Drop new tables
DROP TABLE IF EXISTS shopify_users;
DROP TABLE IF EXISTS shopify_store;

-- Recreate original shopify_store table
CREATE TABLE shopify_store (
    id TEXT PRIMARY KEY,
    user_id TEXT NOT NULL REFERENCES users(id),
    shop_domain TEXT UNIQUE NOT NULL,
    access_token TEXT NOT NULL,
    scope TEXT NOT NULL,
    created_at TIMESTAMP DEFAULT NOW(),
    updated_at TIMESTAMP DEFAULT NOW()
);

-- +goose StatementEnd
