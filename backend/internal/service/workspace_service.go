// Package service implements the business logic for workspace management.
package service

import (
	"context"
	"database/sql"
	"fmt"
	"log"
	"regexp"
	"strings"
	"time"

	"github.com/google/uuid"

	"github.com/RezaAdityaRamadhan26/taskflow/backend/internal/dto"
	"github.com/RezaAdityaRamadhan26/taskflow/backend/internal/repository"
)

// WorkspaceService handles workspace business logic.
type WorkspaceService struct {
	queries *repository.Queries
}

// NewWorkspaceService creates a new WorkspaceService instance.
func NewWorkspaceService(queries *repository.Queries) *WorkspaceService {
	return &WorkspaceService{queries: queries}
}

// slugRegex validates that a slug contains only lowercase letters, numbers, and hyphens.
var slugRegex = regexp.MustCompile(`^[a-z0-9]+(?:-[a-z0-9]+)*$`)

// normalizeSlug converts a string to a valid slug format.
func normalizeSlug(s string) string {
	s = strings.ToLower(strings.TrimSpace(s))
	s = regexp.MustCompile(`[^a-z0-9-]+`).ReplaceAllString(s, "-")
	s = regexp.MustCompile(`-{2,}`).ReplaceAllString(s, "-")
	s = strings.Trim(s, "-")
	return s
}

// CreateWorkspace creates a new workspace and adds the creator as OWNER.
func (s *WorkspaceService) CreateWorkspace(ctx context.Context, userID uuid.UUID, req dto.CreateWorkspaceRequest) (*dto.WorkspaceDTO, error) {
	// Normalize slug
	slug := normalizeSlug(req.Slug)
	if slug == "" {
		return nil, fmt.Errorf("invalid slug format")
	}

	// Check if slug is already taken
	_, err := s.queries.GetWorkspaceBySlug(ctx, slug)
	if err == nil {
		return nil, fmt.Errorf("slug already taken")
	}

	// Create workspace
	var description sql.NullString
	if req.Description != nil && *req.Description != "" {
		description = sql.NullString{String: *req.Description, Valid: true}
	}

	workspace, err := s.queries.CreateWorkspace(ctx, repository.CreateWorkspaceParams{
		ID:          uuid.New(),
		Name:        req.Name,
		Description: description,
		Slug:        slug,
	})
	if err != nil {
		log.Printf("[ERROR] Failed to create workspace: %v", err)
		return nil, fmt.Errorf("failed to create workspace")
	}

	// Add creator as OWNER
	_, err = s.queries.AddWorkspaceMember(ctx, repository.AddWorkspaceMemberParams{
		ID:          uuid.New(),
		UserID:      userID,
		WorkspaceID: workspace.ID,
		Role:        repository.RoleOWNER,
	})
	if err != nil {
		log.Printf("[ERROR] Failed to add owner to workspace: %v", err)
		// Rollback: delete the workspace
		_ = s.queries.DeleteWorkspace(ctx, workspace.ID)
		return nil, fmt.Errorf("failed to setup workspace")
	}

	var desc *string
	if workspace.Description.Valid {
		desc = &workspace.Description.String
	}

	return &dto.WorkspaceDTO{
		ID:          workspace.ID.String(),
		Name:        workspace.Name,
		Description: desc,
		Slug:        workspace.Slug,
		Role:        string(repository.RoleOWNER),
		CreatedAt:   workspace.CreatedAt.Format(time.RFC3339),
		UpdatedAt:   workspace.UpdatedAt.Format(time.RFC3339),
	}, nil
}

// ListWorkspaces returns all workspaces for a given user.
func (s *WorkspaceService) ListWorkspaces(ctx context.Context, userID uuid.UUID) ([]dto.WorkspaceDTO, error) {
	rows, err := s.queries.ListWorkspacesByUserID(ctx, userID)
	if err != nil {
		log.Printf("[ERROR] Failed to list workspaces: %v", err)
		return nil, fmt.Errorf("failed to list workspaces")
	}

	workspaces := make([]dto.WorkspaceDTO, 0, len(rows))
	for _, row := range rows {
		var desc *string
		if row.Description.Valid {
			desc = &row.Description.String
		}
		workspaces = append(workspaces, dto.WorkspaceDTO{
			ID:          row.ID.String(),
			Name:        row.Name,
			Description: desc,
			Slug:        row.Slug,
			Role:        string(row.Role),
			CreatedAt:   row.CreatedAt.Format(time.RFC3339),
			UpdatedAt:   row.UpdatedAt.Format(time.RFC3339),
		})
	}

	return workspaces, nil
}

