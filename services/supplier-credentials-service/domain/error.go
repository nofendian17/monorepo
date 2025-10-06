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
	ErrSupplierNotFound = &AppError{
		Message: "supplier not found",
		Code:    404, // StatusNotFound
	}
	ErrCredentialNotFound = &AppError{
		Message: "credential not found",
		Code:    404, // StatusNotFound
	}
	ErrSupplierCodeRequired = &AppError{
		Message: "supplier code is required",
		Code:    400, // StatusBadRequest
	}
	ErrSupplierNameRequired = &AppError{
		Message: "supplier name is required",
		Code:    400, // StatusBadRequest
	}
	ErrSupplierCodeAlreadyExists = &AppError{
		Message: "supplier with this code already exists",
		Code:    409, // StatusConflict
	}
	ErrIataAgentIDRequired = &AppError{
		Message: "IATA agent ID is required",
		Code:    400, // StatusBadRequest
	}
	ErrSupplierIDRequired = &AppError{
		Message: "supplier ID is required",
		Code:    400, // StatusBadRequest
	}
	ErrCredentialsRequired = &AppError{
		Message: "credentials are required",
		Code:    400, // StatusBadRequest
	}
	ErrCredentialAlreadyExists = &AppError{
		Message: "credential for this agent-supplier pair already exists",
		Code:    409, // StatusConflict
	}
	ErrInvalidID = &AppError{
		Message: "invalid id",
		Code:    400, // StatusBadRequest
	}
)

// Standard error types for repositories
var (
	ErrNotFound = errors.New("not found")
)
