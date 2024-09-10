package serverless

import (
	"fmt"
	"os"
	"time"

	"github.com/langgenius/dify-plugin-daemon/internal/core/plugin_manager/aws_manager"
	"github.com/langgenius/dify-plugin-daemon/internal/utils/cache"
)

var (
	AWS_LAUNCH_LOCK_PREFIX = "aws_launch_lock_"
)

// UploadPlugin uploads the plugin to the AWS Lambda
func UploadPlugin(r *aws_manager.AWSPluginRuntime) error {
	r.Log("Starting to initialize environment")
	// check if the plugin has already been initialized, at most 300s
	if err := cache.Lock(AWS_LAUNCH_LOCK_PREFIX+r.Checksum(), 300*time.Second, 300*time.Second); err != nil {
		return err
	}
	defer cache.Unlock(AWS_LAUNCH_LOCK_PREFIX + r.Checksum())
	r.Log("Started to initialize environment")

	identity, err := r.Identity()
	if err != nil {
		return err
	}
	function, err := FetchLambda(identity.String(), r.Checksum())
	if err != nil {
		if err != ErrNoLambdaFunction {
			return err
		}
	} else {
		// found, return directly
		r.LambdaURL = function.FunctionURL
		r.LambdaName = function.FunctionName
		r.Log(fmt.Sprintf("Found existing lambda function: %s", r.LambdaName))
		return nil
	}

	// create it if not found
	r.Log("Creating new lambda function")

	// create lambda function
	packager := NewPackager(r, r.Decoder)
	context, err := packager.Pack()
	if err != nil {
		return err
	}
	defer os.Remove(context.Name())
	defer context.Close()

	response, err := LaunchLambda(identity.String(), r.Checksum(), context)
	if err != nil {
		return err
	}

	for response.Next() {
		response, err := response.Read()
		if err != nil {
			return err
		}

		switch response.Event {
		case Error:
			return fmt.Errorf("error: %s", response.Message)
		case LambdaUrl:
			r.LambdaURL = response.Message
		case Lambda:
			r.LambdaName = response.Message
		case Info:
			r.Log(fmt.Sprintf("installing: %s", response.Message))
		}
	}

	return nil
}
