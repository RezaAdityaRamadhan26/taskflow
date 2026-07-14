// Package tests provides integration tests for the TaskFlow Auth API.
package tests

import (
	"encoding/json"
	"io"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gofiber/fiber/v2"
)

// ============================================
// TEST: Register
// ============================================

func TestRegister_Success(t *testing.T) {
	ta := SetupTestApp(t)
	email := UniqueEmail()

	resp, apiResp := ta.MakeRequest(t, "POST", "/api/v1/auth/register", map[string]string{
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

	var data AuthResponseData
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
	ta := SetupTestApp(t)
	email := UniqueEmail()

	ta.MakeRequest(t, "POST", "/api/v1/auth/register", map[string]string{
		"email":    email,
		"password": "TestPass123",
		"name":     "Test User",
	}, nil)

	resp, apiResp := ta.MakeRequest(t, "POST", "/api/v1/auth/register", map[string]string{
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
	ta := SetupTestApp(t)

	resp, apiResp := ta.MakeRequest(t, "POST", "/api/v1/auth/register", map[string]string{
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
	ta := SetupTestApp(t)

	resp, apiResp := ta.MakeRequest(t, "POST", "/api/v1/auth/register", map[string]string{
		"email":    UniqueEmail(),
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
	ta := SetupTestApp(t)

	resp, _ := ta.MakeRequest(t, "POST", "/api/v1/auth/register", map[string]string{
		"password": "TestPass123",
		"name":     "Test User",
	}, nil)
	if resp.StatusCode != fiber.StatusUnprocessableEntity {
		t.Errorf("Missing email: expected 422, got %d", resp.StatusCode)
	}

	resp, _ = ta.MakeRequest(t, "POST", "/api/v1/auth/register", map[string]string{
		"email": UniqueEmail(),
		"name":  "Test User",
	}, nil)
	if resp.StatusCode != fiber.StatusUnprocessableEntity {
		t.Errorf("Missing password: expected 422, got %d", resp.StatusCode)
	}

	resp, _ = ta.MakeRequest(t, "POST", "/api/v1/auth/register", map[string]string{
		"email":    UniqueEmail(),
		"password": "TestPass123",
	}, nil)
	if resp.StatusCode != fiber.StatusUnprocessableEntity {
		t.Errorf("Missing name: expected 422, got %d", resp.StatusCode)
	}
}

func TestRegister_EmptyBody(t *testing.T) {
	ta := SetupTestApp(t)

	resp, apiResp := ta.MakeRequest(t, "POST", "/api/v1/auth/register", nil, nil)

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
	ta := SetupTestApp(t)
	email := UniqueEmail()

	ta.MakeRequest(t, "POST", "/api/v1/auth/register", map[string]string{
		"email":    email,
		"password": "TestPass123",
		"name":     "Test User",
	}, nil)

	resp, apiResp := ta.MakeRequest(t, "POST", "/api/v1/auth/login", map[string]string{
		"email":    email,
		"password": "TestPass123",
	}, nil)

	if resp.StatusCode != fiber.StatusOK {
		t.Errorf("Expected status 200, got %d. Error: %s", resp.StatusCode, apiResp.Error)
	}
	if !apiResp.Success {
		t.Errorf("Expected success=true. Error: %s", apiResp.Error)
	}

	var data AuthResponseData
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
	ta := SetupTestApp(t)
	email := UniqueEmail()

	ta.MakeRequest(t, "POST", "/api/v1/auth/register", map[string]string{
		"email":    email,
		"password": "TestPass123",
		"name":     "Test User",
	}, nil)

	resp, apiResp := ta.MakeRequest(t, "POST", "/api/v1/auth/login", map[string]string{
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
	ta := SetupTestApp(t)

	resp, apiResp := ta.MakeRequest(t, "POST", "/api/v1/auth/login", map[string]string{
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
	ta := SetupTestApp(t)

	resp, _ := ta.MakeRequest(t, "POST", "/api/v1/auth/login", map[string]string{
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
	ta := SetupTestApp(t)
	email := UniqueEmail()

	token, _ := ta.RegisterAndLogin(t, email, "TestPass123", "Test User Me")

	resp, meResp := ta.MakeRequest(t, "GET", "/api/v1/auth/me", nil, AuthHeader(token))

	if resp.StatusCode != fiber.StatusOK {
		t.Errorf("Expected status 200, got %d. Error: %s", resp.StatusCode, meResp.Error)
	}
	if !meResp.Success {
		t.Errorf("Expected success=true. Error: %s", meResp.Error)
	}

	var user UserDTO
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
	ta := SetupTestApp(t)

	resp, apiResp := ta.MakeRequest(t, "GET", "/api/v1/auth/me", nil, nil)

	if resp.StatusCode != fiber.StatusUnauthorized {
		t.Errorf("Expected status 401, got %d", resp.StatusCode)
	}
	if apiResp.Success {
		t.Error("Expected success=false without token")
	}
}

func TestMe_InvalidToken(t *testing.T) {
	ta := SetupTestApp(t)

	resp, apiResp := ta.MakeRequest(t, "GET", "/api/v1/auth/me", nil, AuthHeader("invalid-token-here"))

	if resp.StatusCode != fiber.StatusUnauthorized {
		t.Errorf("Expected status 401, got %d", resp.StatusCode)
	}
	if apiResp.Success {
		t.Error("Expected success=false for invalid token")
	}
}

func TestMe_MalformedAuthHeader(t *testing.T) {
	ta := SetupTestApp(t)

	resp, _ := ta.MakeRequest(t, "GET", "/api/v1/auth/me", nil, map[string]string{
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
	ta := SetupTestApp(t)
	email := UniqueEmail()

	resp, _ := ta.MakeRequest(t, "POST", "/api/v1/auth/register", map[string]string{
		"email":    email,
		"password": "TestPass123",
		"name":     "Test User Refresh",
	}, nil)

	var refreshToken string
	for _, c := range resp.Cookies() {
		if c.Name == "refresh_token" {
			refreshToken = c.Value
		}
	}
	if refreshToken == "" {
		t.Fatal("No refresh_token cookie found after registration")
	}

	req := httptest.NewRequest("POST", "/api/v1/auth/refresh", nil)
	req.Header.Set("Content-Type", "application/json")
	req.AddCookie(&http.Cookie{Name: "refresh_token", Value: refreshToken})

	refreshResp, err := ta.App.Test(req, -1)
	if err != nil {
		t.Fatalf("Failed to execute refresh request: %v", err)
	}

	respBody, _ := io.ReadAll(refreshResp.Body)
	defer refreshResp.Body.Close()

	var apiResp APIResponse
	json.Unmarshal(respBody, &apiResp)

	if refreshResp.StatusCode != fiber.StatusOK {
		t.Errorf("Expected status 200, got %d. Error: %s", refreshResp.StatusCode, apiResp.Error)
	}
	if !apiResp.Success {
		t.Errorf("Expected success=true. Error: %s", apiResp.Error)
	}

	var tokenData TokenResponseData
	json.Unmarshal(apiResp.Data, &tokenData)
	if tokenData.AccessToken == "" {
		t.Error("Expected non-empty access_token after refresh")
	}

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
	ta := SetupTestApp(t)

	resp, apiResp := ta.MakeRequest(t, "POST", "/api/v1/auth/refresh", nil, nil)

	if resp.StatusCode != fiber.StatusUnauthorized {
		t.Errorf("Expected status 401, got %d", resp.StatusCode)
	}
	if apiResp.Success {
		t.Error("Expected success=false without refresh token cookie")
	}
}

func TestRefresh_InvalidToken(t *testing.T) {
	ta := SetupTestApp(t)

	req := httptest.NewRequest("POST", "/api/v1/auth/refresh", nil)
	req.Header.Set("Content-Type", "application/json")
	req.AddCookie(&http.Cookie{Name: "refresh_token", Value: "invalid-refresh-token"})

	resp, err := ta.App.Test(req, -1)
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
	ta := SetupTestApp(t)
	email := UniqueEmail()

	token, _ := ta.RegisterAndLogin(t, email, "TestPass123", "Test Logout")

	resp, logoutResp := ta.MakeRequest(t, "POST", "/api/v1/auth/logout", nil, AuthHeader(token))

	if resp.StatusCode != fiber.StatusOK {
		t.Errorf("Expected status 200, got %d. Error: %s", resp.StatusCode, logoutResp.Error)
	}
	if !logoutResp.Success {
		t.Errorf("Expected success=true. Error: %s", logoutResp.Error)
	}
}

func TestLogout_NoToken(t *testing.T) {
	ta := SetupTestApp(t)

	resp, apiResp := ta.MakeRequest(t, "POST", "/api/v1/auth/logout", nil, nil)

	if resp.StatusCode != fiber.StatusUnauthorized {
		t.Errorf("Expected status 401, got %d", resp.StatusCode)
	}
	if apiResp.Success {
		t.Error("Expected success=false without token")
	}
}

// ============================================
// TEST: Full Auth Flow
// ============================================

func TestFullAuthFlow(t *testing.T) {
	ta := SetupTestApp(t)
	email := UniqueEmail()
	password := "FlowTest123"

	// Step 1: Register
	regResp, regAPI := ta.MakeRequest(t, "POST", "/api/v1/auth/register", map[string]string{
		"email":    email,
		"password": password,
		"name":     "Flow Test User",
	}, nil)
	if regResp.StatusCode != fiber.StatusCreated {
		t.Fatalf("Register failed: %d - %s", regResp.StatusCode, regAPI.Error)
	}

	// Step 2: Login
	loginResp, loginAPI := ta.MakeRequest(t, "POST", "/api/v1/auth/login", map[string]string{
		"email":    email,
		"password": password,
	}, nil)
	if loginResp.StatusCode != fiber.StatusOK {
		t.Fatalf("Login failed: %d - %s", loginResp.StatusCode, loginAPI.Error)
	}

	var loginData AuthResponseData
	json.Unmarshal(loginAPI.Data, &loginData)

	// Step 3: Me
	_, meAPI := ta.MakeRequest(t, "GET", "/api/v1/auth/me", nil, AuthHeader(loginData.AccessToken))
	if !meAPI.Success {
		t.Fatalf("Me failed: %s", meAPI.Error)
	}

	// Step 4: Refresh
	var refreshToken string
	for _, c := range loginResp.Cookies() {
		if c.Name == "refresh_token" {
			refreshToken = c.Value
		}
	}

	req := httptest.NewRequest("POST", "/api/v1/auth/refresh", nil)
	req.Header.Set("Content-Type", "application/json")
	req.AddCookie(&http.Cookie{Name: "refresh_token", Value: refreshToken})
	refreshResp, _ := ta.App.Test(req, -1)
	if refreshResp.StatusCode != fiber.StatusOK {
		t.Fatalf("Refresh failed: status %d", refreshResp.StatusCode)
	}

	// Step 5: Logout
	logoutResp, logoutAPI := ta.MakeRequest(t, "POST", "/api/v1/auth/logout", nil, AuthHeader(loginData.AccessToken))
	if logoutResp.StatusCode != fiber.StatusOK {
		t.Fatalf("Logout failed: %d - %s", logoutResp.StatusCode, logoutAPI.Error)
	}

	t.Log("Full auth flow completed successfully: Register -> Login -> Me -> Refresh -> Logout")
}
