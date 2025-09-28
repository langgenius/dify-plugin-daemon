package plugin_manager

import (
	"errors"
	"sync/atomic"

	"github.com/langgenius/dify-plugin-daemon/internal/core/session_manager"
	"github.com/langgenius/dify-plugin-daemon/internal/utils/log"
	"github.com/langgenius/dify-plugin-daemon/pkg/entities/plugin_entities"
)

type runtimeTrafficState struct {
	sessions atomic.Int64
	draining atomic.Bool
	// true: auto stop drained runtime;
	// false: require manual approval
	autoStop atomic.Bool
}

type registerRuntimeOptions struct {
	blueGreen bool
}

type BlueGreenMode string

const (
	BLUE_GREEN_MODE_AUTO   BlueGreenMode = "auto"
	BLUE_GREEN_MODE_MANUAL BlueGreenMode = "manual"
)

type RuntimeTrafficReport struct {
	Identity string
	PluginID string
	Sessions int64
	Draining bool
	AutoStop bool
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
			if current == 0 && state.autoStop.Load() {
				if runtime, ok := p.m.Load(identity); ok {
					log.Info("[blue-green] draining runtime has no sessions, stopping: %s", identity)
					runtime.Stop()
				}

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

	log.Warn("[blue-green] runtime state missing on draining; checking zero-connection stop: %s", identity)
	if session_manager.GetPluginConnections(identity) == 0 {
		// stop immediately and cleanup
		if runtime, ok := p.m.Load(identity); ok {
			log.Info("[blue-green] stopping drained runtime with missing state: %s", identity)
			runtime.Stop()
		}
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
		return
	}

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
	state.autoStop.Store(true)
	log.Info("[blue-green] registered runtime: identity=%s plugin_id=%s blue_green=%v", identityStr, pluginID, options.blueGreen)
}

// finalizeRuntimeRegistration performs traffic cutover after new runtime is ready.
func (p *PluginManager) finalizeRuntimeRegistration(
	identity plugin_entities.PluginUniqueIdentifier,
	mode BlueGreenMode,
) {
	identityStr := identity.String()
	pluginID := identity.PluginID()
	log.Info("[blue-green] finalize runtime registration: identity=%s plugin_id=%s mode=%s", identityStr, pluginID, string(mode))

	if mode == BLUE_GREEN_MODE_AUTO || mode == BLUE_GREEN_MODE_MANUAL {
		p.runtimePluginIDs.Range(func(key, value any) bool {
			identityKey := key.(string)
			if identityKey == identityStr {
				return true
			}
			if value.(string) == pluginID {
				log.Info("[blue-green] will drain old runtime of same plugin_id: %s (mode=%s)", identityKey, string(mode))
				// ensure autoStop is set BEFORE marking draining so that immediate stop can happen when sessions==0
				if state, ok := p.loadRuntimeState(identityKey); ok {
					state.autoStop.Store(mode == BLUE_GREEN_MODE_AUTO)
				}
				p.markRuntimeDraining(identityKey)
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

func (p *PluginManager) FinalizeRuntimeRegistration(
	identity plugin_entities.PluginUniqueIdentifier,
	mode BlueGreenMode,
) {
	p.finalizeRuntimeRegistration(identity, mode)
}

func (p *PluginManager) AcquireRuntime(identity plugin_entities.PluginUniqueIdentifier) (plugin_entities.PluginLifetime, func(), error) {
	runtime, err := p.Get(identity)
	if err != nil {
		return nil, nil, err
	}

	state := p.ensureRuntimeState(identity.String())
	current := state.sessions.Add(1)

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

		if state.draining.Load() && remaining == 0 && state.autoStop.Load() {
			log.Info("[blue-green] draining runtime now has zero sessions, stopping: %s", identity.String())
			if runtime, ok := p.m.Load(identity.String()); ok {
				runtime.Stop()
			}

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
			autoStop bool
		)

		if state, ok := p.loadRuntimeState(identity); ok {
			sessions = state.sessions.Load()
			draining = state.draining.Load()
			autoStop = state.autoStop.Load()
		}

		reports = append(reports, RuntimeTrafficReport{
			Identity: identity,
			PluginID: pluginID,
			Sessions: sessions,
			Draining: draining,
			AutoStop: autoStop,
		})
		return true
	})

	return reports
}

func (p *PluginManager) ApproveBlueGreen(pluginID string) error {
	canStop := true
	ids := make([]string, 0)
	p.runtimePluginIDs.Range(func(key, value any) bool {
		identity := key.(string)
		pid := value.(string)
		if pid != pluginID {
			return true
		}
		if state, ok := p.loadRuntimeState(identity); ok {
			if state.draining.Load() {
				if state.sessions.Load() == 0 {
					ids = append(ids, identity)
				} else {
					canStop = false
				}
			}
		}
		return true
	})

	if !canStop {
		return errors.New("blue-green approve blocked: some draining runtimes still have active sessions")
	}

	for _, identity := range ids {
		if runtime, ok := p.m.Load(identity); ok {
			runtime.Stop()
		}
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
	return nil
}

// Not yet enabled on the frontend
// But the API can be called
func (p *PluginManager) ForceOffline(identity string) {
	if runtime, ok := p.m.Load(identity); ok {
		runtime.Stop()
	}
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

func (p *PluginManager) rollbackRuntimeRegistration(oldIdentity plugin_entities.PluginUniqueIdentifier) {
	oldStr := oldIdentity.String()
	pluginID := oldIdentity.PluginID()
	log.Info("[blue-green] rollback runtime registration: old_identity=%s plugin_id=%s", oldStr, pluginID)

	p.runtimePluginIDs.Range(func(key, value any) bool {
		identityKey := key.(string)
		if value.(string) != pluginID {
			return true
		}
		if state, ok := p.loadRuntimeState(identityKey); ok {
			if identityKey == oldStr {
				state.draining.Store(false)
			} else {
				state.draining.Store(true)
				state.autoStop.Store(true)
				// mark this draining identity to cleanup installed package after stop
				p.drainingCleanup.Store(identityKey, true)
				// if draining runtime already has zero sessions, stop it immediately
				if state.sessions.Load() == 0 {
					log.Info("[blue-green] draining runtime has zero sessions on rollback, stopping: %s", identityKey)
					if runtime, ok := p.m.Load(identityKey); ok {
						runtime.Stop()
					}
					if _, need := p.drainingCleanup.Load(identityKey); need {
						if uid, err := plugin_entities.NewPluginUniqueIdentifier(identityKey); err == nil {
							if err := p.installedBucket.Delete(uid); err != nil {
								log.Warn("[blue-green] cleanup installed package failed: identity=%s err=%s", identityKey, err.Error())
							} else {
								log.Info("[blue-green] cleaned installed package for drained runtime: %s", identityKey)
							}
						}
						p.drainingCleanup.Delete(identityKey)
					}
					p.cleanupRuntime(identityKey)
				}
			}
		}
		return true
	})
}

func (p *PluginManager) RollbackRuntimeRegistration(oldIdentity plugin_entities.PluginUniqueIdentifier) {
	p.rollbackRuntimeRegistration(oldIdentity)
}
