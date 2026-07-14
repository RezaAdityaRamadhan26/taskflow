// Package tests provides integration tests for the TaskFlow List API.
package tests

import (
	"encoding/json"
	"testing"

	"github.com/gofiber/fiber/v2"
)

// ============================================
// TEST: Create List
// ============================================

func TestCreateList_Success(t *testing.T) {
	ta := SetupTestApp(t)

	// Setup Workspace and Board
	email := UniqueEmail()
	token, _ := ta.RegisterAndLogin(t, email, "TestPass123", "Owner")
	_, wsResp := ta.MakeRequest(t, "POST", "/api/v1/workspaces", map[string]interface{}{
		"name": "List WS",
		"slug": UniqueSlug(),
	}, AuthHeader(token))
	var ws WorkspaceDTO
	json.Unmarshal(wsResp.Data, &ws)

	_, boardResp := ta.MakeRequest(t, "POST", "/api/v1/boards", map[string]interface{}{
		"workspace_id": ws.ID,
		"name":         "List Board",
	}, AuthHeader(token))
	var board BoardDTO
	json.Unmarshal(boardResp.Data, &board)

	// Create List 1 (should have pos 65536)
	resp1, apiResp1 := ta.MakeRequest(t, "POST", "/api/v1/lists", map[string]interface{}{
		"board_id": board.ID,
		"name":     "To Do",
	}, AuthHeader(token))

	if resp1.StatusCode != fiber.StatusCreated {
		t.Errorf("Expected 201, got %d. Error: %s", resp1.StatusCode, apiResp1.Error)
	}

	var list1 ListDTO
	json.Unmarshal(apiResp1.Data, &list1)
	if list1.Name != "To Do" {
		t.Errorf("Expected name 'To Do', got '%s'", list1.Name)
	}
	if list1.Position != 65536.0 {
		t.Errorf("Expected position 65536, got %f", list1.Position)
	}

	// Create List 2 (should have pos 131072)
	_, apiResp2 := ta.MakeRequest(t, "POST", "/api/v1/lists", map[string]interface{}{
		"board_id": board.ID,
		"name":     "In Progress",
	}, AuthHeader(token))
	var list2 ListDTO
	json.Unmarshal(apiResp2.Data, &list2)
	if list2.Position != 131072.0 {
		t.Errorf("Expected position 131072, got %f", list2.Position)
	}
}

func TestCreateList_NotMember(t *testing.T) {
	ta := SetupTestApp(t)

	emailOwner := UniqueEmail()
	tokenOwner, _ := ta.RegisterAndLogin(t, emailOwner, "TestPass123", "Owner")
	_, wsResp := ta.MakeRequest(t, "POST", "/api/v1/workspaces", map[string]interface{}{
		"name": "Owner WS",
		"slug": UniqueSlug(),
	}, AuthHeader(tokenOwner))
	var ws WorkspaceDTO
	json.Unmarshal(wsResp.Data, &ws)

	_, boardResp := ta.MakeRequest(t, "POST", "/api/v1/boards", map[string]interface{}{
		"workspace_id": ws.ID,
		"name":         "Owner Board",
	}, AuthHeader(tokenOwner))
	var board BoardDTO
	json.Unmarshal(boardResp.Data, &board)

	// Non-member tries to create list
	emailOther := UniqueEmail()
	tokenOther, _ := ta.RegisterAndLogin(t, emailOther, "TestPass123", "Other")

	resp, _ := ta.MakeRequest(t, "POST", "/api/v1/lists", map[string]interface{}{
		"board_id": board.ID,
		"name":     "Hacked List",
	}, AuthHeader(tokenOther))

	if resp.StatusCode != fiber.StatusForbidden {
		t.Errorf("Expected 403, got %d", resp.StatusCode)
	}
}

// ============================================
// TEST: List Lists (Get all lists in board)
// ============================================

