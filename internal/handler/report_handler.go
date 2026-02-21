package handler

import (
	"net/http"

	"github.com/gin-gonic/gin"
	_ "github.com/mhakimsaputra17/xyz-football-api/internal/dto"
	"github.com/mhakimsaputra17/xyz-football-api/internal/service"
	"github.com/mhakimsaputra17/xyz-football-api/pkg/response"
)

// ReportHandler handles report-related HTTP requests.
type ReportHandler struct {
	reportService service.ReportService
}

// NewReportHandler creates a new ReportHandler instance.
func NewReportHandler(reportService service.ReportService) *ReportHandler {
	return &ReportHandler{reportService: reportService}
}

// GetMatchReports handles GET /api/v1/reports/matches
// Returns a paginated list of all completed match reports.
//
//	@Summary		List match reports
//	@Description	Returns a paginated list of completed match reports with results summary
//	@Tags			Reports
//	@Produce		json
//	@Security		BearerAuth
//	@Param			page		query		int		false	"Page number"		default(1)
//	@Param			per_page	query		int		false	"Items per page"	default(10)
//	@Param			sort_by		query		string	false	"Sort field"		default(created_at)
//	@Param			sort_order	query		string	false	"Sort order"		Enums(asc, desc)	default(desc)
//	@Success		200			{object}	response.Envelope{data=[]dto.MatchReportListItem,meta=response.PaginationMeta}
//	@Failure		401			{object}	response.Envelope
//	@Failure		500			{object}	response.Envelope
//	@Router			/reports/matches [get]
func (h *ReportHandler) GetMatchReports(c *gin.Context) {
	pagination := bindPagination(c)

	reports, meta, err := h.reportService.GetMatchReports(pagination)
	if err != nil {
		handleServiceError(c, err)
		return
	}

	response.SuccessWithPagination(c, http.StatusOK, "Match reports retrieved successfully", reports, meta)
}

// GetMatchReportByID handles GET /api/v1/reports/matches/:id
// Returns a detailed report for a single completed match.
//
//	@Summary		Get match report by ID
//	@Description	Returns a detailed report for a completed match including goals, top scorer, match result, and accumulated total wins
//	@Tags			Reports
//	@Produce		json
//	@Security		BearerAuth
//	@Param			id	path		string	true	"Match UUID"
//	@Success		200	{object}	response.Envelope{data=dto.MatchReportResponse}
//	@Failure		400	{object}	response.Envelope
//	@Failure		401	{object}	response.Envelope
//	@Failure		404	{object}	response.Envelope
//	@Failure		500	{object}	response.Envelope
//	@Router			/reports/matches/{id} [get]
func (h *ReportHandler) GetMatchReportByID(c *gin.Context) {
	id, ok := parseUUID(c, c.Param("id"))
	if !ok {
		return
	}

	report, err := h.reportService.GetMatchReportByID(id)
	if err != nil {
		handleServiceError(c, err)
		return
	}

	response.Success(c, http.StatusOK, "Match report retrieved successfully", report)
}
