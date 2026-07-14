// Package handler implements HTTP handlers for workspace endpoints.
package handler

import (
	"log"
	"regexp"

	"github.com/go-playground/validator/v10"
	"github.com/gofiber/fiber/v2"
	"github.com/google/uuid"

	"github.com/RezaAdityaRamadhan26/taskflow/backend/internal/dto"
	"github.com/RezaAdityaRamadhan26/taskflow/backend/internal/service"
	"github.com/RezaAdityaRamadhan26/taskflow/backend/internal/util"
)

// WorkspaceHandler handles HTTP requests for workspace-related endpoints.
type WorkspaceHandler struct {
	workspaceService *service.WorkspaceService
	validate         *validator.Validate
}

// NewWorkspaceHandler creates a new WorkspaceHandler instance.
func NewWorkspaceHandler(workspaceService *service.WorkspaceService) *WorkspaceHandler {
	v := validator.New()
	// Register custom validation for slug (lowercase alphanum + hyphens)
	v.RegisterValidation("alphanum_with_dash", func(fl validator.FieldLevel) bool {
		return regexp.MustCompile(`^[a-z0-9]+(?:-[a-z0-9]+)*$`).MatchString(fl.Field().String())
	})

	return &WorkspaceHandler{
		workspaceService: workspaceService,
		validate:         v,
	}
}

// Create handles POST /api/v1/workspaces
// Creates a new workspace and adds the creator as OWNER.
func (h *WorkspaceHandler) Create(c *fiber.Ctx) error {
	userID, ok := c.Locals("userID").(uuid.UUID)
	if !ok {
		return util.ErrorResponse(c, fiber.StatusUnauthorized, "Invalid user session")
	}

	var req dto.CreateWorkspaceRequest
	if err := c.BodyParser(&req); err != nil {
		return util.ErrorResponse(c, fiber.StatusBadRequest, "Invalid request body")
	}

	if err := h.validate.Struct(req); err != nil {
		validationErrors := err.(validator.ValidationErrors)
		return util.ValidationErrorResponse(c, formatWorkspaceValidationError(validationErrors))
	}

	workspace, err := h.workspaceService.CreateWorkspace(c.Context(), userID, req)
	if err != nil {
		if err.Error() == "slug already taken" {
			return util.ErrorResponse(c, fiber.StatusConflict, err.Error())
		}
		log.Printf("[ERROR] Create workspace failed: %v", err)
		return util.ErrorResponse(c, fiber.StatusInternalServerError, err.Error())
	}

	return util.SuccessResponse(c, fiber.StatusCreated, "Workspace created successfully", workspace)
}

// List handles GET /api/v1/workspaces
// Returns all workspaces for the authenticated user.
func (h *WorkspaceHandler) List(c *fiber.Ctx) error {
	userID, ok := c.Locals("userID").(uuid.UUID)
	if !ok {
		return util.ErrorResponse(c, fiber.StatusUnauthorized, "Invalid user session")
	}

	workspaces, err := h.workspaceService.ListWorkspaces(c.Context(), userID)
	if err != nil {
		return util.ErrorResponse(c, fiber.StatusInternalServerError, err.Error())
	}

	return util.SuccessResponse(c, fiber.StatusOK, "", workspaces)
}

// Get handles GET /api/v1/workspaces/:id
// Returns a workspace by ID.
func (h *WorkspaceHandler) Get(c *fiber.Ctx) error {
	userID, ok := c.Locals("userID").(uuid.UUID)
	if !ok {
		return util.ErrorResponse(c, fiber.StatusUnauthorized, "Invalid user session")
	}

	workspaceID, err := uuid.Parse(c.Params("id"))
	if err != nil {
		return util.ErrorResponse(c, fiber.StatusBadRequest, "Invalid workspace ID")
	}

	workspace, err := h.workspaceService.GetWorkspace(c.Context(), userID, workspaceID)
	if err != nil {
		if err.Error() == "you are not a member of this workspace" {
			return util.ErrorResponse(c, fiber.StatusForbidden, err.Error())
		}
		return util.ErrorResponse(c, fiber.StatusNotFound, err.Error())
	}

	return util.SuccessResponse(c, fiber.StatusOK, "", workspace)
}

