// Package local_runtime manages local plugin runtimes and provides a lightweight autoscaling
// scheduler. The autoscaling algorithm itself is intentionally stubbed; only the scheduling
// framework is implemented here.
package local_runtime

import (
	"errors"
	"fmt"
	"os"
	"os/exec"
	"sync"
	"time"

	"github.com/langgenius/dify-plugin-daemon/internal/utils/log"
	"github.com/langgenius/dify-plugin-daemon/internal/utils/routine"
	"github.com/langgenius/dify-plugin-daemon/pkg/entities/constants"
	"github.com/langgenius/dify-plugin-daemon/pkg/entities/plugin_entities"
)

/*
================================================================================
 Autoscaling scheduler — design notes
 --------------------------------------------------------------------------------
 * All plugin instances are treated identically. During `launchStageInit`, any
   failure to start **any** instance is considered fatal and is propagated back
   to `StartPlugin` so that callers know startup failed.
 * After the runtime enters `launchStageVerified`, instance exits are considered
   unexpected crashes. The scheduler will automatically start new instances to
   bring the replica count back to the desired target.
 * The replica‑target decision is encapsulated in `getDesiredScale()` so that
   higher‑level code can plug in metrics‑driven logic.
 * `waitChan` semantics from the original implementation remain untouched:
       – it is created at the beginning of `StartPlugin`
       – closed in `gc()` when the runtime tears down
   The scheduler uses its own `scalingStop` channel to terminate the ticker
   goroutine, keeping `waitChan` for external observers only.
================================================================================
*/

//----------------------------------------------------------------------------//
// Additional struct fields (add these to LocalPluginRuntime)
//----------------------------------------------------------------------------//
//   scalingTicker *time.Ticker  // periodic reconciliation ticker
//   scalingStop   chan struct{} // closed when the scheduler should exit
//   scaleInterval time.Duration // reconciliation period (defaults to 5 s)
//----------------------------------------------------------------------------//

//------------------------------------------------------------------------------
// Helper functions
//------------------------------------------------------------------------------

// gc performs garbage collection for the LocalPluginRuntime
func (r *LocalPluginRuntime) gc() {
	if r.waitChan != nil {
		close(r.waitChan)
		r.waitChan = nil
	}
}

// Type returns the runtime type of the plugin
func (r *LocalPluginRuntime) Type() plugin_entities.PluginRuntimeType {
	return plugin_entities.PLUGIN_RUNTIME_TYPE_LOCAL
}

// getCmd prepares the exec.Cmd for the plugin according to its language
func (r *LocalPluginRuntime) getCmd() (*exec.Cmd, error) {
	if r.Config.Meta.Runner.Language == constants.Python {
		cmd := exec.Command(r.pythonInterpreterPath, "-m", r.Config.Meta.Runner.Entrypoint)
		cmd.Dir = r.State.WorkingPath
		cmd.Env = cmd.Environ()
		if r.HttpsProxy != "" {
			cmd.Env = append(cmd.Env, fmt.Sprintf("HTTPS_PROXY=%s", r.HttpsProxy))
		}
		if r.HttpProxy != "" {
			cmd.Env = append(cmd.Env, fmt.Sprintf("HTTP_PROXY=%s", r.HttpProxy))
		}
		if r.NoProxy != "" {
			cmd.Env = append(cmd.Env, fmt.Sprintf("NO_PROXY=%s", r.NoProxy))
		}
		return cmd, nil
	}
	return nil, fmt.Errorf("unsupported language: %s", r.Config.Meta.Runner.Language)
}

//------------------------------------------------------------------------------
// Start / Stop lifecycle (single entry point)
//------------------------------------------------------------------------------

