// Package config handles application configuration loading and management
package config

import (
	"errors"
	"log"

	"github.com/spf13/viper"
)

// Config holds the entire application configuration
// It contains nested configurations for application, server, database, and security settings
type Config struct {
	// Application contains application-level settings
	Application ApplicationConfig `mapstructure:"application"`
	// Server contains HTTP server settings
	Server ServerConfig `mapstructure:"server"`
	// Infrastructure contains infrastructure connection settings
	Infrastructure InfrastructureConfig `mapstructure:"infrastructure"`
	// Security contains security-related settings
	Security SecurityConfig `mapstructure:"security"`
}

// ApplicationConfig holds the application-level configuration
// It contains settings that define the application's identity and behavior
type ApplicationConfig struct {
	// Name specifies the name of the application
	Name string `mapstructure:"name"`
	// Version specifies the version of the application
	Version string `mapstructure:"version"`
}

// ServerConfig holds the server configuration
// It contains settings for HTTP server behavior including timeouts and port
type ServerConfig struct {
	// Port specifies the port number the server will listen on
	Port int `mapstructure:"port"`
	// ReadTimeout defines the maximum duration for reading the entire request, including the body, in seconds
	ReadTimeout int `mapstructure:"read_timeout"` // in seconds
	// WriteTimeout defines the maximum duration before timing out writes of the response, in seconds
	WriteTimeout int `mapstructure:"write_timeout"` // seconds
	// ShutdownTimeout defines the maximum duration the server will wait for active connections to finish during shutdown, in seconds
	ShutdownTimeout int `mapstructure:"shutdown_timeout"` // seconds
}

// InfrastructureConfig holds the infrastructure configuration
// It contains settings for infrastructure connections like databases and message queues
type InfrastructureConfig struct {
	// Postgres contains PostgreSQL-specific settings
	Postgres PostgresConfig `mapstructure:"postgres"`
}

// SecurityConfig holds the security configuration
// It contains settings for security-related features like encryption
type SecurityConfig struct {
	// Encryption contains encryption settings
	Encryption EncryptionConfig `mapstructure:"encryption"`
}

// EncryptionConfig holds the encryption configuration
type EncryptionConfig struct {
	// Key is the encryption key for credentials
	Key string `mapstructure:"key"`
}

// PostgresConfig holds the PostgreSQL database configuration
// It contains all necessary parameters to establish a PostgreSQL connection
type PostgresConfig struct {
	// Host specifies the database server host
	Host string `mapstructure:"host"`
	// Port specifies the database server port
	Port int `mapstructure:"port"`
	// User specifies the database user
	User string `mapstructure:"user"`
	// Password specifies the database password
	Password string `mapstructure:"password"`
	// DBName specifies the database name
	DBName string `mapstructure:"dbname"`
	// Schema specifies the database schema
	Schema string `mapstructure:"schema"`
	// SSLMode specifies the SSL mode for database connection
	SSLMode string `mapstructure:"sslmode"`
	// MaxIdleConns specifies the maximum number of idle connections in the pool
	MaxIdleConns int `mapstructure:"max_idle_conns"`
	// MaxOpenConns specifies the maximum number of open connections to the database
	MaxOpenConns int `mapstructure:"max_open_conns"`
	// ConnMaxIdleTime specifies the maximum amount of time a connection may be idle, in minutes
	ConnMaxIdleTime int `mapstructure:"conn_max_idle_time"` // minutes
	// ConnMaxLifetime specifies the maximum amount of time a connection may be reused, in minutes
	ConnMaxLifetime int `mapstructure:"conn_max_lifetime"` // minutes
	// Debug enables or disables debug mode for database operations
	Debug bool `mapstructure:"debug"`
	// IsUseMigrate specifies whether to use database migration
	IsUseMigrate bool `mapstructure:"is_use_migrate"`
}

// LoadConfig loads the application configuration from various sources
// It first looks for a config.yaml file in the current directory and config directory
// If no config file is found, it uses environment variables and default values
// Returns a Config struct and an error if loading fails
func LoadConfig() (*Config, error) {
	viper.SetConfigName("supplier-credentials")
	viper.SetConfigType("yaml")
	viper.AddConfigPath("../configs")
	viper.AddConfigPath("../../configs")
	viper.AddConfigPath(".")
	viper.AddConfigPath("configs")

	// Set default values
	viper.SetDefault("server.port", "8081")
	viper.SetDefault("server.read_timeout", 15)     // seconds
	viper.SetDefault("server.write_timeout", 15)    // seconds
	viper.SetDefault("server.shutdown_timeout", 30) // seconds
	viper.SetDefault("infrastructure.postgres.host", "localhost")
	viper.SetDefault("infrastructure.postgres.port", 5432)
	// No defaults for user and password - they must be provided
	viper.SetDefault("infrastructure.postgres.dbname", "app_db")
	viper.SetDefault("infrastructure.postgres.schema", "public")
	viper.SetDefault("infrastructure.postgres.sslmode", "disable")
	viper.SetDefault("infrastructure.postgres.max_idle_conns", 10)
	viper.SetDefault("infrastructure.postgres.max_open_conns", 100)
	viper.SetDefault("infrastructure.postgres.conn_max_idle_time", 5) // minutes
	viper.SetDefault("infrastructure.postgres.conn_max_lifetime", 60) // minutes
	viper.SetDefault("infrastructure.postgres.debug", false)
	viper.SetDefault("application.name", "Supplier Credentials Service")
	viper.SetDefault("application.version", "1.0")
	viper.SetDefault("infrastructure.kafka.brokers", []string{"localhost:9092"})
	viper.SetDefault("infrastructure.kafka.topics.password_reset", "supplier-credentials.password.reset")

	if err := viper.ReadInConfig(); err != nil {
		var configFileNotFoundError viper.ConfigFileNotFoundError
		if errors.As(err, &configFileNotFoundError) {
			log.Println("Config file not found, using environment variables and defaults")
		}
	}

	var config Config
	if err := viper.Unmarshal(&config); err != nil {
		return nil, err
	}

	// Validate required secrets
	if config.Security.Encryption.Key == "" {
		return nil, errors.New("encryption key is required")
	}
	if config.Infrastructure.Postgres.User == "" {
		return nil, errors.New("database user is required")
	}
	if config.Infrastructure.Postgres.Password == "" {
		return nil, errors.New("database password is required")
	}

	return &config, nil
}

// GetConfigPath returns the path of the loaded config file
// If no config file was loaded, it returns an empty string
func GetConfigPath() string {
	return viper.ConfigFileUsed()
}
