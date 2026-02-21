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

func newTestPlayerService(t *testing.T) (*playerService, *mocks.MockPlayerRepository, *mocks.MockTeamRepository) {
	playerRepo := mocks.NewMockPlayerRepository(t)
	teamRepo := mocks.NewMockTeamRepository(t)
	svc := &playerService{playerRepo: playerRepo, teamRepo: teamRepo}
	return svc, playerRepo, teamRepo
}

func samplePlayer(teamID uuid.UUID) model.Player {
	return model.Player{
		Base: model.Base{
			ID:        uuid.Must(uuid.NewV7()),
			CreatedAt: time.Now(),
			UpdatedAt: time.Now(),
		},
		TeamID:       teamID,
		Name:         "Bambang Pamungkas",
		Height:       176,
		Weight:       72,
		Position:     "penyerang",
		JerseyNumber: 20,
	}
}

func TestPlayerService_GetAllByTeamID(t *testing.T) {
	teamID := uuid.Must(uuid.NewV7())
	team := sampleTeam()
	team.ID = teamID

	tests := []struct {
		name    string
		setup   func(*mocks.MockPlayerRepository, *mocks.MockTeamRepository)
		wantErr bool
		wantLen int
	}{
		{
			name: "success with players",
			setup: func(pr *mocks.MockPlayerRepository, tr *mocks.MockTeamRepository) {
				tr.EXPECT().FindByID(teamID).Return(&team, nil)
				players := []model.Player{samplePlayer(teamID), samplePlayer(teamID)}
				pr.EXPECT().FindAllByTeamID(teamID, 0, 10, "created_at", "desc").Return(players, nil)
				pr.EXPECT().CountByTeamID(teamID).Return(int64(2), nil)
			},
			wantErr: false,
			wantLen: 2,
		},
		{
			name: "team not found",
			setup: func(pr *mocks.MockPlayerRepository, tr *mocks.MockTeamRepository) {
				tr.EXPECT().FindByID(teamID).Return(nil, gorm.ErrRecordNotFound)
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			svc, playerRepo, teamRepo := newTestPlayerService(t)
			tt.setup(playerRepo, teamRepo)

			pagination := dto.PaginationQuery{Page: 1, PerPage: 10, SortBy: "created_at", SortOrder: "desc"}
			players, meta, err := svc.GetAllByTeamID(teamID, pagination)

			if tt.wantErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
				assert.Len(t, players, tt.wantLen)
				assert.NotNil(t, meta)
			}
			playerRepo.AssertExpectations(t)
			teamRepo.AssertExpectations(t)
		})
	}
}

func TestPlayerService_GetByID(t *testing.T) {
	teamID := uuid.Must(uuid.NewV7())
	player := samplePlayer(teamID)

	tests := []struct {
		name        string
		setup       func(*mocks.MockPlayerRepository)
		wantErr     bool
		errContains string
	}{
		{
			name: "success",
			setup: func(pr *mocks.MockPlayerRepository) {
				pr.EXPECT().FindByID(player.ID).Return(&player, nil)
			},
			wantErr: false,
		},
		{
			name: "not found",
			setup: func(pr *mocks.MockPlayerRepository) {
				pr.EXPECT().FindByID(player.ID).Return(nil, gorm.ErrRecordNotFound)
			},
			wantErr:     true,
			errContains: "Player not found",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			svc, playerRepo, _ := newTestPlayerService(t)
			tt.setup(playerRepo)

			result, err := svc.GetByID(player.ID)

			if tt.wantErr {
				assert.Error(t, err)
				var appErr *errs.AppError
				assert.ErrorAs(t, err, &appErr)
				assert.Contains(t, appErr.Message, tt.errContains)
			} else {
				assert.NoError(t, err)
				assert.NotNil(t, result)
				assert.Equal(t, player.Name, result.Name)
			}
			playerRepo.AssertExpectations(t)
		})
	}
}

func TestPlayerService_Create(t *testing.T) {
	teamID := uuid.Must(uuid.NewV7())
	team := sampleTeam()
	team.ID = teamID

	tests := []struct {
		name        string
		req         dto.CreatePlayerRequest
		setup       func(*mocks.MockPlayerRepository, *mocks.MockTeamRepository)
		wantErr     bool
		errContains string
	}{
		{
			name: "success",
			req: dto.CreatePlayerRequest{
				Name:         "Bambang Pamungkas",
				Height:       176,
				Weight:       72,
				Position:     "penyerang",
				JerseyNumber: 20,
			},
			setup: func(pr *mocks.MockPlayerRepository, tr *mocks.MockTeamRepository) {
				tr.EXPECT().FindByID(teamID).Return(&team, nil)
				pr.EXPECT().FindByTeamIDAndJerseyNumber(teamID, 20).Return(nil, gorm.ErrRecordNotFound)
				pr.EXPECT().Create(mock.AnythingOfType("*model.Player")).Return(nil)
			},
			wantErr: false,
		},
		{
			name: "jersey number conflict",
			req: dto.CreatePlayerRequest{
				Name:         "Another Player",
				Height:       180,
				Weight:       75,
				Position:     "gelandang",
				JerseyNumber: 20,
			},
			setup: func(pr *mocks.MockPlayerRepository, tr *mocks.MockTeamRepository) {
				tr.EXPECT().FindByID(teamID).Return(&team, nil)
				existingPlayer := samplePlayer(teamID)
				existingPlayer.JerseyNumber = 20
				pr.EXPECT().FindByTeamIDAndJerseyNumber(teamID, 20).Return(&existingPlayer, nil)
			},
			wantErr:     true,
			errContains: "Jersey number already used",
		},
		{
			name: "team not found",
			req: dto.CreatePlayerRequest{
				Name:         "Player",
				Height:       175,
				Weight:       70,
				Position:     "bertahan",
				JerseyNumber: 5,
			},
			setup: func(pr *mocks.MockPlayerRepository, tr *mocks.MockTeamRepository) {
				tr.EXPECT().FindByID(teamID).Return(nil, gorm.ErrRecordNotFound)
			},
			wantErr:     true,
			errContains: "Team not found",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			svc, playerRepo, teamRepo := newTestPlayerService(t)
			tt.setup(playerRepo, teamRepo)

			result, err := svc.Create(teamID, tt.req)

			if tt.wantErr {
				assert.Error(t, err)
				var appErr *errs.AppError
				assert.ErrorAs(t, err, &appErr)
				assert.Contains(t, appErr.Message, tt.errContains)
			} else {
				assert.NoError(t, err)
				assert.NotNil(t, result)
				assert.Equal(t, tt.req.Name, result.Name)
				assert.Equal(t, tt.req.JerseyNumber, result.JerseyNumber)
			}
			playerRepo.AssertExpectations(t)
			teamRepo.AssertExpectations(t)
		})
	}
}

