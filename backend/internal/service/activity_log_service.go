// Package service implements business logic for activity logs.
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

// ActivityLogService handles fetching activity logs.
// The actual logging happens inside other services via the internal logActivity function.
type ActivityLogService struct {
	queries *repository.Queries
}

// NewActivityLogService creates a new ActivityLogService.
func NewActivityLogService(queries *repository.Queries) *ActivityLogService {
	return &ActivityLogService{queries: queries}
}

// verifyBoardAccess checks if the user is a member of the workspace that owns the board.
func (s *ActivityLogService) verifyBoardAccess(ctx context.Context, userID, boardID uuid.UUID) error {
	board, err := s.queries.GetBoardByID(ctx, boardID)
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

// ListBoardActivities returns activity logs for a specific board.
func (s *ActivityLogService) ListBoardActivities(ctx context.Context, userID, boardID uuid.UUID, limit, offset int32) ([]dto.ActivityLogDTO, error) {
	if err := s.verifyBoardAccess(ctx, userID, boardID); err != nil {
		return nil, err
	}

	// Default limit
	if limit <= 0 || limit > 100 {
		limit = 50
	}

	rows, err := s.queries.ListActivityLogsByBoard(ctx, repository.ListActivityLogsByBoardParams{
		BoardID: boardID,
		Limit:   limit,
		Offset:  offset,
	})
	if err != nil {
		log.Printf("[ERROR] Failed to list board activities: %v", err)
		return nil, fmt.Errorf("failed to list activities")
	}

	return mapActivityLogRows(rows), nil
}

// ListCardActivities returns activity logs specific to a card.
func (s *ActivityLogService) ListCardActivities(ctx context.Context, userID, cardID uuid.UUID, limit, offset int32) ([]dto.ActivityLogDTO, error) {
	// Need to check board access via card -> list -> board
	card, err := s.queries.GetCardByID(ctx, cardID)
	if err != nil {
		return nil, fmt.Errorf("card not found")
	}

	list, err := s.queries.GetListByID(ctx, card.ListID)
	if err != nil {
		return nil, fmt.Errorf("list not found")
	}

	if err := s.verifyBoardAccess(ctx, userID, list.BoardID); err != nil {
		return nil, err
	}

	// Default limit
	if limit <= 0 || limit > 100 {
		limit = 50
	}

	rows, err := s.queries.ListActivityLogsByCard(ctx, repository.ListActivityLogsByCardParams{
		CardID: uuid.NullUUID{UUID: cardID, Valid: true},
		Limit:  limit,
		Offset: offset,
	})
	if err != nil {
		log.Printf("[ERROR] Failed to list card activities: %v", err)
		return nil, fmt.Errorf("failed to list activities")
	}

	return mapActivityLogRowsCard(rows), nil
}

// Internal helper functions to map rows to DTOs
func mapActivityLogRows(rows []repository.ListActivityLogsByBoardRow) []dto.ActivityLogDTO {
	var logs []dto.ActivityLogDTO
	for _, row := range rows {
		var avatarURL *string
		if row.UserAvatarUrl.Valid {
			avatarURL = &row.UserAvatarUrl.String
		}
		var cardID *string
		if row.CardID.Valid {
			cid := row.CardID.UUID.String()
			cardID = &cid
		}
		var details *string
		if row.Details.Valid {
			details = &row.Details.String
		}

		logs = append(logs, dto.ActivityLogDTO{
			ID:          row.ID.String(),
			BoardID:     row.BoardID.String(),
			CardID:      cardID,
			UserID:      row.UserID.String(),
			UserName:    row.UserName,
			AvatarURL:   avatarURL,
			Action:      row.Action,
			EntityType:  row.EntityType,
			EntityTitle: row.EntityTitle,
			Details:     details,
			CreatedAt:   row.CreatedAt.Format(time.RFC3339),
		})
	}
	return logs
}

func mapActivityLogRowsCard(rows []repository.ListActivityLogsByCardRow) []dto.ActivityLogDTO {
	var logs []dto.ActivityLogDTO
	for _, row := range rows {
		var avatarURL *string
		if row.UserAvatarUrl.Valid {
			avatarURL = &row.UserAvatarUrl.String
		}
		var cardID *string
		if row.CardID.Valid {
			cid := row.CardID.UUID.String()
			cardID = &cid
		}
		var details *string
		if row.Details.Valid {
			details = &row.Details.String
		}

		logs = append(logs, dto.ActivityLogDTO{
			ID:          row.ID.String(),
			BoardID:     row.BoardID.String(),
			CardID:      cardID,
			UserID:      row.UserID.String(),
			UserName:    row.UserName,
			AvatarURL:   avatarURL,
			Action:      row.Action,
			EntityType:  row.EntityType,
			EntityTitle: row.EntityTitle,
			Details:     details,
			CreatedAt:   row.CreatedAt.Format(time.RFC3339),
		})
	}
	return logs
}

// ==========================================
// INTERNAL LOGGING HELPER
// Used by other services (CardService, ListService)
// ==========================================

// logActivity is an internal helper used across services to log user actions.
// We use a fire-and-forget approach (goroutine) so it doesn't block the main API response.
func logActivity(queries *repository.Queries, boardID uuid.UUID, cardID *uuid.UUID, userID uuid.UUID, action, entityType, entityTitle, details string) {
	go func() {
		// Create an independent context with timeout for background task
		ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()

		var cID uuid.NullUUID
		if cardID != nil {
			cID = uuid.NullUUID{UUID: *cardID, Valid: true}
		}

		var d sql.NullString
		if details != "" {
			d = sql.NullString{String: details, Valid: true}
		}

		err := queries.CreateActivityLog(ctx, repository.CreateActivityLogParams{
			ID:          uuid.New(),
			BoardID:     boardID,
			CardID:      cID,
			UserID:      userID,
			Action:      action,
			EntityType:  entityType,
			EntityTitle: entityTitle,
			Details:     d,
		})

		if err != nil {
			log.Printf("[ERROR] Failed to save activity log: %v", err)
		}
	}()
}
