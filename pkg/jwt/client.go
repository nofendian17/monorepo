package jwt

import (
	"context"
	"errors"
	"fmt"
	"monorepo/pkg/redis"
	"time"

	"github.com/golang-jwt/jwt/v5"
)

const (
	// Default secrets for development/testing
	DefaultAccessTokenSecret  = "default-access-secret"
	DefaultRefreshTokenSecret = "default-refresh-secret"

	// Token types
	TokenTypeAccess  = "access"
	TokenTypeRefresh = "refresh"

	// Session status
	SessionStatusActive   = "active"
	SessionStatusInactive = "inactive"

	// Issuer
	DefaultIssuer = "agent-service"

	// Redis key prefixes
	SessionKeyPrefix  = "session:"
	SessionKeyPattern = "session:*"

	// Session expiry (24 hours)
	SessionExpiry = 24 * time.Hour
)

// JWTClient defines the interface for JWT token operations
type JWTClient interface {
	GenerateAccessToken(userID, agentID, agentType string) (string, error)
	GenerateRefreshToken(userID, agentID, agentType string) (string, error)
	ValidateAccessToken(tokenString string) (*TokenClaims, error)
	ValidateRefreshToken(tokenString string) (*TokenClaims, error)
	RefreshAccessToken(refreshToken string) (string, error)
	RevokeRefreshToken(userID, tokenID string) error
	RevokeAllRefreshTokens(userID string) error
	Cleanup() error
	GetConfig() TokenConfig
	IsStateful() bool
	GetTokenExpiration(tokenString string) (time.Time, error)
	GetTokenRemainingTime(tokenString string) (time.Duration, error)
	IsTokenExpired(tokenString string) (bool, error)
	GetAccessTokenExpiry() time.Duration
	GetRefreshTokenExpiry() time.Duration
	CreateSession(ctx context.Context, userID, agentID, agentType, deviceInfo, ipAddress string) (*SessionInfo, string, error)
	GetSession(ctx context.Context, sessionID string) (*SessionInfo, error)
	UpdateSessionLastSeen(ctx context.Context, sessionID string) error
	EndSession(ctx context.Context, sessionID string) error
	GetUserSessions(ctx context.Context, userID string) ([]string, error)
	GenerateTokensWithSession(ctx context.Context, userID, agentID, agentType, deviceInfo, ipAddress string) (string, string, string, error)
}

const (
	// Error messages
	ErrAccessTokenSecretRequired     = "access token secret is required"
	ErrRefreshTokenSecretRequired    = "refresh token secret is required"
	ErrRefreshTokenNotFoundOrInvalid = "refresh token not found or invalid"
	ErrRefreshTokenNotInStore        = "refresh token not found in store"
	ErrInvalidTokenType              = "invalid token type"
	ErrInvalidToken                  = "invalid token"
	ErrRevokeNotSupportedStateless   = "revoke not supported in stateless mode"
	ErrNoStoreConfigured             = "no store configured for stateful mode"
	ErrSessionRequiresStatefulRedis  = "session management requires stateful mode with Redis"
	ErrRedisClientNotConfigured      = "Redis client not configured"
	ErrSessionNotFound               = "session not found"
)

// SessionInfo represents user session information stored in Redis
type SessionInfo struct {
	DeviceInfo string `json:"device_info"`
	IPAddress  string `json:"ip_address"`
	LastSeen   string `json:"last_seen"`
	Status     string `json:"status"`
}

// Client represents a JWT client that handles token operations
type Client struct {
	config      TokenConfig
	store       RefreshTokenStore
	redisClient redis.RedisClient
}

// New creates a new JWT client with the provided options
func New(opts ...Option) (JWTClient, error) {
	// Default configuration
	config := TokenConfig{
		AccessTokenSecret:  DefaultAccessTokenSecret,
		RefreshTokenSecret: DefaultRefreshTokenSecret,
		AccessTokenExpiry:  time.Minute * 15,
		RefreshTokenExpiry: time.Hour * 24 * 7,
		Stateful:           false,
	}

	// Apply options
	for _, opt := range opts {
		opt(&config)
	}

	// Validate configuration
	if config.AccessTokenSecret == "" {
		return nil, errors.New(ErrAccessTokenSecretRequired)
	}
	if config.RefreshTokenSecret == "" {
		return nil, errors.New(ErrRefreshTokenSecretRequired)
	}

	client := &Client{
		config:      config,
		store:       nil, // No store for stateless mode by default
		redisClient: nil,
	}

	return client, nil
}

// NewStateless creates a new JWT client for stateless mode
func NewStateless(opts ...Option) (JWTClient, error) {
	return New(opts...)
}

