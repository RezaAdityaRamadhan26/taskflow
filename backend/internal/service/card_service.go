// Package service implements the business logic for card management.
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

// CardService handles card business logic.
type CardService struct {
	queries *repository.Queries
}

// NewCardService creates a new CardService instance.
func NewCardService(queries *repository.Queries) *CardService {
	return &CardService{queries: queries}
}

// verifyBoardAccess checks if the user is a member of the workspace that owns the board/list.
// For cards, any MEMBER can create, edit, or delete cards.
func (s *CardService) verifyBoardAccess(ctx context.Context, userID, listID uuid.UUID) error {
	list, err := s.queries.GetListByID(ctx, listID)
	if err != nil {
		return fmt.Errorf("list not found")
	}

	board, err := s.queries.GetBoardByID(ctx, list.BoardID)
	if err != nil {
		return fmt.Errorf("board not found")
	}

	_, err = s.queries.GetWorkspaceMember(ctx, repository.GetWorkspaceMemberParams{
		UserID:      userID,
		WorkspaceID: board.WorkspaceID,
	})
	if err != nil {
		return fmt.Errorf("you do not have access to this board")
	}

	return nil
}

// CreateCard creates a new card in a list.
func (s *CardService) CreateCard(ctx context.Context, userID uuid.UUID, req dto.CreateCardRequest) (*dto.CardDTO, error) {
	listID, err := uuid.Parse(req.ListID)
	if err != nil {
		return nil, fmt.Errorf("invalid list ID")
	}

	// Verify access
	if err := s.verifyBoardAccess(ctx, userID, listID); err != nil {
		return nil, err
	}

	// Calculate position
	position := float64(0)
	if req.Position != nil {
		position = *req.Position
	} else {
		// Get max position and add 65536.0
		maxPos, err := s.queries.GetMaxPositionForList(ctx, listID)
		if err != nil {
			maxPos = 0
		}
		position = maxPos + 65536.0
	}

	// Parse priority
	priority := repository.PriorityMEDIUM
	if req.Priority != nil {
		priority = repository.Priority(*req.Priority)
	}

	var description sql.NullString
	if req.Description != nil && *req.Description != "" {
		description = sql.NullString{String: *req.Description, Valid: true}
	}

	var dueDate sql.NullTime
	if req.DueDate != nil {
		dueDate = sql.NullTime{Time: *req.DueDate, Valid: true}
	}

	card, err := s.queries.CreateCard(ctx, repository.CreateCardParams{
		ID:          uuid.New(),
		ListID:      listID,
		Title:       req.Title,
		Description: description,
		Position:    position,
		Priority:    priority,
		DueDate:     dueDate,
	})
	if err != nil {
		log.Printf("[ERROR] Failed to create card: %v", err)
		return nil, fmt.Errorf("failed to create card")
	}

	return mapCardToDTO(card), nil
}

// ListCards returns all cards in a list.
func (s *CardService) ListCards(ctx context.Context, userID, listID uuid.UUID) ([]dto.CardDTO, error) {
	// Verify access
	if err := s.verifyBoardAccess(ctx, userID, listID); err != nil {
		return nil, err
	}

	rows, err := s.queries.ListCardsByList(ctx, listID)
	if err != nil {
		log.Printf("[ERROR] Failed to list cards: %v", err)
		return nil, fmt.Errorf("failed to list cards")
	}

	cards := make([]dto.CardDTO, 0, len(rows))
	for _, row := range rows {
		cards = append(cards, *mapCardToDTO(repository.Card{
			ID:          row.ID,
			ListID:      row.ListID,
			Title:       row.Title,
			Description: row.Description,
			Position:    row.Position,
			Priority:    row.Priority,
			DueDate:     row.DueDate,
			CreatedAt:   row.CreatedAt,
			UpdatedAt:   row.UpdatedAt,
		}))
	}

	return cards, nil
}

// GetCard returns a card by ID.
func (s *CardService) GetCard(ctx context.Context, userID, cardID uuid.UUID) (*dto.CardDTO, error) {
	card, err := s.queries.GetCardByID(ctx, cardID)
	if err != nil {
		return nil, fmt.Errorf("card not found")
	}

	// Verify access via the list it belongs to
	if err := s.verifyBoardAccess(ctx, userID, card.ListID); err != nil {
		return nil, err
	}

	return mapCardToDTO(card), nil
}

// UpdateCard updates a card's fields, including moving it to a different list.
func (s *CardService) UpdateCard(ctx context.Context, userID, cardID uuid.UUID, req dto.UpdateCardRequest) (*dto.CardDTO, error) {
	card, err := s.queries.GetCardByID(ctx, cardID)
	if err != nil {
		return nil, fmt.Errorf("card not found")
	}

	// Verify access to current list
	if err := s.verifyBoardAccess(ctx, userID, card.ListID); err != nil {
		return nil, err
	}

	newListID, err := uuid.Parse(req.ListID)
	if err != nil {
		return nil, fmt.Errorf("invalid list ID")
	}

	// If moving to a new list, verify access to the target list too
	if newListID != card.ListID {
		if err := s.verifyBoardAccess(ctx, userID, newListID); err != nil {
			return nil, fmt.Errorf("target list not found or access denied")
		}
	}

	var description sql.NullString
	if req.Description != nil && *req.Description != "" {
		description = sql.NullString{String: *req.Description, Valid: true}
	}

	var dueDate sql.NullTime
	if req.DueDate != nil {
		dueDate = sql.NullTime{Time: *req.DueDate, Valid: true}
	}

	updatedCard, err := s.queries.UpdateCard(ctx, repository.UpdateCardParams{
		ID:          cardID,
		ListID:      newListID,
		Title:       req.Title,
		Description: description,
		Position:    req.Position,
		Priority:    repository.Priority(req.Priority),
		DueDate:     dueDate,
	})
	if err != nil {
		log.Printf("[ERROR] Failed to update card: %v", err)
		return nil, fmt.Errorf("failed to update card")
	}

	return mapCardToDTO(updatedCard), nil
}

// DeleteCard deletes a card.
func (s *CardService) DeleteCard(ctx context.Context, userID, cardID uuid.UUID) error {
	card, err := s.queries.GetCardByID(ctx, cardID)
	if err != nil {
		return fmt.Errorf("card not found")
	}

	// Verify access
	if err := s.verifyBoardAccess(ctx, userID, card.ListID); err != nil {
		return err
	}

	if err := s.queries.DeleteCard(ctx, cardID); err != nil {
		log.Printf("[ERROR] Failed to delete card: %v", err)
		return fmt.Errorf("failed to delete card")
	}

	return nil
}

// mapCardToDTO maps a repository.Card to dto.CardDTO.
func mapCardToDTO(card repository.Card) *dto.CardDTO {
	var desc *string
	if card.Description.Valid {
		desc = &card.Description.String
	}

	var dueDate *time.Time
	if card.DueDate.Valid {
		dueDate = &card.DueDate.Time
	}

	return &dto.CardDTO{
		ID:          card.ID.String(),
		ListID:      card.ListID.String(),
		Title:       card.Title,
		Description: desc,
		Position:    card.Position,
		Priority:    string(card.Priority),
		DueDate:     dueDate,
		CreatedAt:   card.CreatedAt.Format(time.RFC3339),
		UpdatedAt:   card.UpdatedAt.Format(time.RFC3339),
	}
}
