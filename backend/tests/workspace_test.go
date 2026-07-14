// Package tests provides integration tests for the TaskFlow Workspace API.
package tests

import (
	"encoding/json"
	"fmt"
	"testing"

	"github.com/gofiber/fiber/v2"
)

// ============================================
// TEST: Create Workspace
// ============================================

func TestCreateWorkspace_Success(t *testing.T) {
	ta := SetupTestApp(t)
	email := UniqueEmail()
	token, _ := ta.RegisterAndLogin(t, email, "TestPass123", "WS Creator")

	resp, apiResp := ta.MakeRequest(t, "POST", "/api/v1/workspaces", map[string]interface{}{
		"name":        "My Workspace",
		"description": "A test workspace",
		"slug":        UniqueSlug(),
	}, AuthHeader(token))

	if resp.StatusCode != fiber.StatusCreated {
		t.Errorf("Expected 201, got %d. Error: %s", resp.StatusCode, apiResp.Error)
	}
	if !apiResp.Success {
		t.Errorf("Expected success=true. Error: %s", apiResp.Error)
	}

	var ws WorkspaceDTO
	if err := json.Unmarshal(apiResp.Data, &ws); err != nil {
		t.Fatalf("Failed to parse workspace: %v", err)
	}
	if ws.Name != "My Workspace" {
		t.Errorf("Expected name 'My Workspace', got '%s'", ws.Name)
	}
	if ws.Role != "OWNER" {
		t.Errorf("Expected role 'OWNER', got '%s'", ws.Role)
	}
}

func TestCreateWorkspace_DuplicateSlug(t *testing.T) {
	ta := SetupTestApp(t)
	email := UniqueEmail()
	token, _ := ta.RegisterAndLogin(t, email, "TestPass123", "WS Creator")
	slug := UniqueSlug()

	ta.MakeRequest(t, "POST", "/api/v1/workspaces", map[string]interface{}{
		"name": "Workspace 1",
		"slug": slug,
	}, AuthHeader(token))

	resp, apiResp := ta.MakeRequest(t, "POST", "/api/v1/workspaces", map[string]interface{}{
		"name": "Workspace 2",
		"slug": slug,
	}, AuthHeader(token))

	if resp.StatusCode != fiber.StatusConflict {
		t.Errorf("Expected 409, got %d", resp.StatusCode)
	}
	if apiResp.Success {
		t.Error("Expected success=false for duplicate slug")
	}
}

func TestCreateWorkspace_InvalidSlug(t *testing.T) {
	ta := SetupTestApp(t)
	email := UniqueEmail()
	token, _ := ta.RegisterAndLogin(t, email, "TestPass123", "WS Creator")

	resp, _ := ta.MakeRequest(t, "POST", "/api/v1/workspaces", map[string]interface{}{
		"name": "My Workspace",
		"slug": "INVALID SLUG WITH SPACES",
	}, AuthHeader(token))

	if resp.StatusCode != fiber.StatusUnprocessableEntity {
		t.Errorf("Expected 422 for invalid slug, got %d", resp.StatusCode)
	}
}

func TestCreateWorkspace_MissingName(t *testing.T) {
	ta := SetupTestApp(t)
	email := UniqueEmail()
	token, _ := ta.RegisterAndLogin(t, email, "TestPass123", "WS Creator")

	resp, _ := ta.MakeRequest(t, "POST", "/api/v1/workspaces", map[string]interface{}{
		"slug": UniqueSlug(),
	}, AuthHeader(token))

	if resp.StatusCode != fiber.StatusUnprocessableEntity {
		t.Errorf("Expected 422 for missing name, got %d", resp.StatusCode)
	}
}

func TestCreateWorkspace_NoAuth(t *testing.T) {
	ta := SetupTestApp(t)

	resp, _ := ta.MakeRequest(t, "POST", "/api/v1/workspaces", map[string]interface{}{
		"name": "My Workspace",
		"slug": UniqueSlug(),
	}, nil)

	if resp.StatusCode != fiber.StatusUnauthorized {
		t.Errorf("Expected 401, got %d", resp.StatusCode)
	}
}