// NewStateful creates a new JWT client for stateful mode with a store
func NewStateful(store RefreshTokenStore, opts ...Option) (JWTClient, error) {
	client, err := New(opts...)
	if err != nil {
		return nil, err
	}

	// Type assert to access internal fields
	c := client.(*Client)
	c.store = store
	return client, nil
}

// NewStatefulWithRedis creates a new JWT client for stateful mode with Redis client
func NewStatefulWithRedis(redisClient redis.RedisClient, opts ...Option) (JWTClient, error) {
	client, err := New(opts...)
	if err != nil {
		return nil, err
	}

	// Type assert to access internal fields
	c := client.(*Client)
	c.store = NewRedisStore(redisClient)
	c.redisClient = redisClient
	return client, nil
}

// GenerateAccessToken generates a new access token
func (c *Client) GenerateAccessToken(userID, agentID, agentType string) (string, error) {
	// Create a unique JWT ID for this session
	jti := fmt.Sprintf("%s_%d", userID, time.Now().UnixNano())

	claims := TokenClaims{
		UserID:    userID,
		AgentID:   agentID,
		AgentType: agentType,
		TokenType: TokenTypeAccess,
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(c.config.AccessTokenExpiry)),
			IssuedAt:  jwt.NewNumericDate(time.Now()),
			Issuer:    DefaultIssuer,
			ID:        jti,
		},
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	return token.SignedString([]byte(c.config.AccessTokenSecret))
}

// GenerateRefreshToken generates a new refresh token
func (c *Client) GenerateRefreshToken(userID, agentID, agentType string) (string, error) {
	// Create a unique token ID
	tokenID := fmt.Sprintf("%s_%d", userID, time.Now().UnixNano())

	claims := TokenClaims{
		UserID:    userID,
		AgentID:   agentID,
		AgentType: agentType,
		TokenType: TokenTypeRefresh,
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(c.config.RefreshTokenExpiry)),
			IssuedAt:  jwt.NewNumericDate(time.Now()),
			Issuer:    DefaultIssuer,
			ID:        tokenID,
		},
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	refreshToken, err := token.SignedString([]byte(c.config.RefreshTokenSecret))
	if err != nil {
		return "", err
	}

	// If stateful, save the refresh token to store
	if c.config.Stateful && c.store != nil {
		expiryTime := time.Now().Add(c.config.RefreshTokenExpiry)
		err = c.store.Save(userID, tokenID, refreshToken, expiryTime)
		if err != nil {
			return "", err
		}
	}

	return refreshToken, nil
}

// ValidateAccessToken validates an access token
func (c *Client) ValidateAccessToken(tokenString string) (*TokenClaims, error) {
	return c.validateToken(tokenString, c.config.AccessTokenSecret, "access")
}

// ValidateRefreshToken validates a refresh token
func (c *Client) ValidateRefreshToken(tokenString string) (*TokenClaims, error) {
	claims, err := c.validateToken(tokenString, c.config.RefreshTokenSecret, "refresh")
	if err != nil {
		return nil, err
	}

	// If stateful, check if the token exists in the store
	if c.config.Stateful && c.store != nil {
		storedToken, err := c.store.Get(claims.UserID, claims.ID)
		if err != nil {
			return nil, fmt.Errorf("refresh token not found or invalid: %w", err)
		}

		if storedToken != tokenString {
			return nil, errors.New(ErrRefreshTokenNotInStore)
		}
	}

	return claims, nil
}

// validateToken is a helper function to validate tokens
func (c *Client) validateToken(tokenString, secret, expectedType string) (*TokenClaims, error) {
	token, err := jwt.ParseWithClaims(tokenString, &TokenClaims{}, func(token *jwt.Token) (interface{}, error) {
		return []byte(secret), nil
	})

	if err != nil {
		return nil, err
	}

	if claims, ok := token.Claims.(*TokenClaims); ok && token.Valid {
		if claims.TokenType != expectedType {
			return nil, errors.New(ErrInvalidTokenType)
		}
		return claims, nil
	}

	return nil, errors.New(ErrInvalidToken)
}

// RefreshAccessToken refreshes an access token using a refresh token
func (c *Client) RefreshAccessToken(refreshToken string) (string, error) {
	claims, err := c.ValidateRefreshToken(refreshToken)
	if err != nil {
		return "", err
	}

	// If stateful, remove the used refresh token to prevent reuse
	if c.config.Stateful && c.store != nil {
		err = c.store.Delete(claims.UserID, claims.ID)
		if err != nil {
			return "", fmt.Errorf("failed to invalidate used refresh token: %w", err)
		}
	}

	// Generate new access token with same user details
	return c.GenerateAccessToken(claims.UserID, claims.AgentID, claims.AgentType)
}

