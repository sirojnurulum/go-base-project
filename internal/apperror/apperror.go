package apperror

import (
	"fmt"
	"net/http"
)

// AppError adalah tipe error kustom untuk aplikasi kita.
// Ini mengimplementasikan interface `error`.
type AppError struct {
	Code    int    `json:"code"`
	Message string `json:"message"`
	Err     error  `json:"-"` // Error asli, tidak diekspos ke JSON
}

func (e *AppError) Error() string {
	if e.Err != nil {
		return fmt.Sprintf("%s: %v", e.Message, e.Err)
	}
	return e.Message
}

// NewAppError membuat instance AppError baru.
func NewAppError(code int, message string, err error) *AppError {
	return &AppError{Code: code, Message: message, Err: err}
}

// NewNotFoundError adalah helper untuk error 404.
func NewNotFoundError(resource string) *AppError {
	return NewAppError(http.StatusNotFound, fmt.Sprintf("%s not found", resource), nil)
}

// NewConflictError adalah helper untuk error 409.
func NewConflictError(message string) *AppError {
	return NewAppError(http.StatusConflict, message, nil)
}

// NewUnauthorizedError adalah helper untuk error 401.
func NewUnauthorizedError(message string) *AppError {
	return NewAppError(http.StatusUnauthorized, message, nil)
}

// NewInternalError adalah helper untuk error 500.
func NewInternalError(err error) *AppError {
	return NewAppError(http.StatusInternalServerError, "an unexpected error occurred", err)
}