// GetWorkspace returns a workspace by ID, verifying user membership.
func (s *WorkspaceService) GetWorkspace(ctx context.Context, userID, workspaceID uuid.UUID) (*dto.WorkspaceDTO, error) {
	// Verify membership
	member, err := s.queries.GetWorkspaceMember(ctx, repository.GetWorkspaceMemberParams{
		UserID:      userID,
		WorkspaceID: workspaceID,
	})
	if err != nil {
		return nil, fmt.Errorf("you are not a member of this workspace")
	}

	workspace, err := s.queries.GetWorkspaceByID(ctx, workspaceID)
	if err != nil {
		return nil, fmt.Errorf("workspace not found")
	}

	var desc *string
	if workspace.Description.Valid {
		desc = &workspace.Description.String
	}

	return &dto.WorkspaceDTO{
		ID:          workspace.ID.String(),
		Name:        workspace.Name,
		Description: desc,
		Slug:        workspace.Slug,
		Role:        string(member.Role),
		CreatedAt:   workspace.CreatedAt.Format(time.RFC3339),
		UpdatedAt:   workspace.UpdatedAt.Format(time.RFC3339),
	}, nil
}

// UpdateWorkspace updates a workspace. Only OWNER and ADMIN can update.
func (s *WorkspaceService) UpdateWorkspace(ctx context.Context, userID, workspaceID uuid.UUID, req dto.UpdateWorkspaceRequest) (*dto.WorkspaceDTO, error) {
	// Verify membership and role
	member, err := s.queries.GetWorkspaceMember(ctx, repository.GetWorkspaceMemberParams{
		UserID:      userID,
		WorkspaceID: workspaceID,
	})
	if err != nil {
		return nil, fmt.Errorf("you are not a member of this workspace")
	}
	if member.Role != repository.RoleOWNER && member.Role != repository.RoleADMIN {
		return nil, fmt.Errorf("only OWNER or ADMIN can update workspace")
	}

	// Normalize slug
	slug := normalizeSlug(req.Slug)
	if slug == "" {
		return nil, fmt.Errorf("invalid slug format")
	}

	// Check slug uniqueness (excluding current workspace)
	existingWS, err := s.queries.GetWorkspaceBySlug(ctx, slug)
	if err == nil && existingWS.ID != workspaceID {
		return nil, fmt.Errorf("slug already taken")
	}

	var description sql.NullString
	if req.Description != nil && *req.Description != "" {
		description = sql.NullString{String: *req.Description, Valid: true}
	}

	workspace, err := s.queries.UpdateWorkspace(ctx, repository.UpdateWorkspaceParams{
		ID:          workspaceID,
		Name:        req.Name,
		Description: description,
		Slug:        slug,
	})
	if err != nil {
		log.Printf("[ERROR] Failed to update workspace: %v", err)
		return nil, fmt.Errorf("failed to update workspace")
	}

	var desc *string
	if workspace.Description.Valid {
		desc = &workspace.Description.String
	}

	return &dto.WorkspaceDTO{
		ID:          workspace.ID.String(),
		Name:        workspace.Name,
		Description: desc,
		Slug:        workspace.Slug,
		Role:        string(member.Role),
		CreatedAt:   workspace.CreatedAt.Format(time.RFC3339),
		UpdatedAt:   workspace.UpdatedAt.Format(time.RFC3339),
	}, nil
}

// DeleteWorkspace deletes a workspace. Only OWNER can delete.
func (s *WorkspaceService) DeleteWorkspace(ctx context.Context, userID, workspaceID uuid.UUID) error {
	// Verify OWNER role
	member, err := s.queries.GetWorkspaceMember(ctx, repository.GetWorkspaceMemberParams{
		UserID:      userID,
		WorkspaceID: workspaceID,
	})
	if err != nil {
		return fmt.Errorf("you are not a member of this workspace")
	}
	if member.Role != repository.RoleOWNER {
		return fmt.Errorf("only OWNER can delete workspace")
	}

	if err := s.queries.DeleteWorkspace(ctx, workspaceID); err != nil {
		log.Printf("[ERROR] Failed to delete workspace: %v", err)
		return fmt.Errorf("failed to delete workspace")
	}

	return nil
}

