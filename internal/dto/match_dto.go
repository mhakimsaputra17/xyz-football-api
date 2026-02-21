package dto

// CreateMatchRequest represents the request payload for creating a match schedule.
type CreateMatchRequest struct {
	HomeTeamID string `json:"home_team_id" binding:"required,uuid"`
	AwayTeamID string `json:"away_team_id" binding:"required,uuid"`
	MatchDate  string `json:"match_date" binding:"required"` // YYYY-MM-DD
	MatchTime  string `json:"match_time" binding:"required"` // HH:MM
}

// UpdateMatchRequest represents the request payload for updating a match schedule.
type UpdateMatchRequest struct {
	HomeTeamID string `json:"home_team_id" binding:"required,uuid"`
	AwayTeamID string `json:"away_team_id" binding:"required,uuid"`
	MatchDate  string `json:"match_date" binding:"required"`
	MatchTime  string `json:"match_time" binding:"required"`
}

// MatchResultRequest represents the request payload for submitting match results.
type MatchResultRequest struct {
	Goals []GoalInput `json:"goals" binding:"required,dive"`
}

// GoalInput represents a single goal entry in the match result request.
type GoalInput struct {
	PlayerID string `json:"player_id" binding:"required,uuid"`
	TeamID   string `json:"team_id" binding:"required,uuid"`
	Minute   int    `json:"minute" binding:"required,gte=1"`
}

// MatchResponse represents the match data returned in API responses.
type MatchResponse struct {
	ID         string         `json:"id"`
	HomeTeamID string         `json:"home_team_id"`
	AwayTeamID string         `json:"away_team_id"`
	MatchDate  string         `json:"match_date"`
	MatchTime  string         `json:"match_time"`
	HomeScore  int            `json:"home_score"`
	AwayScore  int            `json:"away_score"`
	Status     string         `json:"status"`
	HomeTeam   *TeamResponse  `json:"home_team,omitempty"`
	AwayTeam   *TeamResponse  `json:"away_team,omitempty"`
	Goals      []GoalResponse `json:"goals,omitempty"`
	CreatedAt  string         `json:"created_at"`
	UpdatedAt  string         `json:"updated_at"`
}

// GoalResponse represents a goal entry in API responses.
type GoalResponse struct {
	ID        string          `json:"id"`
	MatchID   string          `json:"match_id"`
	PlayerID  string          `json:"player_id"`
	TeamID    string          `json:"team_id"`
	Minute    int             `json:"minute"`
	Player    *PlayerResponse `json:"player,omitempty"`
	Team      *TeamResponse   `json:"team,omitempty"`
	CreatedAt string          `json:"created_at"`
}
