package main

import (
	"bufio"
	"errors"
	"fmt"
	"io"
	"net/http"
	"strings"
	"time"

	"github.com/langgenius/dify-plugin-daemon/internal/types/app"
	"github.com/langgenius/dify-plugin-daemon/pkg/utils/log"
)

const (
	alibabaCloudPypiMirrorURL  = "https://mirrors.aliyun.com/pypi/simple/"
	pipMirrorAutoDetectTimeout = 2 * time.Second
)

var cloudflareTraceURL = "https://cloudflare.com/cdn-cgi/trace"

func applyPipMirrorAutoDetect(config *app.Config) {
	loc, applied, err := detectAndApplyPipMirror(config, &http.Client{Timeout: pipMirrorAutoDetectTimeout}, cloudflareTraceURL)
	if err != nil {
		log.Warn("failed to auto-detect pip mirror", "error", err)
		return
	}

	if applied {
		log.Info("auto-detected pip mirror", "location", loc, "mirror_url", config.PipMirrorUrl)
	}
}

func detectAndApplyPipMirror(config *app.Config, client *http.Client, traceURL string) (string, bool, error) {
	if !config.PipMirrorAutoDetect || config.PipMirrorUrl != "" {
		return "", false, nil
	}

	loc, err := detectCloudflareLocation(client, traceURL)
	if err != nil {
		return "", false, err
	}

	if loc != "CN" {
		return loc, false, nil
	}

	config.PipMirrorUrl = alibabaCloudPypiMirrorURL
	return loc, true, nil
}

func detectCloudflareLocation(client *http.Client, traceURL string) (string, error) {
	req, err := http.NewRequest(http.MethodGet, traceURL, nil)
	if err != nil {
		return "", fmt.Errorf("create cloudflare trace request: %w", err)
	}

	resp, err := client.Do(req)
	if err != nil {
		return "", fmt.Errorf("request cloudflare trace: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return "", fmt.Errorf("cloudflare trace returned status %s", resp.Status)
	}

	loc, err := parseCloudflareTraceLocation(resp.Body)
	if err != nil {
		return "", fmt.Errorf("parse cloudflare trace response: %w", err)
	}

	return loc, nil
}

func parseCloudflareTraceLocation(body io.Reader) (string, error) {
	scanner := bufio.NewScanner(body)
	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())
		if line == "" {
			continue
		}

		key, value, found := strings.Cut(line, "=")
		if !found || !strings.EqualFold(strings.TrimSpace(key), "loc") {
			continue
		}

		loc := strings.ToUpper(strings.TrimSpace(value))
		if loc != "" {
			return loc, nil
		}
	}

	if err := scanner.Err(); err != nil {
		return "", err
	}

	return "", errors.New("cloudflare trace response missing loc")
}
