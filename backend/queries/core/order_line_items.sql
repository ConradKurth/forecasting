-- name: GetOrderLineItemByID :one
SELECT id, order_id, external_id, product_id, variant_id, inventory_item_id, quantity, price
FROM order_line_items
WHERE id = $1;

-- name: GetOrderLineItemsByOrderID :many
SELECT id, order_id, external_id, product_id, variant_id, inventory_item_id, quantity, price
FROM order_line_items
WHERE order_id = $1
ORDER BY id;

-- name: CreateOrderLineItem :one
INSERT INTO order_line_items (id, order_id, external_id, product_id, variant_id, inventory_item_id, quantity, price)
VALUES ($1, $2, $3, $4, $5, $6, $7, $8)
RETURNING id, order_id, external_id, product_id, variant_id, inventory_item_id, quantity, price;

-- name: UpsertOrderLineItem :one
INSERT INTO order_line_items (id, order_id, external_id, product_id, variant_id, inventory_item_id, quantity, price)
VALUES ($1, $2, $3, $4, $5, $6, $7, $8)
ON CONFLICT (order_id, external_id)
DO UPDATE SET
    product_id = EXCLUDED.product_id,
    variant_id = EXCLUDED.variant_id,
    inventory_item_id = EXCLUDED.inventory_item_id,
    quantity = EXCLUDED.quantity,
    price = EXCLUDED.price
RETURNING id, order_id, external_id, product_id, variant_id, inventory_item_id, quantity, price;
