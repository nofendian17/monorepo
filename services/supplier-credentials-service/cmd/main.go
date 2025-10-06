// Package main is the entry point for the application
// It initializes all components and starts the HTTP server
package main

import (
	"context"
	"net/http"
	"os"
	"os/signal"
	"strconv"
	"syscall"
	"time"

	"monorepo/pkg/logger"
	"monorepo/pkg/postgres"
	"supplier-credentials-service/config"
	httpDelivery "supplier-credentials-service/delivery/http"
	"supplier-credentials-service/domain/model"
	pgRepository "supplier-credentials-service/repository/postgres"
	"supplier-credentials-service/usecase"
)

// main is the entry point of the application
// It performs the following steps:
// 1. Initializes the logger
// 2. Loads configuration from files or environment variables
// 3. Sets up the database connection
// 4. Runs database migrations
// 5. Initializes the repository, usecase, and handler layers
// 6. Sets up HTTP routes
// 7. Starts the HTTP server with graceful shutdown
func main() {
	// configure logger
	appLogger := logger.NewJSONDefault()

	// Load configuration
	cfg, err := config.LoadConfig()
	if err != nil {
		appLogger.Error("Failed to load configuration", "error", err)
		os.Exit(1)
	}

	// Initialize PostgreSQL client
	postgresClient, err := postgres.NewPostgresClient(postgres.Config{
		Host:            cfg.Infrastructure.Postgres.Host,
		Port:            cfg.Infrastructure.Postgres.Port,
		User:            cfg.Infrastructure.Postgres.User,
		Password:        cfg.Infrastructure.Postgres.Password,
		DBName:          cfg.Infrastructure.Postgres.DBName,
		Schema:          cfg.Infrastructure.Postgres.Schema,
		SSLMode:         cfg.Infrastructure.Postgres.SSLMode,
		MaxIdleConns:    cfg.Infrastructure.Postgres.MaxIdleConns,
		MaxOpenConns:    cfg.Infrastructure.Postgres.MaxOpenConns,
		ConnMaxIdleTime: cfg.Infrastructure.Postgres.ConnMaxIdleTime,
		ConnMaxLifetime: cfg.Infrastructure.Postgres.ConnMaxLifetime,
		Debug:           cfg.Infrastructure.Postgres.Debug,
	})
	if err != nil {
		appLogger.Error("Failed to connect to database", "error", err)
		os.Exit(1)
	}

	if cfg.Infrastructure.Postgres.IsUseMigrate {
		// Run database migrations
		err = postgresClient.Migrate(
			&model.Supplier{},
			&model.AgentSupplierCredential{},
		)
		if err != nil {
			appLogger.Error("Failed to migrate database", "error", err)
			os.Exit(1)
		}
	}

	// Initialize repository
	supplierRepo := pgRepository.NewSupplierRepository(postgresClient.GetDB(), appLogger)
	credentialRepo := pgRepository.NewCredentialRepository(postgresClient.GetDB(), appLogger)

	// Initialize usecase
	supplierUsecase := usecase.NewSupplierUseCase(supplierRepo, appLogger)
	credentialUsecase := usecase.NewCredentialUseCase(credentialRepo, supplierUsecase, appLogger, cfg.Security.Encryption.Key)

	// Initialize handlers
	credentialHandler := httpDelivery.NewCredentialHandler(credentialUsecase, appLogger)
	supplierHandler := httpDelivery.NewSupplierHandler(supplierUsecase, appLogger)
	healthHandler := httpDelivery.NewHealthHandler(appLogger)

	// Initialize router
	router := httpDelivery.NewRouter(credentialHandler, supplierHandler, healthHandler, appLogger)

	// Setup routes
	httpHandler := router.SetupRoutes()

	// Start server
	server := &http.Server{
		Addr:         ":" + strconv.Itoa(cfg.Server.Port),
		Handler:      httpHandler,
		ReadTimeout:  time.Duration(cfg.Server.ReadTimeout) * time.Second,
		WriteTimeout: time.Duration(cfg.Server.WriteTimeout) * time.Second,
	}

	// Create channel to listen for interrupt signal
	quit := make(chan os.Signal, 1)

	// Register the channel to receive specific signals
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)

	// Start HTTP server in a separate goroutine
	go func() {
		appLogger.Info("Service starting", "name", cfg.Application.Name, "version", cfg.Application.Version, "port", cfg.Server.Port)
		if err := server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			appLogger.Error("Failed to start server", "error", err)
			os.Exit(1)
		}
	}()

	// Block until a signal is received
	<-quit
	appLogger.Info("Shutting down server...")

	// Create a context with timeout for graceful shutdown
	ctx, cancel := context.WithTimeout(context.Background(), time.Duration(cfg.Server.ShutdownTimeout)*time.Second)
	defer cancel()

	// Shutdown the server gracefully
	if err := server.Shutdown(ctx); err != nil {
		appLogger.Error("Server forced to shutdown", "error", err)
		os.Exit(1)
	}

	// Close database connection
	if err := postgresClient.Close(); err != nil {
		appLogger.Warn("Error closing database connection", "error", err)
	}

	appLogger.Info("Server exited")
}
