package handler

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/mhakimsaputra17/xyz-football-api/internal/dto"
	"github.com/mhakimsaputra17/xyz-football-api/internal/service"
	"github.com/mhakimsaputra17/xyz-football-api/pkg/response"
)

// TeamHandler handles team-related HTTP requests.
type TeamHandler struct {
	teamService service.TeamService
}

// NewTeamHandler creates a new TeamHandler instance.
func NewTeamHandler(teamService service.TeamService) *TeamHandler {
	return &TeamHandler{teamService: teamService}
}

// GetAll handles GET /api/v1/teams
// Returns a paginated list of all teams.
//
//	@Summary		List all teams
//	@Description	Returns a paginated list of all teams with sorting support
//	@Tags			Teams
//	@Produce		json
//	@Security		BearerAuth
//	@Param			page		query		int		false	"Page number"		default(1)
//	@Param			per_page	query		int		false	"Items per page"	default(10)
//	@Param			sort_by		query		string	false	"Sort field"		default(created_at)
//	@Param			sort_order	query		string	false	"Sort order"		Enums(asc, desc)	default(desc)
//	@Success		200			{object}	response.Envelope{data=[]dto.TeamResponse,meta=response.PaginationMeta}
//	@Failure		401			{object}	response.Envelope
//	@Failure		500			{object}	response.Envelope
//	@Router			/teams [get]
func (h *TeamHandler) GetAll(c *gin.Context) {
	pagination := bindPagination(c)

	teams, meta, err := h.teamService.GetAll(pagination)
	if err != nil {
		handleServiceError(c, err)
		return
	}

	response.SuccessWithPagination(c, http.StatusOK, "Teams retrieved successfully", teams, meta)
}

// GetByID handles GET /api/v1/teams/:id
// Returns details of a single team.
//
//	@Summary		Get team by ID
//	@Description	Returns details of a single team by its UUID
//	@Tags			Teams
//	@Produce		json
//	@Security		BearerAuth
//	@Param			id	path		string	true	"Team UUID"
//	@Success		200	{object}	response.Envelope{data=dto.TeamResponse}
//	@Failure		400	{object}	response.Envelope
//	@Failure		401	{object}	response.Envelope
//	@Failure		404	{object}	response.Envelope
//	@Failure		500	{object}	response.Envelope
//	@Router			/teams/{id} [get]
func (h *TeamHandler) GetByID(c *gin.Context) {
	id, ok := parseUUID(c, c.Param("id"), "id")
	if !ok {
		return
	}

	team, err := h.teamService.GetByID(id)
	if err != nil {
		handleServiceError(c, err)
		return
	}

	response.Success(c, http.StatusOK, "Team retrieved successfully", team)
}

// Create handles POST /api/v1/teams
// Creates a new team.
//
//	@Summary		Create a new team
//	@Description	Creates a new football team
//	@Tags			Teams
//	@Accept			json
//	@Produce		json
//	@Security		BearerAuth
//	@Param			request	body		dto.CreateTeamRequest	true	"Team data"
//	@Success		201		{object}	response.Envelope{data=dto.TeamResponse}
//	@Failure		400		{object}	response.Envelope
//	@Failure		401		{object}	response.Envelope
//	@Failure		500		{object}	response.Envelope
//	@Router			/teams [post]
func (h *TeamHandler) Create(c *gin.Context) {
	var req dto.CreateTeamRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		handleBindingError(c, err)
		return
	}

	team, err := h.teamService.Create(req)
	if err != nil {
		handleServiceError(c, err)
		return
	}

	response.Success(c, http.StatusCreated, "Team created successfully", team)
}

// Update handles PUT /api/v1/teams/:id
// Updates an existing team.
//
//	@Summary		Update a team
//	@Description	Updates an existing team by its UUID
//	@Tags			Teams
//	@Accept			json
//	@Produce		json
//	@Security		BearerAuth
//	@Param			id		path		string					true	"Team UUID"
//	@Param			request	body		dto.UpdateTeamRequest	true	"Updated team data"
//	@Success		200		{object}	response.Envelope{data=dto.TeamResponse}
//	@Failure		400		{object}	response.Envelope
//	@Failure		401		{object}	response.Envelope
//	@Failure		404		{object}	response.Envelope
//	@Failure		500		{object}	response.Envelope
//	@Router			/teams/{id} [put]
func (h *TeamHandler) Update(c *gin.Context) {
	id, ok := parseUUID(c, c.Param("id"), "id")
	if !ok {
		return
	}

	var req dto.UpdateTeamRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		handleBindingError(c, err)
		return
	}

	team, err := h.teamService.Update(id, req)
	if err != nil {
		handleServiceError(c, err)
		return
	}

	response.Success(c, http.StatusOK, "Team updated successfully", team)
}

// Delete handles DELETE /api/v1/teams/:id
// Soft-deletes a team.
//
//	@Summary		Delete a team
//	@Description	Soft-deletes a team by its UUID
//	@Tags			Teams
//	@Produce		json
//	@Security		BearerAuth
//	@Param			id	path		string	true	"Team UUID"
//	@Success		200	{object}	response.Envelope
//	@Failure		400	{object}	response.Envelope
//	@Failure		401	{object}	response.Envelope
//	@Failure		404	{object}	response.Envelope
//	@Failure		500	{object}	response.Envelope
//	@Router			/teams/{id} [delete]
func (h *TeamHandler) Delete(c *gin.Context) {
	id, ok := parseUUID(c, c.Param("id"), "id")
	if !ok {
		return
	}

	if err := h.teamService.Delete(id); err != nil {
		handleServiceError(c, err)
		return
	}

	response.Success(c, http.StatusOK, "Team deleted successfully", nil)
}
