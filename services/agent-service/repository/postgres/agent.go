// Package postgres provides PostgreSQL implementation for agent repository
package postgres

import (
	"context"
	"fmt"

	"agent-service/domain"
	"agent-service/domain/model"
	"agent-service/domain/repository"
	"monorepo/pkg/logger"

	"gorm.io/gorm"
)

// agentRepository implements the Agent repository interface using PostgreSQL
type agentRepository struct {
	// db is the GORM database instance for database operations
	db *gorm.DB
	// logger is used for logging operations within the repository
	logger logger.LoggerInterface
}

// NewAgentRepository creates a new instance of agentRepository
// It takes a GORM database instance and a logger instance
// Returns an implementation of the TransactionalAgent repository interface
func NewAgentRepository(db *gorm.DB, logger logger.LoggerInterface) repository.TransactionalAgent {
	return &agentRepository{
		db:     db,
		logger: logger,
	}
}

// Create adds a new agent to the database
func (r *agentRepository) Create(ctx context.Context, agent *model.Agent) error {
	r.logger.InfoContext(ctx, "Creating agent", "email", agent.Email)

	// Check if there's a transaction in the context
	db := r.db
	if tx, ok := ctx.Value("tx").(*gorm.DB); ok {
		db = tx
	}

	if err := db.WithContext(ctx).Create(agent).Error; err != nil {
		r.logger.ErrorContext(ctx, "Failed to create agent", "email", agent.Email, "error", err)
		return fmt.Errorf("failed to create agent: %w", err)
	}
	r.logger.InfoContext(ctx, "Agent created successfully", "id", agent.ID, "email", agent.Email)
	return nil
}

// GetByID retrieves an agent by their unique identifier
// It takes a context for request-scoped values and the agent ID
// Returns the agent model and an error if the operation fails
func (r *agentRepository) GetByID(ctx context.Context, id string) (*model.Agent, error) {
	r.logger.InfoContext(ctx, "Getting agent by ID", "id", id)
	var agent model.Agent
	if err := r.db.WithContext(ctx).Preload("Parent").Preload("Children").Where("id = ? AND deleted_at IS NULL", id).First(&agent).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			r.logger.WarnContext(ctx, "Agent not found by ID", "id", id)
			return nil, domain.ErrNotFound
		}
		r.logger.ErrorContext(ctx, "Failed to get agent by ID", "id", id, "error", err)
		return nil, fmt.Errorf("failed to get agent: %w", err)
	}
	r.logger.InfoContext(ctx, "Agent retrieved by ID", "id", agent.ID, "email", agent.Email)
	return &agent, nil
}

// GetByEmail retrieves an agent by their email address
func (r *agentRepository) GetByEmail(ctx context.Context, email string) (*model.Agent, error) {
	r.logger.InfoContext(ctx, "Getting agent by email", "email", email)
	var agent model.Agent
	if err := r.db.WithContext(ctx).Preload("Parent").Preload("Children").Where("email = ? AND deleted_at IS NULL", email).First(&agent).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			r.logger.WarnContext(ctx, "Agent not found by email", "email", email)
			return nil, domain.ErrNotFound
		}
		r.logger.ErrorContext(ctx, "Failed to get agent by email", "email", email, "error", err)
		return nil, fmt.Errorf("failed to get agent: %w", err)
	}
	r.logger.InfoContext(ctx, "Agent retrieved by email", "id", agent.ID, "email", agent.Email)
	return &agent, nil
}

// Update modifies an existing agent in the database
func (r *agentRepository) Update(ctx context.Context, agent *model.Agent) error {
	r.logger.InfoContext(ctx, "Updating agent", "id", agent.ID, "email", agent.Email)
	if err := r.db.WithContext(ctx).Model(&model.Agent{}).Where("id = ?", agent.ID).Updates(agent).Error; err != nil {
		r.logger.ErrorContext(ctx, "Failed to update agent", "id", agent.ID, "email", agent.Email, "error", err)
		return fmt.Errorf("failed to update agent: %w", err)
	}
	r.logger.InfoContext(ctx, "Agent updated successfully", "id", agent.ID, "email", agent.Email)
	return nil
}

