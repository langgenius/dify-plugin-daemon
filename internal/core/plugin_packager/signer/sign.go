package signer

import (
	"archive/zip"
	"bytes"
	"crypto/sha256"
	"encoding/base64"
	"io"
	"strconv"
	"time"

	"github.com/langgenius/dify-plugin-daemon/internal/core/license/private_key"
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
	private_key, err := encryption.LoadPrivateKey(private_key.PRIVATE_KEY)
	if err != nil {
		return nil, err
	}

	// construct zip
	zip_reader, err := zip.NewReader(bytes.NewReader(plugin), int64(len(plugin)))
	if err != nil {
		return nil, err
	}

	data := new(bytes.Buffer)
	// read one by one
	for _, file := range zip_reader.File {
		// read file bytes
		file_reader, err := file.Open()
		if err != nil {
			return nil, err
		}
		defer file_reader.Close()

		temp_data := new(bytes.Buffer)
		_, err = temp_data.ReadFrom(file_reader)
		if err != nil {
			return nil, err
		}

		// calculate sha256 hash of the file
		hash := sha256.New()
		hash.Write(temp_data.Bytes())
		hashed := hash.Sum(nil)

		// write the hash into data
		data.Write(hashed)
	}

	// get current time
	ct := time.Now().Unix()

	// convert time to bytes
	time_string := strconv.FormatInt(ct, 10)

	// write the time into data
	data.Write([]byte(time_string))

	// sign the data
	signature, err := encryption.RSASign(private_key, data.Bytes())
	if err != nil {
		return nil, err
	}

	// write the signature into the comment field of the zip file
	zip_buffer := new(bytes.Buffer)
	zip_writer := zip.NewWriter(zip_buffer)
	for _, file := range zip_reader.File {
		file_writer, err := zip_writer.Create(file.Name)
		if err != nil {
			return nil, err
		}

		file_reader, err := file.Open()
		if err != nil {
			return nil, err
		}
		defer file_reader.Close()

		_, err = io.Copy(file_writer, file_reader)
		if err != nil {
			return nil, err
		}
	}

	comments := parser.MarshalJson(map[string]any{
		"signature": base64.StdEncoding.EncodeToString(signature),
		"time":      ct,
	})

	// write signature
	err = zip_writer.SetComment(comments)
	if err != nil {
		return nil, err
	}

	// close the zip writer
	err = zip_writer.Close()
	if err != nil {
		return nil, err
	}

	return zip_buffer.Bytes(), nil
}
