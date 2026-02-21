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
	"github.com/stretchr/testify/mock"
	"gorm.io/gorm"
)

func newTestMatchService(t *testing.T) (*matchService, *mocks.MockMatchRepository, *mocks.MockTeamRepository, *mocks.MockPlayerRepository, *mocks.MockGoalRepository) {
	matchRepo := mocks.NewMockMatchRepository(t)
	teamRepo := mocks.NewMockTeamRepository(t)
	playerRepo := mocks.NewMockPlayerRepository(t)
	goalRepo := mocks.NewMockGoalRepository(t)
	svc := &matchService{
		matchRepo:  matchRepo,
		teamRepo:   teamRepo,
		playerRepo: playerRepo,
		goalRepo:   goalRepo,
	}
	return svc, matchRepo, teamRepo, playerRepo, goalRepo
}

func sampleMatch(homeTeamID, awayTeamID uuid.UUID) model.Match {
	return model.Match{
		Base: model.Base{
			ID:        uuid.Must(uuid.NewV7()),
			CreatedAt: time.Now(),
			UpdatedAt: time.Now(),
		},
		HomeTeamID: homeTeamID,
		AwayTeamID: awayTeamID,
		MatchDate:  "2026-03-15",
		MatchTime:  "19:30",
		HomeScore:  0,
		AwayScore:  0,
		Status:     "scheduled",
	}
}

func TestMatchService_GetAll(t *testing.T) {
	homeID := uuid.Must(uuid.NewV7())
	awayID := uuid.Must(uuid.NewV7())

	tests := []struct {
		name    string
		setup   func(*mocks.MockMatchRepository)
		wantErr bool
		wantLen int
	}{
		{
			name: "success",
			setup: func(mr *mocks.MockMatchRepository) {
				matches := []model.Match{sampleMatch(homeID, awayID)}
				mr.EXPECT().FindAll(0, 10, "created_at", "desc").Return(matches, nil)
				mr.EXPECT().Count().Return(int64(1), nil)
			},
			wantLen: 1,
		},
		{
			name: "db error",
			setup: func(mr *mocks.MockMatchRepository) {
				mr.EXPECT().FindAll(0, 10, "created_at", "desc").Return(nil, gorm.ErrInvalidDB)
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			svc, matchRepo, _, _, _ := newTestMatchService(t)
			tt.setup(matchRepo)

			pagination := dto.PaginationQuery{Page: 1, PerPage: 10, SortBy: "created_at", SortOrder: "desc"}
			matches, meta, err := svc.GetAll(pagination)

			if tt.wantErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
				assert.Len(t, matches, tt.wantLen)
				assert.NotNil(t, meta)
			}
			matchRepo.AssertExpectations(t)
		})
	}
}

