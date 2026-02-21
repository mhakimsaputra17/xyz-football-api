package service

import (
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/mhakimsaputra17/xyz-football-api/internal/mocks"
	"github.com/mhakimsaputra17/xyz-football-api/internal/model"
	"github.com/mhakimsaputra17/xyz-football-api/pkg/errs"
	jwtpkg "github.com/mhakimsaputra17/xyz-football-api/pkg/jwt"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"golang.org/x/crypto/bcrypt"
	"gorm.io/gorm"
)

// newTestAuthService creates an authService with mock dependencies for testing.
func newTestAuthService(t *testing.T) (*authService, *mocks.MockAdminRepository, *mocks.MockRefreshTokenRepository, *jwtpkg.Service) {
	adminRepo := mocks.NewMockAdminRepository(t)
	refreshTokenRepo := mocks.NewMockRefreshTokenRepository(t)
	jwtService := jwtpkg.NewService("test-secret-key-for-unit-testing-256bit", 15*time.Minute, 7*24*time.Hour)

	svc := &authService{
		adminRepo:        adminRepo,
		refreshTokenRepo: refreshTokenRepo,
		jwtService:       jwtService,
	}
	return svc, adminRepo, refreshTokenRepo, jwtService
}

func TestAuthService_Login(t *testing.T) {
	hashedPw, _ := bcrypt.GenerateFromPassword([]byte("password123"), bcrypt.DefaultCost)
	adminID := uuid.Must(uuid.NewV7())

	tests := []struct {
		name        string
		username    string
		password    string
		setup       func(*mocks.MockAdminRepository, *mocks.MockRefreshTokenRepository)
		wantErr     bool
		errContains string
	}{
		{
			name:     "successful login",
			username: "admin",
			password: "password123",
			setup: func(ar *mocks.MockAdminRepository, rr *mocks.MockRefreshTokenRepository) {
				ar.EXPECT().FindByUsername("admin").Return(&model.Admin{
					Base:     model.Base{ID: adminID, CreatedAt: time.Now(), UpdatedAt: time.Now()},
					Username: "admin",
					Password: string(hashedPw),
				}, nil)
				rr.EXPECT().Create(mock.AnythingOfType("*model.RefreshToken")).Return(nil)
			},
			wantErr: false,
		},
		{
			name:     "user not found",
			username: "nonexistent",
			password: "password123",
			setup: func(ar *mocks.MockAdminRepository, rr *mocks.MockRefreshTokenRepository) {
				ar.EXPECT().FindByUsername("nonexistent").Return(nil, gorm.ErrRecordNotFound)
			},
			wantErr:     true,
			errContains: "Invalid username or password",
		},
		{
			name:     "wrong password",
			username: "admin",
			password: "wrongpassword",
			setup: func(ar *mocks.MockAdminRepository, rr *mocks.MockRefreshTokenRepository) {
				ar.EXPECT().FindByUsername("admin").Return(&model.Admin{
					Base:     model.Base{ID: adminID, CreatedAt: time.Now(), UpdatedAt: time.Now()},
					Username: "admin",
					Password: string(hashedPw),
				}, nil)
			},
			wantErr:     true,
			errContains: "Invalid username or password",
		},
		{
			name:     "db error on find",
			username: "admin",
			password: "password123",
			setup: func(ar *mocks.MockAdminRepository, rr *mocks.MockRefreshTokenRepository) {
				ar.EXPECT().FindByUsername("admin").Return(nil, gorm.ErrInvalidDB)
			},
			wantErr:     true,
			errContains: "Internal server error",
		},
		{
			name:     "db error on store refresh token",
			username: "admin",
			password: "password123",
			setup: func(ar *mocks.MockAdminRepository, rr *mocks.MockRefreshTokenRepository) {
				ar.EXPECT().FindByUsername("admin").Return(&model.Admin{
					Base:     model.Base{ID: adminID, CreatedAt: time.Now(), UpdatedAt: time.Now()},
					Username: "admin",
					Password: string(hashedPw),
				}, nil)
				rr.EXPECT().Create(mock.AnythingOfType("*model.RefreshToken")).Return(gorm.ErrInvalidDB)
			},
			wantErr:     true,
			errContains: "Internal server error",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			svc, adminRepo, refreshRepo, _ := newTestAuthService(t)
			tt.setup(adminRepo, refreshRepo)

			tokenPair, admin, err := svc.Login(tt.username, tt.password)

			if tt.wantErr {
				assert.Error(t, err)
				var appErr *errs.AppError
				assert.ErrorAs(t, err, &appErr)
				assert.Contains(t, appErr.Message, tt.errContains)
				assert.Nil(t, tokenPair)
			} else {
				assert.NoError(t, err)
				assert.NotNil(t, tokenPair)
				assert.NotEmpty(t, tokenPair.AccessToken)
				assert.NotEmpty(t, tokenPair.RefreshToken)
				assert.NotNil(t, admin)
				assert.Equal(t, "admin", admin.Username)
			}

			adminRepo.AssertExpectations(t)
			refreshRepo.AssertExpectations(t)
		})
	}
}