// Delete removes an agent from the database (soft delete)
// It takes a context for request-scoped values and the agent ID
// Returns an error if the operation fails
func (r *agentRepository) Delete(ctx context.Context, id string) error {
	r.logger.InfoContext(ctx, "Deleting agent", "id", id)
	agent := &model.Agent{ID: id}

	// Use soft delete
	if err := r.db.WithContext(ctx).Delete(agent).Error; err != nil {
		r.logger.ErrorContext(ctx, "Failed to delete agent", "id", id, "error", err)
		return fmt.Errorf("failed to delete agent: %w", err)
	}

	// Check if record was actually deleted
	var count int64
	r.db.WithContext(ctx).Model(&model.Agent{}).Where("id = ? AND deleted_at IS NULL", id).Count(&count)
	if count > 0 {
		r.logger.WarnContext(ctx, "Agent not found for deletion", "id", id)
		return domain.ErrNotFound
	}

	r.logger.InfoContext(ctx, "Agent deleted successfully", "id", id)
	return nil
}

// List retrieves a paginated list of agents from the database
// It takes a context for request-scoped values, offset for pagination, and limit for page size
// Returns a slice of agent pointers, the real total count, and an error if the operation fails
func (r *agentRepository) List(ctx context.Context, offset, limit int) ([]*model.Agent, int, error) {
	r.logger.InfoContext(ctx, "Listing agents", "offset", offset, "limit", limit)
	var agents []*model.Agent
	var total int64

	// Get total count
	if err := r.db.WithContext(ctx).Model(&model.Agent{}).Where("deleted_at IS NULL").Count(&total).Error; err != nil {
		r.logger.ErrorContext(ctx, "Failed to count agents", "error", err)
		return nil, 0, fmt.Errorf("failed to count agents: %w", err)
	}

	// Get paginated agents
	if err := r.db.WithContext(ctx).Preload("Parent").Preload("Children").Where("deleted_at IS NULL").Offset(offset).Limit(limit).Order("id ASC").Find(&agents).Error; err != nil {
		r.logger.ErrorContext(ctx, "Failed to list agents", "offset", offset, "limit", limit, "error", err)
		return nil, 0, fmt.Errorf("failed to list agents: %w", err)
	}

	r.logger.InfoContext(ctx, "Agents listed successfully", "count", len(agents), "offset", offset, "limit", limit, "total", total)
	return agents, int(total), nil
}

// GetByParentID retrieves agents by their parent agent ID
// It takes a context for request-scoped values and the parent agent ID
// Returns a slice of agent pointers and an error if the operation fails
func (r *agentRepository) GetByParentID(ctx context.Context, parentID string) ([]*model.Agent, error) {
	r.logger.InfoContext(ctx, "Getting agents by parent ID", "parentID", parentID)
	var agents []*model.Agent
	if err := r.db.WithContext(ctx).Preload("Parent").Preload("Children").Where("parent_agent_id = ? AND deleted_at IS NULL", parentID).Find(&agents).Error; err != nil {
		r.logger.ErrorContext(ctx, "Failed to get agents by parent ID", "parentID", parentID, "error", err)
		return nil, fmt.Errorf("failed to get agents by parent ID: %w", err)
	}
	r.logger.InfoContext(ctx, "Agents retrieved by parent ID", "count", len(agents), "parentID", parentID)
	return agents, nil
}

// ExecuteInTransaction executes a function within a database transaction
// The function receives a transaction context that should be used for all operations
// Returns an error if the transaction fails or if the function returns an error
func (r *agentRepository) ExecuteInTransaction(ctx context.Context, fn func(txCtx context.Context) error) error {
	r.logger.InfoContext(ctx, "Executing operation in transaction")
	return r.db.Transaction(func(tx *gorm.DB) error {
		// Create a context that carries the transaction
		txCtx := context.WithValue(ctx, "tx", tx)
		return fn(txCtx)
	})
}
