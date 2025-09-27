package plugin_manager

import (
    "sync/atomic"

    "github.com/langgenius/dify-plugin-daemon/internal/utils/log"
    "github.com/langgenius/dify-plugin-daemon/pkg/entities/plugin_entities"
)

type runtimeTrafficState struct {
    sessions atomic.Int64
    draining atomic.Bool
}

type registerRuntimeOptions struct {
	blueGreen bool
}

type RuntimeTrafficReport struct {
	Identity string
	PluginID string
	Sessions int64
	Draining bool
}

func (p *PluginManager) ensureRuntimeState(identity string) *runtimeTrafficState {
	stateAny, _ := p.runtimeSessions.LoadOrStore(identity, &runtimeTrafficState{})
	state, ok := stateAny.(*runtimeTrafficState)
	if !ok {
		// this should never happen, but protect against panic if the map was polluted
		newState := &runtimeTrafficState{}
		p.runtimeSessions.Store(identity, newState)
		return newState
	}
	return state
}

func (p *PluginManager) cleanupRuntime(identity string) {
	p.runtimeSessions.Delete(identity)
	p.runtimePluginIDs.Delete(identity)
}

func (p *PluginManager) loadRuntimeState(identity string) (*runtimeTrafficState, bool) {
	stateAny, ok := p.runtimeSessions.Load(identity)
	if !ok {
		return nil, false
	}
	state, ok := stateAny.(*runtimeTrafficState)
	if !ok {
		return nil, false
	}
	return state, true
}

func (p *PluginManager) stopRuntime(identity string) {
	log.Info("[blue-green] stop runtime immediately: %s", identity)
	if runtime, ok := p.m.Load(identity); ok {
		runtime.Stop()
	}
	p.cleanupRuntime(identity)
}

func (p *PluginManager) markRuntimeDraining(identity string) {
    log.Info("[blue-green] mark runtime draining: %s", identity)
    stateAny, ok := p.runtimeSessions.Load(identity)
    if ok {
        if state, ok := stateAny.(*runtimeTrafficState); ok {
            current := state.sessions.Load()
            log.Info("[blue-green] set draining=true: identity=%s current_sessions=%d", identity, current)
            state.draining.Store(true)
            if current == 0 {
                if runtime, ok := p.m.Load(identity); ok {
                    log.Info("[blue-green] draining runtime has no sessions, stopping: %s", identity)
                    runtime.Stop()
                }
                // 若记录了需要清理安装包，则在停止后执行删除
                if _, need := p.drainingCleanup.Load(identity); need {
                    if uid, err := plugin_entities.NewPluginUniqueIdentifier(identity); err == nil {
                        if err := p.installedBucket.Delete(uid); err != nil {
                            log.Warn("[blue-green] cleanup installed package failed: identity=%s err=%s", identity, err.Error())
                        } else {
                            log.Info("[blue-green] cleaned installed package for drained runtime: %s", identity)
                        }
                    }
                    p.drainingCleanup.Delete(identity)
                }
                p.cleanupRuntime(identity)
            }
            return
        }
    }

    // 方案A：状态缺失时不立即停止，避免中断在途业务。记录为 draining-pending，后续在 AcquireRuntime 桥接。
    log.Warn("[blue-green] runtime state missing on draining; mark pending and skip stop: %s", identity)
    p.drainingPending.Store(identity, true)
}

// registerRuntime only records mapping and initializes state; it does NOT perform cutover.
func (p *PluginManager) registerRuntime(
	identity plugin_entities.PluginUniqueIdentifier,
	options registerRuntimeOptions,
) {
	identityStr := identity.String()
	pluginID := identity.PluginID()
	p.runtimePluginIDs.Store(identityStr, pluginID)

	state := p.ensureRuntimeState(identityStr)
	state.draining.Store(false)
	log.Info("[blue-green] registered runtime: identity=%s plugin_id=%s blue_green=%v", identityStr, pluginID, options.blueGreen)
}

