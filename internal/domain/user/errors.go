package user

import "errors"

// Sentinel errors for user domain operations.
var (
	// ErrNotFound is returned when a user cannot be found.
	ErrNotFound = errors.New("user not found")
	// ErrInvalidCredentials is returned when login credentials are invalid.
	ErrInvalidCredentials = errors.New("invalid credentials")
	// ErrEmailAlreadyInUse is returned when attempting to register with an existing email.
	ErrEmailAlreadyInUse = errors.New("email already in use")
)
