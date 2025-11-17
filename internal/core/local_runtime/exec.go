package local_runtime

import "errors"

var (
	// Runtime could not find a proper instance to execute the request
	// Usually, it means no instance was ready to accept requests
	ErrNoProperInstance = errors.New("no proper instance")

	// Runtime could not find the sessionId in the mapping
	ErrSessionNotFound = errors.New("session not found")

	// Runtime already started
	ErrRuntimeAlreadyStarted = errors.New("runtime already started")

	// Runtime already stopped
	ErrRuntimeAlreadyStopped = errors.New("runtime already stopped")

	// Runtime is not active
	ErrRuntimeNotActive = errors.New("runtime is not active")
)
