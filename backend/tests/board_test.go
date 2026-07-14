// Package tests provides integration tests for the TaskFlow Board API.
package tests

import (
	"encoding/json"
	"testing"

	"github.com/gofiber/fiber/v2"
)

// ============================================
// TEST: Create Board
// ============================================

func TestCreateBoard_Success(t *testing.T) {
	ta := SetupTestApp(t)

	// Create workspace
	email := UniqueEmail()
	token, _ := ta.RegisterAndLogin(t, email, "TestPass123", "Owner")
	_, wsResp := ta.MakeRequest(t, "POST", "/api/v1/workspaces", map[string]interface{}{
		"name": "My Workspace",
		"slug": UniqueSlug(),
	}, AuthHeader(token))

	var ws WorkspaceDTO
	json.Unmarshal(wsResp.Data, &ws)

	// Create board
	resp, apiResp := ta.MakeRequest(t, "POST", "/api/v1/boards", map[string]interface{}{
		"workspace_id": ws.ID,
		"name":         "Engineering Board",
		"description":  "Sprint planning",
		"color":        "#FF5733",
	}, AuthHeader(token))

	if resp.StatusCode != fiber.StatusCreated {
		t.Errorf("Expected 201, got %d. Error: %s", resp.StatusCode, apiResp.Error)
	}
	if !apiResp.Success {
		t.Errorf("Expected success=true. Error: %s", apiResp.Error)
	}

	var board BoardDTO
	if err := json.Unmarshal(apiResp.Data, &board); err != nil {
		t.Fatalf("Failed to parse board: %v", err)
	}
	if board.Name != "Engineering Board" {
		t.Errorf("Expected name 'Engineering Board', got '%s'", board.Name)
	}
	if board.WorkspaceID != ws.ID {
		t.Errorf("Expected workspace_id %s, got %s", ws.ID, board.WorkspaceID)
	}
	if board.Color == nil || *board.Color != "#FF5733" {
		t.Errorf("Expected color '#FF5733', got %v", board.Color)
	}
}

func TestCreateBoard_NotMember(t *testing.T) {
	ta := SetupTestApp(t)

	// Owner creates workspace
	emailOwner := UniqueEmail()
	tokenOwner, _ := ta.RegisterAndLogin(t, emailOwner, "TestPass123", "Owner")
	_, wsResp := ta.MakeRequest(t, "POST", "/api/v1/workspaces", map[string]interface{}{
		"name": "Owner WS",
		"slug": UniqueSlug(),
	}, AuthHeader(tokenOwner))
	var ws WorkspaceDTO
	json.Unmarshal(wsResp.Data, &ws)

	// Non-member tries to create board
	emailOther := UniqueEmail()
	tokenOther, _ := ta.RegisterAndLogin(t, emailOther, "TestPass123", "Other")

	resp, _ := ta.MakeRequest(t, "POST", "/api/v1/boards", map[string]interface{}{
		"workspace_id": ws.ID,
		"name":         "Hacked Board",
	}, AuthHeader(tokenOther))

	if resp.StatusCode != fiber.StatusForbidden {
		t.Errorf("Expected 403, got %d", resp.StatusCode)
	}
}

func TestCreateBoard_AsMember(t *testing.T) {
	ta := SetupTestApp(t)

	// Owner creates workspace
	emailOwner := UniqueEmail()
	tokenOwner, _ := ta.RegisterAndLogin(t, emailOwner, "TestPass123", "Owner")
	_, wsResp := ta.MakeRequest(t, "POST", "/api/v1/workspaces", map[string]interface{}{
		"name": "Owner WS",
		"slug": UniqueSlug(),
	}, AuthHeader(tokenOwner))
	var ws WorkspaceDTO
	json.Unmarshal(wsResp.Data, &ws)

	// Invite member
	emailMember := UniqueEmail()
	tokenMember, _ := ta.RegisterAndLogin(t, emailMember, "TestPass123", "Member")
	ta.MakeRequest(t, "POST", "/api/v1/workspaces/"+ws.ID+"/members", map[string]interface{}{
		"email": emailMember,
		"role":  "MEMBER",
	}, AuthHeader(tokenOwner))

	// Member tries to create board (should fail)
	resp, _ := ta.MakeRequest(t, "POST", "/api/v1/boards", map[string]interface{}{
		"workspace_id": ws.ID,
		"name":         "Member Board",
	}, AuthHeader(tokenMember))

	if resp.StatusCode != fiber.StatusForbidden {
		t.Errorf("Expected 403 for MEMBER creating board, got %d", resp.StatusCode)
	}
}