func TestMatchService_Create(t *testing.T) {
	homeID := uuid.Must(uuid.NewV7())
	awayID := uuid.Must(uuid.NewV7())
	homeTeam := sampleTeam()
	homeTeam.ID = homeID
	awayTeam := sampleTeam()
	awayTeam.ID = awayID
	awayTeam.Name = "Persib Bandung"

	tests := []struct {
		name        string
		req         dto.CreateMatchRequest
		setup       func(*mocks.MockMatchRepository, *mocks.MockTeamRepository)
		wantErr     bool
		errContains string
	}{
		{
			name: "success",
			req: dto.CreateMatchRequest{
				HomeTeamID: homeID.String(),
				AwayTeamID: awayID.String(),
				MatchDate:  "2026-03-15",
				MatchTime:  "19:30",
			},
			setup: func(mr *mocks.MockMatchRepository, tr *mocks.MockTeamRepository) {
				tr.EXPECT().FindByID(homeID).Return(&homeTeam, nil)
				tr.EXPECT().FindByID(awayID).Return(&awayTeam, nil)
				mr.EXPECT().Create(mock.AnythingOfType("*model.Match")).Return(nil)
				mr.EXPECT().FindByID(mock.AnythingOfType("uuid.UUID")).Return(&model.Match{
					Base:       model.Base{ID: uuid.Must(uuid.NewV7()), CreatedAt: time.Now(), UpdatedAt: time.Now()},
					HomeTeamID: homeID,
					AwayTeamID: awayID,
					MatchDate:  "2026-03-15",
					MatchTime:  "19:30",
					Status:     "scheduled",
					HomeTeam:   &homeTeam,
					AwayTeam:   &awayTeam,
				}, nil)
			},
			wantErr: false,
		},
		{
			name: "same team",
			req: dto.CreateMatchRequest{
				HomeTeamID: homeID.String(),
				AwayTeamID: homeID.String(),
				MatchDate:  "2026-03-15",
				MatchTime:  "19:30",
			},
			setup:       func(mr *mocks.MockMatchRepository, tr *mocks.MockTeamRepository) {},
			wantErr:     true,
			errContains: "Home team and away team cannot be the same",
		},
		{
			name: "home team not found",
			req: dto.CreateMatchRequest{
				HomeTeamID: homeID.String(),
				AwayTeamID: awayID.String(),
				MatchDate:  "2026-03-15",
				MatchTime:  "19:30",
			},
			setup: func(mr *mocks.MockMatchRepository, tr *mocks.MockTeamRepository) {
				tr.EXPECT().FindByID(homeID).Return(nil, gorm.ErrRecordNotFound)
			},
			wantErr:     true,
			errContains: "Home team not found",
		},
		{
			name: "away team not found",
			req: dto.CreateMatchRequest{
				HomeTeamID: homeID.String(),
				AwayTeamID: awayID.String(),
				MatchDate:  "2026-03-15",
				MatchTime:  "19:30",
			},
			setup: func(mr *mocks.MockMatchRepository, tr *mocks.MockTeamRepository) {
				tr.EXPECT().FindByID(homeID).Return(&homeTeam, nil)
				tr.EXPECT().FindByID(awayID).Return(nil, gorm.ErrRecordNotFound)
			},
			wantErr:     true,
			errContains: "Away team not found",
		},
		{
			name: "invalid home team id",
			req: dto.CreateMatchRequest{
				HomeTeamID: "not-a-uuid",
				AwayTeamID: awayID.String(),
				MatchDate:  "2026-03-15",
				MatchTime:  "19:30",
			},
			setup:       func(mr *mocks.MockMatchRepository, tr *mocks.MockTeamRepository) {},
			wantErr:     true,
			errContains: "Invalid home_team_id",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			svc, matchRepo, teamRepo, _, _ := newTestMatchService(t)
			tt.setup(matchRepo, teamRepo)

			result, err := svc.Create(tt.req)

			if tt.wantErr {
				assert.Error(t, err)
				var appErr *errs.AppError
				assert.ErrorAs(t, err, &appErr)
				assert.Contains(t, appErr.Message, tt.errContains)
			} else {
				assert.NoError(t, err)
				assert.NotNil(t, result)
				assert.Equal(t, "scheduled", result.Status)
			}
			matchRepo.AssertExpectations(t)
			teamRepo.AssertExpectations(t)
		})
	}
}

func TestMatchService_Delete(t *testing.T) {
	matchID := uuid.Must(uuid.NewV7())
	homeID := uuid.Must(uuid.NewV7())
	awayID := uuid.Must(uuid.NewV7())

	tests := []struct {
		name        string
		setup       func(*mocks.MockMatchRepository)
		wantErr     bool
		errContains string
	}{
		{
			name: "success",
			setup: func(mr *mocks.MockMatchRepository) {
				m := sampleMatch(homeID, awayID)
				m.ID = matchID
				mr.EXPECT().FindByID(matchID).Return(&m, nil)
				mr.EXPECT().Delete(matchID).Return(nil)
			},
			wantErr: false,
		},
		{
			name: "not found",
			setup: func(mr *mocks.MockMatchRepository) {
				mr.EXPECT().FindByID(matchID).Return(nil, gorm.ErrRecordNotFound)
			},
			wantErr:     true,
			errContains: "Match not found",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			svc, matchRepo, _, _, _ := newTestMatchService(t)
			tt.setup(matchRepo)

			err := svc.Delete(matchID)

			if tt.wantErr {
				assert.Error(t, err)
				var appErr *errs.AppError
				assert.ErrorAs(t, err, &appErr)
				assert.Contains(t, appErr.Message, tt.errContains)
			} else {
				assert.NoError(t, err)
			}
			matchRepo.AssertExpectations(t)
		})
	}
}

