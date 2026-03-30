package signature

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/langgenius/dify-plugin-daemon/pkg/plugin_packager/decoder"
	"github.com/langgenius/dify-plugin-daemon/pkg/plugin_packager/signer/withkey"
	"github.com/langgenius/dify-plugin-daemon/pkg/utils/encryption"
	"github.com/langgenius/dify-plugin-daemon/pkg/utils/log"
)

func Sign(difypkgPath string, privateKeyPath string, verification *decoder.Verification) error {
	// read the plugin and private key
	plugin, err := os.ReadFile(difypkgPath)
	if err != nil {
		log.Error("failed to read plugin file", "error", err)
		return err
	}

	privateKeyBytes, err := os.ReadFile(privateKeyPath)
	if err != nil {
		log.Error("failed to read private key file", "error", err)
		return err
	}

	privateKey, err := encryption.LoadPrivateKey(privateKeyBytes)
	if err != nil {
		log.Error("failed to load private key", "error", err)
		return err
	}

	// sign the plugin
	pluginFile, err := withkey.SignPluginWithPrivateKey(plugin, verification, privateKey)
	if err != nil {
		log.Error("failed to sign plugin", "error", err)
		return err
	}

	// write the signed plugin to a file
	dir := filepath.Dir(difypkgPath)
	base := filepath.Base(difypkgPath)
	ext := filepath.Ext(base)
	name := strings.TrimSuffix(base, ext)
	outputPath := filepath.Join(dir, fmt.Sprintf("%s.signed%s", name, ext))

	err = os.WriteFile(outputPath, pluginFile, 0644)
	if err != nil {
		log.Error("failed to write signed plugin file", "error", err)
		return err
	}

	log.Info("plugin signed successfully", "output_path", outputPath)

	return nil
}