// StartPlugin launches the autoscaling scheduler and blocks until the runtime
// is stopped or until startup fails while still in launchStageInit.
func (r *LocalPluginRuntime) StartPlugin() error {
	r.stage = launchStageInit
	scalingTicker := time.NewTicker(time.Second * 5)
	scalingStop := make(chan bool)
	defer func() {
		close(scalingStop)
		scalingTicker.Stop()
	}()

	defer log.Info("plugin %s stopped", r.Config.Identity())
	defer func() {
		r.waitChanLock.Lock()
		for _, c := range r.waitStoppedChan {
			select {
			case c <- true:
			default:
			}
		}
		r.waitChanLock.Unlock()
	}()

	if r.isNotFirstStart {
		r.SetRestarting()
	} else {
		r.SetLaunching()
		r.isNotFirstStart = true
	}

	// Reset waitChan for external observers (original behaviour)
	r.waitChan = make(chan bool)

	// fatalErrChan propagates startup failures while still in Init stage
	fatalErrChan := make(chan error, 1)

	// Do an immediate reconcile so at least one instance is running
	if err := r.reconcileOnce(fatalErrChan); err != nil {
		r.gc()
		return err
	}

	routine.Submit(map[string]string{
		"module":   "plugin_manager",
		"type":     "local",
		"function": "schedulerLoop",
	}, func() {
		r.schedulerLoop(scalingTicker, scalingStop, fatalErrChan)
	})

	healthCheckTicker := time.NewTicker(time.Second * 10)
	defer healthCheckTicker.Stop()

	// Wait for a fatal startup error or for external Stop()
	for {
		select {
		case err := <-fatalErrChan:
			r.Stop()
			r.gc()
			return err
		case <-scalingStop: // gc() triggered → runtime shutting down
			return nil
		case <-healthCheckTicker.C:
			if r.Stopped() {
				return nil
			}
		}
	}
}

// Stop terminates the runtime and all managed plugin instances.
func (r *LocalPluginRuntime) Stop() {
	// Inherit behaviour from PluginRuntime (sets stopped flag etc.)
	r.PluginRuntime.Stop()

	// Stop every active stdioHolder
	r.stdioHolderLock.Lock()
	for _, h := range r.stdioHolders {
		h.Stop()
	}
	r.stdioHolderLock.Unlock()
}

//------------------------------------------------------------------------------
// Scheduler
//------------------------------------------------------------------------------

// schedulerLoop runs a ticker‑driven reconciliation until scalingStop is closed.
func (r *LocalPluginRuntime) schedulerLoop(scalingTicker *time.Ticker, scalingStop chan bool, fatalErrChan chan<- error) {
	for {
		select {
		case <-scalingTicker.C:
			if r.Stopped() {
				return
			}
			_ = r.reconcileOnce(fatalErrChan) // errors after Init stage are logged only
		case <-scalingStop:
			return
		}
	}
}

// reconcileOnce ensures that the current replica count matches getDesiredScale().
// During launchStageInit any instance‑startup error is fatal and returned, otherwise
// the error is only logged.
func (r *LocalPluginRuntime) reconcileOnce(fatalErrChan chan<- error) error {
	desired := r.getDesiredScale()

	r.stdioHolderLock.Lock()
	current := len(r.stdioHolders)
	r.stdioHolderLock.Unlock()

	switch {
	case desired > current:
		if r.autoScale && desired > 1 {
			log.Info("plugin %s auto scaling from %d to %d", r.Config.Identity(), current, desired)
		}
		return r.scaleUp(desired-current, fatalErrChan)
	case desired < current:
		r.scaleDown(current - desired)
	}
	return nil
}

// getDesiredScale decides the target replica count.
func (r *LocalPluginRuntime) getDesiredScale() int {
	if r.stage == launchStageInit {
		return 1
	}

	if !r.autoScale {
		return 1
	}

	// keep the average cpu usage of the instances below 50%
	totalCpuUsage := 0
	r.stdioHolderLock.Lock()
	for _, h := range r.stdioHolders {
		totalCpuUsage += int(h.cpuUsagePercentSum / _SAMPLES)
	}
	r.stdioHolderLock.Unlock()

	// calculate how many instances are needed to keep the average cpu usage below 50%
	targetInstances := totalCpuUsage / 50

	if targetInstances > r.maxInstances {
		targetInstances = r.maxInstances
	} else if targetInstances < r.minInstances {
		targetInstances = r.minInstances
	}

	return targetInstances
}

// scaleUp starts `n` additional plugin instances.
func (r *LocalPluginRuntime) scaleUp(n int, fatalErrChan chan<- error) error {
	if r.scaling {
		return nil
	}

	r.scaling = true
	defer func() {
		r.scaling = false
	}()

	var firstErr error
	for i := 0; i < n; i++ {
		launched := make(chan bool)
		routine.Submit(map[string]string{
			"module":      "plugin_manager",
			"type":        "local",
			"function":    "AutoScale",
			"innerMethod": "scaleUp.launchOneInstance",
		}, func() {
			holder, err := r.launchOneInstance(launched)
			if err != nil {
				// propagate the first fatal error only when still in Init stage
				if r.stage == launchStageInit && firstErr == nil {
					firstErr = err
					select {
					case fatalErrChan <- err:
					default:
					}
				} else {
					log.Error("plugin %s instance exited unexpectedly: %s", r.Config.Identity(), err.Error())
				}

				if holder != nil {
					err := holder.Error()
					if err != nil {
						log.Error(
							"plugin %s instance exited unexpectedly with stderr messages: %s",
							r.Config.Identity(),
							err.Error(),
						)
					}
				}
			}
			if holder != nil {
				r.removeHolder(holder)
			}
		})
		<-launched // wait until the instance has executed cmd.Start()
	}
	return firstErr
}

