package main

import (
	"bytes"
	"encoding/json"
	"flag"
	"io"
	"log/slog"
	"os"

	"github.com/langgenius/dify-plugin-daemon/pkg/slim"
)

func main() {
	slog.SetDefault(slog.New(slog.NewTextHandler(io.Discard, nil)))
	id := flag.String("id", "", "plugin unique identifier")
	action := flag.String("action", "", "plugin access action")
	args := flag.String("args", "", "plugin invocation parameters (JSON); if omitted, read from stdin")
	configFile := flag.String("config", "", "path to JSON config file (replaces env vars)")
	flag.Parse()

	if *id == "" || *action == "" {
		fatal(slim.NewError(slim.ErrInvalidInput, "usage: slim -id <id> -action <action> [-args '<json>'] [-config <path>]"))
	}

	argsJSON := *args
	if argsJSON == "" {
		b, err := io.ReadAll(os.Stdin)
		if err != nil {
			fatal(slim.NewError(slim.ErrInvalidInput, "failed to read stdin: "+err.Error()))
		}
		if len(bytes.TrimSpace(b)) == 0 {
			fatal(slim.NewError(slim.ErrInvalidInput, "no -args flag and no JSON on stdin"))
		}
		argsJSON = string(b)
	}

	ctx, err := slim.NewInvokeContext(*id, *action, argsJSON)
	if err != nil {
		fatal(err)
	}

	var cfg *slim.SlimConfig
	if *configFile != "" {
		cfg, err = slim.LoadConfigFromFile(*configFile)
	} else {
		cfg, err = slim.LoadConfig()
	}
	if err != nil {
		fatal(err)
	}

	out := slim.NewOutputWriter(os.Stdout)

	switch cfg.Mode {
	case slim.ModeLocal:
		err = slim.RunLocal(ctx, &cfg.Local, out)
	case slim.ModeRemote:
		err = slim.RunRemote(ctx, &cfg.Remote, out)
	default:
		err = slim.NewError(slim.ErrUnknownMode, cfg.Mode)
	}

	if err != nil {
		fatal(err)
	}
}

func fatal(err error) {
	exitCode := slim.ExitPluginError
	if se, ok := err.(*slim.SlimError); ok {
		exitCode = se.ExitCode()
	}
	b, _ := json.Marshal(err)
	os.Stderr.Write(b)
	os.Stderr.Write([]byte("\n"))
	os.Exit(exitCode)
}
