// Package postgres provides PostgreSQL implementation for user repository
package postgres

import (
	"context"
	"fmt"

	"agent-service/domain"
	"agent-service/domain/model"
	"agent-service/domain/repository"
	"monorepo/pkg/logger"

	"gorm.io/gorm"
)

// userRepository implements the User repository interface using PostgreSQL
type userRepository struct {
	// db is the GORM database instance for database operations
	db *gorm.DB
	// logger is used for logging operations within the repository
	logger logger.LoggerInterface
}

// NewUserRepository creates a new instance of userRepository
// It takes a GORM database instance and a logger instance
// Returns an implementation of the TransactionalUser repository interface
func NewUserRepository(db *gorm.DB, logger logger.LoggerInterface) repository.TransactionalUser {
	return &userRepository{
		db:     db,
		logger: logger,
	}
}

// Create adds a new user to the database
// It takes a context for request-scoped values and a pointer to a User model
// Returns an error if the operation fails
func (r *userRepository) Create(ctx context.Context, user *model.User) error {
	r.logger.InfoContext(ctx, "Creating user", "email", user.Email)

	// Check if there's a transaction in the context
	db := r.db
	if tx, ok := ctx.Value("tx").(*gorm.DB); ok {
		db = tx
	}

	if err := db.WithContext(ctx).Create(user).Error; err != nil {
		r.logger.ErrorContext(ctx, "Failed to create user", "email", user.Email, "error", err)
		return fmt.Errorf("failed to create user: %w", err)
	}
	r.logger.InfoContext(ctx, "User created successfully", "id", user.ID, "email", user.Email)
	return nil
}

// GetByID retrieves a user by their unique identifier
// It takes a context for request-scoped values and the user ID
// Returns the user model and an error if the operation fails
func (r *userRepository) GetByID(ctx context.Context, id string) (*model.User, error) {
	r.logger.InfoContext(ctx, "Getting user by ID", "id", id)
	var user model.User
	if err := r.db.WithContext(ctx).Preload("Agent").Where("id = ? AND is_active = ? AND deleted_at IS NULL", id, true).First(&user).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			r.logger.WarnContext(ctx, "User not found by ID", "id", id)
			return nil, domain.ErrNotFound
		}
		r.logger.ErrorContext(ctx, "Failed to get user by ID", "id", id, "error", err)
		return nil, fmt.Errorf("failed to get user: %w", err)
	}
	r.logger.InfoContext(ctx, "User retrieved by ID", "id", user.ID, "email", user.Email)
	return &user, nil
}

// GetByEmail retrieves a user by their email address
// It takes a context for request-scoped values and the email address
// Returns the user model and an error if the operation fails
func (r *userRepository) GetByEmail(ctx context.Context, email string) (*model.User, error) {
	r.logger.InfoContext(ctx, "Getting user by email", "email", email)
	var user model.User
	if err := r.db.WithContext(ctx).Preload("Agent").Where("email = ? AND is_active = ? AND deleted_at IS NULL", email, true).First(&user).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			r.logger.WarnContext(ctx, "User not found by email", "email", email)
			return nil, domain.ErrNotFound
		}
		r.logger.ErrorContext(ctx, "Failed to get user by email", "email", email, "error", err)
		return nil, fmt.Errorf("failed to get user by email: %w", err)
	}
	r.logger.InfoContext(ctx, "User retrieved by email", "id", user.ID, "email", user.Email)
	return &user, nil
}

// Update modifies an existing user in the database
// It takes a context for request-scoped values and a pointer to a User model
// Returns an error if the operation fails
func (r *userRepository) Update(ctx context.Context, user *model.User) error {
	r.logger.InfoContext(ctx, "Updating user", "id", user.ID, "email", user.Email)
	if err := r.db.WithContext(ctx).Model(&model.User{}).Where("id = ?", user.ID).Updates(user).Error; err != nil {
		r.logger.ErrorContext(ctx, "Failed to update user", "id", user.ID, "email", user.Email, "error", err)
		return fmt.Errorf("failed to update user: %w", err)
	}
	r.logger.InfoContext(ctx, "User updated successfully", "id", user.ID, "email", user.Email)
	return nil
}

// UpdatePassword updates only the password of a user
// It takes a context for request-scoped values, user ID, and hashed password
// Returns an error if the operation fails
func (r *userRepository) UpdatePassword(ctx context.Context, id string, hashedPassword string) error {
	r.logger.InfoContext(ctx, "Updating user password", "id", id)
	if err := r.db.WithContext(ctx).Model(&model.User{}).Where("id = ?", id).Update("password", hashedPassword).Error; err != nil {
		r.logger.ErrorContext(ctx, "Failed to update user password", "id", id, "error", err)
		return fmt.Errorf("failed to update user password: %w", err)
	}
	r.logger.InfoContext(ctx, "User password updated successfully", "id", id)
	return nil
}

