-- name: CreateList :one
INSERT INTO lists (id, board_id, name, position, created_at, updated_at)
VALUES ($1, $2, $3, $4, NOW(), NOW())
RETURNING id, board_id, name, position, created_at, updated_at;

-- name: GetListByID :one
SELECT id, board_id, name, position, created_at, updated_at
FROM lists
WHERE id = $1;

-- name: ListListsByBoard :many
SELECT id, board_id, name, position, created_at, updated_at
FROM lists
WHERE board_id = $1
ORDER BY position ASC;

-- name: UpdateList :one
UPDATE lists
SET name = $2, position = $3, updated_at = NOW()
WHERE id = $1
RETURNING id, board_id, name, position, created_at, updated_at;

-- name: DeleteList :exec
DELETE FROM lists WHERE id = $1;

-- name: GetMaxPositionForBoard :one
SELECT COALESCE(MAX(position), 0)::float8 FROM lists WHERE board_id = $1;
