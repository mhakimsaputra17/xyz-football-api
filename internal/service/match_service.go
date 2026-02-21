package service

import (
	"errors"
	"fmt"
	"log/slog"

	"github.com/google/uuid"
	"github.com/mhakimsaputra17/xyz-football-api/internal/dto"
	"github.com/mhakimsaputra17/xyz-football-api/internal/model"
	"github.com/mhakimsaputra17/xyz-football-api/internal/repository"
	"github.com/mhakimsaputra17/xyz-football-api/pkg/errs"
	"github.com/mhakimsaputra17/xyz-football-api/pkg/response"
	"gorm.io/gorm"
)

// MatchService defines the contract for match business logic.
type MatchService interface {
	GetAll(pagination dto.PaginationQuery) ([]dto.MatchResponse, *response.PaginationMeta, error)
	GetByID(id uuid.UUID) (*dto.MatchResponse, error)
	Create(req dto.CreateMatchRequest) (*dto.MatchResponse, error)
	Update(id uuid.UUID, req dto.UpdateMatchRequest) (*dto.MatchResponse, error)
	Delete(id uuid.UUID) error
	SubmitResult(matchID uuid.UUID, req dto.MatchResultRequest) (*dto.MatchResponse, error)
	UpdateResult(matchID uuid.UUID, req dto.MatchResultRequest) (*dto.MatchResponse, error)
}

type matchService struct {
	matchRepo  repository.MatchRepository
	teamRepo   repository.TeamRepository
	playerRepo repository.PlayerRepository
	goalRepo   repository.GoalRepository
}

// NewMatchService creates a new MatchService instance.
func NewMatchService(
	matchRepo repository.MatchRepository,
	teamRepo repository.TeamRepository,
	playerRepo repository.PlayerRepository,
	goalRepo repository.GoalRepository,
) MatchService {
	return &matchService{
		matchRepo:  matchRepo,
		teamRepo:   teamRepo,
		playerRepo: playerRepo,
		goalRepo:   goalRepo,
	}
}

func (s *matchService) GetAll(pagination dto.PaginationQuery) ([]dto.MatchResponse, *response.PaginationMeta, error) {
	pagination.Sanitize()

	matches, err := s.matchRepo.FindAll(pagination.GetOffset(), pagination.PerPage, pagination.SortBy, pagination.SortOrder)
	if err != nil {
		slog.Error("failed to fetch matches", "error", err)
		return nil, nil, errs.ErrInternal("Internal server error")
	}

	total, err := s.matchRepo.Count()
	if err != nil {
		slog.Error("failed to count matches", "error", err)
		return nil, nil, errs.ErrInternal("Internal server error")
	}

	matchResponses := make([]dto.MatchResponse, len(matches))
	for i, match := range matches {
		matchResponses[i] = toMatchResponse(match)
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

	return matchResponses, meta, nil
}

func (s *matchService) GetByID(id uuid.UUID) (*dto.MatchResponse, error) {
	match, err := s.matchRepo.FindByIDWithDetails(id)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, errs.ErrNotFound("Match not found")
		}
		slog.Error("failed to fetch match", "error", err, "match_id", id)
		return nil, errs.ErrInternal("Internal server error")
	}

	resp := toMatchResponse(*match)
	return &resp, nil
}

func (s *matchService) Create(req dto.CreateMatchRequest) (*dto.MatchResponse, error) {
	homeTeamID, err := uuid.Parse(req.HomeTeamID)
	if err != nil {
		return nil, errs.ErrBadRequest("Invalid home_team_id format")
	}
	awayTeamID, err := uuid.Parse(req.AwayTeamID)
	if err != nil {
		return nil, errs.ErrBadRequest("Invalid away_team_id format")
	}

	// Validate: home_team_id != away_team_id
	if homeTeamID == awayTeamID {
		return nil, errs.ErrBadRequest("Home team and away team cannot be the same")
	}

	// Verify both teams exist
	if _, err := s.teamRepo.FindByID(homeTeamID); err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, errs.ErrNotFound("Home team not found")
		}
		slog.Error("failed to fetch home team", "error", err)
		return nil, errs.ErrInternal("Internal server error")
	}
	if _, err := s.teamRepo.FindByID(awayTeamID); err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, errs.ErrNotFound("Away team not found")
		}
		slog.Error("failed to fetch away team", "error", err)
		return nil, errs.ErrInternal("Internal server error")
	}

	match := model.Match{
		HomeTeamID: homeTeamID,
		AwayTeamID: awayTeamID,
		MatchDate:  req.MatchDate,
		MatchTime:  req.MatchTime,
		Status:     "scheduled",
		HomeScore:  0,
		AwayScore:  0,
	}

	if err := s.matchRepo.Create(&match); err != nil {
		slog.Error("failed to create match", "error", err)
		return nil, errs.ErrInternal("Internal server error")
	}

	// Reload with teams preloaded
	created, err := s.matchRepo.FindByID(match.ID)
	if err != nil {
		slog.Error("failed to reload created match", "error", err)
		return nil, errs.ErrInternal("Internal server error")
	}

	resp := toMatchResponse(*created)
	return &resp, nil
}

