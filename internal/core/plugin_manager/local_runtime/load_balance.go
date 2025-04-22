package local_runtime

import (
	"time"

	"github.com/shirou/gopsutil/v3/process"
)

const (
	_SAMPLES = 5
)

func (r *LocalPluginRuntime) getLowestLoadPluginInstance() *pluginInstance {
	// get the lowest load plugin instance by considering both session count and CPU usage
	var lowestInstance *pluginInstance
	var lowestScore int64

	for _, s := range r.pluginInstances {
		// Calculate a combined score based on session count and CPU usage
		// Multiply session count by 5 to give it more weight in the calculation
		// This helps prevent many requests from flooding the same instance
		score := (int64(s.cpuUsagePercentSum) / _SAMPLES) + int64(s.sessionCount*5)

		if lowestInstance == nil || score < lowestScore {
			lowestInstance = s
			lowestScore = score
		}
	}

	return lowestInstance
}

func (s *pluginInstance) startUsageMonitor() {
	ticker := time.NewTicker(2 * time.Second)
	defer ticker.Stop()

	for !s.waitingControllerChanClosed {
		// check usage
		cpuUsagePercent, memoryUsage := s.getUsage()

		// remove the oldest usage and move all the others forward
		s.cpuUsagePercentSum -= int16(s.cpuUsagePercent[0])
		s.memoryUsageSum -= s.memoryUsage[0]

		for i := 0; i < _SAMPLES-1; i++ {
			s.cpuUsagePercent[i] = s.cpuUsagePercent[i+1]
			s.memoryUsage[i] = s.memoryUsage[i+1]
		}

		s.cpuUsagePercent[_SAMPLES-1] = int8(cpuUsagePercent)
		s.memoryUsage[_SAMPLES-1] = memoryUsage

		// increase the sum of the usage
		s.cpuUsagePercentSum += int16(cpuUsagePercent)
		s.memoryUsageSum += memoryUsage

		<-ticker.C
	}
}

func (s *pluginInstance) getUsage() (cpuUsagePercent float64, memoryUsage int64) {
	process, err := process.NewProcess(int32(s.pid))
	if err != nil {
		return 0, 0
	}
	cpuUsagePercent, err = process.CPUPercent()
	if err != nil {
		return 0, 0
	}
	memoryInfo, err := process.MemoryInfo()
	if err != nil {
		return 0, 0
	}
	memoryUsage = int64(memoryInfo.RSS)
	return
}
