package dto

// CreatePlayerRequest represents the request payload for creating a player.
type CreatePlayerRequest struct {
	Name         string `json:"name" binding:"required" example:"Marko Simic"`
	Height       int    `json:"height" binding:"required,gt=0" example:"185"`
	Weight       int    `json:"weight" binding:"required,gt=0" example:"80"`
	Position     string `json:"position" binding:"required,oneof=penyerang gelandang bertahan penjaga_gawang" example:"penyerang"`
	JerseyNumber int    `json:"jersey_number" binding:"required,gt=0" example:"9"`
}

// UpdatePlayerRequest represents the request payload for updating a player.
type UpdatePlayerRequest struct {
	Name         string `json:"name" binding:"required" example:"Marko Simic"`
	Height       int    `json:"height" binding:"required,gt=0" example:"185"`
	Weight       int    `json:"weight" binding:"required,gt=0" example:"80"`
	Position     string `json:"position" binding:"required,oneof=penyerang gelandang bertahan penjaga_gawang" example:"penyerang"`
	JerseyNumber int    `json:"jersey_number" binding:"required,gt=0" example:"9"`
}

// PlayerResponse represents the player data returned in API responses.
type PlayerResponse struct {
	ID           string        `json:"id" example:"019292f0-6b00-7a50-8d00-000000000100"`
	TeamID       string        `json:"team_id" example:"019292f0-6b00-7a50-8d00-000000000010"`
	Name         string        `json:"name" example:"Marko Simic"`
	Height       int           `json:"height" example:"185"`
	Weight       int           `json:"weight" example:"80"`
	Position     string        `json:"position" example:"penyerang"`
	JerseyNumber int           `json:"jersey_number" example:"9"`
	Team         *TeamResponse `json:"team,omitempty"`
	CreatedAt    string        `json:"created_at" example:"2025-01-15T10:30:00Z"`
	UpdatedAt    string        `json:"updated_at" example:"2025-01-15T10:30:00Z"`
}
