-- name: GetUserByID :one
SELECT id, created_at, updated_at
FROM users
WHERE id = $1;

-- name: CreateUser :one
INSERT INTO users (id, created_at, updated_at)
VALUES ($1, NOW(), NOW())
RETURNING id, created_at, updated_at;

-- name: UpdateUser :one
UPDATE users
SET updated_at = NOW()
WHERE id = $1
RETURNING id, created_at, updated_at;
