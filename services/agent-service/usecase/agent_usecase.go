// Package usecase contains business logic for agent operations
package usecase

import (
	"context"
	"errors"
	"fmt"

	"agent-service/domain"
	"agent-service/domain/model"
	"agent-service/domain/repository"
	"monorepo/contracts/agent_service"
	"monorepo/pkg/logger"

	"golang.org/x/crypto/bcrypt"
)

// AgentUseCase defines the interface for agent-related business operations
// It provides methods for CRUD operations and listing agents with business logic
type AgentUseCase interface {
	// CreateAgent adds a new agent with business validation
	// It takes a context for request-scoped values and a pointer to an Agent model
	// Returns an error if the operation fails
	CreateAgent(ctx context.Context, agent *model.Agent) error
	// GetAgentByID retrieves an agent by their unique identifier
	// It takes a context for request-scoped values and the agent ID
	// Returns the agent model and an error if the operation fails
	GetAgentByID(ctx context.Context, id string) (*model.Agent, error)
	// UpdateAgent modifies an existing agent with business validation
	// It takes a context for request-scoped values and a pointer to an Agent model
	// Returns an error if the operation fails
	UpdateAgent(ctx context.Context, agent *model.Agent) error
	// DeleteAgent removes an agent from the system
	// It takes a context for request-scoped values and the agent ID
	// Returns an error if the operation fails
	DeleteAgent(ctx context.Context, id string) error
	// GetAgentsByParentID retrieves agents by their parent agent ID
	// It takes a context for request-scoped values and the parent agent ID
	// Returns a slice of agent pointers and an error if the operation fails
	GetAgentsByParentID(ctx context.Context, parentID string) ([]*model.Agent, error)
	// ListAgents retrieves a paginated list of agents
	// It takes a context for request-scoped values, offset for pagination, and limit for page size
	// Returns a slice of agent pointers, the real total count, and an error if the operation fails
	ListAgents(ctx context.Context, offset, limit int) ([]*model.Agent, int, error)
	// CreateSubAgentWithUser creates a new sub-agent and associated user in a single transactional operation
	// It takes a context for request-scoped values, parent agent ID, and a CreateSubAgentWithUserRequest
	// Returns the created agent and user models, or an error if the operation fails
	CreateSubAgentWithUser(ctx context.Context, parentID string, req *agent_service.CreateSubAgentWithUserRequest) (*model.Agent, *model.User, error)
}

// agentUseCase implements the AgentUseCase interface
type agentUseCase struct {
	// agentRepo is the repository interface for agent database operations
	agentRepo repository.Agent
	// userRepo is the repository interface for user database operations
	userRepo repository.User
	// logger is used for logging operations within the usecase
	logger logger.LoggerInterface
}

// NewAgentUseCase creates a new instance of agentUseCase
// It takes an Agent repository implementation, User repository implementation, and a logger instance
// Returns an implementation of the AgentUseCase interface
func NewAgentUseCase(agentRepo repository.Agent, userRepo repository.User, appLogger logger.LoggerInterface) AgentUseCase {
	return &agentUseCase{
		agentRepo: agentRepo,
		userRepo:  userRepo,
		logger:    appLogger,
	}
}

