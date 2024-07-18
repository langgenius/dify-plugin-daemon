package verifier

import (
	"bytes"
	"crypto/sha256"
	"encoding/base64"
	"path"
	"strconv"

	"github.com/langgenius/dify-plugin-daemon/internal/core/license/public_key"
	"github.com/langgenius/dify-plugin-daemon/internal/core/plugin_packager/decoder"
	"github.com/langgenius/dify-plugin-daemon/internal/utils/encryption"
)

// VerifyPlugin is a function that verifies the signature of a plugin
// It takes a plugin decoder and verifies the signature
func VerifyPlugin(decoder decoder.PluginDecoder) error {
	// load public key
	public_key, err := encryption.LoadPublicKey(public_key.PUBLIC_KEY)
	if err != nil {
		return err
	}

	data := new(bytes.Buffer)
	// read one by one
	err = decoder.Walk(func(filename, dir string) error {
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
	created_at, err := decoder.CreateTime()
	if err != nil {
		return err
	}

	// write the time into data
	data.Write([]byte(strconv.FormatInt(created_at, 10)))

	sig_bytes, err := base64.StdEncoding.DecodeString(signature)
	if err != nil {
		return err
	}

	// verify signature
	err = encryption.VerifySign(public_key, data.Bytes(), sig_bytes)
	return err
}
