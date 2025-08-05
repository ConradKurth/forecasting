-- name: GetInventoryItemByID :one
SELECT id, integration_id, external_id, sku, tracked, cost, created_at, updated_at
FROM inventory_items
WHERE id = $1;

-- name: GetInventoryItemsByIntegrationID :many
SELECT id, integration_id, external_id, sku, tracked, cost, created_at, updated_at
FROM inventory_items
WHERE integration_id = $1
ORDER BY created_at DESC
LIMIT $2 OFFSET $3;

-- name: GetInventoryItemByExternalID :one
SELECT id, integration_id, external_id, sku, tracked, cost, created_at, updated_at
FROM inventory_items
WHERE integration_id = $1 AND external_id = $2;

-- name: CreateInventoryItem :one
INSERT INTO inventory_items (id, integration_id, external_id, sku, tracked, cost, created_at, updated_at)
VALUES ($1, $2, $3, $4, $5, $6, NOW(), NOW())
RETURNING id, integration_id, external_id, sku, tracked, cost, created_at, updated_at;

-- name: UpsertInventoryItem :one
INSERT INTO inventory_items (id, integration_id, external_id, sku, tracked, cost, created_at, updated_at)
VALUES ($1, $2, $3, $4, $5, $6, NOW(), NOW())
ON CONFLICT (integration_id, external_id)
DO UPDATE SET
    sku = EXCLUDED.sku,
    tracked = EXCLUDED.tracked,
    cost = EXCLUDED.cost,
    updated_at = NOW()
RETURNING id, integration_id, external_id, sku, tracked, cost, created_at, updated_at;
