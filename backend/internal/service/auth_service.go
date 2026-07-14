// Package service implements the business logic for authentication.
package service

import (
	"context"
	"database/sql"
	"fmt"
	"log"
	"time"

	"github.com/google/uuid"

	"github.com/RezaAdityaRamadhan26/taskflow/backend/internal/config"
	"github.com/RezaAdityaRamadhan26/taskflow/backend/internal/dto"
	"github.com/RezaAdityaRamadhan26/taskflow/backend/internal/repository"
	"github.com/RezaAdityaRamadhan26/taskflow/backend/internal/util"
)

// AuthService handles authentication business logic.
type AuthService struct {
	queries *repository.Queries
	config  *config.Config
}

// NewAuthService creates a new AuthService instance.
func NewAuthService(queries *repository.Queries, cfg *config.Config) *AuthService {
	return &AuthService{
		queries: queries,
		config:  cfg,
	}
}

// Register creates a new user account and returns auth tokens.
func (s *AuthService) Register(ctx context.Context, req dto.RegisterRequest) (*dto.AuthResponse, string, error) {
	// Check if user already exists
	_, err := s.queries.GetUserByEmail(ctx, req.Email)
	if err == nil {
		return nil, "", fmt.Errorf("email already registered")
	}

	// Hash password
	hashedPassword, err := util.HashPassword(req.Password)
	if err != nil {
		log.Printf("[ERROR] Failed to hash password: %v", err)
		return nil, "", fmt.Errorf("internal server error")
	}

	// Create user
	user, err := s.queries.CreateUser(ctx, repository.CreateUserParams{
		ID:           uuid.New(),
		Email:        req.Email,
		PasswordHash: hashedPassword,
		Name:         req.Name,
		AvatarUrl:    sql.NullString{Valid: false},
	})
	if err != nil {
		log.Printf("[ERROR] Failed to create user: %v", err)
		return nil, "", fmt.Errorf("failed to create user")
	}

	// Generate access token
	accessToken, err := util.GenerateAccessToken(
		user.ID, user.Email,
		s.config.JWTAccessSecret,
		s.config.JWTAccessExpiry,
	)
	if err != nil {
		log.Printf("[ERROR] Failed to generate access token: %v", err)
		return nil, "", fmt.Errorf("failed to generate token")
	}

	// Generate and store refresh token
	refreshTokenStr := util.GenerateRefreshToken()
	_, err = s.queries.CreateRefreshToken(ctx, repository.CreateRefreshTokenParams{
		ID:        uuid.New(),
		Token:     refreshTokenStr,
		UserID:    user.ID,
		ExpiresAt: time.Now().Add(s.config.JWTRefreshExpiry),
	})
	if err != nil {
		log.Printf("[ERROR] Failed to store refresh token: %v", err)
		return nil, "", fmt.Errorf("failed to create session")
	}

	// Build response
	var avatarURL *string
	if user.AvatarUrl.Valid {
		avatarURL = &user.AvatarUrl.String
	}

	response := &dto.AuthResponse{
		User: dto.UserDTO{
			ID:        user.ID.String(),
			Email:     user.Email,
			Name:      user.Name,
			AvatarURL: avatarURL,
			CreatedAt: user.CreatedAt.Format(time.RFC3339),
			UpdatedAt: user.UpdatedAt.Format(time.RFC3339),
		},
		AccessToken: accessToken,
	}

	return response, refreshTokenStr, nil
}