func TestMatchService_SubmitResult(t *testing.T) {
	homeID := uuid.Must(uuid.NewV7())
	awayID := uuid.Must(uuid.NewV7())
	matchID := uuid.Must(uuid.NewV7())
	playerHomeID := uuid.Must(uuid.NewV7())
	playerAwayID := uuid.Must(uuid.NewV7())

	homeTeam := sampleTeam()
	homeTeam.ID = homeID
	awayTeam := sampleTeam()
	awayTeam.ID = awayID
	awayTeam.Name = "Persib Bandung"

	tests := []struct {
		name        string
		req         dto.MatchResultRequest
		setup       func(*mocks.MockMatchRepository, *mocks.MockPlayerRepository, *mocks.MockGoalRepository)
		wantErr     bool
		errContains string
	}{
		{
			name: "success 2-1",
			req: dto.MatchResultRequest{
				Goals: []dto.GoalInput{
					{PlayerID: playerHomeID.String(), TeamID: homeID.String(), Minute: 23},
					{PlayerID: playerAwayID.String(), TeamID: awayID.String(), Minute: 45},
					{PlayerID: playerHomeID.String(), TeamID: homeID.String(), Minute: 78},
				},
			},
			setup: func(mr *mocks.MockMatchRepository, pr *mocks.MockPlayerRepository, gr *mocks.MockGoalRepository) {
				m := sampleMatch(homeID, awayID)
				m.ID = matchID
				m.Status = "scheduled"
				mr.EXPECT().FindByID(matchID).Return(&m, nil)

				// Validate players
				pr.EXPECT().FindByID(playerHomeID).Return(&model.Player{
					Base:   model.Base{ID: playerHomeID},
					TeamID: homeID,
					Name:   "Bambang",
				}, nil).Times(2)
				pr.EXPECT().FindByID(playerAwayID).Return(&model.Player{
					Base:   model.Base{ID: playerAwayID},
					TeamID: awayID,
					Name:   "Atep",
				}, nil)

				gr.EXPECT().CreateBatch(mock.AnythingOfType("[]model.Goal")).Return(nil)
				mr.EXPECT().Update(mock.AnythingOfType("*model.Match")).Return(nil)

				// Reload with details
				completedMatch := m
				completedMatch.HomeScore = 2
				completedMatch.AwayScore = 1
				completedMatch.Status = "completed"
				completedMatch.HomeTeam = &homeTeam
				completedMatch.AwayTeam = &awayTeam
				mr.EXPECT().FindByIDWithDetails(matchID).Return(&completedMatch, nil)
			},
			wantErr: false,
		},
		{
			name: "match already completed",
			req: dto.MatchResultRequest{
				Goals: []dto.GoalInput{
					{PlayerID: playerHomeID.String(), TeamID: homeID.String(), Minute: 10},
				},
			},
			setup: func(mr *mocks.MockMatchRepository, pr *mocks.MockPlayerRepository, gr *mocks.MockGoalRepository) {
				m := sampleMatch(homeID, awayID)
				m.ID = matchID
				m.Status = "completed"
				mr.EXPECT().FindByID(matchID).Return(&m, nil)
			},
			wantErr:     true,
			errContains: "Match result already submitted",
		},
		{
			name: "player does not belong to team",
			req: dto.MatchResultRequest{
				Goals: []dto.GoalInput{
					{PlayerID: playerHomeID.String(), TeamID: homeID.String(), Minute: 23},
				},
			},
			setup: func(mr *mocks.MockMatchRepository, pr *mocks.MockPlayerRepository, gr *mocks.MockGoalRepository) {
				m := sampleMatch(homeID, awayID)
				m.ID = matchID
				m.Status = "scheduled"
				mr.EXPECT().FindByID(matchID).Return(&m, nil)

				// Player belongs to away team but goal says home team
				pr.EXPECT().FindByID(playerHomeID).Return(&model.Player{
					Base:   model.Base{ID: playerHomeID},
					TeamID: awayID, // wrong team!
					Name:   "Wrong Player",
				}, nil)
			},
			wantErr:     true,
			errContains: "Player does not belong to the specified team",
		},
		{
			name: "goal team not in match",
			req: dto.MatchResultRequest{
				Goals: []dto.GoalInput{
					{PlayerID: playerHomeID.String(), TeamID: uuid.Must(uuid.NewV7()).String(), Minute: 23},
				},
			},
			setup: func(mr *mocks.MockMatchRepository, pr *mocks.MockPlayerRepository, gr *mocks.MockGoalRepository) {
				m := sampleMatch(homeID, awayID)
				m.ID = matchID
				m.Status = "scheduled"
				mr.EXPECT().FindByID(matchID).Return(&m, nil)
			},
			wantErr:     true,
			errContains: "Goal team_id must be either home or away team",
		},
		{
			name: "match not found",
			req: dto.MatchResultRequest{
				Goals: []dto.GoalInput{
					{PlayerID: playerHomeID.String(), TeamID: homeID.String(), Minute: 10},
				},
			},
			setup: func(mr *mocks.MockMatchRepository, pr *mocks.MockPlayerRepository, gr *mocks.MockGoalRepository) {
				mr.EXPECT().FindByID(matchID).Return(nil, gorm.ErrRecordNotFound)
			},
			wantErr:     true,
			errContains: "Match not found",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			svc, matchRepo, _, playerRepo, goalRepo := newTestMatchService(t)
			tt.setup(matchRepo, playerRepo, goalRepo)

			result, err := svc.SubmitResult(matchID, tt.req)

			if tt.wantErr {
				assert.Error(t, err)
				var appErr *errs.AppError
				assert.ErrorAs(t, err, &appErr)
				assert.Contains(t, appErr.Message, tt.errContains)
			} else {
				assert.NoError(t, err)
				assert.NotNil(t, result)
				assert.Equal(t, "completed", result.Status)
				assert.Equal(t, 2, result.HomeScore)
				assert.Equal(t, 1, result.AwayScore)
			}
			matchRepo.AssertExpectations(t)
			playerRepo.AssertExpectations(t)
			goalRepo.AssertExpectations(t)
		})
	}
}

