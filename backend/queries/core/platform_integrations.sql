-- name: GetPlatformIntegrationByID :one
SELECT id, shop_id, platform_type, platform_shop_id, is_active, created_at, updated_at
FROM platform_integrations
WHERE id = $1;

-- name: GetPlatformIntegrationsByShopID :many
SELECT id, shop_id, platform_type, platform_shop_id, is_active, created_at, updated_at
FROM platform_integrations
WHERE shop_id = $1 AND is_active = true
ORDER BY created_at DESC;

-- name: GetPlatformIntegrationByShopAndType :one
SELECT id, shop_id, platform_type, platform_shop_id, is_active, created_at, updated_at
FROM platform_integrations
WHERE shop_id = $1 AND platform_type = $2 AND is_active = true;

-- name: CreatePlatformIntegration :one
INSERT INTO platform_integrations (id, shop_id, platform_type, platform_shop_id, is_active, created_at, updated_at)
VALUES ($1, $2, $3, $4, $5, NOW(), NOW())
RETURNING id, shop_id, platform_type, platform_shop_id, is_active, created_at, updated_at;

-- name: UpdatePlatformIntegration :one
UPDATE platform_integrations
SET is_active = $3, updated_at = NOW()
WHERE id = $1 AND shop_id = $2
RETURNING id, shop_id, platform_type, platform_shop_id, is_active, created_at, updated_at;

-- name: UpsertPlatformIntegration :one
INSERT INTO platform_integrations (id, shop_id, platform_type, platform_shop_id, is_active, created_at, updated_at)
VALUES ($1, $2, $3, $4, $5, NOW(), NOW())
ON CONFLICT (platform_shop_id, platform_type)
DO UPDATE SET
    is_active = EXCLUDED.is_active,
    updated_at = NOW()
RETURNING id, shop_id, platform_type, platform_shop_id, is_active, created_at, updated_at;

-- name: DeactivatePlatformIntegration :exec
UPDATE platform_integrations
SET is_active = false, updated_at = NOW()
WHERE id = $1;
