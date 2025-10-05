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

// AgentHandler handles HTTP requests for agent operations
type AgentHandler struct {
	// AgentUseCase contains business logic for agent operations
	AgentUseCase usecase.AgentUseCase
	// Logger is used for logging operations within the handler
	Logger logger.LoggerInterface
	// API provides standardized API response patterns
	API api.Api
}

// NewAgentHandler creates a new instance of AgentHandler
// It takes an AgentUseCase implementation and a logger instance
// Returns a pointer to an AgentHandler
func NewAgentHandler(agentUseCase usecase.AgentUseCase, logger logger.LoggerInterface) *AgentHandler {
	return &AgentHandler{
		AgentUseCase: agentUseCase,
		Logger:       logger,
		API:          api.New(),
	}
}

// CreateHandler handles HTTP requests to create a new agent
// It expects a JSON payload with agent data in the request body
// Returns a 201 status code with the created agent on success
// Returns a 400 status code for invalid request data
// Returns a 422 status code for validation errors
// Returns a 500 status code for internal server errors
func (h *AgentHandler) CreateHandler(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	h.Logger.InfoContext(ctx, "Create agent handler called")

	var req agent_service.CreateAgentRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		h.Logger.ErrorContext(ctx, "Invalid request body for agent creation", "error", err)
		h.API.BadRequest(ctx, w, "Invalid request body")
		return
	}

	// Validate the agent input using the validator
	validationErrors := validator.ValidateStruct(&req)
	if validationErrors != nil {
		h.Logger.WarnContext(ctx, "Validation failed for agent creation", "errors", validationErrors)
		h.API.ValidationError(ctx, w, h.convertValidationErrors(validationErrors))
		return
	}

	agent := agent_service.CreateAgentRequestToModel(&req)
	if err := h.AgentUseCase.CreateAgent(ctx, agent); err != nil {
		switch {
		case err.Error() == domain.ErrEmailRequired.Message:
			h.API.BadRequest(ctx, w, err.Error())
		case err.Error() == domain.ErrAgentNameRequired.Message:
			h.API.BadRequest(ctx, w, err.Error())
		case err.Error() == domain.ErrAgentTypeRequired.Message:
			h.API.BadRequest(ctx, w, err.Error())
		case err.Error() == domain.ErrInvalidAgentType.Message:
			h.API.BadRequest(ctx, w, err.Error())
		case err.Error() == domain.ErrParentAgentNotFound.Message:
			h.API.NotFound(ctx, w, err.Error())
		case err.Error() == domain.ErrCircularReference.Message:
			h.API.BadRequest(ctx, w, err.Error())
		default:
			h.Logger.ErrorContext(ctx, "Unexpected error during agent creation", "email", agent.Email, "error", err)
			h.API.InternalServerError(ctx, w, "Failed to create agent")
		}
		return
	}

	h.Logger.InfoContext(ctx, "Agent created successfully in handler", "id", agent.ID, "email", agent.Email)
	h.API.Created(ctx, w, agent_service.AgentModelToResponse(agent))
}

// handleAgentError handles agent-related errors consistently
func (h *AgentHandler) handleAgentError(ctx context.Context, w http.ResponseWriter, err error, id string) {
	switch {
	case err.Error() == domain.ErrAgentNotFound.Message:
		h.API.NotFound(ctx, w, err.Error())
	case err.Error() == domain.ErrInvalidID.Message:
		h.API.BadRequest(ctx, w, err.Error())
	case err.Error() == domain.ErrEmailRequired.Message:
		h.API.BadRequest(ctx, w, err.Error())
	case err.Error() == domain.ErrAgentNameRequired.Message:
		h.API.BadRequest(ctx, w, err.Error())
	case err.Error() == domain.ErrAgentTypeRequired.Message:
		h.API.BadRequest(ctx, w, err.Error())
	case err.Error() == domain.ErrInvalidAgentType.Message:
		h.API.BadRequest(ctx, w, err.Error())
	case err.Error() == domain.ErrParentAgentNotFound.Message:
		h.API.NotFound(ctx, w, err.Error())
	case err.Error() == domain.ErrCircularReference.Message:
		h.API.BadRequest(ctx, w, err.Error())
	case err.Error() == domain.ErrAgentHasChildren.Message:
		h.API.BadRequest(ctx, w, err.Error())
	default:
		h.Logger.ErrorContext(ctx, "Unexpected error", "id", id, "error", err)
		h.API.InternalServerError(ctx, w, "An unexpected error occurred")
	}
}

