// Package middleware provides HTTP middleware for the Fiber application.
package middleware

import (
	"strings"

	"github.com/gofiber/fiber/v2"

	"github.com/RezaAdityaRamadhan26/taskflow/backend/internal/util"
)

// AuthMiddleware validates the JWT access token from the Authorization header.
// On success, it stores the user ID and email in the Fiber context locals.
func AuthMiddleware(jwtSecret string) fiber.Handler {
	return func(c *fiber.Ctx) error {
		// Extract token from Authorization header
		authHeader := c.Get("Authorization")
		if authHeader == "" {
			return util.ErrorResponse(c, fiber.StatusUnauthorized, "Authorization header is required")
		}

		// Expect format: "Bearer <token>"
		parts := strings.SplitN(authHeader, " ", 2)
		if len(parts) != 2 || !strings.EqualFold(parts[0], "Bearer") {
			return util.ErrorResponse(c, fiber.StatusUnauthorized, "Invalid authorization format. Use: Bearer <token>")
		}

		tokenString := parts[1]
		if tokenString == "" {
			return util.ErrorResponse(c, fiber.StatusUnauthorized, "Token is required")
		}

		// Validate the token
		claims, err := util.ValidateAccessToken(tokenString, jwtSecret)
		if err != nil {
			return util.ErrorResponse(c, fiber.StatusUnauthorized, "Invalid or expired token")
		}

		// Store user info in context for downstream handlers
		c.Locals("userID", claims.UserID)
		c.Locals("userEmail", claims.Email)

		return c.Next()
	}
}
