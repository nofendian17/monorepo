// Package postgres provides PostgreSQL database infrastructure components
package postgres

import (
	"fmt"
	"time"

	"gorm.io/driver/postgres"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

// PostgresClient defines the interface for PostgreSQL database operations
// It provides methods for database migration, getting the database instance, and closing connections
type PostgresClient interface {
	// Migrate runs auto-migration for all models
	// It takes optional model instances to migrate
	// Returns an error if the migration fails
	Migrate(dst ...any) error
	// GetDB returns the underlying gorm.DB instance
	// This allows direct access to the GORM database for custom operations
	GetDB() *gorm.DB
	// Close closes the database connection
	// Returns an error if closing the connection fails
	Close() error
}

// postgresClient manages database connections and operations
type postgresClient struct {
	// DB is the GORM database instance
	DB *gorm.DB
}

// NewPostgresClient creates a new database client based on the configuration
// It takes a Config struct with database connection parameters
// Returns a PostgresClient interface and an error if initialization fails
func NewPostgresClient(cfg Config) (PostgresClient, error) {
	dsn := fmt.Sprintf("host=%s port=%d user=%s password=%s dbname=%s search_path=%s sslmode=%s",
		cfg.Host, cfg.Port, cfg.User, cfg.Password, cfg.DBName, cfg.Schema, cfg.SSLMode)

	// Add connect timeout if specified
	if cfg.ConnectTimeout > 0 {
		dsn += fmt.Sprintf(" connect_timeout=%d", cfg.ConnectTimeout)
	}

	// Set appropriate log level based on config
	var loggerInterface logger.Interface
	if cfg.Debug {
		loggerInterface = logger.Default.LogMode(logger.Info)
	} else {
		loggerInterface = logger.Default.LogMode(logger.Silent)
	}

	// Open database connection with the configured logger
	db, err := gorm.Open(postgres.Open(dsn), &gorm.Config{
		Logger: loggerInterface,
	})
	if err != nil {
		return nil, err
	}

	// Configure connection pool
	dbSQL, err := db.DB()
	if err != nil {
		return nil, err
	}

	dbSQL.SetMaxIdleConns(cfg.MaxIdleConns)
	dbSQL.SetMaxOpenConns(cfg.MaxOpenConns)
	dbSQL.SetConnMaxIdleTime(time.Duration(cfg.ConnMaxIdleTime) * time.Minute)
	dbSQL.SetConnMaxLifetime(time.Duration(cfg.ConnMaxLifetime) * time.Minute)

	// Test the database connection
	if err := dbSQL.Ping(); err != nil {
		return nil, err
	}

	return &postgresClient{
		DB: db,
	}, nil
}

// Migrate runs auto-migration for all models
// Returns an error if the migration fails
func (c *postgresClient) Migrate(dst ...any) error {
	if err := c.DB.AutoMigrate(dst...); err != nil {
		return fmt.Errorf("failed to auto-migrate models: %w", err)
	}
	return nil
}

// GetDB returns the underlying gorm.DB instance
// This allows direct access to the GORM database for custom operations
func (c *postgresClient) GetDB() *gorm.DB {
	return c.DB
}

// Close closes the database connection
// Returns an error if closing the connection fails
func (c *postgresClient) Close() error {
	sqlDB, err := c.DB.DB()
	if err != nil {
		return err
	}
	return sqlDB.Close()
}
