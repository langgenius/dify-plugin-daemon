package persistence

import (
	"encoding/hex"
	"os"
	"testing"

	cloudoss "github.com/langgenius/dify-cloud-kit/oss"
	"github.com/langgenius/dify-cloud-kit/oss/factory"
	"github.com/langgenius/dify-plugin-daemon/internal/db"
	"github.com/langgenius/dify-plugin-daemon/internal/types/app"
	"github.com/langgenius/dify-plugin-daemon/internal/types/models"
	"github.com/langgenius/dify-plugin-daemon/pkg/utils/cache"
	"github.com/langgenius/dify-plugin-daemon/pkg/utils/strings"
	"github.com/stretchr/testify/assert"
)

func TestMain(m *testing.M) {
	if err := cache.InitRedisClient("localhost:6379", "", "difyai123456", false, 0, nil); err != nil {
		panic("Failed to init redis client: " + err.Error())
	}
	defer cache.Close()

	db.Init(&app.Config{
		DBType:     app.DB_TYPE_POSTGRESQL,
		DBUsername: "postgres",
		DBPassword: "difyai123456",
		DBHost:     "localhost",
		DBPort:     5432,
		DBDatabase: "dify_plugin_daemon",
		DBSslMode:  "disable",
	})
	defer db.Close()

	if err := db.AutoMigrate(); err != nil {
		panic("Failed to auto migrate: " + err.Error())
	}

	os.Exit(m.Run())
}

func TestPersistenceStoreAndLoad(t *testing.T) {
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

func TestPersistenceOverwriteAdjustsCounter(t *testing.T) {
	// init deps
	err := cache.InitRedisClient("localhost:6379", "", "difyai123456", false, 0, nil)
	assert.Nil(t, err)
	defer cache.Close()

	db.Init(&app.Config{
		DBType:     app.DB_TYPE_POSTGRESQL,
		DBUsername: "postgres",
		DBPassword: "difyai123456",
		DBHost:     "localhost",
		DBPort:     5432,
		DBDatabase: "dify_plugin_daemon",
		DBSslMode:  "disable",
	})
	defer db.Close()

	oss, err := factory.Load("local", cloudoss.OSSArgs{Local: &cloudoss.Local{Path: "./storage"}})
	assert.Nil(t, err)

	InitPersistence(oss, &app.Config{PersistenceStoragePath: "./persistence_storage", PersistenceStorageMaxSize: 1024 * 1024})

	tenant := "tenant_" + strings.RandomString(6)
	plugin := "plugin_" + strings.RandomString(6)
	key := "k_" + strings.RandomString(6)

	// write 4 bytes
	assert.Nil(t, persistence.Save(tenant, plugin, -1, key, []byte("abcd")))
	st, err := db.GetOne[models.TenantStorage](db.Equal("tenant_id", tenant), db.Equal("plugin_id", plugin))
	assert.Nil(t, err)
	assert.Equal(t, int64(4), st.Size)

	// overwrite with 2 bytes -> size should be 2
	assert.Nil(t, persistence.Save(tenant, plugin, -1, key, []byte("bb")))
	st, err = db.GetOne[models.TenantStorage](db.Equal("tenant_id", tenant), db.Equal("plugin_id", plugin))
	assert.Nil(t, err)
	assert.Equal(t, int64(2), st.Size)

	// overwrite with 3 bytes -> size should be 3
	assert.Nil(t, persistence.Save(tenant, plugin, -1, key, []byte("ccc")))
	st, err = db.GetOne[models.TenantStorage](db.Equal("tenant_id", tenant), db.Equal("plugin_id", plugin))
	assert.Nil(t, err)
	assert.Equal(t, int64(3), st.Size)

	// and data should be latest
	data, err := persistence.Load(tenant, plugin, key)
	assert.Nil(t, err)
	assert.Equal(t, "ccc", string(data))
}

func TestPersistenceOverwriteLimitEnforcedByDelta(t *testing.T) {
	// init deps
	err := cache.InitRedisClient("localhost:6379", "", "difyai123456", false, 0, nil)
	assert.Nil(t, err)
	defer cache.Close()

	db.Init(&app.Config{
		DBType:     app.DB_TYPE_POSTGRESQL,
		DBUsername: "postgres",
		DBPassword: "difyai123456",
		DBHost:     "localhost",
		DBPort:     5432,
		DBDatabase: "dify_plugin_daemon",
		DBSslMode:  "disable",
	})
	defer db.Close()

	oss, err := factory.Load("local", cloudoss.OSSArgs{Local: &cloudoss.Local{Path: "./storage"}})
	assert.Nil(t, err)

	// set a small global limit 5 bytes
	InitPersistence(oss, &app.Config{PersistenceStoragePath: "./persistence_storage", PersistenceStorageMaxSize: 5})

	tenant := "tenant_" + strings.RandomString(6)
	plugin := "plugin_" + strings.RandomString(6)
	key := "k_" + strings.RandomString(6)

	// write 4 bytes OK
	assert.Nil(t, persistence.Save(tenant, plugin, -1, key, []byte("aaaa")))
	st, err := db.GetOne[models.TenantStorage](db.Equal("tenant_id", tenant), db.Equal("plugin_id", plugin))
	assert.Nil(t, err)
	assert.Equal(t, int64(4), st.Size)

	// overwrite with 6 bytes -> delta = +2, 4+2=6 > 5 -> expect error, no change
	if err := persistence.Save(tenant, plugin, -1, key, []byte("abcdef")); err == nil {
		t.Fatalf("expected limit error, got nil")
	}

	st, err = db.GetOne[models.TenantStorage](db.Equal("tenant_id", tenant), db.Equal("plugin_id", plugin))
	assert.Nil(t, err)
	assert.Equal(t, int64(4), st.Size)

	// stored data should remain old value
	data, err := persistence.Load(tenant, plugin, key)
	assert.Nil(t, err)
	assert.Equal(t, "aaaa", string(data))
}

func TestPersistenceSaveAndLoadWithLongKey(t *testing.T) {
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
