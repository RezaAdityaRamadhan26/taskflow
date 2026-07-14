-- name: CreateComment :one
INSERT INTO comments (id, card_id, user_id, content, created_at, updated_at)
VALUES ($1, $2, $3, $4, NOW(), NOW())
RETURNING id, card_id, user_id, content, created_at, updated_at;

-- name: GetCommentByID :one
SELECT id, card_id, user_id, content, created_at, updated_at
FROM comments
WHERE id = $1;

-- name: ListCommentsByCard :many
SELECT c.id, c.card_id, c.user_id, c.content, c.created_at, c.updated_at,
       u.name AS user_name, u.avatar_url AS user_avatar_url
FROM comments c
JOIN users u ON u.id = c.user_id
WHERE c.card_id = $1
ORDER BY c.created_at DESC;

-- name: UpdateComment :one
UPDATE comments
SET content = $2, updated_at = NOW()
WHERE id = $1
RETURNING id, card_id, user_id, content, created_at, updated_at;

-- name: DeleteComment :exec
DELETE FROM comments WHERE id = $1;
