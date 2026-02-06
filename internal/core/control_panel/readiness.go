package controlpanel

import (
	"sync"
	"time"

	"github.com/langgenius/dify-plugin-daemon/pkg/entities/plugin_entities"
)

type LocalReadinessSnapshot struct {
	// ⭐ 核心：readiness 只基于初始插件状态
	// Pod 一旦 ready，永远不会因为运行时新增插件而变为 not ready
	Ready bool

	// 初始插件状态（Pod启动时锁定，之后永不改变）
	InitialPluginsReady bool
	InitialExpected     int
	InitialRunning      int
	InitialMissing      []string
	InitialFailed       []string

	// 运行时新增插件状态（与readiness无关，仅供监控）
	RuntimePluginsLoading int
	RuntimeMissing        []string

	// 全量统计（包含初始+运行时）
	Expected   int
	Running    int
	Missing    []string
	Failed     []string
	UpdatedAt  time.Time
	Platform   string
	Installed  int
	Ignored    int
	MaxRetries int32
}

type initialPluginSet struct {
	lock    sync.RWMutex
	ids     map[string]bool  // plugin id → true
	ready   bool             // 是否已锁定
}

var initialPlugins = &initialPluginSet{
	ids: make(map[string]bool),
}

func (c *ControlPanel) LocalReadiness() (LocalReadinessSnapshot, bool) {
	ptr := c.localReadinessSnapshot.Load()
	if ptr == nil {
		return LocalReadinessSnapshot{}, false
	}
	return *ptr, true
}

func (c *ControlPanel) updateLocalReadinessSnapshot(
	installed []plugin_entities.PluginUniqueIdentifier,
) {
	now := time.Now()

	expected := make([]plugin_entities.PluginUniqueIdentifier, 0, len(installed))
	ignored := 0
	for _, id := range installed {
		if _, ok := c.localPluginWatchIgnoreList.Load(id); ok {
			ignored++
			continue
		}
		expected = append(expected, id)
	}

	// 计算全量插件状态
	missing := make([]string, 0)
	failed := make([]string, 0)
	running := 0
	for _, id := range expected {
		if c.localPluginRuntimes.Exists(id) {
			running++
			continue
		}

		if retry, ok := c.localPluginFailsRecord.Load(id); ok && retry.RetryCount >= c.config.PluginLocalMaxRetryCount {
			failed = append(failed, id.String())
			continue
		}
		missing = append(missing, id.String())
	}

	// 计算初始插件的状态
	initialMissing := make([]string, 0)
	initialFailed := make([]string, 0)
	initialRunning := 0
	initialExpected := 0

	isInitialReady := c.isInitialPluginsReady(expected, &initialExpected, &initialRunning, &initialMissing, &initialFailed)

	// 计算运行时新增插件
	runtimeMissing := make([]string, 0)
	runtimeLoading := 0

	initialSet := c.getInitialPluginSet()
	for _, id := range expected {
		idStr := id.String()
		if !initialSet[idStr] {
			// 这是运行时新增的插件
			if !c.localPluginRuntimes.Exists(id) {
				if retry, ok := c.localPluginFailsRecord.Load(id); !ok || retry.RetryCount < c.config.PluginLocalMaxRetryCount {
					runtimeMissing = append(runtimeMissing, idStr)
					runtimeLoading++
				}
			}
		}
	}

	// 🔑 关键：readiness ONLY depends on initial plugins
	// Once ready, it will never become not ready due to runtime plugin additions
	snapshot := &LocalReadinessSnapshot{
		Ready:                 isInitialReady,
		InitialPluginsReady:   isInitialReady,
		InitialExpected:       initialExpected,
		InitialRunning:        initialRunning,
		InitialMissing:        initialMissing,
		InitialFailed:         initialFailed,
		RuntimePluginsLoading: runtimeLoading,
		RuntimeMissing:        runtimeMissing,
		Expected:              len(expected),
		Installed:             len(installed),
		Ignored:               ignored,
		Running:               running,
		Missing:               missing,
		Failed:                failed,
		UpdatedAt:             now,
		Platform:              string(c.config.Platform),
		MaxRetries:            c.config.PluginLocalMaxRetryCount,
	}
	c.localReadinessSnapshot.Store(snapshot)
}

// isInitialPluginsReady 检查初始插件是否全部启动完成
func (c *ControlPanel) isInitialPluginsReady(
	current []plugin_entities.PluginUniqueIdentifier,
	initialExpected *int,
	initialRunning *int,
	initialMissing *[]string,
	initialFailed *[]string,
) bool {
	initialSet := c.getInitialPluginSet()
	if len(initialSet) == 0 && len(current) > 0 {
		// 首次启动，锁定初始插件集合
		c.lockInitialPlugins(current)
		initialSet = c.getInitialPluginSet()
	}

	missingList := make([]string, 0)
	failedList := make([]string, 0)
	running := 0
	expected := 0

	for _, id := range current {
		idStr := id.String()
		if !initialSet[idStr] {
			continue
		}

		expected++
		if c.localPluginRuntimes.Exists(id) {
			running++
			continue
		}

		if retry, ok := c.localPluginFailsRecord.Load(id); ok && retry.RetryCount >= c.config.PluginLocalMaxRetryCount {
			failedList = append(failedList, idStr)
			continue
		}
		missingList = append(missingList, idStr)
	}

	*initialExpected = expected
	*initialRunning = running
	*initialMissing = missingList
	*initialFailed = failedList

	return len(missingList) == 0
}

// lockInitialPlugins 锁定初始插件集合（仅在首次调用时）
func (c *ControlPanel) lockInitialPlugins(
	plugins []plugin_entities.PluginUniqueIdentifier,
) {
	initialPlugins.lock.Lock()
	defer initialPlugins.lock.Unlock()

	if initialPlugins.ready {
		return
	}

	for _, id := range plugins {
		initialPlugins.ids[id.String()] = true
	}
	initialPlugins.ready = true
}

// getInitialPluginSet 获取初始插件集合（只读）
func (c *ControlPanel) getInitialPluginSet() map[string]bool {
	initialPlugins.lock.RLock()
	defer initialPlugins.lock.RUnlock()

	result := make(map[string]bool)
	for k, v := range initialPlugins.ids {
		result[k] = v
	}
	return result
}

