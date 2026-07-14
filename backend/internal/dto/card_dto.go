// Package dto defines request and response data transfer objects for the card module.
package dto

import "time"

// CreateCardRequest represents the request body for creating a card.
type CreateCardRequest struct {
	ListID      string     `json:"list_id" validate:"required,uuid"`
	Title       string     `json:"title" validate:"required,min=1,max=255"`
	Description *string    `json:"description,omitempty" validate:"omitempty,max=5000"`
	Position    *float64   `json:"position,omitempty"`
	Priority    *string    `json:"priority,omitempty" validate:"omitempty,oneof=LOW MEDIUM HIGH URGENT"`
	DueDate     *time.Time `json:"due_date,omitempty"`
}

// UpdateCardRequest represents the request body for updating a card.
// Including moving between lists.
type UpdateCardRequest struct {
	ListID      string     `json:"list_id" validate:"required,uuid"`
	Title       string     `json:"title" validate:"required,min=1,max=255"`
	Description *string    `json:"description,omitempty" validate:"omitempty,max=5000"`
	Position    float64    `json:"position" validate:"required"`
	Priority    string     `json:"priority" validate:"required,oneof=LOW MEDIUM HIGH URGENT"`
	DueDate     *time.Time `json:"due_date,omitempty"`
}

// CardDTO is the response representation of a card.
type CardDTO struct {
	ID          string     `json:"id"`
	ListID      string     `json:"list_id"`
	Title       string     `json:"title"`
	Description *string    `json:"description,omitempty"`
	Position    float64    `json:"position"`
	Priority    string     `json:"priority"`
	DueDate     *time.Time `json:"due_date,omitempty"`
	CreatedAt   string     `json:"created_at"`
	UpdatedAt   string     `json:"updated_at"`
}