// InviteMember adds a new member to a workspace by email. Only OWNER/ADMIN can invite.
func (s *WorkspaceService) InviteMember(ctx context.Context, inviterID, workspaceID uuid.UUID, req dto.InviteMemberRequest) (*dto.WorkspaceMemberDTO, error) {
	// Verify inviter is OWNER or ADMIN
	inviter, err := s.queries.GetWorkspaceMember(ctx, repository.GetWorkspaceMemberParams{
		UserID:      inviterID,
		WorkspaceID: workspaceID,
	})
	if err != nil {
		return nil, fmt.Errorf("you are not a member of this workspace")
	}
	if inviter.Role != repository.RoleOWNER && inviter.Role != repository.RoleADMIN {
		return nil, fmt.Errorf("only OWNER or ADMIN can invite members")
	}

	// Prevent ADMIN from inviting as ADMIN (only OWNER can elevate to ADMIN)
	if inviter.Role == repository.RoleADMIN && req.Role == "ADMIN" {
		return nil, fmt.Errorf("only OWNER can invite as ADMIN")
	}

	// Find user by email
	user, err := s.queries.GetUserByEmail(ctx, req.Email)
	if err != nil {
		return nil, fmt.Errorf("user with email %s not found", req.Email)
	}

	// Check if already a member
	_, err = s.queries.GetWorkspaceMember(ctx, repository.GetWorkspaceMemberParams{
		UserID:      user.ID,
		WorkspaceID: workspaceID,
	})
	if err == nil {
		return nil, fmt.Errorf("user is already a member of this workspace")
	}

	// Add member
	member, err := s.queries.AddWorkspaceMember(ctx, repository.AddWorkspaceMemberParams{
		ID:          uuid.New(),
		UserID:      user.ID,
		WorkspaceID: workspaceID,
		Role:        repository.Role(req.Role),
	})
	if err != nil {
		log.Printf("[ERROR] Failed to add member: %v", err)
		return nil, fmt.Errorf("failed to invite member")
	}

	var avatarURL *string
	if user.AvatarUrl.Valid {
		avatarURL = &user.AvatarUrl.String
	}

	return &dto.WorkspaceMemberDTO{
		ID:        member.ID.String(),
		UserID:    user.ID.String(),
		Name:      user.Name,
		Email:     user.Email,
		AvatarURL: avatarURL,
		Role:      string(member.Role),
		JoinedAt:  member.JoinedAt.Format(time.RFC3339),
	}, nil
}

// ListMembers returns all members of a workspace.
func (s *WorkspaceService) ListMembers(ctx context.Context, userID, workspaceID uuid.UUID) ([]dto.WorkspaceMemberDTO, error) {
	// Verify membership
	_, err := s.queries.GetWorkspaceMember(ctx, repository.GetWorkspaceMemberParams{
		UserID:      userID,
		WorkspaceID: workspaceID,
	})
	if err != nil {
		return nil, fmt.Errorf("you are not a member of this workspace")
	}

	rows, err := s.queries.ListWorkspaceMembers(ctx, workspaceID)
	if err != nil {
		log.Printf("[ERROR] Failed to list members: %v", err)
		return nil, fmt.Errorf("failed to list members")
	}

	members := make([]dto.WorkspaceMemberDTO, 0, len(rows))
	for _, row := range rows {
		var avatarURL *string
		if row.UserAvatarUrl.Valid {
			avatarURL = &row.UserAvatarUrl.String
		}
		members = append(members, dto.WorkspaceMemberDTO{
			ID:        row.ID.String(),
			UserID:    row.UserID.String(),
			Name:      row.UserName,
			Email:     row.UserEmail,
			AvatarURL: avatarURL,
			Role:      string(row.Role),
			JoinedAt:  row.JoinedAt.Format(time.RFC3339),
		})
	}

	return members, nil
}

