package service

import (
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/mhakimsaputra17/xyz-football-api/internal/dto"
	"github.com/mhakimsaputra17/xyz-football-api/internal/mocks"
	"github.com/mhakimsaputra17/xyz-football-api/internal/model"
	"github.com/mhakimsaputra17/xyz-football-api/pkg/errs"
	"github.com/stretchr/testify/assert"
	"gorm.io/gorm"
)

func newTestReportService(t *testing.T) (*reportService, *mocks.MockMatchRepository, *mocks.MockGoalRepository) {
	matchRepo := mocks.NewMockMatchRepository(t)
	goalRepo := mocks.NewMockGoalRepository(t)
	svc := &reportService{matchRepo: matchRepo, goalRepo: goalRepo}
	return svc, matchRepo, goalRepo
}

func TestReportService_GetMatchReports(t *testing.T) {
	homeID := uuid.Must(uuid.NewV7())
	awayID := uuid.Must(uuid.NewV7())

	homeTeam := model.Team{
		Base: model.Base{ID: homeID, CreatedAt: time.Now(), UpdatedAt: time.Now()},
		Name: "Persija Jakarta",
	}
	awayTeam := model.Team{
		Base: model.Base{ID: awayID, CreatedAt: time.Now(), UpdatedAt: time.Now()},
		Name: "Persib Bandung",
	}

	tests := []struct {
		name    string
		setup   func(*mocks.MockMatchRepository)
		wantErr bool
		wantLen int
	}{
		{
			name: "success with completed matches",
			setup: func(mr *mocks.MockMatchRepository) {
				matches := []model.Match{
					{
						Base:       model.Base{ID: uuid.Must(uuid.NewV7()), CreatedAt: time.Now(), UpdatedAt: time.Now()},
						HomeTeamID: homeID,
						AwayTeamID: awayID,
						MatchDate:  "2026-03-15",
						MatchTime:  "19:30",
						HomeScore:  2,
						AwayScore:  1,
						Status:     "completed",
						HomeTeam:   &homeTeam,
						AwayTeam:   &awayTeam,
					},
				}
				mr.EXPECT().FindCompletedMatches(0, 10).Return(matches, nil)
				mr.EXPECT().CountCompletedMatches().Return(int64(1), nil)
			},
			wantLen: 1,
		},
		{
			name: "success empty",
			setup: func(mr *mocks.MockMatchRepository) {
				mr.EXPECT().FindCompletedMatches(0, 10).Return([]model.Match{}, nil)
				mr.EXPECT().CountCompletedMatches().Return(int64(0), nil)
			},
			wantLen: 0,
		},
		{
			name: "db error",
			setup: func(mr *mocks.MockMatchRepository) {
				mr.EXPECT().FindCompletedMatches(0, 10).Return(nil, gorm.ErrInvalidDB)
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			svc, matchRepo, _ := newTestReportService(t)
			tt.setup(matchRepo)

			pagination := dto.PaginationQuery{Page: 1, PerPage: 10, SortBy: "created_at", SortOrder: "desc"}
			reports, meta, err := svc.GetMatchReports(pagination)

			if tt.wantErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
				assert.Len(t, reports, tt.wantLen)
				if tt.wantLen > 0 {
					assert.NotNil(t, meta)
					assert.Equal(t, "Home Win", reports[0].MatchResult)
				}
			}
			matchRepo.AssertExpectations(t)
		})
	}
}

