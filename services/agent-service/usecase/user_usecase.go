// Package usecase contains business logic for user operations
package usecase

import (
	"context"
	"errors"
	"fmt"

	"agent-service/domain"
	"agent-service/domain/model"
	"agent-service/domain/repository"
	"monorepo/pkg/logger"

	"golang.org/x/crypto/bcrypt"
)

// UserUseCase defines the interface for user-related business operations
// It provides methods for CRUD operations and listing users with business logic
type UserUseCase interface {
	// CreateUser adds a new user with business validation
	// It takes a context for request-scoped values and a pointer to a User model
	// Returns an error if the operation fails
	CreateUser(ctx context.Context, user *model.User) error
	// GetUserByID retrieves a user by their unique identifier
	// It takes a context for request-scoped values and the user ID
	// Returns the user model and an error if the operation fails
	GetUserByID(ctx context.Context, id string) (*model.User, error)
	// GetUserByEmail retrieves a user by their email address
	// It takes a context for request-scoped values and the email address
	// Returns the user model and an error if the operation fails
	GetUserByEmail(ctx context.Context, email string) (*model.User, error)
	// UpdateUser modifies an existing user with business validation
	// It takes a context for request-scoped values and a pointer to a User model
	// Returns an error if the operation fails
	UpdateUser(ctx context.Context, user *model.User) error
	// UpdateUserStatus updates the active status of a user
	// It takes a context for request-scoped values, user ID, and the new active status
	// Returns an error if the operation fails
	UpdateUserStatus(ctx context.Context, id string, isActive bool) error
	// DeleteUser removes a user from the system
	// It takes a context for request-scoped values and the user ID
	// Returns an error if the operation fails
	DeleteUser(ctx context.Context, id string) error
	// GetUsersByAgentID retrieves users by their associated agent ID
	// It takes a context for request-scoped values and the agent ID
	// Returns a slice of user pointers and an error if the operation fails
	GetUsersByAgentID(ctx context.Context, agentID string) ([]*model.User, error)
	// GetActiveUsers retrieves all active users
	// It takes a context for request-scoped values
	// Returns a slice of user pointers and an error if the operation fails
	GetActiveUsers(ctx context.Context) ([]*model.User, error)
	// ListUsers retrieves a paginated list of users
	// It takes a context for request-scoped values, offset for pagination, and limit for page size
	// Returns a slice of user pointers, the real total count, and an error if the operation fails
	ListUsers(ctx context.Context, offset, limit int) ([]*model.User, int, error)
}

// userUseCase implements the UserUseCase interface
type userUseCase struct {
	// userRepo is the repository interface for user database operations
	userRepo repository.User
	// logger is used for logging operations within the usecase
	logger logger.LoggerInterface
}

// hashPassword hashes a plain password using bcrypt
func hashPassword(password string) (string, error) {
	if password == "" {
		return "", nil
	}
	hashed, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	return string(hashed), err
}

// NewUserUseCase creates a new instance of userUseCase
// It takes a User repository implementation and a logger instance
// Returns an implementation of the UserUseCase interface
func NewUserUseCase(userRepo repository.User, appLogger logger.LoggerInterface) UserUseCase {
	return &userUseCase{
		userRepo: userRepo,
		logger:   appLogger,
	}
}

// CreateUser adds a new user with business validation
// It takes a context for request-scoped values and a pointer to a User model
// Returns an error if the operation fails
func (uc *userUseCase) CreateUser(ctx context.Context, user *model.User) error {
	uc.logger.InfoContext(ctx, "Creating user in usecase", "email", user.Email)
	// Business logic validation
	if user.Email == "" {
		uc.logger.WarnContext(ctx, "Email is required for user creation")
		return domain.ErrEmailRequired
	}

	// Check if user with email already exists
	existingUser, err := uc.userRepo.GetByEmail(ctx, user.Email)
	if err != nil && !errors.Is(err, domain.ErrNotFound) {
		uc.logger.ErrorContext(ctx, "Error checking existing user", "email", user.Email, "error", err)
		return fmt.Errorf("error checking existing user: %w", err)
	}

	if existingUser != nil {
		uc.logger.WarnContext(ctx, "User with email already exists", "email", user.Email)
		return domain.ErrEmailAlreadyExists
	}

	// Hash the password before saving
	if user.Password != "" {
		hashedPassword, err := hashPassword(user.Password)
		if err != nil {
			uc.logger.ErrorContext(ctx, "Failed to hash password", "error", err)
			return fmt.Errorf("failed to hash password: %w", err)
		}
		user.Password = hashedPassword
	}

	if err := uc.userRepo.Create(ctx, user); err != nil {
		uc.logger.ErrorContext(ctx, "Failed to create user in repository", "email", user.Email, "error", err)
		return err
	}

	uc.logger.InfoContext(ctx, "User created successfully in usecase", "id", user.ID, "email", user.Email)
	return nil
}

