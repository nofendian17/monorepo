// Package repository defines the interfaces for data access layer
package repository

import (
	"agent-service/domain/model"
	"context"
)

// Agent interface defines the contract for agent-related database operations
// It provides methods for CRUD operations and listing agents
type Agent interface {
	// Create adds a new agent to the database
	// It takes a context for request-scoped values and a pointer to an Agent model
	// Returns an error if the operation fails
	Create(ctx context.Context, agent *model.Agent) error
	// GetByID retrieves an agent by their unique identifier
	// It takes a context for request-scoped values and the agent ID
	// Returns the agent model and an error if the operation fails
	GetByID(ctx context.Context, id string) (*model.Agent, error)
	// GetByEmail retrieves an agent by their email address
	// It takes a context for request-scoped values and the agent email
	// Returns the agent model and an error if the operation fails
	GetByEmail(ctx context.Context, email string) (*model.Agent, error)
	// GetByParentID retrieves agents by their parent agent ID
	// It takes a context for request-scoped values and the parent agent ID
	// Returns a slice of agent pointers and an error if the operation fails
	GetByParentID(ctx context.Context, parentID string) ([]*model.Agent, error)
	// Update modifies an existing agent in the database
	// It takes a context for request-scoped values and a pointer to an Agent model
	// Returns an error if the operation fails
	Update(ctx context.Context, agent *model.Agent) error
	// Delete removes an agent from the database (soft delete)
	// It takes a context for request-scoped values and the agent ID
	// Returns an error if the operation fails
	Delete(ctx context.Context, id string) error
	// List retrieves a paginated list of agents from the database
	// It takes a context for request-scoped values, offset for pagination, and limit for page size
	// Returns a slice of agent pointers and an error if the operation fails
	List(ctx context.Context, offset, limit int) ([]*model.Agent, int, error)
}