// Delete removes a user from the database (soft delete)
// It takes a context for request-scoped values and the user ID
// Returns an error if the operation fails
func (r *userRepository) Delete(ctx context.Context, id string) error {
	r.logger.InfoContext(ctx, "Deleting user", "id", id)
	user := &model.User{ID: id}

	// Use soft delete
	if err := r.db.WithContext(ctx).Delete(user).Error; err != nil {
		r.logger.ErrorContext(ctx, "Failed to delete user", "id", id, "error", err)
		return fmt.Errorf("failed to delete user: %w", err)
	}

	// Check if record was actually deleted
	var count int64
	r.db.WithContext(ctx).Model(&model.User{}).Where("id = ? AND deleted_at IS NULL", id).Count(&count)
	if count > 0 {
		r.logger.WarnContext(ctx, "User not found for deletion", "id", id)
		return domain.ErrNotFound
	}

	r.logger.InfoContext(ctx, "User deleted successfully", "id", id)
	return nil
}

// List retrieves a paginated list of users from the database
// It takes a context for request-scoped values, offset for pagination, and limit for page size
// Returns a slice of user pointers, the real total count, and an error if the operation fails
func (r *userRepository) List(ctx context.Context, offset, limit int) ([]*model.User, int, error) {
	r.logger.InfoContext(ctx, "Listing users", "offset", offset, "limit", limit)
	var users []*model.User
	var total int64

	// Get total count
	if err := r.db.WithContext(ctx).Model(&model.User{}).Where("is_active = ? AND deleted_at IS NULL", true).Count(&total).Error; err != nil {
		r.logger.ErrorContext(ctx, "Failed to count users", "error", err)
		return nil, 0, fmt.Errorf("failed to count users: %w", err)
	}

	// Get paginated users
	if err := r.db.WithContext(ctx).Where("is_active = ? AND deleted_at IS NULL", true).Offset(offset).Limit(limit).Order("id ASC").Find(&users).Error; err != nil {
		r.logger.ErrorContext(ctx, "Failed to list users", "offset", offset, "limit", limit, "error", err)
		return nil, 0, fmt.Errorf("failed to list users: %w", err)
	}

	r.logger.InfoContext(ctx, "Users listed successfully", "count", len(users), "offset", offset, "limit", limit, "total", total)
	return users, int(total), nil
}

// GetByAgentID retrieves users by their associated agent ID
// It takes a context for request-scoped values and the agent ID
// Returns a slice of user pointers and an error if the operation fails
func (r *userRepository) GetByAgentID(ctx context.Context, agentID string) ([]*model.User, error) {
	r.logger.InfoContext(ctx, "Getting users by agent ID", "agentID", agentID)
	var users []*model.User
	if err := r.db.WithContext(ctx).Where("agent_id = ? AND is_active = ? AND deleted_at IS NULL", agentID, true).Find(&users).Error; err != nil {
		r.logger.ErrorContext(ctx, "Failed to get users by agent ID", "agentID", agentID, "error", err)
		return nil, fmt.Errorf("failed to get users by agent ID: %w", err)
	}
	r.logger.InfoContext(ctx, "Users retrieved by agent ID", "count", len(users), "agentID", agentID)
	return users, nil
}

// GetActiveUsers retrieves all active users
// It takes a context for request-scoped values
// Returns a slice of user pointers and an error if the operation fails
func (r *userRepository) GetActiveUsers(ctx context.Context) ([]*model.User, error) {
	r.logger.InfoContext(ctx, "Getting active users")
	var users []*model.User
	if err := r.db.WithContext(ctx).Preload("Agent").Where("is_active = ? AND deleted_at IS NULL", true).Find(&users).Error; err != nil {
		r.logger.ErrorContext(ctx, "Failed to get active users", "error", err)
		return nil, fmt.Errorf("failed to get active users: %w", err)
	}
	r.logger.InfoContext(ctx, "Active users retrieved", "count", len(users))
	return users, nil
}

// ExecuteInTransaction executes a function within a database transaction
// The function receives a transaction context that should be used for all operations
// Returns an error if the transaction fails or if the function returns an error
func (r *userRepository) ExecuteInTransaction(ctx context.Context, fn func(txCtx context.Context) error) error {
	r.logger.InfoContext(ctx, "Executing operation in transaction")
	return r.db.Transaction(func(tx *gorm.DB) error {
		// Create a context that carries the transaction
		txCtx := context.WithValue(ctx, "tx", tx)
		return fn(txCtx)
	})
}
