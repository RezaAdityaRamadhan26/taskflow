// Package dto defines request and response data transfer objects for the workspace module.
package dto

// CreateWorkspaceRequest represents the request body for creating a workspace.
type CreateWorkspaceRequest struct {
	Name        string  `json:"name" validate:"required,min=2,max=100"`
	Description *string `json:"description,omitempty" validate:"omitempty,max=500"`
	Slug        string  `json:"slug" validate:"required,min=2,max=100,alphanum_with_dash"`
}

// UpdateWorkspaceRequest represents the request body for updating a workspace.
type UpdateWorkspaceRequest struct {
	Name        string  `json:"name" validate:"required,min=2,max=100"`
	Description *string `json:"description,omitempty" validate:"omitempty,max=500"`
	Slug        string  `json:"slug" validate:"required,min=2,max=100,alphanum_with_dash"`
}

// InviteMemberRequest represents the request body for inviting a user to a workspace.
type InviteMemberRequest struct {
	Email string `json:"email" validate:"required,email"`
	Role  string `json:"role" validate:"required,oneof=ADMIN MEMBER"`
}

// UpdateMemberRoleRequest represents the request body for changing a member's role.
type UpdateMemberRoleRequest struct {
	Role string `json:"role" validate:"required,oneof=OWNER ADMIN MEMBER"`
}

// WorkspaceDTO is the response representation of a workspace.
type WorkspaceDTO struct {
	ID          string  `json:"id"`
	Name        string  `json:"name"`
	Description *string `json:"description,omitempty"`
	Slug        string  `json:"slug"`
	Role        string  `json:"role,omitempty"`
	CreatedAt   string  `json:"created_at"`
	UpdatedAt   string  `json:"updated_at"`
}

// WorkspaceMemberDTO is the response representation of a workspace member.
type WorkspaceMemberDTO struct {
	ID        string  `json:"id"`
	UserID    string  `json:"user_id"`
	Name      string  `json:"name"`
	Email     string  `json:"email"`
	AvatarURL *string `json:"avatar_url,omitempty"`
	Role      string  `json:"role"`
	JoinedAt  string  `json:"joined_at"`
}
