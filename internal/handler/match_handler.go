package handler

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/mhakimsaputra17/xyz-football-api/internal/dto"
	"github.com/mhakimsaputra17/xyz-football-api/internal/service"
	"github.com/mhakimsaputra17/xyz-football-api/pkg/errs"
	"github.com/mhakimsaputra17/xyz-football-api/pkg/response"
)

// MatchHandler handles match-related HTTP requests.
type MatchHandler struct {
	matchService service.MatchService
}

// NewMatchHandler creates a new MatchHandler instance.
func NewMatchHandler(matchService service.MatchService) *MatchHandler {
	return &MatchHandler{matchService: matchService}
}

// GetAll handles GET /api/v1/matches
// Returns a paginated list of all matches.
//
//	@Summary		List all matches
//	@Description	Returns a paginated list of all matches with home/away team details
//	@Tags			Matches
//	@Produce		json
//	@Security		BearerAuth
//	@Param			page		query		int		false	"Page number"		default(1)
//	@Param			per_page	query		int		false	"Items per page"	default(10)
//	@Param			sort_by		query		string	false	"Sort field"		default(created_at)
//	@Param			sort_order	query		string	false	"Sort order"		Enums(asc, desc)	default(desc)
//	@Success		200			{object}	response.Envelope{data=[]dto.MatchResponse,meta=response.PaginationMeta}
//	@Failure		401			{object}	response.Envelope
//	@Failure		500			{object}	response.Envelope
//	@Router			/matches [get]
func (h *MatchHandler) GetAll(c *gin.Context) {
	pagination := bindPagination(c)

	matches, meta, err := h.matchService.GetAll(pagination)
	if err != nil {
		handleServiceError(c, err)
		return
	}

	response.SuccessWithPagination(c, http.StatusOK, "Matches retrieved successfully", matches, meta)
}

// GetByID handles GET /api/v1/matches/:id
// Returns details of a single match including goals.
//
//	@Summary		Get match by ID
//	@Description	Returns details of a single match including goals, home team, and away team
//	@Tags			Matches
//	@Produce		json
//	@Security		BearerAuth
//	@Param			id	path		string	true	"Match UUID"
//	@Success		200	{object}	response.Envelope{data=dto.MatchResponse}
//	@Failure		400	{object}	response.Envelope
//	@Failure		401	{object}	response.Envelope
//	@Failure		404	{object}	response.Envelope
//	@Failure		500	{object}	response.Envelope
//	@Router			/matches/{id} [get]
func (h *MatchHandler) GetByID(c *gin.Context) {
	id, ok := parseUUID(c, c.Param("id"))
	if !ok {
		return
	}

	match, err := h.matchService.GetByID(id)
	if err != nil {
		handleServiceError(c, err)
		return
	}

	response.Success(c, http.StatusOK, "Match retrieved successfully", match)
}

// Create handles POST /api/v1/matches
// Creates a new match schedule.
//
//	@Summary		Create a new match
//	@Description	Creates a new match schedule between two different teams
//	@Tags			Matches
//	@Accept			json
//	@Produce		json
//	@Security		BearerAuth
//	@Param			request	body		dto.CreateMatchRequest	true	"Match data"
//	@Success		201		{object}	response.Envelope{data=dto.MatchResponse}
//	@Failure		400		{object}	response.Envelope
//	@Failure		401		{object}	response.Envelope
//	@Failure		404		{object}	response.Envelope
//	@Failure		500		{object}	response.Envelope
//	@Router			/matches [post]
func (h *MatchHandler) Create(c *gin.Context) {
	var req dto.CreateMatchRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Error(c, errs.ErrBadRequest("Invalid request body: "+err.Error()))
		return
	}

	match, err := h.matchService.Create(req)
	if err != nil {
		handleServiceError(c, err)
		return
	}

	response.Success(c, http.StatusCreated, "Match created successfully", match)
}

