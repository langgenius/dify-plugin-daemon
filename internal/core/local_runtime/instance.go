package local_runtime

import (
	"bufio"
	"errors"
	"fmt"
	"io"
	"os/exec"
	"sync"
	"time"

	"github.com/google/uuid"
	"github.com/langgenius/dify-plugin-daemon/internal/types/app"
	"github.com/langgenius/dify-plugin-daemon/pkg/entities/plugin_entities"
)

const (
	MAX_ERR_MSG_LEN = 1024

	MAX_HEARTBEAT_INTERVAL = 120 * time.Second
)

type PluginInstance struct {
	instanceId             string
	pluginUniqueIdentifier string
	cmd                    *exec.Cmd
	inWriter               io.WriteCloser
	outReader              io.ReadCloser
	errReader              io.ReadCloser
	l                      *sync.Mutex
	listener               map[string]func([]byte)

	started  bool // mark the instance as started
	shutdown bool // mark the instance as shutdown

	// app config
	appConfig *app.Config

	// error message container
	errMessage              string
	lastErrMessageUpdatedAt time.Time

	// the last time the plugin sent a heartbeat
	lastActiveAt time.Time

	// notifier
	notifiers    []PluginInstanceNotifier
	notifierLock *sync.Mutex
}

type PluginInstanceConfig struct {
	StdoutBufferSize    int
	StdoutMaxBufferSize int
}

func newPluginInstance(
	pluginUniqueIdentifier string,
	e *exec.Cmd,
	writer io.WriteCloser,
	reader io.ReadCloser,
	errReader io.ReadCloser,
	appConfig *app.Config,
) *PluginInstance {
	instanceId, _ := uuid.NewV7()
	instance := &PluginInstance{
		instanceId:             instanceId.String(),
		pluginUniqueIdentifier: pluginUniqueIdentifier,
		cmd:                    e,
		inWriter:               writer,
		outReader:              reader,
		errReader:              errReader,
		l:                      &sync.Mutex{},
		appConfig:              appConfig,

		notifiers:    []PluginInstanceNotifier{},
		notifierLock: &sync.Mutex{},
	}

	return instance
}

func (s *PluginInstance) setupStdioEventListener(session_id string, listener func([]byte)) {
	s.l.Lock()
	defer s.l.Unlock()
	if s.listener == nil {
		s.listener = map[string]func([]byte){}
	}

	s.listener[session_id] = listener
}

func (s *PluginInstance) removeStdioHandlerListener(session_id string) {
	s.l.Lock()
	defer s.l.Unlock()
	delete(s.listener, session_id)
}

func (s *PluginInstance) Error() error {
	if time.Since(s.lastErrMessageUpdatedAt) < 60*time.Second {
		if s.errMessage != "" {
			return errors.New(s.errMessage)
		}
	}

	return nil
}

// Stop stops the stdio, of course, it will shutdown the plugin asynchronously
func (s *PluginInstance) Stop() {
	s.inWriter.Close()
	s.outReader.Close()
	s.errReader.Close()
}

// StartStdout starts to read the stdout of the plugin
// and parse the stdout data to trigger corresponding listeners
func (s *PluginInstance) StartStdout() {
	defer func() {
		// notify shutdown signal
		s.WalkNotifiers(func(notifier PluginInstanceNotifier) {
			notifier.OnInstanceShutdown(s)
		})
	}()

	once := &sync.Once{}

	scanner := bufio.NewScanner(s.outReader)
	scanner.Buffer(
		make([]byte, s.appConfig.GetLocalRuntimeBufferSize()),
		s.appConfig.GetLocalRuntimeMaxBufferSize(),
	)

	for scanner.Scan() {
		// read data, once s.outReader.Close was called
		// scanner.Scan() breaks immediately
		data := scanner.Bytes()

		if len(data) == 0 {
			continue
		}

		// handle stdout
		s.handleStdout(data, once)

		// notify stdout notifiers
		s.WalkNotifiers(func(notifier PluginInstanceNotifier) {
			notifier.OnInstanceStdout(s, data)
		})
	}

	if err := scanner.Err(); err != nil {
		s.WalkNotifiers(func(notifier PluginInstanceNotifier) {
			notifier.OnInstanceErrorLog(
				s,
				fmt.Errorf(
					"plugin %s has an error on stdout: %s",
					s.pluginUniqueIdentifier,
					err,
				),
			)
		})
	}

	// once reader of stdout is closed, kill subprocess
	if err := s.cmd.Process.Kill(); err != nil {
		s.WalkNotifiers(func(notifier PluginInstanceNotifier) {
			notifier.OnInstanceErrorLog(s, fmt.Errorf("failed to kill subprocess: %s", err.Error()))
		})
	}
}