// ============================================
// TEST: List Workspaces
// ============================================

func TestListWorkspaces_Success(t *testing.T) {
	ta := SetupTestApp(t)
	email := UniqueEmail()
	token, _ := ta.RegisterAndLogin(t, email, "TestPass123", "List User")

	// Create 2 workspaces
	ta.MakeRequest(t, "POST", "/api/v1/workspaces", map[string]interface{}{
		"name": "Workspace A",
		"slug": UniqueSlug(),
	}, AuthHeader(token))

	ta.MakeRequest(t, "POST", "/api/v1/workspaces", map[string]interface{}{
		"name": "Workspace B",
		"slug": UniqueSlug(),
	}, AuthHeader(token))

	resp, apiResp := ta.MakeRequest(t, "GET", "/api/v1/workspaces", nil, AuthHeader(token))

	if resp.StatusCode != fiber.StatusOK {
		t.Errorf("Expected 200, got %d. Error: %s", resp.StatusCode, apiResp.Error)
	}

	var workspaces []WorkspaceDTO
	if err := json.Unmarshal(apiResp.Data, &workspaces); err != nil {
		t.Fatalf("Failed to parse workspaces: %v", err)
	}
	if len(workspaces) < 2 {
		t.Errorf("Expected at least 2 workspaces, got %d", len(workspaces))
	}
}

func TestListWorkspaces_Empty(t *testing.T) {
	ta := SetupTestApp(t)
	email := UniqueEmail()
	token, _ := ta.RegisterAndLogin(t, email, "TestPass123", "Empty User")

	resp, apiResp := ta.MakeRequest(t, "GET", "/api/v1/workspaces", nil, AuthHeader(token))

	if resp.StatusCode != fiber.StatusOK {
		t.Errorf("Expected 200, got %d", resp.StatusCode)
	}

	var workspaces []WorkspaceDTO
	json.Unmarshal(apiResp.Data, &workspaces)
	if len(workspaces) != 0 {
		t.Errorf("Expected 0 workspaces for new user, got %d", len(workspaces))
	}
}

// ============================================
// TEST: Get Workspace
// ============================================

func TestGetWorkspace_Success(t *testing.T) {
	ta := SetupTestApp(t)
	email := UniqueEmail()
	token, _ := ta.RegisterAndLogin(t, email, "TestPass123", "Get User")

	slug := UniqueSlug()
	_, createResp := ta.MakeRequest(t, "POST", "/api/v1/workspaces", map[string]interface{}{
		"name":        "Get Test WS",
		"description": "Test description",
		"slug":        slug,
	}, AuthHeader(token))

	var created WorkspaceDTO
	json.Unmarshal(createResp.Data, &created)

	resp, apiResp := ta.MakeRequest(t, "GET", "/api/v1/workspaces/"+created.ID, nil, AuthHeader(token))

	if resp.StatusCode != fiber.StatusOK {
		t.Errorf("Expected 200, got %d. Error: %s", resp.StatusCode, apiResp.Error)
	}

	var ws WorkspaceDTO
	json.Unmarshal(apiResp.Data, &ws)
	if ws.Name != "Get Test WS" {
		t.Errorf("Expected name 'Get Test WS', got '%s'", ws.Name)
	}
	if ws.Description == nil || *ws.Description != "Test description" {
		t.Error("Expected description 'Test description'")
	}
}

