// Package http contains HTTP delivery implementations for the application
package http

import (
	"net/http"

	"monorepo/pkg/api"
	"monorepo/pkg/logger"
)

// HealthHandler handles health check requests
type HealthHandler struct {
	// Logger is used for logging operations within the handler
	Logger logger.LoggerInterface
	// API provides standardized API response patterns
	API api.Api
}

// NewHealthHandler creates a new instance of HealthHandler
func NewHealthHandler(logger logger.LoggerInterface) *HealthHandler {
	return &HealthHandler{
		Logger: logger,
		API:    api.New(),
	}
}

// HealthCheckHandler handles HTTP requests for health checks
// It returns a JSON response indicating the service status
func (h *HealthHandler) HealthCheckHandler(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	h.Logger.InfoContext(ctx, "Health check endpoint called")

	healthData := map[string]interface{}{
		"status":  "healthy",
		"message": "Service is running",
	}

	h.API.Success(ctx, w, healthData)
}
