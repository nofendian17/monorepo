// Package repository defines the interfaces for data access layerpackage repository

package repository

import (
	"context"
	"supplier-credentials-service/domain/model"
)

// Supplier interface defines the contract for supplier-related database operations
type Supplier interface {
	// Create adds a new supplier to the database
	Create(ctx context.Context, supplier *model.Supplier) error
	// GetByID retrieves a supplier by their unique identifier
	GetByID(ctx context.Context, id int) (*model.Supplier, error)
	// GetByCode retrieves a supplier by their code
	GetByCode(ctx context.Context, code string) (*model.Supplier, error)
	// List retrieves a paginated list of suppliers
	List(ctx context.Context, offset, limit int) ([]*model.Supplier, int, error)
	// Update modifies an existing supplier
	Update(ctx context.Context, supplier *model.Supplier) error
	// Delete removes a supplier (soft delete)
	Delete(ctx context.Context, id int) error
}

// Credential interface defines the contract for credential-related database operations
type Credential interface {
	// Create adds a new credential to the database
	Create(ctx context.Context, credential *model.AgentSupplierCredential) error
	// GetByID retrieves a credential by their unique identifier
	GetByID(ctx context.Context, id string) (*model.AgentSupplierCredential, error)
	// GetByAgentID retrieves all credentials for an agent
	GetByAgentID(ctx context.Context, agentID string) ([]*model.AgentSupplierCredential, error)
	// GetAll retrieves all credentials
	GetAll(ctx context.Context) ([]*model.AgentSupplierCredential, error)
	// GetByAgentAndSupplier retrieves a credential by agent and supplier
	GetByAgentAndSupplier(ctx context.Context, agentID string, supplierID int) (*model.AgentSupplierCredential, error)
	// Update modifies an existing credential
	Update(ctx context.Context, credential *model.AgentSupplierCredential) error
	// Delete removes a credential (soft delete)
	Delete(ctx context.Context, id string) error
}