// UpdateMemberRole updates a member's role. Only OWNER can change roles.
func (s *WorkspaceService) UpdateMemberRole(ctx context.Context, updaterID, workspaceID, targetUserID uuid.UUID, req dto.UpdateMemberRoleRequest) (*dto.WorkspaceMemberDTO, error) {
	// Verify updater is OWNER
	updater, err := s.queries.GetWorkspaceMember(ctx, repository.GetWorkspaceMemberParams{
		UserID:      updaterID,
		WorkspaceID: workspaceID,
	})
	if err != nil {
		return nil, fmt.Errorf("you are not a member of this workspace")
	}
	if updater.Role != repository.RoleOWNER {
		return nil, fmt.Errorf("only OWNER can change member roles")
	}

	// Prevent owner from demoting themselves
	if updaterID == targetUserID {
		return nil, fmt.Errorf("owner cannot change their own role")
	}

	// Verify target is a member
	_, err = s.queries.GetWorkspaceMember(ctx, repository.GetWorkspaceMemberParams{
		UserID:      targetUserID,
		WorkspaceID: workspaceID,
	})
	if err != nil {
		return nil, fmt.Errorf("target user is not a member of this workspace")
	}

	// Update role
	member, err := s.queries.UpdateWorkspaceMemberRole(ctx, repository.UpdateWorkspaceMemberRoleParams{
		UserID:      targetUserID,
		WorkspaceID: workspaceID,
		Role:        repository.Role(req.Role),
	})
	if err != nil {
		log.Printf("[ERROR] Failed to update member role: %v", err)
		return nil, fmt.Errorf("failed to update member role")
	}

	// Get user info
	user, err := s.queries.GetUserByID(ctx, targetUserID)
	if err != nil {
		return nil, fmt.Errorf("user not found")
	}

	var avatarURL *string
	if user.AvatarUrl.Valid {
		avatarURL = &user.AvatarUrl.String
	}

	return &dto.WorkspaceMemberDTO{
		ID:        member.ID.String(),
		UserID:    user.ID.String(),
		Name:      user.Name,
		Email:     user.Email,
		AvatarURL: avatarURL,
		Role:      string(member.Role),
		JoinedAt:  member.JoinedAt.Format(time.RFC3339),
	}, nil
}

// RemoveMember removes a member from a workspace. OWNER can remove anyone; ADMIN can remove MEMBERs.
func (s *WorkspaceService) RemoveMember(ctx context.Context, removerID, workspaceID, targetUserID uuid.UUID) error {
	// Verify remover membership
	remover, err := s.queries.GetWorkspaceMember(ctx, repository.GetWorkspaceMemberParams{
		UserID:      removerID,
		WorkspaceID: workspaceID,
	})
	if err != nil {
		return fmt.Errorf("you are not a member of this workspace")
	}

	// Self-removal (leaving workspace) is allowed for non-owners
	if removerID == targetUserID {
		if remover.Role == repository.RoleOWNER {
			return fmt.Errorf("owner cannot leave the workspace; transfer ownership or delete the workspace")
		}
		return s.queries.RemoveWorkspaceMember(ctx, repository.RemoveWorkspaceMemberParams{
			UserID:      targetUserID,
			WorkspaceID: workspaceID,
		})
	}

	// Check permissions for removing others
	target, err := s.queries.GetWorkspaceMember(ctx, repository.GetWorkspaceMemberParams{
		UserID:      targetUserID,
		WorkspaceID: workspaceID,
	})
	if err != nil {
		return fmt.Errorf("target user is not a member of this workspace")
	}

	switch remover.Role {
	case repository.RoleOWNER:
		// Owner can remove anyone except themselves (handled above)
	case repository.RoleADMIN:
		// Admin can only remove MEMBERs
		if target.Role != repository.RoleMEMBER {
			return fmt.Errorf("admin can only remove MEMBER role users")
		}
	default:
		return fmt.Errorf("you do not have permission to remove members")
	}

	return s.queries.RemoveWorkspaceMember(ctx, repository.RemoveWorkspaceMemberParams{
		UserID:      targetUserID,
		WorkspaceID: workspaceID,
	})
}
