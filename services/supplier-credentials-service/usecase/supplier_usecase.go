// Package usecase contains business logic for supplier operations
package usecase

import (
	"context"
	"errors"
	"fmt"

	"monorepo/pkg/logger"
	"supplier-credentials-service/domain"
	"supplier-credentials-service/domain/model"
	"supplier-credentials-service/domain/repository"
)

// SupplierUseCase defines the interface for supplier-related business operations
type SupplierUseCase interface {
	// CreateSupplier adds a new supplier
	CreateSupplier(ctx context.Context, supplier *model.Supplier) error
	// UpdateSupplier modifies an existing supplier
	UpdateSupplier(ctx context.Context, supplier *model.Supplier) error
	// DeleteSupplier removes a supplier
	DeleteSupplier(ctx context.Context, id int) error
	// ListSuppliers retrieves a paginated list of suppliers
	ListSuppliers(ctx context.Context, offset, limit int) ([]*model.Supplier, int, error)
	// GetSupplierByID retrieves a supplier by ID
	GetSupplierByID(ctx context.Context, id int) (*model.Supplier, error)
}

// supplierUseCase implements the SupplierUseCase interface
type supplierUseCase struct {
	// supplierRepo is the repository interface for supplier database operations
	supplierRepo repository.Supplier
	// logger is used for logging operations within the usecase
	logger logger.LoggerInterface
}

// NewSupplierUseCase creates a new instance of supplierUseCase
func NewSupplierUseCase(supplierRepo repository.Supplier, appLogger logger.LoggerInterface) SupplierUseCase {
	return &supplierUseCase{
		supplierRepo: supplierRepo,
		logger:       appLogger,
	}
}

// CreateSupplier adds a new supplier
func (uc *supplierUseCase) CreateSupplier(ctx context.Context, supplier *model.Supplier) error {
	uc.logger.InfoContext(ctx, "Creating supplier in usecase", "code", supplier.SupplierCode, "name", supplier.SupplierName)

	// Business logic validation
	if supplier.SupplierCode == "" {
		uc.logger.WarnContext(ctx, "Supplier code is required for supplier creation")
		return domain.ErrSupplierCodeRequired
	}

	if supplier.SupplierName == "" {
		uc.logger.WarnContext(ctx, "Supplier name is required for supplier creation")
		return domain.ErrSupplierNameRequired
	}

	// Check if supplier code already exists
	existing, err := uc.supplierRepo.GetByCode(ctx, supplier.SupplierCode)
	if err != nil && !errors.Is(err, domain.ErrNotFound) {
		uc.logger.ErrorContext(ctx, "Error checking existing supplier", "code", supplier.SupplierCode, "error", err)
		return fmt.Errorf("error checking existing supplier: %w", err)
	}
	if existing != nil {
		uc.logger.WarnContext(ctx, "Supplier with this code already exists", "code", supplier.SupplierCode)
		return domain.ErrSupplierCodeAlreadyExists
	}

	if err := uc.supplierRepo.Create(ctx, supplier); err != nil {
		uc.logger.ErrorContext(ctx, "Failed to create supplier in repository", "code", supplier.SupplierCode, "error", err)
		return err
	}

	uc.logger.InfoContext(ctx, "Supplier created successfully in usecase", "id", supplier.ID, "code", supplier.SupplierCode)
	return nil
}

