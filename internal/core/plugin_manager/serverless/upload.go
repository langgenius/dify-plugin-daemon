package serverless

import (
	"os"
	"time"

	"github.com/langgenius/dify-plugin-daemon/internal/core/plugin_packager/checksum"
	"github.com/langgenius/dify-plugin-daemon/internal/core/plugin_packager/decoder"
	"github.com/langgenius/dify-plugin-daemon/internal/utils/cache"
	"github.com/langgenius/dify-plugin-daemon/internal/utils/stream"
)

var (
	AWS_LAUNCH_LOCK_PREFIX = "aws_launch_lock_"
)

// UploadPlugin uploads the plugin to the AWS Lambda
// return the lambda url and name
func UploadPlugin(decoder decoder.PluginDecoder) (*stream.Stream[LaunchAWSLambdaFunctionResponse], error) {
	checksum, err := checksum.CalculateChecksum(decoder)
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

	identity := manifest.Identity()
	function, err := FetchLambda(identity, checksum)
	if err != nil {
		if err != ErrNoLambdaFunction {
			return nil, err
		}
	} else {
		// found, return directly
		response := stream.NewStreamResponse[LaunchAWSLambdaFunctionResponse](2)
		response.Write(LaunchAWSLambdaFunctionResponse{
			Event:   LambdaUrl,
			Message: function.FunctionURL,
		})
		response.Write(LaunchAWSLambdaFunctionResponse{
			Event:   Lambda,
			Message: function.FunctionName,
		})
		return response, nil
	}

	// create lambda function
	packager := NewPackager(decoder)
	context, err := packager.Pack()
	if err != nil {
		return nil, err
	}
	defer os.Remove(context.Name())
	defer context.Close()

	response, err := LaunchLambda(identity, checksum, context)
	if err != nil {
		return nil, err
	}

	return response, nil
}
