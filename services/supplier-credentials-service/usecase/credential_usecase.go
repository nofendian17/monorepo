// Package usecase contains business logic for credential operations
package usecase

import (
	"context"
	"crypto/aes"
	"crypto/cipher"
	"crypto/rand"
	"encoding/base64"
	"errors"
	"fmt"

	"monorepo/pkg/logger"
	"supplier-credentials-service/domain"
	"supplier-credentials-service/domain/model"
	"supplier-credentials-service/domain/repository"
)

// CredentialUseCase defines the interface for credential-related business operations
type CredentialUseCase interface {
	// CreateCredential adds a new supplier credential for an agent
	CreateCredential(ctx context.Context, credential *model.AgentSupplierCredential) error
	// GetCredentialByID retrieves a credential by its ID
	GetCredentialByID(ctx context.Context, id string) (*model.AgentSupplierCredential, error)
	// GetCredentialsByAgentID retrieves all credentials for an agent
	GetCredentialsByAgentID(ctx context.Context, agentID string) ([]*model.AgentSupplierCredential, error)
	// GetAllCredentials retrieves all credentials
	GetAllCredentials(ctx context.Context) ([]*model.AgentSupplierCredential, error)
	// UpdateCredential modifies an existing credential
	UpdateCredential(ctx context.Context, credential *model.AgentSupplierCredential) error
	// DeleteCredential removes a credential
	DeleteCredential(ctx context.Context, id string) error
}

// credentialUseCase implements the CredentialUseCase interface
type credentialUseCase struct {
	// credentialRepo is the repository interface for credential database operations
	credentialRepo repository.Credential
	// supplierUseCase is used to validate supplier existence
	supplierUseCase SupplierUseCase
	// logger is used for logging operations within the usecase
	logger logger.LoggerInterface
	// encryptionKey is the key used for encrypting/decrypting credentials
	encryptionKey string
}

// NewCredentialUseCase creates a new instance of credentialUseCase
func NewCredentialUseCase(credentialRepo repository.Credential, supplierUseCase SupplierUseCase, appLogger logger.LoggerInterface, encryptionKey string) CredentialUseCase {
	return &credentialUseCase{
		credentialRepo:  credentialRepo,
		supplierUseCase: supplierUseCase,
		logger:          appLogger,
		encryptionKey:   encryptionKey,
	}
}

// encrypt encrypts the given plaintext using AES-GCM
func (uc *credentialUseCase) encrypt(plaintext string) (string, error) {
	if uc.encryptionKey == "" {
		return "", errors.New("encryption key not set")
	}

	key := []byte(uc.encryptionKey)
	if len(key) != 32 {
		return "", errors.New("encryption key must be 32 bytes")
	}

	block, err := aes.NewCipher(key)
	if err != nil {
		return "", err
	}

	gcm, err := cipher.NewGCM(block)
	if err != nil {
		return "", err
	}

	nonce := make([]byte, gcm.NonceSize())
	if _, err := rand.Read(nonce); err != nil {
		return "", err
	}

	ciphertext := gcm.Seal(nonce, nonce, []byte(plaintext), nil)
	return base64.StdEncoding.EncodeToString(ciphertext), nil
}

// decrypt decrypts the given ciphertext using AES-GCM
func (uc *credentialUseCase) decrypt(ciphertext string) (string, error) {
	if uc.encryptionKey == "" {
		return "", errors.New("encryption key not set")
	}

	key := []byte(uc.encryptionKey)
	if len(key) != 32 {
		return "", errors.New("encryption key must be 32 bytes")
	}

	data, err := base64.StdEncoding.DecodeString(ciphertext)
	if err != nil {
		return "", err
	}

	block, err := aes.NewCipher(key)
	if err != nil {
		return "", err
	}

	gcm, err := cipher.NewGCM(block)
	if err != nil {
		return "", err
	}

	nonceSize := gcm.NonceSize()
	if len(data) < nonceSize {
		return "", errors.New("ciphertext too short")
	}

	nonce, ciphertextBytes := data[:nonceSize], data[nonceSize:]
	plaintext, err := gcm.Open(nil, nonce, ciphertextBytes, nil)
	if err != nil {
		return "", err
	}

	return string(plaintext), nil
}

