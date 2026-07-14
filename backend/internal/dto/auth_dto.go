// Package dto defines request and response data transfer objects for the auth module.
package dto

// RegisterRequest represents the request body for user registration.
type RegisterRequest struct {
	Email    string `json:"email" validate:"required,email,max=255"`
	Password string `json:"password" validate:"required,min=8,max=128"`
	Name     string `json:"name" validate:"required,min=2,max=100"`
}

// LoginRequest represents the request body for user login.
type LoginRequest struct {
	Email    string `json:"email" validate:"required,email"`
	Password string `json:"password" validate:"required"`
}

// AuthResponse represents the response body after successful authentication.
type AuthResponse struct {
	User        UserDTO `json:"user"`
	AccessToken string  `json:"access_token"`
}

// UserDTO is a safe user representation for API responses (no password hash).
type UserDTO struct {
	ID        string  `json:"id"`
	Email     string  `json:"email"`
	Name      string  `json:"name"`
	AvatarURL *string `json:"avatar_url,omitempty"`
	CreatedAt string  `json:"created_at"`
	UpdatedAt string  `json:"updated_at"`
}

// TokenResponse represents the response body for token refresh.
type TokenResponse struct {
	AccessToken string `json:"access_token"`
}
