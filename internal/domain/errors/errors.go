package errors

import "errors"

// Domain errors
var (
	// Server errors
	ErrServerNotFound      = errors.New("server not found")
	ErrServerAlreadyExists = errors.New("server already exists")
	ErrInvalidServerID     = errors.New("invalid server ID")
	ErrInvalidServerName   = errors.New("invalid server name")
	ErrInvalidIPv4         = errors.New("invalid IPv4 address")

	// User errors
	ErrUserNotFound            = errors.New("user not found")
	ErrUserAlreadyExists       = errors.New("user already exists")
	ErrEmailAlreadyExists      = errors.New("email already exists")
	ErrUserCreationFailed      = errors.New("user creation failed")
	ErrUserUpdateFailed        = errors.New("user update failed")
	ErrPasswordHashingFailed   = errors.New("password hashing failed")
	ErrPasswordMismatch        = errors.New("password mismatch")
	ErrInvalidUsername         = errors.New("invalid username")
	ErrInvalidEmail            = errors.New("invalid email")
	ErrInvalidPassword         = errors.New("invalid password")
	ErrInvalidCredentials      = errors.New("invalid credentials")
	ErrInsufficientPermissions = errors.New("insufficient permissions")

	// General errors
	ErrInvalidInput        = errors.New("invalid input")
	ErrNotFound            = errors.New("resource not found")
	ErrUnauthorized        = errors.New("unauthorized")
	ErrForbidden           = errors.New("forbidden")
	ErrInternalServerError = errors.New("internal server error")
)
