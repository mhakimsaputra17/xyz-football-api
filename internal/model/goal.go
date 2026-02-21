package model

import "github.com/google/uuid"

// Goal represents a goal scored in a match.
// The player must belong to one of the two teams in the match (validated in service layer).
type Goal struct {
	Base
	MatchID  uuid.UUID `gorm:"type:uuid;not null;index" json:"match_id"`
	PlayerID uuid.UUID `gorm:"type:uuid;not null;index" json:"player_id"`
	TeamID   uuid.UUID `gorm:"type:uuid;not null" json:"team_id"`
	Minute   int       `gorm:"type:int;not null" json:"minute"` // Must be >= 1
	Match    *Match    `gorm:"foreignKey:MatchID" json:"match,omitempty"`
	Player   *Player   `gorm:"foreignKey:PlayerID" json:"player,omitempty"`
	Team     *Team     `gorm:"foreignKey:TeamID" json:"team,omitempty"`
}

// TableName overrides the default table name.
func (Goal) TableName() string {
	return "goals"
}
