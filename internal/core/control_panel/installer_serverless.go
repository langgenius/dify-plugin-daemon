package controlpanel

import (
	"context"
	"errors"
	"strings"
	"time"

	serverless "github.com/langgenius/dify-plugin-daemon/internal/core/serverless_connector"
	"github.com/langgenius/dify-plugin-daemon/pkg/entities/plugin_entities"
	"github.com/langgenius/dify-plugin-daemon/pkg/utils/stream"
	"golang.org/x/sync/singleflight"
)

// Runtime preparation is shared within this daemon; tenant commits remain independent.
var serverlessInstalls singleflight.Group

func (c *ControlPanel) InstallToServerless(
	ctx context.Context,
	pluginUniqueIdentifier plugin_entities.PluginUniqueIdentifier,
) (
	*stream.Stream[serverless.LaunchFunctionResponse], error,
) {
	if err := ctx.Err(); err != nil {
		return nil, err
	}
	result := serverlessInstalls.DoChan(pluginUniqueIdentifier.String(), func() (any, error) {
		// One caller timing out must not cancel a launch still needed by others.
		// Bound the shared operation independently, including runtime preparation and launch.
		timeout := time.Duration(c.config.PluginInstallTimeout) * time.Minute
		if timeout <= 0 {
			timeout = 15 * time.Minute
		}
		sharedCtx, cancel := context.WithTimeout(context.WithoutCancel(ctx), timeout)
		defer cancel()
		return c.ensureServerless(sharedCtx, pluginUniqueIdentifier)
	})
	select {
	case <-ctx.Done():
		return nil, ctx.Err()
	case r := <-result:
		if err := ctx.Err(); err != nil {
			return nil, err
		}
		if r.Err != nil {
			return nil, r.Err
		}
		function := r.Val.(*serverless.ServerlessFunction)
		// Each caller owns its stream; sharing a consumable stream loses events.
		response := stream.NewStream[serverless.LaunchFunctionResponse](3)
		response.Write(serverless.LaunchFunctionResponse{Event: serverless.FunctionUrl, Message: function.FunctionURL})
		response.Write(serverless.LaunchFunctionResponse{Event: serverless.Function, Message: function.FunctionName})
		response.Write(serverless.LaunchFunctionResponse{Event: serverless.Done})
		response.Close()
		return response, nil
	}
}

func (c *ControlPanel) ensureServerless(ctx context.Context, identifier plugin_entities.PluginUniqueIdentifier) (*serverless.ServerlessFunction, error) {
	// Avoid loading and decoding a package when its runtime already exists.
	if _, err := plugin_entities.NewPluginUniqueIdentifier(identifier.String()); err != nil {
		return nil, err
	}
	var manifest plugin_entities.PluginDeclaration
	manifest.Author = identifier.Author()
	manifest.Name = strings.TrimPrefix(identifier.PluginID(), identifier.Author()+"/")
	manifest.Version = identifier.Version()
	function, err := serverless.FetchFunction(ctx, manifest, identifier.Checksum())
	if err == nil {
		if function.FunctionName == "" || function.FunctionURL == "" {
			return nil, errors.New("serverless runtime is incomplete")
		}
		return function, nil
	}
	if !errors.Is(err, serverless.ErrFunctionNotFound) {
		return nil, err
	}
	if err := ctx.Err(); err != nil {
		return nil, err
	}
	packageDecoder, packageFile, err := c.buildPluginDecoder(identifier)
	if err != nil {
		return nil, err
	}
	response, err := serverless.LaunchPlugin(ctx, identifier, packageFile, packageDecoder,
		c.config.DifyPluginServerlessConnectorLaunchTimeout, false)

	if err != nil {
		return nil, err
	}
	stop := context.AfterFunc(ctx, func() { response.CloseWithError(ctx.Err()) })
	defer stop()
	defer response.Close()
	var launchedFunction serverless.ServerlessFunction
	var launchErr error
	var done bool
	err = response.Process(func(event serverless.LaunchFunctionResponse) {
		switch event.Event {
		case serverless.Function:
			launchedFunction.FunctionName = event.Message
		case serverless.FunctionUrl:
			launchedFunction.FunctionURL = event.Message
		case serverless.Error:
			launchErr = errors.New(event.Message)
		case serverless.Done:
			done = true
		}
	})
	if err := ctx.Err(); err != nil {
		return nil, err
	}
	if err != nil {
		return nil, err
	}
	if launchErr != nil {
		return nil, launchErr
	}
	if !done || launchedFunction.FunctionName == "" || launchedFunction.FunctionURL == "" {
		return nil, errors.New("serverless launch ended without a complete runtime")
	}
	return &launchedFunction, nil
}

func (c *ControlPanel) ReinstallToServerless(
	ctx context.Context,
	pluginUniqueIdentifier plugin_entities.PluginUniqueIdentifier,
) (
	*stream.Stream[serverless.LaunchFunctionResponse], error,
) {
	decoder, packageFile, err := c.buildPluginDecoder(pluginUniqueIdentifier)
	if err != nil {
		return nil, err
	}

	// check valid manifest
	_, err = decoder.Manifest()
	if err != nil {
		return nil, err
	}

	response, err := serverless.LaunchPlugin(
		ctx,
		pluginUniqueIdentifier,
		packageFile,
		decoder,
		c.config.DifyPluginServerlessConnectorLaunchTimeout,
		true, // ignoreIdempotent, true means always reinstall
	)
	if err != nil {
		return nil, err
	}

	return response, nil
}
