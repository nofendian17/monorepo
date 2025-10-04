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

// UpdateStatusHandler handles HTTP requests to update agent active status
// It expects the agent ID as a URL parameter and status data in the request body
// Returns a 200 status code with the updated agent on success
// Returns a 400 status code for invalid ID format or request data
// Returns a 422 status code for validation errors
// Returns a 404 status code if the agent is not found
// Returns a 500 status code for internal server errors
func (h *AgentHandler) UpdateStatusHandler(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	h.Logger.InfoContext(ctx, "Update agent status handler called")

	var req agent_service.UpdateAgentStatusRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		h.Logger.ErrorContext(ctx, "Invalid request body for agent status update", "error", err)
		h.API.BadRequest(ctx, w, "Invalid request body")
		return
	}

	agentID := chi.URLParam(r, "id")
	if agentID == "" {
		h.Logger.WarnContext(ctx, "Agent ID is required for status update")
		h.API.BadRequest(ctx, w, "Agent ID is required")
		return
	}

	// Validate the agent ID
	idReq := agent_service.GetAgentByIDRequest{ID: agentID}
	if err := validator.ValidateStruct(&idReq); err != nil {
		h.Logger.WarnContext(ctx, "Validation failed for agent ID", "errors", err)
		h.API.ValidationError(ctx, w, h.convertValidationErrors(err))
		return
	}

	// Validate the status request
	if err := validator.ValidateStruct(&req); err != nil {
		h.Logger.WarnContext(ctx, "Validation failed for agent status update", "errors", err)
		h.API.ValidationError(ctx, w, h.convertValidationErrors(err))
		return
	}

	if err := h.AgentUseCase.UpdateAgentStatus(ctx, agentID, req.IsActive); err != nil {
		h.handleAgentError(ctx, w, err, agentID)
		return
	}

	// Get the updated agent to return
	agent, err := h.AgentUseCase.GetAgentByID(ctx, agentID)
	if err != nil {
		h.handleAgentError(ctx, w, err, agentID)
		return
	}

	h.Logger.InfoContext(ctx, "Agent status updated successfully in handler", "id", agent.ID, "isActive", agent.IsActive)
	h.API.Success(ctx, w, agent_service.AgentModelToResponse(agent))
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

// GetActiveHandler handles HTTP requests to get all active agents
// Returns a 200 status code with a list of active agents on success
// Returns a 500 status code for internal server errors
func (h *AgentHandler) GetActiveHandler(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	h.Logger.InfoContext(ctx, "Get active agents handler called")

	agents, err := h.AgentUseCase.GetActiveAgents(ctx)
	if err != nil {
		h.Logger.ErrorContext(ctx, "Error getting active agents", "error", err)
		h.API.InternalServerError(ctx, w, "Failed to get active agents")
		return
	}

	h.Logger.InfoContext(ctx, "Active agents retrieved successfully in handler", "count", len(agents))
	h.API.Success(ctx, w, agent_service.AgentModelsToResponses(agents))
}

// GetInactiveHandler handles HTTP requests to get all inactive agents
// Returns a 200 status code with a list of inactive agents on success
// Returns a 500 status code for internal server errors
func (h *AgentHandler) GetInactiveHandler(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	h.Logger.InfoContext(ctx, "Get inactive agents handler called")

	agents, err := h.AgentUseCase.GetInactiveAgents(ctx)
	if err != nil {
		h.Logger.ErrorContext(ctx, "Error getting inactive agents", "error", err)
		h.API.InternalServerError(ctx, w, "Failed to get inactive agents")
		return
	}

	h.Logger.InfoContext(ctx, "Inactive agents retrieved successfully in handler", "count", len(agents))
	h.API.Success(ctx, w, agent_service.AgentModelsToResponses(agents))
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
