// Package http contains HTTP delivery implementations for the application
package http

import (
	"net/http"

	"monorepo/pkg/api"
	"monorepo/pkg/logger"
)

// HealthHandler handles HTTP requests for health check operations
type HealthHandler struct {
	// Logger is used for logging operations within the handler
	Logger logger.LoggerInterface
	// API provides standardized API response patterns
	API api.Api
}

// NewHealthHandler creates a new instance of HealthHandler
// It takes a logger instance
// Returns a pointer to a HealthHandler
func NewHealthHandler(appLogger logger.LoggerInterface) *HealthHandler {
	return &HealthHandler{
		Logger: appLogger,
		API:    api.New(),
	}
}

// HealthCheckHandler handles HTTP requests to check the health of the service
// It returns a JSON response indicating the service status
// Returns a 200 status code with health information
func (h *HealthHandler) HealthCheckHandler(w http.ResponseWriter, req *http.Request) {
	ctx := req.Context()
	h.Logger.InfoContext(ctx, "Health check endpoint called")

	healthData := map[string]interface{}{
		"status":  "healthy",
		"message": "Service is running",
	}

	h.API.Success(ctx, w, healthData)
}
