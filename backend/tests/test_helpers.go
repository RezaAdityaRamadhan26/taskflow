// Package tests provides shared test utilities for integration tests.
package tests

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"math/rand"
	"net/http"
	"net/http/httptest"
	"regexp"
	"testing"
	"time"

	"github.com/go-playground/validator/v10"
	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/middleware/cors"
	"github.com/gofiber/fiber/v2/middleware/recover"

	"github.com/RezaAdityaRamadhan26/taskflow/backend/internal/config"
	"github.com/RezaAdityaRamadhan26/taskflow/backend/internal/database"
	"github.com/RezaAdityaRamadhan26/taskflow/backend/internal/handler"
	"github.com/RezaAdityaRamadhan26/taskflow/backend/internal/middleware"
	"github.com/RezaAdityaRamadhan26/taskflow/backend/internal/repository"
	"github.com/RezaAdityaRamadhan26/taskflow/backend/internal/service"
)

// TestApp holds the Fiber app instance and configuration for tests.
type TestApp struct {
	App *fiber.App
	Cfg *config.Config
}

// APIResponse is a generic API response struct for unmarshaling test results.
type APIResponse struct {
	Success bool            `json:"success"`
	Message string          `json:"message"`
	Data    json.RawMessage `json:"data"`
	Error   string          `json:"error"`
}

// AuthResponseData represents the auth response data for register/login.
type AuthResponseData struct {
	User        UserDTO `json:"user"`
	AccessToken string  `json:"access_token"`
}

// TokenResponseData represents the token refresh response.
type TokenResponseData struct {
	AccessToken string `json:"access_token"`
}

// UserDTO matches the API user response.
type UserDTO struct {
	ID        string  `json:"id"`
	Email     string  `json:"email"`
	Name      string  `json:"name"`
	AvatarURL *string `json:"avatar_url,omitempty"`
	CreatedAt string  `json:"created_at"`
	UpdatedAt string  `json:"updated_at"`
}

// WorkspaceDTO matches the API workspace response.
type WorkspaceDTO struct {
	ID          string  `json:"id"`
	Name        string  `json:"name"`
	Description *string `json:"description,omitempty"`
	Slug        string  `json:"slug"`
	Role        string  `json:"role,omitempty"`
	CreatedAt   string  `json:"created_at"`
	UpdatedAt   string  `json:"updated_at"`
}

// WorkspaceMemberDTO matches the API workspace member response.
type WorkspaceMemberDTO struct {
	ID        string  `json:"id"`
	UserID    string  `json:"user_id"`
	Name      string  `json:"name"`
	Email     string  `json:"email"`
	AvatarURL *string `json:"avatar_url,omitempty"`
	Role      string  `json:"role"`
	JoinedAt  string  `json:"joined_at"`
}

