-- name: GetLocationByID :one
SELECT id, integration_id, external_id, name, address, country, province, is_active, created_at, updated_at
FROM locations
WHERE id = $1;

-- name: GetLocationsByIntegrationID :many
SELECT id, integration_id, external_id, name, address, country, province, is_active, created_at, updated_at
FROM locations
WHERE integration_id = $1 AND is_active = true
ORDER BY created_at DESC
LIMIT $2 OFFSET $3;

-- name: GetLocationByExternalID :one
SELECT id, integration_id, external_id, name, address, country, province, is_active, created_at, updated_at
FROM locations
WHERE integration_id = $1 AND external_id = $2;

-- name: CreateLocation :one
INSERT INTO locations (id, integration_id, external_id, name, address, country, province, is_active, created_at, updated_at)
VALUES ($1, $2, $3, $4, $5, $6, $7, $8, NOW(), NOW())
RETURNING id, integration_id, external_id, name, address, country, province, is_active, created_at, updated_at;

-- name: UpsertLocation :one
INSERT INTO locations (id, integration_id, external_id, name, address, country, province, is_active, created_at, updated_at)
VALUES ($1, $2, $3, $4, $5, $6, $7, $8, NOW(), NOW())
ON CONFLICT (integration_id, external_id)
DO UPDATE SET
    name = EXCLUDED.name,
    address = EXCLUDED.address,
    country = EXCLUDED.country,
    province = EXCLUDED.province,
    is_active = EXCLUDED.is_active,
    updated_at = NOW()
RETURNING id, integration_id, external_id, name, address, country, province, is_active, created_at, updated_at;

-- name: InsertLocationsBatch :copyfrom
INSERT INTO locations (id, integration_id, external_id, name, address, country, province, is_active, created_at, updated_at) VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10);
