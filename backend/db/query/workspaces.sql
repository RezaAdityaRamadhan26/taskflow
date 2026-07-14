-- name: CreateWorkspace :one
INSERT INTO workspaces (id, name, description, slug, created_at, updated_at)
VALUES ($1, $2, $3, $4, NOW(), NOW())
RETURNING id, name, description, slug, created_at, updated_at;

-- name: GetWorkspaceByID :one
SELECT id, name, description, slug, created_at, updated_at
FROM workspaces
WHERE id = $1;

-- name: GetWorkspaceBySlug :one
SELECT id, name, description, slug, created_at, updated_at
FROM workspaces
WHERE slug = $1;

-- name: UpdateWorkspace :one
UPDATE workspaces
SET name = $2, description = $3, slug = $4, updated_at = NOW()
WHERE id = $1
RETURNING id, name, description, slug, created_at, updated_at;

-- name: DeleteWorkspace :exec
DELETE FROM workspaces WHERE id = $1;

-- name: AddWorkspaceMember :one
INSERT INTO workspace_members (id, user_id, workspace_id, role, joined_at)
VALUES ($1, $2, $3, $4, NOW())
RETURNING id, user_id, workspace_id, role, joined_at;

-- name: GetWorkspaceMember :one
SELECT id, user_id, workspace_id, role, joined_at
FROM workspace_members
WHERE user_id = $1 AND workspace_id = $2;

-- name: ListWorkspaceMembers :many
SELECT wm.id, wm.user_id, wm.workspace_id, wm.role, wm.joined_at,
       u.name AS user_name, u.email AS user_email, u.avatar_url AS user_avatar_url
FROM workspace_members wm
JOIN users u ON u.id = wm.user_id
WHERE wm.workspace_id = $1
ORDER BY wm.joined_at ASC;

-- name: UpdateWorkspaceMemberRole :one
UPDATE workspace_members
SET role = $3
WHERE user_id = $1 AND workspace_id = $2
RETURNING id, user_id, workspace_id, role, joined_at;

-- name: RemoveWorkspaceMember :exec
DELETE FROM workspace_members
WHERE user_id = $1 AND workspace_id = $2;

-- name: ListWorkspacesByUserID :many
SELECT w.id, w.name, w.description, w.slug, w.created_at, w.updated_at,
       wm.role
FROM workspaces w
JOIN workspace_members wm ON wm.workspace_id = w.id
WHERE wm.user_id = $1
ORDER BY w.created_at DESC;

-- name: CountWorkspaceMembers :one
SELECT COUNT(*) FROM workspace_members WHERE workspace_id = $1;