// SetupTestApp creates a Fiber app with all routes for testing.
func SetupTestApp(t *testing.T) *TestApp {
	t.Helper()

	cfg, err := config.Load()
	if err != nil {
		t.Fatalf("Failed to load config: %v", err)
	}

	pool, err := database.Connect(cfg.DatabaseURL)
	if err != nil {
		t.Fatalf("Failed to connect to database: %v", err)
	}
	t.Cleanup(func() { pool.Close() })

	queries := repository.New(pool)
	authService := service.NewAuthService(queries, cfg)
	workspaceService := service.NewWorkspaceService(queries)
	authHandler := handler.NewAuthHandler(authService, cfg)
	workspaceHandler := handler.NewWorkspaceHandler(workspaceService)

	// Register custom slug validation for workspace handler
	v := validator.New()
	v.RegisterValidation("alphanum_with_dash", func(fl validator.FieldLevel) bool {
		return regexp.MustCompile(`^[a-z0-9]+(?:-[a-z0-9]+)*$`).MatchString(fl.Field().String())
	})

	app := fiber.New(fiber.Config{
		ErrorHandler: func(c *fiber.Ctx, err error) error {
			code := fiber.StatusInternalServerError
			if e, ok := err.(*fiber.Error); ok {
				code = e.Code
			}
			return c.Status(code).JSON(fiber.Map{"success": false, "error": err.Error()})
		},
	})
	app.Use(recover.New())
	app.Use(cors.New(cors.Config{
		AllowOrigins:     "http://localhost:5173",
		AllowCredentials: true,
	}))

	api := app.Group("/api/v1")

	// Auth routes
	auth := api.Group("/auth")
	auth.Post("/register", authHandler.Register)
	auth.Post("/login", authHandler.Login)
	auth.Post("/refresh", authHandler.Refresh)

	authProtected := auth.Group("", middleware.AuthMiddleware(cfg.JWTAccessSecret))
	authProtected.Post("/logout", authHandler.Logout)
	authProtected.Get("/me", authHandler.Me)

	// Workspace routes
	workspaces := api.Group("/workspaces", middleware.AuthMiddleware(cfg.JWTAccessSecret))
	workspaces.Post("/", workspaceHandler.Create)
	workspaces.Get("/", workspaceHandler.List)
	workspaces.Get("/:id", workspaceHandler.Get)
	workspaces.Put("/:id", workspaceHandler.Update)
	workspaces.Delete("/:id", workspaceHandler.Delete)
	workspaces.Post("/:id/members", workspaceHandler.InviteMember)
	workspaces.Get("/:id/members", workspaceHandler.ListMembers)
	workspaces.Patch("/:id/members/:userId", workspaceHandler.UpdateMemberRole)
	workspaces.Delete("/:id/members/:userId", workspaceHandler.RemoveMember)

	return &TestApp{App: app, Cfg: cfg}
}

// MakeRequest is a helper to create and execute HTTP requests against the test app.
func (ta *TestApp) MakeRequest(t *testing.T, method, path string, body interface{}, headers map[string]string) (*http.Response, APIResponse) {
	t.Helper()

	var reqBody io.Reader
	if body != nil {
		jsonBody, err := json.Marshal(body)
		if err != nil {
			t.Fatalf("Failed to marshal request body: %v", err)
		}
		reqBody = bytes.NewReader(jsonBody)
	}

	req := httptest.NewRequest(method, path, reqBody)
	req.Header.Set("Content-Type", "application/json")
	for key, val := range headers {
		req.Header.Set(key, val)
	}

	resp, err := ta.App.Test(req, -1)
	if err != nil {
		t.Fatalf("Failed to execute request: %v", err)
	}

	respBody, err := io.ReadAll(resp.Body)
	if err != nil {
		t.Fatalf("Failed to read response body: %v", err)
	}
	defer resp.Body.Close()

	var apiResp APIResponse
	if err := json.Unmarshal(respBody, &apiResp); err != nil {
		t.Fatalf("Failed to unmarshal response: %v\nBody: %s", err, string(respBody))
	}

	return resp, apiResp
}

// RegisterAndLogin registers a user and returns the access token.
func (ta *TestApp) RegisterAndLogin(t *testing.T, email, password, name string) (string, AuthResponseData) {
	t.Helper()

	_, apiResp := ta.MakeRequest(t, "POST", "/api/v1/auth/register", map[string]string{
		"email":    email,
		"password": password,
		"name":     name,
	}, nil)

	if !apiResp.Success {
		t.Fatalf("Failed to register user %s: %s", email, apiResp.Error)
	}

	var data AuthResponseData
	if err := json.Unmarshal(apiResp.Data, &data); err != nil {
		t.Fatalf("Failed to parse auth response: %v", err)
	}

	return data.AccessToken, data
}

// AuthHeader returns a map with the Authorization header set.
func AuthHeader(token string) map[string]string {
	return map[string]string{"Authorization": "Bearer " + token}
}

// UniqueEmail generates a unique email for each test to avoid conflicts.
func UniqueEmail() string {
	return fmt.Sprintf("test_%d_%d@test.com", time.Now().UnixNano(), rand.Intn(999999))
}

// UniqueSlug generates a unique slug for each test to avoid conflicts.
func UniqueSlug() string {
	return fmt.Sprintf("test-ws-%d-%d", time.Now().UnixNano()%100000, rand.Intn(9999))
}
