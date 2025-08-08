-- +goose Up
-- +goose StatementBegin

-- Remove the problematic unique constraint on (integration_id, external_id)
-- This constraint prevents a single integration from having multiple locations
-- which is incorrect for platforms like Shopify where stores can have multiple locations
ALTER TABLE locations DROP CONSTRAINT locations_integration_id_external_id_key;

-- Add a unique constraint only on external_id since platform location IDs should be globally unique
-- This ensures the same external location isn't duplicated across the system
ALTER TABLE locations ADD CONSTRAINT locations_external_id_key UNIQUE (external_id);

-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin

-- Restore the original constraint
ALTER TABLE locations DROP CONSTRAINT IF EXISTS locations_external_id_key;
ALTER TABLE locations ADD CONSTRAINT locations_integration_id_external_id_key UNIQUE (integration_id, external_id);

-- +goose StatementEnd
