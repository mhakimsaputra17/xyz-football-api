package handler

import (
	"errors"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/mhakimsaputra17/xyz-football-api/internal/dto"
	"github.com/mhakimsaputra17/xyz-football-api/pkg/errs"
	"github.com/mhakimsaputra17/xyz-football-api/pkg/response"
)

// handleServiceError converts service-layer errors (*AppError) into HTTP responses.
func handleServiceError(c *gin.Context, err error) {
	var appErr *errs.AppError
	if errors.As(err, &appErr) {
		response.Error(c, appErr)
		return
	}
	// Fallback for unexpected errors — generic 500
	response.Error(c, errs.ErrInternal("Internal server error"))
}

// parseUUID parses a string into a UUID and sends a 400 error if invalid.
// Returns the parsed UUID and true if successful, zero UUID and false if not.
func parseUUID(c *gin.Context, param string) (uuid.UUID, bool) {
	id, err := uuid.Parse(param)
	if err != nil {
		response.Error(c, errs.ErrBadRequest("Invalid UUID format for parameter"))
		return uuid.Nil, false
	}
	return id, true
}

// bindPagination parses pagination query parameters from the request.
func bindPagination(c *gin.Context) dto.PaginationQuery {
	var pagination dto.PaginationQuery
	// ShouldBindQuery does not abort on error; defaults are set via Sanitize().
	_ = c.ShouldBindQuery(&pagination)
	pagination.Sanitize()
	return pagination
}
