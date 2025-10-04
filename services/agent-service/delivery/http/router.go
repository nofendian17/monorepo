package http

import (
	"monorepo/pkg/logger"
	"net/http"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
)

type Router struct {
	Handler       *UserHandler
	AgentHandler  *AgentHandler
	HealthHandler *HealthHandler
	AppLogger     logger.LoggerInterface
}

func NewRouter(userHandler *UserHandler, agentHandler *AgentHandler, healthHandler *HealthHandler, appLogger logger.LoggerInterface) *Router {
	return &Router{
		Handler:       userHandler,
		AgentHandler:  agentHandler,
		HealthHandler: healthHandler,
		AppLogger:     appLogger,
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
		// You can add more middleware here if needed
		// User routes
		api.Route("/users", func(users chi.Router) {
			users.Post("/", r.Handler.CreateHandler)
			users.Get("/", r.Handler.ListHandler)
			users.Get("/{id}", r.Handler.GetByIDHandler)
			users.Put("/{id}", r.Handler.UpdateHandler)
			users.Patch("/{id}/status", r.Handler.UpdateStatusHandler)
			users.Delete("/{id}", r.Handler.DeleteHandler)
			users.Get("/email/{email}", r.Handler.GetByEmailHandler)
		})
		// Agent routes
		api.Route("/agents", func(agents chi.Router) {
			agents.Post("/", r.AgentHandler.CreateHandler)
			agents.Get("/", r.AgentHandler.ListHandler)
			agents.Get("/{id}", r.AgentHandler.GetByIDHandler)
			agents.Put("/{id}", r.AgentHandler.UpdateHandler)
			agents.Patch("/{id}/status", r.AgentHandler.UpdateStatusHandler)
			agents.Delete("/{id}", r.AgentHandler.DeleteHandler)
			agents.Get("/active", r.AgentHandler.GetActiveHandler)
			agents.Get("/inactive", r.AgentHandler.GetInactiveHandler)
		})
	})
	return router
}
