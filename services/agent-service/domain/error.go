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
		Message: "invalid id",
		Code:    400, // StatusBadRequest
	}
	ErrEmailRequired = &AppError{
		Message: "email is required",
		Code:    400, // StatusBadRequest
	}
	ErrAgentNameRequired = &AppError{
		Message: "agent name is required",
		Code:    400, // StatusBadRequest
	}
	ErrAgentTypeRequired = &AppError{
		Message: "agent type is required",
		Code:    400, // StatusBadRequest
	}
	ErrInvalidAgentType = &AppError{
		Message: "invalid agent type. Must be IATA or SUB_AGENT",
		Code:    400, // StatusBadRequest
	}
	ErrAgentNotFound = &AppError{
		Message: "agent not found",
		Code:    404, // StatusNotFound
	}
	ErrParentAgentNotFound = &AppError{
		Message: "parent agent not found",
		Code:    404, // StatusNotFound
	}
	ErrCircularReference = &AppError{
		Message: "circular reference detected in agent hierarchy",
		Code:    400, // StatusBadRequest
	}
	ErrAgentHasChildren = &AppError{
		Message: "cannot delete agent with children",
		Code:    400, // StatusBadRequest
	}
)

// Standard error types for repositories
var (
	ErrNotFound = errors.New("not found")
)
