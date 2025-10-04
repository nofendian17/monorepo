package jwt

import (
	"time"
)

// Option is a function that configures TokenConfig
type Option func(*TokenConfig)

// WithAccessTokenSecret sets the access token secret
func WithAccessTokenSecret(secret string) Option {
	return func(c *TokenConfig) {
		c.AccessTokenSecret = secret
	}
}

// WithRefreshTokenSecret sets the refresh token secret
func WithRefreshTokenSecret(secret string) Option {
	return func(c *TokenConfig) {
		c.RefreshTokenSecret = secret
	}
}

// WithAccessTokenExpiry sets the access token expiry duration
func WithAccessTokenExpiry(expiry time.Duration) Option {
	return func(c *TokenConfig) {
		c.AccessTokenExpiry = expiry
	}
}

// WithRefreshTokenExpiry sets the refresh token expiry duration
func WithRefreshTokenExpiry(expiry time.Duration) Option {
	return func(c *TokenConfig) {
		c.RefreshTokenExpiry = expiry
	}
}

// WithStateful enables stateful mode
func WithStateful(stateful bool) Option {
	return func(c *TokenConfig) {
		c.Stateful = stateful
	}
}
