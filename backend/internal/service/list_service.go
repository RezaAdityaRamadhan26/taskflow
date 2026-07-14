// Package service implements the business logic for list management.
package service

import (
	"context"
	"fmt"
	"log"
	"time"

	"github.com/google/uuid"

	"github.com/RezaAdityaRamadhan26/taskflow/backend/internal/dto"
	"github.com/RezaAdityaRamadhan26/taskflow/backend/internal/repository"
)

// ListService handles list business logic.
type ListService struct {
	queries *repository.Queries
}

// NewListService creates a new ListService instance.
func NewListService(queries *repository.Queries) *ListService {
	return &ListService{queries: queries}
}

// verifyBoardAccess checks if the user has access to the board, and optionally if they have write access.
func (s *ListService) verifyBoardAccess(ctx context.Context, userID, boardID uuid.UUID, requireWrite bool) (*repository.Board, error) {
	board, err := s.queries.GetBoardByID(ctx, boardID)
	if err != nil {
		return nil, fmt.Errorf("board not found")
	}

	member, err := s.queries.GetWorkspaceMember(ctx, repository.GetWorkspaceMemberParams{
		UserID:      userID,
		WorkspaceID: board.WorkspaceID,
	})
	if err != nil {
		return nil, fmt.Errorf("you are not a member of this workspace")
	}

	if requireWrite && member.Role != repository.RoleOWNER && member.Role != repository.RoleADMIN {
		// Only OWNER and ADMIN can edit boards/lists right now
		// Future enhancement: Allow MEMBER to edit if they are explicitly invited to the board (if board permissions are granular)
		// For MVP, workspace OWNER/ADMIN manage board structure, MEMBER manages cards.
		return nil, fmt.Errorf("only OWNER or ADMIN can modify lists")
	}

	return &board, nil
}

// CreateList creates a new list in a board.
func (s *ListService) CreateList(ctx context.Context, userID uuid.UUID, req dto.CreateListRequest) (*dto.ListDTO, error) {
	boardID, err := uuid.Parse(req.BoardID)
	if err != nil {
		return nil, fmt.Errorf("invalid board ID")
	}

	// Verify write access
	if _, err := s.verifyBoardAccess(ctx, userID, boardID, true); err != nil {
		return nil, err
	}

	// Calculate position
	position := float64(0)
	if req.Position != nil {
		position = *req.Position
	} else {
		// Get max position and add 65536.0 (standard Trello-like spacing)
		maxPos, err := s.queries.GetMaxPositionForBoard(ctx, boardID)
		if err != nil {
			maxPos = 0
		}
		position = maxPos + 65536.0
	}

	list, err := s.queries.CreateList(ctx, repository.CreateListParams{
		ID:       uuid.New(),
		BoardID:  boardID,
		Name:     req.Name,
		Position: position,
	})
	if err != nil {
		log.Printf("[ERROR] Failed to create list: %v", err)
		return nil, fmt.Errorf("failed to create list")
	}

	logActivity(s.queries, list.BoardID, nil, userID, "CREATED_LIST", "LIST", list.Name, "")

	return mapListToDTO(list), nil
}

// ListLists returns all lists in a board.
func (s *ListService) ListLists(ctx context.Context, userID, boardID uuid.UUID) ([]dto.ListDTO, error) {
	// Verify read access
	if _, err := s.verifyBoardAccess(ctx, userID, boardID, false); err != nil {
		return nil, err
	}

	rows, err := s.queries.ListListsByBoard(ctx, boardID)
	if err != nil {
		log.Printf("[ERROR] Failed to list lists: %v", err)
		return nil, fmt.Errorf("failed to list lists")
	}

	lists := make([]dto.ListDTO, 0, len(rows))
	for _, row := range rows {
		lists = append(lists, *mapListToDTO(repository.List{
			ID:        row.ID,
			BoardID:   row.BoardID,
			Name:      row.Name,
			Position:  row.Position,
			CreatedAt: row.CreatedAt,
			UpdatedAt: row.UpdatedAt,
		}))
	}

	return lists, nil
}

// GetList returns a list by ID.
func (s *ListService) GetList(ctx context.Context, userID, listID uuid.UUID) (*dto.ListDTO, error) {
	list, err := s.queries.GetListByID(ctx, listID)
	if err != nil {
		return nil, fmt.Errorf("list not found")
	}

	// Verify read access
	if _, err := s.verifyBoardAccess(ctx, userID, list.BoardID, false); err != nil {
		return nil, err
	}

	return mapListToDTO(list), nil
}

// UpdateList updates a list's name and/or position.
func (s *ListService) UpdateList(ctx context.Context, userID, listID uuid.UUID, req dto.UpdateListRequest) (*dto.ListDTO, error) {
	list, err := s.queries.GetListByID(ctx, listID)
	if err != nil {
		return nil, fmt.Errorf("list not found")
	}

	// Verify write access
	if _, err := s.verifyBoardAccess(ctx, userID, list.BoardID, true); err != nil {
		return nil, err
	}

	updatedList, err := s.queries.UpdateList(ctx, repository.UpdateListParams{
		ID:       listID,
		Name:     req.Name,
		Position: req.Position,
	})
	if err != nil {
		log.Printf("[ERROR] Failed to update list: %v", err)
		return nil, fmt.Errorf("failed to update list")
	}

	action := "UPDATED_LIST"
	if list.Name != req.Name {
		action = "RENAMED_LIST"
	} else if list.Position != req.Position {
		action = "MOVED_LIST"
	}
	logActivity(s.queries, list.BoardID, nil, userID, action, "LIST", updatedList.Name, "")

	return mapListToDTO(updatedList), nil
}

// DeleteList deletes a list and all its cards (via CASCADE).
func (s *ListService) DeleteList(ctx context.Context, userID, listID uuid.UUID) error {
	list, err := s.queries.GetListByID(ctx, listID)
	if err != nil {
		return fmt.Errorf("list not found")
	}

	// Verify write access
	if _, err := s.verifyBoardAccess(ctx, userID, list.BoardID, true); err != nil {
		return err
	}

	if err := s.queries.DeleteList(ctx, listID); err != nil {
		log.Printf("[ERROR] Failed to delete list: %v", err)
		return fmt.Errorf("failed to delete list")
	}

	logActivity(s.queries, list.BoardID, nil, userID, "DELETED_LIST", "LIST", list.Name, "")

	return nil
}

// mapListToDTO maps a repository.List to dto.ListDTO.
func mapListToDTO(list repository.List) *dto.ListDTO {
	return &dto.ListDTO{
		ID:        list.ID.String(),
		BoardID:   list.BoardID.String(),
		Name:      list.Name,
		Position:  list.Position,
		CreatedAt: list.CreatedAt.Format(time.RFC3339),
		UpdatedAt: list.UpdatedAt.Format(time.RFC3339),
	}
}
