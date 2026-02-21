package service

import (
	"errors"
	"log/slog"

	"github.com/google/uuid"
	"github.com/mhakimsaputra17/xyz-football-api/internal/dto"
	"github.com/mhakimsaputra17/xyz-football-api/internal/model"
	"github.com/mhakimsaputra17/xyz-football-api/internal/repository"
	"github.com/mhakimsaputra17/xyz-football-api/pkg/errs"
	"github.com/mhakimsaputra17/xyz-football-api/pkg/response"
	"gorm.io/gorm"
)

// TeamService defines the contract for team business logic.
type TeamService interface {
	GetAll(pagination dto.PaginationQuery) ([]dto.TeamResponse, *response.PaginationMeta, error)
	GetByID(id uuid.UUID) (*dto.TeamResponse, error)
	Create(req dto.CreateTeamRequest) (*dto.TeamResponse, error)
	Update(id uuid.UUID, req dto.UpdateTeamRequest) (*dto.TeamResponse, error)
	Delete(id uuid.UUID) error
}

type teamService struct {
	teamRepo repository.TeamRepository
}

// NewTeamService creates a new TeamService instance.
func NewTeamService(teamRepo repository.TeamRepository) TeamService {
	return &teamService{teamRepo: teamRepo}
}

func (s *teamService) GetAll(pagination dto.PaginationQuery) ([]dto.TeamResponse, *response.PaginationMeta, error) {
	pagination.Sanitize()

	teams, err := s.teamRepo.FindAll(pagination.GetOffset(), pagination.PerPage, pagination.SortBy, pagination.SortOrder)
	if err != nil {
		slog.Error("failed to fetch teams", "error", err)
		return nil, nil, errs.ErrInternal("Internal server error")
	}

	total, err := s.teamRepo.Count()
	if err != nil {
		slog.Error("failed to count teams", "error", err)
		return nil, nil, errs.ErrInternal("Internal server error")
	}

	teamResponses := make([]dto.TeamResponse, len(teams))
	for i, team := range teams {
		teamResponses[i] = toTeamResponse(team)
	}

	totalPages := int(total) / pagination.PerPage
	if int(total)%pagination.PerPage > 0 {
		totalPages++
	}

	meta := &response.PaginationMeta{
		Page:       pagination.Page,
		PerPage:    pagination.PerPage,
		Total:      total,
		TotalPages: totalPages,
	}

	return teamResponses, meta, nil
}

func (s *teamService) GetByID(id uuid.UUID) (*dto.TeamResponse, error) {
	team, err := s.teamRepo.FindByID(id)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, errs.ErrNotFound("Team not found")
		}
		slog.Error("failed to fetch team", "error", err, "team_id", id)
		return nil, errs.ErrInternal("Internal server error")
	}

	resp := toTeamResponse(*team)
	return &resp, nil
}

func (s *teamService) Create(req dto.CreateTeamRequest) (*dto.TeamResponse, error) {
	team := model.Team{
		Name:        req.Name,
		LogoURL:     req.LogoURL,
		FoundedYear: req.FoundedYear,
		Address:     req.Address,
		City:        req.City,
	}

	if err := s.teamRepo.Create(&team); err != nil {
		slog.Error("failed to create team", "error", err)
		return nil, errs.ErrInternal("Internal server error")
	}

	resp := toTeamResponse(team)
	return &resp, nil
}

func (s *teamService) Update(id uuid.UUID, req dto.UpdateTeamRequest) (*dto.TeamResponse, error) {
	team, err := s.teamRepo.FindByID(id)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, errs.ErrNotFound("Team not found")
		}
		slog.Error("failed to fetch team for update", "error", err, "team_id", id)
		return nil, errs.ErrInternal("Internal server error")
	}

	team.Name = req.Name
	team.LogoURL = req.LogoURL
	team.FoundedYear = req.FoundedYear
	team.Address = req.Address
	team.City = req.City

	if err := s.teamRepo.Update(team); err != nil {
		slog.Error("failed to update team", "error", err, "team_id", id)
		return nil, errs.ErrInternal("Internal server error")
	}

	resp := toTeamResponse(*team)
	return &resp, nil
}

func (s *teamService) Delete(id uuid.UUID) error {
	_, err := s.teamRepo.FindByID(id)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return errs.ErrNotFound("Team not found")
		}
		slog.Error("failed to fetch team for delete", "error", err, "team_id", id)
		return errs.ErrInternal("Internal server error")
	}

	if err := s.teamRepo.Delete(id); err != nil {
		slog.Error("failed to delete team", "error", err, "team_id", id)
		return errs.ErrInternal("Internal server error")
	}

	return nil
}

// toTeamResponse converts a model.Team to dto.TeamResponse.
func toTeamResponse(team model.Team) dto.TeamResponse {
	return dto.TeamResponse{
		ID:          team.ID.String(),
		Name:        team.Name,
		LogoURL:     team.LogoURL,
		FoundedYear: team.FoundedYear,
		Address:     team.Address,
		City:        team.City,
		CreatedAt:   team.CreatedAt.Format("2006-01-02T15:04:05Z"),
		UpdatedAt:   team.UpdatedAt.Format("2006-01-02T15:04:05Z"),
	}
}
