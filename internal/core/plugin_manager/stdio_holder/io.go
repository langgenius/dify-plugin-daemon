package stdio_holder

import (
	"fmt"
	"io"
	"sync"

	"github.com/google/uuid"
)

var (
	stdio_holder sync.Map                        = sync.Map{}
	l            *sync.Mutex                     = &sync.Mutex{}
	listeners    map[string]func(string, []byte) = map[string]func(string, []byte){}
)

type stdioHolder struct {
	id        string
	writer    io.WriteCloser
	reader    io.ReadCloser
	errReader io.ReadCloser
	l         *sync.Mutex
	listener  map[string]func([]byte)
	started   bool
	alive     bool
}

func (s *stdioHolder) Stop() {
	s.alive = false
	s.writer.Close()
	s.reader.Close()
	s.errReader.Close()
}

func (s *stdioHolder) StartStdout() {
	s.started = true
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
}

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

/*
 * @param writer: io.WriteCloser
 * @param reader: io.ReadCloser
 */
func PutStdio(writer io.WriteCloser, reader io.ReadCloser, errReader io.ReadCloser) *stdioHolder {
	id := uuid.New().String()

	holder := &stdioHolder{
		writer:    writer,
		reader:    reader,
		errReader: errReader,
		id:        id,
		l:         &sync.Mutex{},
	}

	stdio_holder.Store(id, holder)
	return holder
}

/*
 * @param id: string
 */
func GetStdio(id string) *stdioHolder {
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
func RemoveStdio(id string) {
	stdio_holder.Delete(id)
}

/*
 * @param id: string
 * @param listener: func(data []byte)
 * @return string - listener identity
 */
func OnStdioEvent(id string, listener func([]byte)) string {
	if v, ok := stdio_holder.Load(id); ok {
		if holder, ok := v.(*stdioHolder); ok {
			holder.l.Lock()
			defer holder.l.Unlock()
			if holder.listener == nil {
				holder.listener = map[string]func([]byte){}
			}

			identity := uuid.New().String()
			holder.listener[identity] = listener
			return identity
		}
	}

	return ""
}

/*
 * @param id: string
 * @param listener: string
 */
func RemoveStdioListener(id string, listener string) {
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
func OnStdioEventGlobal(listener func(string, []byte)) {
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
