package signer

import (
	"archive/zip"
	"bytes"
	"crypto/sha256"
	"encoding/base64"
	"io"
	"path"
	"strconv"
	"time"

	"github.com/langgenius/dify-plugin-daemon/internal/core/license/private_key"
	"github.com/langgenius/dify-plugin-daemon/internal/core/plugin_packager/decoder"
	"github.com/langgenius/dify-plugin-daemon/internal/utils/encryption"
	"github.com/langgenius/dify-plugin-daemon/internal/utils/parser"
)

/*
	DifyPlugin is a file type that represents a plugin, it's designed to based on zip file format.
	When signing a plugin, we use RSA-4096 to create a signature for the plugin and write the signature
	into comment field of the zip file.
*/

// SignPlugin is a function that signs a plugin
// It takes a plugin as a stream of bytes and signs it with RSA-4096
func SignPlugin(plugin []byte) ([]byte, error) {
	// load private key
	privateKey, err := encryption.LoadPrivateKey(private_key.PRIVATE_KEY)
	if err != nil {
		return nil, err
	}

	decoder, err := decoder.NewZipPluginDecoder(plugin)
	if err != nil {
		return nil, err
	}

	// create a new zip writer
	zipBuffer := new(bytes.Buffer)
	zipWriter := zip.NewWriter(zipBuffer)

	defer zipWriter.Close()
	// store temporary hash
	data := new(bytes.Buffer)
	// read one by one
	err = decoder.Walk(func(filename, dir string) error {
		file, err := decoder.ReadFile(path.Join(dir, filename))
		if err != nil {
			return err
		}

		// calculate sha256 hash of the file
		hash := sha256.New()
		hash.Write(file)
		hashed := hash.Sum(nil)

		// write the hash into data
		data.Write(hashed)

		// create a new file in the zip writer
		fileWriter, err := zipWriter.Create(path.Join(dir, filename))
		if err != nil {
			return err
		}

		_, err = io.Copy(fileWriter, bytes.NewReader(file))
		if err != nil {
			return err
		}

		return nil
	})

	if err != nil {
		return nil, err
	}

	// get current time
	ct := time.Now().Unix()

	// convert time to bytes
	timeString := strconv.FormatInt(ct, 10)

	// write the time into data
	data.Write([]byte(timeString))

	// sign the data
	signature, err := encryption.RSASign(privateKey, data.Bytes())
	if err != nil {
		return nil, err
	}

	// write the signature into the comment field of the zip file
	comments := parser.MarshalJson(map[string]any{
		"signature": base64.StdEncoding.EncodeToString(signature),
		"time":      ct,
	})

	// write signature
	err = zipWriter.SetComment(comments)
	if err != nil {
		return nil, err
	}

	// close the zip writer
	err = zipWriter.Close()
	if err != nil {
		return nil, err
	}

	return zipBuffer.Bytes(), nil
}
