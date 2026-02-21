package response

import (
	"github.com/gin-gonic/gin"
	"github.com/mhakimsaputra17/xyz-football-api/pkg/errs"
)

// Envelope is the standard API response format: { status, message, data, meta?, errors? }
type Envelope struct {
	Status  string            `json:"status"`
	Message string            `json:"message"`
	Data    any               `json:"data,omitempty"`
	Meta    *PaginationMeta   `json:"meta,omitempty"`
	Errors  []errs.FieldError `json:"errors,omitempty"`
}

// PaginationMeta holds pagination metadata for list responses.
type PaginationMeta struct {
	Page       int   `json:"page"`
	PerPage    int   `json:"per_page"`
	Total      int64 `json:"total"`
	TotalPages int   `json:"total_pages"`
}

// Success sends a success response with optional data.
func Success(c *gin.Context, code int, message string, data any) {
	c.JSON(code, Envelope{
		Status:  "success",
		Message: message,
		Data:    data,
	})
}

// SuccessWithPagination sends a success response with data and pagination metadata.
func SuccessWithPagination(c *gin.Context, code int, message string, data any, meta *PaginationMeta) {
	c.JSON(code, Envelope{
		Status:  "success",
		Message: message,
		Data:    data,
		Meta:    meta,
	})
}

// Error sends an error response derived from an AppError.
// Detail is logged server-side; only the structured error goes to the client.
func Error(c *gin.Context, err *errs.AppError) {
	c.JSON(err.Code, Envelope{
		Status:  "error",
		Message: err.Message,
		Errors:  err.Errors,
	})
}

// Abort sends an error response and aborts the middleware chain.
func Abort(c *gin.Context, err *errs.AppError) {
	c.AbortWithStatusJSON(err.Code, Envelope{
		Status:  "error",
		Message: err.Message,
		Errors:  err.Errors,
	})
}
