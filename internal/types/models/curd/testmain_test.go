package curd

import (
	"os"
	"testing"

	"github.com/langgenius/dify-plugin-daemon/internal/db"
	"github.com/langgenius/dify-plugin-daemon/internal/types/app"
	"github.com/langgenius/dify-plugin-daemon/pkg/utils/cache"
)

func TestMain(m *testing.M) {
	cfg := &app.Config{
		DBType:     app.DB_TYPE_POSTGRESQL,
		DBUsername: "postgres",
		DBPassword: "difyai123456",
		DBHost:     "localhost",
		DBPort:     5432,
		DBDatabase: "dify_plugin_daemon",
		DBSslMode:  "disable",
	}
	cfg.SetDefault()

	db.Init(cfg)
	defer db.Close()

	if err := db.AutoMigrate(); err != nil {
		panic("Failed to auto migrate: " + err.Error())
	}

	if err := cache.InitRedisClient("127.0.0.1:6379", "", "difyai123456", false, 0, nil); err != nil {
		panic("Failed to init redis client: " + err.Error())
	}
	defer cache.Close()

	os.Exit(m.Run())
}
