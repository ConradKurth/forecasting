-- name: GetSyncState :one
SELECT id, integration_id, entity_type, last_synced_at, sync_status, error_message
FROM sync_states
WHERE integration_id = $1 AND entity_type = $2;

-- name: GetSyncStatesByIntegrationID :many
SELECT id, integration_id, entity_type, last_synced_at, sync_status, error_message
FROM sync_states
WHERE integration_id = $1
ORDER BY last_synced_at DESC;

-- name: CreateSyncState :one
INSERT INTO sync_states (id, integration_id, entity_type, last_synced_at, sync_status, error_message)
VALUES ($1, $2, $3, $4, $5, $6)
RETURNING id, integration_id, entity_type, last_synced_at, sync_status, error_message;

-- name: UpdateSyncState :one
UPDATE sync_states
SET last_synced_at = $3, sync_status = $4, error_message = $5
WHERE integration_id = $1 AND entity_type = $2
RETURNING id, integration_id, entity_type, last_synced_at, sync_status, error_message;

-- name: UpsertSyncState :one
INSERT INTO sync_states (id, integration_id, entity_type, last_synced_at, sync_status, error_message)
VALUES ($1, $2, $3, $4, $5, $6)
ON CONFLICT (integration_id, entity_type)
DO UPDATE SET
    last_synced_at = EXCLUDED.last_synced_at,
    sync_status = EXCLUDED.sync_status,
    error_message = EXCLUDED.error_message
RETURNING id, integration_id, entity_type, last_synced_at, sync_status, error_message;

-- name: DeleteSyncState :exec
DELETE FROM sync_states WHERE integration_id = $1 AND entity_type = $2;