func TestGetWorkspace_NotMember(t *testing.T) {
	ta := SetupTestApp(t)

	// User A creates workspace
	emailA := UniqueEmail()
	tokenA, _ := ta.RegisterAndLogin(t, emailA, "TestPass123", "Owner A")
	_, createResp := ta.MakeRequest(t, "POST", "/api/v1/workspaces", map[string]interface{}{
		"name": "Owner A Workspace",
		"slug": UniqueSlug(),
	}, AuthHeader(tokenA))
	var created WorkspaceDTO
	json.Unmarshal(createResp.Data, &created)

	// User B tries to access
	emailB := UniqueEmail()
	tokenB, _ := ta.RegisterAndLogin(t, emailB, "TestPass123", "User B")

	resp, _ := ta.MakeRequest(t, "GET", "/api/v1/workspaces/"+created.ID, nil, AuthHeader(tokenB))

	if resp.StatusCode != fiber.StatusForbidden {
		t.Errorf("Expected 403 for non-member, got %d", resp.StatusCode)
	}
}

func TestGetWorkspace_InvalidID(t *testing.T) {
	ta := SetupTestApp(t)
	email := UniqueEmail()
	token, _ := ta.RegisterAndLogin(t, email, "TestPass123", "Test User")

	resp, _ := ta.MakeRequest(t, "GET", "/api/v1/workspaces/invalid-id", nil, AuthHeader(token))

	if resp.StatusCode != fiber.StatusBadRequest {
		t.Errorf("Expected 400 for invalid ID, got %d", resp.StatusCode)
	}
}

// ============================================
// TEST: Update Workspace
// ============================================

func TestUpdateWorkspace_Success(t *testing.T) {
	ta := SetupTestApp(t)
	email := UniqueEmail()
	token, _ := ta.RegisterAndLogin(t, email, "TestPass123", "Update User")

	_, createResp := ta.MakeRequest(t, "POST", "/api/v1/workspaces", map[string]interface{}{
		"name": "Original Name",
		"slug": UniqueSlug(),
	}, AuthHeader(token))
	var created WorkspaceDTO
	json.Unmarshal(createResp.Data, &created)

	resp, apiResp := ta.MakeRequest(t, "PUT", "/api/v1/workspaces/"+created.ID, map[string]interface{}{
		"name": "Updated Name",
		"slug": created.Slug,
	}, AuthHeader(token))

	if resp.StatusCode != fiber.StatusOK {
		t.Errorf("Expected 200, got %d. Error: %s", resp.StatusCode, apiResp.Error)
	}

	var ws WorkspaceDTO
	json.Unmarshal(apiResp.Data, &ws)
	if ws.Name != "Updated Name" {
		t.Errorf("Expected name 'Updated Name', got '%s'", ws.Name)
	}
}

func TestUpdateWorkspace_NotOwnerOrAdmin(t *testing.T) {
	ta := SetupTestApp(t)

	// Owner creates workspace
	emailOwner := UniqueEmail()
	tokenOwner, _ := ta.RegisterAndLogin(t, emailOwner, "TestPass123", "Owner")
	_, createResp := ta.MakeRequest(t, "POST", "/api/v1/workspaces", map[string]interface{}{
		"name": "Owner WS",
		"slug": UniqueSlug(),
	}, AuthHeader(tokenOwner))
	var created WorkspaceDTO
	json.Unmarshal(createResp.Data, &created)

	// Invite Member
	emailMember := UniqueEmail()
	tokenMember, _ := ta.RegisterAndLogin(t, emailMember, "TestPass123", "Member")
	ta.MakeRequest(t, "POST", "/api/v1/workspaces/"+created.ID+"/members", map[string]interface{}{
		"email": emailMember,
		"role":  "MEMBER",
	}, AuthHeader(tokenOwner))

	// Member tries to update
	resp, _ := ta.MakeRequest(t, "PUT", "/api/v1/workspaces/"+created.ID, map[string]interface{}{
		"name": "Hacked Name",
		"slug": created.Slug,
	}, AuthHeader(tokenMember))

	if resp.StatusCode != fiber.StatusForbidden {
		t.Errorf("Expected 403 for MEMBER update, got %d", resp.StatusCode)
	}
}

