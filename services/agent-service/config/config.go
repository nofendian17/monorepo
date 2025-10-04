// Package config handles application configuration loading and management
package config

import (
	"errors"
	"log"

	"github.com/spf13/viper"
)

// Config holds the entire application configuration
// It contains nested configurations for application, server, and database settings
type Config struct {
	// Application contains application-level settings
	Application ApplicationConfig `mapstructure:"application"`
	// Server contains HTTP server settings
	Server ServerConfig `mapstructure:"server"`
	// Database contains database connection settings
	Database DatabaseConfig `mapstructure:"database"`
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
	Port string `mapstructure:"port"`
	// ReadTimeout defines the maximum duration for reading the entire request, including the body, in seconds
	ReadTimeout int `mapstructure:"read_timeout"` // in seconds
	// WriteTimeout defines the maximum duration before timing out writes of the response, in seconds
	WriteTimeout int `mapstructure:"write_timeout"` // in seconds
	// ShutdownTimeout defines the maximum duration the server will wait for active connections to finish during shutdown, in seconds
	ShutdownTimeout int `mapstructure:"shutdown_timeout"` // in seconds
}

// DatabaseConfig holds the database configuration
// It contains settings for database connections, currently only PostgreSQL
type DatabaseConfig struct {
	// Postgres contains PostgreSQL-specific settings
	Postgres PostgresConfig `mapstructure:"postgres"`
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
	ConnMaxIdleTime int `mapstructure:"conn_max_idle_time"` // in minutes
	// ConnMaxLifetime specifies the maximum amount of time a connection may be reused, in minutes
	ConnMaxLifetime int `mapstructure:"conn_max_lifetime"` // in minutes
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
	viper.SetConfigName("agent")
	viper.SetConfigType("yaml")
	viper.AddConfigPath("../configs")
	viper.AddConfigPath("../../configs")
	viper.AddConfigPath(".")
	viper.AddConfigPath("configs")

	// Set default values
	viper.SetDefault("server.port", "8080")
	viper.SetDefault("server.read_timeout", 15)     // seconds
	viper.SetDefault("server.write_timeout", 15)    // seconds
	viper.SetDefault("server.shutdown_timeout", 30) // seconds
	viper.SetDefault("database.postgres.host", "localhost")
	viper.SetDefault("database.postgres.port", 5432)
	viper.SetDefault("database.postgres.user", "postgres")
	viper.SetDefault("database.postgres.password", "password")
	viper.SetDefault("database.postgres.dbname", "app_db")
	viper.SetDefault("database.postgres.schema", "public")
	viper.SetDefault("database.postgres.sslmode", "disable")
	viper.SetDefault("database.postgres.max_idle_conns", 10)
	viper.SetDefault("database.postgres.max_open_conns", 100)
	viper.SetDefault("database.postgres.conn_max_idle_time", 5) // minutes
	viper.SetDefault("database.postgres.conn_max_lifetime", 60) // minutes
	viper.SetDefault("database.postgres.debug", false)
	viper.SetDefault("application.name", "Application Service")
	viper.SetDefault("application.version", "1.0")

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

	return &config, nil
}

// GetConfigPath returns the path of the loaded config file
// If no config file was loaded, it returns an empty string
func GetConfigPath() string {
	return viper.ConfigFileUsed()
}
