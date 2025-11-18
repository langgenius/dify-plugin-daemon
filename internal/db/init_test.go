package db

import (
	"testing"

	"github.com/langgenius/dify-plugin-daemon/internal/types/app"
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
