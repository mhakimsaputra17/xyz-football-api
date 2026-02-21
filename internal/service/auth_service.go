package service

import (
	"errors"
	"log/slog"

	"github.com/mhakimsaputra17/xyz-football-api/internal/model"
	"github.com/mhakimsaputra17/xyz-football-api/internal/repository"
	"github.com/mhakimsaputra17/xyz-football-api/pkg/errs"
	jwtpkg "github.com/mhakimsaputra17/xyz-football-api/pkg/jwt"
	"golang.org/x/crypto/bcrypt"
	"gorm.io/gorm"
)

// AuthService defines the contract for authentication business logic.
type AuthService interface {
	Login(username, password string) (*jwtpkg.TokenPair, *model.Admin, error)
	RefreshToken(refreshToken string) (*jwtpkg.TokenPair, error)
	Logout(refreshToken string) error
}

type authService struct {
	adminRepo        repository.AdminRepository
	refreshTokenRepo repository.RefreshTokenRepository
	jwtService       *jwtpkg.Service
}

// NewAuthService creates a new AuthService instance.
func NewAuthService(
	adminRepo repository.AdminRepository,
	refreshTokenRepo repository.RefreshTokenRepository,
	jwtService *jwtpkg.Service,
) AuthService {
	return &authService{
		adminRepo:        adminRepo,
		refreshTokenRepo: refreshTokenRepo,
		jwtService:       jwtService,
	}
}

// Login authenticates an admin and returns a JWT token pair.
func (s *authService) Login(username, password string) (*jwtpkg.TokenPair, *model.Admin, error) {
	admin, err := s.adminRepo.FindByUsername(username)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, nil, errs.ErrUnauthorized("Invalid username or password")
		}
		slog.Error("failed to find admin by username", "error", err)
		return nil, nil, errs.ErrInternal("Internal server error")
	}

	// Compare password with bcrypt hash
	if err := bcrypt.CompareHashAndPassword([]byte(admin.Password), []byte(password)); err != nil {
		return nil, nil, errs.ErrUnauthorized("Invalid username or password")
	}

	// Generate access token
	accessToken, err := s.jwtService.GenerateAccessToken(admin.ID, admin.Username)
	if err != nil {
		slog.Error("failed to generate access token", "error", err)
		return nil, nil, errs.ErrInternal("Internal server error")
	}

	// Generate refresh token and store in DB
	refreshTokenStr, expiresAt, err := s.jwtService.GenerateRefreshToken()
	if err != nil {
		slog.Error("failed to generate refresh token", "error", err)
		return nil, nil, errs.ErrInternal("Internal server error")
	}

	refreshToken := &model.RefreshToken{
		AdminID:   admin.ID,
		Token:     refreshTokenStr,
		ExpiresAt: expiresAt,
	}
	if err := s.refreshTokenRepo.Create(refreshToken); err != nil {
		slog.Error("failed to store refresh token", "error", err)
		return nil, nil, errs.ErrInternal("Internal server error")
	}

	tokenPair := &jwtpkg.TokenPair{
		AccessToken:  accessToken,
		RefreshToken: refreshTokenStr,
	}

	return tokenPair, admin, nil
}

// RefreshToken validates a refresh token and issues a new token pair (token rotation).
func (s *authService) RefreshToken(refreshTokenStr string) (*jwtpkg.TokenPair, error) {
	// Look up refresh token in DB
	storedToken, err := s.refreshTokenRepo.FindByToken(refreshTokenStr)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, errs.ErrUnauthorized("Invalid refresh token")
		}
		slog.Error("failed to find refresh token", "error", err)
		return nil, errs.ErrInternal("Internal server error")
	}

	// Check expiration
	if storedToken.IsExpired() {
		// Clean up expired token
		_ = s.refreshTokenRepo.DeleteByToken(refreshTokenStr)
		return nil, errs.ErrUnauthorized("Refresh token has expired")
	}

	// Look up the admin
	admin, err := s.adminRepo.FindByID(storedToken.AdminID)
	if err != nil {
		slog.Error("failed to find admin for refresh token", "error", err)
		return nil, errs.ErrInternal("Internal server error")
	}

	// Token rotation: delete old refresh token, create new one
	if err := s.refreshTokenRepo.DeleteByToken(refreshTokenStr); err != nil {
		slog.Error("failed to delete old refresh token", "error", err)
		return nil, errs.ErrInternal("Internal server error")
	}

	// Generate new access token
	newAccessToken, err := s.jwtService.GenerateAccessToken(admin.ID, admin.Username)
	if err != nil {
		slog.Error("failed to generate new access token", "error", err)
		return nil, errs.ErrInternal("Internal server error")
	}

	// Generate new refresh token
	newRefreshTokenStr, expiresAt, err := s.jwtService.GenerateRefreshToken()
	if err != nil {
		slog.Error("failed to generate new refresh token", "error", err)
		return nil, errs.ErrInternal("Internal server error")
	}

	newRefreshToken := &model.RefreshToken{
		AdminID:   admin.ID,
		Token:     newRefreshTokenStr,
		ExpiresAt: expiresAt,
	}
	if err := s.refreshTokenRepo.Create(newRefreshToken); err != nil {
		slog.Error("failed to store new refresh token", "error", err)
		return nil, errs.ErrInternal("Internal server error")
	}

	return &jwtpkg.TokenPair{
		AccessToken:  newAccessToken,
		RefreshToken: newRefreshTokenStr,
	}, nil
}

// Logout invalidates a refresh token by hard-deleting it from the database.
func (s *authService) Logout(refreshTokenStr string) error {
	if err := s.refreshTokenRepo.DeleteByToken(refreshTokenStr); err != nil {
		slog.Error("failed to delete refresh token on logout", "error", err)
		return errs.ErrInternal("Internal server error")
	}
	return nil
}
