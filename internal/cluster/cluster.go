package cluster

import (
	"sync"
	"sync/atomic"
	"time"

	"github.com/google/uuid"
	"github.com/langgenius/dify-plugin-daemon/internal/core/plugin_manager"
	"github.com/langgenius/dify-plugin-daemon/internal/types/app"
	"github.com/langgenius/dify-plugin-daemon/internal/utils/mapping"
)

type Cluster struct {
	// id is the unique id of the cluster
	id string

	// i_am_master is the flag to indicate whether the current node is the master node
	iAmMaster bool

	// main http port of the current node
	port uint16

	// plugins stores all the plugin life time of the current node
	plugins    mapping.Map[string, *pluginLifeTime]
	pluginLock sync.RWMutex

	manager *plugin_manager.PluginManager

	// nodes stores all the nodes of the cluster
	nodes mapping.Map[string, node]

	// signals for waiting for the cluster to stop
	stopChan chan bool
	stopped  int32

	isInAutoGcNodes   int32
	isInAutoGcPlugins int32

	// channels to notify cluster event
	notifyBecomeMasterChan            chan bool
	notifyMasterGcChan                chan bool
	notifyMasterGcCompletedChan       chan bool
	notifyVotingChan                  chan bool
	notifyVotingCompletedChan         chan bool
	notifyPluginScheduleChan          chan bool
	notifyPluginScheduleCompletedChan chan bool
	notifyNodeUpdateChan              chan bool
	notifyNodeUpdateCompletedChan     chan bool
	notifyClusterStoppedChan          chan bool

	showLog bool

	masterGcInterval              time.Duration
	masterLockingInterval         time.Duration
	masterLockExpiredTime         time.Duration
	nodeVoteInterval              time.Duration
	nodeDisconnectedTimeout       time.Duration
	updateNodeStatusInterval      time.Duration
	pluginSchedulerInterval       time.Duration
	pluginSchedulerTickerInterval time.Duration
	pluginDeactivatedTimeout      time.Duration
}

func NewCluster(config *app.Config, plugin_manager *plugin_manager.PluginManager) *Cluster {
	return &Cluster{
		id:                            uuid.New().String(),
		port:                          uint16(config.ServerPort),
		stopChan:                      make(chan bool),
		showLog:                       config.DisplayClusterLog,
		masterGcInterval:              MASTER_GC_INTERVAL,
		masterLockingInterval:         MASTER_LOCKING_INTERVAL,
		masterLockExpiredTime:         MASTER_LOCK_EXPIRED_TIME,
		nodeVoteInterval:              NODE_VOTE_INTERVAL,
		nodeDisconnectedTimeout:       NODE_DISCONNECTED_TIMEOUT,
		updateNodeStatusInterval:      UPDATE_NODE_STATUS_INTERVAL,
		pluginSchedulerInterval:       PLUGIN_SCHEDULER_INTERVAL,
		pluginSchedulerTickerInterval: PLUGIN_SCHEDULER_TICKER_INTERVAL,
		pluginDeactivatedTimeout:      PLUGIN_DEACTIVATED_TIMEOUT,

		manager: plugin_manager,

		notifyBecomeMasterChan:            make(chan bool),
		notifyMasterGcChan:                make(chan bool),
		notifyMasterGcCompletedChan:       make(chan bool),
		notifyVotingChan:                  make(chan bool),
		notifyVotingCompletedChan:         make(chan bool),
		notifyPluginScheduleChan:          make(chan bool),
		notifyPluginScheduleCompletedChan: make(chan bool),
		notifyNodeUpdateChan:              make(chan bool),
		notifyNodeUpdateCompletedChan:     make(chan bool),
		notifyClusterStoppedChan:          make(chan bool),
	}
}

func (c *Cluster) Launch() {
	go c.clusterLifetime()
}

func (c *Cluster) Close() error {
	if atomic.CompareAndSwapInt32(&c.stopped, 0, 1) {
		close(c.stopChan)
	}

	return nil
}

func (c *Cluster) ID() string {
	return c.id
}

// trigger for master event
func (c *Cluster) notifyBecomeMaster() {
	if atomic.LoadInt32(&c.stopped) == 1 {
		return
	}

	select {
	case c.notifyBecomeMasterChan <- true:
	default:
	}
}

