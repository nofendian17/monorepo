package domain

import "errors"

// Error types with HTTP status codes
type AppError struct {
	Message string
	Code    int
}

func (e *AppError) Error() string {
	return e.Message
}

// Custom error types
var (
	ErrEmailAlreadyExists = &AppError{
		Message: "user with this email already exists",
		Code:    409, // StatusConflict
	}
	ErrUserNotFound = &AppError{
		Message: "user not found",
		Code:    404, // StatusNotFound
	}
	ErrInvalidID = &AppError{
		Message: "invalid user id",
		Code:    400, // StatusBadRequest
	}
	ErrEmailRequired = &AppError{
		Message: "email is required",
		Code:    400, // StatusBadRequest
	}
)

// Standard error types for repositories
var (
	ErrNotFound = errors.New("not found")
)
