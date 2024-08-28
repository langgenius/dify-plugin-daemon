package persistence

import (
	"encoding/hex"
	"os"
	"testing"

	"github.com/langgenius/dify-plugin-daemon/internal/types/app"
	"github.com/langgenius/dify-plugin-daemon/internal/utils/cache"
	"github.com/langgenius/dify-plugin-daemon/internal/utils/strings"
)

func TestPersistenceStoreAndLoad(t *testing.T) {
	err := cache.InitRedisClient("localhost:6379", "difyai123456")
	if err != nil {
		t.Fatalf("Failed to init redis client: %v", err)
	}
	defer cache.Close()

	p := InitPersistence(&app.Config{
		PersistenceStorageType:      "local",
		PersistenceStorageLocalPath: "./persistence_storage",
	})

	key := strings.RandomString(10)

	if err := p.Save("tenant_id", "plugin_checksum", key, []byte("data")); err != nil {
		t.Fatalf("Failed to save data: %v", err)
	}

	data, err := p.Load("tenant_id", "plugin_checksum", key)
	if err != nil {
		t.Fatalf("Failed to load data: %v", err)
	}

	if string(data) != "data" {
		t.Fatalf("Data mismatch: %s", data)
	}

	// check if the file exists
	if _, err := os.Stat("./persistence_storage/tenant_id/plugin_checksum/" + key); os.IsNotExist(err) {
		t.Fatalf("File not found: %v", err)
	}

	// check if cache is updated
	cache_data, err := cache.GetString("persistence:cache:tenant_id:plugin_checksum:" + key)
	if err != nil {
		t.Fatalf("Failed to get cache data: %v", err)
	}

	cache_data_bytes, err := hex.DecodeString(cache_data)
	if err != nil {
		t.Fatalf("Failed to decode cache data: %v", err)
	}

	if string(cache_data_bytes) != "data" {
		t.Fatalf("Cache data mismatch: %s", cache_data)
	}
}

func TestPersistenceSaveAndLoadWithLongKey(t *testing.T) {
	err := cache.InitRedisClient("localhost:6379", "difyai123456")
	if err != nil {
		t.Fatalf("Failed to init redis client: %v", err)
	}
	defer cache.Close()

	p := InitPersistence(&app.Config{
		PersistenceStorageType:      "local",
		PersistenceStorageLocalPath: "./persistence_storage",
	})

	key := strings.RandomString(65)

	if err := p.Save("tenant_id", "plugin_checksum", key, []byte("data")); err == nil {
		t.Fatalf("Expected error, got nil")
	}
}

func TestPersistenceDelete(t *testing.T) {
	err := cache.InitRedisClient("localhost:6379", "difyai123456")
	if err != nil {
		t.Fatalf("Failed to init redis client: %v", err)
	}
	defer cache.Close()

	p := InitPersistence(&app.Config{
		PersistenceStorageType:      "local",
		PersistenceStorageLocalPath: "./persistence_storage",
	})

	key := strings.RandomString(10)

	if err := p.Save("tenant_id", "plugin_checksum", key, []byte("data")); err != nil {
		t.Fatalf("Failed to save data: %v", err)
	}

	if err := p.Delete("tenant_id", "plugin_checksum", key); err != nil {
		t.Fatalf("Failed to delete data: %v", err)
	}

	// check if the file exists
	if _, err := os.Stat("./persistence_storage/tenant_id/plugin_checksum/" + key); !os.IsNotExist(err) {
		t.Fatalf("File not deleted: %v", err)
	}

	// check if cache is updated
	_, err = cache.GetString("persistence:cache:tenant_id:plugin_checksum:" + key)
	if err != cache.ErrNotFound {
		t.Fatalf("Cache data not deleted: %v", err)
	}
}
