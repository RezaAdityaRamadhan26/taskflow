-- name: CreateAttachment :one
INSERT INTO attachments (id, card_id, user_id, file_name, file_url, file_size, file_type, created_at)
VALUES ($1, $2, $3, $4, $5, $6, $7, NOW())
RETURNING id, card_id, user_id, file_name, file_url, file_size, file_type, created_at;

-- name: GetAttachmentByID :one
SELECT id, card_id, user_id, file_name, file_url, file_size, file_type, created_at
FROM attachments
WHERE id = $1;

-- name: ListAttachmentsByCard :many
SELECT a.id, a.card_id, a.user_id, a.file_name, a.file_url, a.file_size, a.file_type, a.created_at,
       u.name AS user_name, u.avatar_url AS user_avatar_url
FROM attachments a
JOIN users u ON u.id = a.user_id
WHERE a.card_id = $1
ORDER BY a.created_at DESC;

-- name: DeleteAttachment :exec
DELETE FROM attachments WHERE id = $1;