// scaleDown terminates `n` surplus instances (FIFO order).
func (r *LocalPluginRuntime) scaleDown(n int) {
	r.stdioHolderLock.Lock()
	if n > len(r.stdioHolders) {
		n = len(r.stdioHolders)
	}
	toStop := r.stdioHolders[:n]
	r.stdioHolders = r.stdioHolders[n:]
	r.stdioHolderLock.Unlock()

	for _, h := range toStop {
		h.Stop()
	}
}

// removeHolder deletes a finished instance from the holders slice.
func (r *LocalPluginRuntime) removeHolder(target *pluginInstance) {
	r.stdioHolderLock.Lock()
	defer r.stdioHolderLock.Unlock()
	for i, h := range r.stdioHolders {
		if h == target {
			r.stdioHolders = append(r.stdioHolders[:i], r.stdioHolders[i+1:]...)
			return
		}
	}
}

//------------------------------------------------------------------------------
// launchOneInstance — unchanged except for comment tidy‑up
//------------------------------------------------------------------------------

func (r *LocalPluginRuntime) launchOneInstance(launched chan bool) (*pluginInstance, error) {
	once := sync.Once{}
	closeOnce := func() { once.Do(func() { close(launched) }) }
	defer closeOnce()

	cmd, err := r.getCmd()
	if err != nil {
		return nil, err
	}

	cmd.Dir = r.State.WorkingPath
	cmd.Env = append(cmd.Environ(), "INSTALL_METHOD=local", "PATH="+os.Getenv("PATH"))

	stdin, err := cmd.StdinPipe()
	if err != nil {
		return nil, fmt.Errorf("get stdin pipe failed: %s", err.Error())
	}
	defer stdin.Close()
	stdout, err := cmd.StdoutPipe()
	if err != nil {
		return nil, fmt.Errorf("get stdout pipe failed: %s", err.Error())
	}
	defer stdout.Close()
	stderr, err := cmd.StderrPipe()
	if err != nil {
		return nil, fmt.Errorf("get stderr pipe failed: %s", err.Error())
	}
	defer stderr.Close()

	if err := cmd.Start(); err != nil {
		return nil, fmt.Errorf("start plugin failed: %s", err.Error())
	}
	defer cmd.Process.Kill()

	wg := sync.WaitGroup{}
	wg.Add(2)

	holder := newPluginInstance(r.Config.Identity(), cmd.Process.Pid, stdin, stdout, stderr)

	// add holder to r.stdioHolders
	r.stdioHolderLock.Lock()
	r.stdioHolders = append(r.stdioHolders, holder)
	r.stdioHolderLock.Unlock()

	// instance has started successfully; notify caller
	closeOnce()

	routine.Submit(map[string]string{
		"module":   "plugin_manager",
		"type":     "local",
		"function": "StartStdout",
	}, func() {
		defer wg.Done()
		holder.StartStdout(func() {
			r.stage = launchStageVerifiedWorking
		})
	})
	routine.Submit(map[string]string{
		"module":   "plugin_manager",
		"type":     "local",
		"function": "StartStderr",
	}, func() {
		defer wg.Done()
		holder.StartStderr()
	})
	routine.Submit(map[string]string{
		"module":   "plugin_manager",
		"type":     "local",
		"function": "startUsageMonitor",
	}, func() {
		holder.startUsageMonitor()
	})

	err = holder.Wait()
	if err != nil {
		return holder, errors.Join(err, holder.Error())
	}
	wg.Wait()

	return holder, nil
}

//------------------------------------------------------------------------------
// Wait helpers (original logic retained)
//------------------------------------------------------------------------------

func (r *LocalPluginRuntime) Wait() (<-chan bool, error) {
	if r.waitChan == nil {
		return nil, errors.New("plugin not started")
	}
	return r.waitChan, nil
}

func (r *LocalPluginRuntime) WaitStarted() <-chan bool {
	c := make(chan bool)
	r.waitChanLock.Lock()
	r.waitStartedChan = append(r.waitStartedChan, c)
	r.waitChanLock.Unlock()
	return c
}

func (r *LocalPluginRuntime) WaitStopped() <-chan bool {
	c := make(chan bool)
	r.waitChanLock.Lock()
	r.waitStoppedChan = append(r.waitStoppedChan, c)
	r.waitChanLock.Unlock()
	return c
}
