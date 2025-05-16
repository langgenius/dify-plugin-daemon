package test_utils

import (
	"errors"
	"fmt"
	"os"
	"os/exec"
	"path"
	"strings"
	"time"

	_ "embed"

	"github.com/langgenius/dify-plugin-daemon/internal/core/dify_invocation/tester"
	"github.com/langgenius/dify-plugin-daemon/internal/core/plugin_daemon"
	"github.com/langgenius/dify-plugin-daemon/internal/core/plugin_daemon/access_types"
	"github.com/langgenius/dify-plugin-daemon/internal/core/plugin_manager/basic_runtime"
	"github.com/langgenius/dify-plugin-daemon/internal/core/plugin_manager/lifecycle"
	"github.com/langgenius/dify-plugin-daemon/internal/core/plugin_manager/local_runtime"
	"github.com/langgenius/dify-plugin-daemon/internal/core/session_manager"
	"github.com/langgenius/dify-plugin-daemon/internal/utils/log"
	"github.com/langgenius/dify-plugin-daemon/internal/utils/routine"
	"github.com/langgenius/dify-plugin-daemon/internal/utils/stream"
	"github.com/langgenius/dify-plugin-daemon/pkg/entities/plugin_entities"
	"github.com/langgenius/dify-plugin-daemon/pkg/entities/requests"
	"github.com/langgenius/dify-plugin-daemon/pkg/plugin_packager/decoder"
)

// GetRuntime returns a runtime for a plugin
// Please ensure cwd is a valid directory without any file in it
func GetRuntime(pluginZip []byte, cwd string) (*local_runtime.LocalPluginRuntime, error) {
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

	localPluginRuntime := local_runtime.NewLocalPluginRuntime(local_runtime.LocalPluginRuntimeConfig{
		PythonInterpreterPath: os.Getenv("PYTHON_INTERPRETER_PATH"),
		UvPath:                uvPath,
		PythonEnvInitTimeout:  120,
	})

	localPluginRuntime.PluginRuntime = plugin_entities.PluginRuntime{
		Config: manifest,
		State: plugin_entities.PluginRuntimeState{
			Status:      plugin_entities.PLUGIN_RUNTIME_STATUS_PENDING,
			Restarts:    0,
			ActiveAt:    nil,
			Verified:    manifest.Verified,
			WorkingPath: pluginWorkingPath,
		},
	}
	localPluginRuntime.BasicChecksum = basic_runtime.BasicChecksum{
		WorkingPath: pluginWorkingPath,
		Decoder:     decoder,
	}

	launchedChan := make(chan bool)
	errChan := make(chan error)

	// local plugin
	routine.Submit(map[string]string{
		"module":   "plugin_manager",
		"function": "LaunchLocal",
	}, func() {
		defer func() {
			if r := recover(); r != nil {
				log.Error("plugin runtime panic: %v", r)
			}
		}()

		// add max launching lock to prevent too many plugins launching at the same time
		routine.Submit(map[string]string{
			"module":   "plugin_manager",
			"function": "LaunchLocal",
		}, func() {
			// wait for plugin launched
			<-launchedChan
		})

		lifecycle.FullDuplex(localPluginRuntime, launchedChan, errChan)
	})

	// wait for plugin launched
	select {
	case err := <-errChan:
		return nil, err
	case <-launchedChan:
	}

	// wait 1s for stdio to be ready
	time.Sleep(1 * time.Second)

	return localPluginRuntime, nil
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
			BackwardsInvocation:    tester.NewMockedDifyInvocation(),
			IgnoreCache:            true,
		},
	)

	return RunOnceWithSession[T, R](runtime, session, request)
}
