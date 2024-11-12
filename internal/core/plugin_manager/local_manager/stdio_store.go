package local_manager

import (
	"io"
	"sync"

	"github.com/google/uuid"
)

func registerStdioHandler(
	pluginUniqueIdentifier string, writer io.WriteCloser,
	reader io.ReadCloser, err_reader io.ReadCloser,
) *stdioHolder {
	id := uuid.New().String()

	holder := &stdioHolder{
		pluginUniqueIdentifier: pluginUniqueIdentifier,
		writer:                 writer,
		reader:                 reader,
		errReader:              err_reader,
		id:                     id,
		l:                      &sync.Mutex{},

		waitControllerChanLock: &sync.Mutex{},
		waitingControllerChan:  make(chan bool),
	}

	stdio_holder.Store(id, holder)
	return holder
}

func getStdioHandler(id string) *stdioHolder {
	if v, ok := stdio_holder.Load(id); ok {
		if holder, ok := v.(*stdioHolder); ok {
			return holder
		}
	}

	return nil
}

func removeStdioHandler(id string) {
	stdio_holder.Delete(id)
}

func setupStdioEventListener(id string, session_id string, listener func([]byte)) {
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
			if holder.errorListener == nil {
				holder.errorListener = map[string]func([]byte){}
			}

			holder.errorListener[session_id] = listener
		}
	}
}

func removeStdioHandlerListener(id string, listener string) {
	if v, ok := stdio_holder.Load(id); ok {
		if holder, ok := v.(*stdioHolder); ok {
			holder.l.Lock()
			defer holder.l.Unlock()
			delete(holder.listener, listener)
			delete(holder.errorListener, listener)
		}
	}
}

func OnGlobalEvent(listener func(string, []byte)) {
	l.Lock()
	defer l.Unlock()
	listeners[uuid.New().String()] = listener
}

func writeToStdioHandler(id string, data []byte) error {
	if v, ok := stdio_holder.Load(id); ok {
		if holder, ok := v.(*stdioHolder); ok {
			_, err := holder.writer.Write(data)

			return err
		}
	}

	return nil
}
