package main

import (
	"flag"
	"os"

	"github.com/langgenius/dify-plugin-daemon/internal/core/plugin_packager/packager"
	"github.com/langgenius/dify-plugin-daemon/internal/utils/log"
)

func main() {
	var (
		in_path  string
		out_path string
		help     bool
	)

	flag.StringVar(&in_path, "in", "", "directory of plugin to be packaged")
	flag.StringVar(&out_path, "out", "", "package output path")
	flag.BoolVar(&help, "help", false, "show help")
	flag.Parse()

	if help || in_path == "" || out_path == "" {
		flag.PrintDefaults()
		os.Exit(0)
	}

	packager := packager.NewPackager(in_path)
	zip_file, err := packager.Pack()

	if err != nil {
		log.Panic("failed to package plugin %v", err)
	}

	err = os.WriteFile(out_path, zip_file, 0644)
	if err != nil {
		log.Panic("failed to write package file %v", err)
	}

	log.Info("plugin packaged successfully, output path: %s", out_path)
}
