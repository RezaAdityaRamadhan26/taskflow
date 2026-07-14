// Package dto defines request and response data transfer objects for comments and attachments.
package dto

// CreateCommentRequest represents the request body for adding a comment.
type CreateCommentRequest struct {
	CardID  string `json:"card_id" validate:"required,uuid"`
	Content string `json:"content" validate:"required,min=1,max=5000"`
}

// UpdateCommentRequest represents the request body for editing a comment.
type UpdateCommentRequest struct {
	Content string `json:"content" validate:"required,min=1,max=5000"`
}

// CommentDTO is the response representation of a comment.
type CommentDTO struct {
	ID        string  `json:"id"`
	CardID    string  `json:"card_id"`
	UserID    string  `json:"user_id"`
	UserName  string  `json:"user_name"`
	AvatarURL *string `json:"user_avatar_url,omitempty"`
	Content   string  `json:"content"`
	CreatedAt string  `json:"created_at"`
	UpdatedAt string  `json:"updated_at"`
}

// CreateAttachmentRequest represents the request body for adding an attachment link.
type CreateAttachmentRequest struct {
	CardID   string  `json:"card_id" validate:"required,uuid"`
	FileName string  `json:"file_name" validate:"required,min=1,max=255"`
	FileURL  string  `json:"file_url" validate:"required,url"`
	FileSize *int32  `json:"file_size,omitempty"`
	FileType *string `json:"file_type,omitempty" validate:"omitempty,max=100"`
}

// AttachmentDTO is the response representation of an attachment.
type AttachmentDTO struct {
	ID        string  `json:"id"`
	CardID    string  `json:"card_id"`
	UserID    string  `json:"user_id"`
	UserName  string  `json:"user_name"`
	AvatarURL *string `json:"user_avatar_url,omitempty"`
	FileName  string  `json:"file_name"`
	FileURL   string  `json:"file_url"`
	FileSize  *int32  `json:"file_size,omitempty"`
	FileType  *string `json:"file_type,omitempty"`
	CreatedAt string  `json:"created_at"`
}
