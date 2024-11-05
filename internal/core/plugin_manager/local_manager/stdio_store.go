package local_manager

import (
	"io"
	"sync"

	"github.com/google/uuid"
)

func PutStdioIo(
	plugin_unique_identifier string, writer io.WriteCloser,
	reader io.ReadCloser, err_reader io.ReadCloser,
) *stdioHolder {
	id := uuid.New().String()

	holder := &stdioHolder{
		plugin_unique_identifier: plugin_unique_identifier,
		writer:                   writer,
		reader:                   reader,
		err_reader:               err_reader,
		id:                       id,
		l:                        &sync.Mutex{},

		health_chan_lock: &sync.Mutex{},
		health_chan:      make(chan bool),
	}

	stdio_holder.Store(id, holder)
	return holder
}

func Get(id string) *stdioHolder {
	if v, ok := stdio_holder.Load(id); ok {
		if holder, ok := v.(*stdioHolder); ok {
			return holder
		}
	}

	return nil
}

func RemoveStdio(id string) {
	stdio_holder.Delete(id)
}

func OnStdioEvent(id string, session_id string, listener func([]byte)) {
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

func OnError(id string, session_id string, listener func([]byte)) {
	if v, ok := stdio_holder.Load(id); ok {
		if holder, ok := v.(*stdioHolder); ok {
			holder.l.Lock()
			defer holder.l.Unlock()
			if holder.error_listener == nil {
				holder.error_listener = map[string]func([]byte){}
			}

			holder.error_listener[session_id] = listener
		}
	}
}

func RemoveStdioListener(id string, listener string) {
	if v, ok := stdio_holder.Load(id); ok {
		if holder, ok := v.(*stdioHolder); ok {
			holder.l.Lock()
			defer holder.l.Unlock()
			delete(holder.listener, listener)
			delete(holder.error_listener, listener)
		}
	}
}

func OnGlobalEvent(listener func(string, []byte)) {
	l.Lock()
	defer l.Unlock()
	listeners[uuid.New().String()] = listener
}

func WriteToStdio(id string, data []byte) error {
	if v, ok := stdio_holder.Load(id); ok {
		if holder, ok := v.(*stdioHolder); ok {
			_, err := holder.writer.Write(data)

			return err
		}
	}

	return nil
}
