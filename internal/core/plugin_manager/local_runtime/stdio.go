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

const (
	MAX_ERR_MSG_LEN = 1024

	MAX_HEARTBEAT_INTERVAL = 120 * time.Second
)

type stdioHolder struct {
	pluginUniqueIdentifier string
	writer                 io.WriteCloser
	reader                 io.ReadCloser
	errReader              io.ReadCloser
	l                      *sync.Mutex
	listener               map[string]func([]byte)
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

	stdoutBufferSize    int
	stdoutMaxBufferSize int
}

type StdioHolderConfig struct {
	StdoutBufferSize    int
	StdoutMaxBufferSize int
}

func newStdioHolder(
	pluginUniqueIdentifier string, writer io.WriteCloser,
	reader io.ReadCloser, err_reader io.ReadCloser,
	config *StdioHolderConfig,
) *stdioHolder {
	if config == nil {
		config = &StdioHolderConfig{}
	}

	if config.StdoutBufferSize <= 0 {
		config.StdoutBufferSize = 1024
	}
	if config.StdoutMaxBufferSize <= 0 {
		config.StdoutMaxBufferSize = 5 * 1024 * 1024
	}

	holder := &stdioHolder{
		pluginUniqueIdentifier: pluginUniqueIdentifier,
		writer:                 writer,
		reader:                 reader,
		errReader:              err_reader,
		l:                      &sync.Mutex{},

		stdoutBufferSize:       config.StdoutBufferSize,
		stdoutMaxBufferSize:    config.StdoutMaxBufferSize,
		waitControllerChanLock: &sync.Mutex{},
		waitingControllerChan:  make(chan bool),
	}

	return holder
}

func (s *stdioHolder) setupStdioEventListener(session_id string, listener func([]byte)) {
	s.l.Lock()
	defer s.l.Unlock()
	if s.listener == nil {
		s.listener = map[string]func([]byte){}
	}

	s.listener[session_id] = listener
}

func (s *stdioHolder) removeStdioHandlerListener(session_id string) {
	s.l.Lock()
	defer s.l.Unlock()
	delete(s.listener, session_id)
}

func (s *stdioHolder) write(data []byte) error {
	_, err := s.writer.Write(data)
	return err
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
	scanner.Buffer(make([]byte, s.stdoutBufferSize), s.stdoutMaxBufferSize)

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
				// FIX: avoid deadlock to plugin invoke
				s.l.Lock()
				listener := s.listener[session_id]
				s.l.Unlock()
				if listener != nil {
					listener(data)
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
			if time.Since(s.lastActiveAt) > MAX_HEARTBEAT_INTERVAL {
				log.Error(
					"plugin %s is not active for %f seconds, it may be dead, killing and restarting it",
					s.pluginUniqueIdentifier,
					time.Since(s.lastActiveAt).Seconds(),
				)
				return plugin_errors.ErrPluginNotActive
			}
			if time.Since(s.lastActiveAt) > MAX_HEARTBEAT_INTERVAL/2 {
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