// ============================================
// TEST: List Boards
// ============================================

func TestListBoards_Success(t *testing.T) {
	ta := SetupTestApp(t)

	email := UniqueEmail()
	token, _ := ta.RegisterAndLogin(t, email, "TestPass123", "Owner")
	_, wsResp := ta.MakeRequest(t, "POST", "/api/v1/workspaces", map[string]interface{}{
		"name": "List WS",
		"slug": UniqueSlug(),
	}, AuthHeader(token))
	var ws WorkspaceDTO
	json.Unmarshal(wsResp.Data, &ws)

	// Create 2 boards
	ta.MakeRequest(t, "POST", "/api/v1/boards", map[string]interface{}{
		"workspace_id": ws.ID,
		"name":         "Board 1",
	}, AuthHeader(token))

	ta.MakeRequest(t, "POST", "/api/v1/boards", map[string]interface{}{
		"workspace_id": ws.ID,
		"name":         "Board 2",
	}, AuthHeader(token))

	resp, apiResp := ta.MakeRequest(t, "GET", "/api/v1/workspaces/"+ws.ID+"/boards", nil, AuthHeader(token))

	if resp.StatusCode != fiber.StatusOK {
		t.Errorf("Expected 200, got %d. Error: %s", resp.StatusCode, apiResp.Error)
	}

	var boards []BoardDTO
	json.Unmarshal(apiResp.Data, &boards)
	if len(boards) != 2 {
		t.Errorf("Expected 2 boards, got %d", len(boards))
	}
}

func TestListBoards_AsMember(t *testing.T) {
	ta := SetupTestApp(t)

	emailOwner := UniqueEmail()
	tokenOwner, _ := ta.RegisterAndLogin(t, emailOwner, "TestPass123", "Owner")
	_, wsResp := ta.MakeRequest(t, "POST", "/api/v1/workspaces", map[string]interface{}{
		"name": "List Member WS",
		"slug": UniqueSlug(),
	}, AuthHeader(tokenOwner))
	var ws WorkspaceDTO
	json.Unmarshal(wsResp.Data, &ws)

	ta.MakeRequest(t, "POST", "/api/v1/boards", map[string]interface{}{
		"workspace_id": ws.ID,
		"name":         "Board 1",
	}, AuthHeader(tokenOwner))

	// Invite member
	emailMember := UniqueEmail()
	tokenMember, _ := ta.RegisterAndLogin(t, emailMember, "TestPass123", "Member")
	ta.MakeRequest(t, "POST", "/api/v1/workspaces/"+ws.ID+"/members", map[string]interface{}{
		"email": emailMember,
		"role":  "MEMBER",
	}, AuthHeader(tokenOwner))

	// Member should be able to list boards
	resp, apiResp := ta.MakeRequest(t, "GET", "/api/v1/workspaces/"+ws.ID+"/boards", nil, AuthHeader(tokenMember))

	if resp.StatusCode != fiber.StatusOK {
		t.Errorf("Expected 200, got %d. Error: %s", resp.StatusCode, apiResp.Error)
	}

	var boards []BoardDTO
	json.Unmarshal(apiResp.Data, &boards)
	if len(boards) != 1 {
		t.Errorf("Expected 1 board, got %d", len(boards))
	}
}

// ============================================
// TEST: Get Board
// ============================================

