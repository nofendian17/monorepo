// Package repository defines the interfaces for data access layer
package repository

import (
	"agent-service/domain/model"
	"context"
)

// User defines the contract for user-related database operations
type User interface {
	Create(ctx context.Context, user *model.User) error
	GetByID(ctx context.Context, id string) (*model.User, error)
	GetByEmail(ctx context.Context, email string) (*model.User, error)
	GetByAgentID(ctx context.Context, agentID string) ([]*model.User, error)
	GetActiveUsers(ctx context.Context) ([]*model.User, error)
	Update(ctx context.Context, user *model.User) error
	UpdatePassword(ctx context.Context, id string, hashedPassword string) error
	Delete(ctx context.Context, id string) error
	List(ctx context.Context, offset, limit int) ([]*model.User, int, error)
}

// TransactionalUser extends User with transactional operations
type TransactionalUser interface {
	User
	ExecuteInTransaction(ctx context.Context, fn func(txCtx context.Context) error) error
}
