package plugin_manager

import (
	"io"
	"sync"

	"github.com/google/uuid"
)

var (
	stdio_holder sync.Map               = sync.Map{}
	listeners    []func(string, []byte) = []func(string, []byte){}
)

type stdioHolder struct {
	id       string
	writer   io.WriteCloser
	reader   io.ReadCloser
	listener []func(data []byte)
	started  bool
	alive    bool
}

func (s *stdioHolder) Stop() {
	s.alive = false
	s.writer.Close()
	s.reader.Close()
}

func (s *stdioHolder) Start() {
	s.started = true

	go func() {
		s.alive = true
		for s.alive {
			buf := make([]byte, 1024)
			n, err := s.reader.Read(buf)
			if err != nil {
				s.Stop()
				break
			}

			for _, listener := range listeners {
				listener(s.id, buf[:n])
			}

			for _, listener := range s.listener {
				listener(buf[:n])
			}
		}
	}()
}

func PutStdio(writer io.WriteCloser, reader io.ReadCloser) string {
	id := uuid.New().String()

	holder := &stdioHolder{
		writer: writer,
		reader: reader,
		id:     id,
	}

	stdio_holder.Store(id, holder)

	holder.Start()

	return id
}

/*
 * @param id: string
 */
func RemoveStdio(id string) {
	stdio_holder.Delete(id)
}

/*
 * @param listener: func(data []byte)
 */
func OnStdioEvent(id string, listener func([]byte)) {
	if v, ok := stdio_holder.Load(id); ok {
		if holder, ok := v.(*stdioHolder); ok {
			holder.listener = append(holder.listener, listener)
		}
	}
}

/*
 * @param listener: func(id string, data []byte)
 */
func OnStdioEventGlobal(listener func(string, []byte)) {
	listeners = append(listeners, listener)
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
