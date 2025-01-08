package serverless

import (
	"bytes"
	"time"

	"github.com/langgenius/dify-plugin-daemon/internal/core/plugin_packager/decoder"
	"github.com/langgenius/dify-plugin-daemon/internal/utils/cache"
	"github.com/langgenius/dify-plugin-daemon/internal/utils/stream"
)

var (
	AWS_LAUNCH_LOCK_PREFIX = "aws_launch_lock_"
)

// LaunchPlugin uploads the plugin to specific serverless connector
// return the function url and name
func LaunchPlugin(originPackage []byte, decoder decoder.PluginDecoder) (*stream.Stream[LaunchFunctionResponse], error) {
	checksum, err := decoder.Checksum()
	if err != nil {
		return nil, err
	}

	// check if the plugin has already been initialized, at most 300s
	if err := cache.Lock(AWS_LAUNCH_LOCK_PREFIX+checksum, 300*time.Second, 300*time.Second); err != nil {
		return nil, err
	}
	defer cache.Unlock(AWS_LAUNCH_LOCK_PREFIX + checksum)

	manifest, err := decoder.Manifest()
	if err != nil {
		return nil, err
	}

	function, err := FetchFunction(manifest, checksum)
	if err != nil {
		if err != ErrFunctionNotFound {
			return nil, err
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
		return response, nil
	}

	response, err := SetupFunction(manifest, checksum, bytes.NewReader(originPackage))
	if err != nil {
		return nil, err
	}

	return response, nil
}