// Login authenticates a user and returns auth tokens.
func (s *AuthService) Login(ctx context.Context, req dto.LoginRequest) (*dto.AuthResponse, string, error) {
	// Get user by email
	user, err := s.queries.GetUserByEmail(ctx, req.Email)
	if err != nil {
		return nil, "", fmt.Errorf("invalid email or password")
	}

	// Check password
	if err := util.CheckPassword(req.Password, user.PasswordHash); err != nil {
		return nil, "", fmt.Errorf("invalid email or password")
	}

	// Generate access token
	accessToken, err := util.GenerateAccessToken(
		user.ID, user.Email,
		s.config.JWTAccessSecret,
		s.config.JWTAccessExpiry,
	)
	if err != nil {
		log.Printf("[ERROR] Failed to generate access token: %v", err)
		return nil, "", fmt.Errorf("failed to generate token")
	}

	// Generate and store refresh token
	refreshTokenStr := util.GenerateRefreshToken()
	_, err = s.queries.CreateRefreshToken(ctx, repository.CreateRefreshTokenParams{
		ID:        uuid.New(),
		Token:     refreshTokenStr,
		UserID:    user.ID,
		ExpiresAt: time.Now().Add(s.config.JWTRefreshExpiry),
	})
	if err != nil {
		log.Printf("[ERROR] Failed to store refresh token: %v", err)
		return nil, "", fmt.Errorf("failed to create session")
	}

	// Build response
	var avatarURL *string
	if user.AvatarUrl.Valid {
		avatarURL = &user.AvatarUrl.String
	}

	response := &dto.AuthResponse{
		User: dto.UserDTO{
			ID:        user.ID.String(),
			Email:     user.Email,
			Name:      user.Name,
			AvatarURL: avatarURL,
			CreatedAt: user.CreatedAt.Format(time.RFC3339),
			UpdatedAt: user.UpdatedAt.Format(time.RFC3339),
		},
		AccessToken: accessToken,
	}

	return response, refreshTokenStr, nil
}

// RefreshAccessToken validates a refresh token and issues a new access token.
func (s *AuthService) RefreshAccessToken(ctx context.Context, refreshTokenStr string) (*dto.TokenResponse, string, error) {
	if refreshTokenStr == "" {
		return nil, "", fmt.Errorf("refresh token is required")
	}

	// Look up the refresh token
	storedToken, err := s.queries.GetRefreshToken(ctx, refreshTokenStr)
	if err != nil {
		return nil, "", fmt.Errorf("invalid refresh token")
	}

	// Check expiration
	if time.Now().After(storedToken.ExpiresAt) {
		// Delete expired token
		_ = s.queries.DeleteRefreshToken(ctx, refreshTokenStr)
		return nil, "", fmt.Errorf("refresh token has expired")
	}

	// Get user
	user, err := s.queries.GetUserByID(ctx, storedToken.UserID)
	if err != nil {
		return nil, "", fmt.Errorf("user not found")
	}

	// Delete old refresh token (rotate)
	_ = s.queries.DeleteRefreshToken(ctx, refreshTokenStr)

	// Generate new access token
	accessToken, err := util.GenerateAccessToken(
		user.ID, user.Email,
		s.config.JWTAccessSecret,
		s.config.JWTAccessExpiry,
	)
	if err != nil {
		log.Printf("[ERROR] Failed to generate access token: %v", err)
		return nil, "", fmt.Errorf("failed to generate token")
	}

	// Generate new refresh token (rotation)
	newRefreshTokenStr := util.GenerateRefreshToken()
	_, err = s.queries.CreateRefreshToken(ctx, repository.CreateRefreshTokenParams{
		ID:        uuid.New(),
		Token:     newRefreshTokenStr,
		UserID:    user.ID,
		ExpiresAt: time.Now().Add(s.config.JWTRefreshExpiry),
	})
	if err != nil {
		log.Printf("[ERROR] Failed to store new refresh token: %v", err)
		return nil, "", fmt.Errorf("failed to create session")
	}

	response := &dto.TokenResponse{
		AccessToken: accessToken,
	}

	return response, newRefreshTokenStr, nil
}

// Logout invalidates all refresh tokens for the given user.
func (s *AuthService) Logout(ctx context.Context, userID uuid.UUID) error {
	err := s.queries.DeleteRefreshTokensByUserID(ctx, userID)
	if err != nil {
		log.Printf("[ERROR] Failed to delete refresh tokens: %v", err)
		return fmt.Errorf("failed to logout")
	}
	return nil
}

// GetCurrentUser retrieves the user profile by ID.
func (s *AuthService) GetCurrentUser(ctx context.Context, userID uuid.UUID) (*dto.UserDTO, error) {
	user, err := s.queries.GetUserByID(ctx, userID)
	if err != nil {
		return nil, fmt.Errorf("user not found")
	}

	var avatarURL *string
	if user.AvatarUrl.Valid {
		avatarURL = &user.AvatarUrl.String
	}

	return &dto.UserDTO{
		ID:        user.ID.String(),
		Email:     user.Email,
		Name:      user.Name,
		AvatarURL: avatarURL,
		CreatedAt: user.CreatedAt.Format(time.RFC3339),
		UpdatedAt: user.UpdatedAt.Format(time.RFC3339),
	}, nil
}
