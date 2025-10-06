// Package postgres provides PostgreSQL implementation for credential repository
package postgres

import (
	"context"
	"fmt"

	"monorepo/pkg/logger"
	"supplier-credentials-service/domain"
	"supplier-credentials-service/domain/model"
	"supplier-credentials-service/domain/repository"

	"gorm.io/gorm"
)

// credentialRepository implements the Credential repository interface using PostgreSQL
type credentialRepository struct {
	// db is the GORM database instance for database operations
	db *gorm.DB
	// logger is used for logging operations within the repository
	logger logger.LoggerInterface
}

// NewCredentialRepository creates a new instance of credentialRepository
func NewCredentialRepository(db *gorm.DB, logger logger.LoggerInterface) repository.Credential {
	return &credentialRepository{
		db:     db,
		logger: logger,
	}
}

// Create adds a new credential to the database
func (r *credentialRepository) Create(ctx context.Context, credential *model.AgentSupplierCredential) error {
	r.logger.InfoContext(ctx, "Creating credential", "agentID", credential.IataAgentID, "supplierID", credential.SupplierID)
	if err := r.db.WithContext(ctx).Create(credential).Error; err != nil {
		r.logger.ErrorContext(ctx, "Failed to create credential", "agentID", credential.IataAgentID, "supplierID", credential.SupplierID, "error", err)
		return fmt.Errorf("failed to create credential: %w", err)
	}
	r.logger.InfoContext(ctx, "Credential created successfully", "id", credential.ID, "agentID", credential.IataAgentID, "supplierID", credential.SupplierID)
	return nil
}

// GetByID retrieves a credential by their unique identifier
func (r *credentialRepository) GetByID(ctx context.Context, id string) (*model.AgentSupplierCredential, error) {
	r.logger.InfoContext(ctx, "Getting credential by ID", "id", id)
	var credential model.AgentSupplierCredential
	if err := r.db.WithContext(ctx).Preload("Supplier").Where("id = ? AND deleted_at IS NULL", id).First(&credential).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			r.logger.WarnContext(ctx, "Credential not found by ID", "id", id)
			return nil, domain.ErrNotFound
		}
		r.logger.ErrorContext(ctx, "Failed to get credential by ID", "id", id, "error", err)
		return nil, fmt.Errorf("failed to get credential: %w", err)
	}
	r.logger.InfoContext(ctx, "Credential retrieved by ID", "id", credential.ID, "agentID", credential.IataAgentID)
	return &credential, nil
}

// GetByAgentID retrieves all credentials for an agent
func (r *credentialRepository) GetByAgentID(ctx context.Context, agentID string) ([]*model.AgentSupplierCredential, error) {
	r.logger.InfoContext(ctx, "Getting credentials by agent ID", "agentID", agentID)
	var credentials []*model.AgentSupplierCredential
	if err := r.db.WithContext(ctx).Preload("Supplier").Where("iata_agent_id = ? AND deleted_at IS NULL", agentID).Find(&credentials).Error; err != nil {
		r.logger.ErrorContext(ctx, "Failed to get credentials by agent ID", "agentID", agentID, "error", err)
		return nil, fmt.Errorf("failed to get credentials by agent ID: %w", err)
	}
	r.logger.InfoContext(ctx, "Credentials retrieved by agent ID", "count", len(credentials), "agentID", agentID)
	return credentials, nil
}

// GetAll retrieves all credentials
func (r *credentialRepository) GetAll(ctx context.Context) ([]*model.AgentSupplierCredential, error) {
	r.logger.InfoContext(ctx, "Getting all credentials")
	var credentials []*model.AgentSupplierCredential
	if err := r.db.WithContext(ctx).Preload("Supplier").Where("deleted_at IS NULL").Find(&credentials).Error; err != nil {
		r.logger.ErrorContext(ctx, "Failed to get all credentials", "error", err)
		return nil, fmt.Errorf("failed to get all credentials: %w", err)
	}
	r.logger.InfoContext(ctx, "All credentials retrieved", "count", len(credentials))
	return credentials, nil
}

// GetByAgentAndSupplier retrieves a credential by agent and supplier
func (r *credentialRepository) GetByAgentAndSupplier(ctx context.Context, agentID string, supplierID int) (*model.AgentSupplierCredential, error) {
	r.logger.InfoContext(ctx, "Getting credential by agent and supplier", "agentID", agentID, "supplierID", supplierID)
	var credential model.AgentSupplierCredential
	if err := r.db.WithContext(ctx).Preload("Supplier").Where("iata_agent_id = ? AND supplier_id = ? AND deleted_at IS NULL", agentID, supplierID).First(&credential).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			r.logger.WarnContext(ctx, "Credential not found by agent and supplier", "agentID", agentID, "supplierID", supplierID)
			return nil, domain.ErrNotFound
		}
		r.logger.ErrorContext(ctx, "Failed to get credential by agent and supplier", "agentID", agentID, "supplierID", supplierID, "error", err)
		return nil, fmt.Errorf("failed to get credential: %w", err)
	}
	r.logger.InfoContext(ctx, "Credential retrieved by agent and supplier", "id", credential.ID, "agentID", agentID, "supplierID", supplierID)
	return &credential, nil
}

// Update modifies an existing credential
func (r *credentialRepository) Update(ctx context.Context, credential *model.AgentSupplierCredential) error {
	r.logger.InfoContext(ctx, "Updating credential", "id", credential.ID, "agentID", credential.IataAgentID)
	if err := r.db.WithContext(ctx).Model(&model.AgentSupplierCredential{}).Where("id = ?", credential.ID).Updates(credential).Error; err != nil {
		r.logger.ErrorContext(ctx, "Failed to update credential", "id", credential.ID, "agentID", credential.IataAgentID, "error", err)
		return fmt.Errorf("failed to update credential: %w", err)
	}
	r.logger.InfoContext(ctx, "Credential updated successfully", "id", credential.ID, "agentID", credential.IataAgentID)
	return nil
}

// Delete removes a credential (soft delete)
func (r *credentialRepository) Delete(ctx context.Context, id string) error {
	r.logger.InfoContext(ctx, "Deleting credential", "id", id)
	credential := &model.AgentSupplierCredential{ID: id}

	// Use soft delete
	if err := r.db.WithContext(ctx).Delete(credential).Error; err != nil {
		r.logger.ErrorContext(ctx, "Failed to delete credential", "id", id, "error", err)
		return fmt.Errorf("failed to delete credential: %w", err)
	}

	// Check if record was actually deleted
	var count int64
	r.db.WithContext(ctx).Model(&model.AgentSupplierCredential{}).Where("id = ? AND deleted_at IS NULL", id).Count(&count)
	if count > 0 {
		r.logger.WarnContext(ctx, "Credential not found for deletion", "id", id)
		return domain.ErrNotFound
	}

	r.logger.InfoContext(ctx, "Credential deleted successfully", "id", id)
	return nil
}
