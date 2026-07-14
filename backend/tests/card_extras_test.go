// Package tests provides integration tests for the TaskFlow Card Extras API (Comments & Attachments).
package tests

import (
	"encoding/json"
	"testing"

	"github.com/gofiber/fiber/v2"
)

// Helper function to setup workspace -> board -> list -> card
func setupCardEnv(t *testing.T, ta *TestApp, userEmail, userName string) (string, CardDTO) {
	token, _ := ta.RegisterAndLogin(t, userEmail, "TestPass123", userName)

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
		"list_id": list.ID, "title": "My Task",
	}, AuthHeader(token))
	var card CardDTO
	json.Unmarshal(cardResp.Data, &card)

	return token, card
}

// ============================================
// TEST: Comments
// ============================================

func TestCreateComment_Success(t *testing.T) {
	ta := SetupTestApp(t)
	token, card := setupCardEnv(t, ta, UniqueEmail(), "Commenter")

	resp, apiResp := ta.MakeRequest(t, "POST", "/api/v1/comments", map[string]interface{}{
		"card_id": card.ID,
		"content": "This is a test comment",
	}, AuthHeader(token))

	if resp.StatusCode != fiber.StatusCreated {
		t.Errorf("Expected 201, got %d. Error: %s", resp.StatusCode, apiResp.Error)
	}

	var comment CommentDTO
	json.Unmarshal(apiResp.Data, &comment)
	if comment.Content != "This is a test comment" {
		t.Errorf("Expected content 'This is a test comment', got '%s'", comment.Content)
	}
	if comment.CardID != card.ID {
		t.Errorf("Expected card_id %s, got %s", card.ID, comment.CardID)
	}
}

func TestListComments_Success(t *testing.T) {
	ta := SetupTestApp(t)
	token, card := setupCardEnv(t, ta, UniqueEmail(), "Commenter")

	ta.MakeRequest(t, "POST", "/api/v1/comments", map[string]interface{}{
		"card_id": card.ID, "content": "Comment 1",
	}, AuthHeader(token))
	ta.MakeRequest(t, "POST", "/api/v1/comments", map[string]interface{}{
		"card_id": card.ID, "content": "Comment 2",
	}, AuthHeader(token))

	resp, apiResp := ta.MakeRequest(t, "GET", "/api/v1/cards/"+card.ID+"/comments", nil, AuthHeader(token))

	if resp.StatusCode != fiber.StatusOK {
		t.Errorf("Expected 200, got %d", resp.StatusCode)
	}

	var comments []CommentDTO
	json.Unmarshal(apiResp.Data, &comments)
	if len(comments) != 2 {
		t.Errorf("Expected 2 comments, got %d", len(comments))
	}
}

func TestUpdateComment_Success(t *testing.T) {
	ta := SetupTestApp(t)
	token, card := setupCardEnv(t, ta, UniqueEmail(), "Commenter")

	_, cResp := ta.MakeRequest(t, "POST", "/api/v1/comments", map[string]interface{}{
		"card_id": card.ID, "content": "Old Comment",
	}, AuthHeader(token))
	var comment CommentDTO
	json.Unmarshal(cResp.Data, &comment)

	resp, apiResp := ta.MakeRequest(t, "PUT", "/api/v1/comments/"+comment.ID, map[string]interface{}{
		"content": "Updated Comment",
	}, AuthHeader(token))

	if resp.StatusCode != fiber.StatusOK {
		t.Errorf("Expected 200, got %d", resp.StatusCode)
	}

	var updated CommentDTO
	json.Unmarshal(apiResp.Data, &updated)
	if updated.Content != "Updated Comment" {
		t.Errorf("Expected updated content, got '%s'", updated.Content)
	}
}

func TestDeleteComment_Success(t *testing.T) {
	ta := SetupTestApp(t)
	token, card := setupCardEnv(t, ta, UniqueEmail(), "Commenter")

	_, cResp := ta.MakeRequest(t, "POST", "/api/v1/comments", map[string]interface{}{
		"card_id": card.ID, "content": "To Delete",
	}, AuthHeader(token))
	var comment CommentDTO
	json.Unmarshal(cResp.Data, &comment)

	resp, _ := ta.MakeRequest(t, "DELETE", "/api/v1/comments/"+comment.ID, nil, AuthHeader(token))

	if resp.StatusCode != fiber.StatusOK {
		t.Errorf("Expected 200, got %d", resp.StatusCode)
	}
}

// ============================================
// TEST: Attachments
// ============================================

func TestCreateAttachment_Success(t *testing.T) {
	ta := SetupTestApp(t)
	token, card := setupCardEnv(t, ta, UniqueEmail(), "Uploader")

	resp, apiResp := ta.MakeRequest(t, "POST", "/api/v1/attachments", map[string]interface{}{
		"card_id":   card.ID,
		"file_name": "design.png",
		"file_url":  "https://example.com/design.png",
		"file_size": 102400,
		"file_type": "image/png",
	}, AuthHeader(token))

	if resp.StatusCode != fiber.StatusCreated {
		t.Errorf("Expected 201, got %d. Error: %s", resp.StatusCode, apiResp.Error)
	}

	var attachment AttachmentDTO
	json.Unmarshal(apiResp.Data, &attachment)
	if attachment.FileName != "design.png" {
		t.Errorf("Expected name 'design.png', got '%s'", attachment.FileName)
	}
	if attachment.FileURL != "https://example.com/design.png" {
		t.Errorf("Expected URL, got '%s'", attachment.FileURL)
	}
}

func TestListAttachments_Success(t *testing.T) {
	ta := SetupTestApp(t)
	token, card := setupCardEnv(t, ta, UniqueEmail(), "Uploader")

	ta.MakeRequest(t, "POST", "/api/v1/attachments", map[string]interface{}{
		"card_id": card.ID, "file_name": "doc1.pdf", "file_url": "https://a.com/doc1.pdf",
	}, AuthHeader(token))

	resp, apiResp := ta.MakeRequest(t, "GET", "/api/v1/cards/"+card.ID+"/attachments", nil, AuthHeader(token))

	if resp.StatusCode != fiber.StatusOK {
		t.Errorf("Expected 200, got %d", resp.StatusCode)
	}

	var atts []AttachmentDTO
	json.Unmarshal(apiResp.Data, &atts)
	if len(atts) != 1 {
		t.Errorf("Expected 1 attachment, got %d", len(atts))
	}
}

func TestDeleteAttachment_Success(t *testing.T) {
	ta := SetupTestApp(t)
	token, card := setupCardEnv(t, ta, UniqueEmail(), "Uploader")

	_, aResp := ta.MakeRequest(t, "POST", "/api/v1/attachments", map[string]interface{}{
		"card_id": card.ID, "file_name": "doc1.pdf", "file_url": "https://a.com/doc1.pdf",
	}, AuthHeader(token))
	var att AttachmentDTO
	json.Unmarshal(aResp.Data, &att)

	resp, _ := ta.MakeRequest(t, "DELETE", "/api/v1/attachments/"+att.ID, nil, AuthHeader(token))

	if resp.StatusCode != fiber.StatusOK {
		t.Errorf("Expected 200, got %d", resp.StatusCode)
	}
}