// GetByIDHandler handles HTTP requests to retrieve an agent by their ID
// It expects the agent ID as a URL parameter
// Returns a 200 status code with the agent data on success
// Returns a 400 status code for invalid ID format
// Returns a 404 status code if the agent is not found
// Returns a 500 status code for internal server errors
func (h *AgentHandler) GetByIDHandler(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	h.Logger.InfoContext(ctx, "Get agent by ID handler called")

	req := agent_service.GetAgentByIDRequest{ID: chi.URLParam(r, "id")}
	if err := validator.ValidateStruct(&req); err != nil {
		h.Logger.WarnContext(ctx, "Validation failed for get agent by ID", "errors", err)
		h.API.ValidationError(ctx, w, h.convertValidationErrors(err))
		return
	}

	agent, err := h.AgentUseCase.GetAgentByID(ctx, req.ID)
	if err != nil {
		h.handleAgentError(ctx, w, err, req.ID)
		return
	}

	h.Logger.InfoContext(ctx, "Agent retrieved by ID in handler", "id", agent.ID, "email", agent.Email)
	h.API.Success(ctx, w, agent_service.AgentModelToResponse(agent))
}

// UpdateHandler handles HTTP requests to update an existing agent
// It expects the agent ID as a URL parameter and agent data in the request body
// Returns a 200 status code with the updated agent on success
// Returns a 400 status code for invalid ID format or request data
// Returns a 422 status code for validation errors
// Returns a 404 status code if the agent is not found
// Returns a 500 status code for internal server errors
func (h *AgentHandler) UpdateHandler(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	h.Logger.InfoContext(ctx, "Update agent handler called")

	var req agent_service.UpdateAgentRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		h.Logger.ErrorContext(ctx, "Invalid request body for agent update", "error", err)
		h.API.BadRequest(ctx, w, "Invalid request body")
		return
	}

	// Set ID from URL parameter
	req.ID = chi.URLParam(r, "id")

	// Validate the agent input using the validator
	validationErrors := validator.ValidateStruct(&req)
	if validationErrors != nil {
		h.Logger.WarnContext(ctx, "Validation failed for agent update", "id", req.ID, "errors", validationErrors)
		h.API.ValidationError(ctx, w, h.convertValidationErrors(validationErrors))
		return
	}

	// Get existing agent
	existingAgent, err := h.AgentUseCase.GetAgentByID(ctx, req.ID)
	if err != nil {
		h.handleAgentError(ctx, w, err, req.ID)
		return
	}

	// Apply updates
	if req.AgentName != "" {
		existingAgent.AgentName = req.AgentName
	}
	if req.AgentType != "" {
		existingAgent.AgentType = req.AgentType
	}
	if req.ParentAgentID != nil {
		existingAgent.ParentAgentID = req.ParentAgentID
	}
	if req.Email != "" {
		existingAgent.Email = req.Email
	}
	if req.IsActive != nil {
		existingAgent.IsActive = *req.IsActive
	}

	if err := h.AgentUseCase.UpdateAgent(ctx, existingAgent); err != nil {
		h.handleAgentError(ctx, w, err, existingAgent.ID)
		return
	}

	h.Logger.InfoContext(ctx, "Agent updated successfully in handler", "id", existingAgent.ID, "email", existingAgent.Email)
	h.API.Success(ctx, w, agent_service.AgentModelToResponse(existingAgent))
}

// DeleteHandler handles HTTP requests to delete an agent
// It expects the agent ID as a URL parameter
// Returns a 200 status code with a success message on success
// Returns a 400 status code for invalid ID format
// Returns a 404 status code if the agent is not found
// Returns a 500 status code for internal server errors
func (h *AgentHandler) DeleteHandler(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	h.Logger.InfoContext(ctx, "Delete agent handler called")

	req := agent_service.DeleteAgentRequest{ID: chi.URLParam(r, "id")}
	if err := validator.ValidateStruct(&req); err != nil {
		h.Logger.WarnContext(ctx, "Validation failed for delete agent", "errors", err)
		h.API.ValidationError(ctx, w, h.convertValidationErrors(err))
		return
	}

	if err := h.AgentUseCase.DeleteAgent(ctx, req.ID); err != nil {
		h.handleAgentError(ctx, w, err, req.ID)
		return
	}

	h.Logger.InfoContext(ctx, "Agent deleted successfully in handler", "id", req.ID)
	h.API.Success(ctx, w, map[string]string{"message": "Agent deleted successfully"})
}

