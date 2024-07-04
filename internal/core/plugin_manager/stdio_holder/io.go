package stdio_holder

import (
	"bufio"
	"fmt"
	"io"
	"sync"

	"github.com/google/uuid"
	"github.com/langgenius/dify-plugin-daemon/internal/types/entities/plugin_entities"
	"github.com/langgenius/dify-plugin-daemon/internal/utils/log"
	"github.com/langgenius/dify-plugin-daemon/internal/utils/parser"
)

var (
	stdio_holder sync.Map                        = sync.Map{}
	l            *sync.Mutex                     = &sync.Mutex{}
	listeners    map[string]func(string, []byte) = map[string]func(string, []byte){}
)

type stdioHolder struct {
	id             string
	pluginIdentity string
	writer         io.WriteCloser
	reader         io.ReadCloser
	errReader      io.ReadCloser
	l              *sync.Mutex
	listener       map[string]func([]byte)
	started        bool
	alive          bool
}

func (s *stdioHolder) Stop() {
	s.alive = false
	s.writer.Close()
	s.reader.Close()
	s.errReader.Close()

	stdio_holder.Delete(s.id)
}

func (s *stdioHolder) StartStdout() {
	s.started = true
	s.alive = true

	scanner := bufio.NewScanner(s.reader)
	for s.alive {
		for scanner.Scan() {
			data := scanner.Bytes()
			event, err := parser.UnmarshalJsonBytes[plugin_entities.PluginUniversalEvent](data)
			if err != nil {
				log.Error("unmarshal json failed: %s", err.Error())
				continue
			}

			session_id := event.SessionId

			switch event.Event {
			case plugin_entities.PLUGIN_EVENT_LOG:
				if event.Event == plugin_entities.PLUGIN_EVENT_LOG {
					logEvent, err := parser.UnmarshalJsonBytes[plugin_entities.PluginLogEvent](event.Data)
					if err != nil {
						log.Error("unmarshal json failed: %s", err.Error())
						continue
					}

					log.Info("plugin %s: %s", s.pluginIdentity, logEvent.Message)
				}
			case plugin_entities.PLUGIN_EVENT_RESPONSE:
				for _, listener := range listeners {
					listener(s.id, event.Data)
				}

				for listener_session_id, listener := range s.listener {
					if listener_session_id == session_id {
						listener(event.Data)
					}
				}
			case plugin_entities.PLUGIN_EVENT_ERROR:
				log.Error("plugin %s: %s", s.pluginIdentity, event.Data)
			}
		}
	}
}

/*
 * @return error
 */
func (s *stdioHolder) StartStderr() error {
	s.started = true
	s.alive = true
	defer s.Stop()
	for s.alive {
		buf := make([]byte, 1024)
		n, err := s.errReader.Read(buf)
		if err != nil && err != io.EOF {
			return err
		} else if err != nil {
			return nil
		}

		if n > 0 {
			return fmt.Errorf("stderr: %s", buf[:n])
		}
	}

	return nil
}

func (s *stdioHolder) GetID() string {
	return s.id
}

/*
 * @param plugin_identity: string
 * @param writer: io.WriteCloser
 * @param reader: io.ReadCloser
 * @param errReader: io.ReadCloser
 */
func Put(
	plugin_identity string,
	writer io.WriteCloser,
	reader io.ReadCloser,
	errReader io.ReadCloser,
) *stdioHolder {
	id := uuid.New().String()

	holder := &stdioHolder{
		pluginIdentity: plugin_identity,
		writer:         writer,
		reader:         reader,
		errReader:      errReader,
		id:             id,
		l:              &sync.Mutex{},
	}

	stdio_holder.Store(id, holder)
	return holder
}

/*
 * @param id: string
 */
func Get(id string) *stdioHolder {
	if v, ok := stdio_holder.Load(id); ok {
		if holder, ok := v.(*stdioHolder); ok {
			return holder
		}
	}

	return nil
}

/*
 * @param id: string
 */
func Remove(id string) {
	stdio_holder.Delete(id)
}

/*
 * @param id: string
 * @param session_id: string
 * @param listener: func(data []byte)
 * @return string - listener identity
 */
func OnEvent(id string, session_id string, listener func([]byte)) {
	if v, ok := stdio_holder.Load(id); ok {
		if holder, ok := v.(*stdioHolder); ok {
			holder.l.Lock()
			defer holder.l.Unlock()
			if holder.listener == nil {
				holder.listener = map[string]func([]byte){}
			}

			holder.listener[session_id] = listener
		}
	}
}

/*
 * @param id: string
 * @param listener: string
 */
func RemoveListener(id string, listener string) {
	if v, ok := stdio_holder.Load(id); ok {
		if holder, ok := v.(*stdioHolder); ok {
			holder.l.Lock()
			defer holder.l.Unlock()
			delete(holder.listener, listener)
		}
	}
}

/*
 * @param listener: func(id string, data []byte)
 */
func OnGlobalEvent(listener func(string, []byte)) {
	l.Lock()
	defer l.Unlock()
	listeners[uuid.New().String()] = listener
}

/*
 * @param id: string
 * @param data: []byte
 */
func Write(id string, data []byte) error {
	if v, ok := stdio_holder.Load(id); ok {
		if holder, ok := v.(*stdioHolder); ok {
			_, err := holder.writer.Write(data)

			return err
		}
	}

	return nil
}
