package dto

// CreateMatchRequest represents the request payload for creating a match schedule.
type CreateMatchRequest struct {
	HomeTeamID string `json:"home_team_id" binding:"required,uuid" example:"019292f0-6b00-7a50-8d00-000000000010"`
	AwayTeamID string `json:"away_team_id" binding:"required,uuid" example:"019292f0-6b00-7a50-8d00-000000000020"`
	MatchDate  string `json:"match_date" binding:"required" example:"2025-06-15"` // YYYY-MM-DD
	MatchTime  string `json:"match_time" binding:"required" example:"19:30"`      // HH:MM
}

// UpdateMatchRequest represents the request payload for updating a match schedule.
type UpdateMatchRequest struct {
	HomeTeamID string `json:"home_team_id" binding:"required,uuid" example:"019292f0-6b00-7a50-8d00-000000000010"`
	AwayTeamID string `json:"away_team_id" binding:"required,uuid" example:"019292f0-6b00-7a50-8d00-000000000020"`
	MatchDate  string `json:"match_date" binding:"required" example:"2025-06-15"`
	MatchTime  string `json:"match_time" binding:"required" example:"19:30"`
}

// MatchResultRequest represents the request payload for submitting match results.
type MatchResultRequest struct {
	Goals []GoalInput `json:"goals" binding:"required,dive"`
}

// GoalInput represents a single goal entry in the match result request.
type GoalInput struct {
	PlayerID string `json:"player_id" binding:"required,uuid" example:"019292f0-6b00-7a50-8d00-000000000100"`
	TeamID   string `json:"team_id" binding:"required,uuid" example:"019292f0-6b00-7a50-8d00-000000000010"`
	Minute   int    `json:"minute" binding:"required,gte=1" example:"45"`
}

// MatchResponse represents the match data returned in API responses.
type MatchResponse struct {
	ID         string         `json:"id" example:"019292f0-6b00-7a50-8d00-000000001000"`
	HomeTeamID string         `json:"home_team_id" example:"019292f0-6b00-7a50-8d00-000000000010"`
	AwayTeamID string         `json:"away_team_id" example:"019292f0-6b00-7a50-8d00-000000000020"`
	MatchDate  string         `json:"match_date" example:"2025-06-15"`
	MatchTime  string         `json:"match_time" example:"19:30"`
	HomeScore  int            `json:"home_score" example:"2"`
	AwayScore  int            `json:"away_score" example:"1"`
	Status     string         `json:"status" example:"completed"`
	HomeTeam   *TeamResponse  `json:"home_team,omitempty"`
	AwayTeam   *TeamResponse  `json:"away_team,omitempty"`
	Goals      []GoalResponse `json:"goals,omitempty"`
	CreatedAt  string         `json:"created_at" example:"2025-01-15T10:30:00Z"`
	UpdatedAt  string         `json:"updated_at" example:"2025-01-15T10:30:00Z"`
}

// GoalResponse represents a goal entry in API responses.
type GoalResponse struct {
	ID        string          `json:"id" example:"019292f0-6b00-7a50-8d00-000000010000"`
	MatchID   string          `json:"match_id" example:"019292f0-6b00-7a50-8d00-000000001000"`
	PlayerID  string          `json:"player_id" example:"019292f0-6b00-7a50-8d00-000000000100"`
	TeamID    string          `json:"team_id" example:"019292f0-6b00-7a50-8d00-000000000010"`
	Minute    int             `json:"minute" example:"45"`
	Player    *PlayerResponse `json:"player,omitempty"`
	Team      *TeamResponse   `json:"team,omitempty"`
	CreatedAt string          `json:"created_at" example:"2025-01-15T10:30:00Z"`
}
