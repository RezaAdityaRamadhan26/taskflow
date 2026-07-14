// Package service implements the business logic for comments and attachments.
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

// CardExtrasService handles comments and attachments logic.
type CardExtrasService struct {
	queries *repository.Queries
}

// NewCardExtrasService creates a new CardExtrasService instance.
func NewCardExtrasService(queries *repository.Queries) *CardExtrasService {
	return &CardExtrasService{queries: queries}
}

// verifyCardAccess checks if the user is a member of the workspace that owns the card.
func (s *CardExtrasService) verifyCardAccess(ctx context.Context, userID, cardID uuid.UUID) error {
	card, err := s.queries.GetCardByID(ctx, cardID)
	if err != nil {
		return fmt.Errorf("card not found")
	}

	list, err := s.queries.GetListByID(ctx, card.ListID)
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
		return fmt.Errorf("you do not have access to this card")
	}

	return nil
}

// CreateComment adds a comment to a card.
func (s *CardExtrasService) CreateComment(ctx context.Context, userID uuid.UUID, req dto.CreateCommentRequest) (*dto.CommentDTO, error) {
	cardID, err := uuid.Parse(req.CardID)
	if err != nil {
		return nil, fmt.Errorf("invalid card ID")
	}

	if err := s.verifyCardAccess(ctx, userID, cardID); err != nil {
		return nil, err
	}

	comment, err := s.queries.CreateComment(ctx, repository.CreateCommentParams{
		ID:      uuid.New(),
		CardID:  cardID,
		UserID:  userID,
		Content: req.Content,
	})
	if err != nil {
		log.Printf("[ERROR] Failed to create comment: %v", err)
		return nil, fmt.Errorf("failed to create comment")
	}

	// Fetch user info for DTO
	user, _ := s.queries.GetUserByID(ctx, userID)
	var avatarURL *string
	if user.AvatarUrl.Valid {
		avatarURL = &user.AvatarUrl.String
	}

	return &dto.CommentDTO{
		ID:        comment.ID.String(),
		CardID:    comment.CardID.String(),
		UserID:    comment.UserID.String(),
		UserName:  user.Name,
		AvatarURL: avatarURL,
		Content:   comment.Content,
		CreatedAt: comment.CreatedAt.Format(time.RFC3339),
		UpdatedAt: comment.UpdatedAt.Format(time.RFC3339),
	}, nil
}

// ListComments lists all comments for a card.
func (s *CardExtrasService) ListComments(ctx context.Context, userID, cardID uuid.UUID) ([]dto.CommentDTO, error) {
	if err := s.verifyCardAccess(ctx, userID, cardID); err != nil {
		return nil, err
	}

	rows, err := s.queries.ListCommentsByCard(ctx, cardID)
	if err != nil {
		log.Printf("[ERROR] Failed to list comments: %v", err)
		return nil, fmt.Errorf("failed to list comments")
	}

	comments := make([]dto.CommentDTO, 0, len(rows))
	for _, row := range rows {
		var avatarURL *string
		if row.UserAvatarUrl.Valid {
			avatarURL = &row.UserAvatarUrl.String
		}
		comments = append(comments, dto.CommentDTO{
			ID:        row.ID.String(),
			CardID:    row.CardID.String(),
			UserID:    row.UserID.String(),
			UserName:  row.UserName,
			AvatarURL: avatarURL,
			Content:   row.Content,
			CreatedAt: row.CreatedAt.Format(time.RFC3339),
			UpdatedAt: row.UpdatedAt.Format(time.RFC3339),
		})
	}

	return comments, nil
}

// UpdateComment updates a comment's content. Only the author can update.
func (s *CardExtrasService) UpdateComment(ctx context.Context, userID, commentID uuid.UUID, req dto.UpdateCommentRequest) (*dto.CommentDTO, error) {
	comment, err := s.queries.GetCommentByID(ctx, commentID)
	if err != nil {
		return nil, fmt.Errorf("comment not found")
	}

	if comment.UserID != userID {
		return nil, fmt.Errorf("you can only edit your own comments")
	}

	updatedComment, err := s.queries.UpdateComment(ctx, repository.UpdateCommentParams{
		ID:      commentID,
		Content: req.Content,
	})
	if err != nil {
		log.Printf("[ERROR] Failed to update comment: %v", err)
		return nil, fmt.Errorf("failed to update comment")
	}

	user, _ := s.queries.GetUserByID(ctx, userID)
	var avatarURL *string
	if user.AvatarUrl.Valid {
		avatarURL = &user.AvatarUrl.String
	}

	return &dto.CommentDTO{
		ID:        updatedComment.ID.String(),
		CardID:    updatedComment.CardID.String(),
		UserID:    updatedComment.UserID.String(),
		UserName:  user.Name,
		AvatarURL: avatarURL,
		Content:   updatedComment.Content,
		CreatedAt: updatedComment.CreatedAt.Format(time.RFC3339),
		UpdatedAt: updatedComment.UpdatedAt.Format(time.RFC3339),
	}, nil
}

