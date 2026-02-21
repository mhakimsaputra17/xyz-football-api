package handler

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/mhakimsaputra17/xyz-football-api/internal/dto"
	"github.com/mhakimsaputra17/xyz-football-api/internal/service"
	"github.com/mhakimsaputra17/xyz-football-api/pkg/errs"
	"github.com/mhakimsaputra17/xyz-football-api/pkg/response"
)

// AuthHandler handles authentication-related HTTP requests.
type AuthHandler struct {
	authService service.AuthService
}

// NewAuthHandler creates a new AuthHandler instance.
func NewAuthHandler(authService service.AuthService) *AuthHandler {
	return &AuthHandler{authService: authService}
}

// Login handles POST /api/v1/auth/login
// Validates credentials and returns an access + refresh token pair.
//
//	@Summary		Admin login
//	@Description	Authenticate with username and password to receive access and refresh tokens
//	@Tags			Auth
//	@Accept			json
//	@Produce		json
//	@Param			request	body		dto.LoginRequest	true	"Login credentials"
//	@Success		200		{object}	response.Envelope{data=dto.LoginResponse}
//	@Failure		400		{object}	response.Envelope
//	@Failure		401		{object}	response.Envelope
//	@Failure		500		{object}	response.Envelope
//	@Router			/auth/login [post]
func (h *AuthHandler) Login(c *gin.Context) {
	var req dto.LoginRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Error(c, errs.ErrBadRequest("Invalid request body: "+err.Error()))
		return
	}

	tokenPair, admin, err := h.authService.Login(req.Username, req.Password)
	if err != nil {
		handleServiceError(c, err)
		return
	}

	resp := dto.LoginResponse{
		AccessToken:  tokenPair.AccessToken,
		RefreshToken: tokenPair.RefreshToken,
		Admin: dto.AdminResponse{
			ID:       admin.ID.String(),
			Username: admin.Username,
		},
	}

	response.Success(c, http.StatusOK, "Login successful", resp)
}

// Refresh handles POST /api/v1/auth/refresh
// Validates a refresh token and returns a new token pair (token rotation).
//
//	@Summary		Refresh tokens
//	@Description	Exchange a valid refresh token for a new access + refresh token pair (token rotation)
//	@Tags			Auth
//	@Accept			json
//	@Produce		json
//	@Param			request	body		dto.RefreshRequest	true	"Refresh token"
//	@Success		200		{object}	response.Envelope{data=dto.RefreshResponse}
//	@Failure		400		{object}	response.Envelope
//	@Failure		401		{object}	response.Envelope
//	@Failure		500		{object}	response.Envelope
//	@Router			/auth/refresh [post]
func (h *AuthHandler) Refresh(c *gin.Context) {
	var req dto.RefreshRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Error(c, errs.ErrBadRequest("Invalid request body: "+err.Error()))
		return
	}

	tokenPair, err := h.authService.RefreshToken(req.RefreshToken)
	if err != nil {
		handleServiceError(c, err)
		return
	}

	resp := dto.RefreshResponse{
		AccessToken:  tokenPair.AccessToken,
		RefreshToken: tokenPair.RefreshToken,
	}

	response.Success(c, http.StatusOK, "Token refreshed successfully", resp)
}

// Logout handles POST /api/v1/auth/logout
// Invalidates the refresh token by deleting it from the database.
//
//	@Summary		Admin logout
//	@Description	Invalidate a refresh token by removing it from the database
//	@Tags			Auth
//	@Accept			json
//	@Produce		json
//	@Security		BearerAuth
//	@Param			request	body		dto.RefreshRequest	true	"Refresh token to invalidate"
//	@Success		200		{object}	response.Envelope
//	@Failure		400		{object}	response.Envelope
//	@Failure		401		{object}	response.Envelope
//	@Failure		500		{object}	response.Envelope
//	@Router			/auth/logout [post]
func (h *AuthHandler) Logout(c *gin.Context) {
	var req dto.RefreshRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Error(c, errs.ErrBadRequest("Invalid request body: "+err.Error()))
		return
	}

	if err := h.authService.Logout(req.RefreshToken); err != nil {
		handleServiceError(c, err)
		return
	}

	response.Success(c, http.StatusOK, "Logout successful", nil)
}
