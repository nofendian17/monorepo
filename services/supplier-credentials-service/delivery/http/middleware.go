// Package http contains HTTP delivery implementations for the application
package http

import (
	"context"
	"monorepo/pkg/api"
	"monorepo/pkg/logger"
	"net/http"
	"strings"
)

// AgentIATAMiddleware validates the presence and validity of the X-AgentIATA-ID header
// It ensures that only requests with a valid IATA agent ID can access credential-related endpoints
func AgentIATAMiddleware(logger logger.LoggerInterface) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			ctx := r.Context()

			// Extract X-AgentIATA-ID header
			agentIATAID := r.Header.Get("X-AgentIATA-ID")
			if agentIATAID == "" {
				logger.WarnContext(ctx, "Missing X-AgentIATA-ID header")
				apiClient := api.New()
				apiClient.BadRequest(ctx, w, "X-AgentIATA-ID header is required")
				return
			}

			// Validate that the header is not empty (trimmed)
			if len(strings.TrimSpace(agentIATAID)) == 0 {
				logger.WarnContext(ctx, "Empty X-AgentIATA-ID header")
				apiClient := api.New()
				apiClient.BadRequest(ctx, w, "X-AgentIATA-ID header cannot be empty")
				return
			}

			// Add the agent IATA ID to context for potential use in handlers
			ctx = context.WithValue(ctx, "agent_iata_id", agentIATAID)
			r = r.WithContext(ctx)

			next.ServeHTTP(w, r)
		})
	}
}
