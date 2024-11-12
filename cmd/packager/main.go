package main

import (
	"flag"
	"os"

	"github.com/langgenius/dify-plugin-daemon/internal/core/plugin_packager/decoder"
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

	decoder, err := decoder.NewFSPluginDecoder(in_path)
	if err != nil {
		log.Panic("failed to create plugin decoder , plugin path: %s, error: %v", in_path, err)
	}

	packager := packager.NewPackager(decoder)
	zipFile, err := packager.Pack()

	if err != nil {
		log.Panic("failed to package plugin %v", err)
	}

	err = os.WriteFile(out_path, zipFile, 0644)
	if err != nil {
		log.Panic("failed to write package file %v", err)
	}

	log.Info("plugin packaged successfully, output path: %s", out_path)
}