// RevokeRefreshToken revokes a refresh token (only works in stateful mode)
func (c *Client) RevokeRefreshToken(userID, tokenID string) error {
	if !c.config.Stateful {
		return errors.New(ErrRevokeNotSupportedStateless)
	}

	if c.store == nil {
		return errors.New(ErrNoStoreConfigured)
	}

	return c.store.Delete(userID, tokenID)
}

// RevokeAllRefreshTokens revokes all refresh tokens for a user (only works in stateful mode)
func (c *Client) RevokeAllRefreshTokens(userID string) error {
	if !c.config.Stateful {
		return errors.New(ErrRevokeNotSupportedStateless)
	}

	if c.store == nil {
		return errors.New(ErrNoStoreConfigured)
	}

	return c.store.DeleteAll(userID)
}

// Cleanup removes expired tokens from the store (only relevant in stateful mode)
func (c *Client) Cleanup() error {
	if !c.config.Stateful || c.store == nil {
		return nil
	}

	return c.store.Cleanup()
}

// GetConfig returns the current configuration
func (c *Client) GetConfig() TokenConfig {
	return c.config
}

// IsStateful returns whether the client is in stateful mode
func (c *Client) IsStateful() bool {
	return c.config.Stateful
}

// GetTokenExpiration returns the expiration time of a token without full validation
func (c *Client) GetTokenExpiration(tokenString string) (time.Time, error) {
	token, err := jwt.ParseWithClaims(tokenString, &TokenClaims{}, func(token *jwt.Token) (interface{}, error) {
		// Use the appropriate secret based on token type if possible
		// For now, try access token secret first
		return []byte(c.config.AccessTokenSecret), nil
	})

	if err != nil {
		// If access token secret fails, try refresh token secret
		token, err = jwt.ParseWithClaims(tokenString, &TokenClaims{}, func(token *jwt.Token) (interface{}, error) {
			return []byte(c.config.RefreshTokenSecret), nil
		})
		if err != nil {
			return time.Time{}, err
		}
	}

	if claims, ok := token.Claims.(*TokenClaims); ok {
		if claims.ExpiresAt != nil {
			return claims.ExpiresAt.Time, nil
		}
		return time.Time{}, errors.New("token has no expiration")
	}

	return time.Time{}, errors.New("invalid token claims")
}

// GetTokenRemainingTime returns the remaining time until token expiration
func (c *Client) GetTokenRemainingTime(tokenString string) (time.Duration, error) {
	expiry, err := c.GetTokenExpiration(tokenString)
	if err != nil {
		return 0, err
	}

	remaining := time.Until(expiry)
	if remaining < 0 {
		return 0, errors.New("token is expired")
	}

	return remaining, nil
}

// IsTokenExpired checks if a token is expired
func (c *Client) IsTokenExpired(tokenString string) (bool, error) {
	expiry, err := c.GetTokenExpiration(tokenString)
	if err != nil {
		return false, err
	}

	return time.Now().After(expiry), nil
}

// GetAccessTokenExpiry returns the configured access token expiry duration
func (c *Client) GetAccessTokenExpiry() time.Duration {
	return c.config.AccessTokenExpiry
}

// GetRefreshTokenExpiry returns the configured refresh token expiry duration
func (c *Client) GetRefreshTokenExpiry() time.Duration {
	return c.config.RefreshTokenExpiry
}

// CreateSession creates a new user session with device tracking
func (c *Client) CreateSession(ctx context.Context, userID, agentID, agentType, deviceInfo, ipAddress string) (*SessionInfo, string, error) {
	if !c.config.Stateful || c.redisClient == nil {
		return nil, "", errors.New(ErrSessionRequiresStatefulRedis)
	}

	sessionID := fmt.Sprintf("%s_%d", userID, time.Now().UnixNano())
	lastSeen := time.Now().Format(time.RFC3339)

	sessionInfo := &SessionInfo{
		DeviceInfo: deviceInfo,
		IPAddress:  ipAddress,
		LastSeen:   lastSeen,
		Status:     SessionStatusActive,
	}

	// Store session info in Redis hash
	sessionKey := fmt.Sprintf("%s%s", SessionKeyPrefix, sessionID)
	err := c.redisClient.HMSet(ctx, sessionKey, map[string]interface{}{
		"user_id":     userID,
		"agent_id":    agentID,
		"agent_type":  agentType,
		"device_info": deviceInfo,
		"ip_address":  ipAddress,
		"last_seen":   lastSeen,
		"status":      SessionStatusActive,
		"created_at":  time.Now().Format(time.RFC3339),
	})
	if err != nil {
		return nil, "", fmt.Errorf("failed to store session info: %w", err)
	}

	// Set session expiry (24 hours)
	err = c.redisClient.Expire(ctx, sessionKey, SessionExpiry)
	if err != nil {
		return nil, "", fmt.Errorf("failed to set session expiry: %w", err)
	}

	return sessionInfo, sessionID, nil
}

