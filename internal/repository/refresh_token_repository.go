package repository

import (
	"github.com/google/uuid"
	"github.com/mhakimsaputra17/xyz-football-api/internal/model"
	"gorm.io/gorm"
)

// RefreshTokenRepository defines the contract for refresh token data access.
type RefreshTokenRepository interface {
	Create(token *model.RefreshToken) error
	FindByToken(token string) (*model.RefreshToken, error)
	DeleteByToken(token string) error
	DeleteByAdminID(adminID uuid.UUID) error
}

// refreshTokenRepository implements RefreshTokenRepository using GORM.
type refreshTokenRepository struct {
	db *gorm.DB
}

// NewRefreshTokenRepository creates a new RefreshTokenRepository instance.
func NewRefreshTokenRepository(db *gorm.DB) RefreshTokenRepository {
	return &refreshTokenRepository{db: db}
}

func (r *refreshTokenRepository) Create(token *model.RefreshToken) error {
	return r.db.Create(token).Error
}

// FindByToken looks up a refresh token by its string value.
// Returns the token with its associated admin for validation.
func (r *refreshTokenRepository) FindByToken(token string) (*model.RefreshToken, error) {
	var rt model.RefreshToken
	if err := r.db.Where("token = ?", token).First(&rt).Error; err != nil {
		return nil, err
	}
	return &rt, nil
}

// DeleteByToken performs a hard delete (not soft delete) of a refresh token.
func (r *refreshTokenRepository) DeleteByToken(token string) error {
	return r.db.Unscoped().Where("token = ?", token).Delete(&model.RefreshToken{}).Error
}

// DeleteByAdminID performs a hard delete of ALL refresh tokens for an admin.
// Supports "logout from all devices" functionality.
func (r *refreshTokenRepository) DeleteByAdminID(adminID uuid.UUID) error {
	return r.db.Unscoped().Where("admin_id = ?", adminID).Delete(&model.RefreshToken{}).Error
}
