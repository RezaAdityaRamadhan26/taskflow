// Package handler implements HTTP handlers for list endpoints.
package handler

import (
	"log"

	"github.com/go-playground/validator/v10"
	"github.com/gofiber/fiber/v2"
	"github.com/google/uuid"

	"github.com/RezaAdityaRamadhan26/taskflow/backend/internal/dto"
	"github.com/RezaAdityaRamadhan26/taskflow/backend/internal/service"
	"github.com/RezaAdityaRamadhan26/taskflow/backend/internal/util"
)

// ListHandler handles HTTP requests for list-related endpoints.
type ListHandler struct {
	listService *service.ListService
	validate    *validator.Validate
}

// NewListHandler creates a new ListHandler instance.
func NewListHandler(listService *service.ListService) *ListHandler {
	return &ListHandler{
		listService: listService,
		validate:    validator.New(),
	}
}

// Create handles POST /api/v1/lists
// Creates a new list in a board.
func (h *ListHandler) Create(c *fiber.Ctx) error {
	userID, ok := c.Locals("userID").(uuid.UUID)
	if !ok {
		return util.ErrorResponse(c, fiber.StatusUnauthorized, "Invalid user session")
	}

	var req dto.CreateListRequest
	if err := c.BodyParser(&req); err != nil {
		return util.ErrorResponse(c, fiber.StatusBadRequest, "Invalid request body")
	}

	if err := h.validate.Struct(req); err != nil {
		validationErrors := err.(validator.ValidationErrors)
		return util.ValidationErrorResponse(c, formatListValidationError(validationErrors))
	}

	list, err := h.listService.CreateList(c.Context(), userID, req)
	if err != nil {
		switch err.Error() {
		case "invalid board ID", "board not found":
			return util.ErrorResponse(c, fiber.StatusBadRequest, err.Error())
		case "you are not a member of this workspace", "only OWNER or ADMIN can modify lists":
			return util.ErrorResponse(c, fiber.StatusForbidden, err.Error())
		default:
			log.Printf("[ERROR] Create list failed: %v", err)
			return util.ErrorResponse(c, fiber.StatusInternalServerError, err.Error())
		}
	}

	return util.SuccessResponse(c, fiber.StatusCreated, "List created successfully", list)
}

// List handles GET /api/v1/boards/:boardId/lists
// Returns all lists in a board.
func (h *ListHandler) List(c *fiber.Ctx) error {
	userID, ok := c.Locals("userID").(uuid.UUID)
	if !ok {
		return util.ErrorResponse(c, fiber.StatusUnauthorized, "Invalid user session")
	}

	boardID, err := uuid.Parse(c.Params("boardId"))
	if err != nil {
		return util.ErrorResponse(c, fiber.StatusBadRequest, "Invalid board ID")
	}

	lists, err := h.listService.ListLists(c.Context(), userID, boardID)
	if err != nil {
		switch err.Error() {
		case "board not found":
			return util.ErrorResponse(c, fiber.StatusNotFound, err.Error())
		case "you are not a member of this workspace":
			return util.ErrorResponse(c, fiber.StatusForbidden, err.Error())
		default:
			return util.ErrorResponse(c, fiber.StatusInternalServerError, err.Error())
		}
	}

	return util.SuccessResponse(c, fiber.StatusOK, "", lists)
}

// Get handles GET /api/v1/lists/:id
// Returns a list by ID.
func (h *ListHandler) Get(c *fiber.Ctx) error {
	userID, ok := c.Locals("userID").(uuid.UUID)
	if !ok {
		return util.ErrorResponse(c, fiber.StatusUnauthorized, "Invalid user session")
	}

	listID, err := uuid.Parse(c.Params("id"))
	if err != nil {
		return util.ErrorResponse(c, fiber.StatusBadRequest, "Invalid list ID")
	}

	list, err := h.listService.GetList(c.Context(), userID, listID)
	if err != nil {
		switch err.Error() {
		case "list not found", "board not found":
			return util.ErrorResponse(c, fiber.StatusNotFound, err.Error())
		case "you are not a member of this workspace":
			return util.ErrorResponse(c, fiber.StatusForbidden, err.Error())
		default:
			return util.ErrorResponse(c, fiber.StatusInternalServerError, err.Error())
		}
	}

	return util.SuccessResponse(c, fiber.StatusOK, "", list)
}

// Update handles PUT /api/v1/lists/:id
// Updates a list (name and position).
func (h *ListHandler) Update(c *fiber.Ctx) error {
	userID, ok := c.Locals("userID").(uuid.UUID)
	if !ok {
		return util.ErrorResponse(c, fiber.StatusUnauthorized, "Invalid user session")
	}

	listID, err := uuid.Parse(c.Params("id"))
	if err != nil {
		return util.ErrorResponse(c, fiber.StatusBadRequest, "Invalid list ID")
	}

	var req dto.UpdateListRequest
	if err := c.BodyParser(&req); err != nil {
		return util.ErrorResponse(c, fiber.StatusBadRequest, "Invalid request body")
	}

	if err := h.validate.Struct(req); err != nil {
		validationErrors := err.(validator.ValidationErrors)
		return util.ValidationErrorResponse(c, formatListValidationError(validationErrors))
	}

	list, err := h.listService.UpdateList(c.Context(), userID, listID, req)
	if err != nil {
		switch err.Error() {
		case "list not found", "board not found":
			return util.ErrorResponse(c, fiber.StatusNotFound, err.Error())
		case "you are not a member of this workspace", "only OWNER or ADMIN can modify lists":
			return util.ErrorResponse(c, fiber.StatusForbidden, err.Error())
		default:
			return util.ErrorResponse(c, fiber.StatusInternalServerError, err.Error())
		}
	}

	return util.SuccessResponse(c, fiber.StatusOK, "List updated successfully", list)
}

// Delete handles DELETE /api/v1/lists/:id
// Deletes a list.
func (h *ListHandler) Delete(c *fiber.Ctx) error {
	userID, ok := c.Locals("userID").(uuid.UUID)
	if !ok {
		return util.ErrorResponse(c, fiber.StatusUnauthorized, "Invalid user session")
	}

	listID, err := uuid.Parse(c.Params("id"))
	if err != nil {
		return util.ErrorResponse(c, fiber.StatusBadRequest, "Invalid list ID")
	}

	if err := h.listService.DeleteList(c.Context(), userID, listID); err != nil {
		switch err.Error() {
		case "list not found", "board not found":
			return util.ErrorResponse(c, fiber.StatusNotFound, err.Error())
		case "you are not a member of this workspace", "only OWNER or ADMIN can modify lists":
			return util.ErrorResponse(c, fiber.StatusForbidden, err.Error())
		default:
			return util.ErrorResponse(c, fiber.StatusInternalServerError, err.Error())
		}
	}

	return util.SuccessResponse(c, fiber.StatusOK, "List deleted successfully", nil)
}

// formatListValidationError converts validator errors into a human-readable message.
func formatListValidationError(errs validator.ValidationErrors) string {
	if len(errs) == 0 {
		return "Validation failed"
	}

	fieldMessages := map[string]map[string]string{
		"BoardID": {
			"required": "Board ID is required",
			"uuid":     "Board ID must be a valid UUID",
		},
		"Name": {
			"required": "Name is required",
			"min":      "Name must be at least 1 character",
			"max":      "Name must not exceed 100 characters",
		},
		"Position": {
			"required": "Position is required",
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