// CreateCredential adds a new supplier credential for an agent
func (uc *credentialUseCase) CreateCredential(ctx context.Context, credential *model.AgentSupplierCredential) error {
	uc.logger.InfoContext(ctx, "Creating credential in usecase", "agentID", credential.IataAgentID, "supplierID", credential.SupplierID)

	// Business logic validation
	if credential.IataAgentID == "" {
		uc.logger.WarnContext(ctx, "IATA agent ID is required for credential creation")
		return domain.ErrIataAgentIDRequired
	}

	if credential.SupplierID == 0 {
		uc.logger.WarnContext(ctx, "Supplier ID is required for credential creation")
		return domain.ErrSupplierIDRequired
	}

	if credential.Credentials == "" {
		uc.logger.WarnContext(ctx, "Credentials are required for credential creation")
		return domain.ErrCredentialsRequired
	}

	// Check if supplier exists
	_, err := uc.supplierUseCase.GetSupplierByID(ctx, credential.SupplierID)
	if err != nil {
		if errors.Is(err, domain.ErrSupplierNotFound) {
			uc.logger.WarnContext(ctx, "Supplier not found", "supplierID", credential.SupplierID)
			return domain.ErrSupplierNotFound
		}
		uc.logger.ErrorContext(ctx, "Error checking supplier", "supplierID", credential.SupplierID, "error", err)
		return fmt.Errorf("error checking supplier: %w", err)
	}

	// Check if credential already exists for this agent-supplier pair
	existing, err := uc.credentialRepo.GetByAgentAndSupplier(ctx, credential.IataAgentID, credential.SupplierID)
	if err != nil && !errors.Is(err, domain.ErrNotFound) {
		uc.logger.ErrorContext(ctx, "Error checking existing credential", "agentID", credential.IataAgentID, "supplierID", credential.SupplierID, "error", err)
		return fmt.Errorf("error checking existing credential: %w", err)
	}
	if existing != nil {
		uc.logger.WarnContext(ctx, "Credential already exists for this agent-supplier pair", "agentID", credential.IataAgentID, "supplierID", credential.SupplierID)
		return domain.ErrCredentialAlreadyExists
	}

	// Encrypt credentials
	encryptedCredentials, err := uc.encrypt(credential.Credentials)
	if err != nil {
		uc.logger.ErrorContext(ctx, "Failed to encrypt credentials", "error", err)
		return fmt.Errorf("failed to encrypt credentials: %w", err)
	}
	credential.Credentials = encryptedCredentials

	if err := uc.credentialRepo.Create(ctx, credential); err != nil {
		uc.logger.ErrorContext(ctx, "Failed to create credential in repository", "agentID", credential.IataAgentID, "supplierID", credential.SupplierID, "error", err)
		return err
	}

	uc.logger.InfoContext(ctx, "Credential created successfully in usecase", "id", credential.ID, "agentID", credential.IataAgentID, "supplierID", credential.SupplierID)
	return nil
}

// GetCredentialByID retrieves a credential by its ID
func (uc *credentialUseCase) GetCredentialByID(ctx context.Context, id string) (*model.AgentSupplierCredential, error) {
	uc.logger.InfoContext(ctx, "Getting credential by ID in usecase", "id", id)
	if id == "" {
		uc.logger.WarnContext(ctx, "Invalid credential ID provided", "id", id)
		return nil, domain.ErrInvalidID
	}

	credential, err := uc.credentialRepo.GetByID(ctx, id)
	if err != nil {
		if errors.Is(err, domain.ErrNotFound) {
			uc.logger.WarnContext(ctx, "Credential not found", "id", id)
			return nil, domain.ErrCredentialNotFound
		}
		uc.logger.ErrorContext(ctx, "Error getting credential by ID", "id", id, "error", err)
		return nil, fmt.Errorf("error getting credential: %w", err)
	}

	// Decrypt credentials
	decryptedCredentials, err := uc.decrypt(credential.Credentials)
	if err != nil {
		uc.logger.ErrorContext(ctx, "Failed to decrypt credentials", "id", id, "error", err)
		return nil, fmt.Errorf("failed to decrypt credentials: %w", err)
	}
	credential.Credentials = decryptedCredentials

	uc.logger.InfoContext(ctx, "Credential retrieved by ID in usecase", "id", credential.ID, "agentID", credential.IataAgentID)
	return credential, nil
}

// GetCredentialsByAgentID retrieves all credentials for an agent
func (uc *credentialUseCase) GetCredentialsByAgentID(ctx context.Context, agentID string) ([]*model.AgentSupplierCredential, error) {
	uc.logger.InfoContext(ctx, "Getting credentials by agent ID in usecase", "agentID", agentID)
	if agentID == "" {
		uc.logger.WarnContext(ctx, "Invalid agent ID provided", "agentID", agentID)
		return nil, domain.ErrInvalidID
	}

	credentials, err := uc.credentialRepo.GetByAgentID(ctx, agentID)
	if err != nil {
		uc.logger.ErrorContext(ctx, "Error getting credentials by agent ID", "agentID", agentID, "error", err)
		return nil, fmt.Errorf("error getting credentials: %w", err)
	}

	// Decrypt credentials for each
	for _, cred := range credentials {
		decrypted, err := uc.decrypt(cred.Credentials)
		if err != nil {
			uc.logger.ErrorContext(ctx, "Failed to decrypt credentials", "id", cred.ID, "error", err)
			return nil, fmt.Errorf("failed to decrypt credentials for id %s: %w", cred.ID, err)
		}
		cred.Credentials = decrypted
	}

	uc.logger.InfoContext(ctx, "Credentials retrieved by agent ID in usecase", "count", len(credentials), "agentID", agentID)
	return credentials, nil
}