// GetSession retrieves session information by session ID
func (c *Client) GetSession(ctx context.Context, sessionID string) (*SessionInfo, error) {
	if c.redisClient == nil {
		return nil, errors.New(ErrRedisClientNotConfigured)
	}

	sessionKey := fmt.Sprintf("%s%s", SessionKeyPrefix, sessionID)
	exists, err := c.redisClient.Exists(ctx, sessionKey)
	if err != nil {
		return nil, fmt.Errorf("failed to check session existence: %w", err)
	}

	if !exists {
		return nil, errors.New(ErrSessionNotFound)
	}

	fields, err := c.redisClient.HMGet(ctx, sessionKey, "device_info", "ip_address", "last_seen", "status")
	if err != nil {
		return nil, fmt.Errorf("failed to get session info: %w", err)
	}

	sessionInfo := &SessionInfo{
		DeviceInfo: getStringValue(fields[0]),
		IPAddress:  getStringValue(fields[1]),
		LastSeen:   getStringValue(fields[2]),
		Status:     getStringValue(fields[3]),
	}

	return sessionInfo, nil
}

// UpdateSessionLastSeen updates the last seen timestamp for a session
func (c *Client) UpdateSessionLastSeen(ctx context.Context, sessionID string) error {
	if c.redisClient == nil {
		return errors.New(ErrRedisClientNotConfigured)
	}

	sessionKey := fmt.Sprintf("%s%s", SessionKeyPrefix, sessionID)
	lastSeen := time.Now().Format(time.RFC3339)

	err := c.redisClient.HSet(ctx, sessionKey, "last_seen", lastSeen)
	if err != nil {
		return fmt.Errorf("failed to update session last seen: %w", err)
	}

	return nil
}

// EndSession marks a session as inactive
func (c *Client) EndSession(ctx context.Context, sessionID string) error {
	if c.redisClient == nil {
		return errors.New(ErrRedisClientNotConfigured)
	}

	sessionKey := fmt.Sprintf("%s%s", SessionKeyPrefix, sessionID)
	err := c.redisClient.HSet(ctx, sessionKey, "status", SessionStatusInactive)
	if err != nil {
		return fmt.Errorf("failed to end session: %w", err)
	}

	return nil
}

// GetUserSessions retrieves all active sessions for a user
func (c *Client) GetUserSessions(ctx context.Context, userID string) ([]string, error) {
	if c.redisClient == nil {
		return nil, errors.New(ErrRedisClientNotConfigured)
	}

	// Find all session keys for this user
	pattern := SessionKeyPattern
	keys, err := c.redisClient.GetClient().Keys(ctx, pattern).Result()
	if err != nil {
		return nil, fmt.Errorf("failed to find user sessions: %w", err)
	}

	var userSessions []string
	for _, key := range keys {
		// Check if this session belongs to the user
		userIDField, err := c.redisClient.HGet(ctx, key, "user_id")
		if err == nil && userIDField == userID {
			// Extract session ID from key (remove "session:" prefix)
			sessionID := key[len(SessionKeyPrefix):]
			userSessions = append(userSessions, sessionID)
		}
	}

	return userSessions, nil
}

// GenerateTokensWithSession generates access and refresh tokens with session tracking
func (c *Client) GenerateTokensWithSession(ctx context.Context, userID, agentID, agentType, deviceInfo, ipAddress string) (string, string, string, error) {
	// Create session
	sessionInfo, sessionID, err := c.CreateSession(ctx, userID, agentID, agentType, deviceInfo, ipAddress)
	if err != nil {
		return "", "", "", err
	}

	// Generate access token with session info
	accessToken, err := c.GenerateAccessToken(userID, agentID, agentType)
	if err != nil {
		return "", "", "", err
	}

	// Generate refresh token
	refreshToken, err := c.GenerateRefreshToken(userID, agentID, agentType)
	if err != nil {
		return "", "", "", err
	}

	_ = sessionInfo // Use sessionInfo if needed
	return accessToken, refreshToken, sessionID, nil
}

// Helper function to safely get string value from interface{}
func getStringValue(value interface{}) string {
	if value == nil {
		return ""
	}
	if str, ok := value.(string); ok {
		return str
	}
	return ""
}