// ListHandler handles HTTP requests to list agents with pagination
// It expects optional 'offset' and 'limit' query parameters
// Returns a 200 status code with a list of agents on success
// Returns a 500 status code for internal server errors
func (h *AgentHandler) ListHandler(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	h.Logger.InfoContext(ctx, "List agents handler called")

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

	// Get agents and real total from usecase
	agents, total, err := h.AgentUseCase.ListAgents(ctx, offset, limit)
	if err != nil {
		h.Logger.ErrorContext(ctx, "Error listing agents", "offset", offset, "limit", limit, "error", err)
		h.API.InternalServerError(ctx, w, "Failed to list agents")
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

	h.Logger.InfoContext(ctx, "Agents listed successfully in handler", "count", len(agents), "offset", offset, "limit", limit, "total", total)
	h.API.SuccessWithMeta(ctx, w, agent_service.AgentModelsToResponses(agents), meta)
}

// CreateSubAgentHandler handles HTTP requests to create a new sub-agent with an associated user
// It expects the parent agent ID as a URL parameter and sub-agent with user data in the request body
// The sub-agent will have AgentType set to SUB_AGENT and ParentAgentID set to the URL parameter
// Returns a 201 status code with the created sub-agent and user on success
// Returns a 400 status code for invalid request data
// Returns a 422 status code for validation errors
// Returns a 404 status code if the parent agent is not found
// Returns a 500 status code for internal server errors
func (h *AgentHandler) CreateSubAgentHandler(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	parentID := chi.URLParam(r, "id")
	h.Logger.InfoContext(ctx, "Create sub-agent with user handler called", "parent_id", parentID)

	var req agent_service.CreateSubAgentWithUserRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		h.Logger.ErrorContext(ctx, "Invalid request body for sub-agent with user creation", "error", err)
		h.API.BadRequest(ctx, w, "Invalid request body")
		return
	}

	// Validate the sub-agent with user input using the validator
	validationErrors := validator.ValidateStruct(&req)
	if validationErrors != nil {
		h.Logger.WarnContext(ctx, "Validation failed for sub-agent with user creation", "errors", validationErrors)
		h.API.ValidationError(ctx, w, h.convertValidationErrors(validationErrors))
		return
	}

	// Create agent and user in a single transactional operation
	agent, user, err := h.AgentUseCase.CreateSubAgentWithUser(ctx, parentID, &req)
	if err != nil {
		switch {
		case err.Error() == domain.ErrInvalidID.Message:
			h.API.BadRequest(ctx, w, err.Error())
		case err.Error() == domain.ErrParentAgentNotFound.Message:
			h.API.NotFound(ctx, w, err.Error())
		default:
			h.Logger.ErrorContext(ctx, "Unexpected error during sub-agent with user creation", "parent_id", parentID, "error", err)
			h.API.InternalServerError(ctx, w, "Failed to create sub-agent with user")
		}
		return
	}

	// Create response with both agent and user data
	response := map[string]interface{}{
		"agent": agent_service.AgentModelToResponse(agent),
		"user":  agent_service.UserModelToResponse(user),
	}

	h.Logger.InfoContext(ctx, "Sub-agent with user created successfully in handler", "agent_id", agent.ID, "user_id", user.ID, "parent_id", parentID)
	h.API.Created(ctx, w, response)
}

// ListSubAgentsHandler handles HTTP requests to list all sub-agents of a parent agent
// It expects the parent agent ID as a URL parameter
// Returns a 200 status code with a list of sub-agents on success
// Returns a 400 status code for invalid parent ID format
// Returns a 404 status code if the parent agent is not found
// Returns a 500 status code for internal server errors
func (h *AgentHandler) ListSubAgentsHandler(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	parentID := chi.URLParam(r, "id")
	h.Logger.InfoContext(ctx, "List sub-agents handler called", "parent_id", parentID)

	// Validate parent ID
	req := agent_service.GetAgentByIDRequest{ID: parentID}
	if err := validator.ValidateStruct(&req); err != nil {
		h.Logger.WarnContext(ctx, "Validation failed for list sub-agents", "errors", err)
		h.API.ValidationError(ctx, w, h.convertValidationErrors(err))
		return
	}

	// Check if parent agent exists
	_, err := h.AgentUseCase.GetAgentByID(ctx, parentID)
	if err != nil {
		h.handleAgentError(ctx, w, err, parentID)
		return
	}

	// Get sub-agents
	subAgents, err := h.AgentUseCase.GetAgentsByParentID(ctx, parentID)
	if err != nil {
		h.Logger.ErrorContext(ctx, "Error listing sub-agents", "parent_id", parentID, "error", err)
		h.API.InternalServerError(ctx, w, "Failed to list sub-agents")
		return
	}

	h.Logger.InfoContext(ctx, "Sub-agents listed successfully in handler", "count", len(subAgents), "parent_id", parentID)
	h.API.Success(ctx, w, agent_service.AgentModelsToResponses(subAgents))
}

// convertValidationErrors converts validator errors to API error details
func (h *AgentHandler) convertValidationErrors(validationErrors map[string]string) []api.ErrorDetail {
	details := make([]api.ErrorDetail, 0, len(validationErrors))
	for field, message := range validationErrors {
		details = append(details, api.ErrorDetail{
			Field:   field,
			Message: message,
		})
	}
	return details
}
