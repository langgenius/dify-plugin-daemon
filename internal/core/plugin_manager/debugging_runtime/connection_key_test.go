package debugging_runtime

import (
	"testing"

	"github.com/google/uuid"
	"github.com/langgenius/dify-plugin-daemon/internal/utils/cache"
)

func TestConnectionKey(t *testing.T) {
	err := cache.InitRedisClient("0.0.0.0:6379", "difyai123456", false, 0)
	if err != nil {
		t.Errorf("init redis client failed: %v", err)
		return
	}
	defer cache.Close()

	// test connection key
	key, err := GetConnectionKey(ConnectionInfo{
		TenantId: "abc",
	})

	if err != nil {
		t.Errorf("get connection key failed: %v", err)
		return
	}

	defer ClearConnectionKey("abc")

	_, err = uuid.Parse(key)
	if err != nil {
		t.Errorf("connection key is not a valid uuid: %v", err)
		return
	}

	// test connection key with the same tenant id
	key2, err := GetConnectionKey(ConnectionInfo{
		TenantId: "abc",
	})

	if err != nil {
		t.Errorf("get connection key failed: %v", err)
		return
	}

	if key != key2 {
		t.Errorf("connection key is not the same: %s, %s", key, key2)
		return
	}

	connectionInfo, err := GetConnectionInfo(key)
	if err != nil {
		t.Errorf("get connection info failed: %v", err)
		return
	}

	if connectionInfo.TenantId != "abc" {
		t.Errorf("connection info is not the same: %v", connectionInfo)
		return
	}
}
