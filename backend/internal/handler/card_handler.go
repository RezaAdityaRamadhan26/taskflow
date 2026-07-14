// Package handler implements HTTP handlers for card endpoints.
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

// CardHandler handles HTTP requests for card-related endpoints.
type CardHandler struct {
	cardService *service.CardService
	validate    *validator.Validate
}

// NewCardHandler creates a new CardHandler instance.
func NewCardHandler(cardService *service.CardService) *CardHandler {
	return &CardHandler{
		cardService: cardService,
		validate:    validator.New(),
	}
}

// Create handles POST /api/v1/cards
func (h *CardHandler) Create(c *fiber.Ctx) error {
	userID, ok := c.Locals("userID").(uuid.UUID)
	if !ok {
		return util.ErrorResponse(c, fiber.StatusUnauthorized, "Invalid user session")
	}

	var req dto.CreateCardRequest
	if err := c.BodyParser(&req); err != nil {
		return util.ErrorResponse(c, fiber.StatusBadRequest, "Invalid request body")
	}

	if err := h.validate.Struct(req); err != nil {
		validationErrors := err.(validator.ValidationErrors)
		return util.ValidationErrorResponse(c, formatCardValidationError(validationErrors))
	}

	card, err := h.cardService.CreateCard(c.Context(), userID, req)
	if err != nil {
		switch err.Error() {
		case "invalid list ID", "list not found", "board not found":
			return util.ErrorResponse(c, fiber.StatusBadRequest, err.Error())
		case "you do not have access to this board":
			return util.ErrorResponse(c, fiber.StatusForbidden, err.Error())
		default:
			log.Printf("[ERROR] Create card failed: %v", err)
			return util.ErrorResponse(c, fiber.StatusInternalServerError, err.Error())
		}
	}

	return util.SuccessResponse(c, fiber.StatusCreated, "Card created successfully", card)
}

// List handles GET /api/v1/lists/:listId/cards
func (h *CardHandler) List(c *fiber.Ctx) error {
	userID, ok := c.Locals("userID").(uuid.UUID)
	if !ok {
		return util.ErrorResponse(c, fiber.StatusUnauthorized, "Invalid user session")
	}

	listID, err := uuid.Parse(c.Params("listId"))
	if err != nil {
		return util.ErrorResponse(c, fiber.StatusBadRequest, "Invalid list ID")
	}

	cards, err := h.cardService.ListCards(c.Context(), userID, listID)
	if err != nil {
		switch err.Error() {
		case "list not found", "board not found":
			return util.ErrorResponse(c, fiber.StatusNotFound, err.Error())
		case "you do not have access to this board":
			return util.ErrorResponse(c, fiber.StatusForbidden, err.Error())
		default:
			return util.ErrorResponse(c, fiber.StatusInternalServerError, err.Error())
		}
	}

	return util.SuccessResponse(c, fiber.StatusOK, "", cards)
}

// Get handles GET /api/v1/cards/:id
func (h *CardHandler) Get(c *fiber.Ctx) error {
	userID, ok := c.Locals("userID").(uuid.UUID)
	if !ok {
		return util.ErrorResponse(c, fiber.StatusUnauthorized, "Invalid user session")
	}

	cardID, err := uuid.Parse(c.Params("id"))
	if err != nil {
		return util.ErrorResponse(c, fiber.StatusBadRequest, "Invalid card ID")
	}

	card, err := h.cardService.GetCard(c.Context(), userID, cardID)
	if err != nil {
		switch err.Error() {
		case "card not found", "list not found", "board not found":
			return util.ErrorResponse(c, fiber.StatusNotFound, err.Error())
		case "you do not have access to this board":
			return util.ErrorResponse(c, fiber.StatusForbidden, err.Error())
		default:
			return util.ErrorResponse(c, fiber.StatusInternalServerError, err.Error())
		}
	}

	return util.SuccessResponse(c, fiber.StatusOK, "", card)
}

// Update handles PUT /api/v1/cards/:id
func (h *CardHandler) Update(c *fiber.Ctx) error {
	userID, ok := c.Locals("userID").(uuid.UUID)
	if !ok {
		return util.ErrorResponse(c, fiber.StatusUnauthorized, "Invalid user session")
	}

	cardID, err := uuid.Parse(c.Params("id"))
	if err != nil {
		return util.ErrorResponse(c, fiber.StatusBadRequest, "Invalid card ID")
	}

	var req dto.UpdateCardRequest
	if err := c.BodyParser(&req); err != nil {
		return util.ErrorResponse(c, fiber.StatusBadRequest, "Invalid request body")
	}

	if err := h.validate.Struct(req); err != nil {
		validationErrors := err.(validator.ValidationErrors)
		return util.ValidationErrorResponse(c, formatCardValidationError(validationErrors))
	}

	card, err := h.cardService.UpdateCard(c.Context(), userID, cardID, req)
	if err != nil {
		switch err.Error() {
		case "card not found", "list not found", "board not found", "invalid list ID":
			return util.ErrorResponse(c, fiber.StatusNotFound, err.Error())
		case "you do not have access to this board", "target list not found or access denied":
			return util.ErrorResponse(c, fiber.StatusForbidden, err.Error())
		default:
			return util.ErrorResponse(c, fiber.StatusInternalServerError, err.Error())
		}
	}

	return util.SuccessResponse(c, fiber.StatusOK, "Card updated successfully", card)
}

// Delete handles DELETE /api/v1/cards/:id
func (h *CardHandler) Delete(c *fiber.Ctx) error {
	userID, ok := c.Locals("userID").(uuid.UUID)
	if !ok {
		return util.ErrorResponse(c, fiber.StatusUnauthorized, "Invalid user session")
	}

	cardID, err := uuid.Parse(c.Params("id"))
	if err != nil {
		return util.ErrorResponse(c, fiber.StatusBadRequest, "Invalid card ID")
	}

	if err := h.cardService.DeleteCard(c.Context(), userID, cardID); err != nil {
		switch err.Error() {
		case "card not found", "list not found", "board not found":
			return util.ErrorResponse(c, fiber.StatusNotFound, err.Error())
		case "you do not have access to this board":
			return util.ErrorResponse(c, fiber.StatusForbidden, err.Error())
		default:
			return util.ErrorResponse(c, fiber.StatusInternalServerError, err.Error())
		}
	}

	return util.SuccessResponse(c, fiber.StatusOK, "Card deleted successfully", nil)
}

func formatCardValidationError(errs validator.ValidationErrors) string {
	if len(errs) == 0 {
		return "Validation failed"
	}

	fieldMessages := map[string]map[string]string{
		"ListID": {
			"required": "List ID is required",
			"uuid":     "List ID must be a valid UUID",
		},
		"Title": {
			"required": "Title is required",
			"min":      "Title must be at least 1 character",
			"max":      "Title must not exceed 255 characters",
		},
		"Description": {
			"max": "Description must not exceed 5000 characters",
		},
		"Position": {
			"required": "Position is required",
		},
		"Priority": {
			"required": "Priority is required",
			"oneof":    "Priority must be LOW, MEDIUM, HIGH, or URGENT",
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
