package service

import (
	"errors"
	"log/slog"

	"github.com/google/uuid"
	"github.com/mhakimsaputra17/xyz-football-api/internal/dto"
	"github.com/mhakimsaputra17/xyz-football-api/internal/repository"
	"github.com/mhakimsaputra17/xyz-football-api/pkg/errs"
	"github.com/mhakimsaputra17/xyz-football-api/pkg/response"
	"gorm.io/gorm"
)

// ReportService defines the contract for match report business logic.
type ReportService interface {
	GetMatchReports(pagination dto.PaginationQuery) ([]dto.MatchReportListItem, *response.PaginationMeta, error)
	GetMatchReportByID(matchID uuid.UUID) (*dto.MatchReportResponse, error)
}

type reportService struct {
	matchRepo repository.MatchRepository
	goalRepo  repository.GoalRepository
}

// NewReportService creates a new ReportService instance.
func NewReportService(matchRepo repository.MatchRepository, goalRepo repository.GoalRepository) ReportService {
	return &reportService{
		matchRepo: matchRepo,
		goalRepo:  goalRepo,
	}
}

// GetMatchReports returns a paginated list of all completed match reports.
func (s *reportService) GetMatchReports(pagination dto.PaginationQuery) ([]dto.MatchReportListItem, *response.PaginationMeta, error) {
	pagination.Sanitize()

	matches, err := s.matchRepo.FindCompletedMatches(pagination.GetOffset(), pagination.PerPage)
	if err != nil {
		slog.Error("failed to fetch completed matches for report", "error", err)
		return nil, nil, errs.ErrInternal("Internal server error")
	}

	total, err := s.matchRepo.CountCompletedMatches()
	if err != nil {
		slog.Error("failed to count completed matches", "error", err)
		return nil, nil, errs.ErrInternal("Internal server error")
	}

	items := make([]dto.MatchReportListItem, len(matches))
	for i, match := range matches {
		items[i] = dto.MatchReportListItem{
			MatchID:     match.ID.String(),
			MatchDate:   match.MatchDate,
			MatchTime:   match.MatchTime,
			HomeScore:   match.HomeScore,
			AwayScore:   match.AwayScore,
			MatchResult: computeMatchResult(match.HomeScore, match.AwayScore),
		}
		if match.HomeTeam != nil {
			items[i].HomeTeam = toTeamResponse(*match.HomeTeam)
		}
		if match.AwayTeam != nil {
			items[i].AwayTeam = toTeamResponse(*match.AwayTeam)
		}
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

	return items, meta, nil
}

// GetMatchReportByID returns a detailed report for a single completed match.
// Includes: match result, goal list, top scorer, and accumulated total wins for both teams.
func (s *reportService) GetMatchReportByID(matchID uuid.UUID) (*dto.MatchReportResponse, error) {
	match, err := s.matchRepo.FindByIDWithDetails(matchID)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, errs.ErrNotFound("Match not found")
		}
		slog.Error("failed to fetch match for report", "error", err, "match_id", matchID)
		return nil, errs.ErrInternal("Internal server error")
	}

	if match.Status != "completed" {
		return nil, errs.ErrBadRequest("Match has not been completed yet")
	}

	// Build goal list for report
	reportGoals := make([]dto.MatchReportGoal, len(match.Goals))
	// Track goal counts per player for top scorer calculation
	type playerGoalCount struct {
		PlayerName string
		TeamName   string
		Count      int
	}
	playerGoals := make(map[uuid.UUID]*playerGoalCount)

	for i, goal := range match.Goals {
		playerName := ""
		teamName := ""
		if goal.Player != nil {
			playerName = goal.Player.Name
		}
		if goal.Team != nil {
			teamName = goal.Team.Name
		}

		reportGoals[i] = dto.MatchReportGoal{
			PlayerName: playerName,
			TeamName:   teamName,
			Minute:     goal.Minute,
		}

		// Accumulate goal count per player
		if _, exists := playerGoals[goal.PlayerID]; !exists {
			playerGoals[goal.PlayerID] = &playerGoalCount{
				PlayerName: playerName,
				TeamName:   teamName,
				Count:      0,
			}
		}
		playerGoals[goal.PlayerID].Count++
	}

	// Determine top scorer (player with most goals in this match)
	var topScorer *dto.TopScorerResponse
	maxGoals := 0
	for _, pg := range playerGoals {
		if pg.Count > maxGoals {
			maxGoals = pg.Count
			topScorer = &dto.TopScorerResponse{
				PlayerName:   pg.PlayerName,
				TeamName:     pg.TeamName,
				GoalsInMatch: pg.Count,
			}
		}
	}

	// Calculate accumulated total wins for both teams across ALL completed matches
	homeTeamWins, err := s.matchRepo.CountWins(match.HomeTeamID)
	if err != nil {
		slog.Error("failed to count home team wins", "error", err)
		return nil, errs.ErrInternal("Internal server error")
	}
	awayTeamWins, err := s.matchRepo.CountWins(match.AwayTeamID)
	if err != nil {
		slog.Error("failed to count away team wins", "error", err)
		return nil, errs.ErrInternal("Internal server error")
	}

	report := &dto.MatchReportResponse{
		MatchID:           match.ID.String(),
		MatchDate:         match.MatchDate,
		MatchTime:         match.MatchTime,
		HomeScore:         match.HomeScore,
		AwayScore:         match.AwayScore,
		MatchResult:       computeMatchResult(match.HomeScore, match.AwayScore),
		Goals:             reportGoals,
		TopScorer:         topScorer,
		HomeTeamTotalWins: homeTeamWins,
		AwayTeamTotalWins: awayTeamWins,
	}

	if match.HomeTeam != nil {
		report.HomeTeam = toTeamResponse(*match.HomeTeam)
	}
	if match.AwayTeam != nil {
		report.AwayTeam = toTeamResponse(*match.AwayTeam)
	}

	return report, nil
}

// computeMatchResult determines the match outcome string.
func computeMatchResult(homeScore, awayScore int) string {
	switch {
	case homeScore > awayScore:
		return "Home Win"
	case awayScore > homeScore:
		return "Away Win"
	default:
		return "Draw"
	}
}
