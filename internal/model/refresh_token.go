package model

import (
	"time"

	"github.com/google/uuid"
)

// RefreshToken stores issued refresh tokens for JWT authentication.
// Tokens can be invalidated on logout by deleting the record.
type RefreshToken struct {
	Base
	AdminID   uuid.UUID `gorm:"type:uuid;not null;index" json:"admin_id"`
	Token     string    `gorm:"type:text;not null;uniqueIndex" json:"-"`
	ExpiresAt time.Time `gorm:"not null" json:"expires_at"`
	Admin     *Admin    `gorm:"foreignKey:AdminID" json:"admin,omitempty"`
}

// TableName overrides the default table name.
func (RefreshToken) TableName() string {
	return "refresh_tokens"
}

// IsExpired checks whether the refresh token has passed its expiration time.
func (r *RefreshToken) IsExpired() bool {
	return time.Now().After(r.ExpiresAt)
}
