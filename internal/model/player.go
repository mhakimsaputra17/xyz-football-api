package model

import "github.com/google/uuid"

// ValidPositions defines the allowed player positions.
var ValidPositions = []string{"penyerang", "gelandang", "bertahan", "penjaga_gawang"}

// Player represents a football player belonging to a team.
// Jersey number uniqueness per team is validated at the service layer
// (not via DB constraint) because soft-deleted players should free up their numbers.
type Player struct {
	Base
	TeamID       uuid.UUID `gorm:"type:uuid;not null;index" json:"team_id"`
	Name         string    `gorm:"type:text;not null" json:"name"`
	Height       int       `gorm:"type:int" json:"height"` // in cm
	Weight       int       `gorm:"type:int" json:"weight"` // in kg
	Position     string    `gorm:"type:text;not null" json:"position"`
	JerseyNumber int       `gorm:"type:int;not null" json:"jersey_number"`
	Team         *Team     `gorm:"foreignKey:TeamID" json:"team,omitempty"`
}

// TableName overrides the default table name.
func (Player) TableName() string {
	return "players"
}
