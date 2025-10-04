// Package agent_service contains request and response contracts for the agent service
package agent_service

import (
	"time"

	"agent-service/domain/model"
)

// CreateUserRequest represents the request payload for creating a new user
type CreateUserRequest struct {
	AgentID         *string `json:"agent_id,omitempty" validate:"omitempty,ulid"`
	Name            string  `json:"name" validate:"required,min=1,max=255"`
	Email           string  `json:"email" validate:"required,email"`
	Password        string  `json:"password" validate:"required,min=8"`
	PasswordConfirm string  `json:"password_confirm" validate:"required,min=8,eqfield=Password"`
}

// UserResponse represents the response payload for a user
type UserResponse struct {
	ID        string         `json:"id"`
	AgentID   *string        `json:"agent_id,omitempty"`
	Agent     *AgentResponse `json:"agent,omitempty"`
	Name      string         `json:"name"`
	Email     string         `json:"email"`
	IsActive  bool           `json:"is_active"`
	CreatedAt string         `json:"created_at"`
	UpdatedAt string         `json:"updated_at"`
}

// AgentResponse represents the response payload for an agent
type AgentResponse struct {
	ID            string  `json:"id"`
	AgentName     string  `json:"agent_name"`
	AgentType     string  `json:"agent_type"`
	ParentAgentID *string `json:"parent_agent_id,omitempty"`
	Email         string  `json:"email"`
	IsActive      bool    `json:"is_active"`
	CreatedAt     string  `json:"created_at"`
	UpdatedAt     string  `json:"updated_at"`
}

// GetUserByIDRequest represents the request for getting a user by ID
type GetUserByIDRequest struct {
	ID string `validate:"required,ulid"`
}

// GetUserByEmailRequest represents the request for getting a user by email
type GetUserByEmailRequest struct {
	Email string `validate:"required,email"`
}

// DeleteUserRequest represents the request for deleting a user
type DeleteUserRequest struct {
	ID string `validate:"required,ulid"`
}

// UpdateUserRequest represents the request payload for updating an existing user
type UpdateUserRequest struct {
	ID              string  `json:"id" validate:"required,ulid"`
	AgentID         *string `json:"agent_id,omitempty" validate:"omitempty,ulid"`
	Name            string  `json:"name,omitempty" validate:"omitempty,min=1,max=255"`
	Email           string  `json:"email,omitempty" validate:"omitempty,email"`
	Password        string  `json:"password,omitempty" validate:"omitempty,min=8"`
	PasswordConfirm string  `json:"password_confirm,omitempty" validate:"omitempty,min=8,eqfield=Password"`
	IsActive        *bool   `json:"is_active,omitempty"`
}
type UsersListResponse struct {
	Users []UserResponse `json:"users"`
}

// CreateUserRequestToModel converts CreateUserRequest to model.User
func CreateUserRequestToModel(req *CreateUserRequest) *model.User {
	return &model.User{
		AgentID:  req.AgentID,
		Name:     req.Name,
		Email:    req.Email,
		Password: req.Password, // Plain password, will be hashed in usecase
		IsActive: false,        // default for new users
	}
}

// UserModelToResponse converts model.User to UserResponse
func UserModelToResponse(user *model.User) *UserResponse {
	resp := &UserResponse{
		ID:        user.ID,
		AgentID:   user.AgentID,
		Name:      user.Name,
		Email:     user.Email,
		IsActive:  user.IsActive,
		CreatedAt: user.CreatedAt.Format(time.RFC3339),
		UpdatedAt: user.UpdatedAt.Format(time.RFC3339),
	}
	if user.Agent.ID != "" {
		resp.Agent = AgentModelToResponse(&user.Agent)
	}
	return resp
}

// AgentModelToResponse converts model.Agent to AgentResponse
func AgentModelToResponse(agent *model.Agent) *AgentResponse {
	return &AgentResponse{
		ID:            agent.ID,
		AgentName:     agent.AgentName,
		AgentType:     agent.AgentType,
		ParentAgentID: agent.ParentAgentID,
		Email:         agent.Email,
		IsActive:      agent.IsActive,
		CreatedAt:     agent.CreatedAt.Format(time.RFC3339),
		UpdatedAt:     agent.UpdatedAt.Format(time.RFC3339),
	}
}

// UserModelsToResponses converts slice of model.User to slice of UserResponse
func UserModelsToResponses(users []*model.User) []UserResponse {
	responses := make([]UserResponse, len(users))
	for i, user := range users {
		responses[i] = *UserModelToResponse(user)
	}
	return responses
}
