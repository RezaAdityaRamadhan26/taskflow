// Package tests provides integration tests for the TaskFlow API.
// These tests run against a real database and test the full HTTP request/response cycle.
package tests

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"math/rand"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

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

// testApp holds the Fiber app instance and configuration for tests.
type testApp struct {
	app *fiber.App
	cfg *config.Config
}

// apiResponse is a generic API response struct for unmarshaling test results.
type apiResponse struct {
	Success bool            `json:"success"`
	Message string          `json:"message"`
	Data    json.RawMessage `json:"data"`
	Error   string          `json:"error"`
}

// authResponseData represents the auth response data for register/login.
type authResponseData struct {
	User        userDTO `json:"user"`
	AccessToken string  `json:"access_token"`
}

// tokenResponseData represents the token refresh response.
type tokenResponseData struct {
	AccessToken string `json:"access_token"`
}

// userDTO matches the API user response.
type userDTO struct {
	ID        string  `json:"id"`
	Email     string  `json:"email"`
	Name      string  `json:"name"`
	AvatarURL *string `json:"avatar_url,omitempty"`
	CreatedAt string  `json:"created_at"`
	UpdatedAt string  `json:"updated_at"`
}

// setupTestApp creates a Fiber app with all routes for testing.
func setupTestApp(t *testing.T) *testApp {
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
	authHandler := handler.NewAuthHandler(authService, cfg)

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
	auth := api.Group("/auth")
	auth.Post("/register", authHandler.Register)
	auth.Post("/login", authHandler.Login)
	auth.Post("/refresh", authHandler.Refresh)

	authProtected := auth.Group("", middleware.AuthMiddleware(cfg.JWTAccessSecret))
	authProtected.Post("/logout", authHandler.Logout)
	authProtected.Get("/me", authHandler.Me)

	return &testApp{app: app, cfg: cfg}
}

// makeRequest is a helper to create and execute HTTP requests against the test app.
func (ta *testApp) makeRequest(t *testing.T, method, path string, body interface{}, headers map[string]string) (*http.Response, apiResponse) {
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

	resp, err := ta.app.Test(req, -1)
	if err != nil {
		t.Fatalf("Failed to execute request: %v", err)
	}

	respBody, err := io.ReadAll(resp.Body)
	if err != nil {
		t.Fatalf("Failed to read response body: %v", err)
	}
	defer resp.Body.Close()

	var apiResp apiResponse
	if err := json.Unmarshal(respBody, &apiResp); err != nil {
		t.Fatalf("Failed to unmarshal response: %v\nBody: %s", err, string(respBody))
	}

	return resp, apiResp
}

// uniqueEmail generates a unique email for each test to avoid conflicts.
func uniqueEmail() string {
	return fmt.Sprintf("test_%d_%d@test.com", time.Now().UnixNano(), rand.Intn(999999))
}

// randomSuffix generates a random string suffix using timestamp nanoseconds.
func randomSuffix() string {
	return fmt.Sprintf("%d", rand.Intn(999999))
}

// ============================================
// TEST: Register
// ============================================

func TestRegister_Success(t *testing.T) {
	ta := setupTestApp(t)
	email := uniqueEmail()

	resp, apiResp := ta.makeRequest(t, "POST", "/api/v1/auth/register", map[string]string{
		"email":    email,
		"password": "TestPass123",
		"name":     "Test User",
	}, nil)

	if resp.StatusCode != fiber.StatusCreated {
		t.Errorf("Expected status 201, got %d. Error: %s", resp.StatusCode, apiResp.Error)
	}
	if !apiResp.Success {
		t.Errorf("Expected success=true, got false. Error: %s", apiResp.Error)
	}

	// Parse data
	var data authResponseData
	if err := json.Unmarshal(apiResp.Data, &data); err != nil {
		t.Fatalf("Failed to parse auth data: %v", err)
	}

	if data.AccessToken == "" {
		t.Error("Expected non-empty access_token")
	}
	if data.User.Email != email {
		t.Errorf("Expected email %s, got %s", email, data.User.Email)
	}
	if data.User.Name != "Test User" {
		t.Errorf("Expected name 'Test User', got '%s'", data.User.Name)
	}
	if data.User.ID == "" {
		t.Error("Expected non-empty user ID")
	}

	// Check refresh_token cookie
	cookies := resp.Cookies()
	found := false
	for _, c := range cookies {
		if c.Name == "refresh_token" {
			found = true
			if c.Value == "" {
				t.Error("refresh_token cookie is empty")
			}
			if !c.HttpOnly {
				t.Error("refresh_token cookie should be HttpOnly")
			}
		}
	}
	if !found {
		t.Error("Expected refresh_token cookie to be set")
	}
}

