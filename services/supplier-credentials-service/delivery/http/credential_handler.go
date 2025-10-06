// Package http contains HTTP delivery implementations for the application
package http

import (
	"context"
	"encoding/json"
	"net/http"

	"monorepo/contracts/supplier_credentials_service"
	"monorepo/pkg/api"
	"monorepo/pkg/logger"
	"monorepo/pkg/validator"
	"supplier-credentials-service/domain"
	"supplier-credentials-service/domain/model"
	"supplier-credentials-service/usecase"

	"github.com/go-chi/chi/v5"
)

// CredentialHandler handles HTTP requests for credential operations
type CredentialHandler struct {
	// CredentialUseCase contains business logic for credential operations
	CredentialUseCase usecase.CredentialUseCase
	// Logger is used for logging operations within the handler
	Logger logger.LoggerInterface
	// API provides standardized API response patterns
	API api.Api
}

// NewCredentialHandler creates a new instance of CredentialHandler
func NewCredentialHandler(credentialUseCase usecase.CredentialUseCase, logger logger.LoggerInterface) *CredentialHandler {
	return &CredentialHandler{
		CredentialUseCase: credentialUseCase,
		Logger:            logger,
		API:               api.New(),
	}
}

// CreateHandler handles HTTP requests to create a new credential
func (h *CredentialHandler) CreateHandler(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	h.Logger.InfoContext(ctx, "Create credential handler called")

	var req supplier_credentials_service.CreateCredentialRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		h.Logger.ErrorContext(ctx, "Invalid request body for credential creation", "error", err)
		h.API.BadRequest(ctx, w, "Invalid request body")
		return
	}

	req.IataAgentID = ctx.Value("agent_iata_id").(string) // Get IATA agent ID from context (set by middleware)

	// Validate the request
	validationErrors := validator.ValidateStruct(&req)
	if validationErrors != nil {
		h.Logger.WarnContext(ctx, "Validation failed for credential creation", "errors", validationErrors)
		h.API.ValidationError(ctx, w, h.convertValidationErrors(validationErrors))
		return
	}

	// Convert to model
	credential := &model.AgentSupplierCredential{
		IataAgentID: req.IataAgentID,
		SupplierID:  req.SupplierID,
		Credentials: req.Credentials,
	}

	if err := h.CredentialUseCase.CreateCredential(ctx, credential); err != nil {
		h.handleCredentialError(ctx, w, err)
		return
	}

	h.Logger.InfoContext(ctx, "Credential created successfully in handler", "id", credential.ID)
	h.API.Created(ctx, w, h.credentialToResponse(credential))
}

// ListHandler handles HTTP requests to list credentials for the authenticated agent
func (h *CredentialHandler) ListHandler(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	h.Logger.InfoContext(ctx, "List credentials handler called")

	var req supplier_credentials_service.ListCredentialsRequest
	req.IataAgentID = ctx.Value("agent_iata_id").(string) // Get IATA agent ID from context (set by middleware)

	// Validate the request
	validationErrors := validator.ValidateStruct(&req)
	if validationErrors != nil {
		h.Logger.WarnContext(ctx, "Validation failed for list credentials", "errors", validationErrors)
		h.API.ValidationError(ctx, w, h.convertValidationErrors(validationErrors))
		return
	}

	credentials, err := h.CredentialUseCase.GetCredentialsByAgentID(ctx, req.IataAgentID)
	if err != nil {
		h.handleCredentialError(ctx, w, err)
		return
	}

	response := make([]*supplier_credentials_service.CredentialResponse, len(credentials))
	for i, cred := range credentials {
		response[i] = h.credentialToResponse(cred)
	}

	h.Logger.InfoContext(ctx, "Credentials listed successfully", "count", len(response))
	h.API.Success(ctx, w, response)
}

// GetByIDHandler handles HTTP requests to retrieve a credential by ID
func (h *CredentialHandler) GetByIDHandler(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	h.Logger.InfoContext(ctx, "Get credential by ID handler called")

	req := supplier_credentials_service.GetCredentialByIDRequest{ID: chi.URLParam(r, "id")}
	if err := validator.ValidateStruct(&req); err != nil {
		h.Logger.WarnContext(ctx, "Validation failed for get credential by ID", "errors", err)
		h.API.ValidationError(ctx, w, h.convertValidationErrors(err))
		return
	}

	credential, err := h.CredentialUseCase.GetCredentialByID(ctx, req.ID)
	if err != nil {
		h.handleCredentialError(ctx, w, err)
		return
	}

	h.Logger.InfoContext(ctx, "Credential retrieved by ID", "id", credential.ID)
	h.API.Success(ctx, w, h.credentialToResponse(credential))
}

