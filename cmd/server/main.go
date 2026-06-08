package main

import (
	"context"
	"math"
	"net/http"
	"time"

	"github.com/kelseyhightower/envconfig"
	"github.com/langgenius/dify-plugin-daemon/internal/server"
	"github.com/langgenius/dify-plugin-daemon/internal/types/app"
	"github.com/langgenius/dify-plugin-daemon/pkg/utils/log"
)

const (
	officialPypiURL            = "https://pypi.org/simple/"
	pipMirrorAutoDetectTimeout = 3 * time.Second
)

var defaultMirrorCandidates = []string{
	"https://mirrors.aliyun.com/pypi/simple/",
	"https://pypi.tuna.tsinghua.edu.cn/simple/",
}

func main() {
	var config app.Config

	err := envconfig.Process("", &config)
	if err != nil {
		log.Panic("error processing environment variables", "error", err)
	}

	config.SetDefault()

	loc, err := time.LoadLocation(config.ServerTimeZone)
	if err != nil {
		log.Panic("load location error", "error", err)
	}
	time.Local = loc

	logCloser, err := log.Init(config.LogOutputFormat == "json", config.LogFile, config.LogLevel)
	if err != nil {
		log.Panic("failed to init logger", "error", err)
	}
	if logCloser != nil {
		defer func() {
			if err := logCloser.Close(); err != nil {
				log.Error("failed to close log file", "error", err)
			}
		}()
	}
	defer log.RecoverAndExit()

	applyPipMirrorAutoDetect(&config)

	if err = config.Validate(); err != nil {
		log.Panic("invalid configuration", "error", err)
	}

	// Initialize OpenTelemetry if enabled
	if config.EnableOtel {
		shutdown, err := server.InitTelemetry(&config)
		if err != nil {
			log.Panic("failed to init OpenTelemetry", "error", err)
		} else {
			defer shutdown(context.Background())
		}
	}

	(&server.App{}).Run(&config)
}

func applyPipMirrorAutoDetect(config *app.Config) {
	candidates := config.PipMirrorCandidates
	if len(candidates) == 0 {
		candidates = defaultMirrorCandidates
	}

	mirror := detectAndApplyPipMirror(config, &http.Client{Timeout: pipMirrorAutoDetectTimeout}, candidates, officialPypiURL)
	if mirror != "" {
		log.Info(
			"IMPORTANT: pip mirror auto-detect selected a mirror; set PIP_MIRROR_AUTO_DETECT=false to disable or PIP_MIRROR_URL=<mirror_url> to override",
			"mirror_url", mirror,
		)
	}
}

func detectAndApplyPipMirror(config *app.Config, client *http.Client, candidates []string, officialURL string) string {
	if !config.PipMirrorAutoDetect || config.PipMirrorUrl != "" {
		return ""
	}

	ctx, cancel := context.WithTimeout(context.Background(), pipMirrorAutoDetectTimeout)
	defer cancel()

	mirror := selectFastestMirror(ctx, client, candidates, officialURL)
	if mirror != "" {
		config.PipMirrorUrl = mirror
	}
	return mirror
}

type mirrorProbeResult struct {
	url     string
	latency time.Duration
	ok      bool
}

func probeURL(ctx context.Context, client *http.Client, url string) mirrorProbeResult {
	start := time.Now()
	req, err := http.NewRequestWithContext(ctx, http.MethodHead, url, nil)
	if err != nil {
		return mirrorProbeResult{url: url}
	}
	resp, err := client.Do(req)
	if err != nil {
		return mirrorProbeResult{url: url}
	}
	resp.Body.Close()
	if resp.StatusCode >= 400 {
		return mirrorProbeResult{url: url}
	}
	return mirrorProbeResult{url: url, latency: time.Since(start), ok: true}
}

func selectFastestMirror(ctx context.Context, client *http.Client, candidates []string, officialURL string) string {
	all := append([]string{officialURL}, candidates...)
	ch := make(chan mirrorProbeResult, len(all))

	for _, u := range all {
		go func(u string) {
			ch <- probeURL(ctx, client, u)
		}(u)
	}

	officialLatency := time.Duration(math.MaxInt64)
	bestURL := ""
	bestLatency := time.Duration(math.MaxInt64)

	for range all {
		r := <-ch
		if !r.ok {
			continue
		}
		if r.url == officialURL {
			officialLatency = r.latency
		} else if r.latency < bestLatency {
			bestLatency = r.latency
			bestURL = r.url
		}
	}

	if bestURL != "" && bestLatency < officialLatency {
		return bestURL
	}
	return ""
}