func TestRegister_DuplicateEmail(t *testing.T) {
	ta := setupTestApp(t)
	email := uniqueEmail()

	// Register first time
	ta.makeRequest(t, "POST", "/api/v1/auth/register", map[string]string{
		"email":    email,
		"password": "TestPass123",
		"name":     "Test User",
	}, nil)

	// Register again with same email
	resp, apiResp := ta.makeRequest(t, "POST", "/api/v1/auth/register", map[string]string{
		"email":    email,
		"password": "TestPass456",
		"name":     "Test User 2",
	}, nil)

	if resp.StatusCode != fiber.StatusConflict {
		t.Errorf("Expected status 409, got %d", resp.StatusCode)
	}
	if apiResp.Success {
		t.Error("Expected success=false for duplicate email")
	}
}

func TestRegister_InvalidEmail(t *testing.T) {
	ta := setupTestApp(t)

	resp, apiResp := ta.makeRequest(t, "POST", "/api/v1/auth/register", map[string]string{
		"email":    "not-an-email",
		"password": "TestPass123",
		"name":     "Test User",
	}, nil)

	if resp.StatusCode != fiber.StatusUnprocessableEntity {
		t.Errorf("Expected status 422, got %d", resp.StatusCode)
	}
	if apiResp.Success {
		t.Error("Expected success=false for invalid email")
	}
}

func TestRegister_ShortPassword(t *testing.T) {
	ta := setupTestApp(t)

	resp, apiResp := ta.makeRequest(t, "POST", "/api/v1/auth/register", map[string]string{
		"email":    uniqueEmail(),
		"password": "short",
		"name":     "Test User",
	}, nil)

	if resp.StatusCode != fiber.StatusUnprocessableEntity {
		t.Errorf("Expected status 422, got %d", resp.StatusCode)
	}
	if apiResp.Success {
		t.Error("Expected success=false for short password")
	}
}

func TestRegister_MissingFields(t *testing.T) {
	ta := setupTestApp(t)

	// Missing email
	resp, _ := ta.makeRequest(t, "POST", "/api/v1/auth/register", map[string]string{
		"password": "TestPass123",
		"name":     "Test User",
	}, nil)
	if resp.StatusCode != fiber.StatusUnprocessableEntity {
		t.Errorf("Missing email: expected 422, got %d", resp.StatusCode)
	}

	// Missing password
	resp, _ = ta.makeRequest(t, "POST", "/api/v1/auth/register", map[string]string{
		"email": uniqueEmail(),
		"name":  "Test User",
	}, nil)
	if resp.StatusCode != fiber.StatusUnprocessableEntity {
		t.Errorf("Missing password: expected 422, got %d", resp.StatusCode)
	}

	// Missing name
	resp, _ = ta.makeRequest(t, "POST", "/api/v1/auth/register", map[string]string{
		"email":    uniqueEmail(),
		"password": "TestPass123",
	}, nil)
	if resp.StatusCode != fiber.StatusUnprocessableEntity {
		t.Errorf("Missing name: expected 422, got %d", resp.StatusCode)
	}
}

func TestRegister_EmptyBody(t *testing.T) {
	ta := setupTestApp(t)

	resp, apiResp := ta.makeRequest(t, "POST", "/api/v1/auth/register", nil, nil)

	if resp.StatusCode != fiber.StatusUnprocessableEntity && resp.StatusCode != fiber.StatusBadRequest {
		t.Errorf("Expected status 422 or 400, got %d", resp.StatusCode)
	}
	if apiResp.Success {
		t.Error("Expected success=false for empty body")
	}
}

// ============================================
// TEST: Login
// ============================================

