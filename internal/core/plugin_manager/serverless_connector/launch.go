package serverless

import (
	"bytes"
	"sync"
	"time"

	"github.com/langgenius/dify-plugin-daemon/internal/utils/cache"
	"github.com/langgenius/dify-plugin-daemon/internal/utils/stream"
	"github.com/langgenius/dify-plugin-daemon/pkg/entities/plugin_entities"
	"github.com/langgenius/dify-plugin-daemon/pkg/plugin_packager/decoder"
)

var (
	SERVERLESS_LAUNCH_LOCK_PREFIX = "serverless_launch_lock_"

	locks = sync.Map{}
)

func addLock(checksum string) {
	locks.Store(checksum, struct{}{})
}

func removeLock(checksum string) {
	locks.Delete(checksum)
}

func CleanupLocks() error {
	var errs []error
	locks.Range(func(key, value any) bool {
		lockName := (key.(string))
		if err := cache.Unlock(lockName); err != nil {
			errs = append(errs, err)
		}
		return true
	})
	return nil
}

// LaunchPlugin uploads the plugin to specific serverless connector
// return the function url and name
func LaunchPlugin(
	pluginUniqueIdentifier plugin_entities.PluginUniqueIdentifier,
	originPackage []byte,
	decoder decoder.PluginDecoder,
	timeout int, // in seconds
	ignoreIdempotent bool, // if true, never check if the plugin has launched
) (*stream.Stream[LaunchFunctionResponse], error) {
	checksum, err := decoder.Checksum()
	if err != nil {
		return nil, err
	}
	lock := SERVERLESS_LAUNCH_LOCK_PREFIX + checksum
	// check if the plugin has already been initialized
	if err := cache.Lock(
		lock,
		time.Duration(timeout)*time.Second,
		time.Duration(timeout)*time.Second,
	); err != nil {
		return nil, err
	}
	addLock(lock)

	unlock := func(e error) error {
		cache.Unlock(lock)
		removeLock(lock)
		return e
	}

	manifest, err := decoder.Manifest()
	if err != nil {
		return nil, unlock(err)
	}

	if !ignoreIdempotent {
		function, err := FetchFunction(manifest, checksum)
		if err != nil {
			if err != ErrFunctionNotFound {
				return nil, unlock(err)
			}
		} else {
			// found, return directly
			response := stream.NewStream[LaunchFunctionResponse](3)
			response.Write(LaunchFunctionResponse{
				Event:   FunctionUrl,
				Message: function.FunctionURL,
			})
			response.Write(LaunchFunctionResponse{
				Event:   Function,
				Message: function.FunctionName,
			})
			response.Write(LaunchFunctionResponse{
				Event:   Done,
				Message: "",
			})
			response.Close()
			return response, unlock(nil)
		}
	}

	response, err := SetupFunction(pluginUniqueIdentifier, manifest, checksum, bytes.NewReader(originPackage), timeout)
	if err != nil {
		return nil, unlock(err)
	}

	response.BeforeClose(func() { unlock(nil) })
	return response, nil
}
