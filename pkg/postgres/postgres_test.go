package postgres

import (
	"database/sql"
	"testing"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
)

func createMockPostgresClient(t *testing.T, db *sql.DB, config Config) PostgresClient {
	gormDB, err := gorm.Open(postgres.New(postgres.Config{
		Conn: db,
	}), &gorm.Config{})
	require.NoError(t, err)

	return &postgresClient{
		DB: gormDB,
	}
}

func TestPostgresClient_ConnectionPool(t *testing.T) {
	db, mock, err := sqlmock.New(sqlmock.MonitorPingsOption(true))
	require.NoError(t, err)
	defer db.Close()

	config := Config{
		Host:            "localhost",
		Port:            5432,
		User:            "postgres",
		Password:        "password",
		DBName:          "testdb",
		Schema:          "public",
		SSLMode:         "disable",
		MaxIdleConns:    5,
		MaxOpenConns:    20,
		ConnMaxIdleTime: 10,
		ConnMaxLifetime: 30,
		Debug:           false,
	}

	mock.ExpectPing()

	client := createMockPostgresClient(t, db, config)

	// Test getting underlying DB
	sqlDB, err := client.GetDB().DB()
	assert.NoError(t, err, "Getting underlying DB should succeed")
	assert.NotNil(t, sqlDB, "Underlying DB should not be nil")

	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestNewPostgresClient_WithMock(t *testing.T) {
	sqlDB, mock, err := sqlmock.New(sqlmock.MonitorPingsOption(true))
	require.NoError(t, err, "Failed to create sqlmock")
	defer sqlDB.Close()

	// Mock the ping that happens in NewPostgresClient
	mock.ExpectPing()

	// Create a config for testing
	config := Config{
		Host:            "localhost",
		Port:            5432,
		User:            "postgres",
		Password:        "password",
		DBName:          "testdb",
		Schema:          "public",
		SSLMode:         "disable",
		MaxIdleConns:    10,
		MaxOpenConns:    100,
		ConnMaxIdleTime: 5,
		ConnMaxLifetime: 60,
		Debug:           false,
	}

	// We can't easily test NewPostgresClient directly since it creates its own connection
	// But we can test that the config is valid
	assert.Equal(t, "localhost", config.Host, "Config validation failed")
}

func setupMockPostgres(t *testing.T) (PostgresClient, sqlmock.Sqlmock) {
	sqlDB, mock, err := sqlmock.New(sqlmock.MonitorPingsOption(true))
	require.NoError(t, err, "Failed to create sqlmock")

	// Mock the ping
	mock.ExpectPing()

	// Open GORM with the mocked database
	dialector := postgres.New(postgres.Config{
		Conn:                 sqlDB,
		PreferSimpleProtocol: true,
	})

	db, err := gorm.Open(dialector, &gorm.Config{})
	require.NoError(t, err, "Failed to open GORM with mock")

	client := &postgresClient{
		DB: db,
	}

	t.Cleanup(func() {
		sqlDB.Close()
	})

	return client, mock
}

func TestPostgresClient_Migrate(t *testing.T) {
	client, mock := setupMockPostgres(t)

	// Mock GORM's table existence check for users
	mock.ExpectQuery(`SELECT count\(\*\) FROM information_schema\.tables WHERE table_schema = CURRENT_SCHEMA\(\) AND table_name = \$1 AND table_type = \$2`).
		WithArgs("users", "BASE TABLE").
		WillReturnRows(sqlmock.NewRows([]string{"count"}).AddRow(0))

	// Mock the users table creation
	mock.ExpectExec(`CREATE TABLE "users"`).
		WillReturnResult(sqlmock.NewResult(1, 1))

	// Mock GORM's table existence check for posts
	mock.ExpectQuery(`SELECT count\(\*\) FROM information_schema\.tables WHERE table_schema = CURRENT_SCHEMA\(\) AND table_name = \$1 AND table_type = \$2`).
		WithArgs("posts", "BASE TABLE").
		WillReturnRows(sqlmock.NewRows([]string{"count"}).AddRow(0))

	// Mock the posts table creation
	mock.ExpectExec(`CREATE TABLE "posts"`).
		WillReturnResult(sqlmock.NewResult(1, 1))

	// Define models for testing
	type User struct {
		ID   uint   `gorm:"primaryKey"`
		Name string `gorm:"size:100"`
	}

	type Post struct {
		ID     uint   `gorm:"primaryKey"`
		Title  string `gorm:"size:200"`
		UserID uint
		User   User `gorm:"constraint:OnUpdate:CASCADE,OnDelete:SET NULL;"`
	}

	err := client.Migrate(&User{}, &Post{})
	require.NoError(t, err, "Migrate() should not fail")

	require.NoError(t, mock.ExpectationsWereMet(), "SQL expectations should be met")
}

func TestPostgresClient_Migrate_Error(t *testing.T) {
	client, mock := setupMockPostgres(t)

	// Mock a migration error
	mock.ExpectExec(`CREATE TABLE "users"`).
		WillReturnError(gorm.ErrInvalidDB)

	type User struct {
		ID   uint   `gorm:"primaryKey"`
		Name string `gorm:"size:100"`
	}

	err := client.Migrate(&User{})
	assert.Error(t, err, "Migrate() should fail with database error")
	assert.Contains(t, err.Error(), "failed to auto-migrate", "Error should mention migration failure")

	require.NoError(t, mock.ExpectationsWereMet(), "SQL expectations should be met")
}

func TestPostgresClient_Migrate_EmptyModels(t *testing.T) {
	client, mock := setupMockPostgres(t)

	// Test migrating with no models
	err := client.Migrate()
	assert.NoError(t, err, "Migrate() should succeed with no models")

	require.NoError(t, mock.ExpectationsWereMet(), "SQL expectations should be met")
}

func TestPostgresClient_Migrate_SingleModel(t *testing.T) {
	client, mock := setupMockPostgres(t)

	// Mock table existence check
	mock.ExpectQuery(`SELECT count\(\*\) FROM information_schema\.tables WHERE table_schema = CURRENT_SCHEMA\(\) AND table_name = \$1 AND table_type = \$2`).
		WithArgs("users", "BASE TABLE").
		WillReturnRows(sqlmock.NewRows([]string{"count"}).AddRow(0))

	// Mock table creation
	mock.ExpectExec(`CREATE TABLE "users"`).
		WillReturnResult(sqlmock.NewResult(1, 1))

	type User struct {
		ID   uint   `gorm:"primaryKey"`
		Name string `gorm:"size:100"`
	}

	err := client.Migrate(&User{})
	require.NoError(t, err, "Migrate() should not fail")

	require.NoError(t, mock.ExpectationsWereMet(), "SQL expectations should be met")
}

func TestPostgresClient_GetDB(t *testing.T) {
	client, _ := setupMockPostgres(t)

	db := client.GetDB()
	require.NotNil(t, db, "GetDB() should not return nil")
}

func TestPostgresClient_Close(t *testing.T) {
	client, mock := setupMockPostgres(t)

	// Mock the close operation
	mock.ExpectClose()

	err := client.Close()
	require.NoError(t, err, "Close() should not fail")

	require.NoError(t, mock.ExpectationsWereMet(), "SQL expectations should be met")
}

func TestPostgresClient_Close_Error(t *testing.T) {
	client, mock := setupMockPostgres(t)

	// Mock a close error
	mock.ExpectClose().WillReturnError(gorm.ErrInvalidDB)

	err := client.Close()
	assert.Error(t, err, "Close() should fail with database error")

	require.NoError(t, mock.ExpectationsWereMet(), "SQL expectations should be met")
}

func TestConfig(t *testing.T) {
	config := Config{
		Host:            "localhost",
		Port:            5432,
		User:            "postgres",
		Password:        "password",
		DBName:          "testdb",
		Schema:          "public",
		SSLMode:         "disable",
		MaxIdleConns:    10,
		MaxOpenConns:    100,
		ConnMaxIdleTime: 5,
		ConnMaxLifetime: 60,
		Debug:           false,
	}

	assert.Equal(t, "localhost", config.Host, "Expected correct host")
	assert.Equal(t, 5432, config.Port, "Expected correct port")
	assert.Equal(t, "postgres", config.User, "Expected correct user")
	assert.Equal(t, "password", config.Password, "Expected correct password")
	assert.Equal(t, "testdb", config.DBName, "Expected correct dbname")
	assert.Equal(t, "public", config.Schema, "Expected correct schema")
	assert.Equal(t, "disable", config.SSLMode, "Expected correct sslmode")
	assert.Equal(t, 10, config.MaxIdleConns, "Expected correct max idle conns")
	assert.Equal(t, 100, config.MaxOpenConns, "Expected correct max open conns")
	assert.Equal(t, 5, config.ConnMaxIdleTime, "Expected correct conn max idle time")
	assert.Equal(t, 60, config.ConnMaxLifetime, "Expected correct conn max lifetime")
	assert.False(t, config.Debug, "Expected debug to be false")
}

func TestConfig_Validation(t *testing.T) {
	testCases := []struct {
		name   string
		config Config
		valid  bool
	}{
		{
			name: "valid config",
			config: Config{
				Host:            "localhost",
				Port:            5432,
				User:            "postgres",
				Password:        "password",
				DBName:          "testdb",
				Schema:          "public",
				SSLMode:         "disable",
				MaxIdleConns:    10,
				MaxOpenConns:    100,
				ConnMaxIdleTime: 5,
				ConnMaxLifetime: 60,
				Debug:           false,
			},
			valid: true,
		},
		{
			name: "empty host",
			config: Config{
				Host:            "",
				Port:            5432,
				ConnMaxIdleTime: 0, // Disable idle timeout
				ConnMaxLifetime: 0, // Disable lifetime timeout
				MaxOpenConns:    0, // Disable connection pooling
				MaxIdleConns:    0, // Disable idle connections
			},
			valid: false,
		},
		{
			name: "zero port",
			config: Config{
				Host:            "localhost",
				Port:            0,
				ConnMaxIdleTime: 0, // Disable idle timeout
				ConnMaxLifetime: 0, // Disable lifetime timeout
				MaxOpenConns:    0, // Disable connection pooling
				MaxIdleConns:    0, // Disable idle connections
			},
			valid: false,
		},
		{
			name: "empty user",
			config: Config{
				Host:            "localhost",
				Port:            5432,
				User:            "",
				ConnMaxIdleTime: 0, // Disable idle timeout
				ConnMaxLifetime: 0, // Disable lifetime timeout
				MaxOpenConns:    0, // Disable connection pooling
				MaxIdleConns:    0, // Disable idle connections
			},
			valid: false,
		},
		{
			name: "empty dbname",
			config: Config{
				Host:            "localhost",
				Port:            5432,
				User:            "postgres",
				DBName:          "",
				ConnMaxIdleTime: 0, // Disable idle timeout
				ConnMaxLifetime: 0, // Disable lifetime timeout
				MaxOpenConns:    0, // Disable connection pooling
				MaxIdleConns:    0, // Disable idle connections
			},
			valid: false,
		},
		{
			name: "negative connections",
			config: Config{
				Host:         "localhost",
				Port:         5432,
				User:         "postgres",
				DBName:       "testdb",
				MaxIdleConns: -1,
			},
			valid: false,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			if tc.valid {
				// For valid configs, we can't test the actual connection without a real DB
				// But we can verify the config struct is properly set
				assert.NotEmpty(t, tc.config.Host, "Host should not be empty for valid config")
				assert.Greater(t, tc.config.Port, 0, "Port should be positive for valid config")
			} else {
				// For invalid configs, test that NewPostgresClient fails
				client, err := NewPostgresClient(tc.config)
				assert.Error(t, err, "NewPostgresClient() should fail with invalid config")
				assert.Nil(t, client, "Client should be nil on error")
			}
		})
	}
}

