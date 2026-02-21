package errs

import "net/http"

// AppError represents an application-level error with an HTTP status code.
// This follows the error-handling-patterns skill: errors carry context about
// how they should be presented to the client.
type AppError struct {
	Code    int          `json:"-"`
	Message string       `json:"message"`
	Errors  []FieldError `json:"errors,omitempty"`
}

// FieldError represents a validation error on a specific field.
type FieldError struct {
	Field   string `json:"field"`
	Message string `json:"message"`
}

func (e *AppError) Error() string {
	return e.Message
}

// New creates a new AppError with the given HTTP status code and message.
func New(code int, message string) *AppError {
	return &AppError{
		Code:    code,
		Message: message,
	}
}

// WithFields adds field-level validation errors to the AppError.
func (e *AppError) WithFields(fields []FieldError) *AppError {
	e.Errors = fields
	return e
}

// --- Predefined error constructors ---

// ErrBadRequest returns a 400 error.
func ErrBadRequest(message string) *AppError {
	return New(http.StatusBadRequest, message)
}

// ErrUnauthorized returns a 401 error.
func ErrUnauthorized(message string) *AppError {
	return New(http.StatusUnauthorized, message)
}

// ErrForbidden returns a 403 error.
func ErrForbidden(message string) *AppError {
	return New(http.StatusForbidden, message)
}

// ErrNotFound returns a 404 error.
func ErrNotFound(message string) *AppError {
	return New(http.StatusNotFound, message)
}

// ErrConflict returns a 409 error.
func ErrConflict(message string) *AppError {
	return New(http.StatusConflict, message)
}

// ErrInternal returns a 500 error.
// The actual error detail should be logged server-side; only a generic message goes to the client.
func ErrInternal(message string) *AppError {
	return New(http.StatusInternalServerError, message)
}

// ErrValidation returns a 400 error with field-level details.
func ErrValidation(fields []FieldError) *AppError {
	return New(http.StatusBadRequest, "Validation failed").WithFields(fields)
}
