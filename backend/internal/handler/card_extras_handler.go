// Package handler implements HTTP handlers for comments and attachments.
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

// CardExtrasHandler handles HTTP requests for comments and attachments endpoints.
type CardExtrasHandler struct {
	extrasService *service.CardExtrasService
	validate      *validator.Validate
}

// NewCardExtrasHandler creates a new CardExtrasHandler instance.
func NewCardExtrasHandler(extrasService *service.CardExtrasService) *CardExtrasHandler {
	return &CardExtrasHandler{
		extrasService: extrasService,
		validate:      validator.New(),
	}
}

// CreateComment handles POST /api/v1/comments
func (h *CardExtrasHandler) CreateComment(c *fiber.Ctx) error {
	userID, ok := c.Locals("userID").(uuid.UUID)
	if !ok {
		return util.ErrorResponse(c, fiber.StatusUnauthorized, "Invalid user session")
	}

	var req dto.CreateCommentRequest
	if err := c.BodyParser(&req); err != nil {
		return util.ErrorResponse(c, fiber.StatusBadRequest, "Invalid request body")
	}

	if err := h.validate.Struct(req); err != nil {
		validationErrors := err.(validator.ValidationErrors)
		return util.ValidationErrorResponse(c, formatCardExtrasValidationError(validationErrors))
	}

	comment, err := h.extrasService.CreateComment(c.Context(), userID, req)
	if err != nil {
		switch err.Error() {
		case "invalid card ID", "card not found", "list not found", "board not found":
			return util.ErrorResponse(c, fiber.StatusBadRequest, err.Error())
		case "you do not have access to this card":
			return util.ErrorResponse(c, fiber.StatusForbidden, err.Error())
		default:
			log.Printf("[ERROR] Create comment failed: %v", err)
			return util.ErrorResponse(c, fiber.StatusInternalServerError, err.Error())
		}
	}

	return util.SuccessResponse(c, fiber.StatusCreated, "Comment added successfully", comment)
}

// ListComments handles GET /api/v1/cards/:cardId/comments
func (h *CardExtrasHandler) ListComments(c *fiber.Ctx) error {
	userID, ok := c.Locals("userID").(uuid.UUID)
	if !ok {
		return util.ErrorResponse(c, fiber.StatusUnauthorized, "Invalid user session")
	}

	cardID, err := uuid.Parse(c.Params("cardId"))
	if err != nil {
		return util.ErrorResponse(c, fiber.StatusBadRequest, "Invalid card ID")
	}

	comments, err := h.extrasService.ListComments(c.Context(), userID, cardID)
	if err != nil {
		switch err.Error() {
		case "card not found", "list not found", "board not found":
			return util.ErrorResponse(c, fiber.StatusNotFound, err.Error())
		case "you do not have access to this card":
			return util.ErrorResponse(c, fiber.StatusForbidden, err.Error())
		default:
			return util.ErrorResponse(c, fiber.StatusInternalServerError, err.Error())
		}
	}

	return util.SuccessResponse(c, fiber.StatusOK, "", comments)
}

// UpdateComment handles PUT /api/v1/comments/:id
func (h *CardExtrasHandler) UpdateComment(c *fiber.Ctx) error {
	userID, ok := c.Locals("userID").(uuid.UUID)
	if !ok {
		return util.ErrorResponse(c, fiber.StatusUnauthorized, "Invalid user session")
	}

	commentID, err := uuid.Parse(c.Params("id"))
	if err != nil {
		return util.ErrorResponse(c, fiber.StatusBadRequest, "Invalid comment ID")
	}

	var req dto.UpdateCommentRequest
	if err := c.BodyParser(&req); err != nil {
		return util.ErrorResponse(c, fiber.StatusBadRequest, "Invalid request body")
	}

	if err := h.validate.Struct(req); err != nil {
		validationErrors := err.(validator.ValidationErrors)
		return util.ValidationErrorResponse(c, formatCardExtrasValidationError(validationErrors))
	}

	comment, err := h.extrasService.UpdateComment(c.Context(), userID, commentID, req)
	if err != nil {
		switch err.Error() {
		case "comment not found":
			return util.ErrorResponse(c, fiber.StatusNotFound, err.Error())
		case "you can only edit your own comments":
			return util.ErrorResponse(c, fiber.StatusForbidden, err.Error())
		default:
			return util.ErrorResponse(c, fiber.StatusInternalServerError, err.Error())
		}
	}

	return util.SuccessResponse(c, fiber.StatusOK, "Comment updated successfully", comment)
}