func TestLogin_Success(t *testing.T) {
	ta := setupTestApp(t)
	email := uniqueEmail()

	// Register first
	ta.makeRequest(t, "POST", "/api/v1/auth/register", map[string]string{
		"email":    email,
		"password": "TestPass123",
		"name":     "Test User",
	}, nil)

	// Login
	resp, apiResp := ta.makeRequest(t, "POST", "/api/v1/auth/login", map[string]string{
		"email":    email,
		"password": "TestPass123",
	}, nil)

	if resp.StatusCode != fiber.StatusOK {
		t.Errorf("Expected status 200, got %d. Error: %s", resp.StatusCode, apiResp.Error)
	}
	if !apiResp.Success {
		t.Errorf("Expected success=true. Error: %s", apiResp.Error)
	}

	var data authResponseData
	if err := json.Unmarshal(apiResp.Data, &data); err != nil {
		t.Fatalf("Failed to parse auth data: %v", err)
	}
	if data.AccessToken == "" {
		t.Error("Expected non-empty access_token")
	}
	if data.User.Email != email {
		t.Errorf("Expected email %s, got %s", email, data.User.Email)
	}
}

func TestLogin_WrongPassword(t *testing.T) {
	ta := setupTestApp(t)
	email := uniqueEmail()

	// Register
	ta.makeRequest(t, "POST", "/api/v1/auth/register", map[string]string{
		"email":    email,
		"password": "TestPass123",
		"name":     "Test User",
	}, nil)

	// Login with wrong password
	resp, apiResp := ta.makeRequest(t, "POST", "/api/v1/auth/login", map[string]string{
		"email":    email,
		"password": "WrongPass123",
	}, nil)

	if resp.StatusCode != fiber.StatusUnauthorized {
		t.Errorf("Expected status 401, got %d", resp.StatusCode)
	}
	if apiResp.Success {
		t.Error("Expected success=false for wrong password")
	}
}

func TestLogin_NonExistentEmail(t *testing.T) {
	ta := setupTestApp(t)

	resp, apiResp := ta.makeRequest(t, "POST", "/api/v1/auth/login", map[string]string{
		"email":    "nonexistent@test.com",
		"password": "TestPass123",
	}, nil)

	if resp.StatusCode != fiber.StatusUnauthorized {
		t.Errorf("Expected status 401, got %d", resp.StatusCode)
	}
	if apiResp.Success {
		t.Error("Expected success=false for non-existent email")
	}
}

func TestLogin_InvalidInput(t *testing.T) {
	ta := setupTestApp(t)

	// Missing password
	resp, _ := ta.makeRequest(t, "POST", "/api/v1/auth/login", map[string]string{
		"email": "test@test.com",
	}, nil)
	if resp.StatusCode != fiber.StatusUnprocessableEntity {
		t.Errorf("Missing password: expected 422, got %d", resp.StatusCode)
	}
}

// ============================================
// TEST: Me (Protected)
// ============================================

func TestMe_Success(t *testing.T) {
	ta := setupTestApp(t)
	email := uniqueEmail()

	// Register and get token
	_, apiResp := ta.makeRequest(t, "POST", "/api/v1/auth/register", map[string]string{
		"email":    email,
		"password": "TestPass123",
		"name":     "Test User Me",
	}, nil)

	var data authResponseData
	json.Unmarshal(apiResp.Data, &data)

	// Get profile
	resp, meResp := ta.makeRequest(t, "GET", "/api/v1/auth/me", nil, map[string]string{
		"Authorization": "Bearer " + data.AccessToken,
	})

	if resp.StatusCode != fiber.StatusOK {
		t.Errorf("Expected status 200, got %d. Error: %s", resp.StatusCode, meResp.Error)
	}
	if !meResp.Success {
		t.Errorf("Expected success=true. Error: %s", meResp.Error)
	}

	var user userDTO
	if err := json.Unmarshal(meResp.Data, &user); err != nil {
		t.Fatalf("Failed to parse user data: %v", err)
	}
	if user.Email != email {
		t.Errorf("Expected email %s, got %s", email, user.Email)
	}
	if user.Name != "Test User Me" {
		t.Errorf("Expected name 'Test User Me', got '%s'", user.Name)
	}
}

func TestMe_NoToken(t *testing.T) {
	ta := setupTestApp(t)

	resp, apiResp := ta.makeRequest(t, "GET", "/api/v1/auth/me", nil, nil)

	if resp.StatusCode != fiber.StatusUnauthorized {
		t.Errorf("Expected status 401, got %d", resp.StatusCode)
	}
	if apiResp.Success {
		t.Error("Expected success=false without token")
	}
}