// receive the master event
func (c *Cluster) NotifyBecomeMaster() <-chan bool {
	if atomic.LoadInt32(&c.stopped) == 1 {
		return nil
	}
	return c.notifyBecomeMasterChan
}

// trigger for master gc event
func (c *Cluster) notifyMasterGC() {
	if atomic.LoadInt32(&c.stopped) == 1 {
		return
	}

	select {
	case c.notifyMasterGcChan <- true:
	default:
	}
}

// trigger for master gc completed event
func (c *Cluster) notifyMasterGCCompleted() {
	if atomic.LoadInt32(&c.stopped) == 1 {
		return
	}

	select {
	case c.notifyMasterGcCompletedChan <- true:
	default:
	}
}

// trigger for voting event
func (c *Cluster) notifyVoting() {
	if atomic.LoadInt32(&c.stopped) == 1 {
		return
	}

	select {
	case c.notifyVotingChan <- true:
	default:
	}
}

// trigger for voting completed event
func (c *Cluster) notifyVotingCompleted() {
	if atomic.LoadInt32(&c.stopped) == 1 {
		return
	}

	select {
	case c.notifyVotingCompletedChan <- true:
	default:
	}
}

// trigger for plugin schedule event
func (c *Cluster) notifyPluginSchedule() {
	if atomic.LoadInt32(&c.stopped) == 1 {
		return
	}

	select {
	case c.notifyPluginScheduleChan <- true:
	default:
	}
}

// trigger for plugin schedule completed event
func (c *Cluster) notifyPluginScheduleCompleted() {
	if atomic.LoadInt32(&c.stopped) == 1 {
		return
	}

	select {
	case c.notifyPluginScheduleCompletedChan <- true:
	default:
	}
}

// trigger for node update event
func (c *Cluster) notifyNodeUpdate() {
	if atomic.LoadInt32(&c.stopped) == 1 {
		return
	}

	select {
	case c.notifyNodeUpdateChan <- true:
	default:
	}
}

// trigger for node update completed event
func (c *Cluster) notifyNodeUpdateCompleted() {
	if atomic.LoadInt32(&c.stopped) == 1 {
		return
	}

	select {
	case c.notifyNodeUpdateCompletedChan <- true:
	default:
	}
}

// trigger for cluster stopped event
func (c *Cluster) notifyClusterStopped() {
	select {
	case c.notifyClusterStoppedChan <- true:
	default:
	}
}

// receive the master gc event
func (c *Cluster) NotifyMasterGC() <-chan bool {
	if atomic.LoadInt32(&c.stopped) == 1 {
		return nil
	}
	return c.notifyMasterGcChan
}

// receive the master gc completed event
func (c *Cluster) NotifyMasterGCCompleted() <-chan bool {
	if atomic.LoadInt32(&c.stopped) == 1 {
		return nil
	}
	return c.notifyMasterGcCompletedChan
}

// receive the voting event
func (c *Cluster) NotifyVoting() <-chan bool {
	if atomic.LoadInt32(&c.stopped) == 1 {
		return nil
	}
	return c.notifyVotingChan
}

// receive the voting completed event
func (c *Cluster) NotifyVotingCompleted() <-chan bool {
	if atomic.LoadInt32(&c.stopped) == 1 {
		return nil
	}
	return c.notifyVotingCompletedChan
}

// receive the plugin schedule event
func (c *Cluster) NotifyPluginSchedule() <-chan bool {
	if atomic.LoadInt32(&c.stopped) == 1 {
		return nil
	}
	return c.notifyPluginScheduleChan
}

// receive the plugin schedule completed event
func (c *Cluster) NotifyPluginScheduleCompleted() <-chan bool {
	if atomic.LoadInt32(&c.stopped) == 1 {
		return nil
	}
	return c.notifyPluginScheduleCompletedChan
}

// receive the node update event
func (c *Cluster) NotifyNodeUpdate() <-chan bool {
	if atomic.LoadInt32(&c.stopped) == 1 {
		return nil
	}
	return c.notifyNodeUpdateChan
}

// receive the node update completed event
func (c *Cluster) NotifyNodeUpdateCompleted() <-chan bool {
	if atomic.LoadInt32(&c.stopped) == 1 {
		return nil
	}
	return c.notifyNodeUpdateCompletedChan
}

// receive the cluster stopped event
func (c *Cluster) NotifyClusterStopped() <-chan bool {
	return c.notifyClusterStoppedChan
}