func TestPlayerService_Update(t *testing.T) {
	teamID := uuid.Must(uuid.NewV7())
	player := samplePlayer(teamID)

	tests := []struct {
		name        string
		req         dto.UpdatePlayerRequest
		setup       func(*mocks.MockPlayerRepository)
		wantErr     bool
		errContains string
	}{
		{
			name: "success same jersey number",
			req: dto.UpdatePlayerRequest{
				Name:         "Bambang Updated",
				Height:       178,
				Weight:       73,
				Position:     "penyerang",
				JerseyNumber: 20, // same as existing
			},
			setup: func(pr *mocks.MockPlayerRepository) {
				playerCopy := player
				pr.EXPECT().FindByID(player.ID).Return(&playerCopy, nil)
				pr.EXPECT().Update(mock.AnythingOfType("*model.Player")).Return(nil)
			},
			wantErr: false,
		},
		{
			name: "success changed jersey number",
			req: dto.UpdatePlayerRequest{
				Name:         "Bambang Updated",
				Height:       178,
				Weight:       73,
				Position:     "penyerang",
				JerseyNumber: 10, // different
			},
			setup: func(pr *mocks.MockPlayerRepository) {
				playerCopy := player
				pr.EXPECT().FindByID(player.ID).Return(&playerCopy, nil)
				pr.EXPECT().FindByTeamIDAndJerseyNumber(teamID, 10).Return(nil, gorm.ErrRecordNotFound)
				pr.EXPECT().Update(mock.AnythingOfType("*model.Player")).Return(nil)
			},
			wantErr: false,
		},
		{
			name: "jersey number conflict on update",
			req: dto.UpdatePlayerRequest{
				Name:         "Bambang Updated",
				Height:       178,
				Weight:       73,
				Position:     "penyerang",
				JerseyNumber: 7, // taken by another player
			},
			setup: func(pr *mocks.MockPlayerRepository) {
				playerCopy := player
				pr.EXPECT().FindByID(player.ID).Return(&playerCopy, nil)
				otherPlayer := samplePlayer(teamID)
				otherPlayer.JerseyNumber = 7
				pr.EXPECT().FindByTeamIDAndJerseyNumber(teamID, 7).Return(&otherPlayer, nil)
			},
			wantErr:     true,
			errContains: "Jersey number already used",
		},
		{
			name: "player not found",
			req:  dto.UpdatePlayerRequest{Name: "Test", Height: 175, Weight: 70, Position: "bertahan", JerseyNumber: 5},
			setup: func(pr *mocks.MockPlayerRepository) {
				pr.EXPECT().FindByID(player.ID).Return(nil, gorm.ErrRecordNotFound)
			},
			wantErr:     true,
			errContains: "Player not found",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			svc, playerRepo, _ := newTestPlayerService(t)
			tt.setup(playerRepo)

			result, err := svc.Update(player.ID, tt.req)

			if tt.wantErr {
				assert.Error(t, err)
				var appErr *errs.AppError
				assert.ErrorAs(t, err, &appErr)
				assert.Contains(t, appErr.Message, tt.errContains)
			} else {
				assert.NoError(t, err)
				assert.NotNil(t, result)
				assert.Equal(t, tt.req.Name, result.Name)
			}
			playerRepo.AssertExpectations(t)
		})
	}
}

func TestPlayerService_Delete(t *testing.T) {
	teamID := uuid.Must(uuid.NewV7())
	playerID := uuid.Must(uuid.NewV7())

	tests := []struct {
		name        string
		setup       func(*mocks.MockPlayerRepository)
		wantErr     bool
		errContains string
	}{
		{
			name: "success",
			setup: func(pr *mocks.MockPlayerRepository) {
				player := samplePlayer(teamID)
				player.ID = playerID
				pr.EXPECT().FindByID(playerID).Return(&player, nil)
				pr.EXPECT().Delete(playerID).Return(nil)
			},
			wantErr: false,
		},
		{
			name: "not found",
			setup: func(pr *mocks.MockPlayerRepository) {
				pr.EXPECT().FindByID(playerID).Return(nil, gorm.ErrRecordNotFound)
			},
			wantErr:     true,
			errContains: "Player not found",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			svc, playerRepo, _ := newTestPlayerService(t)
			tt.setup(playerRepo)

			err := svc.Delete(playerID)

			if tt.wantErr {
				assert.Error(t, err)
				var appErr *errs.AppError
				assert.ErrorAs(t, err, &appErr)
				assert.Contains(t, appErr.Message, tt.errContains)
			} else {
				assert.NoError(t, err)
			}
			playerRepo.AssertExpectations(t)
		})
	}
}