func TestMe_InvalidToken(t *testing.T) {
	ta := setupTestApp(t)

	resp, apiResp := ta.makeRequest(t, "GET", "/api/v1/auth/me", nil, map[string]string{
		"Authorization": "Bearer invalid-token-here",
	})

	if resp.StatusCode != fiber.StatusUnauthorized {
		t.Errorf("Expected status 401, got %d", resp.StatusCode)
	}
	if apiResp.Success {
		t.Error("Expected success=false for invalid token")
	}
}

func TestMe_MalformedAuthHeader(t *testing.T) {
	ta := setupTestApp(t)

	// No "Bearer" prefix
	resp, _ := ta.makeRequest(t, "GET", "/api/v1/auth/me", nil, map[string]string{
		"Authorization": "just-a-token",
	})
	if resp.StatusCode != fiber.StatusUnauthorized {
		t.Errorf("Expected status 401 for malformed auth header, got %d", resp.StatusCode)
	}
}

// ============================================
// TEST: Refresh Token
// ============================================

func TestRefresh_Success(t *testing.T) {
	ta := setupTestApp(t)
	email := uniqueEmail()

	// Register and get refresh token cookie
	resp, _ := ta.makeRequest(t, "POST", "/api/v1/auth/register", map[string]string{
		"email":    email,
		"password": "TestPass123",
		"name":     "Test User Refresh",
	}, nil)

	// Extract refresh token from cookie
	var refreshToken string
	for _, c := range resp.Cookies() {
		if c.Name == "refresh_token" {
			refreshToken = c.Value
		}
	}
	if refreshToken == "" {
		t.Fatal("No refresh_token cookie found after registration")
	}

	// Make refresh request with cookie
	req := httptest.NewRequest("POST", "/api/v1/auth/refresh", nil)
	req.Header.Set("Content-Type", "application/json")
	req.AddCookie(&http.Cookie{Name: "refresh_token", Value: refreshToken})

	refreshResp, err := ta.app.Test(req, -1)
	if err != nil {
		t.Fatalf("Failed to execute refresh request: %v", err)
	}

	respBody, _ := io.ReadAll(refreshResp.Body)
	defer refreshResp.Body.Close()

	var apiResp apiResponse
	json.Unmarshal(respBody, &apiResp)

	if refreshResp.StatusCode != fiber.StatusOK {
		t.Errorf("Expected status 200, got %d. Error: %s", refreshResp.StatusCode, apiResp.Error)
	}
	if !apiResp.Success {
		t.Errorf("Expected success=true. Error: %s", apiResp.Error)
	}

	var tokenData tokenResponseData
	json.Unmarshal(apiResp.Data, &tokenData)
	if tokenData.AccessToken == "" {
		t.Error("Expected non-empty access_token after refresh")
	}

	// Verify new refresh token cookie was set (rotation)
	newRefreshFound := false
	for _, c := range refreshResp.Cookies() {
		if c.Name == "refresh_token" && c.Value != "" {
			newRefreshFound = true
		}
	}
	if !newRefreshFound {
		t.Error("Expected new refresh_token cookie after token rotation")
	}
}

func TestRefresh_NoCookie(t *testing.T) {
	ta := setupTestApp(t)

	resp, apiResp := ta.makeRequest(t, "POST", "/api/v1/auth/refresh", nil, nil)

	if resp.StatusCode != fiber.StatusUnauthorized {
		t.Errorf("Expected status 401, got %d", resp.StatusCode)
	}
	if apiResp.Success {
		t.Error("Expected success=false without refresh token cookie")
	}
}

func TestRefresh_InvalidToken(t *testing.T) {
	ta := setupTestApp(t)

	req := httptest.NewRequest("POST", "/api/v1/auth/refresh", nil)
	req.Header.Set("Content-Type", "application/json")
	req.AddCookie(&http.Cookie{Name: "refresh_token", Value: "invalid-refresh-token"})

	resp, err := ta.app.Test(req, -1)
	if err != nil {
		t.Fatalf("Failed to execute request: %v", err)
	}

	if resp.StatusCode != fiber.StatusUnauthorized {
		t.Errorf("Expected status 401, got %d", resp.StatusCode)
	}
}

