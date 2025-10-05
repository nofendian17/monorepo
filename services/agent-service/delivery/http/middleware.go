// Package http contains HTTP delivery implementations for the application
package http

import (
	"agent-service/domain/model"
	"context"
	"net/http"
	"time"

	"monorepo/pkg/api"
	"monorepo/pkg/jwt"
	"monorepo/pkg/logger"

	"github.com/go-chi/chi/v5/middleware"
)

// LoggingMiddleware adds detailed request logging
// It takes a logger instance and returns a middleware function
// The middleware logs information about each HTTP request including method, path, status, duration, and client information
func LoggingMiddleware(logger logger.LoggerInterface) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			start := time.Now()
			ww := middleware.NewWrapResponseWriter(w, r.ProtoMajor)

			next.ServeHTTP(ww, r)

			logger.InfoContext(r.Context(), "HTTP request completed",
				"method", r.Method,
				"path", r.URL.Path,
				"status", ww.Status(),
				"duration", time.Since(start).String(),
				"remote_addr", r.RemoteAddr,
				"user_agent", r.UserAgent(),
			)
		})
	}
}

// JWTMiddleware validates JWT tokens for protected routes
// It extracts the Authorization header, validates the token, and adds user claims to the request context
// Returns a 401 status code for missing or invalid tokens
func JWTMiddleware(jwtClient jwt.JWTClient, logger logger.LoggerInterface, apiClient api.Api) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			ctx := r.Context()

			// Extract token from Authorization header
			authHeader := r.Header.Get("Authorization")
			if authHeader == "" {
				logger.WarnContext(ctx, "Missing Authorization header")
				apiClient.Unauthorized(ctx, w, "Missing Authorization header")
				return
			}

			// Check for Bearer token format
			const bearerPrefix = "Bearer "
			if len(authHeader) <= len(bearerPrefix) || authHeader[:len(bearerPrefix)] != bearerPrefix {
				logger.WarnContext(ctx, "Invalid Authorization header format")
				apiClient.Unauthorized(ctx, w, "Invalid Authorization header format")
				return
			}

			tokenString := authHeader[len(bearerPrefix):]

			// Validate the access token
			claims, err := jwtClient.ValidateAccessToken(tokenString)
			if err != nil {
				logger.WarnContext(ctx, "Invalid access token", "error", err)
				apiClient.Unauthorized(ctx, w, "Invalid access token")
				return
			}

			// Add claims to context for use in handlers
			ctx = context.WithValue(ctx, "user_id", claims.UserID)
			ctx = context.WithValue(ctx, "agent_id", claims.AgentID)
			ctx = context.WithValue(ctx, "agent_type", claims.AgentType)

			// Update request with new context
			r = r.WithContext(ctx)

			next.ServeHTTP(w, r)
		})
	}
}

// AgentTypeMiddleware validates that the JWT token has the specified agent_type
// It should be used after JWTMiddleware
// Returns a 403 status code if the agent type does not match the required type
func AgentTypeMiddleware(requiredAgentType string, logger logger.LoggerInterface, apiClient api.Api) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			ctx := r.Context()

			// Get agent_type from context (set by JWTMiddleware)
			agentType, ok := ctx.Value("agent_type").(string)
			if !ok || agentType != requiredAgentType {
				logger.WarnContext(ctx, "Access denied: agent type does not match required type", "agent_type", agentType, "required_type", requiredAgentType)
				apiClient.Forbidden(ctx, w, "Access denied: insufficient agent permissions")
				return
			}

			next.ServeHTTP(w, r)
		})
	}
}

// IATAAgentMiddleware validates that the JWT token has agent_type = "IATA"
// It should be used after JWTMiddleware
// Returns a 403 status code if the agent type is not IATA
func IATAAgentMiddleware(logger logger.LoggerInterface, apiClient api.Api) func(http.Handler) http.Handler {
	return AgentTypeMiddleware(model.AgentTypeIATA, logger, apiClient)
}
