package repository

import (
	"github.com/google/uuid"
	"github.com/mhakimsaputra17/xyz-football-api/internal/model"
	"gorm.io/gorm"
)

// GoalRepository defines the contract for goal data access.
type GoalRepository interface {
	Create(goal *model.Goal) error
	CreateBatch(goals []model.Goal) error
	FindByMatchID(matchID uuid.UUID) ([]model.Goal, error)
	DeleteByMatchID(matchID uuid.UUID) error
}

// goalRepository implements GoalRepository using GORM.
type goalRepository struct {
	db *gorm.DB
}

// NewGoalRepository creates a new GoalRepository instance.
func NewGoalRepository(db *gorm.DB) GoalRepository {
	return &goalRepository{db: db}
}

func (r *goalRepository) Create(goal *model.Goal) error {
	return r.db.Create(goal).Error
}

// CreateBatch inserts multiple goal records in a single operation.
func (r *goalRepository) CreateBatch(goals []model.Goal) error {
	if len(goals) == 0 {
		return nil
	}
	return r.db.Create(&goals).Error
}

func (r *goalRepository) FindByMatchID(matchID uuid.UUID) ([]model.Goal, error) {
	var goals []model.Goal
	err := r.db.
		Preload("Player").
		Preload("Team").
		Where("match_id = ?", matchID).
		Order("minute asc").
		Find(&goals).Error
	if err != nil {
		return nil, err
	}
	return goals, nil
}

// DeleteByMatchID performs a soft delete of all goals for a match.
// Used when updating match results (delete old goals, insert new ones).
func (r *goalRepository) DeleteByMatchID(matchID uuid.UUID) error {
	return r.db.Where("match_id = ?", matchID).Delete(&model.Goal{}).Error
}
