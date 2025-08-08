-- name: GetProductVariantByID :one
SELECT id, product_id, external_id, sku, price, inventory_item_id, created_at, updated_at
FROM product_variants
WHERE id = $1;

-- name: GetProductVariantsByProductID :many
SELECT id, product_id, external_id, sku, price, inventory_item_id, created_at, updated_at
FROM product_variants
WHERE product_id = $1
ORDER BY created_at DESC;

-- name: GetProductVariantsByIntegrationID :many
SELECT pv.id, pv.product_id, pv.external_id, pv.sku, pv.price, pv.inventory_item_id, pv.created_at, pv.updated_at
FROM product_variants pv
JOIN products p ON pv.product_id = p.id
WHERE p.integration_id = $1
ORDER BY pv.created_at DESC
LIMIT $2 OFFSET $3;

-- name: GetProductVariantByExternalID :one
SELECT pv.id, pv.product_id, pv.external_id, pv.sku, pv.price, pv.inventory_item_id, pv.created_at, pv.updated_at
FROM product_variants pv
JOIN products p ON pv.product_id = p.id
WHERE p.integration_id = $1 AND pv.external_id = $2;

-- name: CreateProductVariant :one
INSERT INTO product_variants (id, product_id, external_id, sku, price, inventory_item_id, created_at, updated_at)
VALUES ($1, $2, $3, $4, $5, $6, NOW(), NOW())
RETURNING id, product_id, external_id, sku, price, inventory_item_id, created_at, updated_at;

-- name: UpsertProductVariant :one
INSERT INTO product_variants (id, product_id, external_id, sku, price, inventory_item_id, created_at, updated_at)
VALUES ($1, $2, $3, $4, $5, $6, NOW(), NOW())
ON CONFLICT (external_id)
DO UPDATE SET
    sku = EXCLUDED.sku,
    price = EXCLUDED.price,
    inventory_item_id = EXCLUDED.inventory_item_id,
    updated_at = NOW()
RETURNING id, product_id, external_id, sku, price, inventory_item_id, created_at, updated_at;

-- name: InsertProductVariantsBatch :batchexec
INSERT INTO product_variants (id, product_id, external_id, sku, price, inventory_item_id, created_at, updated_at) 
VALUES ($1, $2, $3, $4, $5, $6, $7, $8)
ON CONFLICT (external_id)
DO UPDATE SET
    sku = EXCLUDED.sku,
    price = EXCLUDED.price,
    inventory_item_id = EXCLUDED.inventory_item_id,
    updated_at = EXCLUDED.updated_at;
