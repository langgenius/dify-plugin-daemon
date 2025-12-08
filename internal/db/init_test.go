package db

import (
	"testing"
	"time"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"gorm.io/driver/mysql"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"

	"github.com/langgenius/dify-plugin-daemon/internal/types/app"
	"github.com/langgenius/dify-plugin-daemon/internal/types/models"
)

func TestInitDB(t *testing.T) {
	var pgConfig = &app.Config{
		DBType:     app.DB_TYPE_POSTGRESQL,
		DBUsername: "postgres",
		DBPassword: "difyai123456",
		DBHost:     "localhost",
		DBPort:     5432,
		DBDatabase: "dify_plugin_daemon",
		DBSslMode:  "disable",
	}

	var mysqlConfig = &app.Config{
		DBType:     app.DB_TYPE_MYSQL,
		DBUsername: "root",
		DBPassword: "difyai123456",
		DBHost:     "localhost",
		DBPort:     3306,
		DBDatabase: "dify_plugin_daemon",
		DBSslMode:  "disable",
	}

	var testConfigs = []*app.Config{
		pgConfig,
		mysqlConfig,
	}

	for _, config := range testConfigs {
		Init(config)
	}
}

func Test_cleanupDuplicatePluginInstallations_PostgreSQL(t *testing.T) {
	tests := []struct {
		name           string
		duplicateCount int64
		setupMocks     func(sqlmock.Sqlmock)
		expectedError  bool
	}{
		{
			name:           "no duplicates exist",
			duplicateCount: 0,
			setupMocks: func(mock sqlmock.Sqlmock) {
				// Mock table existence check
				mock.ExpectQuery(`SELECT count\(\*\) FROM information_schema.tables WHERE table_schema = CURRENT_SCHEMA\(\) AND table_name = \$1 AND table_type = \$2`).
					WithArgs("plugin_installations", "BASE TABLE").
					WillReturnRows(sqlmock.NewRows([]string{"count"}).AddRow(1))
				// Mock duplicate count query
				mock.ExpectQuery(`SELECT COUNT\(\*\) FROM \(`).
					WillReturnRows(sqlmock.NewRows([]string{"count"}).AddRow(0))
			},
			expectedError: false,
		},
		{
			name:           "duplicates exist and are cleaned up",
			duplicateCount: 2,
			setupMocks: func(mock sqlmock.Sqlmock) {
				// Mock table existence check
				mock.ExpectQuery(`SELECT count\(\*\) FROM information_schema.tables WHERE table_schema = CURRENT_SCHEMA\(\) AND table_name = \$1 AND table_type = \$2`).
					WithArgs("plugin_installations", "BASE TABLE").
					WillReturnRows(sqlmock.NewRows([]string{"count"}).AddRow(1))
				// First query to count duplicates
				mock.ExpectQuery(`SELECT COUNT\(\*\) FROM \(`).
					WillReturnRows(sqlmock.NewRows([]string{"count"}).AddRow(2))

				// PostgreSQL-specific delete duplicates query using DISTINCT ON
				mock.ExpectExec(`DELETE FROM plugin_installations`).
					WillReturnResult(sqlmock.NewResult(0, 2)) // 2 rows deleted
			},
			expectedError: false,
		},
		{
			name:           "error counting duplicates",
			duplicateCount: 0,
			setupMocks: func(mock sqlmock.Sqlmock) {
				// Mock table existence check
				mock.ExpectQuery(`SELECT count\(\*\) FROM information_schema.tables WHERE table_schema = CURRENT_SCHEMA\(\) AND table_name = \$1 AND table_type = \$2`).
					WithArgs("plugin_installations", "BASE TABLE").
					WillReturnRows(sqlmock.NewRows([]string{"count"}).AddRow(1))
				// Mock duplicate count query with error
				mock.ExpectQuery(`SELECT COUNT\(\*\) FROM \(`).
					WillReturnError(assert.AnError)
			},
			expectedError: true,
		},
		{
			name:           "error deleting duplicates",
			duplicateCount: 1,
			setupMocks: func(mock sqlmock.Sqlmock) {
				// Mock table existence check
				mock.ExpectQuery(`SELECT count\(\*\) FROM information_schema.tables WHERE table_schema = CURRENT_SCHEMA\(\) AND table_name = \$1 AND table_type = \$2`).
					WithArgs("plugin_installations", "BASE TABLE").
					WillReturnRows(sqlmock.NewRows([]string{"count"}).AddRow(1))
				// First query to count duplicates
				mock.ExpectQuery(`SELECT COUNT\(\*\) FROM \(`).
					WillReturnRows(sqlmock.NewRows([]string{"count"}).AddRow(1))

				// PostgreSQL delete duplicates query fails
				mock.ExpectExec(`DELETE FROM plugin_installations`).
					WillReturnError(assert.AnError)
			},
			expectedError: true,
		},
		{
			name:           "table doesn't exist",
			duplicateCount: 0,
			setupMocks: func(mock sqlmock.Sqlmock) {
				// Mock table existence check - table doesn't exist
				mock.ExpectQuery(`SELECT count\(\*\) FROM information_schema.tables WHERE table_schema = CURRENT_SCHEMA\(\) AND table_name = \$1 AND table_type = \$2`).
					WithArgs("plugin_installations", "BASE TABLE").
					WillReturnRows(sqlmock.NewRows([]string{"count"}).AddRow(0))
			},
			expectedError: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			db, mock := setupMockDB(t)
			defer func() {
				sqlDB, _ := db.DB()
				sqlDB.Close()
			}()

			// Set up global variable for test
			originalDB := DifyPluginDB
			DifyPluginDB = db
			defer func() { DifyPluginDB = originalDB }()

			tt.setupMocks(mock)

			err := cleanupDuplicatePluginInstallations()

			if tt.expectedError {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}

			assert.NoError(t, mock.ExpectationsWereMet())
		})
	}
}