// ============================================
// TEST: Delete Workspace
// ============================================

func TestDeleteWorkspace_Success(t *testing.T) {
	ta := SetupTestApp(t)
	email := UniqueEmail()
	token, _ := ta.RegisterAndLogin(t, email, "TestPass123", "Delete User")

	_, createResp := ta.MakeRequest(t, "POST", "/api/v1/workspaces", map[string]interface{}{
		"name": "To Delete",
		"slug": UniqueSlug(),
	}, AuthHeader(token))
	var created WorkspaceDTO
	json.Unmarshal(createResp.Data, &created)

	resp, apiResp := ta.MakeRequest(t, "DELETE", "/api/v1/workspaces/"+created.ID, nil, AuthHeader(token))

	if resp.StatusCode != fiber.StatusOK {
		t.Errorf("Expected 200, got %d. Error: %s", resp.StatusCode, apiResp.Error)
	}

	// Verify deleted - CASCADE deletes workspace_members too, so GET returns 403
	getResp, _ := ta.MakeRequest(t, "GET", "/api/v1/workspaces/"+created.ID, nil, AuthHeader(token))
	if getResp.StatusCode != fiber.StatusForbidden && getResp.StatusCode != fiber.StatusNotFound {
		t.Errorf("Expected 403 or 404 after delete, got %d", getResp.StatusCode)
	}
}

func TestDeleteWorkspace_NotOwner(t *testing.T) {
	ta := SetupTestApp(t)

	emailOwner := UniqueEmail()
	tokenOwner, _ := ta.RegisterAndLogin(t, emailOwner, "TestPass123", "Owner")
	_, createResp := ta.MakeRequest(t, "POST", "/api/v1/workspaces", map[string]interface{}{
		"name": "Owner Only WS",
		"slug": UniqueSlug(),
	}, AuthHeader(tokenOwner))
	var created WorkspaceDTO
	json.Unmarshal(createResp.Data, &created)

	// Invite admin
	emailAdmin := UniqueEmail()
	tokenAdmin, _ := ta.RegisterAndLogin(t, emailAdmin, "TestPass123", "Admin")
	ta.MakeRequest(t, "POST", "/api/v1/workspaces/"+created.ID+"/members", map[string]interface{}{
		"email": emailAdmin,
		"role":  "ADMIN",
	}, AuthHeader(tokenOwner))

	// Admin tries to delete
	resp, _ := ta.MakeRequest(t, "DELETE", "/api/v1/workspaces/"+created.ID, nil, AuthHeader(tokenAdmin))

	if resp.StatusCode != fiber.StatusForbidden {
		t.Errorf("Expected 403 for non-owner delete, got %d", resp.StatusCode)
	}
}

// ============================================
// TEST: Member Management
// ============================================

func TestInviteMember_Success(t *testing.T) {
	ta := SetupTestApp(t)

	emailOwner := UniqueEmail()
	tokenOwner, _ := ta.RegisterAndLogin(t, emailOwner, "TestPass123", "Owner")
	_, createResp := ta.MakeRequest(t, "POST", "/api/v1/workspaces", map[string]interface{}{
		"name": "Invite WS",
		"slug": UniqueSlug(),
	}, AuthHeader(tokenOwner))
	var created WorkspaceDTO
	json.Unmarshal(createResp.Data, &created)

	emailMember := UniqueEmail()
	ta.RegisterAndLogin(t, emailMember, "TestPass123", "New Member")

	resp, apiResp := ta.MakeRequest(t, "POST", "/api/v1/workspaces/"+created.ID+"/members", map[string]interface{}{
		"email": emailMember,
		"role":  "MEMBER",
	}, AuthHeader(tokenOwner))

	if resp.StatusCode != fiber.StatusCreated {
		t.Errorf("Expected 201, got %d. Error: %s", resp.StatusCode, apiResp.Error)
	}

	var member WorkspaceMemberDTO
	json.Unmarshal(apiResp.Data, &member)
	if member.Role != "MEMBER" {
		t.Errorf("Expected role 'MEMBER', got '%s'", member.Role)
	}
	if member.Email != emailMember {
		t.Errorf("Expected email %s, got %s", emailMember, member.Email)
	}
}

