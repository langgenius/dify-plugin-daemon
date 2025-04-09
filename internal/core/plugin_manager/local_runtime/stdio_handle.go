package local_runtime

import (
	"bufio"
	"errors"
	"fmt"
	"io"
	"sync"
	"time"

	"github.com/langgenius/dify-plugin-daemon/internal/core/plugin_manager/plugin_errors"
	"github.com/langgenius/dify-plugin-daemon/internal/utils/log"
	"github.com/langgenius/dify-plugin-daemon/pkg/entities/plugin_entities"
)

var (
	stdio_holder sync.Map                        = sync.Map{}
	l            *sync.Mutex                     = &sync.Mutex{}
	listeners    map[string]func(string, []byte) = map[string]func(string, []byte){}
)

type stdioHolder struct {
	id                     string
	pluginUniqueIdentifier string
	writer                 io.WriteCloser
	reader                 io.ReadCloser
	errReader              io.ReadCloser
	l                      *sync.Mutex
	listener               map[string]func([]byte)
	errorListener          map[string]func([]byte)
	started                bool

	// error message container
	errMessage              string
	lastErrMessageUpdatedAt time.Time

	// waiting controller channel to notify the exit signal to the Wait() function
	waitingControllerChan       chan bool
	waitingControllerChanClosed bool
	waitControllerChanLock      *sync.Mutex

	// the last time the plugin sent a heartbeat
	lastActiveAt time.Time
}

func (s *stdioHolder) Error() error {
	if time.Since(s.lastErrMessageUpdatedAt) < 60*time.Second {
		if s.errMessage != "" {
			return errors.New(s.errMessage)
		}
	}

	return nil
}

// Stop stops the stdio, of course, it will shutdown the plugin asynchronously
// by closing a channel to notify the `Wait()` function to exit
func (s *stdioHolder) Stop() {
	s.writer.Close()
	s.reader.Close()
	s.errReader.Close()

	s.waitControllerChanLock.Lock()
	if !s.waitingControllerChanClosed {
		close(s.waitingControllerChan)
		s.waitingControllerChanClosed = true
	}
	s.waitControllerChanLock.Unlock()

	stdio_holder.Delete(s.id)
}

// StartStdout starts to read the stdout of the plugin
// it will notify the heartbeat function when the plugin is active
// and parse the stdout data to trigger corresponding listeners
func (s *stdioHolder) StartStdout(notify_heartbeat func()) {
	s.started = true
	s.lastActiveAt = time.Now()
	defer s.Stop()

	scanner := bufio.NewScanner(s.reader)

	// TODO: set a reasonable buffer size or use a reader, this is a temporary solution
	scanner.Buffer(make([]byte, 1024), 5*1024*1024)

	for scanner.Scan() {
		data := scanner.Bytes()

		if len(data) == 0 {
			continue
		}

		// update the last active time on each time the plugin sends data
		s.lastActiveAt = time.Now()

		plugin_entities.ParsePluginUniversalEvent(
			data,
			"",
			func(session_id string, data []byte) {
				for _, listener := range listeners {
					listener(s.id, data)
				}
				// FIX: avoid deadlock to plugin invoke
				s.l.Lock()
				tasks := []func(){}
				for listener_session_id, listener := range s.listener {
					// copy the listener to avoid reference issue
					listener := listener
					if listener_session_id == session_id {
						tasks = append(tasks, func() {
							listener(data)
						})
					}
				}
				s.l.Unlock()
				for _, t := range tasks {
					t()
				}
			},
			func() {
				// notify launched
				notify_heartbeat()
			},
			func(err string) {
				log.Error("plugin %s: %s", s.pluginUniqueIdentifier, err)
			},
			func(message string) {
				log.Info("plugin %s: %s", s.pluginUniqueIdentifier, message)
			},
		)
	}

	if err := scanner.Err(); err != nil {
		log.Error("plugin %s has an error on stdout: %s", s.pluginUniqueIdentifier, err)
	}
}

// WriteError writes the error message to the stdio holder
// it will keep the last 1024 bytes of the error message
func (s *stdioHolder) WriteError(msg string) {
	const MAX_ERR_MSG_LEN = 1024
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
func (s *stdioHolder) StartStderr() {
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

// Wait waits for the plugin to exit
// it will return an error if the plugin is not active
// you can also call `Stop()` to stop the waiting process
func (s *stdioHolder) Wait() error {
	s.waitControllerChanLock.Lock()
	if s.waitingControllerChanClosed {
		s.waitControllerChanLock.Unlock()
		return errors.New("you need to start the health check before waiting")
	}
	s.waitControllerChanLock.Unlock()

	ticker := time.NewTicker(5 * time.Second)
	defer ticker.Stop()

	// check status of plugin every 5 seconds
	for {
		s.waitControllerChanLock.Lock()
		if s.waitingControllerChanClosed {
			s.waitControllerChanLock.Unlock()
			break
		}
		s.waitControllerChanLock.Unlock()
		select {
		case <-ticker.C:
			// check heartbeat
			if time.Since(s.lastActiveAt) > 120*time.Second {
				log.Error(
					"plugin %s is not active for 120 seconds, it may be dead, killing and restarting it",
					s.pluginUniqueIdentifier,
				)
				return plugin_errors.ErrPluginNotActive
			}
			if time.Since(s.lastActiveAt) > 60*time.Second {
				log.Warn(
					"plugin %s is not active for %f seconds, it may be dead",
					s.pluginUniqueIdentifier,
					time.Since(s.lastActiveAt).Seconds(),
				)
			}
		case <-s.waitingControllerChan:
			// closed
			return s.Error()
		}
	}

	return nil
}

// GetID returns the id of the stdio holder
func (s *stdioHolder) GetID() string {
	return s.id
}
