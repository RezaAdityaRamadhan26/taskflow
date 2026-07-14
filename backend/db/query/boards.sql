-- name: CreateBoard :one
INSERT INTO boards (id, workspace_id, name, description, color, created_at, updated_at)
VALUES ($1, $2, $3, $4, $5, NOW(), NOW())
RETURNING id, workspace_id, name, description, color, created_at, updated_at;

-- name: GetBoardByID :one
SELECT id, workspace_id, name, description, color, created_at, updated_at
FROM boards
WHERE id = $1;

-- name: ListBoardsByWorkspace :many
SELECT id, workspace_id, name, description, color, created_at, updated_at
FROM boards
WHERE workspace_id = $1
ORDER BY created_at DESC;

-- name: UpdateBoard :one
UPDATE boards
SET name = $2, description = $3, color = $4, updated_at = NOW()
WHERE id = $1
RETURNING id, workspace_id, name, description, color, created_at, updated_at;

-- name: DeleteBoard :exec
DELETE FROM boards WHERE id = $1;
