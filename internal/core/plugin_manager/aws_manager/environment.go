package aws_manager

import (
	"fmt"
	"os"
	"time"

	"github.com/langgenius/dify-plugin-daemon/internal/utils/cache"
)

var (
	AWS_LAUNCH_LOCK_PREFIX = "aws_launch_lock_"
)

func (r *AWSPluginRuntime) InitEnvironment() error {
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
	function, err := fetchLambda(identity, r.Checksum())
	if err != nil {
		if err != ErrNoLambdaFunction {
			return err
		}
	} else {
		// found, return directly
		r.lambda_url = function.FunctionURL
		r.lambda_name = function.FunctionName
		r.Log(fmt.Sprintf("Found existing lambda function: %s", r.lambda_name))
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

	response, err := launchLambda(identity, r.Checksum(), context)
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
			r.lambda_url = response.Message
		case Lambda:
			r.lambda_name = response.Message
		case Info:
			r.Log(fmt.Sprintf("installing: %s", response.Message))
		}
	}

	return nil
}

func (r *AWSPluginRuntime) Checksum() string {
	return ""
}
