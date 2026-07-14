// Package handler implements HTTP handlers for board endpoints.
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

// BoardHandler handles HTTP requests for board-related endpoints.
type BoardHandler struct {
	boardService *service.BoardService
	validate     *validator.Validate
}

// NewBoardHandler creates a new BoardHandler instance.
func NewBoardHandler(boardService *service.BoardService) *BoardHandler {
	return &BoardHandler{
		boardService: boardService,
		validate:     validator.New(),
	}
}

// Create handles POST /api/v1/boards
// Creates a new board in a workspace.
func (h *BoardHandler) Create(c *fiber.Ctx) error {
	userID, ok := c.Locals("userID").(uuid.UUID)
	if !ok {
		return util.ErrorResponse(c, fiber.StatusUnauthorized, "Invalid user session")
	}

	var req dto.CreateBoardRequest
	if err := c.BodyParser(&req); err != nil {
		return util.ErrorResponse(c, fiber.StatusBadRequest, "Invalid request body")
	}

	if err := h.validate.Struct(req); err != nil {
		validationErrors := err.(validator.ValidationErrors)
		return util.ValidationErrorResponse(c, formatBoardValidationError(validationErrors))
	}

	board, err := h.boardService.CreateBoard(c.Context(), userID, req)
	if err != nil {
		switch err.Error() {
		case "invalid workspace ID":
			return util.ErrorResponse(c, fiber.StatusBadRequest, err.Error())
		case "you are not a member of this workspace", "only OWNER or ADMIN can create boards":
			return util.ErrorResponse(c, fiber.StatusForbidden, err.Error())
		default:
			log.Printf("[ERROR] Create board failed: %v", err)
			return util.ErrorResponse(c, fiber.StatusInternalServerError, err.Error())
		}
	}

	return util.SuccessResponse(c, fiber.StatusCreated, "Board created successfully", board)
}

// List handles GET /api/v1/workspaces/:workspaceId/boards
// Returns all boards in a workspace.
func (h *BoardHandler) List(c *fiber.Ctx) error {
	userID, ok := c.Locals("userID").(uuid.UUID)
	if !ok {
		return util.ErrorResponse(c, fiber.StatusUnauthorized, "Invalid user session")
	}

	workspaceID, err := uuid.Parse(c.Params("workspaceId"))
	if err != nil {
		return util.ErrorResponse(c, fiber.StatusBadRequest, "Invalid workspace ID")
	}

	boards, err := h.boardService.ListBoards(c.Context(), userID, workspaceID)
	if err != nil {
		if err.Error() == "you are not a member of this workspace" {
			return util.ErrorResponse(c, fiber.StatusForbidden, err.Error())
		}
		return util.ErrorResponse(c, fiber.StatusInternalServerError, err.Error())
	}

	return util.SuccessResponse(c, fiber.StatusOK, "", boards)
}

// Get handles GET /api/v1/boards/:id
// Returns a board by ID.
func (h *BoardHandler) Get(c *fiber.Ctx) error {
	userID, ok := c.Locals("userID").(uuid.UUID)
	if !ok {
		return util.ErrorResponse(c, fiber.StatusUnauthorized, "Invalid user session")
	}

	boardID, err := uuid.Parse(c.Params("id"))
	if err != nil {
		return util.ErrorResponse(c, fiber.StatusBadRequest, "Invalid board ID")
	}

	board, err := h.boardService.GetBoard(c.Context(), userID, boardID)
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

	return util.SuccessResponse(c, fiber.StatusOK, "", board)
}

// Update handles PUT /api/v1/boards/:id
// Updates a board.
func (h *BoardHandler) Update(c *fiber.Ctx) error {
	userID, ok := c.Locals("userID").(uuid.UUID)
	if !ok {
		return util.ErrorResponse(c, fiber.StatusUnauthorized, "Invalid user session")
	}

	boardID, err := uuid.Parse(c.Params("id"))
	if err != nil {
		return util.ErrorResponse(c, fiber.StatusBadRequest, "Invalid board ID")
	}

	var req dto.UpdateBoardRequest
	if err := c.BodyParser(&req); err != nil {
		return util.ErrorResponse(c, fiber.StatusBadRequest, "Invalid request body")
	}

	if err := h.validate.Struct(req); err != nil {
		validationErrors := err.(validator.ValidationErrors)
		return util.ValidationErrorResponse(c, formatBoardValidationError(validationErrors))
	}

	board, err := h.boardService.UpdateBoard(c.Context(), userID, boardID, req)
	if err != nil {
		switch err.Error() {
		case "board not found":
			return util.ErrorResponse(c, fiber.StatusNotFound, err.Error())
		case "you are not a member of this workspace", "only OWNER or ADMIN can update boards":
			return util.ErrorResponse(c, fiber.StatusForbidden, err.Error())
		default:
			return util.ErrorResponse(c, fiber.StatusInternalServerError, err.Error())
		}
	}

	return util.SuccessResponse(c, fiber.StatusOK, "Board updated successfully", board)
}

// Delete handles DELETE /api/v1/boards/:id
// Deletes a board.
func (h *BoardHandler) Delete(c *fiber.Ctx) error {
	userID, ok := c.Locals("userID").(uuid.UUID)
	if !ok {
		return util.ErrorResponse(c, fiber.StatusUnauthorized, "Invalid user session")
	}

	boardID, err := uuid.Parse(c.Params("id"))
	if err != nil {
		return util.ErrorResponse(c, fiber.StatusBadRequest, "Invalid board ID")
	}

	if err := h.boardService.DeleteBoard(c.Context(), userID, boardID); err != nil {
		switch err.Error() {
		case "board not found":
			return util.ErrorResponse(c, fiber.StatusNotFound, err.Error())
		case "you are not a member of this workspace", "only OWNER or ADMIN can delete boards":
			return util.ErrorResponse(c, fiber.StatusForbidden, err.Error())
		default:
			return util.ErrorResponse(c, fiber.StatusInternalServerError, err.Error())
		}
	}

	return util.SuccessResponse(c, fiber.StatusOK, "Board deleted successfully", nil)
}

// formatBoardValidationError converts validator errors into a human-readable message.
func formatBoardValidationError(errs validator.ValidationErrors) string {
	if len(errs) == 0 {
		return "Validation failed"
	}

	fieldMessages := map[string]map[string]string{
		"WorkspaceID": {
			"required": "Workspace ID is required",
			"uuid":     "Workspace ID must be a valid UUID",
		},
		"Name": {
			"required": "Name is required",
			"min":      "Name must be at least 2 characters",
			"max":      "Name must not exceed 100 characters",
		},
		"Description": {
			"max": "Description must not exceed 500 characters",
		},
		"Color": {
			"hexcolor": "Color must be a valid hex color code (e.g., #FFFFFF)",
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