func TestMatchService_UpdateResult(t *testing.T) {
	homeID := uuid.Must(uuid.NewV7())
	awayID := uuid.Must(uuid.NewV7())
	matchID := uuid.Must(uuid.NewV7())
	playerID := uuid.Must(uuid.NewV7())

	homeTeam := sampleTeam()
	homeTeam.ID = homeID

	tests := []struct {
		name        string
		req         dto.MatchResultRequest
		setup       func(*mocks.MockMatchRepository, *mocks.MockPlayerRepository, *mocks.MockGoalRepository)
		wantErr     bool
		errContains string
	}{
		{
			name: "success update",
			req: dto.MatchResultRequest{
				Goals: []dto.GoalInput{
					{PlayerID: playerID.String(), TeamID: homeID.String(), Minute: 55},
				},
			},
			setup: func(mr *mocks.MockMatchRepository, pr *mocks.MockPlayerRepository, gr *mocks.MockGoalRepository) {
				m := sampleMatch(homeID, awayID)
				m.ID = matchID
				m.Status = "completed"
				mr.EXPECT().FindByID(matchID).Return(&m, nil)
				gr.EXPECT().DeleteByMatchID(matchID).Return(nil)

				pr.EXPECT().FindByID(playerID).Return(&model.Player{
					Base:   model.Base{ID: playerID},
					TeamID: homeID,
					Name:   "Bambang",
				}, nil)

				gr.EXPECT().CreateBatch(mock.AnythingOfType("[]model.Goal")).Return(nil)
				mr.EXPECT().Update(mock.AnythingOfType("*model.Match")).Return(nil)

				updatedMatch := m
				updatedMatch.HomeScore = 1
				updatedMatch.AwayScore = 0
				updatedMatch.HomeTeam = &homeTeam
				mr.EXPECT().FindByIDWithDetails(matchID).Return(&updatedMatch, nil)
			},
			wantErr: false,
		},
		{
			name: "match not completed yet",
			req: dto.MatchResultRequest{
				Goals: []dto.GoalInput{
					{PlayerID: playerID.String(), TeamID: homeID.String(), Minute: 10},
				},
			},
			setup: func(mr *mocks.MockMatchRepository, pr *mocks.MockPlayerRepository, gr *mocks.MockGoalRepository) {
				m := sampleMatch(homeID, awayID)
				m.ID = matchID
				m.Status = "scheduled" // not completed
				mr.EXPECT().FindByID(matchID).Return(&m, nil)
			},
			wantErr:     true,
			errContains: "Cannot update result of a match that has not been completed",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			svc, matchRepo, _, playerRepo, goalRepo := newTestMatchService(t)
			tt.setup(matchRepo, playerRepo, goalRepo)

			result, err := svc.UpdateResult(matchID, tt.req)

			if tt.wantErr {
				assert.Error(t, err)
				var appErr *errs.AppError
				assert.ErrorAs(t, err, &appErr)
				assert.Contains(t, appErr.Message, tt.errContains)
			} else {
				assert.NoError(t, err)
				assert.NotNil(t, result)
			}
			matchRepo.AssertExpectations(t)
			playerRepo.AssertExpectations(t)
			goalRepo.AssertExpectations(t)
		})
	}
}

