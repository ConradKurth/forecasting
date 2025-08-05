-- name: GetInventoryLevelByID :one
SELECT id, inventory_item_id, location_id, available, updated_at
FROM inventory_levels
WHERE id = $1;

-- name: GetInventoryLevelsByInventoryItemID :many
SELECT id, inventory_item_id, location_id, available, updated_at
FROM inventory_levels
WHERE inventory_item_id = $1
ORDER BY location_id;

-- name: GetInventoryLevelsByLocationID :many
SELECT id, inventory_item_id, location_id, available, updated_at
FROM inventory_levels
WHERE location_id = $1
ORDER BY inventory_item_id;

-- name: CreateInventoryLevel :one
INSERT INTO inventory_levels (id, inventory_item_id, location_id, available, updated_at)
VALUES ($1, $2, $3, $4, NOW())
RETURNING id, inventory_item_id, location_id, available, updated_at;

-- name: UpsertInventoryLevel :one
INSERT INTO inventory_levels (id, inventory_item_id, location_id, available, updated_at)
VALUES ($1, $2, $3, $4, NOW())
ON CONFLICT (inventory_item_id, location_id)
DO UPDATE SET
    available = EXCLUDED.available,
    updated_at = NOW()
RETURNING id, inventory_item_id, location_id, available, updated_at;
