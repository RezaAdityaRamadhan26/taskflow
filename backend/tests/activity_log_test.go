// Package tests provides integration tests for the TaskFlow Activity Log API.
package tests

import (
	"encoding/json"
	"testing"
	"time"

	"github.com/gofiber/fiber/v2"
)

// ============================================
// TEST: Activity Logs
// ============================================

func TestActivityLogs_E2E(t *testing.T) {
	ta := SetupTestApp(t)

	// User registers and logs in
	email := UniqueEmail()
	token, _ := ta.RegisterAndLogin(t, email, "TestPass123", "Audit User")

	// 1. Create Workspace
	_, wsResp := ta.MakeRequest(t, "POST", "/api/v1/workspaces", map[string]interface{}{
		"name": "Audit WS", "slug": UniqueSlug(),
	}, AuthHeader(token))
	var ws WorkspaceDTO
	json.Unmarshal(wsResp.Data, &ws)

	// 2. Create Board
	_, boardResp := ta.MakeRequest(t, "POST", "/api/v1/boards", map[string]interface{}{
		"workspace_id": ws.ID, "name": "Audit Board",
	}, AuthHeader(token))
	var board BoardDTO
	json.Unmarshal(boardResp.Data, &board)

	// 3. Create List
	_, listResp := ta.MakeRequest(t, "POST", "/api/v1/lists", map[string]interface{}{
		"board_id": board.ID, "name": "To Do",
	}, AuthHeader(token))
	var list ListDTO
	json.Unmarshal(listResp.Data, &list)

	// 4. Create Card
	_, cardResp := ta.MakeRequest(t, "POST", "/api/v1/cards", map[string]interface{}{
		"list_id": list.ID, "title": "Audit Task",
	}, AuthHeader(token))
	var card CardDTO
	json.Unmarshal(cardResp.Data, &card)

	// Wait a moment for async logs to be written
	time.Sleep(100 * time.Millisecond)

	// ========================================
	// Verify Board Activities
	// ========================================
	resp, apiResp := ta.MakeRequest(t, "GET", "/api/v1/boards/"+board.ID+"/activities", nil, AuthHeader(token))

	if resp.StatusCode != fiber.StatusOK {
		t.Errorf("Expected 200, got %d", resp.StatusCode)
	}

	var logs []ActivityLogDTO
	json.Unmarshal(apiResp.Data, &logs)

	// We expect 3 logs: CREATED_BOARD, CREATED_LIST, CREATED_CARD
	if len(logs) < 3 {
		t.Errorf("Expected at least 3 activities, got %d", len(logs))
	}

	// Verify the most recent log (CREATED_CARD)
	if logs[0].Action != "CREATED_CARD" {
		t.Errorf("Expected latest action CREATED_CARD, got %s", logs[0].Action)
	}
	if logs[0].EntityTitle != "Audit Task" {
		t.Errorf("Expected entity_title 'Audit Task', got '%s'", logs[0].EntityTitle)
	}

	// ========================================
	// Verify Card Activities
	// ========================================
	cardLogsResp, cardLogsAPI := ta.MakeRequest(t, "GET", "/api/v1/cards/"+card.ID+"/activities", nil, AuthHeader(token))

	if cardLogsResp.StatusCode != fiber.StatusOK {
		t.Errorf("Expected 200, got %d", cardLogsResp.StatusCode)
	}

	var cLogs []ActivityLogDTO
	json.Unmarshal(cardLogsAPI.Data, &cLogs)

	// Should only have 1 log for the card itself (CREATED_CARD)
	if len(cLogs) != 1 {
		t.Errorf("Expected 1 card activity, got %d", len(cLogs))
	}
	if cLogs[0].Action != "CREATED_CARD" {
		t.Errorf("Expected action CREATED_CARD, got %s", cLogs[0].Action)
	}
}

func TestActivityLogs_NoAccess(t *testing.T) {
	ta := SetupTestApp(t)

	// Owner creates board
	emailOwner := UniqueEmail()
	tokenOwner, _ := ta.RegisterAndLogin(t, emailOwner, "TestPass123", "Owner")
	_, wsResp := ta.MakeRequest(t, "POST", "/api/v1/workspaces", map[string]interface{}{
		"name": "Audit WS", "slug": UniqueSlug(),
	}, AuthHeader(tokenOwner))
	var ws WorkspaceDTO
	json.Unmarshal(wsResp.Data, &ws)

	_, boardResp := ta.MakeRequest(t, "POST", "/api/v1/boards", map[string]interface{}{
		"workspace_id": ws.ID, "name": "Audit Board",
	}, AuthHeader(tokenOwner))
	var board BoardDTO
	json.Unmarshal(boardResp.Data, &board)

	// Non-member tries to read logs
	emailOther := UniqueEmail()
	tokenOther, _ := ta.RegisterAndLogin(t, emailOther, "TestPass123", "Other User")

	resp, _ := ta.MakeRequest(t, "GET", "/api/v1/boards/"+board.ID+"/activities", nil, AuthHeader(tokenOther))

	if resp.StatusCode != fiber.StatusForbidden {
		t.Errorf("Expected 403, got %d", resp.StatusCode)
	}
}
