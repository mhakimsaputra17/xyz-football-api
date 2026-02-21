package dto

// MatchReportResponse represents the detailed match report for a completed match.
type MatchReportResponse struct {
	MatchID           string             `json:"match_id"`
	MatchDate         string             `json:"match_date"`
	MatchTime         string             `json:"match_time"`
	HomeTeam          TeamResponse       `json:"home_team"`
	AwayTeam          TeamResponse       `json:"away_team"`
	HomeScore         int                `json:"home_score"`
	AwayScore         int                `json:"away_score"`
	MatchResult       string             `json:"match_result"` // "Home Win", "Away Win", "Draw"
	Goals             []MatchReportGoal  `json:"goals"`
	TopScorer         *TopScorerResponse `json:"top_scorer"`
	HomeTeamTotalWins int                `json:"home_team_total_wins"`
	AwayTeamTotalWins int                `json:"away_team_total_wins"`
}

// MatchReportGoal represents a goal entry in the match report.
type MatchReportGoal struct {
	PlayerName string `json:"player_name"`
	TeamName   string `json:"team_name"`
	Minute     int    `json:"minute"`
}

// TopScorerResponse represents the top scorer of a match.
type TopScorerResponse struct {
	PlayerName   string `json:"player_name"`
	TeamName     string `json:"team_name"`
	GoalsInMatch int    `json:"goals_in_match"`
}

// MatchReportListItem represents a summary item in the match report list.
type MatchReportListItem struct {
	MatchID     string       `json:"match_id"`
	MatchDate   string       `json:"match_date"`
	MatchTime   string       `json:"match_time"`
	HomeTeam    TeamResponse `json:"home_team"`
	AwayTeam    TeamResponse `json:"away_team"`
	HomeScore   int          `json:"home_score"`
	AwayScore   int          `json:"away_score"`
	MatchResult string       `json:"match_result"`
}
