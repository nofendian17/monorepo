// Package usecase contains business logic for authentication operations
package usecase

import (
	"context"
	"errors"
	"fmt"
	"time"

	"agent-service/domain"
	"agent-service/domain/repository"
	"monorepo/contracts/agent_service"
	"monorepo/pkg/jwt"
	"monorepo/pkg/logger"

	"golang.org/x/crypto/bcrypt"
)

// AuthUseCase defines the interface for authentication-related business operations
type AuthUseCase interface {
	// Login authenticates a user with email and password
	// It takes a context for request-scoped values, a LoginRequest, user agent, and IP address
	// Returns a LoginResponse with tokens, or an error if authentication fails
	Login(ctx context.Context, req agent_service.LoginRequest, userAgent, ipAddress string) (*agent_service.LoginResponse, error)
	// Refresh generates new access and refresh tokens using a valid refresh token
	// It implements fail-fast token rotation: the old refresh token must be successfully revoked
	// before new tokens are issued to prevent having both old and new tokens valid simultaneously
	// It takes a context for request-scoped values and a RefreshTokenRequest
	// Returns a RefreshTokenResponse with new tokens, or an error if refresh fails
	Refresh(ctx context.Context, req agent_service.RefreshTokenRequest) (*agent_service.RefreshTokenResponse, error)
	// Profile retrieves the authenticated user's profile information
	// It takes a context for request-scoped values with user claims
	// Returns a UserResponse with user profile data, or an error if retrieval fails
	Profile(ctx context.Context) (*agent_service.UserResponse, error)
}

// authUseCase implements the AuthUseCase interface
type authUseCase struct {
	// userRepo is the repository interface for user database operations
	userRepo repository.User
	// agentRepo is the repository interface for agent database operations
	agentRepo repository.Agent
	// jwtClient is the JWT client for token generation and validation
	jwtClient jwt.JWTClient
	// logger is used for logging operations within the usecase
	logger logger.LoggerInterface
}

// NewAuthUseCase creates a new instance of authUseCase
// It takes a User repository implementation, Agent repository implementation, JWT client, and a logger instance
// Returns an implementation of the AuthUseCase interface
func NewAuthUseCase(userRepo repository.User, agentRepo repository.Agent, jwtClient jwt.JWTClient, appLogger logger.LoggerInterface) AuthUseCase {
	return &authUseCase{
		userRepo:  userRepo,
		agentRepo: agentRepo,
		jwtClient: jwtClient,
		logger:    appLogger,
	}
}

