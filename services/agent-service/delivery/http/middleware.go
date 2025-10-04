// Package http contains HTTP delivery implementations for the application
package http

import (
	"net/http"
	"time"

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
