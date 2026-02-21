package dto

// MatchReportResponse represents the detailed match report for a completed match.
type MatchReportResponse struct {
	MatchID           string             `json:"match_id" example:"019292f0-6b00-7a50-8d00-000000001000"`
	MatchDate         string             `json:"match_date" example:"2025-06-15"`
	MatchTime         string             `json:"match_time" example:"19:30"`
	HomeTeam          TeamResponse       `json:"home_team"`
	AwayTeam          TeamResponse       `json:"away_team"`
	HomeScore         int                `json:"home_score" example:"2"`
	AwayScore         int                `json:"away_score" example:"1"`
	MatchResult       string             `json:"match_result" example:"Home Win"` // "Home Win", "Away Win", "Draw"
	Goals             []MatchReportGoal  `json:"goals"`
	TopScorer         *TopScorerResponse `json:"top_scorer"`
	HomeTeamTotalWins int                `json:"home_team_total_wins" example:"5"`
	AwayTeamTotalWins int                `json:"away_team_total_wins" example:"3"`
}

// MatchReportGoal represents a goal entry in the match report.
type MatchReportGoal struct {
	PlayerName string `json:"player_name" example:"Marko Simic"`
	TeamName   string `json:"team_name" example:"Persija Jakarta"`
	Minute     int    `json:"minute" example:"45"`
}

// TopScorerResponse represents the top scorer of a match.
type TopScorerResponse struct {
	PlayerName   string `json:"player_name" example:"Marko Simic"`
	TeamName     string `json:"team_name" example:"Persija Jakarta"`
	GoalsInMatch int    `json:"goals_in_match" example:"2"`
}

// MatchReportListItem represents a summary item in the match report list.
type MatchReportListItem struct {
	MatchID     string       `json:"match_id" example:"019292f0-6b00-7a50-8d00-000000001000"`
	MatchDate   string       `json:"match_date" example:"2025-06-15"`
	MatchTime   string       `json:"match_time" example:"19:30"`
	HomeTeam    TeamResponse `json:"home_team"`
	AwayTeam    TeamResponse `json:"away_team"`
	HomeScore   int          `json:"home_score" example:"2"`
	AwayScore   int          `json:"away_score" example:"1"`
	MatchResult string       `json:"match_result" example:"Home Win"`
}