// UpdateSupplier modifies an existing supplier
func (uc *supplierUseCase) UpdateSupplier(ctx context.Context, supplier *model.Supplier) error {
	uc.logger.InfoContext(ctx, "Updating supplier in usecase", "id", supplier.ID, "code", supplier.SupplierCode, "name", supplier.SupplierName)

	// Business logic validation
	if supplier.ID == 0 {
		uc.logger.WarnContext(ctx, "Supplier ID is required for supplier update")
		return domain.ErrSupplierIDRequired
	}

	if supplier.SupplierCode == "" {
		uc.logger.WarnContext(ctx, "Supplier code is required for supplier update")
		return domain.ErrSupplierCodeRequired
	}

	if supplier.SupplierName == "" {
		uc.logger.WarnContext(ctx, "Supplier name is required for supplier update")
		return domain.ErrSupplierNameRequired
	}

	// Check if supplier exists
	existing, err := uc.supplierRepo.GetByID(ctx, supplier.ID)
	if err != nil {
		if errors.Is(err, domain.ErrNotFound) {
			uc.logger.WarnContext(ctx, "Supplier not found for update", "id", supplier.ID)
			return domain.ErrSupplierNotFound
		}
		uc.logger.ErrorContext(ctx, "Error checking existing supplier", "id", supplier.ID, "error", err)
		return fmt.Errorf("error checking existing supplier: %w", err)
	}

	// Check if supplier code is being changed and if it conflicts with another supplier
	if existing.SupplierCode != supplier.SupplierCode {
		codeExists, err := uc.supplierRepo.GetByCode(ctx, supplier.SupplierCode)
		if err != nil && !errors.Is(err, domain.ErrNotFound) {
			uc.logger.ErrorContext(ctx, "Error checking supplier code conflict", "code", supplier.SupplierCode, "error", err)
			return fmt.Errorf("error checking supplier code conflict: %w", err)
		}
		if codeExists != nil && codeExists.ID != supplier.ID {
			uc.logger.WarnContext(ctx, "Supplier code already exists for another supplier", "code", supplier.SupplierCode)
			return domain.ErrSupplierCodeAlreadyExists
		}
	}

	if err := uc.supplierRepo.Update(ctx, supplier); err != nil {
		uc.logger.ErrorContext(ctx, "Failed to update supplier in repository", "id", supplier.ID, "error", err)
		return err
	}

	uc.logger.InfoContext(ctx, "Supplier updated successfully in usecase", "id", supplier.ID, "code", supplier.SupplierCode)
	return nil
}

// ListSuppliers retrieves a paginated list of suppliers
func (uc *supplierUseCase) ListSuppliers(ctx context.Context, offset, limit int) ([]*model.Supplier, int, error) {
	uc.logger.InfoContext(ctx, "Listing suppliers in usecase", "offset", offset, "limit", limit)

	suppliers, total, err := uc.supplierRepo.List(ctx, offset, limit)
	if err != nil {
		uc.logger.ErrorContext(ctx, "Failed to list suppliers in repository", "offset", offset, "limit", limit, "error", err)
		return nil, 0, err
	}

	uc.logger.InfoContext(ctx, "Suppliers listed successfully in usecase", "count", len(suppliers), "offset", offset, "limit", limit, "total", total)
	return suppliers, total, nil
}

// GetSupplierByID retrieves a supplier by ID
func (uc *supplierUseCase) GetSupplierByID(ctx context.Context, id int) (*model.Supplier, error) {
	uc.logger.InfoContext(ctx, "Getting supplier by ID in usecase", "id", id)

	supplier, err := uc.supplierRepo.GetByID(ctx, id)
	if err != nil {
		if errors.Is(err, domain.ErrNotFound) {
			uc.logger.WarnContext(ctx, "Supplier not found", "id", id)
			return nil, domain.ErrSupplierNotFound
		}
		uc.logger.ErrorContext(ctx, "Error getting supplier by ID", "id", id, "error", err)
		return nil, fmt.Errorf("error getting supplier: %w", err)
	}

	uc.logger.InfoContext(ctx, "Supplier retrieved by ID in usecase", "id", supplier.ID)
	return supplier, nil
}

// DeleteSupplier removes a supplier
func (uc *supplierUseCase) DeleteSupplier(ctx context.Context, id int) error {
	uc.logger.InfoContext(ctx, "Deleting supplier in usecase", "id", id)

	// Check if supplier exists first
	_, err := uc.supplierRepo.GetByID(ctx, id)
	if err != nil {
		if errors.Is(err, domain.ErrNotFound) {
			uc.logger.WarnContext(ctx, "Supplier not found for deletion", "id", id)
			return domain.ErrSupplierNotFound
		}
		uc.logger.ErrorContext(ctx, "Error checking supplier existence before deletion", "id", id, "error", err)
		return fmt.Errorf("error checking supplier existence: %w", err)
	}

	// Delete the supplier
	err = uc.supplierRepo.Delete(ctx, id)
	if err != nil {
		uc.logger.ErrorContext(ctx, "Error deleting supplier", "id", id, "error", err)
		return fmt.Errorf("error deleting supplier: %w", err)
	}

	uc.logger.InfoContext(ctx, "Supplier deleted successfully in usecase", "id", id)
	return nil
}