// Login authenticates a user with email and password
// It validates the credentials, generates access and refresh tokens
// Returns a LoginResponse with tokens, or an error if authentication fails
func (uc *authUseCase) Login(ctx context.Context, req agent_service.LoginRequest, userAgent, ipAddress string) (*agent_service.LoginResponse, error) {
	uc.logger.InfoContext(ctx, "Login attempt", "email", req.Email)

	// Get user by email
	user, err := uc.userRepo.GetByEmail(ctx, req.Email)
	if err != nil {
		if errors.Is(err, domain.ErrNotFound) {
			uc.logger.WarnContext(ctx, "User not found", "email", req.Email)
			return nil, domain.ErrInvalidCredentials
		}
		uc.logger.ErrorContext(ctx, "Error retrieving user", "email", req.Email, "error", err)
		return nil, fmt.Errorf("error retrieving user: %w", err)
	}

	// Check if user is active
	if !user.IsActive {
		uc.logger.WarnContext(ctx, "User is not active", "email", req.Email)
		return nil, errors.New("user account is not active")
	}

	// Verify password
	err = bcrypt.CompareHashAndPassword([]byte(user.Password), []byte(req.Password))
	if err != nil {
		uc.logger.WarnContext(ctx, "Invalid password", "email", req.Email)
		return nil, domain.ErrInvalidCredentials
	}

	// Generate access token
	agentID := ""
	agentType := ""
	if user.AgentID != nil {
		agentID = *user.AgentID
		// Get agent type
		agent, err := uc.agentRepo.GetByID(ctx, agentID)
		if err != nil {
			uc.logger.WarnContext(ctx, "Error retrieving agent for token generation", "agentID", agentID, "error", err)
			// Continue with empty agentType - token will still work
		} else {
			agentType = agent.AgentType
		}
	}

	var accessToken, refreshToken string
	var sessionID string

	// Generate tokens based on JWT client mode (stateful or stateless)
	if uc.jwtClient.IsStateful() {
		// Stateful mode: Generate tokens with session tracking in Redis
		accessToken, refreshToken, sessionID, err = uc.jwtClient.GenerateTokensWithSession(
			ctx, user.ID, agentID, agentType, userAgent, ipAddress,
		)
		if err != nil {
			uc.logger.ErrorContext(ctx, "Error generating tokens with session", "userID", user.ID, "error", err)
			return nil, fmt.Errorf("error generating tokens with session: %w", err)
		}
		uc.logger.InfoContext(ctx, "Login successful (stateful)", "userID", user.ID, "email", req.Email, "sessionID", sessionID)
	} else {
		// Stateless mode: Generate tokens without session tracking
		accessToken, err = uc.jwtClient.GenerateAccessToken(user.ID, agentID, agentType)
		if err != nil {
			uc.logger.ErrorContext(ctx, "Error generating access token", "userID", user.ID, "error", err)
			return nil, fmt.Errorf("error generating access token: %w", err)
		}

		refreshToken, err = uc.jwtClient.GenerateRefreshToken(user.ID, agentID, agentType)
		if err != nil {
			uc.logger.ErrorContext(ctx, "Error generating refresh token", "userID", user.ID, "error", err)
			return nil, fmt.Errorf("error generating refresh token: %w", err)
		}

		uc.logger.InfoContext(ctx, "Login successful (stateless)", "userID", user.ID, "email", req.Email)
	}

	// Get token expiration times
	accessTokenExpire, err := uc.jwtClient.GetTokenExpiration(accessToken)
	if err != nil {
		uc.logger.ErrorContext(ctx, "Error getting access token expiration", "userID", user.ID, "error", err)
		return nil, fmt.Errorf("error getting access token expiration: %w", err)
	}

	refreshTokenExpire, err := uc.jwtClient.GetTokenExpiration(refreshToken)
	if err != nil {
		uc.logger.ErrorContext(ctx, "Error getting refresh token expiration", "userID", user.ID, "error", err)
		return nil, fmt.Errorf("error getting refresh token expiration: %w", err)
	}

	return &agent_service.LoginResponse{
		AccessToken:        accessToken,
		RefreshToken:       refreshToken,
		AccessTokenExpire:  int64(time.Until(accessTokenExpire).Seconds()),
		RefreshTokenExpire: int64(time.Until(refreshTokenExpire).Seconds()),
	}, nil
}