// UpdateHandler handles HTTP requests to update an existing credential
func (h *CredentialHandler) UpdateHandler(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	h.Logger.InfoContext(ctx, "Update credential handler called")

	var req supplier_credentials_service.UpdateCredentialRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		h.Logger.ErrorContext(ctx, "Invalid request body for credential update", "error", err)
		h.API.BadRequest(ctx, w, "Invalid request body")
		return
	}

	// Set ID from URL parameter
	req.ID = chi.URLParam(r, "id")

	// Validate the request
	validationErrors := validator.ValidateStruct(&req)
	if validationErrors != nil {
		h.Logger.WarnContext(ctx, "Validation failed for credential update", "errors", validationErrors)
		h.API.ValidationError(ctx, w, h.convertValidationErrors(validationErrors))
		return
	}

	credential := &model.AgentSupplierCredential{
		ID:          req.ID,
		Credentials: req.Credentials,
	}

	if err := h.CredentialUseCase.UpdateCredential(ctx, credential); err != nil {
		h.handleCredentialError(ctx, w, err)
		return
	}

	h.Logger.InfoContext(ctx, "Credential updated successfully", "id", req.ID)
	h.API.Success(ctx, w, h.credentialToResponse(credential))
}

// DeleteHandler handles HTTP requests to delete a credential
func (h *CredentialHandler) DeleteHandler(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	h.Logger.InfoContext(ctx, "Delete credential handler called")

	req := supplier_credentials_service.DeleteCredentialRequest{ID: chi.URLParam(r, "id")}
	if err := validator.ValidateStruct(&req); err != nil {
		h.Logger.WarnContext(ctx, "Validation failed for delete credential", "errors", err)
		h.API.ValidationError(ctx, w, h.convertValidationErrors(err))
		return
	}

	if err := h.CredentialUseCase.DeleteCredential(ctx, req.ID); err != nil {
		h.handleCredentialError(ctx, w, err)
		return
	}

	h.Logger.InfoContext(ctx, "Credential deleted successfully", "id", req.ID)
	h.API.Success(ctx, w, map[string]string{"message": "Credential deleted successfully"})
}

// InternalListHandler handles internal requests to list credentials
func (h *CredentialHandler) InternalListHandler(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	h.Logger.InfoContext(ctx, "Internal list credentials handler called")

	credentials, err := h.CredentialUseCase.GetAllCredentials(ctx)
	if err != nil {
		h.handleCredentialError(ctx, w, err)
		return
	}

	response := make([]*supplier_credentials_service.CredentialResponse, len(credentials))
	for i, cred := range credentials {
		response[i] = h.credentialToResponse(cred)
	}

	h.Logger.InfoContext(ctx, "Credentials listed for internal use", "count", len(response))
	h.API.Success(ctx, w, response)
}

// handleCredentialError handles credential-related errors consistently
func (h *CredentialHandler) handleCredentialError(ctx context.Context, w http.ResponseWriter, err error) {
	switch {
	case err.Error() == domain.ErrCredentialNotFound.Message:
		h.API.NotFound(ctx, w, err.Error())
	case err.Error() == domain.ErrSupplierNotFound.Message:
		h.API.NotFound(ctx, w, err.Error())
	case err.Error() == domain.ErrInvalidID.Message:
		h.API.BadRequest(ctx, w, err.Error())
	case err.Error() == domain.ErrIataAgentIDRequired.Message:
		h.API.BadRequest(ctx, w, err.Error())
	case err.Error() == domain.ErrSupplierIDRequired.Message:
		h.API.BadRequest(ctx, w, err.Error())
	case err.Error() == domain.ErrCredentialsRequired.Message:
		h.API.BadRequest(ctx, w, err.Error())
	case err.Error() == domain.ErrCredentialAlreadyExists.Message:
		h.API.BadRequest(ctx, w, err.Error())
	default:
		h.API.InternalServerError(ctx, w, "Internal server error")
	}
}

// convertValidationErrors converts validator errors to API error details
func (h *CredentialHandler) convertValidationErrors(validationErrors map[string]string) []api.ErrorDetail {
	errorDetails := make([]api.ErrorDetail, 0, len(validationErrors))
	for field, message := range validationErrors {
		errorDetails = append(errorDetails, api.ErrorDetail{
			Field:   field,
			Message: message,
		})
	}
	return errorDetails
}

// credentialToResponse converts a model to response
func (h *CredentialHandler) credentialToResponse(cred *model.AgentSupplierCredential) *supplier_credentials_service.CredentialResponse {
	response := &supplier_credentials_service.CredentialResponse{
		ID:          cred.ID,
		IataAgentID: cred.IataAgentID,
		SupplierID:  cred.SupplierID,
		Credentials: cred.Credentials,
		CreatedAt:   cred.CreatedAt.Format("2006-01-02T15:04:05Z07:00"),
		UpdatedAt:   cred.UpdatedAt.Format("2006-01-02T15:04:05Z07:00"),
	}
	if cred.Supplier.ID != 0 {
		response.Supplier = &supplier_credentials_service.SupplierResponse{
			ID:           cred.Supplier.ID,
			SupplierCode: cred.Supplier.SupplierCode,
			SupplierName: cred.Supplier.SupplierName,
		}
	}
	return response
}