// Update handles PUT /api/v1/matches/:id
// Updates an existing match schedule.
//
//	@Summary		Update a match
//	@Description	Updates an existing match schedule. Cannot update a completed match.
//	@Tags			Matches
//	@Accept			json
//	@Produce		json
//	@Security		BearerAuth
//	@Param			id		path		string					true	"Match UUID"
//	@Param			request	body		dto.UpdateMatchRequest	true	"Updated match data"
//	@Success		200		{object}	response.Envelope{data=dto.MatchResponse}
//	@Failure		400		{object}	response.Envelope
//	@Failure		401		{object}	response.Envelope
//	@Failure		404		{object}	response.Envelope
//	@Failure		500		{object}	response.Envelope
//	@Router			/matches/{id} [put]
func (h *MatchHandler) Update(c *gin.Context) {
	id, ok := parseUUID(c, c.Param("id"))
	if !ok {
		return
	}

	var req dto.UpdateMatchRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Error(c, errs.ErrBadRequest("Invalid request body: "+err.Error()))
		return
	}

	match, err := h.matchService.Update(id, req)
	if err != nil {
		handleServiceError(c, err)
		return
	}

	response.Success(c, http.StatusOK, "Match updated successfully", match)
}

// Delete handles DELETE /api/v1/matches/:id
// Soft-deletes a match.
//
//	@Summary		Delete a match
//	@Description	Soft-deletes a match by its UUID
//	@Tags			Matches
//	@Produce		json
//	@Security		BearerAuth
//	@Param			id	path		string	true	"Match UUID"
//	@Success		200	{object}	response.Envelope
//	@Failure		400	{object}	response.Envelope
//	@Failure		401	{object}	response.Envelope
//	@Failure		404	{object}	response.Envelope
//	@Failure		500	{object}	response.Envelope
//	@Router			/matches/{id} [delete]
func (h *MatchHandler) Delete(c *gin.Context) {
	id, ok := parseUUID(c, c.Param("id"))
	if !ok {
		return
	}

	if err := h.matchService.Delete(id); err != nil {
		handleServiceError(c, err)
		return
	}

	response.Success(c, http.StatusOK, "Match deleted successfully", nil)
}

// SubmitResult handles POST /api/v1/matches/:id/result
// Submits match results (goals), auto-computes scores, transitions status to completed.
//
//	@Summary		Submit match result
//	@Description	Submits goals for a scheduled match, auto-computes scores, and marks the match as completed
//	@Tags			Matches
//	@Accept			json
//	@Produce		json
//	@Security		BearerAuth
//	@Param			id		path		string					true	"Match UUID"
//	@Param			request	body		dto.MatchResultRequest	true	"Match result with goals"
//	@Success		200		{object}	response.Envelope{data=dto.MatchResponse}
//	@Failure		400		{object}	response.Envelope
//	@Failure		401		{object}	response.Envelope
//	@Failure		404		{object}	response.Envelope
//	@Failure		500		{object}	response.Envelope
//	@Router			/matches/{id}/result [post]
func (h *MatchHandler) SubmitResult(c *gin.Context) {
	id, ok := parseUUID(c, c.Param("id"))
	if !ok {
		return
	}

	var req dto.MatchResultRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Error(c, errs.ErrBadRequest("Invalid request body: "+err.Error()))
		return
	}

	match, err := h.matchService.SubmitResult(id, req)
	if err != nil {
		handleServiceError(c, err)
		return
	}

	response.Success(c, http.StatusOK, "Match result submitted successfully", match)
}

// UpdateResult handles PUT /api/v1/matches/:id/result
// Replaces existing match results with new data.
//
//	@Summary		Update match result
//	@Description	Replaces existing goals for a completed match with new result data and recomputes scores
//	@Tags			Matches
//	@Accept			json
//	@Produce		json
//	@Security		BearerAuth
//	@Param			id		path		string					true	"Match UUID"
//	@Param			request	body		dto.MatchResultRequest	true	"Updated match result with goals"
//	@Success		200		{object}	response.Envelope{data=dto.MatchResponse}
//	@Failure		400		{object}	response.Envelope
//	@Failure		401		{object}	response.Envelope
//	@Failure		404		{object}	response.Envelope
//	@Failure		500		{object}	response.Envelope
//	@Router			/matches/{id}/result [put]
func (h *MatchHandler) UpdateResult(c *gin.Context) {
	id, ok := parseUUID(c, c.Param("id"))
	if !ok {
		return
	}

	var req dto.MatchResultRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Error(c, errs.ErrBadRequest("Invalid request body: "+err.Error()))
		return
	}

	match, err := h.matchService.UpdateResult(id, req)
	if err != nil {
		handleServiceError(c, err)
		return
	}

	response.Success(c, http.StatusOK, "Match result updated successfully", match)
}
