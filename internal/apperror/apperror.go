package apperror

import (
	"fmt"
	"net/http"
)

// AppError adalah tipe error kustom untuk aplikasi kita.
// Ini mengimplementasikan interface `error`.
type AppError struct {
	Code      int    `json:"code"`
	Message   string `json:"message"`
	ErrorCode string `json:"error_code,omitempty"` // Machine-readable error code
	Err       error  `json:"-"`                    // Error asli, tidak diekspos ke JSON
}

func (e *AppError) Error() string {
	if e.Err != nil {
		return fmt.Sprintf("[%s] %s: %v", e.ErrorCode, e.Message, e.Err)
	}
	return fmt.Sprintf("[%s] %s", e.ErrorCode, e.Message)
}

// NewAppError membuat instance AppError baru.
func NewAppError(code int, message string, err error) *AppError {
	return &AppError{Code: code, Message: message, Err: err}
}

// NewAppErrorWithCode membuat instance AppError dengan error code.
func NewAppErrorWithCode(code int, message, errorCode string, err error) *AppError {
	return &AppError{Code: code, Message: message, ErrorCode: errorCode, Err: err}
}

// NewNotFoundError adalah helper untuk error 404.
func NewNotFoundError(resource string) *AppError {
	return NewAppErrorWithCode(http.StatusNotFound, fmt.Sprintf("%s not found", resource), "RESOURCE_NOT_FOUND", nil)
}

// NewConflictError adalah helper untuk error 409.
func NewConflictError(message string) *AppError {
	return NewAppErrorWithCode(http.StatusConflict, message, "RESOURCE_CONFLICT", nil)
}

// NewUnauthorizedError adalah helper untuk error 401.
func NewUnauthorizedError(message string) *AppError {
	return NewAppErrorWithCode(http.StatusUnauthorized, message, "UNAUTHORIZED", nil)
}

// NewForbiddenError adalah helper untuk error 403.
func NewForbiddenError(message string) *AppError {
	return NewAppErrorWithCode(http.StatusForbidden, message, "FORBIDDEN", nil)
}

// NewValidationError adalah helper untuk error 400 validation.
func NewValidationError(message string) *AppError {
	return NewAppErrorWithCode(http.StatusBadRequest, message, "VALIDATION_ERROR", nil)
}

// NewInternalError adalah helper untuk error 500.
func NewInternalError(err error) *AppError {
	return NewAppErrorWithCode(http.StatusInternalServerError, "an unexpected error occurred", "INTERNAL_ERROR", err)
}
