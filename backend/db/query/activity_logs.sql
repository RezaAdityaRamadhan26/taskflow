-- name: CreateActivityLog :exec
INSERT INTO activity_logs (id, board_id, card_id, user_id, action, entity_type, entity_title, details, created_at)
VALUES ($1, $2, $3, $4, $5, $6, $7, $8, NOW());

-- name: ListActivityLogsByBoard :many
SELECT a.id, a.board_id, a.card_id, a.user_id, a.action, a.entity_type, a.entity_title, a.details, a.created_at,
       u.name AS user_name, u.avatar_url AS user_avatar_url
FROM activity_logs a
JOIN users u ON u.id = a.user_id
WHERE a.board_id = $1
ORDER BY a.created_at DESC
LIMIT $2 OFFSET $3;

-- name: ListActivityLogsByCard :many
SELECT a.id, a.board_id, a.card_id, a.user_id, a.action, a.entity_type, a.entity_title, a.details, a.created_at,
       u.name AS user_name, u.avatar_url AS user_avatar_url
FROM activity_logs a
JOIN users u ON u.id = a.user_id
WHERE a.card_id = $1
ORDER BY a.created_at DESC
LIMIT $2 OFFSET $3;