func TestListLists_Success(t *testing.T) {
	ta := SetupTestApp(t)

	email := UniqueEmail()
	token, _ := ta.RegisterAndLogin(t, email, "TestPass123", "Owner")
	_, wsResp := ta.MakeRequest(t, "POST", "/api/v1/workspaces", map[string]interface{}{
		"name": "WS",
		"slug": UniqueSlug(),
	}, AuthHeader(token))
	var ws WorkspaceDTO
	json.Unmarshal(wsResp.Data, &ws)

	_, boardResp := ta.MakeRequest(t, "POST", "/api/v1/boards", map[string]interface{}{
		"workspace_id": ws.ID,
		"name":         "Board",
	}, AuthHeader(token))
	var board BoardDTO
	json.Unmarshal(boardResp.Data, &board)

	// Create 3 lists
	ta.MakeRequest(t, "POST", "/api/v1/lists", map[string]interface{}{
		"board_id": board.ID, "name": "To Do",
	}, AuthHeader(token))
	ta.MakeRequest(t, "POST", "/api/v1/lists", map[string]interface{}{
		"board_id": board.ID, "name": "In Progress",
	}, AuthHeader(token))
	ta.MakeRequest(t, "POST", "/api/v1/lists", map[string]interface{}{
		"board_id": board.ID, "name": "Done",
	}, AuthHeader(token))

	resp, apiResp := ta.MakeRequest(t, "GET", "/api/v1/boards/"+board.ID+"/lists", nil, AuthHeader(token))

	if resp.StatusCode != fiber.StatusOK {
		t.Errorf("Expected 200, got %d", resp.StatusCode)
	}

	var lists []ListDTO
	json.Unmarshal(apiResp.Data, &lists)
	if len(lists) != 3 {
		t.Errorf("Expected 3 lists, got %d", len(lists))
	}
	// Verify order
	if lists[0].Name != "To Do" || lists[2].Name != "Done" {
		t.Errorf("Lists are not ordered correctly")
	}
}

// ============================================
// TEST: Update List
// ============================================

func TestUpdateList_Success(t *testing.T) {
	ta := SetupTestApp(t)

	email := UniqueEmail()
	token, _ := ta.RegisterAndLogin(t, email, "TestPass123", "Owner")
	_, wsResp := ta.MakeRequest(t, "POST", "/api/v1/workspaces", map[string]interface{}{
		"name": "WS",
		"slug": UniqueSlug(),
	}, AuthHeader(token))
	var ws WorkspaceDTO
	json.Unmarshal(wsResp.Data, &ws)

	_, boardResp := ta.MakeRequest(t, "POST", "/api/v1/boards", map[string]interface{}{
		"workspace_id": ws.ID,
		"name":         "Board",
	}, AuthHeader(token))
	var board BoardDTO
	json.Unmarshal(boardResp.Data, &board)

	_, listResp := ta.MakeRequest(t, "POST", "/api/v1/lists", map[string]interface{}{
		"board_id": board.ID,
		"name":     "Old Name",
	}, AuthHeader(token))
	var list ListDTO
	json.Unmarshal(listResp.Data, &list)

	// Update name and position
	resp, apiResp := ta.MakeRequest(t, "PUT", "/api/v1/lists/"+list.ID, map[string]interface{}{
		"name":     "New Name",
		"position": 1000.5,
	}, AuthHeader(token))

	if resp.StatusCode != fiber.StatusOK {
		t.Errorf("Expected 200, got %d", resp.StatusCode)
	}

	var updated ListDTO
	json.Unmarshal(apiResp.Data, &updated)
	if updated.Name != "New Name" {
		t.Errorf("Expected name 'New Name', got '%s'", updated.Name)
	}
	if updated.Position != 1000.5 {
		t.Errorf("Expected pos 1000.5, got %f", updated.Position)
	}
}

// ============================================
// TEST: Delete List
// ============================================

func TestDeleteList_Success(t *testing.T) {
	ta := SetupTestApp(t)

	email := UniqueEmail()
	token, _ := ta.RegisterAndLogin(t, email, "TestPass123", "Owner")
	_, wsResp := ta.MakeRequest(t, "POST", "/api/v1/workspaces", map[string]interface{}{
		"name": "WS",
		"slug": UniqueSlug(),
	}, AuthHeader(token))
	var ws WorkspaceDTO
	json.Unmarshal(wsResp.Data, &ws)

	_, boardResp := ta.MakeRequest(t, "POST", "/api/v1/boards", map[string]interface{}{
		"workspace_id": ws.ID,
		"name":         "Board",
	}, AuthHeader(token))
	var board BoardDTO
	json.Unmarshal(boardResp.Data, &board)

	_, listResp := ta.MakeRequest(t, "POST", "/api/v1/lists", map[string]interface{}{
		"board_id": board.ID,
		"name":     "To Delete",
	}, AuthHeader(token))
	var list ListDTO
	json.Unmarshal(listResp.Data, &list)

	resp, _ := ta.MakeRequest(t, "DELETE", "/api/v1/lists/"+list.ID, nil, AuthHeader(token))

	if resp.StatusCode != fiber.StatusOK {
		t.Errorf("Expected 200, got %d", resp.StatusCode)
	}

	// Verify deleted
	getResp, _ := ta.MakeRequest(t, "GET", "/api/v1/lists/"+list.ID, nil, AuthHeader(token))
	if getResp.StatusCode != fiber.StatusNotFound {
		t.Errorf("Expected 404 after delete, got %d", getResp.StatusCode)
	}
}
