// Package repository defines the interfaces for data access layer
package repository

import (
	"agent-service/domain/model"
	"context"
)

// Agent defines the contract for agent-related database operations
type Agent interface {
	Create(ctx context.Context, agent *model.Agent) error
	GetByID(ctx context.Context, id string) (*model.Agent, error)
	GetByEmail(ctx context.Context, email string) (*model.Agent, error)
	GetByParentID(ctx context.Context, parentID string) ([]*model.Agent, error)
	Update(ctx context.Context, agent *model.Agent) error
	Delete(ctx context.Context, id string) error
	List(ctx context.Context, offset, limit int) ([]*model.Agent, int, error)
}

// TransactionalAgent extends Agent with transactional operations
type TransactionalAgent interface {
	Agent
	ExecuteInTransaction(ctx context.Context, fn func(txCtx context.Context) error) error
}
