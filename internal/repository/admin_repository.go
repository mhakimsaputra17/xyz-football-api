package repository

import (
	"github.com/google/uuid"
	"github.com/mhakimsaputra17/xyz-football-api/internal/model"
	"gorm.io/gorm"
)

// AdminRepository defines the contract for admin data access.
type AdminRepository interface {
	FindByUsername(username string) (*model.Admin, error)
	FindByID(id uuid.UUID) (*model.Admin, error)
	Create(admin *model.Admin) error
}

// adminRepository implements AdminRepository using GORM.
type adminRepository struct {
	db *gorm.DB
}

// NewAdminRepository creates a new AdminRepository instance.
func NewAdminRepository(db *gorm.DB) AdminRepository {
	return &adminRepository{db: db}
}

func (r *adminRepository) FindByUsername(username string) (*model.Admin, error) {
	var admin model.Admin
	if err := r.db.Where("username = ?", username).First(&admin).Error; err != nil {
		return nil, err
	}
	return &admin, nil
}

func (r *adminRepository) FindByID(id uuid.UUID) (*model.Admin, error) {
	var admin model.Admin
	if err := r.db.Where("id = ?", id).First(&admin).Error; err != nil {
		return nil, err
	}
	return &admin, nil
}

func (r *adminRepository) Create(admin *model.Admin) error {
	return r.db.Create(admin).Error
}
