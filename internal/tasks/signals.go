package tasks

import (
	"os"
	"os/signal"
	"sync"
	"syscall"

	"github.com/langgenius/dify-plugin-daemon/pkg/utils/log"
)

var (
	signalChanInterrupt = make(chan os.Signal, 1)
	signalChanTerminate = make(chan os.Signal, 1)
	signalChanKill      = make(chan os.Signal, 1)

	finalizers = []Finalizer{}

	lock = sync.Mutex{}
)

// Finalizer is a function that will be called before the program exits
// It should return an error if it fails to cleanup
// errors returned by finalizers won't block other finalizers from being called
// and the program will exit with code 1 if any finalizer fails
type Finalizer func() error

func SetupSignalHandler() {
	signal.Notify(signalChanInterrupt, os.Interrupt)
	signal.Notify(signalChanTerminate, syscall.SIGTERM)
	signal.Notify(signalChanKill, os.Interrupt)

	go func() {
		select {
		case <-signalChanInterrupt:
		case <-signalChanTerminate:
		case <-signalChanKill:
		}

		hasError := false
		lock.Lock()
		defer lock.Unlock()
		for _, finalizer := range finalizers {
			err := finalizer()
			if err != nil {
				log.Error("finalizer failed: %s", err.Error())
				hasError = true
			}
		}
		if hasError {
			os.Exit(1)
		} else {
			os.Exit(0)
		}
	}()
}

func RegisterFinalizers(fns ...Finalizer) {
	lock.Lock()
	defer lock.Unlock()
	finalizers = append(finalizers, fns...)
}