// Refresh generates new access and refresh tokens using a valid refresh token
// It implements fail-fast token rotation: the old refresh token must be successfully revoked
// before new tokens are issued to prevent having both old and new tokens valid simultaneously
// It takes a context for request-scoped values and a RefreshTokenRequest
// Returns a RefreshTokenResponse with new tokens, or an error if refresh fails
func (uc *authUseCase) Refresh(ctx context.Context, req agent_service.RefreshTokenRequest) (*agent_service.RefreshTokenResponse, error) {
	uc.logger.InfoContext(ctx, "Refresh token attempt")

	// Validate the refresh token
	claims, err := uc.jwtClient.ValidateRefreshToken(req.RefreshToken)
	if err != nil {
		uc.logger.WarnContext(ctx, "Invalid refresh token", "error", err)
		return nil, errors.New("invalid refresh token")
	}

	// Check if the user exists
	user, err := uc.userRepo.GetByID(ctx, claims.UserID)
	if err != nil {
		uc.logger.ErrorContext(ctx, "Error retrieving user by ID", "userID", claims.UserID, "error", err)
		return nil, fmt.Errorf("error retrieving user: %w", err)
	}

	// Check if user is active
	if !user.IsActive {
		uc.logger.WarnContext(ctx, "User is not active", "userID", claims.UserID)
		return nil, errors.New("user account is not active")
	}

	// Revoke the old refresh token (only in stateful mode)
	// This is a fail-fast approach: if revocation fails, the entire refresh operation fails
	// to prevent having both old and new tokens valid simultaneously
	if uc.jwtClient.IsStateful() {
		err = uc.jwtClient.RevokeRefreshToken(claims.UserID, claims.ID)
		if err != nil {
			uc.logger.ErrorContext(ctx, "Failed to revoke old refresh token - aborting refresh to maintain security", "userID", claims.UserID, "tokenID", claims.ID, "error", err)
			return nil, fmt.Errorf("failed to revoke old refresh token: %w", err)
		}
		uc.logger.InfoContext(ctx, "Old refresh token revoked successfully", "userID", claims.UserID, "tokenID", claims.ID)
	}

	// Generate new tokens
	var accessToken, refreshToken string
	if uc.jwtClient.IsStateful() {
		// Stateful mode: Generate tokens with session tracking in Redis
		accessToken, refreshToken, _, err = uc.jwtClient.GenerateTokensWithSession(
			ctx, user.ID, claims.AgentID, claims.AgentType, "", "",
		)
		if err != nil {
			uc.logger.ErrorContext(ctx, "Error generating new tokens with session", "userID", user.ID, "error", err)
			return nil, fmt.Errorf("error generating new tokens with session: %w", err)
		}
		uc.logger.InfoContext(ctx, "Token refresh successful (stateful)", "userID", user.ID)
	} else {
		// Stateless mode: Generate tokens without session tracking
		accessToken, err = uc.jwtClient.GenerateAccessToken(user.ID, claims.AgentID, claims.AgentType)
		if err != nil {
			uc.logger.ErrorContext(ctx, "Error generating new access token", "userID", user.ID, "error", err)
			return nil, fmt.Errorf("error generating new access token: %w", err)
		}

		refreshToken, err = uc.jwtClient.GenerateRefreshToken(user.ID, claims.AgentID, claims.AgentType)
		if err != nil {
			uc.logger.ErrorContext(ctx, "Error generating new refresh token", "userID", user.ID, "error", err)
			return nil, fmt.Errorf("error generating new refresh token: %w", err)
		}

		uc.logger.InfoContext(ctx, "Token refresh successful (stateless)", "userID", user.ID)
	}

	// Get token expiration times
	accessTokenExpire, err := uc.jwtClient.GetTokenExpiration(accessToken)
	if err != nil {
		uc.logger.ErrorContext(ctx, "Error getting new access token expiration", "userID", user.ID, "error", err)
		return nil, fmt.Errorf("error getting new access token expiration: %w", err)
	}

	refreshTokenExpire, err := uc.jwtClient.GetTokenExpiration(refreshToken)
	if err != nil {
		uc.logger.ErrorContext(ctx, "Error getting new refresh token expiration", "userID", user.ID, "error", err)
		return nil, fmt.Errorf("error getting new refresh token expiration: %w", err)
	}

	return &agent_service.RefreshTokenResponse{
		AccessToken:        accessToken,
		RefreshToken:       refreshToken,
		AccessTokenExpire:  int64(time.Until(accessTokenExpire).Seconds()),
		RefreshTokenExpire: int64(time.Until(refreshTokenExpire).Seconds()),
	}, nil
}

// Profile retrieves the authenticated user's profile information
// It extracts the user ID from the context and fetches the user data
// Returns a UserResponse with user profile data, or an error if retrieval fails
func (uc *authUseCase) Profile(ctx context.Context) (*agent_service.UserResponse, error) {
	uc.logger.InfoContext(ctx, "Profile request")

	// Extract user ID from context (set by JWT middleware)
	userID, ok := ctx.Value("user_id").(string)
	if !ok || userID == "" {
		uc.logger.WarnContext(ctx, "User ID not found in context")
		return nil, errors.New("unauthorized: user ID not found")
	}

	// Get user by ID
	user, err := uc.userRepo.GetByID(ctx, userID)
	if err != nil {
		if errors.Is(err, domain.ErrNotFound) {
			uc.logger.WarnContext(ctx, "User not found", "userID", userID)
			return nil, domain.ErrNotFound
		}
		uc.logger.ErrorContext(ctx, "Error retrieving user", "userID", userID, "error", err)
		return nil, fmt.Errorf("error retrieving user: %w", err)
	}

	uc.logger.InfoContext(ctx, "Profile retrieved successfully", "userID", userID)
	return agent_service.UserModelToResponse(user), nil
}
