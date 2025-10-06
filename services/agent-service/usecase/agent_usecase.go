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

// AgentUseCase defines business operations for agents
type AgentUseCase interface {
	CreateAgent(ctx context.Context, agent *model.Agent) error
	GetAgentByID(ctx context.Context, id string) (*model.Agent, error)
	UpdateAgent(ctx context.Context, agent *model.Agent) error
	DeleteAgent(ctx context.Context, id string) error
	GetAgentsByParentID(ctx context.Context, parentID string) ([]*model.Agent, error)
	ListAgents(ctx context.Context, offset, limit int) ([]*model.Agent, int, error)
	CreateSubAgentWithUser(ctx context.Context, parentID string, req *agent_service.CreateSubAgentWithUserRequest) (*model.Agent, *model.User, error)
}

// agentUseCase implements the AgentUseCase interface
type agentUseCase struct {
	// agentRepo is the repository interface for agent database operations
	agentRepo repository.TransactionalAgent
	// userRepo is the repository interface for user database operations
	userRepo repository.TransactionalUser
	// logger is used for logging operations within the usecase
	logger logger.LoggerInterface
}

// NewAgentUseCase creates a new instance of agentUseCase
func NewAgentUseCase(agentRepo repository.TransactionalAgent, userRepo repository.TransactionalUser, appLogger logger.LoggerInterface) AgentUseCase {
	return &agentUseCase{
		agentRepo: agentRepo,
		userRepo:  userRepo,
		logger:    appLogger,
	}
}

// CreateAgent creates a new agent
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

	// Check if email already exists
	existingAgent, err := uc.agentRepo.GetByEmail(ctx, agent.Email)
	if err != nil && !errors.Is(err, domain.ErrNotFound) {
		uc.logger.ErrorContext(ctx, "Error checking email uniqueness", "email", agent.Email, "error", err)
		return fmt.Errorf("error checking email uniqueness: %w", err)
	}
	if existingAgent != nil {
		uc.logger.WarnContext(ctx, "Agent with this email already exists", "email", agent.Email)
		return domain.ErrAgentEmailAlreadyExists
	}

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

// GetAgentByID retrieves an agent by ID
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

// UpdateAgent updates an existing agent
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

	// Check if email already exists for another agent
	existingAgent, err := uc.agentRepo.GetByEmail(ctx, agent.Email)
	if err != nil && !errors.Is(err, domain.ErrNotFound) {
		uc.logger.ErrorContext(ctx, "Error checking email uniqueness", "email", agent.Email, "error", err)
		return fmt.Errorf("error checking email uniqueness: %w", err)
	}
	if existingAgent != nil && existingAgent.ID != agent.ID {
		uc.logger.WarnContext(ctx, "Agent with this email already exists", "email", agent.Email, "existingAgentID", existingAgent.ID)
		return domain.ErrAgentEmailAlreadyExists
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

// DeleteAgent deletes an agent
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

// ListAgents returns a paginated list of agents
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

// GetAgentsByParentID retrieves agents by parent ID
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

// CreateSubAgentWithUser creates a sub-agent with user
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
	// Both agent and user creation must succeed or both must fail
	err = uc.agentRepo.ExecuteInTransaction(ctx, func(txCtx context.Context) error {
		// Create the agent within the transaction
		if err := uc.agentRepo.Create(txCtx, agent); err != nil {
			uc.logger.ErrorContext(ctx, "Error creating agent in transaction", "email", agent.Email, "error", err)
			return fmt.Errorf("error creating agent: %w", err)
		}

		// Set the agent ID in the user
		user.AgentID = &agent.ID

		// Create the user within the same transaction
		if err := uc.userRepo.Create(txCtx, user); err != nil {
			uc.logger.ErrorContext(ctx, "Error creating user in transaction", "email", user.Email, "error", err)
			return fmt.Errorf("error creating user: %w", err)
		}

		return nil // Commit the transaction
	})

	if err != nil {
		uc.logger.ErrorContext(ctx, "Transaction failed for sub-agent with user creation", "parentID", parentID, "error", err)
		return nil, nil, err
	}

	uc.logger.InfoContext(ctx, "Sub-agent with user created successfully in usecase", "agentID", agent.ID, "userID", user.ID)
	return agent, user, nil
}
