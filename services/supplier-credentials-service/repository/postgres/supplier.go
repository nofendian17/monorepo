// Package postgres provides PostgreSQL implementation for supplier repositorypackage postgres

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

// supplierRepository implements the Supplier repository interface using PostgreSQL
type supplierRepository struct {
	// db is the GORM database instance for database operations
	db *gorm.DB
	// logger is used for logging operations within the repository
	logger logger.LoggerInterface
}

// NewSupplierRepository creates a new instance of supplierRepository
func NewSupplierRepository(db *gorm.DB, logger logger.LoggerInterface) repository.Supplier {
	return &supplierRepository{
		db:     db,
		logger: logger,
	}
}

// Create adds a new supplier to the database
func (r *supplierRepository) Create(ctx context.Context, supplier *model.Supplier) error {
	r.logger.InfoContext(ctx, "Creating supplier", "code", supplier.SupplierCode)
	if err := r.db.WithContext(ctx).Create(supplier).Error; err != nil {
		r.logger.ErrorContext(ctx, "Failed to create supplier", "code", supplier.SupplierCode, "error", err)
		return fmt.Errorf("failed to create supplier: %w", err)
	}
	r.logger.InfoContext(ctx, "Supplier created successfully", "id", supplier.ID, "code", supplier.SupplierCode)
	return nil
}

// GetByID retrieves a supplier by their unique identifier
func (r *supplierRepository) GetByID(ctx context.Context, id string) (*model.Supplier, error) {
	r.logger.InfoContext(ctx, "Getting supplier by ID", "id", id)
	var supplier model.Supplier
	if err := r.db.WithContext(ctx).Where("id = ? AND deleted_at IS NULL", id).First(&supplier).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			r.logger.WarnContext(ctx, "Supplier not found by ID", "id", id)
			return nil, domain.ErrNotFound
		}
		r.logger.ErrorContext(ctx, "Failed to get supplier by ID", "id", id, "error", err)
		return nil, fmt.Errorf("failed to get supplier: %w", err)
	}
	r.logger.InfoContext(ctx, "Supplier retrieved by ID", "id", supplier.ID, "code", supplier.SupplierCode)
	return &supplier, nil
}

// GetByCode retrieves a supplier by their code
func (r *supplierRepository) GetByCode(ctx context.Context, code string) (*model.Supplier, error) {
	r.logger.InfoContext(ctx, "Getting supplier by code", "code", code)
	var supplier model.Supplier
	if err := r.db.WithContext(ctx).Where("supplier_code = ? AND deleted_at IS NULL", code).First(&supplier).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			r.logger.WarnContext(ctx, "Supplier not found by code", "code", code)
			return nil, domain.ErrNotFound
		}
		r.logger.ErrorContext(ctx, "Failed to get supplier by code", "code", code, "error", err)
		return nil, fmt.Errorf("failed to get supplier: %w", err)
	}
	r.logger.InfoContext(ctx, "Supplier retrieved by code", "id", supplier.ID, "code", supplier.SupplierCode)
	return &supplier, nil
}

// List retrieves a paginated list of suppliers
func (r *supplierRepository) List(ctx context.Context, offset, limit int) ([]*model.Supplier, int, error) {
	r.logger.InfoContext(ctx, "Listing suppliers", "offset", offset, "limit", limit)
	var suppliers []*model.Supplier
	var total int64

	// Get total count
	if err := r.db.WithContext(ctx).Model(&model.Supplier{}).Where("deleted_at IS NULL").Count(&total).Error; err != nil {
		r.logger.ErrorContext(ctx, "Failed to count suppliers", "error", err)
		return nil, 0, fmt.Errorf("failed to count suppliers: %w", err)
	}

	// Get paginated suppliers
	if err := r.db.WithContext(ctx).Where("deleted_at IS NULL").Offset(offset).Limit(limit).Order("id ASC").Find(&suppliers).Error; err != nil {
		r.logger.ErrorContext(ctx, "Failed to list suppliers", "offset", offset, "limit", limit, "error", err)
		return nil, 0, fmt.Errorf("failed to list suppliers: %w", err)
	}

	r.logger.InfoContext(ctx, "Suppliers listed successfully", "count", len(suppliers), "offset", offset, "limit", limit, "total", total)
	return suppliers, int(total), nil
}

// Update modifies an existing supplier
func (r *supplierRepository) Update(ctx context.Context, supplier *model.Supplier) error {
	r.logger.InfoContext(ctx, "Updating supplier", "id", supplier.ID, "code", supplier.SupplierCode)
	if err := r.db.WithContext(ctx).Model(&model.Supplier{}).Where("id = ?", supplier.ID).Updates(supplier).Error; err != nil {
		r.logger.ErrorContext(ctx, "Failed to update supplier", "id", supplier.ID, "code", supplier.SupplierCode, "error", err)
		return fmt.Errorf("failed to update supplier: %w", err)
	}
	r.logger.InfoContext(ctx, "Supplier updated successfully", "id", supplier.ID, "code", supplier.SupplierCode)
	return nil
}

// Delete removes a supplier (soft delete)
func (r *supplierRepository) Delete(ctx context.Context, id string) error {
	r.logger.InfoContext(ctx, "Deleting supplier", "id", id)
	supplier := &model.Supplier{ID: id}

	// Use soft delete
	if err := r.db.WithContext(ctx).Delete(supplier).Error; err != nil {
		r.logger.ErrorContext(ctx, "Failed to delete supplier", "id", id, "error", err)
		return fmt.Errorf("failed to delete supplier: %w", err)
	}

	// Check if record was actually deleted
	var count int64
	r.db.WithContext(ctx).Model(&model.Supplier{}).Where("id = ? AND deleted_at IS NULL", id).Count(&count)
	if count > 0 {
		r.logger.WarnContext(ctx, "Supplier not found for deletion", "id", id)
		return domain.ErrNotFound
	}

	r.logger.InfoContext(ctx, "Supplier deleted successfully", "id", id)
	return nil
}
