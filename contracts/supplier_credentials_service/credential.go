// Package supplier_credentials_service contains request and response contracts for the supplier-credentials-service
package supplier_credentials_service

// CreateCredentialRequest represents the request payload for creating a credential
type CreateCredentialRequest struct {
	IataAgentID string `json:"iata_agent_id" validate:"required,ulid"`
	SupplierID  int    `json:"supplier_id" validate:"required,min=1"`
	Credentials string `json:"credentials" validate:"required"`
}

// ListCredentialsRequest represents the request for listing credentials
type ListCredentialsRequest struct {
	IataAgentID string `validate:"required,ulid"`
}

// UpdateCredentialRequest represents the request payload for updating a credential
type UpdateCredentialRequest struct {
	ID          string `json:"id" validate:"required,ulid"`
	Credentials string `json:"credentials" validate:"required"`
}

// GetCredentialByIDRequest represents the request for getting a credential by ID
type GetCredentialByIDRequest struct {
	ID string `validate:"required,ulid"`
}

// DeleteCredentialRequest represents the request for deleting a credential
type DeleteCredentialRequest struct {
	ID string `validate:"required,ulid"`
}

// CredentialResponse represents the response payload for a credential
type CredentialResponse struct {
	ID          string            `json:"id"`
	IataAgentID string            `json:"iata_agent_id"`
	SupplierID  int               `json:"supplier_id"`
	Supplier    *SupplierResponse `json:"supplier,omitempty"`
	Credentials string            `json:"credentials"`
	CreatedAt   string            `json:"created_at"`
	UpdatedAt   string            `json:"updated_at"`
}

// SupplierResponse represents the response payload for a supplier
type SupplierResponse struct {
	ID           int    `json:"id"`
	SupplierCode string `json:"supplier_code"`
	SupplierName string `json:"supplier_name"`
}
