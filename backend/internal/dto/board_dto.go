// Package dto defines request and response data transfer objects for the board module.
package dto

// CreateBoardRequest represents the request body for creating a board.
type CreateBoardRequest struct {
	WorkspaceID string  `json:"workspace_id" validate:"required,uuid"`
	Name        string  `json:"name" validate:"required,min=2,max=100"`
	Description *string `json:"description,omitempty" validate:"omitempty,max=500"`
	Color       *string `json:"color,omitempty" validate:"omitempty,hexcolor"`
}

// UpdateBoardRequest represents the request body for updating a board.
type UpdateBoardRequest struct {
	Name        string  `json:"name" validate:"required,min=2,max=100"`
	Description *string `json:"description,omitempty" validate:"omitempty,max=500"`
	Color       *string `json:"color,omitempty" validate:"omitempty,hexcolor"`
}

// BoardDTO is the response representation of a board.
type BoardDTO struct {
	ID          string  `json:"id"`
	WorkspaceID string  `json:"workspace_id"`
	Name        string  `json:"name"`
	Description *string `json:"description,omitempty"`
	Color       *string `json:"color,omitempty"`
	CreatedAt   string  `json:"created_at"`
	UpdatedAt   string  `json:"updated_at"`
}
