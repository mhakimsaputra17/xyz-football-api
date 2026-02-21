package model

import "github.com/google/uuid"

// ValidMatchStatuses defines the allowed match statuses.
var ValidMatchStatuses = []string{"scheduled", "completed"}

// Match represents a football match between two teams.
// Scores are computed automatically from the goals table.
type Match struct {
	Base
	HomeTeamID uuid.UUID `gorm:"type:uuid;not null;index" json:"home_team_id"`
	AwayTeamID uuid.UUID `gorm:"type:uuid;not null;index" json:"away_team_id"`
	MatchDate  string    `gorm:"type:text;not null" json:"match_date"` // YYYY-MM-DD
	MatchTime  string    `gorm:"type:text;not null" json:"match_time"` // HH:MM
	HomeScore  int       `gorm:"type:int;not null;default:0" json:"home_score"`
	AwayScore  int       `gorm:"type:int;not null;default:0" json:"away_score"`
	Status     string    `gorm:"type:text;not null;default:'scheduled'" json:"status"`
	HomeTeam   *Team     `gorm:"foreignKey:HomeTeamID" json:"home_team,omitempty"`
	AwayTeam   *Team     `gorm:"foreignKey:AwayTeamID" json:"away_team,omitempty"`
	Goals      []Goal    `gorm:"foreignKey:MatchID" json:"goals,omitempty"`
}

// TableName overrides the default table name.
func (Match) TableName() string {
	return "matches"
}
