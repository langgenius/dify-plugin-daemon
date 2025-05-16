package run

import (
	"bufio"
	"errors"
	"fmt"
	"os"
	"os/signal"
	"syscall"

	"github.com/google/uuid"
	"github.com/langgenius/dify-plugin-daemon/internal/core/dify_invocation/tester"
	"github.com/langgenius/dify-plugin-daemon/internal/core/plugin_manager/local_runtime"
	"github.com/langgenius/dify-plugin-daemon/internal/core/plugin_manager/test_utils"
	"github.com/langgenius/dify-plugin-daemon/internal/core/session_manager"
	"github.com/langgenius/dify-plugin-daemon/internal/utils/log"
	"github.com/langgenius/dify-plugin-daemon/internal/utils/parser"
	"github.com/langgenius/dify-plugin-daemon/internal/utils/routine"
	"github.com/langgenius/dify-plugin-daemon/internal/utils/stream"
	"github.com/langgenius/dify-plugin-daemon/pkg/entities/plugin_entities"
	"github.com/langgenius/dify-plugin-daemon/pkg/plugin_packager/decoder"
)

func logResponse(response GenericResponse, responseFormat string, client client) {
	var responseBytes []byte
	if responseFormat == "json" {
		responseBytes = parser.MarshalJsonBytes(response)
	} else if responseFormat == "text" {
		responseBytes = []byte(fmt.Sprintf("[%s] %s\n", response.Type, response.Response))
	}

	// add a newline to the response
	responseBytes = append(responseBytes, '\n')

	if _, err := client.writer.Write(responseBytes); err != nil {
		systemLog(GenericResponse{
			Type:     GENERIC_RESPONSE_TYPE_ERROR,
			Response: map[string]any{"error": err.Error()},
		}, responseFormat)
	}
}

func systemLog(response GenericResponse, responseFormat string) {
	if responseFormat == "json" {
		responseBytes := parser.MarshalJsonBytes(response)
		fmt.Println(string(responseBytes))
	} else if responseFormat == "text" {
		switch response.Type {
		case GENERIC_RESPONSE_TYPE_INFO:
			logger.Output(3, log.LOG_LEVEL_DEBUG_COLOR+"[INFO]"+response.Response["info"].(string)+log.LOG_LEVEL_COLOR_END)
		case GENERIC_RESPONSE_TYPE_ERROR:
			logger.Output(3, log.LOG_LEVEL_ERROR_COLOR+"[ERROR]"+response.Response["error"].(string)+log.LOG_LEVEL_COLOR_END)
		}
	}
}

func handleClient(
	client client,
	declaration *plugin_entities.PluginDeclaration,
	runtime *local_runtime.LocalPluginRuntime,
	responseFormat string,
) {
	// handle request from client
	scanner := bufio.NewScanner(client.reader)
	scanner.Buffer(make([]byte, 1024*1024), 15*1024*1024)

	// generate a random user id, tenant id and cluster id
	userID := uuid.New().String()
	tenantID := uuid.New().String()
	clusterID := uuid.New().String()

	// runtime.Identity() has already been checked in RunPlugin
	pluginUniqueIdentifier, _ := runtime.Identity()

	// mocked invocation
	mockedInvocation := tester.NewMockedDifyInvocation()

	logResponse(GenericResponse{
		Type:     GENERIC_RESPONSE_TYPE_PLUGIN_READY,
		Response: map[string]any{"info": "plugin loaded"},
	}, responseFormat, client)

	for scanner.Scan() {
		payload := scanner.Bytes()
		logResponse(GenericResponse{
			Type:     GENERIC_RESPONSE_TYPE_INFO,
			Response: map[string]any{"info": "received request"},
		}, responseFormat, client)

		invokePayload, err := parser.UnmarshalJsonBytes[InvokePluginPayload](payload)
		if err != nil {
			logResponse(GenericResponse{
				InvokeID: invokePayload.InvokeID,
				Type:     GENERIC_RESPONSE_TYPE_ERROR,
				Response: map[string]any{"error": err.Error()},
			}, responseFormat, client)
			continue
		}

		if invokePayload.Action == "" || invokePayload.Type == "" {
			logResponse(GenericResponse{
				InvokeID: invokePayload.InvokeID,
				Type:     GENERIC_RESPONSE_TYPE_ERROR,
				Response: map[string]any{"error": "action and type are required"},
			}, responseFormat, client)
			continue
		}

		session := session_manager.NewSession(
			session_manager.NewSessionPayload{
				UserID:                 userID,
				TenantID:               tenantID,
				PluginUniqueIdentifier: pluginUniqueIdentifier,
				ClusterID:              clusterID,
				InvokeFrom:             invokePayload.Type,
				Action:                 invokePayload.Action,
				Declaration:            declaration,
				BackwardsInvocation:    mockedInvocation,
				IgnoreCache:            true,
			},
		)

		stream, err := test_utils.RunOnceWithSession[map[string]any, map[string]any](
			runtime,
			session,
			invokePayload.Request,
		)

		if err != nil {
			logResponse(GenericResponse{
				InvokeID: invokePayload.InvokeID,
				Type:     GENERIC_RESPONSE_TYPE_ERROR,
				Response: map[string]any{"error": err.Error()},
			}, responseFormat, client)
			continue
		}

		routine.Submit(nil, func() {
			for stream.Next() {
				response, err := stream.Read()
				if err != nil {
					logResponse(GenericResponse{
						InvokeID: invokePayload.InvokeID,
						Type:     GENERIC_RESPONSE_TYPE_ERROR,
						Response: map[string]any{"error": err.Error()},
					}, responseFormat, client)
					continue
				}

				logResponse(GenericResponse{
					InvokeID: invokePayload.InvokeID,
					Type:     GENERIC_RESPONSE_TYPE_PLUGIN_RESPONSE,
					Response: response,
				}, responseFormat, client)
			}

			logResponse(GenericResponse{
				InvokeID: invokePayload.InvokeID,
				Type:     GENERIC_RESPONSE_TYPE_PLUGIN_INVOKE_END,
				Response: map[string]any{"info": "plugin invoke end"},
			}, responseFormat, client)
		})
	}

}

