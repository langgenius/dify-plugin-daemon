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
func VerifyPlugin(decoder PluginDecoder) error {
	var publicKeys []*rsa.PublicKey

	// load official public key
	officialPublicKey, err := encryption.LoadPublicKey(public_key.PUBLIC_KEY)
	if err != nil {
		return err
	}
	publicKeys = append(publicKeys, officialPublicKey)

	// verify the plugin
	return VerifyPluginWithPublicKeys(decoder, publicKeys)
}

// VerifyPluginWithPublicKeyPaths is a function that verifies the signature of a plugin
// It takes a plugin decoder and a list of public key paths to verify the signature
func VerifyPluginWithPublicKeyPaths(decoder PluginDecoder, publicKeyPaths []string) error {
	var publicKeys []*rsa.PublicKey

	// load official public key
	officialPublicKey, err := encryption.LoadPublicKey(public_key.PUBLIC_KEY)
	if err != nil {
		return err
	}
	publicKeys = append(publicKeys, officialPublicKey)

	// load keys provided in the arguments
	for _, publicKeyPath := range publicKeyPaths {
		// open file by trimming the spaces in path
		keyBytes, err := os.ReadFile(strings.TrimSpace(publicKeyPath))
		if err != nil {
			return err
		}
		publicKey, err := encryption.LoadPublicKey(keyBytes)
		if err != nil {
			return err
		}
		publicKeys = append(publicKeys, publicKey)
	}

	return VerifyPluginWithPublicKeys(decoder, publicKeys)
}

// VerifyPluginWithPublicKeys is a function that verifies the signature of a plugin
// It takes a plugin decoder and a list of public keys to verify the signature
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
