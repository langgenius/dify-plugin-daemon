package local_manager

import (
	"bufio"
	"errors"
	"fmt"
	"io"
	"sync"
	"time"

	"github.com/langgenius/dify-plugin-daemon/internal/core/plugin_manager/plugin_errors"
	"github.com/langgenius/dify-plugin-daemon/internal/types/entities/plugin_entities"
	"github.com/langgenius/dify-plugin-daemon/internal/utils/log"
)

var (
	stdio_holder sync.Map                        = sync.Map{}
	l            *sync.Mutex                     = &sync.Mutex{}
	listeners    map[string]func(string, []byte) = map[string]func(string, []byte){}
)

type stdioHolder struct {
	id                       string
	plugin_unique_identifier string
	writer                   io.WriteCloser
	reader                   io.ReadCloser
	err_reader               io.ReadCloser
	l                        *sync.Mutex
	listener                 map[string]func([]byte)
	error_listener           map[string]func([]byte)
	started                  bool

	// error message container
	err_message                 string
	last_err_message_updated_at time.Time

	// waiting controller channel to notify the exit signal to the Wait() function
	waiting_controller_chan        chan bool
	waiting_controller_chan_closed bool
	wait_controller_chan_lock      *sync.Mutex

	// the last time the plugin sent a heartbeat
	last_active_at time.Time
}

func (s *stdioHolder) Error() error {
	if time.Since(s.last_err_message_updated_at) < 60*time.Second {
		if s.err_message != "" {
			return errors.New(s.err_message)
		}
	}

	return nil
}

// Stop stops the stdio, of course, it will shutdown the plugin asynchronously
// by closing a channel to notify the `Wait()` function to exit
func (s *stdioHolder) Stop() {
	s.writer.Close()
	s.reader.Close()
	s.err_reader.Close()

	s.wait_controller_chan_lock.Lock()
	if !s.waiting_controller_chan_closed {
		close(s.waiting_controller_chan)
		s.waiting_controller_chan_closed = true
	}
	s.wait_controller_chan_lock.Unlock()

	stdio_holder.Delete(s.id)
}

// StartStdout starts to read the stdout of the plugin
// it will notify the heartbeat function when the plugin is active
// and parse the stdout data to trigger corresponding listeners
func (s *stdioHolder) StartStdout(notify_heartbeat func()) {
	s.started = true
	s.last_active_at = time.Now()
	defer s.Stop()

	scanner := bufio.NewScanner(s.reader)
	for scanner.Scan() {
		data := scanner.Bytes()
		if len(data) == 0 {
			continue
		}

		plugin_entities.ParsePluginUniversalEvent(
			data,
			func(session_id string, data []byte) {
				for _, listener := range listeners {
					listener(s.id, data)
				}
				for listener_session_id, listener := range s.listener {
					if listener_session_id == session_id {
						listener(data)
					}
				}
			},
			func() {
				s.last_active_at = time.Now()
				// notify launched
				notify_heartbeat()
			},
			func(err string) {
				log.Error("plugin %s: %s", s.plugin_unique_identifier, err)
			},
			func(message string) {
				log.Info("plugin %s: %s", s.plugin_unique_identifier, message)
			},
		)
	}
}

// WriteError writes the error message to the stdio holder
// it will keep the last 1024 bytes of the error message
func (s *stdioHolder) WriteError(msg string) {
	const MAX_ERR_MSG_LEN = 1024
	reduce := len(msg) + len(s.err_message) - MAX_ERR_MSG_LEN
	if reduce > 0 {
		if reduce > len(s.err_message) {
			s.err_message = ""
		} else {
			s.err_message = s.err_message[reduce:]
		}
	}

	s.err_message += msg
	s.last_err_message_updated_at = time.Now()
}

// StartStderr starts to read the stderr of the plugin
// it will write the error message to the stdio holder
func (s *stdioHolder) StartStderr() {
	for {
		buf := make([]byte, 1024)
		n, err := s.err_reader.Read(buf)
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
	s.wait_controller_chan_lock.Lock()
	if s.waiting_controller_chan_closed {
		s.wait_controller_chan_lock.Unlock()
		return errors.New("you need to start the health check before waiting")
	}
	s.wait_controller_chan_lock.Unlock()

	ticker := time.NewTicker(5 * time.Second)
	defer ticker.Stop()

	// check status of plugin every 5 seconds
	for {
		s.wait_controller_chan_lock.Lock()
		if s.waiting_controller_chan_closed {
			s.wait_controller_chan_lock.Unlock()
			break
		}
		s.wait_controller_chan_lock.Unlock()
		select {
		case <-ticker.C:
			// check heartbeat
			if time.Since(s.last_active_at) > 60*time.Second {
				return plugin_errors.ErrPluginNotActive
			}
		case <-s.waiting_controller_chan:
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
