package curd

import (
	"fmt"
	"os"
	"testing"
	"time"

	"github.com/langgenius/dify-plugin-daemon/internal/db"
	"github.com/langgenius/dify-plugin-daemon/internal/types/app"
	"github.com/langgenius/dify-plugin-daemon/pkg/utils/cache"
	"github.com/langgenius/dify-plugin-daemon/pkg/utils/network"
)

func TestMain(m *testing.M) {
	if !network.IsTCPReachable("localhost:5432", 500*time.Millisecond) {
		fmt.Fprintln(os.Stderr, "skipping curd tests: postgres is unavailable at localhost:5432")
		os.Exit(0)
	}

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

	if !network.IsTCPReachable("127.0.0.1:6379", 500*time.Millisecond) {
		fmt.Fprintln(os.Stderr, "skipping curd tests: redis is unavailable at 127.0.0.1:6379")
		os.Exit(0)
	}

	if err := cache.InitRedisClient("127.0.0.1:6379", "", "difyai123456", false, 0, nil); err != nil {
		fmt.Fprintf(os.Stderr, "skipping curd tests: failed to init redis client: %v\n", err)
		os.Exit(0)
	}
	defer cache.Close()

	os.Exit(m.Run())
}