func TestReportService_GetMatchReportByID(t *testing.T) {
	homeID := uuid.Must(uuid.NewV7())
	awayID := uuid.Must(uuid.NewV7())
	matchID := uuid.Must(uuid.NewV7())
	playerHomeID := uuid.Must(uuid.NewV7())
	playerAwayID := uuid.Must(uuid.NewV7())

	homeTeam := model.Team{
		Base: model.Base{ID: homeID, CreatedAt: time.Now(), UpdatedAt: time.Now()},
		Name: "Persija Jakarta",
	}
	awayTeam := model.Team{
		Base: model.Base{ID: awayID, CreatedAt: time.Now(), UpdatedAt: time.Now()},
		Name: "Persib Bandung",
	}

	tests := []struct {
		name        string
		setup       func(*mocks.MockMatchRepository)
		wantErr     bool
		errContains string
		wantResult  string // expected match_result
		wantTopGoal int    // expected top scorer goals
	}{
		{
			name: "success home win with top scorer",
			setup: func(mr *mocks.MockMatchRepository) {
				mr.EXPECT().FindByIDWithDetails(matchID).Return(&model.Match{
					Base:       model.Base{ID: matchID, CreatedAt: time.Now(), UpdatedAt: time.Now()},
					HomeTeamID: homeID,
					AwayTeamID: awayID,
					MatchDate:  "2026-03-15",
					MatchTime:  "19:30",
					HomeScore:  2,
					AwayScore:  1,
					Status:     "completed",
					HomeTeam:   &homeTeam,
					AwayTeam:   &awayTeam,
					Goals: []model.Goal{
						{
							Base:     model.Base{ID: uuid.Must(uuid.NewV7())},
							MatchID:  matchID,
							PlayerID: playerHomeID,
							TeamID:   homeID,
							Minute:   23,
							Player:   &model.Player{Base: model.Base{ID: playerHomeID}, Name: "Bambang"},
							Team:     &homeTeam,
						},
						{
							Base:     model.Base{ID: uuid.Must(uuid.NewV7())},
							MatchID:  matchID,
							PlayerID: playerAwayID,
							TeamID:   awayID,
							Minute:   45,
							Player:   &model.Player{Base: model.Base{ID: playerAwayID}, Name: "Atep"},
							Team:     &awayTeam,
						},
						{
							Base:     model.Base{ID: uuid.Must(uuid.NewV7())},
							MatchID:  matchID,
							PlayerID: playerHomeID,
							TeamID:   homeID,
							Minute:   78,
							Player:   &model.Player{Base: model.Base{ID: playerHomeID}, Name: "Bambang"},
							Team:     &homeTeam,
						},
					},
				}, nil)
				mr.EXPECT().CountWins(homeID).Return(5, nil)
				mr.EXPECT().CountWins(awayID).Return(3, nil)
			},
			wantResult:  "Home Win",
			wantTopGoal: 2,
		},
		{
			name: "success draw",
			setup: func(mr *mocks.MockMatchRepository) {
				mr.EXPECT().FindByIDWithDetails(matchID).Return(&model.Match{
					Base:       model.Base{ID: matchID, CreatedAt: time.Now(), UpdatedAt: time.Now()},
					HomeTeamID: homeID,
					AwayTeamID: awayID,
					MatchDate:  "2026-03-20",
					MatchTime:  "20:00",
					HomeScore:  1,
					AwayScore:  1,
					Status:     "completed",
					HomeTeam:   &homeTeam,
					AwayTeam:   &awayTeam,
					Goals: []model.Goal{
						{
							Base:     model.Base{ID: uuid.Must(uuid.NewV7())},
							MatchID:  matchID,
							PlayerID: playerHomeID,
							TeamID:   homeID,
							Minute:   30,
							Player:   &model.Player{Base: model.Base{ID: playerHomeID}, Name: "Bambang"},
							Team:     &homeTeam,
						},
						{
							Base:     model.Base{ID: uuid.Must(uuid.NewV7())},
							MatchID:  matchID,
							PlayerID: playerAwayID,
							TeamID:   awayID,
							Minute:   60,
							Player:   &model.Player{Base: model.Base{ID: playerAwayID}, Name: "Atep"},
							Team:     &awayTeam,
						},
					},
				}, nil)
				mr.EXPECT().CountWins(homeID).Return(2, nil)
				mr.EXPECT().CountWins(awayID).Return(2, nil)
			},
			wantResult:  "Draw",
			wantTopGoal: 1,
		},
		{
			name: "match not found",
			setup: func(mr *mocks.MockMatchRepository) {
				mr.EXPECT().FindByIDWithDetails(matchID).Return(nil, gorm.ErrRecordNotFound)
			},
			wantErr:     true,
			errContains: "Match not found",
		},
		{
			name: "match not completed",
			setup: func(mr *mocks.MockMatchRepository) {
				mr.EXPECT().FindByIDWithDetails(matchID).Return(&model.Match{
					Base:       model.Base{ID: matchID, CreatedAt: time.Now(), UpdatedAt: time.Now()},
					HomeTeamID: homeID,
					AwayTeamID: awayID,
					Status:     "scheduled", // not completed
				}, nil)
			},
			wantErr:     true,
			errContains: "Match has not been completed yet",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			svc, matchRepo, _ := newTestReportService(t)
			tt.setup(matchRepo)

			report, err := svc.GetMatchReportByID(matchID)

			if tt.wantErr {
				assert.Error(t, err)
				var appErr *errs.AppError
				assert.ErrorAs(t, err, &appErr)
				assert.Contains(t, appErr.Message, tt.errContains)
			} else {
				assert.NoError(t, err)
				assert.NotNil(t, report)
				assert.Equal(t, tt.wantResult, report.MatchResult)
				if report.TopScorer != nil {
					assert.Equal(t, tt.wantTopGoal, report.TopScorer.GoalsInMatch)
				}
			}
			matchRepo.AssertExpectations(t)
		})
	}
}

// TestComputeMatchResult tests the match result computation helper.
func TestComputeMatchResult(t *testing.T) {
	tests := []struct {
		name      string
		homeScore int
		awayScore int
		want      string
	}{
		{name: "home win", homeScore: 3, awayScore: 1, want: "Home Win"},
		{name: "away win", homeScore: 0, awayScore: 2, want: "Away Win"},
		{name: "draw", homeScore: 1, awayScore: 1, want: "Draw"},
		{name: "draw zero", homeScore: 0, awayScore: 0, want: "Draw"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := computeMatchResult(tt.homeScore, tt.awayScore)
			assert.Equal(t, tt.want, result)
		})
	}
}
