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
	router.Use(middleware.Heartbeat("/"))

	// Health check endpoint
	router.Get("/health", r.HealthHandler.HealthCheckHandler)

	// User routes
	router.Post("/users", r.Handler.CreateHandler)
	router.Get("/users", r.Handler.ListHandler)
	router.Get("/users/{id}", r.Handler.GetByIDHandler)
	router.Put("/users/{id}", r.Handler.UpdateHandler)
	router.Delete("/users/{id}", r.Handler.DeleteHandler)
	router.Get("/users/email/{email}", r.Handler.GetByEmailHandler)

	return router
}
