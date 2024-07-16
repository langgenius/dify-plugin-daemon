package plugin_manager

import (
	"sync"
	"time"

	"github.com/google/uuid"
	"github.com/langgenius/dify-plugin-daemon/internal/types/app"
	"github.com/langgenius/dify-plugin-daemon/internal/types/entities"
	"github.com/langgenius/dify-plugin-daemon/internal/types/entities/plugin_entities"
	"github.com/langgenius/dify-plugin-daemon/internal/utils/cache"
	"github.com/langgenius/dify-plugin-daemon/internal/utils/log"
	"github.com/langgenius/dify-plugin-daemon/internal/utils/parser"
)

const (
	KEY_PLUGIN_LIFETIME_STATE             = "lifetime_state"
	KEY_PLUGIN_LIFETIME_STATE_MODIFY_LOCK = "lifetime_state_modify_lock"
)

type PluginLifeTime struct {
	Identity string                            `json:"identity"`
	Restarts int                               `json:"restarts"`
	Status   string                            `json:"status"`
	Config   plugin_entities.PluginDeclaration `json:"configuration"`
}

type pluginLifeCollection struct {
	Collection  map[string]PluginLifeTime `json:"state"`
	ID          string                    `json:"id"`
	LastCheckAt time.Time                 `json:"last_check_at"`
}

func (p pluginLifeCollection) MarshalBinary() ([]byte, error) {
	return parser.MarshalJsonBytes(p), nil
}

var (
	instanceId = uuid.New().String()

	pluginLifetimeStateLock  = sync.RWMutex{}
	pluginLifetimeCollection = pluginLifeCollection{
		Collection: map[string]PluginLifeTime{},
		ID:         instanceId,
	}
)

func startLifeTimeManager(config *app.Config) {
	go func() {
		// do check immediately
		doClusterLifetimeCheck(config.LifetimeCollectionGCInterval)

		duration := time.Duration(config.LifetimeCollectionHeartbeatInterval) * time.Second
		for range time.NewTicker(duration).C {
			doClusterLifetimeCheck(config.LifetimeCollectionGCInterval)
		}
	}()
}

func doClusterLifetimeCheck(heartbeat_interval int) {
	// check and update self lifetime state
	if err := updateCurrentInstanceLifetimeCollection(); err != nil {
		log.Error("update current instance lifetime state failed: %s", err.Error())
		return
	}

	// lock cluster and do cluster lifetime check
	if cache.Lock(KEY_PLUGIN_LIFETIME_STATE_MODIFY_LOCK, time.Second*10, time.Second*10) != nil {
		log.Error("update lifetime state failed: lock failed")
		return
	}
	defer cache.Unlock(KEY_PLUGIN_LIFETIME_STATE_MODIFY_LOCK)

	cluster_lifetime_collections, err := fetchClusterPluginLifetimeCollections()
	if err != nil {
		log.Error("fetch cluster plugin lifetime state failed: %s", err.Error())
		return
	}

	for cluster_id, state := range cluster_lifetime_collections {
		if state.ID == instanceId {
			continue
		}

		// skip if last check has been done in $LIFETIME_COLLECTION_CG_INTERVAL
		cg_interval := time.Duration(heartbeat_interval) * time.Second
		if time.Since(state.LastCheckAt) < cg_interval {
			continue
		}

		// if last check has not been done in $LIFETIME_COLLECTION_CG_INTERVAL * 2, delete it
		if time.Since(state.LastCheckAt) > cg_interval*2 {
			if err := cache.DelMapField(KEY_PLUGIN_LIFETIME_STATE, cluster_id); err != nil {
				log.Error("delete cluster plugin lifetime state failed: %s", err.Error())
			} else {
				log.Info("delete cluster plugin lifetime state due to no longer active: %s", cluster_id)
			}
		}
	}
}

func newLifetimeFromRuntimeState(state entities.PluginRuntimeInterface) PluginLifeTime {
	s := state.RuntimeState()
	c := state.Configuration()

	return PluginLifeTime{
		Identity: c.Identity(),
		Restarts: s.Restarts,
		Status:   s.Status,
		Config:   *c,
	}
}

func addLifetimeState(state entities.PluginRuntimeInterface) {
	pluginLifetimeStateLock.Lock()
	defer pluginLifetimeStateLock.Unlock()

	pluginLifetimeCollection.Collection[state.Configuration().Identity()] = newLifetimeFromRuntimeState(state)
}

func deleteLifetimeState(state entities.PluginRuntimeInterface) {
	pluginLifetimeStateLock.Lock()
	defer pluginLifetimeStateLock.Unlock()

	delete(pluginLifetimeCollection.Collection, state.Configuration().Identity())
}

func updateCurrentInstanceLifetimeCollection() error {
	pluginLifetimeStateLock.Lock()
	defer pluginLifetimeStateLock.Unlock()

	pluginLifetimeCollection.LastCheckAt = time.Now()

	m.Range(func(key, value interface{}) bool {
		if v, ok := value.(entities.PluginRuntimeInterface); ok {
			pluginLifetimeCollection.Collection[v.Configuration().Identity()] = newLifetimeFromRuntimeState(v)
		}
		return true
	})

	return cache.SetMapOneField(KEY_PLUGIN_LIFETIME_STATE, instanceId, pluginLifetimeCollection)
}

func fetchClusterPluginLifetimeCollections() (map[string]pluginLifeCollection, error) {
	return cache.GetMap[pluginLifeCollection](KEY_PLUGIN_LIFETIME_STATE)
}
