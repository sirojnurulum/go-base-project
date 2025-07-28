package constant

import "errors"

// Sentinel errors returned by services.
// These are part of the public contract of the service layer and are safe to check against.
var (
	// User-related errors
	ErrUserNotFound   = errors.New("user not found")
	ErrUsernameExists = errors.New("username already exists")
	ErrEmailExists    = errors.New("email already exists")

	// Auth-related errors
	ErrInvalidCredentials = errors.New("invalid credentials")
)