// DeleteComment deletes a comment. Only the author can delete.
func (s *CardExtrasService) DeleteComment(ctx context.Context, userID, commentID uuid.UUID) error {
	comment, err := s.queries.GetCommentByID(ctx, commentID)
	if err != nil {
		return fmt.Errorf("comment not found")
	}

	if comment.UserID != userID {
		return fmt.Errorf("you can only delete your own comments")
	}

	if err := s.queries.DeleteComment(ctx, commentID); err != nil {
		log.Printf("[ERROR] Failed to delete comment: %v", err)
		return fmt.Errorf("failed to delete comment")
	}

	return nil
}

// CreateAttachment adds an attachment to a card.
func (s *CardExtrasService) CreateAttachment(ctx context.Context, userID uuid.UUID, req dto.CreateAttachmentRequest) (*dto.AttachmentDTO, error) {
	cardID, err := uuid.Parse(req.CardID)
	if err != nil {
		return nil, fmt.Errorf("invalid card ID")
	}

	if err := s.verifyCardAccess(ctx, userID, cardID); err != nil {
		return nil, err
	}

	var fileSize sql.NullInt32
	if req.FileSize != nil {
		fileSize = sql.NullInt32{Int32: *req.FileSize, Valid: true}
	}

	var fileType sql.NullString
	if req.FileType != nil && *req.FileType != "" {
		fileType = sql.NullString{String: *req.FileType, Valid: true}
	}

	attachment, err := s.queries.CreateAttachment(ctx, repository.CreateAttachmentParams{
		ID:       uuid.New(),
		CardID:   cardID,
		UserID:   userID,
		FileName: req.FileName,
		FileUrl:  req.FileURL,
		FileSize: fileSize,
		FileType: fileType,
	})
	if err != nil {
		log.Printf("[ERROR] Failed to create attachment: %v", err)
		return nil, fmt.Errorf("failed to create attachment")
	}

	user, _ := s.queries.GetUserByID(ctx, userID)
	var avatarURL *string
	if user.AvatarUrl.Valid {
		avatarURL = &user.AvatarUrl.String
	}

	var fs *int32
	if attachment.FileSize.Valid {
		fs = &attachment.FileSize.Int32
	}
	var ft *string
	if attachment.FileType.Valid {
		ft = &attachment.FileType.String
	}

	return &dto.AttachmentDTO{
		ID:        attachment.ID.String(),
		CardID:    attachment.CardID.String(),
		UserID:    attachment.UserID.String(),
		UserName:  user.Name,
		AvatarURL: avatarURL,
		FileName:  attachment.FileName,
		FileURL:   attachment.FileUrl,
		FileSize:  fs,
		FileType:  ft,
		CreatedAt: attachment.CreatedAt.Format(time.RFC3339),
	}, nil
}

// ListAttachments lists all attachments for a card.
func (s *CardExtrasService) ListAttachments(ctx context.Context, userID, cardID uuid.UUID) ([]dto.AttachmentDTO, error) {
	if err := s.verifyCardAccess(ctx, userID, cardID); err != nil {
		return nil, err
	}

	rows, err := s.queries.ListAttachmentsByCard(ctx, cardID)
	if err != nil {
		log.Printf("[ERROR] Failed to list attachments: %v", err)
		return nil, fmt.Errorf("failed to list attachments")
	}

	attachments := make([]dto.AttachmentDTO, 0, len(rows))
	for _, row := range rows {
		var avatarURL *string
		if row.UserAvatarUrl.Valid {
			avatarURL = &row.UserAvatarUrl.String
		}

		var fs *int32
		if row.FileSize.Valid {
			fs = &row.FileSize.Int32
		}
		var ft *string
		if row.FileType.Valid {
			ft = &row.FileType.String
		}

		attachments = append(attachments, dto.AttachmentDTO{
			ID:        row.ID.String(),
			CardID:    row.CardID.String(),
			UserID:    row.UserID.String(),
			UserName:  row.UserName,
			AvatarURL: avatarURL,
			FileName:  row.FileName,
			FileURL:   row.FileUrl,
			FileSize:  fs,
			FileType:  ft,
			CreatedAt: row.CreatedAt.Format(time.RFC3339),
		})
	}

	return attachments, nil
}

// DeleteAttachment deletes an attachment.
func (s *CardExtrasService) DeleteAttachment(ctx context.Context, userID, attachmentID uuid.UUID) error {
	attachment, err := s.queries.GetAttachmentByID(ctx, attachmentID)
	if err != nil {
		return fmt.Errorf("attachment not found")
	}

	// Verify access to the card
	if err := s.verifyCardAccess(ctx, userID, attachment.CardID); err != nil {
		return err
	}

	// Only author can delete their attachment
	if attachment.UserID != userID {
		return fmt.Errorf("you can only delete your own attachments")
	}

	if err := s.queries.DeleteAttachment(ctx, attachmentID); err != nil {
		log.Printf("[ERROR] Failed to delete attachment: %v", err)
		return fmt.Errorf("failed to delete attachment")
	}

	return nil
}
