// Package agent_service contains request and response contracts for the agent service
package agent_service

import (
	"agent-service/domain/model"
)

// CreateAgentRequest represents the request payload for creating a new agent
type CreateAgentRequest struct {
	AgentName     string  `json:"agent_name" validate:"required,min=1,max=255"`
	AgentType     string  `json:"agent_type" validate:"required,oneof=IATA SUB_AGENT"`
	ParentAgentID *string `json:"parent_agent_id,omitempty" validate:"omitempty,ulid"`
	Email         string  `json:"email" validate:"required,email"`
}

// GetAgentByIDRequest represents the request for getting an agent by ID
type GetAgentByIDRequest struct {
	ID string `validate:"required,ulid"`
}

// GetAgentByEmailRequest represents the request for getting an agent by email
type GetAgentByEmailRequest struct {
	Email string `validate:"required,email"`
}

// DeleteAgentRequest represents the request for deleting an agent
type DeleteAgentRequest struct {
	ID string `validate:"required,ulid"`
}

// UpdateAgentRequest represents the request payload for updating an existing agent
type UpdateAgentRequest struct {
	ID            string  `json:"id" validate:"required,ulid"`
	AgentName     string  `json:"agent_name,omitempty" validate:"omitempty,min=1,max=255"`
	AgentType     string  `json:"agent_type,omitempty" validate:"omitempty,oneof=IATA SUB_AGENT"`
	ParentAgentID *string `json:"parent_agent_id,omitempty" validate:"omitempty,ulid"`
	Email         string  `json:"email,omitempty" validate:"omitempty,email"`
	IsActive      *bool   `json:"is_active,omitempty"`
}

// UpdateAgentStatusRequest represents the request payload for updating agent active status
type UpdateAgentStatusRequest struct {
	IsActive bool `json:"is_active" validate:"required"`
}

type AgentsListResponse struct {
	Agents []AgentResponse `json:"agents"`
}

// CreateAgentRequestToModel converts CreateAgentRequest to model.Agent
func CreateAgentRequestToModel(req *CreateAgentRequest) *model.Agent {
	agent := &model.Agent{
		AgentName:     req.AgentName,
		AgentType:     req.AgentType,
		ParentAgentID: req.ParentAgentID,
		Email:         req.Email,
		IsActive:      false, // default for new agents
	}

	return agent
}

// AgentModelsToResponses converts slice of model.Agent to slice of AgentResponse
func AgentModelsToResponses(agents []*model.Agent) []AgentResponse {
	responses := make([]AgentResponse, len(agents))
	for i, agent := range agents {
		responses[i] = *AgentModelToResponse(agent)
	}
	return responses
}
