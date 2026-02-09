package controlpanel

import (
	"sync"
	"time"

	"github.com/langgenius/dify-plugin-daemon/pkg/entities/plugin_entities"
)

type LocalReadinessSnapshot struct {
	// ⭐ Core: readiness is only based on initial plugin state
	// Once Pod is ready, it will never become not ready due to runtime new plugins
	Ready bool

	// Initial plugin state (locked at Pod startup, never changed afterward)
	InitialPluginsReady bool
	InitialExpected     int
	InitialRunning      int
	InitialMissing      []string
	InitialFailed       []string

	// Runtime added plugin state (not related to readiness, for monitoring only)
	RuntimePluginsLoading int
	RuntimeMissing        []string

	// Total statistics (including initial + runtime)
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
	lock  sync.RWMutex
	ids   map[string]bool // plugin id → true
	ready bool            // whether it has been locked
}

type initialPluginsStatus struct {
	ready   bool
	expected int
	running int
	missing []string
	failed  []string
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

	// Calculate total plugin state
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

	// Calculate initial plugin state
	initialStatus := c.isInitialPluginsReady(expected)

	// Calculate runtime added plugins
	runtimeMissing := make([]string, 0)
	runtimeLoading := 0

	initialSet := c.getInitialPluginSet()
	for _, id := range expected {
		idStr := id.String()
		if !initialSet[idStr] {
			// This is a plugin added at runtime
			if !c.localPluginRuntimes.Exists(id) {
				if retry, ok := c.localPluginFailsRecord.Load(id); !ok || retry.RetryCount < c.config.PluginLocalMaxRetryCount {
					runtimeMissing = append(runtimeMissing, idStr)
					runtimeLoading++
				}
			}
		}
	}

	// 🔑 Key: readiness ONLY depends on initial plugins
	// Once ready, it will never become not ready due to runtime plugin additions
	snapshot := &LocalReadinessSnapshot{
		Ready:                 initialStatus.ready,
		InitialPluginsReady:   initialStatus.ready,
		InitialExpected:       initialStatus.expected,
		InitialRunning:        initialStatus.running,
		InitialMissing:        initialStatus.missing,
		InitialFailed:         initialStatus.failed,
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

// isInitialPluginsReady checks if all initial plugins have been started
func (c *ControlPanel) isInitialPluginsReady(
	current []plugin_entities.PluginUniqueIdentifier,
) initialPluginsStatus {
	initialSet := c.getInitialPluginSet()
	if len(initialSet) == 0 && len(current) > 0 {
		// First startup, lock the initial plugin set
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

	return initialPluginsStatus{
		ready:    len(missingList) == 0,
		expected: expected,
		running:  running,
		missing:  missingList,
		failed:   failedList,
	}
}

// lockInitialPlugins locks the initial plugin set (only on first call)
func (c *ControlPanel) lockInitialPlugins(
	plugins []plugin_entities.PluginUniqueIdentifier,
) {
	c.initialPlugins.lock.Lock()
	defer c.initialPlugins.lock.Unlock()

	if c.initialPlugins.ready {
		return
	}

	for _, id := range plugins {
		c.initialPlugins.ids[id.String()] = true
	}
	c.initialPlugins.ready = true
}

// getInitialPluginSet returns the initial plugin set (read-only)
func (c *ControlPanel) getInitialPluginSet() map[string]bool {
	c.initialPlugins.lock.RLock()
	defer c.initialPlugins.lock.RUnlock()

	result := make(map[string]bool)
	for k, v := range c.initialPlugins.ids {
		result[k] = v
	}
	return result
}
