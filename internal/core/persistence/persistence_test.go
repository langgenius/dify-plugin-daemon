package persistence

import (
	"encoding/hex"
	"testing"

	"github.com/langgenius/dify-plugin-daemon/internal/db"
	"github.com/langgenius/dify-plugin-daemon/internal/oss/local"
	"github.com/langgenius/dify-plugin-daemon/internal/types/app"
	"github.com/langgenius/dify-plugin-daemon/internal/utils/cache"
	"github.com/langgenius/dify-plugin-daemon/internal/utils/strings"
)

func TestPersistenceStoreAndLoad(t *testing.T) {
	err := cache.InitRedisClient("localhost:6379", "difyai123456", false, 0)
	if err != nil {
		t.Fatalf("Failed to init redis client: %v", err)
	}
	defer cache.Close()

	db.Init(&app.Config{
		DBType:            "postgresql",
		DBUsername:        "postgres",
		DBPassword:        "difyai123456",
		DBHost:            "localhost",
		DBDefaultDatabase: "postgres",
		DBPort:            5432,
		DBDatabase:        "dify_plugin_daemon",
		DBSslMode:         "disable",
	})
	defer db.Close()

	oss := local.NewLocalStorage("./storage")

	InitPersistence(oss, &app.Config{
		PersistenceStoragePath:    "./persistence_storage",
		PersistenceStorageMaxSize: 1024 * 1024 * 1024,
	})

	key := strings.RandomString(10)

	if err := persistence.Save("tenant_id", "plugin_checksum", -1, key, []byte("data")); err != nil {
		t.Fatalf("Failed to save data: %v", err)
	}

	data, err := persistence.Load("tenant_id", "plugin_checksum", key)
	if err != nil {
		t.Fatalf("Failed to load data: %v", err)
	}

	if string(data) != "data" {
		t.Fatalf("Data mismatch: %s", data)
	}

	// check if the file exists
	if _, err := oss.Load("./persistence_storage/tenant_id/plugin_checksum/" + key); err != nil {
		t.Fatalf("File not found: %v", err)
	}

	// check if cache is updated
	cacheData, err := cache.GetString("persistence:cache:tenant_id:plugin_checksum:" + key)
	if err != nil {
		t.Fatalf("Failed to get cache data: %v", err)
	}

	cacheDataBytes, err := hex.DecodeString(cacheData)
	if err != nil {
		t.Fatalf("Failed to decode cache data: %v", err)
	}

	if string(cacheDataBytes) != "data" {
		t.Fatalf("Cache data mismatch: %s", cacheData)
	}
}

func TestPersistenceSaveAndLoadWithLongKey(t *testing.T) {
	err := cache.InitRedisClient("localhost:6379", "difyai123456", false, 0)
	if err != nil {
		t.Fatalf("Failed to init redis client: %v", err)
	}
	defer cache.Close()
	db.Init(&app.Config{
		DBType:     "postgresql",
		DBUsername: "postgres",
		DBPassword: "difyai123456",
		DBHost:     "localhost",
		DBPort:     5432,
		DBDatabase: "dify_plugin_daemon",
		DBSslMode:  "disable",
	})
	defer db.Close()

	InitPersistence(local.NewLocalStorage("./storage"), &app.Config{
		PersistenceStoragePath:    "./persistence_storage",
		PersistenceStorageMaxSize: 1024 * 1024 * 1024,
	})

	key := strings.RandomString(257)

	if err := persistence.Save("tenant_id", "plugin_checksum", -1, key, []byte("data")); err == nil {
		t.Fatalf("Expected error, got nil")
	}
}

func TestPersistenceDelete(t *testing.T) {
	err := cache.InitRedisClient("localhost:6379", "difyai123456", false, 0)
	if err != nil {
		t.Fatalf("Failed to init redis client: %v", err)
	}
	defer cache.Close()
	db.Init(&app.Config{
		DBType:     "postgresql",
		DBUsername: "postgres",
		DBPassword: "difyai123456",
		DBHost:     "localhost",
		DBPort:     5432,
		DBDatabase: "dify_plugin_daemon",
		DBSslMode:  "disable",
	})
	defer db.Close()

	oss := local.NewLocalStorage("./storage")

	InitPersistence(oss, &app.Config{
		PersistenceStoragePath:    "./persistence_storage",
		PersistenceStorageMaxSize: 1024 * 1024 * 1024,
	})

	key := strings.RandomString(10)

	if err := persistence.Save("tenant_id", "plugin_checksum", -1, key, []byte("data")); err != nil {
		t.Fatalf("Failed to save data: %v", err)
	}

	if err := persistence.Delete("tenant_id", "plugin_checksum", key); err != nil {
		t.Fatalf("Failed to delete data: %v", err)
	}

	// check if the file exists
	if _, err := oss.Load("./persistence_storage/tenant_id/plugin_checksum/" + key); err == nil {
		t.Fatalf("File not deleted: %v", err)
	}

	// check if cache is updated
	_, err = cache.GetString("persistence:cache:tenant_id:plugin_checksum:" + key)
	if err != cache.ErrNotFound {
		t.Fatalf("Cache data not deleted: %v", err)
	}
}