func Test_cleanupDuplicatePluginInstallations_MySQL(t *testing.T) {
	tests := []struct {
		name           string
		duplicateCount int64
		setupMocks     func(sqlmock.Sqlmock)
		expectedError  bool
	}{
		{
			name:           "MySQL main scenario - table doesn't exist (fixes the original CI issue)",
			duplicateCount: 0,
			setupMocks: func(mock sqlmock.Sqlmock) {
				// No mocks needed - MySQL's HasTable returns false immediately for mock connections
				// This is the main scenario that was causing CI failures
			},
			expectedError: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			db, mock := setupMockMySQLDB(t)
			defer func() {
				sqlDB, _ := db.DB()
				sqlDB.Close()
			}()

			// Set up global variable for test
			originalDB := DifyPluginDB
			DifyPluginDB = db
			defer func() { DifyPluginDB = originalDB }()

			tt.setupMocks(mock)

			err := cleanupDuplicatePluginInstallations()

			if tt.expectedError {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}

			assert.NoError(t, mock.ExpectationsWereMet())
		})
	}
}

func Test_autoMigrate_withCleanup(t *testing.T) {
	tests := []struct {
		name          string
		setupMocks    func(sqlmock.Sqlmock)
		expectedError bool
	}{
		{
			name: "auto migrate fails during cleanup",
			setupMocks: func(mock sqlmock.Sqlmock) {
				// Mock table existence check
				mock.ExpectQuery(`SELECT count\(\*\) FROM information_schema.tables WHERE table_schema = CURRENT_SCHEMA\(\) AND table_name = \$1 AND table_type = \$2`).
					WithArgs("plugin_installations", "BASE TABLE").
					WillReturnRows(sqlmock.NewRows([]string{"count"}).AddRow(1))
				// Cleanup duplicates check fails
				mock.ExpectQuery(`SELECT COUNT\(\*\) FROM \(`).
					WillReturnError(assert.AnError)
			},
			expectedError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			db, mock := setupMockDB(t)
			defer func() {
				sqlDB, _ := db.DB()
				sqlDB.Close()
			}()

			originalDB := DifyPluginDB
			DifyPluginDB = db
			defer func() { DifyPluginDB = originalDB }()

			tt.setupMocks(mock)

			err := autoMigrate()

			if tt.expectedError {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}

			assert.NoError(t, mock.ExpectationsWereMet())
		})
	}
}

// setupMockDB creates a mock PostgreSQL database connection for testing
func setupMockDB(t *testing.T) (*gorm.DB, sqlmock.Sqlmock) {
	sqlDB, mock, err := sqlmock.New()
	require.NoError(t, err)

	gormDB, err := gorm.Open(postgres.New(postgres.Config{
		Conn: sqlDB,
	}), &gorm.Config{})
	require.NoError(t, err)

	return gormDB, mock
}

// setupMockMySQLDB creates a mock MySQL database connection for testing
func setupMockMySQLDB(t *testing.T) (*gorm.DB, sqlmock.Sqlmock) {
	sqlDB, mock, err := sqlmock.New()
	require.NoError(t, err)

	// MySQL driver queries version and schema on initialization
	mock.ExpectQuery("SELECT VERSION()").WillReturnRows(sqlmock.NewRows([]string{"version"}).AddRow("8.0.0"))
	mock.ExpectQuery("SELECT SCHEMA_NAME from Information_schema.SCHEMATA").WillReturnRows(sqlmock.NewRows([]string{"SCHEMA_NAME"}).AddRow("test_db"))

	gormDB, err := gorm.Open(mysql.New(mysql.Config{
		Conn: sqlDB,
	}), &gorm.Config{})
	require.NoError(t, err)

	return gormDB, mock
}

// Test helper to create test plugin installations
func createTestPluginInstallation(tenantID, pluginID string, createdAt time.Time) *models.PluginInstallation {
	return &models.PluginInstallation{
		TenantID:    tenantID,
		PluginID:    pluginID,
		RuntimeType: "local",
		Meta:        make(map[string]any),
		Model:       models.Model{CreatedAt: createdAt},
	}
}

func Test_createTestPluginInstallation(t *testing.T) {
	tenantID := "test-tenant-uuid"
	pluginID := "test-plugin"
	createdAt := time.Now()

	installation := createTestPluginInstallation(tenantID, pluginID, createdAt)

	assert.Equal(t, tenantID, installation.TenantID)
	assert.Equal(t, pluginID, installation.PluginID)
	assert.Equal(t, "local", installation.RuntimeType)
	assert.Equal(t, createdAt, installation.CreatedAt)
	assert.NotNil(t, installation.Meta)
}
