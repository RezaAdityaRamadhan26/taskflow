// Package handler implements HTTP handlers for authentication endpoints.
package handler

import (
	"log"
	"time"

	"github.com/go-playground/validator/v10"
	"github.com/gofiber/fiber/v2"
	"github.com/google/uuid"

	"github.com/RezaAdityaRamadhan26/taskflow/backend/internal/config"
	"github.com/RezaAdityaRamadhan26/taskflow/backend/internal/dto"
	"github.com/RezaAdityaRamadhan26/taskflow/backend/internal/service"
	"github.com/RezaAdityaRamadhan26/taskflow/backend/internal/util"
)

// AuthHandler handles HTTP requests for auth-related endpoints.
type AuthHandler struct {
	authService *service.AuthService
	validate    *validator.Validate
	config      *config.Config
}

// NewAuthHandler creates a new AuthHandler instance.
func NewAuthHandler(authService *service.AuthService, cfg *config.Config) *AuthHandler {
	return &AuthHandler{
		authService: authService,
		validate:    validator.New(),
		config:      cfg,
	}
}

// Register handles POST /api/v1/auth/register
// Creates a new user account and returns JWT tokens.
func (h *AuthHandler) Register(c *fiber.Ctx) error {
	var req dto.RegisterRequest
	if err := c.BodyParser(&req); err != nil {
		return util.ErrorResponse(c, fiber.StatusBadRequest, "Invalid request body")
	}

	// Validate input
	if err := h.validate.Struct(req); err != nil {
		validationErrors := err.(validator.ValidationErrors)
		return util.ValidationErrorResponse(c, formatValidationError(validationErrors))
	}

	// Call service
	response, refreshToken, err := h.authService.Register(c.Context(), req)
	if err != nil {
		// Differentiate between "email already exists" and internal errors
		if err.Error() == "email already registered" {
			return util.ErrorResponse(c, fiber.StatusConflict, err.Error())
		}
		log.Printf("[ERROR] Register failed: %v", err)
		return util.ErrorResponse(c, fiber.StatusInternalServerError, "Registration failed")
	}

	// Set refresh token as httpOnly cookie
	h.setRefreshTokenCookie(c, refreshToken)

	return util.SuccessResponse(c, fiber.StatusCreated, "User registered successfully", response)
}

// Login handles POST /api/v1/auth/login
// Authenticates user credentials and returns JWT tokens.
func (h *AuthHandler) Login(c *fiber.Ctx) error {
	var req dto.LoginRequest
	if err := c.BodyParser(&req); err != nil {
		return util.ErrorResponse(c, fiber.StatusBadRequest, "Invalid request body")
	}

	// Validate input
	if err := h.validate.Struct(req); err != nil {
		validationErrors := err.(validator.ValidationErrors)
		return util.ValidationErrorResponse(c, formatValidationError(validationErrors))
	}

	// Call service
	response, refreshToken, err := h.authService.Login(c.Context(), req)
	if err != nil {
		return util.ErrorResponse(c, fiber.StatusUnauthorized, "Invalid email or password")
	}

	// Set refresh token as httpOnly cookie
	h.setRefreshTokenCookie(c, refreshToken)

	return util.SuccessResponse(c, fiber.StatusOK, "Login successful", response)
}

// Refresh handles POST /api/v1/auth/refresh
// Validates refresh token and issues new access + refresh tokens.
func (h *AuthHandler) Refresh(c *fiber.Ctx) error {
	// Get refresh token from cookie
	refreshToken := c.Cookies("refresh_token")
	if refreshToken == "" {
		return util.ErrorResponse(c, fiber.StatusUnauthorized, "Refresh token not found")
	}

	// Call service
	response, newRefreshToken, err := h.authService.RefreshAccessToken(c.Context(), refreshToken)
	if err != nil {
		// Clear invalid cookie
		h.clearRefreshTokenCookie(c)
		return util.ErrorResponse(c, fiber.StatusUnauthorized, err.Error())
	}

	// Set new refresh token cookie (rotation)
	h.setRefreshTokenCookie(c, newRefreshToken)

	return util.SuccessResponse(c, fiber.StatusOK, "Token refreshed successfully", response)
}

// Logout handles POST /api/v1/auth/logout
// Invalidates all refresh tokens for the authenticated user.
func (h *AuthHandler) Logout(c *fiber.Ctx) error {
	userID, ok := c.Locals("userID").(uuid.UUID)
	if !ok {
		return util.ErrorResponse(c, fiber.StatusUnauthorized, "Invalid user session")
	}

	if err := h.authService.Logout(c.Context(), userID); err != nil {
		log.Printf("[ERROR] Logout failed for user %s: %v", userID, err)
		return util.ErrorResponse(c, fiber.StatusInternalServerError, "Logout failed")
	}

	// Clear refresh token cookie
	h.clearRefreshTokenCookie(c)

	return util.SuccessResponse(c, fiber.StatusOK, "Logged out successfully", nil)
}

// Me handles GET /api/v1/auth/me
// Returns the currently authenticated user's profile.
func (h *AuthHandler) Me(c *fiber.Ctx) error {
	userID, ok := c.Locals("userID").(uuid.UUID)
	if !ok {
		return util.ErrorResponse(c, fiber.StatusUnauthorized, "Invalid user session")
	}

	user, err := h.authService.GetCurrentUser(c.Context(), userID)
	if err != nil {
		return util.ErrorResponse(c, fiber.StatusNotFound, "User not found")
	}

	return util.SuccessResponse(c, fiber.StatusOK, "", user)
}

// setRefreshTokenCookie sets the refresh token as an httpOnly, Secure, SameSite cookie.
func (h *AuthHandler) setRefreshTokenCookie(c *fiber.Ctx, token string) {
	c.Cookie(&fiber.Cookie{
		Name:     "refresh_token",
		Value:    token,
		Path:     "/api/v1/auth",
		Expires:  time.Now().Add(h.config.JWTRefreshExpiry),
		HTTPOnly: true,
		Secure:   !h.config.IsDevelopment(), // false in dev for HTTP testing
		SameSite: "Lax",
	})
}

// clearRefreshTokenCookie removes the refresh token cookie.
func (h *AuthHandler) clearRefreshTokenCookie(c *fiber.Ctx) {
	c.Cookie(&fiber.Cookie{
		Name:     "refresh_token",
		Value:    "",
		Path:     "/api/v1/auth",
		Expires:  time.Now().Add(-1 * time.Hour),
		HTTPOnly: true,
		Secure:   !h.config.IsDevelopment(),
		SameSite: "Lax",
	})
}

// formatValidationError converts validator errors into a human-readable message.
func formatValidationError(errs validator.ValidationErrors) string {
	if len(errs) == 0 {
		return "Validation failed"
	}

	fieldMessages := map[string]map[string]string{
		"Email": {
			"required": "Email is required",
			"email":    "Email must be a valid email address",
			"max":      "Email must not exceed 255 characters",
		},
		"Password": {
			"required": "Password is required",
			"min":      "Password must be at least 8 characters",
			"max":      "Password must not exceed 128 characters",
		},
		"Name": {
			"required": "Name is required",
			"min":      "Name must be at least 2 characters",
			"max":      "Name must not exceed 100 characters",
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
