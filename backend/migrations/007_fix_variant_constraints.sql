-- +goose Up
-- +goose StatementBegin

-- Fix product_variants constraint - external_id should be globally unique
ALTER TABLE product_variants DROP CONSTRAINT product_variants_product_id_external_id_key;
ALTER TABLE product_variants ADD CONSTRAINT product_variants_external_id_key UNIQUE (external_id);

-- Fix inventory_items constraint if it exists
ALTER TABLE inventory_items DROP CONSTRAINT IF EXISTS inventory_items_integration_id_external_id_key;
ALTER TABLE inventory_items ADD CONSTRAINT inventory_items_external_id_key UNIQUE (external_id);

-- Fix orders constraint if it exists  
ALTER TABLE orders DROP CONSTRAINT IF EXISTS orders_integration_id_external_id_key;
ALTER TABLE orders ADD CONSTRAINT orders_external_id_key UNIQUE (external_id);

-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin

-- Restore original constraints
ALTER TABLE product_variants DROP CONSTRAINT IF EXISTS product_variants_external_id_key;
ALTER TABLE product_variants ADD CONSTRAINT product_variants_product_id_external_id_key UNIQUE (product_id, external_id);

ALTER TABLE inventory_items DROP CONSTRAINT IF EXISTS inventory_items_external_id_key;
ALTER TABLE inventory_items ADD CONSTRAINT inventory_items_integration_id_external_id_key UNIQUE (integration_id, external_id);

ALTER TABLE orders DROP CONSTRAINT IF EXISTS orders_external_id_key;
ALTER TABLE orders ADD CONSTRAINT orders_integration_id_external_id_key UNIQUE (integration_id, external_id);

-- +goose StatementEnd
