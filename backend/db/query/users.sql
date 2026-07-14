-- name: CreateUser :one
INSERT INTO users (id, email, password_hash, name, avatar_url, created_at, updated_at)
VALUES ($1, $2, $3, $4, $5, NOW(), NOW())
RETURNING id, email, password_hash, name, avatar_url, created_at, updated_at;

-- name: GetUserByEmail :one
SELECT id, email, password_hash, name, avatar_url, created_at, updated_at
FROM users
WHERE email = $1;

-- name: GetUserByID :one
SELECT id, email, password_hash, name, avatar_url, created_at, updated_at
FROM users
WHERE id = $1;

-- name: UpdateUser :one
UPDATE users
SET name = $2, avatar_url = $3, updated_at = NOW()
WHERE id = $1
RETURNING id, email, password_hash, name, avatar_url, created_at, updated_at;

-- name: DeleteUser :exec
DELETE FROM users WHERE id = $1;

-- name: CreateRefreshToken :one
INSERT INTO refresh_tokens (id, token, user_id, expires_at, created_at)
VALUES ($1, $2, $3, $4, NOW())
RETURNING id, token, user_id, expires_at, created_at;

-- name: GetRefreshToken :one
SELECT id, token, user_id, expires_at, created_at
FROM refresh_tokens
WHERE token = $1;

-- name: DeleteRefreshToken :exec
DELETE FROM refresh_tokens WHERE token = $1;

-- name: DeleteRefreshTokensByUserID :exec
DELETE FROM refresh_tokens WHERE user_id = $1;

-- name: DeleteExpiredRefreshTokens :exec
DELETE FROM refresh_tokens WHERE expires_at < NOW();