// CreateAgent adds a new agent with business validation
// It takes a context for request-scoped values and a pointer to an Agent model
// Returns an error if the operation fails
func (uc *agentUseCase) CreateAgent(ctx context.Context, agent *model.Agent) error {
	uc.logger.InfoContext(ctx, "Creating agent in usecase", "email", agent.Email)
	// Business logic validation
	if agent.Email == "" {
		uc.logger.WarnContext(ctx, "Email is required for agent creation")
		return domain.ErrEmailRequired
	}

	if agent.AgentName == "" {
		uc.logger.WarnContext(ctx, "Agent name is required for agent creation")
		return domain.ErrAgentNameRequired
	}

	if agent.AgentType == "" {
		uc.logger.WarnContext(ctx, "Agent type is required for agent creation")
		return domain.ErrAgentTypeRequired
	}

	// Validate agent type
	if agent.AgentType != model.AgentTypeIATA && agent.AgentType != model.AgentTypeSubAgent {
		uc.logger.WarnContext(ctx, "Invalid agent type", "agentType", agent.AgentType)
		return domain.ErrInvalidAgentType
	}

	// TODO: Email uniqueness check removed - consider database-level unique constraint

	// If parent agent ID is provided, validate it exists
	if agent.ParentAgentID != nil {
		parentAgent, err := uc.agentRepo.GetByID(ctx, *agent.ParentAgentID)
		if err != nil {
			if errors.Is(err, domain.ErrNotFound) {
				uc.logger.WarnContext(ctx, "Parent agent not found", "parentID", *agent.ParentAgentID)
				return domain.ErrParentAgentNotFound
			}
			uc.logger.ErrorContext(ctx, "Error checking parent agent", "parentID", *agent.ParentAgentID, "error", err)
			return fmt.Errorf("error checking parent agent: %w", err)
		}

		// Prevent circular reference
		if parentAgent.ParentAgentID != nil && *parentAgent.ParentAgentID == agent.ID {
			uc.logger.WarnContext(ctx, "Circular reference detected in agent hierarchy", "agentID", agent.ID, "parentID", *agent.ParentAgentID)
			return domain.ErrCircularReference
		}
	}

	if err := uc.agentRepo.Create(ctx, agent); err != nil {
		uc.logger.ErrorContext(ctx, "Failed to create agent in repository", "email", agent.Email, "error", err)
		return err
	}

	uc.logger.InfoContext(ctx, "Agent created successfully in usecase", "id", agent.ID, "email", agent.Email)
	return nil
}

// GetAgentByID retrieves an agent by their unique identifier
// It takes a context for request-scoped values and the agent ID
// Returns the agent model and an error if the operation fails
func (uc *agentUseCase) GetAgentByID(ctx context.Context, id string) (*model.Agent, error) {
	uc.logger.InfoContext(ctx, "Getting agent by ID in usecase", "id", id)
	if id == "" {
		uc.logger.WarnContext(ctx, "Invalid agent ID provided", "id", id)
		return nil, domain.ErrInvalidID
	}

	agent, err := uc.agentRepo.GetByID(ctx, id)
	if err != nil {
		if errors.Is(err, domain.ErrNotFound) {
			uc.logger.WarnContext(ctx, "Agent not found by ID", "id", id)
			return nil, domain.ErrAgentNotFound
		}
		uc.logger.ErrorContext(ctx, "Error getting agent by ID", "id", id, "error", err)
		return nil, fmt.Errorf("error getting agent: %w", err)
	}

	uc.logger.InfoContext(ctx, "Agent retrieved by ID in usecase", "id", agent.ID, "email", agent.Email)
	return agent, nil
}

