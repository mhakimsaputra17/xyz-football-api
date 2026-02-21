package repository

import (
	"github.com/google/uuid"
	"github.com/mhakimsaputra17/xyz-football-api/internal/model"
	"gorm.io/gorm"
)

// PlayerRepository defines the contract for player data access.
type PlayerRepository interface {
	FindAllByTeamID(teamID uuid.UUID, offset, limit int, sortBy, sortOrder string) ([]model.Player, error)
	FindByID(id uuid.UUID) (*model.Player, error)
	Create(player *model.Player) error
	Update(player *model.Player) error
	Delete(id uuid.UUID) error
	CountByTeamID(teamID uuid.UUID) (int64, error)
	FindByTeamIDAndJerseyNumber(teamID uuid.UUID, jerseyNumber int) (*model.Player, error)
}

// playerRepository implements PlayerRepository using GORM.
type playerRepository struct {
	db *gorm.DB
}

// NewPlayerRepository creates a new PlayerRepository instance.
func NewPlayerRepository(db *gorm.DB) PlayerRepository {
	return &playerRepository{db: db}
}

func (r *playerRepository) FindAllByTeamID(teamID uuid.UUID, offset, limit int, sortBy, sortOrder string) ([]model.Player, error) {
	var players []model.Player
	query := r.db.Where("team_id = ?", teamID).Offset(offset).Limit(limit)

	allowedSorts := map[string]bool{
		"created_at":    true,
		"name":          true,
		"jersey_number": true,
		"position":      true,
	}
	if allowedSorts[sortBy] {
		query = query.Order(sortBy + " " + sortOrder)
	} else {
		query = query.Order("created_at desc")
	}

	if err := query.Find(&players).Error; err != nil {
		return nil, err
	}
	return players, nil
}

func (r *playerRepository) FindByID(id uuid.UUID) (*model.Player, error) {
	var player model.Player
	if err := r.db.Preload("Team").Where("id = ?", id).First(&player).Error; err != nil {
		return nil, err
	}
	return &player, nil
}

func (r *playerRepository) Create(player *model.Player) error {
	return r.db.Create(player).Error
}

func (r *playerRepository) Update(player *model.Player) error {
	return r.db.Save(player).Error
}

func (r *playerRepository) Delete(id uuid.UUID) error {
	return r.db.Where("id = ?", id).Delete(&model.Player{}).Error
}

func (r *playerRepository) CountByTeamID(teamID uuid.UUID) (int64, error) {
	var count int64
	if err := r.db.Model(&model.Player{}).Where("team_id = ?", teamID).Count(&count).Error; err != nil {
		return 0, err
	}
	return count, nil
}

// FindByTeamIDAndJerseyNumber checks jersey number uniqueness per team.
// Only considers non-soft-deleted records (GORM default behavior).
func (r *playerRepository) FindByTeamIDAndJerseyNumber(teamID uuid.UUID, jerseyNumber int) (*model.Player, error) {
	var player model.Player
	err := r.db.Where("team_id = ? AND jersey_number = ?", teamID, jerseyNumber).First(&player).Error
	if err != nil {
		return nil, err
	}
	return &player, nil
}
