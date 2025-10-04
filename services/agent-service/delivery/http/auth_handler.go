// Package http contains HTTP delivery implementations for the application
package http

import (
	"encoding/json"
	"net"
	"net/http"
	"strings"

	"agent-service/domain"
	"agent-service/usecase"
	"monorepo/contracts/agent_service"
	"monorepo/pkg/api"
	"monorepo/pkg/logger"
	"monorepo/pkg/validator"
)

// AuthHandler handles HTTP requests for authentication operations
type AuthHandler struct {
	// AuthUseCase contains business logic for authentication operations
	AuthUseCase usecase.AuthUseCase
	// Logger is used for logging operations within the handler
	Logger logger.LoggerInterface
	// API provides standardized API response patterns
	API api.Api
}

// NewAuthHandler creates a new instance of AuthHandler
// It takes an AuthUseCase implementation and a logger instance
// Returns a pointer to an AuthHandler
func NewAuthHandler(authUseCase usecase.AuthUseCase, logger logger.LoggerInterface) *AuthHandler {
	return &AuthHandler{
		AuthUseCase: authUseCase,
		Logger:      logger,
		API:         api.New(),
	}
}

// LoginHandler handles HTTP requests for user login
// It expects a JSON payload with email and password in the request body
// Returns a 200 status code with access and refresh tokens on success
// Returns a 400 status code for invalid request data
// Returns a 401 status code for invalid credentials
// Returns a 500 status code for internal server errors
func (h *AuthHandler) LoginHandler(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	h.Logger.InfoContext(ctx, "Login handler called")

	var req agent_service.LoginRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		h.Logger.ErrorContext(ctx, "Failed to decode login request", "error", err)
		h.API.BadRequest(ctx, w, "Invalid request body")
		return
	}

	// Validate request
	if validationErrors := validator.ValidateStruct(req); validationErrors != nil {
		h.Logger.WarnContext(ctx, "Validation failed for login request", "errors", validationErrors)
		h.API.ValidationError(ctx, w, h.convertValidationErrors(validationErrors))
		return
	}

	// Extract session information from request
	userAgent := r.Header.Get("User-Agent")
	ipAddress := h.getClientIP(r)

	// Call usecase with session information
	response, err := h.AuthUseCase.Login(ctx, req, userAgent, ipAddress)
	if err != nil {
		h.Logger.WarnContext(ctx, "Login failed", "email", req.Email, "error", err)

		// Check if it's a domain error with status code
		if appErr, ok := err.(*domain.AppError); ok {
			switch appErr.Code {
			case 401:
				h.API.Unauthorized(ctx, w, appErr.Message)
			default:
				h.API.BadRequest(ctx, w, appErr.Message)
			}
			return
		}

		// Generic error
		h.API.InternalServerError(ctx, w, "Login failed")
		return
	}

	h.Logger.InfoContext(ctx, "Login successful")
	h.API.Success(ctx, w, response)
}

// RefreshHandler handles HTTP requests for token refresh
// It expects a JSON payload with refresh_token in the request body
// Returns a 200 status code with new access token on success
// Returns a 400 status code for invalid request data
// Returns a 401 status code for invalid refresh token
// Returns a 500 status code for internal server errors
func (h *AuthHandler) RefreshHandler(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	h.Logger.InfoContext(ctx, "Refresh token handler called")

	var req agent_service.RefreshTokenRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		h.Logger.ErrorContext(ctx, "Failed to decode refresh request", "error", err)
		h.API.BadRequest(ctx, w, "Invalid request body")
		return
	}

	// Validate request
	if validationErrors := validator.ValidateStruct(req); validationErrors != nil {
		h.Logger.WarnContext(ctx, "Validation failed for refresh request", "errors", validationErrors)
		h.API.ValidationError(ctx, w, h.convertValidationErrors(validationErrors))
		return
	}

	// Call usecase
	response, err := h.AuthUseCase.Refresh(ctx, req)
	if err != nil {
		h.Logger.WarnContext(ctx, "Token refresh failed", "error", err)

		// Check if it's a domain error with status code
		if appErr, ok := err.(*domain.AppError); ok {
			switch appErr.Code {
			case 401:
				h.API.Unauthorized(ctx, w, appErr.Message)
			default:
				h.API.BadRequest(ctx, w, appErr.Message)
			}
			return
		}

		// Generic error
		h.API.InternalServerError(ctx, w, "Token refresh failed")
		return
	}

	h.Logger.InfoContext(ctx, "Token refresh successful")
	h.API.Success(ctx, w, response)
}

// ProfileHandler handles HTTP requests for authenticated user profile
// It retrieves the user profile information from the authenticated user's context
// Returns a 200 status code with user profile data on success
// Returns a 401 status code for unauthorized access
// Returns a 404 status code if user is not found
// Returns a 500 status code for internal server errors
func (h *AuthHandler) ProfileHandler(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	h.Logger.InfoContext(ctx, "Profile handler called")

	// Call usecase (user ID is extracted from JWT middleware)
	response, err := h.AuthUseCase.Profile(ctx)
	if err != nil {
		h.Logger.WarnContext(ctx, "Profile retrieval failed", "error", err)

		// Check if it's a domain error with status code
		if appErr, ok := err.(*domain.AppError); ok {
			switch appErr.Code {
			case 401:
				h.API.Unauthorized(ctx, w, appErr.Message)
			case 404:
				h.API.NotFound(ctx, w, appErr.Message)
			default:
				h.API.BadRequest(ctx, w, appErr.Message)
			}
			return
		}

		// Check for specific error messages
		if err.Error() == "unauthorized: user ID not found" {
			h.API.Unauthorized(ctx, w, "Unauthorized")
			return
		}

		// Generic error
		h.API.InternalServerError(ctx, w, "Profile retrieval failed")
		return
	}

	h.Logger.InfoContext(ctx, "Profile retrieved successfully")
	h.API.Success(ctx, w, response)
}

// convertValidationErrors converts validator errors to API error details
func (h *AuthHandler) convertValidationErrors(validationErrors map[string]string) []api.ErrorDetail {
	details := make([]api.ErrorDetail, 0, len(validationErrors))
	for field, message := range validationErrors {
		details = append(details, api.ErrorDetail{
			Field:   field,
			Message: message,
		})
	}
	return details
}

// getClientIP extracts the real client IP address from the request
// It checks X-Forwarded-For, X-Real-IP headers, and falls back to RemoteAddr
func (h *AuthHandler) getClientIP(r *http.Request) string {
	// Check X-Forwarded-For header (common with proxies/load balancers)
	xff := r.Header.Get("X-Forwarded-For")
	if xff != "" {
		// X-Forwarded-For can contain multiple IPs, take the first one
		if idx := strings.Index(xff, ","); idx > 0 {
			return strings.TrimSpace(xff[:idx])
		}
		return strings.TrimSpace(xff)
	}

	// Check X-Real-IP header (used by some proxies)
	xri := r.Header.Get("X-Real-IP")
	if xri != "" {
		return strings.TrimSpace(xri)
	}

	// Fall back to RemoteAddr
	ip, _, err := net.SplitHostPort(r.RemoteAddr)
	if err != nil {
		return r.RemoteAddr
	}
	return ip
}
