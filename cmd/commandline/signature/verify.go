package signature

import (
	"crypto/rsa"
	"os"

	"github.com/langgenius/dify-plugin-daemon/pkg/plugin_packager/decoder"
	"github.com/langgenius/dify-plugin-daemon/pkg/utils/encryption"
	"github.com/langgenius/dify-plugin-daemon/pkg/utils/log"
)

func Verify(difypkgPath string, publicKeyPath string) error {
	// read the plugin
	plugin, err := os.ReadFile(difypkgPath)
	if err != nil {
		log.Error("failed to read plugin file", "error", err)
		return err
	}

	decoderInstance, err := decoder.NewZipPluginDecoder(plugin)
	if err != nil {
		log.Error("failed to create plugin decoder", "plugin_path", difypkgPath, "error", err)
		return err
	}

	if publicKeyPath == "" {
		// verify the plugin with the official (bundled) public key
		err = decoder.VerifyPlugin(decoderInstance)
		if err != nil {
			log.Error("failed to verify plugin with official public key", "error", err)
			return err
		}
	} else {
		// read the public key
		publicKeyBytes, err := os.ReadFile(publicKeyPath)
		if err != nil {
			log.Error("failed to read public key file", "error", err)
			return err
		}

		publicKey, err := encryption.LoadPublicKey(publicKeyBytes)
		if err != nil {
			log.Error("failed to load public key", "error", err)
			return err
		}

		// verify the plugin
		err = decoder.VerifyPluginWithPublicKeys(decoderInstance, []*rsa.PublicKey{publicKey})
		if err != nil {
			log.Error("failed to verify plugin with provided public key", "error", err)
			return err
		}
	}

	log.Info("plugin verified successfully")
	return nil
}