// GetUserByID retrieves a user by their unique identifier
// It takes a context for request-scoped values and the user ID
// Returns the user model and an error if the operation fails
func (uc *userUseCase) GetUserByID(ctx context.Context, id string) (*model.User, error) {
	uc.logger.InfoContext(ctx, "Getting user by ID in usecase", "id", id)
	if id == "" {
		uc.logger.WarnContext(ctx, "Invalid user ID provided", "id", id)
		return nil, domain.ErrInvalidID
	}

	user, err := uc.userRepo.GetByID(ctx, id)
	if err != nil {
		if errors.Is(err, domain.ErrNotFound) {
			uc.logger.WarnContext(ctx, "User not found by ID", "id", id)
			return nil, domain.ErrUserNotFound
		}
		uc.logger.ErrorContext(ctx, "Error getting user by ID", "id", id, "error", err)
		return nil, fmt.Errorf("error getting user: %w", err)
	}

	uc.logger.InfoContext(ctx, "User retrieved by ID in usecase", "id", user.ID, "email", user.Email)
	return user, nil
}

// GetUserByEmail retrieves a user by their email address
// It takes a context for request-scoped values and the email address
// Returns the user model and an error if the operation fails
func (uc *userUseCase) GetUserByEmail(ctx context.Context, email string) (*model.User, error) {
	uc.logger.InfoContext(ctx, "Getting user by email in usecase", "email", email)
	if email == "" {
		uc.logger.WarnContext(ctx, "Email is required for user lookup")
		return nil, domain.ErrEmailRequired
	}

	user, err := uc.userRepo.GetByEmail(ctx, email)
	if err != nil {
		if errors.Is(err, domain.ErrNotFound) {
			uc.logger.WarnContext(ctx, "User not found by email", "email", email)
			return nil, domain.ErrUserNotFound
		}
		uc.logger.ErrorContext(ctx, "Error getting user by email", "email", email, "error", err)
		return nil, fmt.Errorf("error getting user by email: %w", err)
	}

	uc.logger.InfoContext(ctx, "User retrieved by email in usecase", "id", user.ID, "email", user.Email)
	return user, nil
}

// UpdateUser modifies an existing user with business validation
// It takes a context for request-scoped values and a pointer to a User model
// Returns an error if the operation fails
func (uc *userUseCase) UpdateUser(ctx context.Context, user *model.User) error {
	uc.logger.InfoContext(ctx, "Updating user in usecase", "id", user.ID, "email", user.Email)
	if user.ID == "" {
		uc.logger.WarnContext(ctx, "Invalid user ID for update", "id", user.ID)
		return domain.ErrInvalidID
	}

	if user.Email == "" {
		uc.logger.WarnContext(ctx, "Email is required for user update", "id", user.ID)
		return domain.ErrEmailRequired
	}

	// Check if user with email already exists (excluding current user)
	existingUser, err := uc.userRepo.GetByEmail(ctx, user.Email)
	if err != nil && !errors.Is(err, domain.ErrNotFound) {
		uc.logger.ErrorContext(ctx, "Error checking existing user during update", "email", user.Email, "error", err)
		return fmt.Errorf("error checking existing user: %w", err)
	}

	if existingUser != nil && existingUser.ID != user.ID {
		uc.logger.WarnContext(ctx, "Email already exists for another user", "email", user.Email, "existing_id", existingUser.ID, "update_id", user.ID)
		return domain.ErrEmailAlreadyExists
	}

	// Hash the password if it's provided
	if user.Password != "" {
		hashedPassword, err := hashPassword(user.Password)
		if err != nil {
			uc.logger.ErrorContext(ctx, "Failed to hash password during update", "id", user.ID, "error", err)
			return fmt.Errorf("failed to hash password: %w", err)
		}
		user.Password = hashedPassword
	}

	if err := uc.userRepo.Update(ctx, user); err != nil {
		uc.logger.ErrorContext(ctx, "Failed to update user in repository", "id", user.ID, "email", user.Email, "error", err)
		return err
	}

	uc.logger.InfoContext(ctx, "User updated successfully in usecase", "id", user.ID, "email", user.Email)
	return nil
}

