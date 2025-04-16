package plugin_manager

import (
	"errors"
	"fmt"
	"os"
	"path"
	"strings"
	"sync"
	"testing"
	"time"

	_ "embed"

	"github.com/langgenius/dify-plugin-daemon/internal/core/dify_invocation/tester"
	"github.com/langgenius/dify-plugin-daemon/internal/core/plugin_daemon/access_types"
	"github.com/langgenius/dify-plugin-daemon/internal/core/plugin_daemon/generic_invoke"
	"github.com/langgenius/dify-plugin-daemon/internal/core/plugin_manager/basic_runtime"
	"github.com/langgenius/dify-plugin-daemon/internal/core/plugin_manager/local_runtime"
	"github.com/langgenius/dify-plugin-daemon/internal/core/session_manager"
	"github.com/langgenius/dify-plugin-daemon/internal/utils/log"
	"github.com/langgenius/dify-plugin-daemon/internal/utils/parser"
	"github.com/langgenius/dify-plugin-daemon/internal/utils/routine"
	"github.com/langgenius/dify-plugin-daemon/internal/utils/stream"
	"github.com/langgenius/dify-plugin-daemon/pkg/entities/model_entities"
	"github.com/langgenius/dify-plugin-daemon/pkg/entities/plugin_entities"
	"github.com/langgenius/dify-plugin-daemon/pkg/entities/requests"
	"github.com/langgenius/dify-plugin-daemon/pkg/plugin_packager/decoder"
)

const (
	_testingPath = "./benchmark_testing"
)

func getRuntime(pluginZip []byte) (*local_runtime.LocalPluginRuntime, error) {
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
	pluginWorkingPath := path.Join(_testingPath, fmt.Sprintf("%s@%s", identity, checksum))
	if _, err := os.Stat(pluginWorkingPath); err != nil {
		if err := decoder.ExtractTo(pluginWorkingPath); err != nil {
			return nil, errors.Join(err, fmt.Errorf("extract plugin to working directory error"))
		}
	}

	localPluginRuntime := local_runtime.NewLocalPluginRuntime(local_runtime.LocalPluginRuntimeConfig{
		PythonInterpreterPath: os.Getenv("PYTHON_INTERPRETER_PATH"),
		UvPath:                os.Getenv("UV_PATH"),
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

		fullDuplexLifecycle(localPluginRuntime, launchedChan, errChan)
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

func runOnce(
	b *testing.B,
	runtime *local_runtime.LocalPluginRuntime,
	request requests.RequestInvokeLLM,
) {
	session := session_manager.NewSession(
		session_manager.NewSessionPayload{
			UserID:                 "test",
			TenantID:               "test",
			PluginUniqueIdentifier: plugin_entities.PluginUniqueIdentifier(""),
			ClusterID:              "test",
			InvokeFrom:             access_types.PLUGIN_ACCESS_TYPE_MODEL,
			Action:                 access_types.PLUGIN_ACCESS_ACTION_INVOKE_LLM,
			Declaration:            nil,
			BackwardsInvocation:    tester.NewMockedDifyInvocation(),
			IgnoreCache:            true,
		},
	)
	session.BindRuntime(runtime)

	response := stream.NewStream[model_entities.LLMResultChunk](1024)

	listener := runtime.Listen(session.ID)
	listener.Listen(func(chunk plugin_entities.SessionMessage) {
		switch chunk.Type {
		case plugin_entities.SESSION_MESSAGE_TYPE_STREAM:
			chunk, err := parser.UnmarshalJsonBytes[model_entities.LLMResultChunk](chunk.Data)
			if err != nil {
				response.WriteError(errors.New(parser.MarshalJson(map[string]string{
					"error_type": "unmarshal_error",
					"message":    fmt.Sprintf("unmarshal json failed: %s", err.Error()),
				})))
				response.Close()
				return
			} else {
				response.Write(chunk)
			}
		case plugin_entities.SESSION_MESSAGE_TYPE_END:
			response.Close()
		case plugin_entities.SESSION_MESSAGE_TYPE_ERROR:
			e, err := parser.UnmarshalJsonBytes[plugin_entities.ErrorResponse](chunk.Data)
			if err != nil {
				break
			}
			response.WriteError(errors.New(e.Error()))
			response.Close()
		default:
			response.WriteError(errors.New(parser.MarshalJson(map[string]string{
				"error_type": "unknown_stream_message_type",
				"message":    "unknown stream message type: " + string(chunk.Type),
			})))
			response.Close()
		}
	})

	// close the listener if stream outside is closed due to close of connection
	response.OnClose(func() {
		listener.Close()
	})

	session.Write(
		session_manager.PLUGIN_IN_STREAM_EVENT_REQUEST,
		session.Action,
		generic_invoke.GetInvokePluginMap(
			session,
			request,
		),
	)

	for response.Next() {
		response.Read()
	}
}

//go:embed testdata/openai.difypkg
var openaiPluginZip []byte

func BenchmarkLocalOpenAILLMInvocation(b *testing.B) {
	log.SetShowLog(false)

	routine.InitPool(10000)

	const concurrency = 100
	r := b.N

	wg := sync.WaitGroup{}
	sem := make(chan struct{}, concurrency)

	// get runtime
	runtime, err := getRuntime(openaiPluginZip)
	if err != nil {
		b.Fatal(err)
	}
	defer func() {
		//runtime.Stop()
		// os.RemoveAll(runtime.PluginRuntime.State.WorkingPath)
		// os.RemoveAll(_testingPath)
	}()

	// launch fake openai server
	port, _ := StartFakeOpenAIServer()

	b.ResetTimer()
	for i := 0; i < r; i++ {
		sem <- struct{}{}
		wg.Add(1)
		go func() {
			defer func() {
				<-sem
				wg.Done()
			}()
			runOnce(b, runtime, requests.RequestInvokeLLM{
				BaseRequestInvokeModel: requests.BaseRequestInvokeModel{
					Provider: "openai",
					Model:    "gpt-3.5-turbo",
				},
				Credentials: requests.Credentials{
					Credentials: map[string]any{
						"openai_api_key":  "test",
						"openai_api_base": fmt.Sprintf("http://localhost:%d", port),
					},
				},
				InvokeLLMSchema: requests.InvokeLLMSchema{
					ModelParameters: map[string]any{
						"temperature": 0.5,
					},
					PromptMessages: []model_entities.PromptMessage{
						{
							Role:    "user",
							Content: "Hello, world!",
						},
					},
					Tools:  []model_entities.PromptMessageTool{},
					Stop:   []string{},
					Stream: true,
				},
				ModelType: model_entities.MODEL_TYPE_LLM,
			})
		}()
	}

	wg.Wait()
}
