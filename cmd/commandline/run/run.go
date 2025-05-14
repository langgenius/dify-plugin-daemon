package run

import (
	"bufio"
	"errors"
	"fmt"
	"os"

	"github.com/langgenius/dify-plugin-daemon/internal/core/plugin_manager/local_runtime"
	"github.com/langgenius/dify-plugin-daemon/internal/core/plugin_manager/test_utils"
	"github.com/langgenius/dify-plugin-daemon/internal/utils/log"
	"github.com/langgenius/dify-plugin-daemon/internal/utils/parser"
	"github.com/langgenius/dify-plugin-daemon/internal/utils/routine"
	"github.com/langgenius/dify-plugin-daemon/internal/utils/stream"
	"github.com/langgenius/dify-plugin-daemon/pkg/plugin_packager/decoder"
)

func handleClient(client client, runtime *local_runtime.LocalPluginRuntime) {
	// handle request from client
	scanner := bufio.NewScanner(client.reader)
	scanner.Buffer(make([]byte, 1024*1024), 15*1024*1024)

	for scanner.Scan() {
		payload := scanner.Bytes()
		invokePayload, err := parser.UnmarshalJsonBytes2Map(payload)
	}

}

func RunPlugin(payload RunPluginPayload) error {
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

	// try decode the plugin
	pluginFile, err := os.ReadFile(payload.PluginPath)
	if err != nil {
		return errors.Join(err, fmt.Errorf("read plugin file error"))
	}

	_, err = decoder.NewZipPluginDecoder(pluginFile)
	if err != nil {
		return errors.Join(err, fmt.Errorf("decode plugin file error"))
	}

	runtime, err := test_utils.GetRuntime(pluginFile, dir)
	if err != nil {
		return err
	}

	var stream *stream.Stream[client]
	if payload.RunMode == RUN_MODE_STDIO {
		stream = createStdioServer(payload)
	} else if payload.RunMode == RUN_MODE_TCP {
		stream, err = createTcpServer(payload)
		if err != nil {
			return err
		}
	}

	// start a routine to handle the client stream
	for stream.Next() {
		client, err := stream.Read()
		if err != nil {
			continue
		}

		routine.Submit(nil, func() {
			handleClient(client, runtime)
		})
	}

	return nil
}