// Update handles PUT /api/v1/workspaces/:id
// Updates a workspace. Only OWNER/ADMIN can update.
func (h *WorkspaceHandler) Update(c *fiber.Ctx) error {
	userID, ok := c.Locals("userID").(uuid.UUID)
	if !ok {
		return util.ErrorResponse(c, fiber.StatusUnauthorized, "Invalid user session")
	}

	workspaceID, err := uuid.Parse(c.Params("id"))
	if err != nil {
		return util.ErrorResponse(c, fiber.StatusBadRequest, "Invalid workspace ID")
	}

	var req dto.UpdateWorkspaceRequest
	if err := c.BodyParser(&req); err != nil {
		return util.ErrorResponse(c, fiber.StatusBadRequest, "Invalid request body")
	}

	if err := h.validate.Struct(req); err != nil {
		validationErrors := err.(validator.ValidationErrors)
		return util.ValidationErrorResponse(c, formatWorkspaceValidationError(validationErrors))
	}

	workspace, err := h.workspaceService.UpdateWorkspace(c.Context(), userID, workspaceID, req)
	if err != nil {
		switch err.Error() {
		case "you are not a member of this workspace":
			return util.ErrorResponse(c, fiber.StatusForbidden, err.Error())
		case "only OWNER or ADMIN can update workspace":
			return util.ErrorResponse(c, fiber.StatusForbidden, err.Error())
		case "slug already taken":
			return util.ErrorResponse(c, fiber.StatusConflict, err.Error())
		default:
			return util.ErrorResponse(c, fiber.StatusInternalServerError, err.Error())
		}
	}

	return util.SuccessResponse(c, fiber.StatusOK, "Workspace updated successfully", workspace)
}

// Delete handles DELETE /api/v1/workspaces/:id
// Deletes a workspace. Only OWNER can delete.
func (h *WorkspaceHandler) Delete(c *fiber.Ctx) error {
	userID, ok := c.Locals("userID").(uuid.UUID)
	if !ok {
		return util.ErrorResponse(c, fiber.StatusUnauthorized, "Invalid user session")
	}

	workspaceID, err := uuid.Parse(c.Params("id"))
	if err != nil {
		return util.ErrorResponse(c, fiber.StatusBadRequest, "Invalid workspace ID")
	}

	if err := h.workspaceService.DeleteWorkspace(c.Context(), userID, workspaceID); err != nil {
		switch err.Error() {
		case "you are not a member of this workspace":
			return util.ErrorResponse(c, fiber.StatusForbidden, err.Error())
		case "only OWNER can delete workspace":
			return util.ErrorResponse(c, fiber.StatusForbidden, err.Error())
		default:
			return util.ErrorResponse(c, fiber.StatusInternalServerError, err.Error())
		}
	}

	return util.SuccessResponse(c, fiber.StatusOK, "Workspace deleted successfully", nil)
}

// InviteMember handles POST /api/v1/workspaces/:id/members
// Invites a user to the workspace by email.
func (h *WorkspaceHandler) InviteMember(c *fiber.Ctx) error {
	userID, ok := c.Locals("userID").(uuid.UUID)
	if !ok {
		return util.ErrorResponse(c, fiber.StatusUnauthorized, "Invalid user session")
	}

	workspaceID, err := uuid.Parse(c.Params("id"))
	if err != nil {
		return util.ErrorResponse(c, fiber.StatusBadRequest, "Invalid workspace ID")
	}

	var req dto.InviteMemberRequest
	if err := c.BodyParser(&req); err != nil {
		return util.ErrorResponse(c, fiber.StatusBadRequest, "Invalid request body")
	}

	if err := h.validate.Struct(req); err != nil {
		validationErrors := err.(validator.ValidationErrors)
		return util.ValidationErrorResponse(c, formatWorkspaceValidationError(validationErrors))
	}

	member, err := h.workspaceService.InviteMember(c.Context(), userID, workspaceID, req)
	if err != nil {
		switch err.Error() {
		case "you are not a member of this workspace":
			return util.ErrorResponse(c, fiber.StatusForbidden, err.Error())
		case "only OWNER or ADMIN can invite members":
			return util.ErrorResponse(c, fiber.StatusForbidden, err.Error())
		case "only OWNER can invite as ADMIN":
			return util.ErrorResponse(c, fiber.StatusForbidden, err.Error())
		case "user is already a member of this workspace":
			return util.ErrorResponse(c, fiber.StatusConflict, err.Error())
		default:
			if len(err.Error()) > 15 && err.Error()[:15] == "user with email" {
				return util.ErrorResponse(c, fiber.StatusNotFound, err.Error())
			}
			return util.ErrorResponse(c, fiber.StatusInternalServerError, err.Error())
		}
	}

	return util.SuccessResponse(c, fiber.StatusCreated, "Member invited successfully", member)
}

// ListMembers handles GET /api/v1/workspaces/:id/members
// Returns all members of a workspace.
func (h *WorkspaceHandler) ListMembers(c *fiber.Ctx) error {
	userID, ok := c.Locals("userID").(uuid.UUID)
	if !ok {
		return util.ErrorResponse(c, fiber.StatusUnauthorized, "Invalid user session")
	}

	workspaceID, err := uuid.Parse(c.Params("id"))
	if err != nil {
		return util.ErrorResponse(c, fiber.StatusBadRequest, "Invalid workspace ID")
	}

	members, err := h.workspaceService.ListMembers(c.Context(), userID, workspaceID)
	if err != nil {
		if err.Error() == "you are not a member of this workspace" {
			return util.ErrorResponse(c, fiber.StatusForbidden, err.Error())
		}
		return util.ErrorResponse(c, fiber.StatusInternalServerError, err.Error())
	}

	return util.SuccessResponse(c, fiber.StatusOK, "", members)
}

