package main

import (
	"bytes"
	"encoding/json"
	"flag"
	"fmt"
	"io"
	"log/slog"
	"os"

	"github.com/langgenius/dify-plugin-daemon/pkg/slim"
)

func main() {
	slog.SetDefault(slog.New(slog.NewTextHandler(io.Discard, nil)))
	if len(os.Args) > 1 && os.Args[1] == "extract" {
		runExtractCommand(os.Args[2:])
		return
	}

	id := flag.String("id", "", "plugin unique identifier")
	action := flag.String("action", "", "plugin access action")
	args := flag.String("args", "", "plugin invocation parameters (JSON); if omitted, read from stdin")
	configFile := flag.String("config", "", "path to JSON config file (replaces env vars)")
	flag.Usage = rootUsage
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

func runExtractCommand(args []string) {
	fs := flag.NewFlagSet("extract", flag.ContinueOnError)
	fs.SetOutput(os.Stderr)
	id := fs.String("id", "", "plugin unique identifier")
	action := fs.String("action", "", "generate -args data for this Slim action")
	path := fs.String("path", "", "local plugin directory or .difypkg path")
	output := fs.String("output", slim.OutputJSON, "output format")
	configFile := fs.String("config", "", "path to JSON config file (replaces env vars)")
	fs.Usage = func() {
		extractUsage(fs)
	}
	if err := fs.Parse(args); err != nil {
		if err == flag.ErrHelp {
			os.Exit(slim.ExitOK)
		}
		fatal(slim.NewError(slim.ErrInvalidInput, err.Error()))
	}
	if fs.NArg() != 0 {
		fatal(slim.NewError(slim.ErrInvalidInput, "unexpected positional arguments"))
	}

	cfg, err := slim.LoadExtractConfig(*configFile, *path != "")
	if err != nil {
		fatal(err)
	}

	if err := slim.RunExtract(cfg, slim.ExtractOptions{
		PluginID: *id,
		Action:   *action,
		Path:     *path,
		Output:   *output,
	}, os.Stdout); err != nil {
		fatal(err)
	}
}

func rootUsage() {
	w := flag.CommandLine.Output()
	fmt.Fprintln(w, "Usage:")
	fmt.Fprintln(w, "  slim -id <plugin_unique_identifier> -action <action> [-args '<json>'] [-config <path>]")
	fmt.Fprintln(w, "  slim extract -id <plugin_unique_identifier> [-config <path>] [-output json]")
	fmt.Fprintln(w, "  slim extract -id <plugin_unique_identifier> -action <action> [-config <path>] [-output json]")
	fmt.Fprintln(w, "  slim extract -path <plugin-dir-or-difypkg> [-config <path>] [-output json]")
	fmt.Fprintln(w, "  slim extract -path <plugin-dir-or-difypkg> -action <action> [-config <path>] [-output json]")
	fmt.Fprintln(w)
	fmt.Fprintln(w, "Commands:")
	fmt.Fprintln(w, "  extract    Parse plugin schema/declaration from daemon or local files")
	fmt.Fprintln(w)
	fmt.Fprintln(w, "Invocation flags:")
	flag.PrintDefaults()
}

func extractUsage(fs *flag.FlagSet) {
	w := fs.Output()
	fmt.Fprintln(w, "Usage:")
	fmt.Fprintln(w, "  slim extract -id <plugin_unique_identifier> [-config <path>] [-output json]")
	fmt.Fprintln(w, "  slim extract -id <plugin_unique_identifier> -action <action> [-config <path>] [-output json]")
	fmt.Fprintln(w, "  slim extract -path <plugin-dir-or-difypkg> [-config <path>] [-output json]")
	fmt.Fprintln(w, "  slim extract -path <plugin-dir-or-difypkg> -action <action> [-config <path>] [-output json]")
	fmt.Fprintln(w)
	fmt.Fprintln(w, "Extract flags:")
	fs.PrintDefaults()
}

func fatal(err error) {
	exitCode := slim.ExitPluginError
	var errorToMarshal *slim.SlimError

	if se, ok := err.(*slim.SlimError); ok {
		errorToMarshal = se
		exitCode = se.ExitCode()
	} else {
		// Wrap non-SlimError types to ensure they are marshalled to JSON correctly.
		errorToMarshal = slim.NewError(slim.ErrPluginExec, err.Error())
	}

	b, marshalErr := json.Marshal(errorToMarshal)
	if marshalErr != nil {
		// This should be practically impossible since SlimError is designed for JSON.
		// As a last resort, print a hardcoded error message.
		os.Stderr.Write([]byte(`{"code":"INTERNAL_ERROR","message":"failed to marshal error to JSON"}`))
		os.Stderr.Write([]byte("\n"))
		os.Exit(slim.ExitPluginError)
	}

	os.Stderr.Write(b)
	os.Stderr.Write([]byte("\n"))
	os.Exit(exitCode)
}
