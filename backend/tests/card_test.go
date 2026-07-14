// Package tests provides integration tests for the TaskFlow Card API.
package tests

import (
	"encoding/json"
	"testing"
	"time"

	"github.com/gofiber/fiber/v2"
)

// ============================================
// TEST: Create Card
// ============================================

func TestCreateCard_Success(t *testing.T) {
	ta := SetupTestApp(t)

	// Setup Workspace, Board, List
	email := UniqueEmail()
	token, _ := ta.RegisterAndLogin(t, email, "TestPass123", "Owner")
	_, wsResp := ta.MakeRequest(t, "POST", "/api/v1/workspaces", map[string]interface{}{
		"name": "WS", "slug": UniqueSlug(),
	}, AuthHeader(token))
	var ws WorkspaceDTO
	json.Unmarshal(wsResp.Data, &ws)

	_, boardResp := ta.MakeRequest(t, "POST", "/api/v1/boards", map[string]interface{}{
		"workspace_id": ws.ID, "name": "Board",
	}, AuthHeader(token))
	var board BoardDTO
	json.Unmarshal(boardResp.Data, &board)

	_, listResp := ta.MakeRequest(t, "POST", "/api/v1/lists", map[string]interface{}{
		"board_id": board.ID, "name": "To Do",
	}, AuthHeader(token))
	var list ListDTO
	json.Unmarshal(listResp.Data, &list)

	dueDate := time.Now().Add(24 * time.Hour).Format(time.RFC3339)

	// Create Card
	resp, apiResp := ta.MakeRequest(t, "POST", "/api/v1/cards", map[string]interface{}{
		"list_id":     list.ID,
		"title":       "My Task",
		"description": "Task details",
		"priority":    "HIGH",
		"due_date":    dueDate,
	}, AuthHeader(token))

	if resp.StatusCode != fiber.StatusCreated {
		t.Errorf("Expected 201, got %d. Error: %s", resp.StatusCode, apiResp.Error)
	}

	var card CardDTO
	json.Unmarshal(apiResp.Data, &card)
	if card.Title != "My Task" {
		t.Errorf("Expected title 'My Task', got '%s'", card.Title)
	}
	if card.Priority != "HIGH" {
		t.Errorf("Expected priority 'HIGH', got '%s'", card.Priority)
	}
	if card.Description == nil || *card.Description != "Task details" {
		t.Error("Expected description 'Task details'")
	}
	if card.DueDate == nil {
		t.Error("Expected due date to be set")
	}
	if card.Position != 65536.0 {
		t.Errorf("Expected position 65536.0, got %f", card.Position)
	}
}

func TestCreateCard_AsMember(t *testing.T) {
	ta := SetupTestApp(t)

	emailOwner := UniqueEmail()
	tokenOwner, _ := ta.RegisterAndLogin(t, emailOwner, "TestPass123", "Owner")
	_, wsResp := ta.MakeRequest(t, "POST", "/api/v1/workspaces", map[string]interface{}{
		"name": "WS", "slug": UniqueSlug(),
	}, AuthHeader(tokenOwner))
	var ws WorkspaceDTO
	json.Unmarshal(wsResp.Data, &ws)

	_, boardResp := ta.MakeRequest(t, "POST", "/api/v1/boards", map[string]interface{}{
		"workspace_id": ws.ID, "name": "Board",
	}, AuthHeader(tokenOwner))
	var board BoardDTO
	json.Unmarshal(boardResp.Data, &board)

	_, listResp := ta.MakeRequest(t, "POST", "/api/v1/lists", map[string]interface{}{
		"board_id": board.ID, "name": "To Do",
	}, AuthHeader(tokenOwner))
	var list ListDTO
	json.Unmarshal(listResp.Data, &list)

	// Invite member
	emailMember := UniqueEmail()
	tokenMember, _ := ta.RegisterAndLogin(t, emailMember, "TestPass123", "Member")
	ta.MakeRequest(t, "POST", "/api/v1/workspaces/"+ws.ID+"/members", map[string]interface{}{
		"email": emailMember, "role": "MEMBER",
	}, AuthHeader(tokenOwner))

	// Member creates card
	resp, apiResp := ta.MakeRequest(t, "POST", "/api/v1/cards", map[string]interface{}{
		"list_id": list.ID,
		"title":   "Member Task",
	}, AuthHeader(tokenMember))

	if resp.StatusCode != fiber.StatusCreated {
		t.Errorf("Expected 201 for MEMBER creating card, got %d. Error: %s", resp.StatusCode, apiResp.Error)
	}
}

// ============================================
// TEST: List Cards
// ============================================

func TestListCards_Success(t *testing.T) {
	ta := SetupTestApp(t)

	email := UniqueEmail()
	token, _ := ta.RegisterAndLogin(t, email, "TestPass123", "Owner")
	_, wsResp := ta.MakeRequest(t, "POST", "/api/v1/workspaces", map[string]interface{}{
		"name": "WS", "slug": UniqueSlug(),
	}, AuthHeader(token))
	var ws WorkspaceDTO
	json.Unmarshal(wsResp.Data, &ws)

	_, boardResp := ta.MakeRequest(t, "POST", "/api/v1/boards", map[string]interface{}{
		"workspace_id": ws.ID, "name": "Board",
	}, AuthHeader(token))
	var board BoardDTO
	json.Unmarshal(boardResp.Data, &board)

	_, listResp := ta.MakeRequest(t, "POST", "/api/v1/lists", map[string]interface{}{
		"board_id": board.ID, "name": "To Do",
	}, AuthHeader(token))
	var list ListDTO
	json.Unmarshal(listResp.Data, &list)

	ta.MakeRequest(t, "POST", "/api/v1/cards", map[string]interface{}{
		"list_id": list.ID, "title": "Task 1",
	}, AuthHeader(token))
	ta.MakeRequest(t, "POST", "/api/v1/cards", map[string]interface{}{
		"list_id": list.ID, "title": "Task 2",
	}, AuthHeader(token))

	resp, apiResp := ta.MakeRequest(t, "GET", "/api/v1/lists/"+list.ID+"/cards", nil, AuthHeader(token))

	if resp.StatusCode != fiber.StatusOK {
		t.Errorf("Expected 200, got %d", resp.StatusCode)
	}

	var cards []CardDTO
	json.Unmarshal(apiResp.Data, &cards)
	if len(cards) != 2 {
		t.Errorf("Expected 2 cards, got %d", len(cards))
	}
	if cards[0].Title != "Task 1" || cards[1].Title != "Task 2" {
		t.Errorf("Cards are not ordered correctly by position")
	}
}

