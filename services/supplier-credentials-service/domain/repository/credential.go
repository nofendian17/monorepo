// Package repository defines the interfaces for data access layer
package repository

import (
	"context"
	"supplier-credentials-service/domain/model"
)

// Supplier defines supplier-related database operations
type Supplier interface {
	Create(ctx context.Context, supplier *model.Supplier) error
	GetByID(ctx context.Context, id string) (*model.Supplier, error)
	GetByCode(ctx context.Context, code string) (*model.Supplier, error)
	List(ctx context.Context, offset, limit int) ([]*model.Supplier, int, error)
	Update(ctx context.Context, supplier *model.Supplier) error
	Delete(ctx context.Context, id string) error
}

// Credential defines credential-related database operations
type Credential interface {
	Create(ctx context.Context, credential *model.AgentSupplierCredential) error
	GetByID(ctx context.Context, id string) (*model.AgentSupplierCredential, error)
	GetByAgentID(ctx context.Context, agentID string) ([]*model.AgentSupplierCredential, error)
	GetAll(ctx context.Context) ([]*model.AgentSupplierCredential, error)
	GetByAgentAndSupplier(ctx context.Context, agentID string, supplierID string) (*model.AgentSupplierCredential, error)
	Update(ctx context.Context, credential *model.AgentSupplierCredential) error
	Delete(ctx context.Context, id string) error
}
