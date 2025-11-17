package local_runtime

import "sync/atomic"

// load balancing is a mechanism to distribute the workload across multiple instances of the plugin
// NOTE: currently, we only support round robin
func (r *LocalPluginRuntime) pickLowestLoadInstance() (*PluginInstance, error) {
	// Just a round robin
	idx := atomic.AddInt64(&r.roundRobinIndex, 1)

	// lock the instances to avoid array out of bounds
	r.instanceLocker.RLock()
	defer r.instanceLocker.RUnlock()

	return r.instances[idx%int64(len(r.instances))], nil
}
