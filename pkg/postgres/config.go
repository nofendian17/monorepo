// Package postgres provides PostgreSQL database infrastructure components
package postgres

// Config holds the PostgreSQL database configuration
// It contains all the necessary parameters to establish a database connection
type Config struct {
	// Host specifies the database server host
	Host string
	// Port specifies the database server port
	Port int
	// User specifies the database user
	User string
	// Password specifies the database password
	Password string
	// DBName specifies the database name
	DBName string
	// Schema specifies the database schema
	Schema string
	// SSLMode specifies the SSL mode for database connection
	SSLMode string
	// MaxIdleConns specifies the maximum number of idle connections in the pool
	MaxIdleConns int
	// MaxOpenConns specifies the maximum number of open connections to the database
	MaxOpenConns int
	// ConnMaxIdleTime specifies the maximum amount of time a connection may be idle, in minutes
	ConnMaxIdleTime int
	// ConnMaxLifetime specifies the maximum amount of time a connection may be reused, in minutes
	ConnMaxLifetime int
	// Debug enables or disables debug mode for database operations
	Debug bool
	// ConnectTimeout specifies the connection timeout in seconds
	ConnectTimeout int
}
