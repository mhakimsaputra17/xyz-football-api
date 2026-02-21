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

// PlayerService defines the contract for player business logic.
type PlayerService interface {
	GetAllByTeamID(teamID uuid.UUID, pagination dto.PaginationQuery) ([]dto.PlayerResponse, *response.PaginationMeta, error)
	GetByID(id uuid.UUID) (*dto.PlayerResponse, error)
	Create(teamID uuid.UUID, req dto.CreatePlayerRequest) (*dto.PlayerResponse, error)
	Update(id uuid.UUID, req dto.UpdatePlayerRequest) (*dto.PlayerResponse, error)
	Delete(id uuid.UUID) error
}

type playerService struct {
	playerRepo repository.PlayerRepository
	teamRepo   repository.TeamRepository
}

// NewPlayerService creates a new PlayerService instance.
func NewPlayerService(playerRepo repository.PlayerRepository, teamRepo repository.TeamRepository) PlayerService {
	return &playerService{
		playerRepo: playerRepo,
		teamRepo:   teamRepo,
	}
}

func (s *playerService) GetAllByTeamID(teamID uuid.UUID, pagination dto.PaginationQuery) ([]dto.PlayerResponse, *response.PaginationMeta, error) {
	pagination.Sanitize()

	// Verify team exists
	if _, err := s.teamRepo.FindByID(teamID); err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, nil, errs.ErrNotFound("Team not found")
		}
		slog.Error("failed to fetch team", "error", err, "team_id", teamID)
		return nil, nil, errs.ErrInternal("Internal server error")
	}

	players, err := s.playerRepo.FindAllByTeamID(teamID, pagination.GetOffset(), pagination.PerPage, pagination.SortBy, pagination.SortOrder)
	if err != nil {
		slog.Error("failed to fetch players", "error", err, "team_id", teamID)
		return nil, nil, errs.ErrInternal("Internal server error")
	}

	total, err := s.playerRepo.CountByTeamID(teamID)
	if err != nil {
		slog.Error("failed to count players", "error", err, "team_id", teamID)
		return nil, nil, errs.ErrInternal("Internal server error")
	}

	playerResponses := make([]dto.PlayerResponse, len(players))
	for i, player := range players {
		playerResponses[i] = toPlayerResponse(player)
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

	return playerResponses, meta, nil
}

func (s *playerService) GetByID(id uuid.UUID) (*dto.PlayerResponse, error) {
	player, err := s.playerRepo.FindByID(id)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, errs.ErrNotFound("Player not found")
		}
		slog.Error("failed to fetch player", "error", err, "player_id", id)
		return nil, errs.ErrInternal("Internal server error")
	}

	resp := toPlayerResponse(*player)
	return &resp, nil
}

// Create adds a new player to a team.
// Jersey number uniqueness per team is validated here (service layer) per PRD design.
func (s *playerService) Create(teamID uuid.UUID, req dto.CreatePlayerRequest) (*dto.PlayerResponse, error) {
	// Verify team exists
	if _, err := s.teamRepo.FindByID(teamID); err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, errs.ErrNotFound("Team not found")
		}
		slog.Error("failed to fetch team for player creation", "error", err, "team_id", teamID)
		return nil, errs.ErrInternal("Internal server error")
	}

	// Check jersey number uniqueness within the team (non-soft-deleted players only)
	existing, err := s.playerRepo.FindByTeamIDAndJerseyNumber(teamID, req.JerseyNumber)
	if err != nil && !errors.Is(err, gorm.ErrRecordNotFound) {
		slog.Error("failed to check jersey number uniqueness", "error", err)
		return nil, errs.ErrInternal("Internal server error")
	}
	if existing != nil {
		return nil, errs.ErrConflict("Jersey number already used in this team")
	}

	player := model.Player{
		TeamID:       teamID,
		Name:         req.Name,
		Height:       req.Height,
		Weight:       req.Weight,
		Position:     req.Position,
		JerseyNumber: req.JerseyNumber,
	}

	if err := s.playerRepo.Create(&player); err != nil {
		slog.Error("failed to create player", "error", err)
		return nil, errs.ErrInternal("Internal server error")
	}

	resp := toPlayerResponse(player)
	return &resp, nil
}

func (s *playerService) Update(id uuid.UUID, req dto.UpdatePlayerRequest) (*dto.PlayerResponse, error) {
	player, err := s.playerRepo.FindByID(id)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, errs.ErrNotFound("Player not found")
		}
		slog.Error("failed to fetch player for update", "error", err, "player_id", id)
		return nil, errs.ErrInternal("Internal server error")
	}

	// Check jersey number uniqueness if it changed
	if req.JerseyNumber != player.JerseyNumber {
		existing, err := s.playerRepo.FindByTeamIDAndJerseyNumber(player.TeamID, req.JerseyNumber)
		if err != nil && !errors.Is(err, gorm.ErrRecordNotFound) {
			slog.Error("failed to check jersey number uniqueness", "error", err)
			return nil, errs.ErrInternal("Internal server error")
		}
		if existing != nil {
			return nil, errs.ErrConflict("Jersey number already used in this team")
		}
	}

	player.Name = req.Name
	player.Height = req.Height
	player.Weight = req.Weight
	player.Position = req.Position
	player.JerseyNumber = req.JerseyNumber

	if err := s.playerRepo.Update(player); err != nil {
		slog.Error("failed to update player", "error", err, "player_id", id)
		return nil, errs.ErrInternal("Internal server error")
	}

	resp := toPlayerResponse(*player)
	return &resp, nil
}

func (s *playerService) Delete(id uuid.UUID) error {
	_, err := s.playerRepo.FindByID(id)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return errs.ErrNotFound("Player not found")
		}
		slog.Error("failed to fetch player for delete", "error", err, "player_id", id)
		return errs.ErrInternal("Internal server error")
	}

	if err := s.playerRepo.Delete(id); err != nil {
		slog.Error("failed to delete player", "error", err, "player_id", id)
		return errs.ErrInternal("Internal server error")
	}

	return nil
}

// toPlayerResponse converts a model.Player to dto.PlayerResponse.
func toPlayerResponse(player model.Player) dto.PlayerResponse {
	resp := dto.PlayerResponse{
		ID:           player.ID.String(),
		TeamID:       player.TeamID.String(),
		Name:         player.Name,
		Height:       player.Height,
		Weight:       player.Weight,
		Position:     player.Position,
		JerseyNumber: player.JerseyNumber,
		CreatedAt:    player.CreatedAt.Format("2006-01-02T15:04:05Z"),
		UpdatedAt:    player.UpdatedAt.Format("2006-01-02T15:04:05Z"),
	}

	if player.Team != nil {
		teamResp := toTeamResponse(*player.Team)
		resp.Team = &teamResp
	}

	return resp
}