// finalizeRuntimeRegistration performs traffic cutover after new runtime is ready.
func (p *PluginManager) finalizeRuntimeRegistration(
    identity plugin_entities.PluginUniqueIdentifier,
    blueGreen bool,
) {
	identityStr := identity.String()
	pluginID := identity.PluginID()
	log.Info("[blue-green] finalize runtime registration: identity=%s plugin_id=%s blue_green=%v", identityStr, pluginID, blueGreen)

    if blueGreen {
        p.runtimePluginIDs.Range(func(key, value any) bool {
            identityKey := key.(string)
            if identityKey == identityStr {
                return true
            }
            if value.(string) == pluginID {
                log.Info("[blue-green] will drain old runtime of same plugin_id: %s", identityKey)
                p.markRuntimeDraining(identityKey)
                // 记录该旧版本在停止后需要清理安装包
                p.drainingCleanup.Store(identityKey, true)
            }
            return true
        })
        return
    }

	p.runtimePluginIDs.Range(func(key, value any) bool {
		identityKey := key.(string)
		if identityKey == identityStr {
			return true
		}
		if value.(string) == pluginID {
			log.Info("[blue-green] will stop old runtime of same plugin_id: %s", identityKey)
			p.stopRuntime(identityKey)
		}
		return true
	})
}

// FinalizeRuntimeRegistration exposes finalize for other packages (e.g., service layer)
func (p *PluginManager) FinalizeRuntimeRegistration(
    identity plugin_entities.PluginUniqueIdentifier,
    blueGreen bool,
) {
    p.finalizeRuntimeRegistration(identity, blueGreen)
}

func (p *PluginManager) AcquireRuntime(identity plugin_entities.PluginUniqueIdentifier) (plugin_entities.PluginLifetime, func(), error) {
    runtime, err := p.Get(identity)
    if err != nil {
        return nil, nil, err
    }

    state := p.ensureRuntimeState(identity.String())
    current := state.sessions.Add(1)
    // 若此前标记了 draining-pending，则在首次获取时桥接为真实 draining 状态
    if pendingAny, ok := p.drainingPending.Load(identity.String()); ok {
        if pending, _ := pendingAny.(bool); pending {
            state.draining.Store(true)
            p.drainingPending.Delete(identity.String())
            log.Info("[blue-green] bridge pending->draining on acquire: identity=%s active_sessions=%d", identity.String(), current)
        }
    }
    log.Info("[traffic] acquire runtime: identity=%s active_sessions=%d draining=%v", identity.String(), current, state.draining.Load())

    var released atomic.Bool
    release := func() {
		if !released.CompareAndSwap(false, true) {
			return
		}

		remaining := state.sessions.Add(-1)
		log.Info("[traffic] release runtime: identity=%s remaining_sessions=%d draining=%v", identity.String(), remaining, state.draining.Load())
		if remaining < 0 {
			state.sessions.Store(0)
			remaining = 0
		}

        if state.draining.Load() && remaining == 0 {
            log.Info("[blue-green] draining runtime now has zero sessions, stopping: %s", identity.String())
            if runtime, ok := p.m.Load(identity.String()); ok {
                runtime.Stop()
            }
            // 若记录了需要清理安装包，则在停止后执行删除
            if _, need := p.drainingCleanup.Load(identity.String()); need {
                if uid, err := plugin_entities.NewPluginUniqueIdentifier(identity.String()); err == nil {
                    if err := p.installedBucket.Delete(uid); err != nil {
                        log.Warn("[blue-green] cleanup installed package failed: identity=%s err=%s", identity.String(), err.Error())
                    } else {
                        log.Info("[blue-green] cleaned installed package for drained runtime: %s", identity.String())
                    }
                }
                p.drainingCleanup.Delete(identity.String())
            }
            p.cleanupRuntime(identity.String())
        }
    }

	return runtime, release, nil
}

func (p *PluginManager) CollectRuntimeTraffic(filterPluginID string) []RuntimeTrafficReport {
	reports := make([]RuntimeTrafficReport, 0)

	p.runtimePluginIDs.Range(func(key, value any) bool {
		identity := key.(string)
		pluginID := value.(string)
		if filterPluginID != "" && filterPluginID != pluginID {
			return true
		}

		var (
			sessions int64
			draining bool
		)

		if state, ok := p.loadRuntimeState(identity); ok {
			sessions = state.sessions.Load()
			draining = state.draining.Load()
		}

		reports = append(reports, RuntimeTrafficReport{
			Identity: identity,
			PluginID: pluginID,
			Sessions: sessions,
			Draining: draining,
		})
		return true
	})

	return reports
}