// DeleteComment handles DELETE /api/v1/comments/:id
func (h *CardExtrasHandler) DeleteComment(c *fiber.Ctx) error {
	userID, ok := c.Locals("userID").(uuid.UUID)
	if !ok {
		return util.ErrorResponse(c, fiber.StatusUnauthorized, "Invalid user session")
	}

	commentID, err := uuid.Parse(c.Params("id"))
	if err != nil {
		return util.ErrorResponse(c, fiber.StatusBadRequest, "Invalid comment ID")
	}

	if err := h.extrasService.DeleteComment(c.Context(), userID, commentID); err != nil {
		switch err.Error() {
		case "comment not found":
			return util.ErrorResponse(c, fiber.StatusNotFound, err.Error())
		case "you can only delete your own comments":
			return util.ErrorResponse(c, fiber.StatusForbidden, err.Error())
		default:
			return util.ErrorResponse(c, fiber.StatusInternalServerError, err.Error())
		}
	}

	return util.SuccessResponse(c, fiber.StatusOK, "Comment deleted successfully", nil)
}

// CreateAttachment handles POST /api/v1/attachments
func (h *CardExtrasHandler) CreateAttachment(c *fiber.Ctx) error {
	userID, ok := c.Locals("userID").(uuid.UUID)
	if !ok {
		return util.ErrorResponse(c, fiber.StatusUnauthorized, "Invalid user session")
	}

	var req dto.CreateAttachmentRequest
	if err := c.BodyParser(&req); err != nil {
		return util.ErrorResponse(c, fiber.StatusBadRequest, "Invalid request body")
	}

	if err := h.validate.Struct(req); err != nil {
		validationErrors := err.(validator.ValidationErrors)
		return util.ValidationErrorResponse(c, formatCardExtrasValidationError(validationErrors))
	}

	attachment, err := h.extrasService.CreateAttachment(c.Context(), userID, req)
	if err != nil {
		switch err.Error() {
		case "invalid card ID", "card not found", "list not found", "board not found":
			return util.ErrorResponse(c, fiber.StatusBadRequest, err.Error())
		case "you do not have access to this card":
			return util.ErrorResponse(c, fiber.StatusForbidden, err.Error())
		default:
			log.Printf("[ERROR] Create attachment failed: %v", err)
			return util.ErrorResponse(c, fiber.StatusInternalServerError, err.Error())
		}
	}

	return util.SuccessResponse(c, fiber.StatusCreated, "Attachment added successfully", attachment)
}

// ListAttachments handles GET /api/v1/cards/:cardId/attachments
func (h *CardExtrasHandler) ListAttachments(c *fiber.Ctx) error {
	userID, ok := c.Locals("userID").(uuid.UUID)
	if !ok {
		return util.ErrorResponse(c, fiber.StatusUnauthorized, "Invalid user session")
	}

	cardID, err := uuid.Parse(c.Params("cardId"))
	if err != nil {
		return util.ErrorResponse(c, fiber.StatusBadRequest, "Invalid card ID")
	}

	attachments, err := h.extrasService.ListAttachments(c.Context(), userID, cardID)
	if err != nil {
		switch err.Error() {
		case "card not found", "list not found", "board not found":
			return util.ErrorResponse(c, fiber.StatusNotFound, err.Error())
		case "you do not have access to this card":
			return util.ErrorResponse(c, fiber.StatusForbidden, err.Error())
		default:
			return util.ErrorResponse(c, fiber.StatusInternalServerError, err.Error())
		}
	}

	return util.SuccessResponse(c, fiber.StatusOK, "", attachments)
}

// DeleteAttachment handles DELETE /api/v1/attachments/:id
func (h *CardExtrasHandler) DeleteAttachment(c *fiber.Ctx) error {
	userID, ok := c.Locals("userID").(uuid.UUID)
	if !ok {
		return util.ErrorResponse(c, fiber.StatusUnauthorized, "Invalid user session")
	}

	attachmentID, err := uuid.Parse(c.Params("id"))
	if err != nil {
		return util.ErrorResponse(c, fiber.StatusBadRequest, "Invalid attachment ID")
	}

	if err := h.extrasService.DeleteAttachment(c.Context(), userID, attachmentID); err != nil {
		switch err.Error() {
		case "attachment not found", "card not found", "list not found", "board not found":
			return util.ErrorResponse(c, fiber.StatusNotFound, err.Error())
		case "you do not have access to this card", "you can only delete your own attachments":
			return util.ErrorResponse(c, fiber.StatusForbidden, err.Error())
		default:
			return util.ErrorResponse(c, fiber.StatusInternalServerError, err.Error())
		}
	}

	return util.SuccessResponse(c, fiber.StatusOK, "Attachment deleted successfully", nil)
}

func formatCardExtrasValidationError(errs validator.ValidationErrors) string {
	if len(errs) == 0 {
		return "Validation failed"
	}

	fieldMessages := map[string]map[string]string{
		"CardID": {
			"required": "Card ID is required",
			"uuid":     "Card ID must be a valid UUID",
		},
		"Content": {
			"required": "Comment content is required",
			"min":      "Comment must not be empty",
			"max":      "Comment must not exceed 5000 characters",
		},
		"FileName": {
			"required": "File name is required",
			"min":      "File name must not be empty",
			"max":      "File name must not exceed 255 characters",
		},
		"FileURL": {
			"required": "File URL is required",
			"url":      "File URL must be a valid URL",
		},
		"FileType": {
			"max": "File type must not exceed 100 characters",
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
