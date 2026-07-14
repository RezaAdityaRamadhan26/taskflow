-- name: CreateCard :one
INSERT INTO cards (id, list_id, title, description, position, priority, due_date, created_at, updated_at)
VALUES ($1, $2, $3, $4, $5, $6, $7, NOW(), NOW())
RETURNING id, list_id, title, description, position, priority, due_date, created_at, updated_at;

-- name: GetCardByID :one
SELECT id, list_id, title, description, position, priority, due_date, created_at, updated_at
FROM cards
WHERE id = $1;

-- name: ListCardsByList :many
SELECT id, list_id, title, description, position, priority, due_date, created_at, updated_at
FROM cards
WHERE list_id = $1
ORDER BY position ASC;

-- name: UpdateCard :one
UPDATE cards
SET list_id = $2, title = $3, description = $4, position = $5, priority = $6, due_date = $7, updated_at = NOW()
WHERE id = $1
RETURNING id, list_id, title, description, position, priority, due_date, created_at, updated_at;

-- name: DeleteCard :exec
DELETE FROM cards WHERE id = $1;

-- name: GetMaxPositionForList :one
SELECT COALESCE(MAX(position), 0)::float8 FROM cards WHERE list_id = $1;