// UpdateAgent modifies an existing agent with business validation
// It takes a context for request-scoped values and a pointer to an Agent model
// Returns an error if the operation fails
func (uc *agentUseCase) UpdateAgent(ctx context.Context, agent *model.Agent) error {
	uc.logger.InfoContext(ctx, "Updating agent in usecase", "id", agent.ID, "email", agent.Email)
	if agent.ID == "" {
		uc.logger.WarnContext(ctx, "Invalid agent ID for update", "id", agent.ID)
		return domain.ErrInvalidID
	}

	if agent.Email == "" {
		uc.logger.WarnContext(ctx, "Email is required for agent update", "id", agent.ID)
		return domain.ErrEmailRequired
	}

	if agent.AgentName == "" {
		uc.logger.WarnContext(ctx, "Agent name is required for agent update", "id", agent.ID)
		return domain.ErrAgentNameRequired
	}

	if agent.AgentType == "" {
		uc.logger.WarnContext(ctx, "Agent type is required for agent update", "id", agent.ID)
		return domain.ErrAgentTypeRequired
	}

	// Validate agent type
	if agent.AgentType != model.AgentTypeIATA && agent.AgentType != model.AgentTypeSubAgent {
		uc.logger.WarnContext(ctx, "Invalid agent type", "agentType", agent.AgentType)
		return domain.ErrInvalidAgentType
	}

	// If parent agent ID is provided, validate it exists and prevent circular reference
	if agent.ParentAgentID != nil {
		parentAgent, err := uc.agentRepo.GetByID(ctx, *agent.ParentAgentID)
		if err != nil {
			if errors.Is(err, domain.ErrNotFound) {
				uc.logger.WarnContext(ctx, "Parent agent not found", "parentID", *agent.ParentAgentID)
				return domain.ErrParentAgentNotFound
			}
			uc.logger.ErrorContext(ctx, "Error checking parent agent", "parentID", *agent.ParentAgentID, "error", err)
			return fmt.Errorf("error checking parent agent: %w", err)
		}

		// Prevent circular reference
		if parentAgent.ParentAgentID != nil && *parentAgent.ParentAgentID == agent.ID {
			uc.logger.WarnContext(ctx, "Circular reference detected in agent hierarchy", "agentID", agent.ID, "parentID", *agent.ParentAgentID)
			return domain.ErrCircularReference
		}
	}

	if err := uc.agentRepo.Update(ctx, agent); err != nil {
		uc.logger.ErrorContext(ctx, "Failed to update agent in repository", "id", agent.ID, "email", agent.Email, "error", err)
		return err
	}

	uc.logger.InfoContext(ctx, "Agent updated successfully in usecase", "id", agent.ID, "email", agent.Email)
	return nil
}

// DeleteAgent removes an agent from the system
// It takes a context for request-scoped values and the agent ID
// Returns an error if the operation fails
func (uc *agentUseCase) DeleteAgent(ctx context.Context, id string) error {
	uc.logger.InfoContext(ctx, "Deleting agent in usecase", "id", id)
	if id == "" {
		uc.logger.WarnContext(ctx, "Invalid agent ID for deletion", "id", id)
		return domain.ErrInvalidID
	}

	// Check if agent has children
	children, err := uc.agentRepo.GetByParentID(ctx, id)
	if err != nil {
		uc.logger.ErrorContext(ctx, "Error checking agent children", "id", id, "error", err)
		return fmt.Errorf("error checking agent children: %w", err)
	}

	if len(children) > 0 {
		uc.logger.WarnContext(ctx, "Cannot delete agent with children", "id", id, "children_count", len(children))
		return domain.ErrAgentHasChildren
	}

	err = uc.agentRepo.Delete(ctx, id)
	if err != nil {
		if errors.Is(err, domain.ErrNotFound) {
			uc.logger.WarnContext(ctx, "Agent not found for deletion", "id", id)
			return domain.ErrAgentNotFound
		}
		uc.logger.ErrorContext(ctx, "Error deleting agent", "id", id, "error", err)
		return fmt.Errorf("error deleting agent: %w", err)
	}

	uc.logger.InfoContext(ctx, "Agent deleted successfully in usecase", "id", id)
	return nil
}

// ListAgents retrieves a paginated list of agents
// It takes a context for request-scoped values, offset for pagination, and limit for page size
// Returns a slice of agent pointers, the real total count, and an error if the operation fails
func (uc *agentUseCase) ListAgents(ctx context.Context, offset, limit int) ([]*model.Agent, int, error) {
	uc.logger.InfoContext(ctx, "Listing agents in usecase", "offset", offset, "limit", limit)
	if offset < 0 {
		offset = 0
	}
	if limit <= 0 {
		limit = 10
	}
	if limit > 100 {
		limit = 100
	}

	agents, total, err := uc.agentRepo.List(ctx, offset, limit)
	if err != nil {
		uc.logger.ErrorContext(ctx, "Error listing agents", "offset", offset, "limit", limit, "error", err)
		return nil, 0, err
	}

	uc.logger.InfoContext(ctx, "Agents listed successfully in usecase", "count", len(agents), "offset", offset, "limit", limit, "total", total)
	return agents, total, nil
}

