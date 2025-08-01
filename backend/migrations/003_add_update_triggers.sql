-- +goose Up
-- +goose StatementBegin

-- Create a function that updates the updated_at column
CREATE OR REPLACE FUNCTION update_updated_at_column()
RETURNS TRIGGER AS $$
BEGIN
    NEW.updated_at = NOW();
    RETURN NEW;
END;
$$ language 'plpgsql';

-- Add triggers to all existing tables with updated_at columns

-- Trigger for users table
CREATE TRIGGER update_users_updated_at
    BEFORE UPDATE ON users
    FOR EACH ROW
    EXECUTE FUNCTION update_updated_at_column();

-- Trigger for shopify_store table
CREATE TRIGGER update_shopify_store_updated_at
    BEFORE UPDATE ON shopify_store
    FOR EACH ROW
    EXECUTE FUNCTION update_updated_at_column();

-- Trigger for shopify_users table
CREATE TRIGGER update_shopify_users_updated_at
    BEFORE UPDATE ON shopify_users
    FOR EACH ROW
    EXECUTE FUNCTION update_updated_at_column();

-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin

-- Drop all triggers
DROP TRIGGER IF EXISTS update_users_updated_at ON users;
DROP TRIGGER IF EXISTS update_shopify_store_updated_at ON shopify_store;
DROP TRIGGER IF EXISTS update_shopify_users_updated_at ON shopify_users;

-- Drop the trigger function
DROP FUNCTION IF EXISTS update_updated_at_column();

-- +goose StatementEnd 