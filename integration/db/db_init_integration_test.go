package dbintegration_test

import (
	"testing"

	"github.com/langgenius/dify-plugin-daemon/internal/db"
	"github.com/langgenius/dify-plugin-daemon/internal/types/app"
)

func TestInitDB(t *testing.T) {
	cases := []struct {
		name   string
		config app.Config
	}{
		{
			name: "postgresql",
			config: app.Config{
				DBType:     app.DB_TYPE_POSTGRESQL,
				DBUsername: "postgres",
				DBPassword: "difyai123456",
				DBHost:     "localhost",
				DBPort:     5432,
				DBDatabase: "dify_plugin_daemon",
				DBSslMode:  "disable",
			},
		},
		{
			name: "pgbouncer",
			config: app.Config{
				DBType:     app.DB_TYPE_PG_BOUNCER,
				DBUsername: "postgres",
				DBPassword: "difyai123456",
				DBHost:     "localhost",
				DBPort:     6432,
				DBDatabase: "dify_plugin_daemon",
				DBSslMode:  "disable",
			},
		},
		{
			name: "mysql",
			config: app.Config{
				DBType:     app.DB_TYPE_MYSQL,
				DBUsername: "root",
				DBPassword: "difyai123456",
				DBHost:     "localhost",
				DBPort:     3306,
				DBDatabase: "dify_plugin_daemon",
				DBSslMode:  "disable",
			},
		},
	}

	for _, tc := range cases {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			tc.config.SetDefault()
			db.Init(&tc.config)
			db.Close()
		})
	}
}
