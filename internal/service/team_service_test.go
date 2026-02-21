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

func newTestTeamService(t *testing.T) (*teamService, *mocks.MockTeamRepository) {
	teamRepo := mocks.NewMockTeamRepository(t)
	svc := &teamService{teamRepo: teamRepo}
	return svc, teamRepo
}

func sampleTeam() model.Team {
	return model.Team{
		Base: model.Base{
			ID:        uuid.Must(uuid.NewV7()),
			CreatedAt: time.Now(),
			UpdatedAt: time.Now(),
		},
		Name:        "Persija Jakarta",
		LogoURL:     "https://example.com/logo.png",
		FoundedYear: 1928,
		Address:     "Jl. Casablanca",
		City:        "Jakarta",
	}
}

func TestTeamService_GetAll(t *testing.T) {
	tests := []struct {
		name    string
		setup   func(*mocks.MockTeamRepository)
		wantErr bool
		wantLen int
	}{
		{
			name: "success with teams",
			setup: func(tr *mocks.MockTeamRepository) {
				teams := []model.Team{sampleTeam(), sampleTeam()}
				tr.EXPECT().FindAll(0, 10, "created_at", "desc").Return(teams, nil)
				tr.EXPECT().Count().Return(int64(2), nil)
			},
			wantErr: false,
			wantLen: 2,
		},
		{
			name: "success empty list",
			setup: func(tr *mocks.MockTeamRepository) {
				tr.EXPECT().FindAll(0, 10, "created_at", "desc").Return([]model.Team{}, nil)
				tr.EXPECT().Count().Return(int64(0), nil)
			},
			wantErr: false,
			wantLen: 0,
		},
		{
			name: "db error on find",
			setup: func(tr *mocks.MockTeamRepository) {
				tr.EXPECT().FindAll(0, 10, "created_at", "desc").Return(nil, gorm.ErrInvalidDB)
			},
			wantErr: true,
		},
		{
			name: "db error on count",
			setup: func(tr *mocks.MockTeamRepository) {
				tr.EXPECT().FindAll(0, 10, "created_at", "desc").Return([]model.Team{}, nil)
				tr.EXPECT().Count().Return(int64(0), gorm.ErrInvalidDB)
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			svc, teamRepo := newTestTeamService(t)
			tt.setup(teamRepo)

			pagination := dto.PaginationQuery{Page: 1, PerPage: 10, SortBy: "created_at", SortOrder: "desc"}
			teams, meta, err := svc.GetAll(pagination)

			if tt.wantErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
				assert.Len(t, teams, tt.wantLen)
				assert.NotNil(t, meta)
			}
			teamRepo.AssertExpectations(t)
		})
	}
}

func TestTeamService_GetByID(t *testing.T) {
	team := sampleTeam()

	tests := []struct {
		name        string
		id          uuid.UUID
		setup       func(*mocks.MockTeamRepository)
		wantErr     bool
		errContains string
	}{
		{
			name: "success",
			id:   team.ID,
			setup: func(tr *mocks.MockTeamRepository) {
				tr.EXPECT().FindByID(team.ID).Return(&team, nil)
			},
			wantErr: false,
		},
		{
			name: "not found",
			id:   uuid.Must(uuid.NewV7()),
			setup: func(tr *mocks.MockTeamRepository) {
				tr.EXPECT().FindByID(mock.AnythingOfType("uuid.UUID")).Return(nil, gorm.ErrRecordNotFound)
			},
			wantErr:     true,
			errContains: "Team not found",
		},
		{
			name: "db error",
			id:   uuid.Must(uuid.NewV7()),
			setup: func(tr *mocks.MockTeamRepository) {
				tr.EXPECT().FindByID(mock.AnythingOfType("uuid.UUID")).Return(nil, gorm.ErrInvalidDB)
			},
			wantErr:     true,
			errContains: "Internal server error",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			svc, teamRepo := newTestTeamService(t)
			tt.setup(teamRepo)

			result, err := svc.GetByID(tt.id)

			if tt.wantErr {
				assert.Error(t, err)
				var appErr *errs.AppError
				assert.ErrorAs(t, err, &appErr)
				assert.Contains(t, appErr.Message, tt.errContains)
			} else {
				assert.NoError(t, err)
				assert.NotNil(t, result)
				assert.Equal(t, team.Name, result.Name)
			}
			teamRepo.AssertExpectations(t)
		})
	}
}

