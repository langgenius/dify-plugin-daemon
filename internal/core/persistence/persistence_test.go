package persistence

import (
	"encoding/hex"
	"testing"

	cloudoss "github.com/langgenius/dify-cloud-kit/oss"
	"github.com/langgenius/dify-cloud-kit/oss/factory"
	"github.com/langgenius/dify-plugin-daemon/internal/db"
	"github.com/langgenius/dify-plugin-daemon/internal/types/app"
	"github.com/langgenius/dify-plugin-daemon/internal/utils/cache"
	"github.com/langgenius/dify-plugin-daemon/internal/utils/strings"
	"github.com/stretchr/testify/assert"
)

func TestPersistenceStoreAndLoad(t *testing.T) {
	err := cache.InitRedisClient("localhost:6379", "", "difyai123456", false, 0)
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

	oss, err := factory.Load("local", cloudoss.OSSArgs{
		Local: &cloudoss.Local{
			Path: "./storage",
		},
	},
	)
	if err != nil {
		t.Error("failed to load local storage", err.Error())
	}

	InitPersistence(oss, &app.Config{
		PersistenceStoragePath:    "./persistence_storage",
		PersistenceStorageMaxSize: 1024 * 1024 * 1024,
	})

	key := strings.RandomString(10)

	assert.Nil(t, persistence.Save("tenant_id", "plugin_checksum", -1, key, []byte("data")))

	data, err := persistence.Load("tenant_id", "plugin_checksum", key)
	assert.Nil(t, err)
	assert.Equal(t, string(data), "data")

	// check if the file exists
	if _, err := oss.Load("./persistence_storage/tenant_id/plugin_checksum/" + key); err != nil {
		t.Fatalf("File not found: %v", err)
	}

	// check if cache is updated
	cacheData, err := cache.GetString("persistence:cache:tenant_id:plugin_checksum:" + key)
	assert.Nil(t, err)

	cacheDataBytes, err := hex.DecodeString(cacheData)
	assert.Nil(t, err)
	assert.Equal(t, string(cacheDataBytes), "data")
}

func TestPersistenceSaveAndLoadWithLongKey(t *testing.T) {
	err := cache.InitRedisClient("localhost:6379", "", "difyai123456", false, 0)
	assert.Nil(t, err)
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

	oss, err := factory.Load("local", cloudoss.OSSArgs{
		Local: &cloudoss.Local{
			Path: "./storage",
		},
	})
	assert.Nil(t, err)

	InitPersistence(oss, &app.Config{
		PersistenceStoragePath:    "./persistence_storage",
		PersistenceStorageMaxSize: 1024 * 1024 * 1024,
	})

	key := strings.RandomString(257)

	if err := persistence.Save("tenant_id", "plugin_checksum", -1, key, []byte("data")); err == nil {
		t.Fatalf("Expected error, got nil")
	}
}

func TestPersistenceDelete(t *testing.T) {
	err := cache.InitRedisClient("localhost:6379", "", "difyai123456", false, 0)
	assert.Nil(t, err)
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

	oss, err := factory.Load("local", cloudoss.OSSArgs{
		Local: &cloudoss.Local{
			Path: "./storage",
		},
	})
	assert.Nil(t, err)

	InitPersistence(oss, &app.Config{
		PersistenceStoragePath:    "./persistence_storage",
		PersistenceStorageMaxSize: 1024 * 1024 * 1024,
	})

	key := strings.RandomString(10)

	if err := persistence.Save("tenant_id", "plugin_checksum", -1, key, []byte("data")); err != nil {
		t.Fatalf("Failed to save data: %v", err)
	}

	if _, err := persistence.Delete("tenant_id", "plugin_checksum", key); err != nil {
		t.Fatalf("Failed to delete data: %v", err)
	}

	// check if the file exists
	if _, err := oss.Load("./persistence_storage/tenant_id/plugin_checksum/" + key); err == nil {
		t.Fatalf("File not deleted: %v", err)
	}

	// check if cache is updated
	_, err = cache.GetString("persistence:cache:tenant_id:plugin_checksum:" + key)
	assert.Equal(t, err, cache.ErrNotFound)
}

func TestPersistencePathTraversal(t *testing.T) {
	err := cache.InitRedisClient("localhost:6379", "", "difyai123456", false, 0)
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

	oss, err := factory.Load("local", cloudoss.OSSArgs{
		Local: &cloudoss.Local{
			Path: "./storage",
		},
	})
	assert.Nil(t, err)

	InitPersistence(oss, &app.Config{
		PersistenceStoragePath:    "./persistence_storage",
		PersistenceStorageMaxSize: 1024 * 1024 * 1024,
	})

	// Test cases for path traversal
	testCases := []struct {
		name    string
		key     string
		wantErr bool
	}{
		{
			name:    "normal key",
			key:     "test.txt",
			wantErr: false,
		},
		{
			name:    "parent directory traversal",
			key:     "../test.txt",
			wantErr: true,
		},
		{
			name:    "multiple parent directory traversal",
			key:     "../../test.txt",
			wantErr: true,
		},
		{
			name:    "double slash",
			key:     "test//test.txt",
			wantErr: false,
		},
		{
			name:    "backslash",
			key:     "test\\test.txt",
			wantErr: true,
		},
		{
			name:    "mixed traversal",
			key:     "test/../test.txt",
			wantErr: false,
		},
		{
			name:    "absolute path",
			key:     "/etc/passwd",
			wantErr: false,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			// Test Save
			err := persistence.Save("tenant_id", "plugin_checksum", -1, tc.key, []byte("data"))
			assert.Equal(t, err != nil, tc.wantErr)

			// Test Load
			_, err = persistence.Load("tenant_id", "plugin_checksum", tc.key)
			assert.Equal(t, err != nil, tc.wantErr)

			// Test Delete
			_, err = persistence.Delete("tenant_id", "plugin_checksum", tc.key)
			assert.Equal(t, err != nil, tc.wantErr)
		})
	}
}
