package testutils

import (
	"errors"
	"fmt"
	"os"
	"os/exec"
	"path"
	"strings"
	"sync"

	_ "embed"

	"github.com/langgenius/dify-plugin-daemon/internal/core/dify_invocation/mock"
	plugin_daemon "github.com/langgenius/dify-plugin-daemon/internal/core/io_tunnel"
	"github.com/langgenius/dify-plugin-daemon/internal/core/io_tunnel/access_types"
	"github.com/langgenius/dify-plugin-daemon/internal/core/local_runtime"
	"github.com/langgenius/dify-plugin-daemon/internal/core/session_manager"
	"github.com/langgenius/dify-plugin-daemon/internal/types/app"
	"github.com/langgenius/dify-plugin-daemon/pkg/entities/plugin_entities"
	"github.com/langgenius/dify-plugin-daemon/pkg/entities/requests"
	"github.com/langgenius/dify-plugin-daemon/pkg/plugin_packager/decoder"
	"github.com/langgenius/dify-plugin-daemon/pkg/utils/stream"
)

// GetRuntime returns a runtime for a plugin
// Please ensure cwd is a valid directory without any file in it
func GetRuntime(pluginZip []byte, cwd string, instanceNums int) (*local_runtime.LocalPluginRuntime, error) {
	decoder, err := decoder.NewZipPluginDecoder(pluginZip)
	if err != nil {
		return nil, errors.Join(err, fmt.Errorf("create plugin decoder error"))
	}

	// get manifest
	manifest, err := decoder.Manifest()
	if err != nil {
		return nil, errors.Join(err, fmt.Errorf("get plugin manifest error"))
	}

	identity := manifest.Identity()
	identity = strings.ReplaceAll(identity, ":", "-")

	checksum, err := decoder.Checksum()
	if err != nil {
		return nil, errors.Join(err, fmt.Errorf("calculate checksum error"))
	}

	// check if the working directory exists, if not, create it, otherwise, launch it directly
	pluginWorkingPath := path.Join(cwd, fmt.Sprintf("%s@%s", identity, checksum))
	if _, err := os.Stat(pluginWorkingPath); err != nil {
		if err := decoder.ExtractTo(pluginWorkingPath); err != nil {
			return nil, errors.Join(err, fmt.Errorf("extract plugin to working directory error"))
		}
	}

	uvPath := os.Getenv("UV_PATH")
	if uvPath == "" {
		if path, err := exec.LookPath("uv"); err == nil {
			uvPath = path
		}
	}

	config := &app.Config{
		PythonInterpreterPath:      os.Getenv("PYTHON_INTERPRETER_PATH"),
		UvPath:                     uvPath,
		PythonEnvInitTimeout:       120,
		PluginWorkingPath:          cwd,
		PluginRuntimeBufferSize:    1024,
		PluginRuntimeMaxBufferSize: 5242880,
	}

	// FIXME: cli test command should give a timeout for launching
	runtime, err := local_runtime.ConstructPluginRuntime(config, decoder)
	if err != nil {
		return nil, errors.Join(err, fmt.Errorf("construct plugin runtime error"))
	}

	// initialize environment
	if err := runtime.InitPythonEnvironment(); err != nil {
		return nil, errors.Join(err, fmt.Errorf("initialize python environment error"))
	}

	errChan := make(chan error, 1)
	launchedChan := make(chan bool, 1)
	once := sync.Once{}

	notifier := local_runtime.PluginRuntimeNotifierTemplate{
		OnInstanceLaunchFailedImpl: func(instance *local_runtime.PluginInstance, err error) {
			once.Do(func() {
				errChan <- err
			})
		},
		OnInstanceReadyImpl: func(instance *local_runtime.PluginInstance) {
			once.Do(func() {
				launchedChan <- true
			})
		},
	}

	// scale up to the expected instance nums
	for i := 0; i < instanceNums; i++ {
		runtime.ScaleUp()
	}

	runtime.AddNotifier(&notifier)
	if err := runtime.Schedule(); err != nil {
		return nil, errors.Join(err, fmt.Errorf("schedule plugin runtime error"))
	}

	// wait for plugin launched
	select {
	case err := <-errChan:
		return nil, errors.Join(err, fmt.Errorf("plugin runtime failed"))
	case <-launchedChan:
	}

	return runtime, nil
}

func ClearTestingPath(cwd string) {
	os.RemoveAll(cwd)
}

type RunOnceRequest interface {
	requests.RequestInvokeLLM | requests.RequestInvokeTextEmbedding | requests.RequestInvokeRerank |
		requests.RequestInvokeTTS | requests.RequestInvokeSpeech2Text | requests.RequestInvokeModeration |
		requests.RequestValidateProviderCredentials | requests.RequestValidateModelCredentials |
		requests.RequestGetTTSModelVoices | requests.RequestGetTextEmbeddingNumTokens |
		requests.RequestGetLLMNumTokens | requests.RequestGetAIModelSchema | requests.RequestInvokeAgentStrategy |
		requests.RequestOAuthGetAuthorizationURL | requests.RequestOAuthGetCredentials |
		requests.RequestInvokeEndpoint |
		map[string]any
}

// RunOnceWithSession sends a request to plugin and returns a stream of responses
// It requires a session to be provided
func RunOnceWithSession[T RunOnceRequest, R any](
	runtime *local_runtime.LocalPluginRuntime,
	session *session_manager.Session,
	request T,
) (*stream.Stream[R], error) {
	// bind the runtime to the session, plugin_daemon.GenericInvokePlugin uses it
	session.BindRuntime(runtime)

	return plugin_daemon.GenericInvokePlugin[T, R](session, &request, 1024)
}

// RunOnce sends a request to plugin and returns a stream of responses
// It automatically generates a session for the request
func RunOnce[T RunOnceRequest, R any](
	runtime *local_runtime.LocalPluginRuntime,
	accessType access_types.PluginAccessType,
	action access_types.PluginAccessAction,
	request T,
) (*stream.Stream[R], error) {
	session := session_manager.NewSession(
		session_manager.NewSessionPayload{
			UserID:                 "test",
			TenantID:               "test",
			PluginUniqueIdentifier: plugin_entities.PluginUniqueIdentifier(""),
			ClusterID:              "test",
			InvokeFrom:             accessType,
			Action:                 action,
			Declaration:            nil,
			BackwardsInvocation:    mock.NewMockedDifyInvocation(),
			IgnoreCache:            true,
		},
	)

	return RunOnceWithSession[T, R](runtime, session, request)
}