// ============================================
// TEST: Update Card (and move between lists)
// ============================================

func TestUpdateCard_Success(t *testing.T) {
	ta := SetupTestApp(t)

	email := UniqueEmail()
	token, _ := ta.RegisterAndLogin(t, email, "TestPass123", "Owner")
	_, wsResp := ta.MakeRequest(t, "POST", "/api/v1/workspaces", map[string]interface{}{
		"name": "WS", "slug": UniqueSlug(),
	}, AuthHeader(token))
	var ws WorkspaceDTO
	json.Unmarshal(wsResp.Data, &ws)

	_, boardResp := ta.MakeRequest(t, "POST", "/api/v1/boards", map[string]interface{}{
		"workspace_id": ws.ID, "name": "Board",
	}, AuthHeader(token))
	var board BoardDTO
	json.Unmarshal(boardResp.Data, &board)

	_, listResp1 := ta.MakeRequest(t, "POST", "/api/v1/lists", map[string]interface{}{
		"board_id": board.ID, "name": "To Do",
	}, AuthHeader(token))
	var list1 ListDTO
	json.Unmarshal(listResp1.Data, &list1)

	_, listResp2 := ta.MakeRequest(t, "POST", "/api/v1/lists", map[string]interface{}{
		"board_id": board.ID, "name": "Done",
	}, AuthHeader(token))
	var list2 ListDTO
	json.Unmarshal(listResp2.Data, &list2)

	_, cardResp := ta.MakeRequest(t, "POST", "/api/v1/cards", map[string]interface{}{
		"list_id": list1.ID, "title": "Old Task",
	}, AuthHeader(token))
	var card CardDTO
	json.Unmarshal(cardResp.Data, &card)

	// Move to list2, rename, change position and priority
	resp, apiResp := ta.MakeRequest(t, "PUT", "/api/v1/cards/"+card.ID, map[string]interface{}{
		"list_id":  list2.ID,
		"title":    "New Task Name",
		"position": 5000.5,
		"priority": "URGENT",
	}, AuthHeader(token))

	if resp.StatusCode != fiber.StatusOK {
		t.Errorf("Expected 200, got %d. Error: %s", resp.StatusCode, apiResp.Error)
	}

	var updated CardDTO
	json.Unmarshal(apiResp.Data, &updated)
	if updated.ListID != list2.ID {
		t.Errorf("Expected card to be moved to list %s, got %s", list2.ID, updated.ListID)
	}
	if updated.Title != "New Task Name" {
		t.Errorf("Expected title 'New Task Name', got '%s'", updated.Title)
	}
	if updated.Position != 5000.5 {
		t.Errorf("Expected pos 5000.5, got %f", updated.Position)
	}
	if updated.Priority != "URGENT" {
		t.Errorf("Expected priority 'URGENT', got '%s'", updated.Priority)
	}
}

// ============================================
// TEST: Delete Card
// ============================================

func TestDeleteCard_Success(t *testing.T) {
	ta := SetupTestApp(t)

	email := UniqueEmail()
	token, _ := ta.RegisterAndLogin(t, email, "TestPass123", "Owner")
	_, wsResp := ta.MakeRequest(t, "POST", "/api/v1/workspaces", map[string]interface{}{
		"name": "WS", "slug": UniqueSlug(),
	}, AuthHeader(token))
	var ws WorkspaceDTO
	json.Unmarshal(wsResp.Data, &ws)

	_, boardResp := ta.MakeRequest(t, "POST", "/api/v1/boards", map[string]interface{}{
		"workspace_id": ws.ID, "name": "Board",
	}, AuthHeader(token))
	var board BoardDTO
	json.Unmarshal(boardResp.Data, &board)

	_, listResp := ta.MakeRequest(t, "POST", "/api/v1/lists", map[string]interface{}{
		"board_id": board.ID, "name": "To Do",
	}, AuthHeader(token))
	var list ListDTO
	json.Unmarshal(listResp.Data, &list)

	_, cardResp := ta.MakeRequest(t, "POST", "/api/v1/cards", map[string]interface{}{
		"list_id": list.ID, "title": "To Delete",
	}, AuthHeader(token))
	var card CardDTO
	json.Unmarshal(cardResp.Data, &card)

	resp, _ := ta.MakeRequest(t, "DELETE", "/api/v1/cards/"+card.ID, nil, AuthHeader(token))

	if resp.StatusCode != fiber.StatusOK {
		t.Errorf("Expected 200, got %d", resp.StatusCode)
	}

	getResp, _ := ta.MakeRequest(t, "GET", "/api/v1/cards/"+card.ID, nil, AuthHeader(token))
	if getResp.StatusCode != fiber.StatusNotFound {
		t.Errorf("Expected 404 after delete, got %d", getResp.StatusCode)
	}
}