func TestInviteMember_AlreadyMember(t *testing.T) {
	ta := SetupTestApp(t)

	emailOwner := UniqueEmail()
	tokenOwner, _ := ta.RegisterAndLogin(t, emailOwner, "TestPass123", "Owner")
	_, createResp := ta.MakeRequest(t, "POST", "/api/v1/workspaces", map[string]interface{}{
		"name": "Invite WS",
		"slug": UniqueSlug(),
	}, AuthHeader(tokenOwner))
	var created WorkspaceDTO
	json.Unmarshal(createResp.Data, &created)

	emailMember := UniqueEmail()
	ta.RegisterAndLogin(t, emailMember, "TestPass123", "Member")

	// First invite succeeds
	ta.MakeRequest(t, "POST", "/api/v1/workspaces/"+created.ID+"/members", map[string]interface{}{
		"email": emailMember,
		"role":  "MEMBER",
	}, AuthHeader(tokenOwner))

	// Second invite fails
	resp, _ := ta.MakeRequest(t, "POST", "/api/v1/workspaces/"+created.ID+"/members", map[string]interface{}{
		"email": emailMember,
		"role":  "MEMBER",
	}, AuthHeader(tokenOwner))

	if resp.StatusCode != fiber.StatusConflict {
		t.Errorf("Expected 409 for duplicate member, got %d", resp.StatusCode)
	}
}

func TestInviteMember_NonExistentEmail(t *testing.T) {
	ta := SetupTestApp(t)

	emailOwner := UniqueEmail()
	tokenOwner, _ := ta.RegisterAndLogin(t, emailOwner, "TestPass123", "Owner")
	_, createResp := ta.MakeRequest(t, "POST", "/api/v1/workspaces", map[string]interface{}{
		"name": "Invite WS",
		"slug": UniqueSlug(),
	}, AuthHeader(tokenOwner))
	var created WorkspaceDTO
	json.Unmarshal(createResp.Data, &created)

	resp, _ := ta.MakeRequest(t, "POST", "/api/v1/workspaces/"+created.ID+"/members", map[string]interface{}{
		"email": "nobody@test.com",
		"role":  "MEMBER",
	}, AuthHeader(tokenOwner))

	if resp.StatusCode != fiber.StatusNotFound {
		t.Errorf("Expected 404 for non-existent user, got %d", resp.StatusCode)
	}
}

func TestListMembers_Success(t *testing.T) {
	ta := SetupTestApp(t)

	emailOwner := UniqueEmail()
	tokenOwner, _ := ta.RegisterAndLogin(t, emailOwner, "TestPass123", "Owner")
	_, createResp := ta.MakeRequest(t, "POST", "/api/v1/workspaces", map[string]interface{}{
		"name": "Member List WS",
		"slug": UniqueSlug(),
	}, AuthHeader(tokenOwner))
	var created WorkspaceDTO
	json.Unmarshal(createResp.Data, &created)

	// Invite 2 members
	for i := 0; i < 2; i++ {
		email := fmt.Sprintf("member%d_%s@test.com", i, UniqueSlug())
		ta.RegisterAndLogin(t, email, "TestPass123", fmt.Sprintf("Member %d", i))
		ta.MakeRequest(t, "POST", "/api/v1/workspaces/"+created.ID+"/members", map[string]interface{}{
			"email": email,
			"role":  "MEMBER",
		}, AuthHeader(tokenOwner))
	}

	resp, apiResp := ta.MakeRequest(t, "GET", "/api/v1/workspaces/"+created.ID+"/members", nil, AuthHeader(tokenOwner))

	if resp.StatusCode != fiber.StatusOK {
		t.Errorf("Expected 200, got %d. Error: %s", resp.StatusCode, apiResp.Error)
	}

	var members []WorkspaceMemberDTO
	json.Unmarshal(apiResp.Data, &members)
	if len(members) != 3 {
		t.Errorf("Expected 3 members (1 owner + 2 invited), got %d", len(members))
	}
}