// UpdateUserStatus updates the active status of a user
// It takes a context for request-scoped values, user ID, and the new active status
// Returns an error if the operation fails
func (uc *userUseCase) UpdateUserStatus(ctx context.Context, id string, isActive bool) error {
	uc.logger.InfoContext(ctx, "Updating user status in usecase", "id", id, "isActive", isActive)
	if id == "" {
		uc.logger.WarnContext(ctx, "Invalid user ID for status update", "id", id)
		return domain.ErrInvalidID
	}

	// Get existing user
	user, err := uc.userRepo.GetByID(ctx, id)
	if err != nil {
		if errors.Is(err, domain.ErrNotFound) {
			uc.logger.WarnContext(ctx, "User not found for status update", "id", id)
			return domain.ErrUserNotFound
		}
		uc.logger.ErrorContext(ctx, "Error getting user for status update", "id", id, "error", err)
		return fmt.Errorf("error getting user: %w", err)
	}

	// Update the status
	user.IsActive = isActive

	if err := uc.userRepo.Update(ctx, user); err != nil {
		uc.logger.ErrorContext(ctx, "Failed to update user status in repository", "id", user.ID, "isActive", isActive, "error", err)
		return err
	}

	uc.logger.InfoContext(ctx, "User status updated successfully in usecase", "id", user.ID, "isActive", isActive)
	return nil
}

// DeleteUser removes a user from the system
// It takes a context for request-scoped values and the user ID
// Returns an error if the operation fails
func (uc *userUseCase) DeleteUser(ctx context.Context, id string) error {
	uc.logger.InfoContext(ctx, "Deleting user in usecase", "id", id)
	if id == "" {
		uc.logger.WarnContext(ctx, "Invalid user ID for deletion", "id", id)
		return domain.ErrInvalidID
	}

	err := uc.userRepo.Delete(ctx, id)
	if err != nil {
		if errors.Is(err, domain.ErrNotFound) {
			uc.logger.WarnContext(ctx, "User not found for deletion", "id", id)
			return domain.ErrUserNotFound
		}
		uc.logger.ErrorContext(ctx, "Error deleting user", "id", id, "error", err)
		return fmt.Errorf("error deleting user: %w", err)
	}

	uc.logger.InfoContext(ctx, "User deleted successfully in usecase", "id", id)
	return nil
}

// ListUsers retrieves a paginated list of users
// It takes a context for request-scoped values, offset for pagination, and limit for page size
// Returns a slice of user pointers, the real total count, and an error if the operation fails
func (uc *userUseCase) ListUsers(ctx context.Context, offset, limit int) ([]*model.User, int, error) {
	uc.logger.InfoContext(ctx, "Listing users in usecase", "offset", offset, "limit", limit)
	if offset < 0 {
		offset = 0
	}
	if limit <= 0 {
		limit = 10
	}
	if limit > 100 {
		limit = 100
	}

	users, total, err := uc.userRepo.List(ctx, offset, limit)
	if err != nil {
		uc.logger.ErrorContext(ctx, "Error listing users", "offset", offset, "limit", limit, "error", err)
		return nil, 0, err
	}

	uc.logger.InfoContext(ctx, "Users listed successfully in usecase", "count", len(users), "offset", offset, "limit", limit, "total", total)
	return users, total, nil
}

// GetUsersByAgentID retrieves users by their associated agent ID
// It takes a context for request-scoped values and the agent ID
// Returns a slice of user pointers and an error if the operation fails
func (uc *userUseCase) GetUsersByAgentID(ctx context.Context, agentID string) ([]*model.User, error) {
	uc.logger.InfoContext(ctx, "Getting users by agent ID in usecase", "agentID", agentID)
	if agentID == "" {
		uc.logger.WarnContext(ctx, "Agent ID is required for user lookup by agent")
		return nil, domain.ErrInvalidID
	}

	users, err := uc.userRepo.GetByAgentID(ctx, agentID)
	if err != nil {
		uc.logger.ErrorContext(ctx, "Error getting users by agent ID", "agentID", agentID, "error", err)
		return nil, fmt.Errorf("error getting users by agent ID: %w", err)
	}

	uc.logger.InfoContext(ctx, "Users retrieved by agent ID in usecase", "count", len(users), "agentID", agentID)
	return users, nil
}

// GetActiveUsers retrieves all active users
// It takes a context for request-scoped values
// Returns a slice of user pointers and an error if the operation fails
func (uc *userUseCase) GetActiveUsers(ctx context.Context) ([]*model.User, error) {
	uc.logger.InfoContext(ctx, "Getting active users in usecase")

	users, err := uc.userRepo.GetActiveUsers(ctx)
	if err != nil {
		uc.logger.ErrorContext(ctx, "Error getting active users", "error", err)
		return nil, fmt.Errorf("error getting active users: %w", err)
	}

	uc.logger.InfoContext(ctx, "Active users retrieved in usecase", "count", len(users))
	return users, nil
}
