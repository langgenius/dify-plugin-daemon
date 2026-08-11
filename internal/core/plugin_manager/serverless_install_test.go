package plugin_manager

import (
	"strings"
	"sync"
	"testing"
	"time"

	miniredis "github.com/alicebob/miniredis/v2"
	"github.com/langgenius/dify-plugin-daemon/internal/db"
	"github.com/langgenius/dify-plugin-daemon/internal/types/models"
	"github.com/langgenius/dify-plugin-daemon/pkg/entities/plugin_entities"
	"github.com/langgenius/dify-plugin-daemon/pkg/utils/cache"
	"github.com/stretchr/testify/require"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

func TestPersistServerlessRuntimeConvergesUnderConcurrency(t *testing.T) {
	redisServer := miniredis.RunT(t)
	require.NoError(t, cache.InitRedisClient(redisServer.Addr(), cache.RedisCredentials{}, false, 0, nil))
	t.Cleanup(func() {
		_ = cache.Close()
	})

	dsn := "file:" + strings.ReplaceAll(t.Name(), "/", "_") + "?mode=memory&cache=shared&_busy_timeout=5000"
	gormDB, err := gorm.Open(sqlite.Open(dsn), &gorm.Config{})
	require.NoError(t, err)
	require.NoError(t, gormDB.AutoMigrate(&models.ServerlessRuntime{}))
	sqlDB, err := gormDB.DB()
	require.NoError(t, err)
	sqlDB.SetMaxOpenConns(1)
	db.DifyPluginDB = gormDB
	t.Cleanup(func() {
		_ = sqlDB.Close()
	})

	identifier := plugin_entities.PluginUniqueIdentifier(
		"langgenius/serverless_concurrency:1.0.0@0123456789abcdef0123456789abcdef",
	)
	manager := &PluginManager{}
	require.NoError(t, cache.Store(
		manager.getServerlessRuntimeCacheKey(identifier),
		models.ServerlessRuntime{FunctionName: "stale"},
		time.Minute,
	))

	const workers = 8
	start := make(chan struct{})
	var ready sync.WaitGroup
	var finished sync.WaitGroup
	ready.Add(workers)
	finished.Add(workers)
	errs := make(chan error, workers)
	for range workers {
		go func() {
			defer finished.Done()
			ready.Done()
			<-start
			errs <- manager.persistServerlessRuntime(
				identifier,
				"https://runtime.example.test",
				"serverless-concurrency",
			)
		}()
	}
	ready.Wait()
	close(start)
	finished.Wait()
	close(errs)

	for err := range errs {
		require.NoError(t, err)
	}

	runtimes, err := db.GetAll[models.ServerlessRuntime](
		db.Equal("plugin_unique_identifier", identifier.String()),
	)
	require.NoError(t, err)
	require.Len(t, runtimes, 1)
	require.Equal(t, "https://runtime.example.test", runtimes[0].FunctionURL)
	require.Equal(t, "serverless-concurrency", runtimes[0].FunctionName)

	_, err = cache.Get[models.ServerlessRuntime](manager.getServerlessRuntimeCacheKey(identifier))
	require.ErrorIs(t, err, cache.ErrNotFound)
}
