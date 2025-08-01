-- name: GetShopifyUserByUserAndStore :one
SELECT id, user_id, shopify_store_id, access_token, scope, expires_at, created_at, updated_at
FROM shopify_users
WHERE user_id = $1 AND shopify_store_id = $2;

-- name: GetShopifyUserByUserAndDomain :one
SELECT su.id, su.user_id, su.shopify_store_id, su.access_token, su.scope, su.expires_at, su.created_at, su.updated_at
FROM shopify_users su
JOIN shopify_store ss ON su.shopify_store_id = ss.id
WHERE su.user_id = $1 AND ss.shop_domain = $2;

-- name: CreateShopifyUser :one
INSERT INTO shopify_users (id, user_id, shopify_store_id, access_token, scope, expires_at, created_at, updated_at)
VALUES ($1, $2, $3, $4, $5, $6, NOW(), NOW())
RETURNING id, user_id, shopify_store_id, access_token, scope, expires_at, created_at, updated_at;

-- name: UpdateShopifyUserToken :one
UPDATE shopify_users
SET access_token = $3, scope = $4, expires_at = $5, updated_at = NOW()
WHERE user_id = $1 AND shopify_store_id = $2
RETURNING id, user_id, shopify_store_id, access_token, scope, expires_at, created_at, updated_at;

-- name: CreateOrUpdateShopifyUser :one
INSERT INTO shopify_users (id, user_id, shopify_store_id, access_token, scope, expires_at, created_at, updated_at)
VALUES ($1, $2, $3, $4, $5, $6, NOW(), NOW())
ON CONFLICT (user_id, shopify_store_id)
DO UPDATE SET
    access_token = EXCLUDED.access_token,
    scope = EXCLUDED.scope,
    expires_at = EXCLUDED.expires_at,
    updated_at = NOW()
RETURNING id, user_id, shopify_store_id, access_token, scope, expires_at, created_at, updated_at;

-- name: GetShopifyUsersByStore :many
SELECT id, user_id, shopify_store_id, access_token, scope, expires_at, created_at, updated_at
FROM shopify_users
WHERE shopify_store_id = $1;

-- name: GetShopifyUsersByUser :many
SELECT id, user_id, shopify_store_id, access_token, scope, expires_at, created_at, updated_at
FROM shopify_users
WHERE user_id = $1;
