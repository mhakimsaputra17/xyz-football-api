package model

// Team represents a football team managed by Perusahaan XYZ.
type Team struct {
	Base
	Name        string   `gorm:"type:text;not null" json:"name"`
	LogoURL     string   `gorm:"type:text" json:"logo_url"`
	FoundedYear int      `gorm:"type:int" json:"founded_year"`
	Address     string   `gorm:"type:text" json:"address"`
	City        string   `gorm:"type:text" json:"city"`
	Players     []Player `gorm:"foreignKey:TeamID" json:"players,omitempty"`
}

// TableName overrides the default table name.
func (Team) TableName() string {
	return "teams"
}
