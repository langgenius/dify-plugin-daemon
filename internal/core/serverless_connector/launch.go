package serverless

import (
	"bytes"
	"context"
	"errors"
	"sync"
	"time"

	"github.com/langgenius/dify-plugin-daemon/pkg/entities/plugin_entities"
	"github.com/langgenius/dify-plugin-daemon/pkg/plugin_packager/decoder"
	"github.com/langgenius/dify-plugin-daemon/pkg/utils/cache"
	"github.com/langgenius/dify-plugin-daemon/pkg/utils/log"
	"github.com/langgenius/dify-plugin-daemon/pkg/utils/stream"
)

var (
	SERVERLESS_LAUNCH_LOCK_PREFIX = "serverless_launch_lock_"
)

// LaunchPlugin uploads the plugin to specific serverless connector
// return the function url and name
func LaunchPlugin(
	ctx context.Context,
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

	// check if the plugin has already been initialized
	lock, err := cache.AcquireOwnedLock(
		SERVERLESS_LAUNCH_LOCK_PREFIX+checksum,
		time.Duration(timeout)*time.Second,
		time.Duration(timeout)*time.Second,
	)
	if err != nil {
		log.Warn(
			"failed to acquire serverless launch lock; continuing with idempotent launch",
			"checksum", checksum,
			"error", err.Error(),
		)
	}

	var stopRenew context.CancelFunc
	var renewResult <-chan error
	if lock != nil {
		var renewCtx context.Context
		renewCtx, stopRenew = context.WithCancel(context.Background())
		renewResult = lock.KeepAlive(renewCtx, time.Duration(timeout)*time.Second)
	}

	var releaseOnce sync.Once
	unlock := func(e error) error {
		releaseOnce.Do(func() {
			if lock == nil {
				return
			}
			stopRenew()
			if renewErr := <-renewResult; renewErr != nil {
				log.Warn(
					"lost serverless launch lock; runtime launch remains authoritative",
					"checksum", checksum,
					"error", renewErr.Error(),
				)
			}
			if unlockErr := lock.Unlock(); unlockErr != nil && !errors.Is(unlockErr, cache.ErrLockNotOwned) {
				log.Warn("failed to release serverless launch lock", "checksum", checksum, "error", unlockErr.Error())
			}
		})
		return e
	}

	manifest, err := decoder.Manifest()
	if err != nil {
		return nil, unlock(err)
	}

	if !ignoreIdempotent {
		function, err := FetchFunction(ctx, manifest, checksum)
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

	response, err := SetupFunction(ctx, pluginUniqueIdentifier, manifest, checksum, bytes.NewReader(originPackage), timeout)
	if err != nil {
		return nil, unlock(err)
	}

	response.BeforeClose(func() { unlock(nil) })
	return response, nil
}
