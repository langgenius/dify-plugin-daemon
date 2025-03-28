package decoder

import (
	"bytes"
	"crypto/rsa"
	"crypto/sha256"
	"encoding/base64"
	"os"
	"path"
	"strconv"
	"strings"

	"github.com/langgenius/dify-plugin-daemon/internal/core/license/public_key"
	"github.com/langgenius/dify-plugin-daemon/internal/utils/encryption"
)

// VerifyPlugin is a function that verifies the signature of a plugin
// It takes a plugin decoder and verifies the signature with a bundled public key
// and the public keys specified in the environment variable
func VerifyPlugin(decoder PluginDecoder) error {

	var publicKeys []*rsa.PublicKey

	// load official public key
	officialPublicKey, err := encryption.LoadPublicKey(public_key.PUBLIC_KEY)
	if err != nil {
		return err
	}
	publicKeys = append(publicKeys, officialPublicKey)

	// load keys specified in environment variable if third party signature verification is enabled
	if strings.ToLower(os.Getenv("THIRD_PARTY_SIGNATURE_VERIFICATION_ENABLED")) == "true" {
		thirdPartyPublicKeys := os.Getenv("THIRD_PARTY_SIGNATURE_VERIFICATION_PUBLIC_KEYS")
		if thirdPartyPublicKeys != "" {
			keyPaths := strings.Split(thirdPartyPublicKeys, ",")
			for _, keyPath := range keyPaths {
				keyBytes, err := os.ReadFile(strings.TrimSpace(keyPath))
				if err != nil {
					return err
				}
				publicKey, err := encryption.LoadPublicKey(keyBytes)
				if err != nil {
					return err
				}
				publicKeys = append(publicKeys, publicKey)
			}
		}
	}

	// verify the plugin
	return VerifyPluginWithPublicKeys(decoder, publicKeys)
}

// VerifyPluginWithPublicKeys is a function that verifies the signature of a plugin
// It takes a plugin decoder and a public key to verify the signature
func VerifyPluginWithPublicKeys(decoder PluginDecoder, publicKeys []*rsa.PublicKey) error {
	data := new(bytes.Buffer)
	// read one by one
	err := decoder.Walk(func(filename, dir string) error {
		// read file bytes
		file, err := decoder.ReadFile(path.Join(dir, filename))
		if err != nil {
			return err
		}

		hash := sha256.New()
		hash.Write(file)

		// write the hash into data
		data.Write(hash.Sum(nil))
		return nil
	})

	if err != nil {
		return err
	}

	// get the signature
	signature, err := decoder.Signature()
	if err != nil {
		return err
	}

	// get the time
	createdAt, err := decoder.CreateTime()
	if err != nil {
		return err
	}

	// write the time into data
	data.Write([]byte(strconv.FormatInt(createdAt, 10)))

	sigBytes, err := base64.StdEncoding.DecodeString(signature)
	if err != nil {
		return err
	}

	// verify signature
	var lastErr error
	for _, publicKey := range publicKeys {
		lastErr = encryption.VerifySign(publicKey, data.Bytes(), sigBytes)
		if lastErr == nil {
			return nil
		}
	}
	return lastErr
}