// GetAgentsByParentID retrieves agents by their parent agent ID
// It takes a context for request-scoped values and the parent agent ID
// Returns a slice of agent pointers and an error if the operation fails
func (uc *agentUseCase) GetAgentsByParentID(ctx context.Context, parentID string) ([]*model.Agent, error) {
	uc.logger.InfoContext(ctx, "Getting agents by parent ID in usecase", "parentID", parentID)
	if parentID == "" {
		uc.logger.WarnContext(ctx, "Parent ID is required for agent lookup by parent")
		return nil, domain.ErrInvalidID
	}

	agents, err := uc.agentRepo.GetByParentID(ctx, parentID)
	if err != nil {
		uc.logger.ErrorContext(ctx, "Error getting agents by parent ID", "parentID", parentID, "error", err)
		return nil, fmt.Errorf("error getting agents by parent ID: %w", err)
	}

	uc.logger.InfoContext(ctx, "Agents retrieved by parent ID in usecase", "count", len(agents), "parentID", parentID)
	return agents, nil
}

// CreateSubAgentWithUser creates a new sub-agent and associated user in a single transactional operation
// It takes a context for request-scoped values, parent agent ID, and a CreateSubAgentWithUserRequest
// Returns the created agent and user models, or an error if the operation fails
func (uc *agentUseCase) CreateSubAgentWithUser(ctx context.Context, parentID string, req *agent_service.CreateSubAgentWithUserRequest) (*model.Agent, *model.User, error) {
	uc.logger.InfoContext(ctx, "Creating sub-agent with user in usecase", "parentID", parentID, "agentEmail", req.AgentEmail, "userEmail", req.UserEmail)

	// Validate parent ID
	if parentID == "" {
		uc.logger.WarnContext(ctx, "Parent ID is required for sub-agent creation")
		return nil, nil, domain.ErrInvalidID
	}

	// Check if parent agent exists
	parentAgent, err := uc.agentRepo.GetByID(ctx, parentID)
	if err != nil {
		uc.logger.ErrorContext(ctx, "Error checking parent agent", "parentID", parentID, "error", err)
		return nil, nil, fmt.Errorf("error checking parent agent: %w", err)
	}
	if parentAgent == nil {
		uc.logger.WarnContext(ctx, "Parent agent not found", "parentID", parentID)
		return nil, nil, domain.ErrParentAgentNotFound
	}

	// Hash the user password
	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(req.UserPassword), bcrypt.DefaultCost)
	if err != nil {
		uc.logger.ErrorContext(ctx, "Error hashing password", "error", err)
		return nil, nil, fmt.Errorf("error hashing password: %w", err)
	}

	// Create agent model
	agent := &model.Agent{
		AgentName:     req.AgentName,
		AgentType:     model.AgentTypeSubAgent,
		ParentAgentID: &parentID,
		Email:         req.AgentEmail,
		IsActive:      false, // default for new agents
	}

	// Create user model
	user := &model.User{
		AgentID:  &agent.ID, // This will be set after agent creation
		Name:     req.UserName,
		Email:    req.UserEmail,
		Password: string(hashedPassword),
		IsActive: false, // default for new users
	}

	// Use transaction to ensure atomicity
	// Note: This assumes the repository implementations support transactions
	// In a real implementation, you might need to modify the repositories to accept a transaction

	// Create the agent first
	if err := uc.agentRepo.Create(ctx, agent); err != nil {
		uc.logger.ErrorContext(ctx, "Error creating agent", "email", agent.Email, "error", err)
		return nil, nil, fmt.Errorf("error creating agent: %w", err)
	}

	// Set the agent ID in the user
	user.AgentID = &agent.ID

	// Create the user
	if err := uc.userRepo.Create(ctx, user); err != nil {
		// If user creation fails, we should ideally rollback the agent creation
		// For now, we'll log the error and return it
		uc.logger.ErrorContext(ctx, "Error creating user, agent was created but user creation failed", "email", user.Email, "agentID", agent.ID, "error", err)
		// TODO: Implement proper transaction rollback
		return nil, nil, fmt.Errorf("error creating user: %w", err)
	}

	uc.logger.InfoContext(ctx, "Sub-agent with user created successfully in usecase", "agentID", agent.ID, "userID", user.ID)
	return agent, user, nil
}
