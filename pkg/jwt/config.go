package jwt

import (
	"time"
)

// TokenConfig holds the configuration for JWT tokens
type TokenConfig struct {
	AccessTokenSecret  string
	RefreshTokenSecret string
	AccessTokenExpiry  time.Duration
	RefreshTokenExpiry time.Duration
	Stateful           bool
}

// NewWithConfig creates a new JWT client from a config struct
func NewWithConfig(config TokenConfig) (JWTClient, error) {
	opts := []Option{
		WithAccessTokenSecret(config.AccessTokenSecret),
		WithRefreshTokenSecret(config.RefreshTokenSecret),
		WithAccessTokenExpiry(config.AccessTokenExpiry),
		WithRefreshTokenExpiry(config.RefreshTokenExpiry),
		WithStateful(config.Stateful),
	}
	return New(opts...)
}