// ============================================
// TEST: Logout
// ============================================

func TestLogout_Success(t *testing.T) {
	ta := setupTestApp(t)
	email := uniqueEmail()

	// Register
	_, apiResp := ta.makeRequest(t, "POST", "/api/v1/auth/register", map[string]string{
		"email":    email,
		"password": "TestPass123",
		"name":     "Test Logout",
	}, nil)

	var data authResponseData
	json.Unmarshal(apiResp.Data, &data)

	// Logout
	resp, logoutResp := ta.makeRequest(t, "POST", "/api/v1/auth/logout", nil, map[string]string{
		"Authorization": "Bearer " + data.AccessToken,
	})

	if resp.StatusCode != fiber.StatusOK {
		t.Errorf("Expected status 200, got %d. Error: %s", resp.StatusCode, logoutResp.Error)
	}
	if !logoutResp.Success {
		t.Errorf("Expected success=true. Error: %s", logoutResp.Error)
	}
}

func TestLogout_NoToken(t *testing.T) {
	ta := setupTestApp(t)

	resp, apiResp := ta.makeRequest(t, "POST", "/api/v1/auth/logout", nil, nil)

	if resp.StatusCode != fiber.StatusUnauthorized {
		t.Errorf("Expected status 401, got %d", resp.StatusCode)
	}
	if apiResp.Success {
		t.Error("Expected success=false without token")
	}
}

// ============================================
// TEST: Full Auth Flow (Register -> Login -> Me -> Refresh -> Logout)
// ============================================

func TestFullAuthFlow(t *testing.T) {
	ta := setupTestApp(t)
	email := uniqueEmail()
	password := "FlowTest123"

	// Step 1: Register
	regResp, regAPI := ta.makeRequest(t, "POST", "/api/v1/auth/register", map[string]string{
		"email":    email,
		"password": password,
		"name":     "Flow Test User",
	}, nil)

	if regResp.StatusCode != fiber.StatusCreated {
		t.Fatalf("Register failed: %d - %s", regResp.StatusCode, regAPI.Error)
	}

	// Step 2: Login with same credentials
	loginResp, loginAPI := ta.makeRequest(t, "POST", "/api/v1/auth/login", map[string]string{
		"email":    email,
		"password": password,
	}, nil)

	if loginResp.StatusCode != fiber.StatusOK {
		t.Fatalf("Login failed: %d - %s", loginResp.StatusCode, loginAPI.Error)
	}

	var loginData authResponseData
	json.Unmarshal(loginAPI.Data, &loginData)

	// Step 3: Get profile
	_, meAPI := ta.makeRequest(t, "GET", "/api/v1/auth/me", nil, map[string]string{
		"Authorization": "Bearer " + loginData.AccessToken,
	})

	if !meAPI.Success {
		t.Fatalf("Me failed: %s", meAPI.Error)
	}

	var user userDTO
	json.Unmarshal(meAPI.Data, &user)
	if user.Email != email {
		t.Errorf("Me returned wrong email: %s", user.Email)
	}

	// Step 4: Refresh token
	var refreshToken string
	for _, c := range loginResp.Cookies() {
		if c.Name == "refresh_token" {
			refreshToken = c.Value
		}
	}

	req := httptest.NewRequest("POST", "/api/v1/auth/refresh", nil)
	req.Header.Set("Content-Type", "application/json")
	req.AddCookie(&http.Cookie{Name: "refresh_token", Value: refreshToken})

	refreshResp, _ := ta.app.Test(req, -1)
	if refreshResp.StatusCode != fiber.StatusOK {
		t.Fatalf("Refresh failed: status %d", refreshResp.StatusCode)
	}

	// Step 5: Logout
	logoutResp, logoutAPI := ta.makeRequest(t, "POST", "/api/v1/auth/logout", nil, map[string]string{
		"Authorization": "Bearer " + loginData.AccessToken,
	})

	if logoutResp.StatusCode != fiber.StatusOK {
		t.Fatalf("Logout failed: %d - %s", logoutResp.StatusCode, logoutAPI.Error)
	}

	t.Log("Full auth flow completed successfully: Register -> Login -> Me -> Refresh -> Logout")
}
