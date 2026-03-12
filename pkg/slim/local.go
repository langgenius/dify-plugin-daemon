package slim

import (
	"bufio"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"time"

	"github.com/langgenius/dify-plugin-daemon/internal/core/local_runtime"
	"github.com/langgenius/dify-plugin-daemon/internal/types/app"
	"github.com/langgenius/dify-plugin-daemon/pkg/entities/plugin_entities"
	"github.com/langgenius/dify-plugin-daemon/pkg/plugin_packager/decoder"
	"github.com/langgenius/dify-plugin-daemon/pkg/utils/parser"
	"github.com/langgenius/dify-plugin-daemon/pkg/utils/routine"
)

const downloadTimeout = 60 * time.Second

func RunLocal(ctx *InvokeContext, local *LocalConfig, out *OutputWriter) error {
	workingPath := pluginWorkingPath(local.Folder, ctx.PluginID)

	dec, err := decoder.NewFSPluginDecoder(workingPath)
	if err != nil {
		out.Message("download", fmt.Sprintf("downloading plugin %s from marketplace", ctx.PluginID))
		dec, err = downloadAndExtract(local, ctx.PluginID, workingPath)
		if err != nil {
			return err
		}
		out.Message("download", "plugin downloaded and extracted")
	}

	if !routine.IsInit() {
		routine.InitPool(4)
	}

	appConfig := local.toAppConfig()
	rt, err := buildRuntime(appConfig, dec, workingPath)
	if err != nil {
		return NewError(ErrPluginInit, fmt.Sprintf("build runtime: %s", err))
	}

	out.Message("init", "initializing python environment")
	if err := rt.InitPythonEnvironment(); err != nil {
		return NewError(ErrPluginInit, fmt.Sprintf("init python env: %s", err))
	}
	out.Message("init", "python environment ready")

	reqBytes, sessionID, err := TransformRequest(ctx)
	if err != nil {
		return err
	}

	return execPlugin(rt, local, reqBytes, sessionID, appConfig, out)
}

func buildRuntime(
	appConfig *app.Config,
	dec *decoder.FSPluginDecoder,
	workingPath string,
) (*local_runtime.LocalPluginRuntime, error) {
	manifest, err := dec.Manifest()
	if err != nil {
		return nil, fmt.Errorf("read manifest: %w", err)
	}

	return local_runtime.NewLocalPluginRuntime(
		appConfig,
		dec,
		manifest,
		workingPath,
	), nil
}

func downloadAndExtract(local *LocalConfig, pluginID, workingPath string) (*decoder.FSPluginDecoder, error) {
	pkgBytes, err := downloadFromMarketplace(local.MarketplaceURL, pluginID)
	if err != nil {
		return nil, err
	}

	zipDec, err := decoder.NewZipPluginDecoder(pkgBytes)
	if err != nil {
		return nil, NewError(ErrPluginPackageInvalid, fmt.Sprintf("invalid plugin package: %s", err))
	}

	if err := os.MkdirAll(workingPath, 0755); err != nil {
		return nil, NewError(ErrPluginExtract, fmt.Sprintf("create working dir: %s", err))
	}

	if err := zipDec.ExtractTo(workingPath); err != nil {
		os.RemoveAll(workingPath)
		return nil, NewError(ErrPluginExtract, fmt.Sprintf("extract package: %s", err))
	}

	fsDec, err := decoder.NewFSPluginDecoder(workingPath)
	if err != nil {
		os.RemoveAll(workingPath)
		return nil, NewError(ErrPluginExtract, fmt.Sprintf("load extracted plugin: %s", err))
	}

	return fsDec, nil
}

func downloadFromMarketplace(marketplaceURL, pluginID string) ([]byte, error) {
	u, err := url.Parse(strings.TrimRight(marketplaceURL, "/") + "/api/v1/plugins/download")
	if err != nil {
		return nil, NewError(ErrPluginDownload, fmt.Sprintf("parse marketplace url: %s", err))
	}
	q := u.Query()
	q.Set("unique_identifier", pluginID)
	u.RawQuery = q.Encode()

	client := &http.Client{Timeout: downloadTimeout}
	resp, err := client.Get(u.String())
	if err != nil {
		if os.IsTimeout(err) {
			return nil, NewError(ErrPluginDownloadTimeout,
				fmt.Sprintf("marketplace download timed out after %s for %s", downloadTimeout, pluginID))
		}
		return nil, NewError(ErrPluginDownload, fmt.Sprintf("marketplace request failed: %s", err))
	}
	defer resp.Body.Close()

	if resp.StatusCode == http.StatusNotFound {
		return nil, NewError(ErrPluginNotFound,
			fmt.Sprintf("plugin %s not found in marketplace (%s)", pluginID, marketplaceURL))
	}
	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(io.LimitReader(resp.Body, 1024))
		return nil, NewError(ErrPluginDownload,
			fmt.Sprintf("marketplace returned status %d for %s: %s", resp.StatusCode, pluginID, string(body)))
	}

	const maxSize = 15 * 1024 * 1024
	data, err := io.ReadAll(io.LimitReader(resp.Body, maxSize+1))
	if err != nil {
		return nil, NewError(ErrPluginDownload, fmt.Sprintf("read response body: %s", err))
	}
	if len(data) > maxSize {
		return nil, NewError(ErrPluginPackageTooLarge,
			fmt.Sprintf("plugin package %s exceeds 15 MiB size limit", pluginID))
	}

	return data, nil
}

