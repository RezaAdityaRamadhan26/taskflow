// Package handler implements HTTP handlers for activity logs.
package handler

import (
	"strconv"

	"github.com/gofiber/fiber/v2"
	"github.com/google/uuid"

	"github.com/RezaAdityaRamadhan26/taskflow/backend/internal/service"
	"github.com/RezaAdityaRamadhan26/taskflow/backend/internal/util"
)

// ActivityLogHandler handles HTTP requests for activity logs endpoints.
type ActivityLogHandler struct {
	activityLogService *service.ActivityLogService
}

// NewActivityLogHandler creates a new ActivityLogHandler instance.
func NewActivityLogHandler(activityLogService *service.ActivityLogService) *ActivityLogHandler {
	return &ActivityLogHandler{
		activityLogService: activityLogService,
	}
}

// ListBoardActivities handles GET /api/v1/boards/:boardId/activities
func (h *ActivityLogHandler) ListBoardActivities(c *fiber.Ctx) error {
	userID, ok := c.Locals("userID").(uuid.UUID)
	if !ok {
		return util.ErrorResponse(c, fiber.StatusUnauthorized, "Invalid user session")
	}

	boardID, err := uuid.Parse(c.Params("boardId"))
	if err != nil {
		return util.ErrorResponse(c, fiber.StatusBadRequest, "Invalid board ID")
	}

	limit := int32(50)
	if l := c.Query("limit"); l != "" {
		if val, err := strconv.Atoi(l); err == nil {
			limit = int32(val)
		}
	}

	offset := int32(0)
	if o := c.Query("offset"); o != "" {
		if val, err := strconv.Atoi(o); err == nil {
			offset = int32(val)
		}
	}

	logs, err := h.activityLogService.ListBoardActivities(c.Context(), userID, boardID, limit, offset)
	if err != nil {
		switch err.Error() {
		case "board not found":
			return util.ErrorResponse(c, fiber.StatusNotFound, err.Error())
		case "you do not have access to this board":
			return util.ErrorResponse(c, fiber.StatusForbidden, err.Error())
		default:
			return util.ErrorResponse(c, fiber.StatusInternalServerError, err.Error())
		}
	}

	return util.SuccessResponse(c, fiber.StatusOK, "", logs)
}

// ListCardActivities handles GET /api/v1/cards/:cardId/activities
func (h *ActivityLogHandler) ListCardActivities(c *fiber.Ctx) error {
	userID, ok := c.Locals("userID").(uuid.UUID)
	if !ok {
		return util.ErrorResponse(c, fiber.StatusUnauthorized, "Invalid user session")
	}

	cardID, err := uuid.Parse(c.Params("cardId"))
	if err != nil {
		return util.ErrorResponse(c, fiber.StatusBadRequest, "Invalid card ID")
	}

	limit := int32(50)
	if l := c.Query("limit"); l != "" {
		if val, err := strconv.Atoi(l); err == nil {
			limit = int32(val)
		}
	}

	offset := int32(0)
	if o := c.Query("offset"); o != "" {
		if val, err := strconv.Atoi(o); err == nil {
			offset = int32(val)
		}
	}

	logs, err := h.activityLogService.ListCardActivities(c.Context(), userID, cardID, limit, offset)
	if err != nil {
		switch err.Error() {
		case "card not found", "list not found":
			return util.ErrorResponse(c, fiber.StatusNotFound, err.Error())
		case "you do not have access to this board":
			return util.ErrorResponse(c, fiber.StatusForbidden, err.Error())
		default:
			return util.ErrorResponse(c, fiber.StatusInternalServerError, err.Error())
		}
	}

	return util.SuccessResponse(c, fiber.StatusOK, "", logs)
}
