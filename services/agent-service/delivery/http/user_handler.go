// Package http contains HTTP delivery implementations for the application
package http

import (
	"context"
	"encoding/json"
	"net/http"
	"strconv"

	"agent-service/domain"
	"agent-service/domain/model"
	"agent-service/usecase"
	"monorepo/pkg/api"
	"monorepo/pkg/logger"
	"monorepo/pkg/validator"

	"github.com/go-chi/chi/v5"
)

// UserHandler handles HTTP requests for user operations
type UserHandler struct {
	// UserUseCase contains business logic for user operations
	UserUseCase usecase.UserUseCase
	// Logger is used for logging operations within the handler
	Logger logger.LoggerInterface
	// API provides standardized API response patterns
	API api.Api
}

// NewUserHandler creates a new instance of UserHandler
// It takes a UserUseCase implementation and a logger instance
// Returns a pointer to a UserHandler
func NewUserHandler(userUseCase usecase.UserUseCase, logger logger.LoggerInterface) *UserHandler {
	return &UserHandler{
		UserUseCase: userUseCase,
		Logger:      logger,
		API:         api.New(),
	}
}

// CreateHandler handles HTTP requests to create a new user
// It expects a JSON payload with user data in the request body
// Returns a 201 status code with the created user on success
// Returns a 400 status code for invalid request data
// Returns a 42 status code for validation errors
// Returns a 500 status code for internal server errors
func (h *UserHandler) CreateHandler(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	h.Logger.InfoContext(ctx, "Create user handler called")

	var user model.User
	if err := json.NewDecoder(r.Body).Decode(&user); err != nil {
		h.Logger.ErrorContext(ctx, "Invalid request body for user creation", "error", err)
		h.API.BadRequest(ctx, w, "Invalid request body")
		return
	}

	// Validate the user input using the new validator
	validationErrors := validator.ValidateStruct(&user)
	if validationErrors != nil {
		h.Logger.WarnContext(ctx, "Validation failed for user creation", "errors", validationErrors)

		// For now, return a generic validation error
		// In a real implementation, you'd convert the validation errors properly
		h.API.ValidationError(ctx, w, []api.ErrorDetail{
			{
				Field:   "validation",
				Message: "Input validation failed",
			},
		})
		return
	}

	if err := h.UserUseCase.CreateUser(ctx, &user); err != nil {
		switch {
		case err.Error() == domain.ErrEmailRequired.Message:
			h.API.BadRequest(ctx, w, err.Error())
		case err.Error() == domain.ErrEmailAlreadyExists.Message:
			h.API.Error(ctx, w, domain.ErrEmailAlreadyExists.Code, &api.Error{
				Code:    "EMAIL_EXISTS",
				Message: err.Error(),
			})
		default:
			h.Logger.ErrorContext(ctx, "Unexpected error during user creation", "email", user.Email, "error", err)
			h.API.InternalServerError(ctx, w, "Failed to create user")
		}
		return
	}

	h.Logger.InfoContext(ctx, "User created successfully in handler", "id", user.ID, "email", user.Email)
	h.API.Success(ctx, w, user)
}

// parseULID parses a string ID parameter to ULID string, returning appropriate error responses
func (h *UserHandler) parseULID(ctx context.Context, w http.ResponseWriter, r *http.Request, paramName string) (string, bool) {
	idStr := chi.URLParam(r, paramName)
	if idStr == "" {
		h.Logger.ErrorContext(ctx, "User ID parameter is required", "param", paramName)
		h.API.BadRequest(ctx, w, "User ID parameter is required")
		return "", false
	}
	return idStr, true
}

// handleUserError handles user-related errors consistently
func (h *UserHandler) handleUserError(ctx context.Context, w http.ResponseWriter, err error, id string) {
	switch {
	case err.Error() == domain.ErrUserNotFound.Message:
		h.API.NotFound(ctx, w, err.Error())
	case err.Error() == domain.ErrInvalidID.Message:
		h.API.BadRequest(ctx, w, err.Error())
	case err.Error() == domain.ErrEmailRequired.Message:
		h.API.BadRequest(ctx, w, err.Error())
	case err.Error() == domain.ErrEmailAlreadyExists.Message:
		h.API.Error(ctx, w, domain.ErrEmailAlreadyExists.Code, &api.Error{
			Code:    "EMAIL_EXISTS",
			Message: err.Error(),
		})
	default:
		h.Logger.ErrorContext(ctx, "Unexpected error", "id", id, "error", err)
		h.API.InternalServerError(ctx, w, "An unexpected error occurred")
	}
}

// GetByIDHandler handles HTTP requests to retrieve a user by their ID
// It expects the user ID as a URL parameter
// Returns a 200 status code with the user data on success
// Returns a 400 status code for invalid ID format
// Returns a 404 status code if the user is not found
// Returns a 500 status code for internal server errors
func (h *UserHandler) GetByIDHandler(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	h.Logger.InfoContext(ctx, "Get user by ID handler called")

	id, ok := h.parseULID(ctx, w, r, "id")
	if !ok {
		return
	}

	user, err := h.UserUseCase.GetUserByID(ctx, id)
	if err != nil {
		h.handleUserError(ctx, w, err, id)
		return
	}

	h.Logger.InfoContext(ctx, "User retrieved by ID in handler", "id", user.ID, "email", user.Email)
	h.API.Success(ctx, w, user)
}

