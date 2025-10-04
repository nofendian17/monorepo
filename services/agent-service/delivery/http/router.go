package http

import (
	"monorepo/pkg/logger"
	"net/http"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
)

type Router struct {
	Handler       *UserHandler
	HealthHandler *HealthHandler
	AppLogger     logger.LoggerInterface
}

func NewRouter(userHandler *UserHandler, healthHandler *HealthHandler, appLogger logger.LoggerInterface) *Router {
	return &Router{
		Handler:       userHandler,
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
			users.Delete("/{id}", r.Handler.DeleteHandler)
			users.Get("/email/{email}", r.Handler.GetByEmailHandler)
		})
	})
	return router
}
