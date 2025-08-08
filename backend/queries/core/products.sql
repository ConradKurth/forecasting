-- name: GetProductByID :one
SELECT id, integration_id, external_id, title, handle, product_type, status, created_at, updated_at
FROM products
WHERE id = $1;

-- name: GetProductsByIntegrationID :many
SELECT id, integration_id, external_id, title, handle, product_type, status, created_at, updated_at
FROM products
WHERE integration_id = $1
ORDER BY created_at DESC
LIMIT $2 OFFSET $3;

-- name: GetProductByHandle :one
SELECT id, integration_id, external_id, title, handle, product_type, status, created_at, updated_at
FROM products
WHERE integration_id = $1 AND handle = $2;

-- name: GetProductByExternalID :one
SELECT id, integration_id, external_id, title, handle, product_type, status, created_at, updated_at
FROM products
WHERE integration_id = $1 AND external_id = $2;

-- name: CreateProduct :one
INSERT INTO products (id, integration_id, external_id, title, handle, product_type, status, created_at, updated_at)
VALUES ($1, $2, $3, $4, $5, $6, $7, NOW(), NOW())
RETURNING id, integration_id, external_id, title, handle, product_type, status, created_at, updated_at;

-- name: UpdateProduct :one
UPDATE products
SET title = $3, handle = $4, product_type = $5, status = $6, updated_at = NOW()
WHERE id = $1 AND integration_id = $2
RETURNING id, integration_id, external_id, title, handle, product_type, status, created_at, updated_at;

-- name: UpsertProduct :one
INSERT INTO products (id, integration_id, external_id, title, handle, product_type, status, created_at, updated_at)
VALUES ($1, $2, $3, $4, $5, $6, $7, NOW(), NOW())
ON CONFLICT (integration_id, handle)
DO UPDATE SET
    external_id = EXCLUDED.external_id,
    title = EXCLUDED.title,
    product_type = EXCLUDED.product_type,
    status = EXCLUDED.status,
    updated_at = NOW()
RETURNING id, integration_id, external_id, title, handle, product_type, status, created_at, updated_at;

-- name: InsertProductsBatch :batchexec
INSERT INTO products (id, integration_id, external_id, title, handle, product_type, status, created_at, updated_at) 
VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9)
ON CONFLICT (integration_id, handle)
DO UPDATE SET
    external_id = EXCLUDED.external_id,
    title = EXCLUDED.title,
    product_type = EXCLUDED.product_type,
    status = EXCLUDED.status,
    updated_at = EXCLUDED.updated_at;

-- name: DeleteProduct :exec
DELETE FROM products WHERE id = $1 AND integration_id = $2;
