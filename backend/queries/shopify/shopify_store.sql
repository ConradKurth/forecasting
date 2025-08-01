-- name: GetShopifyStoreByDomain :one
SELECT id, shop_domain, shop_name, timezone, currency, created_at, updated_at
FROM shopify_store
WHERE shop_domain = $1;

-- name: GetShopifyStoreByID :one
SELECT id, shop_domain, shop_name, timezone, currency, created_at, updated_at
FROM shopify_store
WHERE id = $1;

-- name: CreateShopifyStore :one
INSERT INTO shopify_store (id, shop_domain, shop_name, timezone, currency, created_at, updated_at)
VALUES ($1, $2, $3, $4, $5, NOW(), NOW())
RETURNING id, shop_domain, shop_name, timezone, currency, created_at, updated_at;

-- name: UpdateShopifyStore :one
UPDATE shopify_store
SET shop_name = $2, timezone = $3, currency = $4, updated_at = NOW()
WHERE shop_domain = $1
RETURNING id, shop_domain, shop_name, timezone, currency, created_at, updated_at;

-- name: CreateOrUpdateShopifyStore :one
INSERT INTO shopify_store (id, shop_domain, shop_name, timezone, currency, created_at, updated_at)
VALUES ($1, $2, $3, $4, $5, NOW(), NOW())
ON CONFLICT (shop_domain)
DO UPDATE SET
    shop_name = EXCLUDED.shop_name,
    timezone = EXCLUDED.timezone,
    currency = EXCLUDED.currency,
    updated_at = NOW()
RETURNING id, shop_domain, shop_name, timezone, currency, created_at, updated_at;