func RunPlugin(payload RunPluginPayload) {
	if err := runPlugin(payload); err != nil {
		systemLog(GenericResponse{
			Type:     GENERIC_RESPONSE_TYPE_ERROR,
			Response: map[string]any{"error": err.Error()},
		}, payload.ResponseFormat)
		os.Exit(1)
	}
}

func setupSignalHandler(dir string) {
	signalChanInterrupt := make(chan os.Signal, 1)
	signalChanTerminate := make(chan os.Signal, 1)
	signalChanKill := make(chan os.Signal, 1)
	signal.Notify(signalChanInterrupt, os.Interrupt)
	signal.Notify(signalChanTerminate, syscall.SIGTERM)
	signal.Notify(signalChanKill, os.Interrupt)

	go func() {
		select {
		case <-signalChanInterrupt:
		case <-signalChanTerminate:
		case <-signalChanKill:
		}
		os.RemoveAll(dir)
		os.Exit(0)
	}()
}

func runPlugin(payload RunPluginPayload) error {
	// disable logs
	log.SetLogVisibility(payload.EnableLogs)

	// init routine pool
	routine.InitPool(10000)

	// generate a random cwd
	tempDir := os.TempDir()
	dir, err := os.MkdirTemp(tempDir, "plugin-run-*")
	if err != nil {
		return errors.Join(err, fmt.Errorf("create temp directory error"))
	}
	defer test_utils.ClearTestingPath(dir)

	// remove the temp directory when the program shuts down
	setupSignalHandler(dir)

	// try decode the plugin zip file
	pluginFile, err := os.ReadFile(payload.PluginPath)
	if err != nil {
		return errors.Join(err, fmt.Errorf("read plugin file error"))
	}
	zipDecoder, err := decoder.NewZipPluginDecoder(pluginFile)
	if err != nil {
		return errors.Join(err, fmt.Errorf("decode plugin file error"))
	}

	// get the declaration of the plugin
	declaration, err := zipDecoder.Manifest()
	if err != nil {
		return errors.Join(err, fmt.Errorf("get declaration error"))
	}

	systemLog(GenericResponse{
		Type:     GENERIC_RESPONSE_TYPE_INFO,
		Response: map[string]any{"info": "loading plugin"},
	}, payload.ResponseFormat)

	// launch the plugin locally and returns a local runtime
	runtime, err := test_utils.GetRuntime(pluginFile, dir)
	if err != nil {
		return err
	}

	// check the identity of the plugin
	_, err = runtime.Identity()
	if err != nil {
		return err
	}

	var stream *stream.Stream[client]
	switch payload.RunMode {
	case RUN_MODE_STDIO:
		// create a stream of clients that are connected to the plugin through stdin and stdout
		// NOTE: for stdio, there will only be one client and the stream will never close
		stream = createStdioServer()
	case RUN_MODE_TCP:
		// create a stream of clients that are connected to the plugin through a TCP connection
		// NOTE: for tcp, there will be multiple clients and the stream will close when the client is closed
		stream, err = createTCPServer(&payload)
		if err != nil {
			return err
		}

		systemLog(GenericResponse{
			Type: GENERIC_RESPONSE_TYPE_INFO,
			Response: map[string]any{
				"info": fmt.Sprintf("plugin is running on %s:%d", payload.TcpServerHost, payload.TcpServerPort),
				"host": payload.TcpServerHost,
				"port": payload.TcpServerPort,
			},
		}, payload.ResponseFormat)
	default:
		return fmt.Errorf("invalid run mode: %s", payload.RunMode)
	}

	// start a routine to handle the client stream
	for stream.Next() {
		client, err := stream.Read()
		if err != nil {
			systemLog(GenericResponse{
				Type:     GENERIC_RESPONSE_TYPE_ERROR,
				Response: map[string]any{"error": err.Error()},
			}, payload.ResponseFormat)
			continue
		}

		routine.Submit(nil, func() {
			handleClient(client, &declaration, runtime, payload.ResponseFormat)
		})
	}

	return nil
}
