package run

import (
	"bufio"
	"errors"
	"fmt"
	"os"

	"github.com/google/uuid"
	"github.com/langgenius/dify-plugin-daemon/internal/core/dify_invocation/tester"
	"github.com/langgenius/dify-plugin-daemon/internal/core/plugin_daemon/access_types"
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

func logResponse(response GenericResponse, client client) {
	responseBytes := parser.MarshalJsonBytes(response)
	if _, err := client.writer.Write(responseBytes); err != nil {
		log.Error("write response to client error", "error", err)
	}
}

func logResponseToStdout(response GenericResponse) {
	responseBytes := parser.MarshalJsonBytes(response)
	fmt.Println(string(responseBytes))
}

func handleClient(client client, declaration *plugin_entities.PluginDeclaration, runtime *local_runtime.LocalPluginRuntime) {
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

	for scanner.Scan() {
		payload := scanner.Bytes()
		invokePayload, err := parser.UnmarshalJsonBytes2Map(payload)
		if err != nil {
			logResponse(GenericResponse{
				Type:     GENERIC_RESPONSE_TYPE_ERROR,
				Response: map[string]any{"error": err.Error()},
			}, client)
			continue
		}

		session := session_manager.NewSession(
			session_manager.NewSessionPayload{
				UserID:                 userID,
				TenantID:               tenantID,
				PluginUniqueIdentifier: pluginUniqueIdentifier,
				ClusterID:              clusterID,
				InvokeFrom:             access_types.PLUGIN_ACCESS_TYPE_MODEL,
				Action:                 access_types.PLUGIN_ACCESS_ACTION_INVOKE_LLM,
				Declaration:            declaration,
				BackwardsInvocation:    mockedInvocation,
				IgnoreCache:            true,
			},
		)

		stream, err := test_utils.RunOnceWithSession[map[string]any, map[string]any](runtime, session, invokePayload)
		if err != nil {
			logResponse(GenericResponse{
				Type:     GENERIC_RESPONSE_TYPE_ERROR,
				Response: map[string]any{"error": err.Error()},
			}, client)
			continue
		}

		for stream.Next() {
			response, err := stream.Read()
			if err != nil {
				logResponse(GenericResponse{
					Type:     GENERIC_RESPONSE_TYPE_ERROR,
					Response: map[string]any{"error": err.Error()},
				}, client)
				continue
			}

			logResponse(GenericResponse{
				Type:     GENERIC_RESPONSE_TYPE_PLUGIN_RESPONSE,
				Response: response,
			}, client)
		}
	}

}

func RunPlugin(payload RunPluginPayload) {
	if err := runPlugin(payload); err != nil {
		logResponseToStdout(GenericResponse{
			Type:     GENERIC_RESPONSE_TYPE_ERROR,
			Response: map[string]any{"error": err.Error()},
		})
		os.Exit(1)
	}
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

	logResponseToStdout(GenericResponse{
		Type:     GENERIC_RESPONSE_TYPE_INFO,
		Response: map[string]any{"info": "loading plugin"},
	})

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
		stream = createStdioServer(payload)
	case RUN_MODE_TCP:
		// create a stream of clients that are connected to the plugin through a TCP connection
		// NOTE: for tcp, there will be multiple clients and the stream will close when the client is closed
		stream, err = createTCPServer(payload)
		if err != nil {
			return err
		}
	default:
		return fmt.Errorf("invalid run mode: %s", payload.RunMode)
	}

	// start a routine to handle the client stream
	for stream.Next() {
		client, err := stream.Read()
		if err != nil {
			continue
		}

		routine.Submit(nil, func() {
			handleClient(client, &declaration, runtime)
		})
	}

	return nil
}
