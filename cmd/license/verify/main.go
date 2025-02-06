package main

import (
	"flag"
	"os"

	"github.com/langgenius/dify-plugin-daemon/internal/utils/log"
	"github.com/langgenius/dify-plugin-daemon/pkg/plugin_packager/decoder"
)

func main() {
	var (
		in_path string
		help    bool
	)

	flag.StringVar(&in_path, "in", "", "input plugin file path")
	flag.BoolVar(&help, "help", false, "show help")
	flag.Parse()

	if help || in_path == "" {
		flag.PrintDefaults()
		os.Exit(0)
	}

	// read plugin
	plugin, err := os.ReadFile(in_path)
	if err != nil {
		log.Panic("failed to read plugin file %v", err)
	}

	// decode
	decoderInstance, err := decoder.NewZipPluginDecoder(plugin)
	if err != nil {
		log.Panic("failed to create plugin decoder , plugin path: %s, error: %v", in_path, err)
	}

	// sign plugin
	err = decoder.VerifyPlugin(decoderInstance)
	if err != nil {
		log.Panic("failed to verify plugin %v", err)
	}

	log.Info("plugin verified successfully")
}
