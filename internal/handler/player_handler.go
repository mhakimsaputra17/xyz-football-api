package handler

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/mhakimsaputra17/xyz-football-api/internal/dto"
	"github.com/mhakimsaputra17/xyz-football-api/internal/service"
	"github.com/mhakimsaputra17/xyz-football-api/pkg/errs"
	"github.com/mhakimsaputra17/xyz-football-api/pkg/response"
)

// PlayerHandler handles player-related HTTP requests.
type PlayerHandler struct {
	playerService service.PlayerService
}

// NewPlayerHandler creates a new PlayerHandler instance.
func NewPlayerHandler(playerService service.PlayerService) *PlayerHandler {
	return &PlayerHandler{playerService: playerService}
}

// GetAllByTeamID handles GET /api/v1/teams/:id/players
// Returns a paginated list of players belonging to the specified team.
//
//	@Summary		List players by team
//	@Description	Returns a paginated list of players belonging to the specified team
//	@Tags			Players
//	@Produce		json
//	@Security		BearerAuth
//	@Param			id			path		string	true	"Team UUID"
//	@Param			page		query		int		false	"Page number"		default(1)
//	@Param			per_page	query		int		false	"Items per page"	default(10)
//	@Param			sort_by		query		string	false	"Sort field"		default(created_at)
//	@Param			sort_order	query		string	false	"Sort order"		Enums(asc, desc)	default(desc)
//	@Success		200			{object}	response.Envelope{data=[]dto.PlayerResponse,meta=response.PaginationMeta}
//	@Failure		400			{object}	response.Envelope
//	@Failure		401			{object}	response.Envelope
//	@Failure		404			{object}	response.Envelope
//	@Failure		500			{object}	response.Envelope
//	@Router			/teams/{id}/players [get]
func (h *PlayerHandler) GetAllByTeamID(c *gin.Context) {
	teamID, ok := parseUUID(c, c.Param("id"))
	if !ok {
		return
	}

	pagination := bindPagination(c)

	players, meta, err := h.playerService.GetAllByTeamID(teamID, pagination)
	if err != nil {
		handleServiceError(c, err)
		return
	}

	response.SuccessWithPagination(c, http.StatusOK, "Players retrieved successfully", players, meta)
}

// GetByID handles GET /api/v1/players/:id
// Returns details of a single player.
//
//	@Summary		Get player by ID
//	@Description	Returns details of a single player by its UUID
//	@Tags			Players
//	@Produce		json
//	@Security		BearerAuth
//	@Param			id	path		string	true	"Player UUID"
//	@Success		200	{object}	response.Envelope{data=dto.PlayerResponse}
//	@Failure		400	{object}	response.Envelope
//	@Failure		401	{object}	response.Envelope
//	@Failure		404	{object}	response.Envelope
//	@Failure		500	{object}	response.Envelope
//	@Router			/players/{id} [get]
func (h *PlayerHandler) GetByID(c *gin.Context) {
	id, ok := parseUUID(c, c.Param("id"))
	if !ok {
		return
	}

	player, err := h.playerService.GetByID(id)
	if err != nil {
		handleServiceError(c, err)
		return
	}

	response.Success(c, http.StatusOK, "Player retrieved successfully", player)
}

// Create handles POST /api/v1/teams/:id/players
// Creates a new player under the specified team.
//
//	@Summary		Create a new player
//	@Description	Creates a new player under the specified team. Jersey number must be unique within the team.
//	@Tags			Players
//	@Accept			json
//	@Produce		json
//	@Security		BearerAuth
//	@Param			id		path		string					true	"Team UUID"
//	@Param			request	body		dto.CreatePlayerRequest	true	"Player data"
//	@Success		201		{object}	response.Envelope{data=dto.PlayerResponse}
//	@Failure		400		{object}	response.Envelope
//	@Failure		401		{object}	response.Envelope
//	@Failure		404		{object}	response.Envelope
//	@Failure		409		{object}	response.Envelope
//	@Failure		500		{object}	response.Envelope
//	@Router			/teams/{id}/players [post]
func (h *PlayerHandler) Create(c *gin.Context) {
	teamID, ok := parseUUID(c, c.Param("id"))
	if !ok {
		return
	}

	var req dto.CreatePlayerRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Error(c, errs.ErrBadRequest("Invalid request body: "+err.Error()))
		return
	}

	player, err := h.playerService.Create(teamID, req)
	if err != nil {
		handleServiceError(c, err)
		return
	}

	response.Success(c, http.StatusCreated, "Player created successfully", player)
}

// Update handles PUT /api/v1/players/:id
// Updates an existing player.
//
//	@Summary		Update a player
//	@Description	Updates an existing player by its UUID. Jersey number must remain unique within the team.
//	@Tags			Players
//	@Accept			json
//	@Produce		json
//	@Security		BearerAuth
//	@Param			id		path		string					true	"Player UUID"
//	@Param			request	body		dto.UpdatePlayerRequest	true	"Updated player data"
//	@Success		200		{object}	response.Envelope{data=dto.PlayerResponse}
//	@Failure		400		{object}	response.Envelope
//	@Failure		401		{object}	response.Envelope
//	@Failure		404		{object}	response.Envelope
//	@Failure		409		{object}	response.Envelope
//	@Failure		500		{object}	response.Envelope
//	@Router			/players/{id} [put]
func (h *PlayerHandler) Update(c *gin.Context) {
	id, ok := parseUUID(c, c.Param("id"))
	if !ok {
		return
	}

	var req dto.UpdatePlayerRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Error(c, errs.ErrBadRequest("Invalid request body: "+err.Error()))
		return
	}

	player, err := h.playerService.Update(id, req)
	if err != nil {
		handleServiceError(c, err)
		return
	}

	response.Success(c, http.StatusOK, "Player updated successfully", player)
}

// Delete handles DELETE /api/v1/players/:id
// Soft-deletes a player.
//
//	@Summary		Delete a player
//	@Description	Soft-deletes a player by its UUID
//	@Tags			Players
//	@Produce		json
//	@Security		BearerAuth
//	@Param			id	path		string	true	"Player UUID"
//	@Success		200	{object}	response.Envelope
//	@Failure		400	{object}	response.Envelope
//	@Failure		401	{object}	response.Envelope
//	@Failure		404	{object}	response.Envelope
//	@Failure		500	{object}	response.Envelope
//	@Router			/players/{id} [delete]
func (h *PlayerHandler) Delete(c *gin.Context) {
	id, ok := parseUUID(c, c.Param("id"))
	if !ok {
		return
	}

	if err := h.playerService.Delete(id); err != nil {
		handleServiceError(c, err)
		return
	}

	response.Success(c, http.StatusOK, "Player deleted successfully", nil)
}
