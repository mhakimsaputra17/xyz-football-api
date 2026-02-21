package repository

import (
	"github.com/google/uuid"
	"github.com/mhakimsaputra17/xyz-football-api/internal/model"
	"gorm.io/gorm"
)

// TeamRepository defines the contract for team data access.
type TeamRepository interface {
	FindAll(offset, limit int, sortBy, sortOrder string) ([]model.Team, error)
	FindByID(id uuid.UUID) (*model.Team, error)
	Create(team *model.Team) error
	Update(team *model.Team) error
	Delete(id uuid.UUID) error
	Count() (int64, error)
}

// teamRepository implements TeamRepository using GORM.
type teamRepository struct {
	db *gorm.DB
}

// NewTeamRepository creates a new TeamRepository instance.
func NewTeamRepository(db *gorm.DB) TeamRepository {
	return &teamRepository{db: db}
}

func (r *teamRepository) FindAll(offset, limit int, sortBy, sortOrder string) ([]model.Team, error) {
	var teams []model.Team
	query := r.db.Offset(offset).Limit(limit)

	// Whitelist allowed sort columns to prevent SQL injection
	allowedSorts := map[string]bool{
		"created_at":   true,
		"name":         true,
		"founded_year": true,
		"city":         true,
	}
	if allowedSorts[sortBy] {
		query = query.Order(sortBy + " " + sortOrder)
	} else {
		query = query.Order("created_at desc")
	}

	if err := query.Find(&teams).Error; err != nil {
		return nil, err
	}
	return teams, nil
}

func (r *teamRepository) FindByID(id uuid.UUID) (*model.Team, error) {
	var team model.Team
	if err := r.db.Where("id = ?", id).First(&team).Error; err != nil {
		return nil, err
	}
	return &team, nil
}

func (r *teamRepository) Create(team *model.Team) error {
	return r.db.Create(team).Error
}

func (r *teamRepository) Update(team *model.Team) error {
	return r.db.Save(team).Error
}

func (r *teamRepository) Delete(id uuid.UUID) error {
	return r.db.Where("id = ?", id).Delete(&model.Team{}).Error
}

func (r *teamRepository) Count() (int64, error) {
	var count int64
	if err := r.db.Model(&model.Team{}).Count(&count).Error; err != nil {
		return 0, err
	}
	return count, nil
}