func (s *matchService) Update(id uuid.UUID, req dto.UpdateMatchRequest) (*dto.MatchResponse, error) {
	match, err := s.matchRepo.FindByID(id)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, errs.ErrNotFound("Match not found")
		}
		slog.Error("failed to fetch match for update", "error", err, "match_id", id)
		return nil, errs.ErrInternal("Internal server error")
	}

	// Cannot update a completed match schedule
	if match.Status == "completed" {
		return nil, errs.ErrBadRequest("Cannot update schedule of a completed match")
	}

	homeTeamID, err := uuid.Parse(req.HomeTeamID)
	if err != nil {
		return nil, errs.ErrBadRequest("Invalid home_team_id format")
	}
	awayTeamID, err := uuid.Parse(req.AwayTeamID)
	if err != nil {
		return nil, errs.ErrBadRequest("Invalid away_team_id format")
	}

	if homeTeamID == awayTeamID {
		return nil, errs.ErrBadRequest("Home team and away team cannot be the same")
	}

	// Verify both teams exist
	if _, err := s.teamRepo.FindByID(homeTeamID); err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, errs.ErrNotFound("Home team not found")
		}
		slog.Error("failed to fetch home team for update", "error", err, "match_id", id)
		return nil, errs.ErrInternal("Internal server error")
	}
	if _, err := s.teamRepo.FindByID(awayTeamID); err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, errs.ErrNotFound("Away team not found")
		}
		slog.Error("failed to fetch away team for update", "error", err, "match_id", id)
		return nil, errs.ErrInternal("Internal server error")
	}

	match.HomeTeamID = homeTeamID
	match.AwayTeamID = awayTeamID
	match.MatchDate = req.MatchDate
	match.MatchTime = req.MatchTime

	if err := s.matchRepo.Update(match); err != nil {
		slog.Error("failed to update match", "error", err, "match_id", id)
		return nil, errs.ErrInternal("Internal server error")
	}

	resp := toMatchResponse(*match)
	return &resp, nil
}

func (s *matchService) Delete(id uuid.UUID) error {
	_, err := s.matchRepo.FindByID(id)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return errs.ErrNotFound("Match not found")
		}
		slog.Error("failed to fetch match for delete", "error", err, "match_id", id)
		return errs.ErrInternal("Internal server error")
	}

	if err := s.matchRepo.Delete(id); err != nil {
		slog.Error("failed to delete match", "error", err, "match_id", id)
		return errs.ErrInternal("Internal server error")
	}

	return nil
}

// SubmitResult processes match results: validates goals, calculates scores, and transitions match status.
func (s *matchService) SubmitResult(matchID uuid.UUID, req dto.MatchResultRequest) (*dto.MatchResponse, error) {
	match, err := s.matchRepo.FindByID(matchID)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, errs.ErrNotFound("Match not found")
		}
		slog.Error("failed to fetch match for result", "error", err, "match_id", matchID)
		return nil, errs.ErrInternal("Internal server error")
	}

	if match.Status == "completed" {
		return nil, errs.ErrBadRequest("Match result already submitted. Use PUT to update.")
	}

	return s.processResult(match, req)
}

// UpdateResult replaces existing match results with new ones.
func (s *matchService) UpdateResult(matchID uuid.UUID, req dto.MatchResultRequest) (*dto.MatchResponse, error) {
	match, err := s.matchRepo.FindByID(matchID)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, errs.ErrNotFound("Match not found")
		}
		slog.Error("failed to fetch match for result update", "error", err, "match_id", matchID)
		return nil, errs.ErrInternal("Internal server error")
	}

	if match.Status != "completed" {
		return nil, errs.ErrBadRequest("Cannot update result of a match that has not been completed. Use POST to submit first.")
	}

	// Delete old goals before inserting new ones
	if err := s.goalRepo.DeleteByMatchID(matchID); err != nil {
		slog.Error("failed to delete old goals", "error", err, "match_id", matchID)
		return nil, errs.ErrInternal("Internal server error")
	}

	return s.processResult(match, req)
}

