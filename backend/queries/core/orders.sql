-- name: GetOrderByID :one
SELECT id, integration_id, external_id, created_at, financial_status, fulfillment_status, total_price, cancelled_at
FROM orders
WHERE id = $1;

-- name: GetOrdersByIntegrationID :many
SELECT id, integration_id, external_id, created_at, financial_status, fulfillment_status, total_price, cancelled_at
FROM orders
WHERE integration_id = $1
ORDER BY created_at DESC
LIMIT $2 OFFSET $3;

-- name: GetOrdersByIntegrationIDSince :many
SELECT id, integration_id, external_id, created_at, financial_status, fulfillment_status, total_price, cancelled_at
FROM orders
WHERE integration_id = $1 AND created_at >= $2
ORDER BY created_at DESC
LIMIT $3 OFFSET $4;

-- name: GetOrderByExternalID :one
SELECT id, integration_id, external_id, created_at, financial_status, fulfillment_status, total_price, cancelled_at
FROM orders
WHERE integration_id = $1 AND external_id = $2;

-- name: CreateOrder :one
INSERT INTO orders (id, integration_id, external_id, created_at, financial_status, fulfillment_status, total_price, cancelled_at)
VALUES ($1, $2, $3, $4, $5, $6, $7, $8)
RETURNING id, integration_id, external_id, created_at, financial_status, fulfillment_status, total_price, cancelled_at;

-- name: UpdateOrder :one
UPDATE orders
SET financial_status = $3, fulfillment_status = $4, total_price = $5, cancelled_at = $6
WHERE id = $1 AND integration_id = $2
RETURNING id, integration_id, external_id, created_at, financial_status, fulfillment_status, total_price, cancelled_at;

-- name: UpsertOrder :one
INSERT INTO orders (id, integration_id, external_id, created_at, financial_status, fulfillment_status, total_price, cancelled_at)
VALUES ($1, $2, $3, $4, $5, $6, $7, $8)
ON CONFLICT (integration_id, external_id)
DO UPDATE SET
    financial_status = EXCLUDED.financial_status,
    fulfillment_status = EXCLUDED.fulfillment_status,
    total_price = EXCLUDED.total_price,
    cancelled_at = EXCLUDED.cancelled_at
RETURNING id, integration_id, external_id, created_at, financial_status, fulfillment_status, total_price, cancelled_at;

-- name: InsertOrdersBatch :copyfrom
INSERT INTO orders (id, integration_id, external_id, created_at, financial_status, fulfillment_status, total_price, cancelled_at) VALUES ($1, $2, $3, $4, $5, $6, $7, $8);

-- name: DeleteOrder :exec
DELETE FROM orders WHERE id = $1 AND integration_id = $2;