func TestUpdateMemberRole_Success(t *testing.T) {
	ta := SetupTestApp(t)

	emailOwner := UniqueEmail()
	tokenOwner, _ := ta.RegisterAndLogin(t, emailOwner, "TestPass123", "Owner")
	_, createResp := ta.MakeRequest(t, "POST", "/api/v1/workspaces", map[string]interface{}{
		"name": "Role Update WS",
		"slug": UniqueSlug(),
	}, AuthHeader(tokenOwner))
	var created WorkspaceDTO
	json.Unmarshal(createResp.Data, &created)

	emailMember := UniqueEmail()
	ta.RegisterAndLogin(t, emailMember, "TestPass123", "To Promote")
	_, inviteAPI := ta.MakeRequest(t, "POST", "/api/v1/workspaces/"+created.ID+"/members", map[string]interface{}{
		"email": emailMember,
		"role":  "MEMBER",
	}, AuthHeader(tokenOwner))
	var member WorkspaceMemberDTO
	json.Unmarshal(inviteAPI.Data, &member)

	resp, apiResp := ta.MakeRequest(t, "PATCH", "/api/v1/workspaces/"+created.ID+"/members/"+member.UserID, map[string]interface{}{
		"role": "ADMIN",
	}, AuthHeader(tokenOwner))

	if resp.StatusCode != fiber.StatusOK {
		t.Errorf("Expected 200, got %d. Error: %s", resp.StatusCode, apiResp.Error)
	}

	var updated WorkspaceMemberDTO
	json.Unmarshal(apiResp.Data, &updated)
	if updated.Role != "ADMIN" {
		t.Errorf("Expected role 'ADMIN', got '%s'", updated.Role)
	}
}

func TestUpdateMemberRole_NotOwner(t *testing.T) {
	ta := SetupTestApp(t)

	emailOwner := UniqueEmail()
	tokenOwner, _ := ta.RegisterAndLogin(t, emailOwner, "TestPass123", "Owner")
	_, createResp := ta.MakeRequest(t, "POST", "/api/v1/workspaces", map[string]interface{}{
		"name": "Role WS",
		"slug": UniqueSlug(),
	}, AuthHeader(tokenOwner))
	var created WorkspaceDTO
	json.Unmarshal(createResp.Data, &created)

	emailAdmin := UniqueEmail()
	tokenAdmin, _ := ta.RegisterAndLogin(t, emailAdmin, "TestPass123", "Admin")
	ta.MakeRequest(t, "POST", "/api/v1/workspaces/"+created.ID+"/members", map[string]interface{}{
		"email": emailAdmin,
		"role":  "ADMIN",
	}, AuthHeader(tokenOwner))

	emailMember := UniqueEmail()
	ta.RegisterAndLogin(t, emailMember, "TestPass123", "Member")
	_, inviteAPI := ta.MakeRequest(t, "POST", "/api/v1/workspaces/"+created.ID+"/members", map[string]interface{}{
		"email": emailMember,
		"role":  "MEMBER",
	}, AuthHeader(tokenOwner))
	var member WorkspaceMemberDTO
	json.Unmarshal(inviteAPI.Data, &member)

	// Admin tries to change role
	resp, _ := ta.MakeRequest(t, "PATCH", "/api/v1/workspaces/"+created.ID+"/members/"+member.UserID, map[string]interface{}{
		"role": "ADMIN",
	}, AuthHeader(tokenAdmin))

	if resp.StatusCode != fiber.StatusForbidden {
		t.Errorf("Expected 403 for admin role update, got %d", resp.StatusCode)
	}
}

