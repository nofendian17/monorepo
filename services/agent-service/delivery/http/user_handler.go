// Package http contains HTTP delivery implementations for the application
package http

import (
	"context"
	"encoding/json"
	"net/http"
	"strconv"

	"agent-service/domain"
	"agent-service/usecase"
	"monorepo/contracts/agent_service"
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
// Returns a 422 status code for validation errors
// Returns a 500 status code for internal server errors
func (h *UserHandler) CreateHandler(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	h.Logger.InfoContext(ctx, "Create user handler called")

	var req agent_service.CreateUserRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		h.Logger.ErrorContext(ctx, "Invalid request body for user creation", "error", err)
		h.API.BadRequest(ctx, w, "Invalid request body")
		return
	}

	// Validate the user input using the new validator
	validationErrors := validator.ValidateStruct(&req)
	if validationErrors != nil {
		h.Logger.WarnContext(ctx, "Validation failed for user creation", "errors", validationErrors)
		h.API.ValidationError(ctx, w, h.convertValidationErrors(validationErrors))
		return
	}

	user := agent_service.CreateUserRequestToModel(&req)
	if err := h.UserUseCase.CreateUser(ctx, user); err != nil {
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
	h.API.Created(ctx, w, agent_service.UserModelToResponse(user))
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

	req := agent_service.GetUserByIDRequest{ID: chi.URLParam(r, "id")}
	if err := validator.ValidateStruct(&req); err != nil {
		h.Logger.WarnContext(ctx, "Validation failed for get user by ID", "errors", err)
		h.API.ValidationError(ctx, w, h.convertValidationErrors(err))
		return
	}

	user, err := h.UserUseCase.GetUserByID(ctx, req.ID)
	if err != nil {
		h.handleUserError(ctx, w, err, req.ID)
		return
	}

	h.Logger.InfoContext(ctx, "User retrieved by ID in handler", "id", user.ID, "email", user.Email)
	h.API.Success(ctx, w, agent_service.UserModelToResponse(user))
}

// GetByEmailHandler handles HTTP requests to retrieve a user by their email
// It expects the email as a URL parameter
// Returns a 200 status code with the user data on success
// Returns a 400 status code if the email parameter is missing
// Returns a 404 status code if the user is not found
// Returns a 500 status code for internal server errors
func (h *UserHandler) GetByEmailHandler(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	h.Logger.InfoContext(ctx, "Get user by email handler called")

	req := agent_service.GetUserByEmailRequest{Email: chi.URLParam(r, "email")}
	if err := validator.ValidateStruct(&req); err != nil {
		h.Logger.WarnContext(ctx, "Validation failed for get user by email", "errors", err)
		h.API.ValidationError(ctx, w, h.convertValidationErrors(err))
		return
	}

	user, err := h.UserUseCase.GetUserByEmail(ctx, req.Email)
	if err != nil {
		switch {
		case err.Error() == domain.ErrUserNotFound.Message:
			h.API.NotFound(ctx, w, err.Error())
		case err.Error() == domain.ErrEmailRequired.Message:
			h.API.BadRequest(ctx, w, err.Error())
		default:
			h.Logger.ErrorContext(ctx, "Unexpected error getting user by email", "email", req.Email, "error", err)
			h.API.InternalServerError(ctx, w, "Failed to retrieve user")
		}
		return
	}

	h.Logger.InfoContext(ctx, "User retrieved by email in handler", "id", user.ID, "email", user.Email)
	h.API.Success(ctx, w, agent_service.UserModelToResponse(user))
}

// UpdateHandler handles HTTP requests to update an existing user
// It expects the user ID as a URL parameter and user data in the request body
// Returns a 200 status code with the updated user on success
// Returns a 400 status code for invalid ID format or request data
// Returns a 422 status code for validation errors
// Returns a 404 status code if the user is not found
// Returns a 500 status code for internal server errors
func (h *UserHandler) UpdateHandler(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	h.Logger.InfoContext(ctx, "Update user handler called")

	var req agent_service.UpdateUserRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		h.Logger.ErrorContext(ctx, "Invalid request body for user update", "error", err)
		h.API.BadRequest(ctx, w, "Invalid request body")
		return
	}

	// Set ID from URL parameter
	req.ID = chi.URLParam(r, "id")

	// Validate the user input using the new validator
	validationErrors := validator.ValidateStruct(&req)
	if validationErrors != nil {
		h.Logger.WarnContext(ctx, "Validation failed for user update", "id", req.ID, "errors", validationErrors)
		h.API.ValidationError(ctx, w, h.convertValidationErrors(validationErrors))
		return
	}

	// Get existing user
	existingUser, err := h.UserUseCase.GetUserByID(ctx, req.ID)
	if err != nil {
		h.handleUserError(ctx, w, err, req.ID)
		return
	}

	// Apply updates
	if req.AgentID != nil {
		existingUser.AgentID = req.AgentID
	}
	if req.Name != "" {
		existingUser.Name = req.Name
	}
	if req.Email != "" {
		existingUser.Email = req.Email
	}
	if req.Password != "" {
		existingUser.Password = req.Password // Plain password, will be hashed in usecase
	}
	if req.IsActive != nil {
		existingUser.IsActive = *req.IsActive
	}

	if err := h.UserUseCase.UpdateUser(ctx, existingUser); err != nil {
		h.handleUserError(ctx, w, err, existingUser.ID)
		return
	}

	h.Logger.InfoContext(ctx, "User updated successfully in handler", "id", existingUser.ID, "email", existingUser.Email)
	h.API.Success(ctx, w, agent_service.UserModelToResponse(existingUser))
}

