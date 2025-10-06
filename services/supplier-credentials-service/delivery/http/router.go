package http

import (
	"monorepo/pkg/logger"
	"net/http"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
)

type Router struct {
	CredentialHandler *CredentialHandler
	HealthHandler     *HealthHandler
	AppLogger         logger.LoggerInterface
}

func NewRouter(credentialHandler *CredentialHandler, healthHandler *HealthHandler, appLogger logger.LoggerInterface) *Router {
	return &Router{
		CredentialHandler: credentialHandler,
		HealthHandler:     healthHandler,
		AppLogger:         appLogger,
	}
}

func (r *Router) SetupRoutes() http.Handler {
	router := chi.NewRouter()

	// Add middleware
	router.Use(middleware.Recoverer)
	router.Use(middleware.RequestID)
	router.Use(middleware.Heartbeat("/ping"))

	// Health check endpoint
	router.Get("/health", r.HealthHandler.HealthCheckHandler)

	router.Route("/api/v1", func(api chi.Router) {
		// Credentials routes - require X-AgentIATA-ID header
		api.Route("/credentials", func(credentials chi.Router) {
			credentials.Use(AgentIATAMiddleware(r.AppLogger))
			credentials.Post("/", r.CredentialHandler.CreateHandler)
			credentials.Get("/", r.CredentialHandler.ListHandler)
			credentials.Get("/{id}", r.CredentialHandler.GetByIDHandler)
			credentials.Put("/{id}", r.CredentialHandler.UpdateHandler)
			credentials.Delete("/{id}", r.CredentialHandler.DeleteHandler)
		})
	})

	// Internal routes
	router.Route("/internal", func(internal chi.Router) {
		// Internal credentials route - no header validation required for internal calls
		internal.Get("/credentials", r.CredentialHandler.InternalListHandler)
	})

	return router
}
