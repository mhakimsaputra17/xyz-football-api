package model

// Admin represents a system administrator who can manage all resources.
// Only admins can access CRUD operations after authentication.
type Admin struct {
	Base
	Username string `gorm:"type:text;not null;uniqueIndex" json:"username"`
	Password string `gorm:"type:text;not null" json:"-"` // Never exposed in JSON responses
}

// TableName overrides the default table name.
func (Admin) TableName() string {
	return "admins"
}