func TestAuthService_RefreshToken(t *testing.T) {
	adminID := uuid.Must(uuid.NewV7())

	tests := []struct {
		name        string
		token       string
		setup       func(*mocks.MockAdminRepository, *mocks.MockRefreshTokenRepository)
		wantErr     bool
		errContains string
	}{
		{
			name:  "successful refresh",
			token: "valid-refresh-token",
			setup: func(ar *mocks.MockAdminRepository, rr *mocks.MockRefreshTokenRepository) {
				rr.EXPECT().FindByToken("valid-refresh-token").Return(&model.RefreshToken{
					Base:      model.Base{ID: uuid.Must(uuid.NewV7())},
					AdminID:   adminID,
					Token:     "valid-refresh-token",
					ExpiresAt: time.Now().Add(24 * time.Hour),
				}, nil)
				ar.EXPECT().FindByID(adminID).Return(&model.Admin{
					Base:     model.Base{ID: adminID},
					Username: "admin",
				}, nil)
				rr.EXPECT().DeleteByToken("valid-refresh-token").Return(nil)
				rr.EXPECT().Create(mock.AnythingOfType("*model.RefreshToken")).Return(nil)
			},
			wantErr: false,
		},
		{
			name:  "token not found",
			token: "invalid-token",
			setup: func(ar *mocks.MockAdminRepository, rr *mocks.MockRefreshTokenRepository) {
				rr.EXPECT().FindByToken("invalid-token").Return(nil, gorm.ErrRecordNotFound)
			},
			wantErr:     true,
			errContains: "Invalid refresh token",
		},
		{
			name:  "expired token",
			token: "expired-token",
			setup: func(ar *mocks.MockAdminRepository, rr *mocks.MockRefreshTokenRepository) {
				rr.EXPECT().FindByToken("expired-token").Return(&model.RefreshToken{
					Base:      model.Base{ID: uuid.Must(uuid.NewV7())},
					AdminID:   adminID,
					Token:     "expired-token",
					ExpiresAt: time.Now().Add(-1 * time.Hour), // already expired
				}, nil)
				rr.EXPECT().DeleteByToken("expired-token").Return(nil)
			},
			wantErr:     true,
			errContains: "Refresh token has expired",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			svc, adminRepo, refreshRepo, _ := newTestAuthService(t)
			tt.setup(adminRepo, refreshRepo)

			tokenPair, err := svc.RefreshToken(tt.token)

			if tt.wantErr {
				assert.Error(t, err)
				var appErr *errs.AppError
				assert.ErrorAs(t, err, &appErr)
				assert.Contains(t, appErr.Message, tt.errContains)
			} else {
				assert.NoError(t, err)
				assert.NotNil(t, tokenPair)
				assert.NotEmpty(t, tokenPair.AccessToken)
				assert.NotEmpty(t, tokenPair.RefreshToken)
			}

			adminRepo.AssertExpectations(t)
			refreshRepo.AssertExpectations(t)
		})
	}
}

func TestAuthService_Logout(t *testing.T) {
	tests := []struct {
		name        string
		token       string
		setup       func(*mocks.MockRefreshTokenRepository)
		wantErr     bool
		errContains string
	}{
		{
			name:  "successful logout",
			token: "valid-token",
			setup: func(rr *mocks.MockRefreshTokenRepository) {
				rr.EXPECT().DeleteByToken("valid-token").Return(nil)
			},
			wantErr: false,
		},
		{
			name:  "db error on delete",
			token: "some-token",
			setup: func(rr *mocks.MockRefreshTokenRepository) {
				rr.EXPECT().DeleteByToken("some-token").Return(gorm.ErrInvalidDB)
			},
			wantErr:     true,
			errContains: "Internal server error",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			svc, _, refreshRepo, _ := newTestAuthService(t)
			tt.setup(refreshRepo)

			err := svc.Logout(tt.token)

			if tt.wantErr {
				assert.Error(t, err)
				var appErr *errs.AppError
				assert.ErrorAs(t, err, &appErr)
				assert.Contains(t, appErr.Message, tt.errContains)
			} else {
				assert.NoError(t, err)
			}

			refreshRepo.AssertExpectations(t)
		})
	}
}
