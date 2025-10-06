package http

import (
	"monorepo/pkg/logger"
	"net/http"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
)

type Router struct {
	CredentialHandler *CredentialHandler
	SupplierHandler   *SupplierHandler
	HealthHandler     *HealthHandler
	AppLogger         logger.LoggerInterface
}

func NewRouter(credentialHandler *CredentialHandler, supplierHandler *SupplierHandler, healthHandler *HealthHandler, appLogger logger.LoggerInterface) *Router {
	return &Router{
		CredentialHandler: credentialHandler,
		SupplierHandler:   supplierHandler,
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
		// Protected routes - require X-AgentIATA-ID header
		api.Route("/", func(protected chi.Router) {
			protected.Use(AgentIATAMiddleware(r.AppLogger))

			// Suppliers routes - require authentication
			protected.Get("/suppliers", r.SupplierHandler.ListSuppliersHandler)

			// Credentials routes - require authentication
			protected.Route("/credentials", func(credentials chi.Router) {
				credentials.Post("/", r.CredentialHandler.CreateHandler)
				credentials.Get("/", r.CredentialHandler.ListHandler)
				credentials.Get("/{id}", r.CredentialHandler.GetByIDHandler)
				credentials.Put("/{id}", r.CredentialHandler.UpdateHandler)
				credentials.Delete("/{id}", r.CredentialHandler.DeleteHandler)
			})
		})
	})

	// Internal routes
	router.Route("/internal", func(internal chi.Router) {
		// Internal credentials route - no header validation required for internal calls
		internal.Get("/credentials", r.CredentialHandler.InternalListHandler)

		// Internal supplier routes - no header validation required for internal calls
		internal.Get("/supplier", r.SupplierHandler.ListSuppliersHandler)
		internal.Post("/supplier", r.SupplierHandler.CreateSupplierHandler)
		internal.Put("/supplier/{id}", r.SupplierHandler.UpdateSupplierHandler)
		internal.Delete("/supplier/{id}", r.SupplierHandler.DeleteSupplierHandler)
	})

	return router
}