func TestGetBoard_Success(t *testing.T) {
	ta := SetupTestApp(t)

	email := UniqueEmail()
	token, _ := ta.RegisterAndLogin(t, email, "TestPass123", "Owner")
	_, wsResp := ta.MakeRequest(t, "POST", "/api/v1/workspaces", map[string]interface{}{
		"name": "Get WS",
		"slug": UniqueSlug(),
	}, AuthHeader(token))
	var ws WorkspaceDTO
	json.Unmarshal(wsResp.Data, &ws)

	_, createResp := ta.MakeRequest(t, "POST", "/api/v1/boards", map[string]interface{}{
		"workspace_id": ws.ID,
		"name":         "Get Board",
	}, AuthHeader(token))
	var board BoardDTO
	json.Unmarshal(createResp.Data, &board)

	resp, apiResp := ta.MakeRequest(t, "GET", "/api/v1/boards/"+board.ID, nil, AuthHeader(token))

	if resp.StatusCode != fiber.StatusOK {
		t.Errorf("Expected 200, got %d. Error: %s", resp.StatusCode, apiResp.Error)
	}

	var fetched BoardDTO
	json.Unmarshal(apiResp.Data, &fetched)
	if fetched.Name != "Get Board" {
		t.Errorf("Expected name 'Get Board', got '%s'", fetched.Name)
	}
}

// ============================================
// TEST: Update Board
// ============================================

func TestUpdateBoard_Success(t *testing.T) {
	ta := SetupTestApp(t)

	email := UniqueEmail()
	token, _ := ta.RegisterAndLogin(t, email, "TestPass123", "Owner")
	_, wsResp := ta.MakeRequest(t, "POST", "/api/v1/workspaces", map[string]interface{}{
		"name": "Update WS",
		"slug": UniqueSlug(),
	}, AuthHeader(token))
	var ws WorkspaceDTO
	json.Unmarshal(wsResp.Data, &ws)

	_, createResp := ta.MakeRequest(t, "POST", "/api/v1/boards", map[string]interface{}{
		"workspace_id": ws.ID,
		"name":         "Old Board",
	}, AuthHeader(token))
	var board BoardDTO
	json.Unmarshal(createResp.Data, &board)

	resp, apiResp := ta.MakeRequest(t, "PUT", "/api/v1/boards/"+board.ID, map[string]interface{}{
		"name":  "New Board Name",
		"color": "#000000",
	}, AuthHeader(token))

	if resp.StatusCode != fiber.StatusOK {
		t.Errorf("Expected 200, got %d. Error: %s", resp.StatusCode, apiResp.Error)
	}

	var updated BoardDTO
	json.Unmarshal(apiResp.Data, &updated)
	if updated.Name != "New Board Name" {
		t.Errorf("Expected name 'New Board Name', got '%s'", updated.Name)
	}
	if updated.Color == nil || *updated.Color != "#000000" {
		t.Errorf("Expected color '#000000'")
	}
}

// ============================================
// TEST: Delete Board
// ============================================

func TestDeleteBoard_Success(t *testing.T) {
	ta := SetupTestApp(t)

	email := UniqueEmail()
	token, _ := ta.RegisterAndLogin(t, email, "TestPass123", "Owner")
	_, wsResp := ta.MakeRequest(t, "POST", "/api/v1/workspaces", map[string]interface{}{
		"name": "Delete WS",
		"slug": UniqueSlug(),
	}, AuthHeader(token))
	var ws WorkspaceDTO
	json.Unmarshal(wsResp.Data, &ws)

	_, createResp := ta.MakeRequest(t, "POST", "/api/v1/boards", map[string]interface{}{
		"workspace_id": ws.ID,
		"name":         "Delete Board",
	}, AuthHeader(token))
	var board BoardDTO
	json.Unmarshal(createResp.Data, &board)

	resp, apiResp := ta.MakeRequest(t, "DELETE", "/api/v1/boards/"+board.ID, nil, AuthHeader(token))

	if resp.StatusCode != fiber.StatusOK {
		t.Errorf("Expected 200, got %d. Error: %s", resp.StatusCode, apiResp.Error)
	}

	// Verify deleted
	getResp, _ := ta.MakeRequest(t, "GET", "/api/v1/boards/"+board.ID, nil, AuthHeader(token))
	if getResp.StatusCode != fiber.StatusNotFound {
		t.Errorf("Expected 404 after delete, got %d", getResp.StatusCode)
	}
}