// GetByEmailHandler handles HTTP requests to retrieve a user by their email
// It expects the email as a URL parameter
// Returns a 20 status code with the user data on success
// Returns a 400 status code if the email parameter is missing
// Returns a 404 status code if the user is not found
// Returns a 500 status code for internal server errors
func (h *UserHandler) GetByEmailHandler(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	h.Logger.InfoContext(ctx, "Get user by email handler called")
	email := chi.URLParam(r, "email")
	if email == "" {
		h.Logger.WarnContext(ctx, "Email parameter is required")
		h.API.BadRequest(ctx, w, "Email parameter is required")
		return
	}

	user, err := h.UserUseCase.GetUserByEmail(ctx, email)
	if err != nil {
		switch {
		case err.Error() == domain.ErrUserNotFound.Message:
			h.API.NotFound(ctx, w, err.Error())
		case err.Error() == domain.ErrEmailRequired.Message:
			h.API.BadRequest(ctx, w, err.Error())
		default:
			h.Logger.ErrorContext(ctx, "Unexpected error getting user by email", "email", email, "error", err)
			h.API.InternalServerError(ctx, w, "Failed to retrieve user")
		}
		return
	}

	h.Logger.InfoContext(ctx, "User retrieved by email in handler", "id", user.ID, "email", user.Email)
	h.API.Success(ctx, w, user)
}

// UpdateHandler handles HTTP requests to update an existing user
// It expects the user ID as a URL parameter and user data in the request body
// Returns a 200 status code with the updated user on success
// Returns a 40 status code for invalid ID format or request data
// Returns a 422 status code for validation errors
// Returns a 404 status code if the user is not found
// Returns a 500 status code for internal server errors
func (h *UserHandler) UpdateHandler(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	h.Logger.InfoContext(ctx, "Update user handler called")

	id, ok := h.parseULID(ctx, w, r, "id")
	if !ok {
		return
	}

	var user model.User
	if err := json.NewDecoder(r.Body).Decode(&user); err != nil {
		h.Logger.ErrorContext(ctx, "Invalid request body for user update", "id", id, "error", err)
		h.API.BadRequest(ctx, w, "Invalid request body")
		return
	}
	user.ID = id

	// Validate the user input using the new validator
	validationErrors := validator.ValidateStruct(&user)
	if validationErrors != nil {
		h.Logger.WarnContext(ctx, "Validation failed for user update", "id", user.ID, "errors", validationErrors)

		// For now, return a generic validation error
		// In a real implementation, you'd convert the validation errors properly
		h.API.ValidationError(ctx, w, []api.ErrorDetail{
			{
				Field:   "validation",
				Message: "Input validation failed",
			},
		})
		return
	}

	if err := h.UserUseCase.UpdateUser(ctx, &user); err != nil {
		h.handleUserError(ctx, w, err, user.ID)
		return
	}

	h.Logger.InfoContext(ctx, "User updated successfully in handler", "id", user.ID, "email", user.Email)
	h.API.Success(ctx, w, user)
}

// DeleteHandler handles HTTP requests to delete a user
// It expects the user ID as a URL parameter
// Returns a 20 status code with a success message on success
// Returns a 40 status code for invalid ID format
// Returns a 404 status code if the user is not found
// Returns a 500 status code for internal server errors
func (h *UserHandler) DeleteHandler(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	h.Logger.InfoContext(ctx, "Delete user handler called")

	id, ok := h.parseULID(ctx, w, r, "id")
	if !ok {
		return
	}

	if err := h.UserUseCase.DeleteUser(ctx, id); err != nil {
		h.handleUserError(ctx, w, err, id)
		return
	}

	h.Logger.InfoContext(ctx, "User deleted successfully in handler", "id", id)
	h.API.Success(ctx, w, map[string]string{"message": "User deleted successfully"})
}

// ListHandler handles HTTP requests to list users with pagination
// It expects optional 'offset' and 'limit' query parameters
// Returns a 200 status code with a list of users on success
// Returns a 500 status code for internal server errors
func (h *UserHandler) ListHandler(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	h.Logger.InfoContext(ctx, "List users handler called")

	// Parse query parameters for pagination
	offset, err := strconv.Atoi(r.URL.Query().Get("offset"))
	if err != nil || offset < 0 {
		offset = 0
	}

	limit, err := strconv.Atoi(r.URL.Query().Get("limit"))
	if err != nil || limit <= 0 {
		limit = 10
	}

	if limit > 100 {
		limit = 100
	}

	// Get users and real total from usecase
	users, total, err := h.UserUseCase.ListUsers(ctx, offset, limit)
	if err != nil {
		h.Logger.ErrorContext(ctx, "Error listing users", "offset", offset, "limit", limit, "error", err)
		h.API.InternalServerError(ctx, w, "Failed to list users")
		return
	}

	// Pagination calculation
	if limit <= 0 {
		limit = 10
	}

	if total < 0 {
		total = 0
	}

	// Calculate totalPages (0 if no data, else ceiling division)
	totalPages := 0
	if total > 0 {
		totalPages = (total + limit - 1) / limit
	}

	// Calculate current page (1-based)
	page := 1
	if total > 0 && offset < total {
		page = offset/limit + 1
	} else if total > 0 && offset >= total {
		page = totalPages
	}

	// HasNextPage: true if there are more records after this page
	hasNextPage := false
	if total > 0 && offset+limit < total {
		hasNextPage = true
	}

	// HasPrevPage: true if offset > 0 and there is data
	hasPrevPage := false
	if total > 0 && offset > 0 {
		hasPrevPage = true
	}

	pagination := &api.Pagination{
		Page:        page,
		Limit:       limit,
		Total:       total,
		TotalPages:  totalPages,
		HasNextPage: hasNextPage,
		HasPrevPage: hasPrevPage,
	}

	meta := &api.Meta{
		Pagination: pagination,
	}

	h.Logger.InfoContext(ctx, "Users listed successfully in handler", "count", len(users), "offset", offset, "limit", limit, "total", total)
	h.API.SuccessWithMeta(ctx, w, users, meta)
}