// UpdateStatusHandler handles HTTP requests to update user active status
// It expects the user ID as a URL parameter and status data in the request body
// Returns a 200 status code with the updated user on success
// Returns a 400 status code for invalid ID format or request data
// Returns a 422 status code for validation errors
// Returns a 404 status code if the user is not found
// Returns a 500 status code for internal server errors
func (h *UserHandler) UpdateStatusHandler(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	h.Logger.InfoContext(ctx, "Update user status handler called")

	var req agent_service.UpdateUserStatusRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		h.Logger.ErrorContext(ctx, "Invalid request body for user status update", "error", err)
		h.API.BadRequest(ctx, w, "Invalid request body")
		return
	}

	userID := chi.URLParam(r, "id")
	if userID == "" {
		h.Logger.WarnContext(ctx, "User ID is required for status update")
		h.API.BadRequest(ctx, w, "User ID is required")
		return
	}

	// Validate the user ID
	idReq := agent_service.GetUserByIDRequest{ID: userID}
	if err := validator.ValidateStruct(&idReq); err != nil {
		h.Logger.WarnContext(ctx, "Validation failed for user ID", "errors", err)
		h.API.ValidationError(ctx, w, h.convertValidationErrors(err))
		return
	}

	// Validate the status request
	if err := validator.ValidateStruct(&req); err != nil {
		h.Logger.WarnContext(ctx, "Validation failed for user status update", "errors", err)
		h.API.ValidationError(ctx, w, h.convertValidationErrors(err))
		return
	}

	if err := h.UserUseCase.UpdateUserStatus(ctx, userID, req.IsActive); err != nil {
		h.handleUserError(ctx, w, err, userID)
		return
	}

	// Get the updated user to return
	user, err := h.UserUseCase.GetUserByID(ctx, userID)
	if err != nil {
		h.handleUserError(ctx, w, err, userID)
		return
	}

	h.Logger.InfoContext(ctx, "User status updated successfully in handler", "id", user.ID, "isActive", user.IsActive)
	h.API.Success(ctx, w, agent_service.UserModelToResponse(user))
}

// DeleteHandler handles HTTP requests to delete a user
// It expects the user ID as a URL parameter
// Returns a 200 status code with a success message on success
// Returns a 400 status code for invalid ID format
// Returns a 404 status code if the user is not found
// Returns a 500 status code for internal server errors
func (h *UserHandler) DeleteHandler(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	h.Logger.InfoContext(ctx, "Delete user handler called")

	req := agent_service.DeleteUserRequest{ID: chi.URLParam(r, "id")}
	if err := validator.ValidateStruct(&req); err != nil {
		h.Logger.WarnContext(ctx, "Validation failed for delete user", "errors", err)
		h.API.ValidationError(ctx, w, h.convertValidationErrors(err))
		return
	}

	if err := h.UserUseCase.DeleteUser(ctx, req.ID); err != nil {
		h.handleUserError(ctx, w, err, req.ID)
		return
	}

	h.Logger.InfoContext(ctx, "User deleted successfully in handler", "id", req.ID)
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
	h.API.SuccessWithMeta(ctx, w, agent_service.UserModelsToResponses(users), meta)
}

// convertValidationErrors converts validator errors to API error details
func (h *UserHandler) convertValidationErrors(validationErrors map[string]string) []api.ErrorDetail {
	details := make([]api.ErrorDetail, 0, len(validationErrors))
	for field, message := range validationErrors {
		details = append(details, api.ErrorDetail{
			Field:   field,
			Message: message,
		})
	}
	return details
}
