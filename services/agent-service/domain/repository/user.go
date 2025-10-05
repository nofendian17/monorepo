// Package repository defines the interfaces for data access layer
package repository

import (
	"agent-service/domain/model"
	"context"
)

// User interface defines the contract for user-related database operations
// It provides methods for CRUD operations and listing users
type User interface {
	// Create adds a new user to the database
	// It takes a context for request-scoped values and a pointer to a User model
	// Returns an error if the operation fails
	Create(ctx context.Context, user *model.User) error
	// GetByID retrieves a user by their unique identifier
	// It takes a context for request-scoped values and the user ID
	// Returns the user model and an error if the operation fails
	GetByID(ctx context.Context, id string) (*model.User, error)
	// GetByEmail retrieves a user by their email address
	// It takes a context for request-scoped values and the email address
	// Returns the user model and an error if the operation fails
	GetByEmail(ctx context.Context, email string) (*model.User, error)
	// GetByAgentID retrieves users by their associated agent ID
	// It takes a context for request-scoped values and the agent ID
	// Returns a slice of user pointers and an error if the operation fails
	GetByAgentID(ctx context.Context, agentID string) ([]*model.User, error)
	// GetActiveUsers retrieves all active users
	// It takes a context for request-scoped values
	// Returns a slice of user pointers and an error if the operation fails
	GetActiveUsers(ctx context.Context) ([]*model.User, error)
	// Update modifies an existing user in the database
	// It takes a context for request-scoped values and a pointer to a User model
	// Returns an error if the operation fails
	Update(ctx context.Context, user *model.User) error
	// UpdatePassword updates only the password of a user
	// It takes a context for request-scoped values, user ID, and hashed password
	// Returns an error if the operation fails
	UpdatePassword(ctx context.Context, id string, hashedPassword string) error
	// Delete removes a user from the database (soft delete)
	// It takes a context for request-scoped values and the user ID
	// Returns an error if the operation fails
	Delete(ctx context.Context, id string) error
	// List retrieves a paginated list of users from the database
	// It takes a context for request-scoped values, offset for pagination, and limit for page size
	// Returns a slice of user pointers and an error if the operation fails
	List(ctx context.Context, offset, limit int) ([]*model.User, int, error)
}