// GetAllCredentials retrieves all credentials
func (uc *credentialUseCase) GetAllCredentials(ctx context.Context) ([]*model.AgentSupplierCredential, error) {
	uc.logger.InfoContext(ctx, "Getting all credentials in usecase")

	credentials, err := uc.credentialRepo.GetAll(ctx)
	if err != nil {
		uc.logger.ErrorContext(ctx, "Error getting all credentials", "error", err)
		return nil, fmt.Errorf("error getting all credentials: %w", err)
	}

	// Decrypt credentials for each
	for _, cred := range credentials {
		decrypted, err := uc.decrypt(cred.Credentials)
		if err != nil {
			uc.logger.ErrorContext(ctx, "Failed to decrypt credentials", "id", cred.ID, "error", err)
			return nil, fmt.Errorf("failed to decrypt credentials for id %s: %w", cred.ID, err)
		}
		cred.Credentials = decrypted
	}

	uc.logger.InfoContext(ctx, "All credentials retrieved in usecase", "count", len(credentials))
	return credentials, nil
}

// UpdateCredential modifies an existing credential
func (uc *credentialUseCase) UpdateCredential(ctx context.Context, credential *model.AgentSupplierCredential) error {
	uc.logger.InfoContext(ctx, "Updating credential in usecase", "id", credential.ID, "agentID", credential.IataAgentID)

	// Business logic validation
	if credential.ID == "" {
		uc.logger.WarnContext(ctx, "Credential ID is required for update")
		return domain.ErrInvalidID
	}

	if credential.Credentials == "" {
		uc.logger.WarnContext(ctx, "Credentials are required for update")
		return domain.ErrCredentialsRequired
	}

	// Check if credential exists
	existing, err := uc.credentialRepo.GetByID(ctx, credential.ID)
	if err != nil {
		if errors.Is(err, domain.ErrNotFound) {
			uc.logger.WarnContext(ctx, "Credential not found for update", "id", credential.ID)
			return domain.ErrCredentialNotFound
		}
		uc.logger.ErrorContext(ctx, "Error checking existing credential", "id", credential.ID, "error", err)
		return fmt.Errorf("error checking existing credential: %w", err)
	}

	// Encrypt new credentials
	encryptedCredentials, err := uc.encrypt(credential.Credentials)
	if err != nil {
		uc.logger.ErrorContext(ctx, "Failed to encrypt credentials", "error", err)
		return fmt.Errorf("failed to encrypt credentials: %w", err)
	}
	credential.Credentials = encryptedCredentials

	// Preserve agent and supplier IDs
	credential.IataAgentID = existing.IataAgentID
	credential.SupplierID = existing.SupplierID

	if err := uc.credentialRepo.Update(ctx, credential); err != nil {
		uc.logger.ErrorContext(ctx, "Failed to update credential in repository", "id", credential.ID, "error", err)
		return err
	}

	uc.logger.InfoContext(ctx, "Credential updated successfully in usecase", "id", credential.ID, "agentID", credential.IataAgentID)
	return nil
}

// DeleteCredential removes a credential
func (uc *credentialUseCase) DeleteCredential(ctx context.Context, id string) error {
	uc.logger.InfoContext(ctx, "Deleting credential in usecase", "id", id)
	if id == "" {
		uc.logger.WarnContext(ctx, "Invalid credential ID provided", "id", id)
		return domain.ErrInvalidID
	}

	// Check if credential exists
	_, err := uc.credentialRepo.GetByID(ctx, id)
	if err != nil {
		if errors.Is(err, domain.ErrNotFound) {
			uc.logger.WarnContext(ctx, "Credential not found for deletion", "id", id)
			return domain.ErrCredentialNotFound
		}
		uc.logger.ErrorContext(ctx, "Error checking existing credential", "id", id, "error", err)
		return fmt.Errorf("error checking existing credential: %w", err)
	}

	if err := uc.credentialRepo.Delete(ctx, id); err != nil {
		uc.logger.ErrorContext(ctx, "Failed to delete credential in repository", "id", id, "error", err)
		return err
	}

	uc.logger.InfoContext(ctx, "Credential deleted successfully in usecase", "id", id)
	return nil
}
