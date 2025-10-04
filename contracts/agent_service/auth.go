// Package agent_service contains request and response contracts for the agent service
package agent_service

// LoginRequest represents the request payload for user login
type LoginRequest struct {
	Email    string `json:"email" validate:"required,email"`
	Password string `json:"password" validate:"required,min=8"`
}

// LoginResponse represents the response payload for user login
type LoginResponse struct {
	AccessToken        string `json:"access_token"`
	RefreshToken       string `json:"refresh_token"`
	AccessTokenExpire  int64  `json:"access_token_expire"`
	RefreshTokenExpire int64  `json:"refresh_token_expire"`
}

// RefreshTokenRequest represents the request payload for token refresh
type RefreshTokenRequest struct {
	RefreshToken string `json:"refresh_token" validate:"required"`
}

// RefreshTokenResponse represents the response payload for token refresh
type RefreshTokenResponse struct {
	AccessToken        string `json:"access_token"`
	RefreshToken       string `json:"refresh_token"`
	AccessTokenExpire  int64  `json:"access_token_expire"`
	RefreshTokenExpire int64  `json:"refresh_token_expire"`
}
