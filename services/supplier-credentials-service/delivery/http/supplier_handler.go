// Package http contains HTTP delivery implementations for the application
package http

import (
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"strconv"

	"monorepo/contracts/supplier_credentials_service"
	"monorepo/pkg/api"
	"monorepo/pkg/logger"
	"monorepo/pkg/validator"
	"supplier-credentials-service/domain"
	"supplier-credentials-service/domain/model"
	"supplier-credentials-service/usecase"

	"github.com/go-chi/chi/v5"
)

// SupplierHandler handles HTTP requests for supplier operations
type SupplierHandler struct {
	// SupplierUseCase contains business logic for supplier operations
	SupplierUseCase usecase.SupplierUseCase
	// Logger is used for logging operations within the handler
	Logger logger.LoggerInterface
	// API provides standardized API response patterns
	API api.Api
}

// NewSupplierHandler creates a new instance of SupplierHandler
func NewSupplierHandler(supplierUseCase usecase.SupplierUseCase, logger logger.LoggerInterface) *SupplierHandler {
	return &SupplierHandler{
		SupplierUseCase: supplierUseCase,
		Logger:          logger,
		API:             api.New(),
	}
}

// ListSuppliersHandler handles HTTP requests to list suppliers with pagination
func (h *SupplierHandler) ListSuppliersHandler(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	h.Logger.InfoContext(ctx, "List suppliers handler called")

	// Parse query parameters for pagination
	offset, err := strconv.Atoi(r.URL.Query().Get("offset"))
	if err != nil || offset < 0 {
		offset = 0
	}

	limit, err := strconv.Atoi(r.URL.Query().Get("limit"))
	if err != nil || limit <= 0 {
		limit = 10
	}

	if limit > 100 {
		limit = 100
	}

	// Get suppliers and real total from usecase
	suppliers, total, err := h.SupplierUseCase.ListSuppliers(ctx, offset, limit)
	if err != nil {
		h.Logger.ErrorContext(ctx, "Error listing suppliers", "offset", offset, "limit", limit, "error", err)
		h.API.InternalServerError(ctx, w, "Failed to list suppliers")
		return
	}

	if total < 0 {
		total = 0
	}

	// Calculate totalPages (0 if no data, else ceiling division)
	totalPages := 0
	if total > 0 {
		totalPages = (total + limit - 1) / limit
	}

	// Calculate current page (1-based)
	page := 1
	if total > 0 && offset < total {
		page = offset/limit + 1
	} else if total > 0 && offset >= total {
		page = totalPages
	}

	// HasNextPage: true if there are more records after this page
	hasNextPage := false
	if total > 0 && offset+limit < total {
		hasNextPage = true
	}

	// HasPrevPage: true if offset > 0 and there is data
	hasPrevPage := false
	if total > 0 && offset > 0 {
		hasPrevPage = true
	}

	pagination := &api.Pagination{
		Page:        page,
		Limit:       limit,
		Total:       total,
		TotalPages:  totalPages,
		HasNextPage: hasNextPage,
		HasPrevPage: hasPrevPage,
	}

	meta := &api.Meta{
		Pagination: pagination,
	}

	h.Logger.InfoContext(ctx, "Suppliers listed successfully in handler", "count", len(suppliers), "offset", offset, "limit", limit, "total", total)
	h.API.SuccessWithMeta(ctx, w, supplierModelsToResponses(suppliers), meta)
}

// CreateSupplierHandler handles HTTP requests to create a supplier
func (h *SupplierHandler) CreateSupplierHandler(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	h.Logger.InfoContext(ctx, "Create supplier handler called")

	var req supplier_credentials_service.CreateSupplierRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		h.Logger.ErrorContext(ctx, "Invalid request body for supplier creation", "error", err)
		h.API.BadRequest(ctx, w, "Invalid request body")
		return
	}

	// Validate request
	validationErrors := validator.ValidateStruct(req)
	if validationErrors != nil {
		h.Logger.WarnContext(ctx, "Validation failed for supplier creation", "errors", validationErrors)
		h.API.ValidationError(ctx, w, h.convertValidationErrors(validationErrors))
		return
	}

	supplier := &model.Supplier{
		SupplierCode: req.SupplierCode,
		SupplierName: req.SupplierName,
	}

	if err := h.SupplierUseCase.CreateSupplier(ctx, supplier); err != nil {
		h.Logger.ErrorContext(ctx, "Error creating supplier", "code", req.SupplierCode, "error", err)
		h.handleSupplierError(ctx, w, err)
		return
	}

	response := &supplier_credentials_service.SupplierResponse{
		ID:           supplier.ID,
		SupplierCode: supplier.SupplierCode,
		SupplierName: supplier.SupplierName,
	}

	h.Logger.InfoContext(ctx, "Supplier created successfully in handler", "id", supplier.ID, "code", supplier.SupplierCode)
	h.API.Created(ctx, w, response)
}