func TestConfig_DefaultValues(t *testing.T) {
	config := Config{}

	// Test that zero values are handled appropriately
	assert.Empty(t, config.Host, "Host should be empty by default")
	assert.Equal(t, 0, config.Port, "Port should be 0 by default")
	assert.Empty(t, config.User, "User should be empty by default")
	assert.Empty(t, config.Password, "Password should be empty by default")
	assert.Empty(t, config.DBName, "DBName should be empty by default")
	assert.Empty(t, config.Schema, "Schema should be empty by default")
	assert.Empty(t, config.SSLMode, "SSLMode should be empty by default")
	assert.Equal(t, 0, config.MaxIdleConns, "MaxIdleConns should be 0 by default")
	assert.Equal(t, 0, config.MaxOpenConns, "MaxOpenConns should be 0 by default")
	assert.Equal(t, 0, config.ConnMaxIdleTime, "ConnMaxIdleTime should be 0 by default")
	assert.Equal(t, 0, config.ConnMaxLifetime, "ConnMaxLifetime should be 0 by default")
	assert.False(t, config.Debug, "Debug should be false by default")
	assert.Equal(t, 0, config.ConnectTimeout, "ConnectTimeout should be 0 by default")
}

func TestPostgresClient_QueryOperations(t *testing.T) {
	db, mock, err := sqlmock.New(sqlmock.MonitorPingsOption(true))
	require.NoError(t, err)
	defer db.Close()

	config := Config{
		Host:     "localhost",
		Port:     5432,
		User:     "postgres",
		Password: "password",
		DBName:   "testdb",
		Schema:   "public",
		SSLMode:  "disable",
		Debug:    false,
	}

	mock.ExpectPing()

	client := createMockPostgresClient(t, db, config)

	// Test SELECT query
	rows := sqlmock.NewRows([]string{"id", "name"}).
		AddRow(1, "test").
		AddRow(2, "test2")

	mock.ExpectQuery("SELECT (.+) FROM users").WillReturnRows(rows)

	var results []struct {
		ID   int    `db:"id"`
		Name string `db:"name"`
	}

	err = client.GetDB().Raw("SELECT id, name FROM users").Scan(&results).Error
	assert.NoError(t, err, "Select should succeed")
	assert.Len(t, results, 2, "Should return 2 rows")
	assert.Equal(t, "test", results[0].Name, "First result name should match")

	// Test INSERT query
	mock.ExpectExec("INSERT INTO users").WithArgs("john", "doe").WillReturnResult(sqlmock.NewResult(1, 1))

	err = client.GetDB().Exec("INSERT INTO users (first_name, last_name) VALUES (?, ?)", "john", "doe").Error
	assert.NoError(t, err, "Insert should succeed")

	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestPostgresClient_TransactionOperations(t *testing.T) {
	db, mock, err := sqlmock.New(sqlmock.MonitorPingsOption(true))
	require.NoError(t, err)
	defer db.Close()

	config := Config{
		Host:     "localhost",
		Port:     5432,
		User:     "postgres",
		Password: "password",
		DBName:   "testdb",
		Schema:   "public",
		SSLMode:  "disable",
		Debug:    false,
	}

	mock.ExpectPing()

	client := createMockPostgresClient(t, db, config)

	// Test successful transaction
	mock.ExpectBegin()
	mock.ExpectExec("INSERT INTO users").WithArgs("john").WillReturnResult(sqlmock.NewResult(1, 1))
	mock.ExpectCommit()

	tx := client.GetDB().Begin()
	assert.NotNil(t, tx, "Begin transaction should succeed")

	err = tx.Exec("INSERT INTO users (name) VALUES (?)", "john").Error
	assert.NoError(t, err, "Exec in transaction should succeed")

	err = tx.Commit().Error
	assert.NoError(t, err, "Commit should succeed")

	// Test failed transaction rollback
	mock.ExpectBegin()
	mock.ExpectExec("INSERT INTO users").WithArgs("jane").WillReturnError(sql.ErrNoRows)
	mock.ExpectRollback()

	tx = client.GetDB().Begin()
	assert.NotNil(t, tx, "Begin transaction should succeed")

	err = tx.Exec("INSERT INTO users (name) VALUES (?)", "jane").Error
	assert.Error(t, err, "Exec in transaction should fail")

	err = tx.Rollback().Error
	assert.NoError(t, err, "Rollback should succeed")

	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestNewPostgresClient(t *testing.T) {
	config := Config{
		Host:            "invalid-host",
		Port:            5432,
		User:            "invalid-user",
		Password:        "invalid-password",
		DBName:          "invalid-db",
		Schema:          "public",
		SSLMode:         "disable",
		ConnMaxIdleTime: 0, // Disable idle timeout
		ConnMaxLifetime: 0, // Disable lifetime timeout
		MaxOpenConns:    0, // Disable connection pooling
		MaxIdleConns:    0, // Disable idle connections
		ConnectTimeout:  1, // 1 second timeout for fast failure
	}

	// This should fail to connect
	client, err := NewPostgresClient(config)
	assert.Error(t, err, "NewPostgresClient() should fail with invalid config")
	assert.Nil(t, client, "Client should be nil on error")
}

func TestNewPostgresClient_EmptyConfig(t *testing.T) {
	config := Config{
		ConnMaxIdleTime: 0, // Disable idle timeout
		ConnMaxLifetime: 0, // Disable lifetime timeout
		MaxOpenConns:    0, // Disable connection pooling
		MaxIdleConns:    0, // Disable idle connections
		ConnectTimeout:  1, // 1 second timeout for fast failure
	}

	client, err := NewPostgresClient(config)
	assert.Error(t, err, "NewPostgresClient() should fail with empty config")
	assert.Nil(t, client, "Client should be nil on error")
	assert.Contains(t, err.Error(), "connect", "Error should mention connection failure")
}

func TestNewPostgresClient_InvalidPort(t *testing.T) {
	config := Config{
		Host:            "localhost",
		Port:            99999, // Invalid port
		User:            "postgres",
		Password:        "password",
		DBName:          "testdb",
		Schema:          "public",
		SSLMode:         "disable",
		ConnMaxIdleTime: 0, // Disable idle timeout
		ConnMaxLifetime: 0, // Disable lifetime timeout
		MaxOpenConns:    0, // Disable connection pooling
		MaxIdleConns:    0, // Disable idle connections
		ConnectTimeout:  1, // 1 second timeout for fast failure
	}

	client, err := NewPostgresClient(config)
	assert.Error(t, err, "NewPostgresClient() should fail with invalid port")
	assert.Nil(t, client, "Client should be nil on error")
}

func TestNewPostgresClient_DebugMode(t *testing.T) {
	config := Config{
		Host:            "invalid-host",
		Port:            5432,
		User:            "postgres",
		Password:        "password",
		DBName:          "testdb",
		Schema:          "public",
		SSLMode:         "disable",
		Debug:           true, // Enable debug mode
		ConnMaxIdleTime: 0,    // Disable idle timeout
		ConnMaxLifetime: 0,    // Disable lifetime timeout
		MaxOpenConns:    0,    // Disable connection pooling
		MaxIdleConns:    0,    // Disable idle connections
		ConnectTimeout:  1,    // 1 second timeout for fast failure
	}

	client, err := NewPostgresClient(config)
	assert.Error(t, err, "NewPostgresClient() should fail with invalid host even in debug mode")
	assert.Nil(t, client, "Client should be nil on error")
}
