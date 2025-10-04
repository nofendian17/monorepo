package api

import (
	"context"
	"encoding/json"
	"net/http"

	"github.com/go-chi/chi/v5/middleware"
)

const (
	StatusSuccess = "success"
	StatusError   = "error"
)

// Response represents the standard API response format
type Response struct {
	RequestID string `json:"request_id"`
	Status    string `json:"status"`
	Data      any    `json:"data,omitempty"`
	Error     *Error `json:"error,omitempty"`
	Meta      *Meta  `json:"meta,omitempty"`
}

// Meta contains metadata for API responses
type Meta struct {
	Pagination *Pagination `json:"pagination,omitempty"`
}

// Pagination contains pagination information
type Pagination struct {
	Page        int  `json:"page"`
	Limit       int  `json:"limit"`
	Total       int  `json:"total"`
	TotalPages  int  `json:"total_pages"`
	HasNextPage bool `json:"has_next_page"`
	HasPrevPage bool `json:"has_prev_page"`
}

// Error represents the standard error format
type Error struct {
	Code    string        `json:"code"`
	Message string        `json:"message"`
	Details []ErrorDetail `json:"details,omitempty"`
}

// ErrorDetail contains detailed error information for specific fields
type ErrorDetail struct {
	Field   string `json:"field"`
	Message string `json:"message"`
}

// Api interface defines methods for standard API responses
type Api interface {
	Success(ctx context.Context, w http.ResponseWriter, data any)
	Created(ctx context.Context, w http.ResponseWriter, data any)
	Error(ctx context.Context, w http.ResponseWriter, statusCode int, apiErr *Error)
	SuccessWithMeta(ctx context.Context, w http.ResponseWriter, data any, meta *Meta)
	SuccessWithCode(ctx context.Context, w http.ResponseWriter, data any)
	SuccessWithCodeAndMeta(ctx context.Context, w http.ResponseWriter, data any, meta *Meta)
	BadRequest(ctx context.Context, w http.ResponseWriter, message string)
	Unauthorized(ctx context.Context, w http.ResponseWriter, message string)
	Forbidden(ctx context.Context, w http.ResponseWriter, message string)
	NotFound(ctx context.Context, w http.ResponseWriter, message string)
	InternalServerError(ctx context.Context, w http.ResponseWriter, message string)
	ValidationError(ctx context.Context, w http.ResponseWriter, details []ErrorDetail)
}

type api struct {
}

// New creates a new instance of the API response handler
func New() Api {
	return &api{}
}

// getRequestID safely extracts the request ID from context
func (a *api) getRequestID(ctx context.Context) string {
	return middleware.GetReqID(ctx)
}

// buildResponse creates a basic response structure
func (a *api) buildResponse(ctx context.Context, status string, data any, meta *Meta, apiErr *Error) Response {
	response := Response{
		RequestID: a.getRequestID(ctx),
		Status:    status,
	}

	if data != nil {
		response.Data = data
	}

	if meta != nil {
		response.Meta = meta
	}

	if apiErr != nil {
		response.Error = apiErr
	}

	return response
}

// writeJSONResponse writes a JSON response and handles encoding errors
func (a *api) writeJSONResponse(w http.ResponseWriter, response Response) error {
	return json.NewEncoder(w).Encode(response)
}

// Success sends a successful response with data
func (a *api) Success(ctx context.Context, w http.ResponseWriter, data any) {
	response := a.buildResponse(ctx, StatusSuccess, data, nil, nil)

	w.Header().Set("Content-Type", "application/json")
	if err := a.writeJSONResponse(w, response); err != nil {
		// Log error but don't expose it to client
		// In a real implementation, you'd want to use a proper logger here
		_ = err
	}
}

// Created sends a 201 Created response with data
func (a *api) Created(ctx context.Context, w http.ResponseWriter, data any) {
	response := a.buildResponse(ctx, StatusSuccess, data, nil, nil)

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	if err := a.writeJSONResponse(w, response); err != nil {
		// Log error but don't expose it to client
		_ = err
	}
}

// SuccessWithCode sends a successful response with data and business code
func (a *api) SuccessWithCode(ctx context.Context, w http.ResponseWriter, data any) {
	response := a.buildResponse(ctx, StatusSuccess, data, nil, nil)

	w.Header().Set("Content-Type", "application/json")
	if err := a.writeJSONResponse(w, response); err != nil {
		// Log error but don't expose it to client
		_ = err
	}
}

// SuccessWithMeta sends a successful response with data and metadata
func (a *api) SuccessWithMeta(ctx context.Context, w http.ResponseWriter, data any, meta *Meta) {
	response := a.buildResponse(ctx, StatusSuccess, data, meta, nil)

	w.Header().Set("Content-Type", "application/json")
	if err := a.writeJSONResponse(w, response); err != nil {
		// Log error but don't expose it to client
		_ = err
	}
}

// SuccessWithCodeAndMeta sends a successful response with data, business code, and metadata
func (a *api) SuccessWithCodeAndMeta(ctx context.Context, w http.ResponseWriter, data any, meta *Meta) {
	response := a.buildResponse(ctx, StatusSuccess, data, meta, nil)

	w.Header().Set("Content-Type", "application/json")
	if err := a.writeJSONResponse(w, response); err != nil {
		// Log error but don't expose it to client
		_ = err
	}
}

// Error sends an error response with specific HTTP status code and error details
func (a *api) Error(ctx context.Context, w http.ResponseWriter, statusCode int, apiErr *Error) {
	response := a.buildResponse(ctx, StatusError, nil, nil, apiErr)

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(statusCode)
	if err := a.writeJSONResponse(w, response); err != nil {
		// Log error but don't expose it to client
		_ = err
	}
}

// BadRequest sends a 400 Bad Request response
func (a *api) BadRequest(ctx context.Context, w http.ResponseWriter, message string) {
	apiErr := &Error{
		Code:    "BAD_REQUEST",
		Message: message,
	}

	a.Error(ctx, w, http.StatusBadRequest, apiErr)
}

// Unauthorized sends a 401 Unauthorized response
func (a *api) Unauthorized(ctx context.Context, w http.ResponseWriter, message string) {
	apiErr := &Error{
		Code:    "UNAUTHORIZED",
		Message: message,
	}

	a.Error(ctx, w, http.StatusUnauthorized, apiErr)
}

// Forbidden sends a 403 Forbidden response
func (a *api) Forbidden(ctx context.Context, w http.ResponseWriter, message string) {
	apiErr := &Error{
		Code:    "FORBIDDEN",
		Message: message,
	}

	a.Error(ctx, w, http.StatusForbidden, apiErr)
}

// NotFound sends a 404 Not Found response
func (a *api) NotFound(ctx context.Context, w http.ResponseWriter, message string) {
	apiErr := &Error{
		Code:    "NOT_FOUND",
		Message: message,
	}

	a.Error(ctx, w, http.StatusNotFound, apiErr)
}

// InternalServerError sends a 500 Internal Server Error response
func (a *api) InternalServerError(ctx context.Context, w http.ResponseWriter, message string) {
	apiErr := &Error{
		Code:    "INTERNAL_SERVER_ERROR",
		Message: message,
	}

	a.Error(ctx, w, http.StatusInternalServerError, apiErr)
}

// ValidationError sends a 422 Unprocessable Entity response with validation details
func (a *api) ValidationError(ctx context.Context, w http.ResponseWriter, details []ErrorDetail) {
	apiErr := &Error{
		Code:    "VALIDATION_ERROR",
		Message: "Validation failed",
		Details: details,
	}

	a.Error(ctx, w, http.StatusUnprocessableEntity, apiErr)
}
