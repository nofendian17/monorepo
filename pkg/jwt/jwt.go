package jwt

import (
	"time"

	"github.com/golang-jwt/jwt/v5"
)

// TokenClaims represents the claims in a JWT token
type TokenClaims struct {
	UserID    string `json:"user_id"`
	AgentID   string `json:"agent_id"`
	AgentType string `json:"agent_type"`
	TokenType string `json:"token_type"`
	jwt.RegisteredClaims
}

// RefreshTokenStore defines the interface for storing and managing refresh tokens in stateful mode.
type RefreshTokenStore interface {
	Save(userID, tokenID, token string, expiry time.Time) error
	Get(userID, tokenID string) (string, error)
	Delete(userID, tokenID string) error
	DeleteAll(userID string) error
	Cleanup() error
}

// JWTManager handles JWT token operations (alias for Client).
// Deprecated: Use Client directly instead.
type JWTManager = Client

// NewJWTManager creates a new JWT manager with the given configuration.
// Deprecated: Use NewWithConfig instead.
func NewJWTManager(config TokenConfig, store RefreshTokenStore) (JWTClient, error) {
	opts := []Option{
		WithAccessTokenSecret(config.AccessTokenSecret),
		WithRefreshTokenSecret(config.RefreshTokenSecret),
		WithAccessTokenExpiry(config.AccessTokenExpiry),
		WithRefreshTokenExpiry(config.RefreshTokenExpiry),
		WithStateful(config.Stateful),
	}
	client, err := New(opts...)
	if err != nil {
		return nil, err
	}
	c := client.(*Client)
	c.store = store
	return client, nil
}

// NewJWTManagerStateless creates a new JWT manager for stateless mode.
// Deprecated: Use NewWithConfig instead.
func NewJWTManagerStateless(config TokenConfig) (JWTClient, error) {
	opts := []Option{
		WithAccessTokenSecret(config.AccessTokenSecret),
		WithRefreshTokenSecret(config.RefreshTokenSecret),
		WithAccessTokenExpiry(config.AccessTokenExpiry),
		WithRefreshTokenExpiry(config.RefreshTokenExpiry),
		WithStateful(false),
	}
	return New(opts...)
}
