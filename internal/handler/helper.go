package handler

import (
	"errors"
	"fmt"
	"strings"
	"unicode"

	"github.com/gin-gonic/gin"
	"github.com/go-playground/validator/v10"
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

// handleBindingError converts GIN binding/validation errors into structured
// field-level error responses using errs.ErrValidation.
// For validator.ValidationErrors it maps each field to a human-readable message.
// For other errors (e.g., malformed JSON) it returns a generic 400.
func handleBindingError(c *gin.Context, err error) {
	var ve validator.ValidationErrors
	if !errors.As(err, &ve) {
		// Not a validation error — likely malformed JSON
		response.Error(c, errs.ErrBadRequest("Invalid request body"))
		return
	}

	fields := make([]errs.FieldError, len(ve))
	for i, fe := range ve {
		fields[i] = errs.FieldError{
			Field:   fieldName(fe),
			Message: validationMessage(fe),
		}
	}

	response.Error(c, errs.ErrValidation(fields))
}

// parseUUID parses a string into a UUID and sends a 400 error if invalid.
// paramName is included in the error message for context (e.g., "id").
// Returns the parsed UUID and true if successful, zero UUID and false if not.
func parseUUID(c *gin.Context, raw string, paramName string) (uuid.UUID, bool) {
	id, err := uuid.Parse(raw)
	if err != nil {
		msg := fmt.Sprintf("Invalid UUID format for '%s' parameter", paramName)
		response.Error(c, errs.ErrBadRequest(msg))
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

// fieldName extracts a JSON-style field path from a validator.FieldError.
// Converts PascalCase struct field names to snake_case and preserves array indices.
// Example: "Goals[0].PlayerID" → "goals[0].player_id"
func fieldName(fe validator.FieldError) string {
	ns := fe.Namespace()

	// Remove the struct type prefix (e.g., "CreateTeamRequest.Name" → "Name")
	if idx := strings.Index(ns, "."); idx >= 0 {
		ns = ns[idx+1:]
	}

	// Convert each dot-separated segment to snake_case
	var result strings.Builder
	parts := strings.Split(ns, ".")
	for i, part := range parts {
		if i > 0 {
			result.WriteByte('.')
		}
		// Preserve array indices: "Goals[0]" → "goals[0]"
		if bracketIdx := strings.Index(part, "["); bracketIdx >= 0 {
			result.WriteString(toSnakeCase(part[:bracketIdx]))
			result.WriteString(part[bracketIdx:])
		} else {
			result.WriteString(toSnakeCase(part))
		}
	}
	return result.String()
}

// validationMessage returns a human-readable message for a validator.FieldError.
func validationMessage(fe validator.FieldError) string {
	field := fieldName(fe)

	switch fe.Tag() {
	case "required":
		return field + " is required"
	case "gt":
		return field + " must be greater than " + fe.Param()
	case "gte":
		return field + " must be at least " + fe.Param()
	case "min":
		return field + " must be at least " + fe.Param()
	case "max":
		return field + " must be at most " + fe.Param()
	case "url":
		return field + " must be a valid URL"
	case "uuid":
		return field + " must be a valid UUID"
	case "oneof":
		return field + " must be one of: " + strings.ReplaceAll(fe.Param(), " ", ", ")
	default:
		return field + " is invalid"
	}
}

// toSnakeCase converts a PascalCase string to snake_case.
// Handles consecutive uppercase sequences (e.g., "LogoURL" → "logo_url").
func toSnakeCase(s string) string {
	runes := []rune(s)
	var result []rune

	for i, r := range runes {
		if unicode.IsUpper(r) {
			if i > 0 {
				prev := runes[i-1]
				if !unicode.IsUpper(prev) {
					// Transition from lower to upper: add underscore
					result = append(result, '_')
				} else if i+1 < len(runes) && !unicode.IsUpper(runes[i+1]) {
					// Middle of uppercase sequence followed by lowercase: add underscore
					// e.g., "URL" in "LogoURL" → the final 'L' before end stays grouped
					result = append(result, '_')
				}
			}
			result = append(result, unicode.ToLower(r))
		} else {
			result = append(result, r)
		}
	}

	return string(result)
}
