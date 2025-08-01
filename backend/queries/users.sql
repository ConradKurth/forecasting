-- name: GetUserByShopDomain :one
SELECT id, shop_domain, access_token, created_at, updated_at
FROM users
WHERE shop_domain = $1;

-- name: CreateUser :one
INSERT INTO users (id, shop_domain, access_token, created_at, updated_at)
VALUES ($1, $2, $3, NOW(), NOW())
RETURNING id, shop_domain, access_token, created_at, updated_at;

-- name: UpdateUserAccessToken :one
UPDATE users
SET access_token = $2, updated_at = NOW()
WHERE shop_domain = $1
RETURNING id, shop_domain, access_token, created_at, updated_at;

-- name: CreateOrUpdateUser :one
INSERT INTO users (id, shop_domain, access_token, created_at, updated_at)
VALUES ($1, $2, $3, NOW(), NOW())
ON CONFLICT (shop_domain)
DO UPDATE SET
    access_token = EXCLUDED.access_token,
    updated_at = NOW()
RETURNING id, shop_domain, access_token, created_at, updated_at;
