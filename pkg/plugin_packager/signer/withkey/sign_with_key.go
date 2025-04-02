package withkey

import (
	"archive/zip"
	"bytes"
	"crypto/rsa"
	"crypto/sha256"
	"encoding/base64"
	"io"
	"path"
	"strconv"
	"time"

	"github.com/langgenius/dify-plugin-daemon/internal/utils/encryption"
	"github.com/langgenius/dify-plugin-daemon/internal/utils/parser"
	"github.com/langgenius/dify-plugin-daemon/pkg/plugin_packager/decoder"
)

// SignPluginWithPrivateKey is a function that signs a plugin
// It takes a plugin as a stream of bytes and a private key to sign it with RSA-4096
func SignPluginWithPrivateKey(plugin []byte, privateKey *rsa.PrivateKey) ([]byte, error) {
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
