package jwt

import (
	"time"

	"github.com/golang-jwt/jwt/v5"
	"github.com/google/uuid"
)

// Claims represents the custom JWT claims payload.
type Claims struct {
	AdminID  uuid.UUID `json:"admin_id"`
	Username string    `json:"username"`
	jwt.RegisteredClaims
}

// TokenPair holds an access token and refresh token pair.
type TokenPair struct {
	AccessToken  string `json:"access_token"`
	RefreshToken string `json:"refresh_token"`
}

// Service handles JWT token generation and validation.
type Service struct {
	secret            []byte
	accessExpiration  time.Duration
	refreshExpiration time.Duration
}

// NewService creates a new JWT service with the given configuration.
func NewService(secret string, accessExp, refreshExp time.Duration) *Service {
	return &Service{
		secret:            []byte(secret),
		accessExpiration:  accessExp,
		refreshExpiration: refreshExp,
	}
}

// GenerateAccessToken creates a signed JWT access token for the given admin.
func (s *Service) GenerateAccessToken(adminID uuid.UUID, username string) (string, error) {
	now := time.Now()
	claims := Claims{
		AdminID:  adminID,
		Username: username,
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(now.Add(s.accessExpiration)),
			IssuedAt:  jwt.NewNumericDate(now),
			NotBefore: jwt.NewNumericDate(now),
			Issuer:    "xyz-football-api",
			Subject:   adminID.String(),
		},
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	return token.SignedString(s.secret)
}

// GenerateRefreshToken creates a random refresh token string and returns
// it along with its expiration time.
func (s *Service) GenerateRefreshToken() (string, time.Time, error) {
	id, err := uuid.NewV7()
	if err != nil {
		return "", time.Time{}, err
	}
	expiresAt := time.Now().Add(s.refreshExpiration)
	return id.String(), expiresAt, nil
}

// ValidateAccessToken parses and validates an access token, returning the claims.
func (s *Service) ValidateAccessToken(tokenString string) (*Claims, error) {
	token, err := jwt.ParseWithClaims(tokenString, &Claims{}, func(token *jwt.Token) (any, error) {
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, jwt.ErrSignatureInvalid
		}
		return s.secret, nil
	})
	if err != nil {
		return nil, err
	}

	claims, ok := token.Claims.(*Claims)
	if !ok || !token.Valid {
		return nil, jwt.ErrSignatureInvalid
	}

	return claims, nil
}

// GetRefreshExpiration returns the configured refresh token expiration duration.
func (s *Service) GetRefreshExpiration() time.Duration {
	return s.refreshExpiration
}
