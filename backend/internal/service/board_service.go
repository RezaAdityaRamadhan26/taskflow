// Package service implements the business logic for board management.
package service

import (
	"context"
	"database/sql"
	"fmt"
	"log"
	"time"

	"github.com/google/uuid"

	"github.com/RezaAdityaRamadhan26/taskflow/backend/internal/dto"
	"github.com/RezaAdityaRamadhan26/taskflow/backend/internal/repository"
)

// BoardService handles board business logic.
type BoardService struct {
	queries *repository.Queries
}

// NewBoardService creates a new BoardService instance.
func NewBoardService(queries *repository.Queries) *BoardService {
	return &BoardService{queries: queries}
}

// CreateBoard creates a new board. Only OWNER or ADMIN can create boards in a workspace.
func (s *BoardService) CreateBoard(ctx context.Context, userID uuid.UUID, req dto.CreateBoardRequest) (*dto.BoardDTO, error) {
	workspaceID, err := uuid.Parse(req.WorkspaceID)
	if err != nil {
		return nil, fmt.Errorf("invalid workspace ID")
	}

	// Verify user is a member of the workspace
	member, err := s.queries.GetWorkspaceMember(ctx, repository.GetWorkspaceMemberParams{
		UserID:      userID,
		WorkspaceID: workspaceID,
	})
	if err != nil {
		return nil, fmt.Errorf("you are not a member of this workspace")
	}

	// Only OWNER and ADMIN can create boards
	if member.Role != repository.RoleOWNER && member.Role != repository.RoleADMIN {
		return nil, fmt.Errorf("only OWNER or ADMIN can create boards")
	}

	// Prepare nullable fields
	var description, color sql.NullString
	if req.Description != nil && *req.Description != "" {
		description = sql.NullString{String: *req.Description, Valid: true}
	}
	if req.Color != nil && *req.Color != "" {
		color = sql.NullString{String: *req.Color, Valid: true}
	}

	// Create board
	board, err := s.queries.CreateBoard(ctx, repository.CreateBoardParams{
		ID:          uuid.New(),
		WorkspaceID: workspaceID,
		Name:        req.Name,
		Description: description,
		Color:       color,
	})
	if err != nil {
		log.Printf("[ERROR] Failed to create board: %v", err)
		return nil, fmt.Errorf("failed to create board")
	}

	logActivity(s.queries, board.ID, nil, userID, "CREATED_BOARD", "BOARD", board.Name, "")

	return mapBoardToDTO(board), nil
}

// ListBoards returns all boards in a workspace. All members can view.
func (s *BoardService) ListBoards(ctx context.Context, userID, workspaceID uuid.UUID) ([]dto.BoardDTO, error) {
	// Verify membership
	_, err := s.queries.GetWorkspaceMember(ctx, repository.GetWorkspaceMemberParams{
		UserID:      userID,
		WorkspaceID: workspaceID,
	})
	if err != nil {
		return nil, fmt.Errorf("you are not a member of this workspace")
	}

	rows, err := s.queries.ListBoardsByWorkspace(ctx, workspaceID)
	if err != nil {
		log.Printf("[ERROR] Failed to list boards: %v", err)
		return nil, fmt.Errorf("failed to list boards")
	}

	boards := make([]dto.BoardDTO, 0, len(rows))
	for _, row := range rows {
		// Use type conversion logic matching generated code structure
		boards = append(boards, *mapBoardToDTO(repository.Board{
			ID:          row.ID,
			WorkspaceID: row.WorkspaceID,
			Name:        row.Name,
			Description: row.Description,
			Color:       row.Color,
			CreatedAt:   row.CreatedAt,
			UpdatedAt:   row.UpdatedAt,
		}))
	}

	return boards, nil
}

