// Package dto defines data transfer objects for activity logs.
package dto

// ActivityLogDTO is the response representation of an activity log.
type ActivityLogDTO struct {
	ID          string  `json:"id"`
	BoardID     string  `json:"board_id"`
	CardID      *string `json:"card_id,omitempty"`
	UserID      string  `json:"user_id"`
	UserName    string  `json:"user_name"`
	AvatarURL   *string `json:"user_avatar_url,omitempty"`
	Action      string  `json:"action"`
	EntityType  string  `json:"entity_type"`
	EntityTitle string  `json:"entity_title"`
	Details     *string `json:"details,omitempty"`
	CreatedAt   string  `json:"created_at"`
}
