// Package dto defines request and response data transfer objects for the list module.
package dto

// CreateListRequest represents the request body for creating a list.
type CreateListRequest struct {
	BoardID  string   `json:"board_id" validate:"required,uuid"`
	Name     string   `json:"name" validate:"required,min=1,max=100"`
	Position *float64 `json:"position,omitempty"` // If not provided, it will be placed at the end
}

// UpdateListRequest represents the request body for updating a list.
type UpdateListRequest struct {
	Name     string  `json:"name" validate:"required,min=1,max=100"`
	Position float64 `json:"position" validate:"required"` // Position is required for update (even if just renaming, client must send current pos)
}

// ListDTO is the response representation of a list.
type ListDTO struct {
	ID        string  `json:"id"`
	BoardID   string  `json:"board_id"`
	Name      string  `json:"name"`
	Position  float64 `json:"position"`
	CreatedAt string  `json:"created_at"`
	UpdatedAt string  `json:"updated_at"`
}
