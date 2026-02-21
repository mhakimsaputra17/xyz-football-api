package dto

// CreatePlayerRequest represents the request payload for creating a player.
type CreatePlayerRequest struct {
	Name         string `json:"name" binding:"required"`
	Height       int    `json:"height" binding:"required,gt=0"`
	Weight       int    `json:"weight" binding:"required,gt=0"`
	Position     string `json:"position" binding:"required,oneof=penyerang gelandang bertahan penjaga_gawang"`
	JerseyNumber int    `json:"jersey_number" binding:"required,gt=0"`
}

// UpdatePlayerRequest represents the request payload for updating a player.
type UpdatePlayerRequest struct {
	Name         string `json:"name" binding:"required"`
	Height       int    `json:"height" binding:"required,gt=0"`
	Weight       int    `json:"weight" binding:"required,gt=0"`
	Position     string `json:"position" binding:"required,oneof=penyerang gelandang bertahan penjaga_gawang"`
	JerseyNumber int    `json:"jersey_number" binding:"required,gt=0"`
}

// PlayerResponse represents the player data returned in API responses.
type PlayerResponse struct {
	ID           string        `json:"id"`
	TeamID       string        `json:"team_id"`
	Name         string        `json:"name"`
	Height       int           `json:"height"`
	Weight       int           `json:"weight"`
	Position     string        `json:"position"`
	JerseyNumber int           `json:"jersey_number"`
	Team         *TeamResponse `json:"team,omitempty"`
	CreatedAt    string        `json:"created_at"`
	UpdatedAt    string        `json:"updated_at"`
}