// GetBoard returns a board by ID. All members can view.
func (s *BoardService) GetBoard(ctx context.Context, userID, boardID uuid.UUID) (*dto.BoardDTO, error) {
	board, err := s.queries.GetBoardByID(ctx, boardID)
	if err != nil {
		return nil, fmt.Errorf("board not found")
	}

	// Verify membership
	_, err = s.queries.GetWorkspaceMember(ctx, repository.GetWorkspaceMemberParams{
		UserID:      userID,
		WorkspaceID: board.WorkspaceID,
	})
	if err != nil {
		return nil, fmt.Errorf("you are not a member of this workspace")
	}

	return mapBoardToDTO(board), nil
}

// UpdateBoard updates a board. Only OWNER or ADMIN can update.
func (s *BoardService) UpdateBoard(ctx context.Context, userID, boardID uuid.UUID, req dto.UpdateBoardRequest) (*dto.BoardDTO, error) {
	board, err := s.queries.GetBoardByID(ctx, boardID)
	if err != nil {
		return nil, fmt.Errorf("board not found")
	}

	// Verify membership and role
	member, err := s.queries.GetWorkspaceMember(ctx, repository.GetWorkspaceMemberParams{
		UserID:      userID,
		WorkspaceID: board.WorkspaceID,
	})
	if err != nil {
		return nil, fmt.Errorf("you are not a member of this workspace")
	}

	if member.Role != repository.RoleOWNER && member.Role != repository.RoleADMIN {
		return nil, fmt.Errorf("only OWNER or ADMIN can update boards")
	}

	// Prepare nullable fields
	var description, color sql.NullString
	if req.Description != nil && *req.Description != "" {
		description = sql.NullString{String: *req.Description, Valid: true}
	}
	if req.Color != nil && *req.Color != "" {
		color = sql.NullString{String: *req.Color, Valid: true}
	}

	updatedBoard, err := s.queries.UpdateBoard(ctx, repository.UpdateBoardParams{
		ID:          boardID,
		Name:        req.Name,
		Description: description,
		Color:       color,
	})
	if err != nil {
		log.Printf("[ERROR] Failed to update board: %v", err)
		return nil, fmt.Errorf("failed to update board")
	}

	logActivity(s.queries, boardID, nil, userID, "UPDATED_BOARD", "BOARD", updatedBoard.Name, "")

	return mapBoardToDTO(updatedBoard), nil
}

// DeleteBoard deletes a board. Only OWNER or ADMIN can delete.
func (s *BoardService) DeleteBoard(ctx context.Context, userID, boardID uuid.UUID) error {
	board, err := s.queries.GetBoardByID(ctx, boardID)
	if err != nil {
		return fmt.Errorf("board not found")
	}

	// Verify membership and role
	member, err := s.queries.GetWorkspaceMember(ctx, repository.GetWorkspaceMemberParams{
		UserID:      userID,
		WorkspaceID: board.WorkspaceID,
	})
	if err != nil {
		return fmt.Errorf("you are not a member of this workspace")
	}

	if member.Role != repository.RoleOWNER && member.Role != repository.RoleADMIN {
		return fmt.Errorf("only OWNER or ADMIN can delete boards")
	}

	if err := s.queries.DeleteBoard(ctx, boardID); err != nil {
		log.Printf("[ERROR] Failed to delete board: %v", err)
		return fmt.Errorf("failed to delete board")
	}

	return nil
}

// mapBoardToDTO maps a repository.Board to dto.BoardDTO.
func mapBoardToDTO(board repository.Board) *dto.BoardDTO {
	var desc *string
	if board.Description.Valid {
		desc = &board.Description.String
	}
	var color *string
	if board.Color.Valid {
		color = &board.Color.String
	}

	return &dto.BoardDTO{
		ID:          board.ID.String(),
		WorkspaceID: board.WorkspaceID.String(),
		Name:        board.Name,
		Description: desc,
		Color:       color,
		CreatedAt:   board.CreatedAt.Format(time.RFC3339),
		UpdatedAt:   board.UpdatedAt.Format(time.RFC3339),
	}
}