func execPlugin(
	rt *local_runtime.LocalPluginRuntime,
	local *LocalConfig,
	reqBytes []byte,
	sessionID string,
	appConfig *app.Config,
	out *OutputWriter,
) error {
	pythonPath, err := filepath.Abs(filepath.Join(rt.State.WorkingPath, ".venv", "bin", "python"))
	if err != nil {
		return NewError(ErrPluginExec, fmt.Sprintf("resolve python path: %s", err))
	}

	cmd := exec.Command(pythonPath, "-m", rt.Config.Meta.Runner.Entrypoint)
	cmd.Dir = rt.State.WorkingPath
	cmd.Env = append(os.Environ(), "INSTALL_METHOD=local", "PATH="+os.Getenv("PATH"))

	stdin, err := cmd.StdinPipe()
	if err != nil {
		return NewError(ErrPluginExec, fmt.Sprintf("stdin pipe: %s", err))
	}
	stdout, err := cmd.StdoutPipe()
	if err != nil {
		return NewError(ErrPluginExec, fmt.Sprintf("stdout pipe: %s", err))
	}
	stderr, err := cmd.StderrPipe()
	if err != nil {
		return NewError(ErrPluginExec, fmt.Sprintf("stderr pipe: %s", err))
	}

	if err := cmd.Start(); err != nil {
		return NewError(ErrPluginExec, fmt.Sprintf("start subprocess: %s", err))
	}

	stderrCh := make(chan string, 1)
	go func() {
		b, _ := io.ReadAll(stderr)
		stderrCh <- string(b)
	}()

	if _, err := stdin.Write(append(reqBytes, '\n')); err != nil {
		cmd.Process.Kill()
		return NewError(ErrPluginExec, fmt.Sprintf("write stdin: %s", err))
	}
	stdin.Close()

	timeout := time.Duration(local.MaxExecutionTimeout) * time.Second
	deadline := time.Now().Add(timeout)

	scanner := bufio.NewScanner(stdout)
	scanner.Buffer(
		make([]byte, appConfig.GetLocalRuntimeBufferSize()),
		appConfig.GetLocalRuntimeMaxBufferSize(),
	)

	var execErr error
	done := false

	for scanner.Scan() {
		if time.Now().After(deadline) {
			execErr = NewError(ErrPluginExec, "execution timeout")
			break
		}

		data := scanner.Bytes()
		if len(data) == 0 {
			continue
		}

		plugin_entities.ParsePluginUniversalEvent(
			data,
			"",
			func(sid string, payload []byte) {
				if sid != sessionID {
					return
				}
				msg, err := parser.UnmarshalJsonBytes[plugin_entities.SessionMessage](payload)
				if err != nil {
					execErr = NewError(ErrStreamParse, err.Error())
					done = true
					return
				}
				switch msg.Type {
				case plugin_entities.SESSION_MESSAGE_TYPE_STREAM:
					out.Chunk(json.RawMessage(msg.Data))
				case plugin_entities.SESSION_MESSAGE_TYPE_END:
					out.Done()
					done = true
				case plugin_entities.SESSION_MESSAGE_TYPE_ERROR:
					errResp, parseErr := parser.UnmarshalJsonBytes[plugin_entities.ErrorResponse](msg.Data)
					if parseErr != nil {
						out.Error(ErrPluginExec, string(msg.Data))
					} else {
						out.Error(ErrPluginExec, errResp.Error())
					}
					done = true
				}
			},
			func() {
				deadline = time.Now().Add(timeout)
			},
			func(errMsg string) {
				out.Error(ErrPluginExec, errMsg)
				done = true
			},
			func(logEvent plugin_entities.PluginLogEvent) {
			},
		)

		if done || execErr != nil {
			break
		}
	}

	if scanErr := scanner.Err(); scanErr != nil && execErr == nil {
		execErr = NewError(ErrStreamRead, scanErr.Error())
	}

	cmd.Process.Kill()
	cmd.Wait()

	if execErr != nil {
		stderrMsg := <-stderrCh
		if stderrMsg != "" {
			return NewError(execErr.(*SlimError).Code, execErr.Error()+"; stderr: "+truncate(stderrMsg, 512))
		}
		return execErr
	}

	return nil
}

func pluginWorkingPath(folder, pluginID string) string {
	normalized := strings.ReplaceAll(pluginID, ":", "-")
	return filepath.Join(folder, normalized)
}

func truncate(s string, max int) string {
	if len(s) <= max {
		return s
	}
	return s[:max] + "..."
}
