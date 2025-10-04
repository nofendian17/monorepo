// Package main is the entry point for the application
// It initializes all components and starts the HTTP server
package main

import (
	"context"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"agent-service/config"
	httpDelivery "agent-service/delivery/http"
	"agent-service/domain/model"
	pgRepository "agent-service/repository/postgres"
	"agent-service/usecase"
	"monorepo/pkg/logger"
	"monorepo/pkg/postgres"
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

	// Initialize database client
	postgresClient, err := postgres.NewPostgresClient(postgres.Config{
		Host:            cfg.Database.Postgres.Host,
		Port:            cfg.Database.Postgres.Port,
		User:            cfg.Database.Postgres.User,
		Password:        cfg.Database.Postgres.Password,
		DBName:          cfg.Database.Postgres.DBName,
		Schema:          cfg.Database.Postgres.Schema,
		SSLMode:         cfg.Database.Postgres.SSLMode,
		MaxIdleConns:    cfg.Database.Postgres.MaxIdleConns,
		MaxOpenConns:    cfg.Database.Postgres.MaxOpenConns,
		ConnMaxIdleTime: cfg.Database.Postgres.ConnMaxIdleTime,
		ConnMaxLifetime: cfg.Database.Postgres.ConnMaxLifetime,
		Debug:           cfg.Database.Postgres.Debug,
	})
	if err != nil {
		appLogger.Error("Failed to connect to database", "error", err)
		os.Exit(1)
	}

	if cfg.Database.Postgres.IsUseMigrate {
		// Run database migrations
		err = postgresClient.Migrate(
			&model.User{},
			&model.Agent{},
		)
		if err != nil {
			appLogger.Error("Failed to migrate database", "error", err)
			os.Exit(1)
		}
	}

	// Initialize repository
	userRepo := pgRepository.NewUserRepository(postgresClient.GetDB(), appLogger)
	agentRepo := pgRepository.NewAgentRepository(postgresClient.GetDB(), appLogger)

	// Initialize usecase
	userUsecase := usecase.NewUserUseCase(userRepo, appLogger)
	agentUsecase := usecase.NewAgentUseCase(agentRepo, appLogger)

	// Initialize handlers
	userHandler := httpDelivery.NewUserHandler(userUsecase, appLogger)
	agentHandler := httpDelivery.NewAgentHandler(agentUsecase, appLogger)
	healthHandler := httpDelivery.NewHealthHandler(appLogger)

	// Initialize router
	router := httpDelivery.NewRouter(userHandler, agentHandler, healthHandler, appLogger)

	// Setup routes
	httpHandler := router.SetupRoutes()

	// Start server
	server := &http.Server{
		Addr:         ":" + cfg.Server.Port,
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