func TestMatchService_Update(t *testing.T) {
	homeID := uuid.Must(uuid.NewV7())
	awayID := uuid.Must(uuid.NewV7())
	newAwayID := uuid.Must(uuid.NewV7())
	matchID := uuid.Must(uuid.NewV7())

	homeTeam := sampleTeam()
	homeTeam.ID = homeID
	awayTeam := sampleTeam()
	awayTeam.ID = newAwayID

	tests := []struct {
		name        string
		req         dto.UpdateMatchRequest
		setup       func(*mocks.MockMatchRepository, *mocks.MockTeamRepository)
		wantErr     bool
		errContains string
	}{
		{
			name: "success",
			req: dto.UpdateMatchRequest{
				HomeTeamID: homeID.String(),
				AwayTeamID: newAwayID.String(),
				MatchDate:  "2026-04-01",
				MatchTime:  "20:00",
			},
			setup: func(mr *mocks.MockMatchRepository, tr *mocks.MockTeamRepository) {
				m := sampleMatch(homeID, awayID)
				m.ID = matchID
				mr.EXPECT().FindByID(matchID).Return(&m, nil)
				tr.EXPECT().FindByID(homeID).Return(&homeTeam, nil)
				tr.EXPECT().FindByID(newAwayID).Return(&awayTeam, nil)
				mr.EXPECT().Update(mock.AnythingOfType("*model.Match")).Return(nil)
			},
			wantErr: false,
		},
		{
			name: "cannot update completed match",
			req: dto.UpdateMatchRequest{
				HomeTeamID: homeID.String(),
				AwayTeamID: newAwayID.String(),
				MatchDate:  "2026-04-01",
				MatchTime:  "20:00",
			},
			setup: func(mr *mocks.MockMatchRepository, tr *mocks.MockTeamRepository) {
				m := sampleMatch(homeID, awayID)
				m.ID = matchID
				m.Status = "completed"
				mr.EXPECT().FindByID(matchID).Return(&m, nil)
			},
			wantErr:     true,
			errContains: "Cannot update schedule of a completed match",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			svc, matchRepo, teamRepo, _, _ := newTestMatchService(t)
			tt.setup(matchRepo, teamRepo)

			result, err := svc.Update(matchID, tt.req)

			if tt.wantErr {
				assert.Error(t, err)
				var appErr *errs.AppError
				assert.ErrorAs(t, err, &appErr)
				assert.Contains(t, appErr.Message, tt.errContains)
			} else {
				assert.NoError(t, err)
				assert.NotNil(t, result)
			}
			matchRepo.AssertExpectations(t)
			teamRepo.AssertExpectations(t)
		})
	}
}