func TestRemoveMember_Success(t *testing.T) {
	ta := SetupTestApp(t)

	emailOwner := UniqueEmail()
	tokenOwner, _ := ta.RegisterAndLogin(t, emailOwner, "TestPass123", "Owner")
	_, createResp := ta.MakeRequest(t, "POST", "/api/v1/workspaces", map[string]interface{}{
		"name": "Remove WS",
		"slug": UniqueSlug(),
	}, AuthHeader(tokenOwner))
	var created WorkspaceDTO
	json.Unmarshal(createResp.Data, &created)

	emailMember := UniqueEmail()
	ta.RegisterAndLogin(t, emailMember, "TestPass123", "To Remove")
	_, inviteAPI := ta.MakeRequest(t, "POST", "/api/v1/workspaces/"+created.ID+"/members", map[string]interface{}{
		"email": emailMember,
		"role":  "MEMBER",
	}, AuthHeader(tokenOwner))
	var member WorkspaceMemberDTO
	json.Unmarshal(inviteAPI.Data, &member)

	resp, apiResp := ta.MakeRequest(t, "DELETE", "/api/v1/workspaces/"+created.ID+"/members/"+member.UserID, nil, AuthHeader(tokenOwner))

	if resp.StatusCode != fiber.StatusOK {
		t.Errorf("Expected 200, got %d. Error: %s", resp.StatusCode, apiResp.Error)
	}

	// Verify removed
	_, getAPI := ta.MakeRequest(t, "GET", "/api/v1/workspaces/"+created.ID+"/members", nil, AuthHeader(tokenOwner))
	var members []WorkspaceMemberDTO
	json.Unmarshal(getAPI.Data, &members)
	for _, m := range members {
		if m.UserID == member.UserID {
			t.Error("Member should have been removed")
		}
	}
}

func TestRemoveMember_NonMember(t *testing.T) {
	ta := SetupTestApp(t)

	emailOwner := UniqueEmail()
	tokenOwner, _ := ta.RegisterAndLogin(t, emailOwner, "TestPass123", "Owner")
	_, createResp := ta.MakeRequest(t, "POST", "/api/v1/workspaces", map[string]interface{}{
		"name": "Remove WS",
		"slug": UniqueSlug(),
	}, AuthHeader(tokenOwner))
	var created WorkspaceDTO
	json.Unmarshal(createResp.Data, &created)

	// Try to remove non-existent user
	resp, _ := ta.MakeRequest(t, "DELETE", "/api/v1/workspaces/"+created.ID+"/members/"+"00000000-0000-0000-0000-000000000000", nil, AuthHeader(tokenOwner))

	if resp.StatusCode != fiber.StatusNotFound {
		t.Errorf("Expected 404 for non-member removal, got %d", resp.StatusCode)
	}
}

func TestOwnerCannotLeave(t *testing.T) {
	ta := SetupTestApp(t)

	emailOwner := UniqueEmail()
	tokenOwner, userData := ta.RegisterAndLogin(t, emailOwner, "TestPass123", "Owner")
	_, createResp := ta.MakeRequest(t, "POST", "/api/v1/workspaces", map[string]interface{}{
		"name": "Owner Leave WS",
		"slug": UniqueSlug(),
	}, AuthHeader(tokenOwner))
	var created WorkspaceDTO
	json.Unmarshal(createResp.Data, &created)

	// Owner tries to remove themselves
	resp, _ := ta.MakeRequest(t, "DELETE", "/api/v1/workspaces/"+created.ID+"/members/"+userData.User.ID, nil, AuthHeader(tokenOwner))

	// Owner cannot leave workspace
	if resp.StatusCode != fiber.StatusForbidden {
		t.Errorf("Expected 403 for owner self-removal, got %d", resp.StatusCode)
	}
}