// UpdateMemberRole handles PATCH /api/v1/workspaces/:id/members/:userId
// Updates a member's role. Only OWNER can change roles.
func (h *WorkspaceHandler) UpdateMemberRole(c *fiber.Ctx) error {
	userID, ok := c.Locals("userID").(uuid.UUID)
	if !ok {
		return util.ErrorResponse(c, fiber.StatusUnauthorized, "Invalid user session")
	}

	workspaceID, err := uuid.Parse(c.Params("id"))
	if err != nil {
		return util.ErrorResponse(c, fiber.StatusBadRequest, "Invalid workspace ID")
	}

	targetUserID, err := uuid.Parse(c.Params("userId"))
	if err != nil {
		return util.ErrorResponse(c, fiber.StatusBadRequest, "Invalid user ID")
	}

	var req dto.UpdateMemberRoleRequest
	if err := c.BodyParser(&req); err != nil {
		return util.ErrorResponse(c, fiber.StatusBadRequest, "Invalid request body")
	}

	if err := h.validate.Struct(req); err != nil {
		validationErrors := err.(validator.ValidationErrors)
		return util.ValidationErrorResponse(c, formatWorkspaceValidationError(validationErrors))
	}

	member, err := h.workspaceService.UpdateMemberRole(c.Context(), userID, workspaceID, targetUserID, req)
	if err != nil {
		switch err.Error() {
		case "you are not a member of this workspace", "only OWNER can change member roles", "owner cannot change their own role":
			return util.ErrorResponse(c, fiber.StatusForbidden, err.Error())
		case "target user is not a member of this workspace":
			return util.ErrorResponse(c, fiber.StatusNotFound, err.Error())
		default:
			return util.ErrorResponse(c, fiber.StatusInternalServerError, err.Error())
		}
	}

	return util.SuccessResponse(c, fiber.StatusOK, "Member role updated successfully", member)
}

// RemoveMember handles DELETE /api/v1/workspaces/:id/members/:userId
// Removes a member from the workspace.
func (h *WorkspaceHandler) RemoveMember(c *fiber.Ctx) error {
	userID, ok := c.Locals("userID").(uuid.UUID)
	if !ok {
		return util.ErrorResponse(c, fiber.StatusUnauthorized, "Invalid user session")
	}

	workspaceID, err := uuid.Parse(c.Params("id"))
	if err != nil {
		return util.ErrorResponse(c, fiber.StatusBadRequest, "Invalid workspace ID")
	}

	targetUserID, err := uuid.Parse(c.Params("userId"))
	if err != nil {
		return util.ErrorResponse(c, fiber.StatusBadRequest, "Invalid user ID")
	}

	if err := h.workspaceService.RemoveMember(c.Context(), userID, workspaceID, targetUserID); err != nil {
		switch err.Error() {
		case "you are not a member of this workspace",
			"you do not have permission to remove members",
			"admin can only remove MEMBER role users",
			"owner cannot leave the workspace; transfer ownership or delete the workspace":
			return util.ErrorResponse(c, fiber.StatusForbidden, err.Error())
		case "target user is not a member of this workspace":
			return util.ErrorResponse(c, fiber.StatusNotFound, err.Error())
		default:
			return util.ErrorResponse(c, fiber.StatusInternalServerError, err.Error())
		}
	}

	return util.SuccessResponse(c, fiber.StatusOK, "Member removed successfully", nil)
}

// formatWorkspaceValidationError converts validator errors into a human-readable message.
func formatWorkspaceValidationError(errs validator.ValidationErrors) string {
	if len(errs) == 0 {
		return "Validation failed"
	}

	fieldMessages := map[string]map[string]string{
		"Name": {
			"required": "Name is required",
			"min":      "Name must be at least 2 characters",
			"max":      "Name must not exceed 100 characters",
		},
		"Description": {
			"max": "Description must not exceed 500 characters",
		},
		"Slug": {
			"required":            "Slug is required",
			"min":                 "Slug must be at least 2 characters",
			"max":                 "Slug must not exceed 100 characters",
			"alphanum_with_dash":  "Slug must contain only lowercase letters, numbers, and hyphens",
		},
		"Email": {
			"required": "Email is required",
			"email":    "Email must be a valid email address",
		},
		"Role": {
			"required": "Role is required",
			"oneof":    "Role must be one of: OWNER, ADMIN, MEMBER",
		},
	}

	firstErr := errs[0]
	if msgs, ok := fieldMessages[firstErr.Field()]; ok {
		if msg, ok := msgs[firstErr.Tag()]; ok {
			return msg
		}
	}

	return "Validation failed for field: " + firstErr.Field()
}
