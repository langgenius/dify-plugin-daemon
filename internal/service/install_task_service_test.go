package service

import (
	"fmt"
	"sync"
	"sync/atomic"
	"testing"
	"time"

	"github.com/langgenius/dify-plugin-daemon/internal/db"
	"github.com/langgenius/dify-plugin-daemon/internal/types/models"
	"golang.org/x/sync/singleflight"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

func setupTestDB(t *testing.T) *gorm.DB {
	t.Helper()
	testDB, err := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{})
	if err != nil {
		t.Fatalf("failed to open test db: %v", err)
	}
	sqlDB, err := testDB.DB()
	if err != nil {
		t.Fatalf("failed to get underlying sql.DB: %v", err)
	}
	sqlDB.SetMaxOpenConns(1)

	if err := testDB.AutoMigrate(&models.InstallTask{}); err != nil {
		t.Fatalf("failed to auto migrate: %v", err)
	}
	db.DifyPluginDB = testDB
	t.Cleanup(func() {
		sqlDB.Close()
	})
	return testDB
}

func TestFetchPluginInstallationTasks_Singleflight(t *testing.T) {
	testDB := setupTestDB(t)
	installationTasksGroup = singleflight.Group{}

	var queryCount atomic.Int32
	testDB.Callback().Query().Before("gorm:query").Register("test:count_tasks", func(tx *gorm.DB) {
		queryCount.Add(1)
		time.Sleep(100 * time.Millisecond)
	})
	defer testDB.Callback().Query().Remove("test:count_tasks")

	const concurrency = 50
	var wg sync.WaitGroup
	wg.Add(concurrency)
	start := make(chan struct{})
	errs := make([]int, concurrency)

	for i := 0; i < concurrency; i++ {
		go func(idx int) {
			defer wg.Done()
			<-start
			resp := FetchPluginInstallationTasks("tenant-1", 1, 10)
			errs[idx] = resp.Code
		}(i)
	}

	close(start)
	wg.Wait()

	for i, code := range errs {
		if code != 0 {
			t.Errorf("goroutine %d: expected code 0, got %d", i, code)
		}
	}

	if count := queryCount.Load(); count != 1 {
		t.Errorf("singleflight not working: expected 1 db query for same key, got %d", count)
	}
}

func TestFetchPluginInstallationTask_Singleflight(t *testing.T) {
	testDB := setupTestDB(t)
	installationTaskGroup = singleflight.Group{}

	// Insert a test record before registering the callback.
	task := models.InstallTask{
		TenantID:     "tenant-1",
		Status:       models.InstallTaskStatusPending,
		TotalPlugins: 1,
	}
	if err := testDB.Create(&task).Error; err != nil {
		t.Fatalf("failed to create test task: %v", err)
	}

	var queryCount atomic.Int32
	testDB.Callback().Query().Before("gorm:query").Register("test:count_task", func(tx *gorm.DB) {
		queryCount.Add(1)
		time.Sleep(100 * time.Millisecond)
	})
	defer testDB.Callback().Query().Remove("test:count_task")

	const concurrency = 50
	var wg sync.WaitGroup
	wg.Add(concurrency)
	start := make(chan struct{})
	errs := make([]int, concurrency)

	for i := 0; i < concurrency; i++ {
		go func(idx int) {
			defer wg.Done()
			<-start
			resp := FetchPluginInstallationTask("tenant-1", task.ID)
			errs[idx] = resp.Code
		}(i)
	}

	close(start)
	wg.Wait()

	for i, code := range errs {
		if code != 0 {
			t.Errorf("goroutine %d: expected code 0, got %d", i, code)
		}
	}

	if count := queryCount.Load(); count != 1 {
		t.Errorf("singleflight not working: expected 1 db query for same key, got %d", count)
	}
}

func TestFetchPluginInstallationTasks_DifferentKeysNotDeduplicated(t *testing.T) {
	testDB := setupTestDB(t)
	installationTasksGroup = singleflight.Group{}

	var queryCount atomic.Int32
	testDB.Callback().Query().Before("gorm:query").Register("test:count_diff", func(tx *gorm.DB) {
		queryCount.Add(1)
		time.Sleep(50 * time.Millisecond)
	})
	defer testDB.Callback().Query().Remove("test:count_diff")

	const numKeys = 3
	var wg sync.WaitGroup
	wg.Add(numKeys)
	start := make(chan struct{})

	for i := 0; i < numKeys; i++ {
		go func(idx int) {
			defer wg.Done()
			<-start
			FetchPluginInstallationTasks(fmt.Sprintf("tenant-%d", idx), 1, 10)
		}(i)
	}

	close(start)
	wg.Wait()

	if count := queryCount.Load(); count != int32(numKeys) {
		t.Errorf("different keys should not be deduplicated: expected %d db queries, got %d", numKeys, count)
	}
}