// handles stdout data and notify corresponding listeners
func (s *PluginInstance) handleStdout(data []byte, once *sync.Once) {
	plugin_entities.ParsePluginUniversalEvent(
		data,
		"",
		func(sessionId string, data []byte) {
			// FIX: avoid deadlock to plugin invoke
			s.l.Lock()
			listener := s.listener[sessionId]
			s.l.Unlock()
			if listener != nil {
				listener(data)
			}
		},
		func() {
			// heartbeat
			s.WalkNotifiers(func(notifier PluginInstanceNotifier) {
				notifier.OnInstanceHeartbeat(s)
			})
			// only first heartbeat will trigger this
			once.Do(func() {
				s.WalkNotifiers(func(notifier PluginInstanceNotifier) {
					notifier.OnInstanceReady(s)
				})
			})
		},
		func(err string) {
			// error log
			s.WalkNotifiers(func(notifier PluginInstanceNotifier) {
				notifier.OnInstanceErrorLog(s, errors.New(err))
			})
		},
		func(message string) {
			// plain text log
			s.WalkNotifiers(func(notifier PluginInstanceNotifier) {
				notifier.OnInstanceLog(s, message)
			})
		},
	)
}

// WriteError writes the error message to the stdio holder
// it will keep the last 1024 bytes of the error message
func (s *PluginInstance) WriteError(msg string) {
	if len(msg) > MAX_ERR_MSG_LEN {
		msg = msg[:MAX_ERR_MSG_LEN]
	}

	reduce := len(msg) + len(s.errMessage) - MAX_ERR_MSG_LEN
	if reduce > 0 {
		if reduce > len(s.errMessage) {
			s.errMessage = ""
		} else {
			s.errMessage = s.errMessage[reduce:]
		}
	}

	s.errMessage += msg
	s.lastErrMessageUpdatedAt = time.Now()
}

// StartStderr starts to read the stderr of the plugin
// it will write the error message to the stdio holder
func (s *PluginInstance) StartStderr() {
	for {
		buf := make([]byte, 1024)
		n, err := s.errReader.Read(buf)
		if err != nil && err != io.EOF {
			break
		} else if err != nil {
			s.WriteError(fmt.Sprintf("%s\n", buf[:n]))
			break
		}

		if n > 0 {
			s.WriteError(fmt.Sprintf("%s\n", buf[:n]))
		}
	}
}

// Monitor monitors the plugin instance
// it will return an error if the plugin is not active
// you can also call `Stop()` to stop the monitoring process
func (s *PluginInstance) Monitor() error {
	ticker := time.NewTicker(30 * time.Second)
	defer ticker.Stop()

	// check status of plugin every 5 seconds
	for range ticker.C {
		// check heartbeat
		if time.Since(s.lastActiveAt) > MAX_HEARTBEAT_INTERVAL {
			s.WalkNotifiers(func(notifier PluginInstanceNotifier) {
				// notify handlers
				notifier.OnInstanceLaunchFailed(
					s,
					fmt.Errorf(
						"plugin %s is not active for %f seconds, it may be dead, captured error logs: %v",
						s.pluginUniqueIdentifier,
						time.Since(s.lastActiveAt).Seconds(),
						s.Error(),
					),
				)
			})
			// dead instance detected, kill it
			s.Stop()
			return ErrRuntimeNotActive
		}
		if time.Since(s.lastActiveAt) > MAX_HEARTBEAT_INTERVAL/2 {
			// notify handlers
			s.WalkNotifiers(func(notifier PluginInstanceNotifier) {
				notifier.OnInstanceWarningLog(
					s,
					fmt.Sprintf(
						"plugin %s is not active for %f seconds, it may be dead",
						s.pluginUniqueIdentifier,
						time.Since(s.lastActiveAt).Seconds(),
					),
				)
			})
		}
	}

	return nil
}

func (s *PluginInstance) Write(data []byte) error {
	// write bytes into instance's stdin
	_, err := s.inWriter.Write(data)
	return err
}

// GracefulStop stops the instance gracefully
// wait for at most maxWaitTime to shutdown, forcefully kill it if timeout reached
func (s *PluginInstance) GracefulStop(maxWaitTime time.Duration) {
	timeout := time.NewTimer(maxWaitTime)
	defer timeout.Stop()
	ticker := time.NewTicker(5 * time.Second)
	defer ticker.Stop()

	for {
		s.l.Lock()
		listeners := len(s.listener)
		s.l.Unlock()

		if listeners == 0 {
			break
		}

		select {
		case <-timeout.C:
			// timeout reached, forcefully kill the instance
			s.Stop()
			return
		case <-ticker.C:
			// do nothing
		}
	}

	// all listeners are closed, stop the instance
	s.Stop()
}
