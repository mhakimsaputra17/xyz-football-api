package repository

import (
	"github.com/google/uuid"
	"github.com/mhakimsaputra17/xyz-football-api/internal/model"
	"gorm.io/gorm"
)

// MatchRepository defines the contract for match data access.
type MatchRepository interface {
	FindAll(offset, limit int, sortBy, sortOrder string) ([]model.Match, error)
	FindByID(id uuid.UUID) (*model.Match, error)
	FindByIDWithDetails(id uuid.UUID) (*model.Match, error)
	Create(match *model.Match) error
	Update(match *model.Match) error
	Delete(id uuid.UUID) error
	Count() (int64, error)
	FindCompletedMatches(offset, limit int) ([]model.Match, error)
	CountCompletedMatches() (int64, error)
	CountWins(teamID uuid.UUID) (int, error)
}

// matchRepository implements MatchRepository using GORM.
type matchRepository struct {
	db *gorm.DB
}

// NewMatchRepository creates a new MatchRepository instance.
func NewMatchRepository(db *gorm.DB) MatchRepository {
	return &matchRepository{db: db}
}

func (r *matchRepository) FindAll(offset, limit int, sortBy, sortOrder string) ([]model.Match, error) {
	var matches []model.Match
	query := r.db.Preload("HomeTeam").Preload("AwayTeam").Offset(offset).Limit(limit)

	allowedSorts := map[string]bool{
		"created_at": true,
		"match_date": true,
		"status":     true,
	}
	if allowedSorts[sortBy] {
		query = query.Order(sortBy + " " + sortOrder)
	} else {
		query = query.Order("created_at desc")
	}

	if err := query.Find(&matches).Error; err != nil {
		return nil, err
	}
	return matches, nil
}

func (r *matchRepository) FindByID(id uuid.UUID) (*model.Match, error) {
	var match model.Match
	if err := r.db.Preload("HomeTeam").Preload("AwayTeam").Where("id = ?", id).First(&match).Error; err != nil {
		return nil, err
	}
	return &match, nil
}

// FindByIDWithDetails loads a match with all associations: HomeTeam, AwayTeam, Goals, Goals.Player, Goals.Team.
func (r *matchRepository) FindByIDWithDetails(id uuid.UUID) (*model.Match, error) {
	var match model.Match
	err := r.db.
		Preload("HomeTeam").
		Preload("AwayTeam").
		Preload("Goals", func(db *gorm.DB) *gorm.DB {
			return db.Order("minute asc")
		}).
		Preload("Goals.Player").
		Preload("Goals.Team").
		Where("id = ?", id).
		First(&match).Error
	if err != nil {
		return nil, err
	}
	return &match, nil
}

func (r *matchRepository) Create(match *model.Match) error {
	return r.db.Create(match).Error
}

func (r *matchRepository) Update(match *model.Match) error {
	return r.db.Save(match).Error
}

func (r *matchRepository) Delete(id uuid.UUID) error {
	return r.db.Where("id = ?", id).Delete(&model.Match{}).Error
}

func (r *matchRepository) Count() (int64, error) {
	var count int64
	if err := r.db.Model(&model.Match{}).Count(&count).Error; err != nil {
		return 0, err
	}
	return count, nil
}

func (r *matchRepository) FindCompletedMatches(offset, limit int) ([]model.Match, error) {
	var matches []model.Match
	err := r.db.
		Preload("HomeTeam").
		Preload("AwayTeam").
		Where("status = ?", "completed").
		Order("match_date desc").
		Offset(offset).
		Limit(limit).
		Find(&matches).Error
	if err != nil {
		return nil, err
	}
	return matches, nil
}

func (r *matchRepository) CountCompletedMatches() (int64, error) {
	var count int64
	if err := r.db.Model(&model.Match{}).Where("status = ?", "completed").Count(&count).Error; err != nil {
		return 0, err
	}
	return count, nil
}

// CountWins calculates the total number of wins for a team across ALL completed matches.
// A win is when the team is home and home_score > away_score, or away and away_score > home_score.
func (r *matchRepository) CountWins(teamID uuid.UUID) (int, error) {
	var count int64
	err := r.db.Model(&model.Match{}).
		Where("status = ? AND ((home_team_id = ? AND home_score > away_score) OR (away_team_id = ? AND away_score > home_score))",
			"completed", teamID, teamID).
		Count(&count).Error
	if err != nil {
		return 0, err
	}
	return int(count), nil
}
