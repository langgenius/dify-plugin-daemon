package plugin_manager

import (
	"testing"
	"time"

	"github.com/langgenius/dify-plugin-daemon/internal/core/plugin_manager/local_runtime"
	"github.com/langgenius/dify-plugin-daemon/internal/core/plugin_manager/test_utils"
	"github.com/langgenius/dify-plugin-daemon/internal/utils/log"
	"github.com/langgenius/dify-plugin-daemon/internal/utils/routine"
)

func TestLocalLaunch(t *testing.T) {
	log.SetShowLog(false)

	routine.InitPool(100000)
	defer test_utils.ClearTestingPath()

	runtime, err := test_utils.GetOpenAIRuntime(false, 1, 1)
	if err != nil {
		t.Fatal(err)
	}

	// wait for plugin launched
	ticker := time.NewTicker(time.Second * 1)
	defer ticker.Stop()
	timeout := time.After(time.Second * 20)

	for {
		select {
		case <-ticker.C:
			if runtime.Stage() == local_runtime.LAUNCH_STAGE_VERIFIED_WORKING {
				return
			}
		case <-timeout:
			t.Fatal("plugin not launched")
		}
	}
}

func TestLocalLaunchFailed(t *testing.T) {
	log.SetShowLog(false)

	routine.InitPool(100000)
	defer test_utils.ClearTestingPath()

	_, err := test_utils.GetBrokenRuntime(false, 1, 1)
	if err == nil {
		t.Fatal("plugin launched")
	}
}