func TestTeamService_Create(t *testing.T) {
	tests := []struct {
		name    string
		req     dto.CreateTeamRequest
		setup   func(*mocks.MockTeamRepository)
		wantErr bool
	}{
		{
			name: "success",
			req: dto.CreateTeamRequest{
				Name:        "Persija Jakarta",
				LogoURL:     "https://example.com/logo.png",
				FoundedYear: 1928,
				Address:     "Jl. Casablanca",
				City:        "Jakarta",
			},
			setup: func(tr *mocks.MockTeamRepository) {
				tr.EXPECT().Create(mock.AnythingOfType("*model.Team")).Return(nil)
			},
			wantErr: false,
		},
		{
			name: "db error",
			req: dto.CreateTeamRequest{
				Name: "Persija Jakarta",
			},
			setup: func(tr *mocks.MockTeamRepository) {
				tr.EXPECT().Create(mock.AnythingOfType("*model.Team")).Return(gorm.ErrInvalidDB)
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			svc, teamRepo := newTestTeamService(t)
			tt.setup(teamRepo)

			result, err := svc.Create(tt.req)

			if tt.wantErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
				assert.NotNil(t, result)
				assert.Equal(t, tt.req.Name, result.Name)
			}
			teamRepo.AssertExpectations(t)
		})
	}
}

func TestTeamService_Update(t *testing.T) {
	team := sampleTeam()

	tests := []struct {
		name        string
		id          uuid.UUID
		req         dto.UpdateTeamRequest
		setup       func(*mocks.MockTeamRepository)
		wantErr     bool
		errContains string
	}{
		{
			name: "success",
			id:   team.ID,
			req: dto.UpdateTeamRequest{
				Name:        "Persija Updated",
				LogoURL:     "https://example.com/new-logo.png",
				FoundedYear: 1928,
				Address:     "Jl. New Address",
				City:        "Jakarta",
			},
			setup: func(tr *mocks.MockTeamRepository) {
				teamCopy := team
				tr.EXPECT().FindByID(team.ID).Return(&teamCopy, nil)
				tr.EXPECT().Update(mock.AnythingOfType("*model.Team")).Return(nil)
			},
			wantErr: false,
		},
		{
			name: "not found",
			id:   uuid.Must(uuid.NewV7()),
			req:  dto.UpdateTeamRequest{Name: "Test"},
			setup: func(tr *mocks.MockTeamRepository) {
				tr.EXPECT().FindByID(mock.AnythingOfType("uuid.UUID")).Return(nil, gorm.ErrRecordNotFound)
			},
			wantErr:     true,
			errContains: "Team not found",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			svc, teamRepo := newTestTeamService(t)
			tt.setup(teamRepo)

			result, err := svc.Update(tt.id, tt.req)

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
			teamRepo.AssertExpectations(t)
		})
	}
}

func TestTeamService_Delete(t *testing.T) {
	teamID := uuid.Must(uuid.NewV7())

	tests := []struct {
		name        string
		id          uuid.UUID
		setup       func(*mocks.MockTeamRepository)
		wantErr     bool
		errContains string
	}{
		{
			name: "success",
			id:   teamID,
			setup: func(tr *mocks.MockTeamRepository) {
				team := sampleTeam()
				team.ID = teamID
				tr.EXPECT().FindByID(teamID).Return(&team, nil)
				tr.EXPECT().Delete(teamID).Return(nil)
			},
			wantErr: false,
		},
		{
			name: "not found",
			id:   uuid.Must(uuid.NewV7()),
			setup: func(tr *mocks.MockTeamRepository) {
				tr.EXPECT().FindByID(mock.AnythingOfType("uuid.UUID")).Return(nil, gorm.ErrRecordNotFound)
			},
			wantErr:     true,
			errContains: "Team not found",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			svc, teamRepo := newTestTeamService(t)
			tt.setup(teamRepo)

			err := svc.Delete(tt.id)

			if tt.wantErr {
				assert.Error(t, err)
				var appErr *errs.AppError
				assert.ErrorAs(t, err, &appErr)
				assert.Contains(t, appErr.Message, tt.errContains)
			} else {
				assert.NoError(t, err)
			}
			teamRepo.AssertExpectations(t)
		})
	}
}