// processResult validates goals, calculates scores, and saves everything.
func (s *matchService) processResult(match *model.Match, req dto.MatchResultRequest) (*dto.MatchResponse, error) {
	homeScore := 0
	awayScore := 0
	goals := make([]model.Goal, 0, len(req.Goals))

	for i, goalInput := range req.Goals {
		playerID, err := uuid.Parse(goalInput.PlayerID)
		if err != nil {
			return nil, errs.ErrBadRequest(fmt.Sprintf("Goal #%d: invalid player_id format", i+1))
		}
		teamID, err := uuid.Parse(goalInput.TeamID)
		if err != nil {
			return nil, errs.ErrBadRequest(fmt.Sprintf("Goal #%d: invalid team_id format", i+1))
		}

		// Validate team_id is either home or away team
		if teamID != match.HomeTeamID && teamID != match.AwayTeamID {
			return nil, errs.ErrBadRequest(fmt.Sprintf("Goal #%d: team_id must be either home or away team", i+1))
		}

		// Validate player belongs to the specified team
		player, err := s.playerRepo.FindByID(playerID)
		if err != nil {
			if errors.Is(err, gorm.ErrRecordNotFound) {
				return nil, errs.ErrNotFound(fmt.Sprintf("Goal #%d: player not found", i+1))
			}
			slog.Error("failed to fetch player for goal validation", "error", err)
			return nil, errs.ErrInternal("Internal server error")
		}
		if player.TeamID != teamID {
			return nil, errs.ErrBadRequest(fmt.Sprintf("Goal #%d: player does not belong to the specified team", i+1))
		}

		// Count scores
		if teamID == match.HomeTeamID {
			homeScore++
		} else {
			awayScore++
		}

		goals = append(goals, model.Goal{
			MatchID:  match.ID,
			PlayerID: playerID,
			TeamID:   teamID,
			Minute:   goalInput.Minute,
		})
	}

	// Batch insert goals
	if len(goals) > 0 {
		if err := s.goalRepo.CreateBatch(goals); err != nil {
			slog.Error("failed to create goals", "error", err)
			return nil, errs.ErrInternal("Internal server error")
		}
	}

	// Update match scores and status
	match.HomeScore = homeScore
	match.AwayScore = awayScore
	match.Status = "completed"

	if err := s.matchRepo.Update(match); err != nil {
		slog.Error("failed to update match with results", "error", err)
		return nil, errs.ErrInternal("Internal server error")
	}

	// Reload with full details
	updated, err := s.matchRepo.FindByIDWithDetails(match.ID)
	if err != nil {
		slog.Error("failed to reload match after result", "error", err)
		return nil, errs.ErrInternal("Internal server error")
	}

	resp := toMatchResponse(*updated)
	return &resp, nil
}

// toMatchResponse converts a model.Match to dto.MatchResponse.
func toMatchResponse(match model.Match) dto.MatchResponse {
	resp := dto.MatchResponse{
		ID:         match.ID.String(),
		HomeTeamID: match.HomeTeamID.String(),
		AwayTeamID: match.AwayTeamID.String(),
		MatchDate:  match.MatchDate,
		MatchTime:  match.MatchTime,
		HomeScore:  match.HomeScore,
		AwayScore:  match.AwayScore,
		Status:     match.Status,
		CreatedAt:  match.CreatedAt.Format("2006-01-02T15:04:05Z"),
		UpdatedAt:  match.UpdatedAt.Format("2006-01-02T15:04:05Z"),
	}

	if match.HomeTeam != nil {
		homeTeam := toTeamResponse(*match.HomeTeam)
		resp.HomeTeam = &homeTeam
	}
	if match.AwayTeam != nil {
		awayTeam := toTeamResponse(*match.AwayTeam)
		resp.AwayTeam = &awayTeam
	}

	if len(match.Goals) > 0 {
		resp.Goals = make([]dto.GoalResponse, len(match.Goals))
		for i, goal := range match.Goals {
			resp.Goals[i] = toGoalResponse(goal)
		}
	}

	return resp
}

// toGoalResponse converts a model.Goal to dto.GoalResponse.
func toGoalResponse(goal model.Goal) dto.GoalResponse {
	resp := dto.GoalResponse{
		ID:        goal.ID.String(),
		MatchID:   goal.MatchID.String(),
		PlayerID:  goal.PlayerID.String(),
		TeamID:    goal.TeamID.String(),
		Minute:    goal.Minute,
		CreatedAt: goal.CreatedAt.Format("2006-01-02T15:04:05Z"),
	}

	if goal.Player != nil {
		playerResp := toPlayerResponse(*goal.Player)
		resp.Player = &playerResp
	}
	if goal.Team != nil {
		teamResp := toTeamResponse(*goal.Team)
		resp.Team = &teamResp
	}

	return resp
}