// UpdateSupplierHandler handles HTTP requests to update a supplier
func (h *SupplierHandler) UpdateSupplierHandler(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	h.Logger.InfoContext(ctx, "Update supplier handler called")

	idStr := chi.URLParam(r, "id")
	if idStr == "" {
		h.Logger.ErrorContext(ctx, "Invalid supplier ID", "id", idStr)
		h.API.BadRequest(ctx, w, "Invalid supplier ID")
		return
	}

	var req supplier_credentials_service.UpdateSupplierRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		h.Logger.ErrorContext(ctx, "Invalid request body for supplier update", "error", err)
		h.API.BadRequest(ctx, w, "Invalid request body")
		return
	}

	// Validate request
	validationErrors := validator.ValidateStruct(req)
	if validationErrors != nil {
		h.Logger.WarnContext(ctx, "Validation failed for supplier update", "errors", validationErrors)
		h.API.ValidationError(ctx, w, h.convertValidationErrors(validationErrors))
		return
	}

	supplier := &model.Supplier{
		ID:           idStr,
		SupplierCode: req.SupplierCode,
		SupplierName: req.SupplierName,
	}

	if err := h.SupplierUseCase.UpdateSupplier(ctx, supplier); err != nil {
		h.Logger.ErrorContext(ctx, "Error updating supplier", "id", idStr, "error", err)
		h.handleSupplierError(ctx, w, err)
		return
	}

	response := &supplier_credentials_service.SupplierResponse{
		ID:           supplier.ID,
		SupplierCode: supplier.SupplierCode,
		SupplierName: supplier.SupplierName,
	}

	h.Logger.InfoContext(ctx, "Supplier updated successfully in handler", "id", supplier.ID, "code", supplier.SupplierCode)
	h.API.Success(ctx, w, response)
}

// DeleteSupplierHandler handles HTTP requests to delete a supplier
func (h *SupplierHandler) DeleteSupplierHandler(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	h.Logger.InfoContext(ctx, "Delete supplier handler called")

	idStr := chi.URLParam(r, "id")
	id, err := strconv.Atoi(idStr)
	if err != nil || id <= 0 {
		h.Logger.ErrorContext(ctx, "Invalid supplier ID", "id", idStr, "error", err)
		h.API.BadRequest(ctx, w, "Invalid supplier ID")
		return
	}

	if err := h.SupplierUseCase.DeleteSupplier(ctx, idStr); err != nil {
		h.Logger.ErrorContext(ctx, "Error deleting supplier", "id", idStr, "error", err)
		h.handleSupplierError(ctx, w, err)
		return
	}

	h.Logger.InfoContext(ctx, "Supplier deleted successfully in handler", "id", idStr)
	h.API.Success(ctx, w, map[string]string{"message": "Supplier deleted successfully"})
}

// handleSupplierError handles supplier-related errors
func (h *SupplierHandler) handleSupplierError(ctx context.Context, w http.ResponseWriter, err error) {
	switch {
	case errors.Is(err, domain.ErrSupplierNotFound):
		h.API.NotFound(ctx, w, err.Error())
	case errors.Is(err, domain.ErrSupplierCodeRequired):
		h.API.BadRequest(ctx, w, err.Error())
	case errors.Is(err, domain.ErrSupplierNameRequired):
		h.API.BadRequest(ctx, w, err.Error())
	case errors.Is(err, domain.ErrSupplierCodeAlreadyExists):
		h.API.Conflict(ctx, w, err.Error())
	default:
		h.API.InternalServerError(ctx, w, "Internal server error")
	}
}

// convertValidationErrors converts validation errors to API format
func (h *SupplierHandler) convertValidationErrors(validationErrors map[string]string) []api.ErrorDetail {
	errorDetails := make([]api.ErrorDetail, 0, len(validationErrors))
	for field, message := range validationErrors {
		errorDetails = append(errorDetails, api.ErrorDetail{
			Field:   field,
			Message: message,
		})
	}
	return errorDetails
}

// supplierModelsToResponses converts supplier models to response format
func supplierModelsToResponses(suppliers []*model.Supplier) []*supplier_credentials_service.SupplierResponse {
	responses := make([]*supplier_credentials_service.SupplierResponse, len(suppliers))
	for i, supplier := range suppliers {
		responses[i] = &supplier_credentials_service.SupplierResponse{
			ID:           supplier.ID,
			SupplierCode: supplier.SupplierCode,
			SupplierName: supplier.SupplierName,
		}
	}
	return responses
}
