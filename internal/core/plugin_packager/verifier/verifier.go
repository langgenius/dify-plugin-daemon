package verifier

import (
	"archive/zip"
	"bytes"
	"crypto/sha256"
	"encoding/base64"
	"strconv"

	"github.com/langgenius/dify-plugin-daemon/internal/core/license/public_key"
	"github.com/langgenius/dify-plugin-daemon/internal/utils/encryption"
	"github.com/langgenius/dify-plugin-daemon/internal/utils/parser"
)

// VerifyPlugin is a function that verifies the signature of a plugin
// It takes a plugin as a stream of bytes and verifies the signature
func VerifyPlugin(archive []byte) error {
	// load public key
	public_key, err := encryption.LoadPublicKey(public_key.PUBLIC_KEY)
	if err != nil {
		return err
	}

	// construct zip
	zip_reader, err := zip.NewReader(bytes.NewReader(archive), int64(len(archive)))
	if err != nil {
		return err
	}

	data := new(bytes.Buffer)
	// read one by one
	for _, file := range zip_reader.File {
		// read file bytes
		file_reader, err := file.Open()
		if err != nil {
			return err
		}
		defer file_reader.Close()

		temp_data := new(bytes.Buffer)
		_, err = temp_data.ReadFrom(file_reader)
		if err != nil {
			return err
		}

		hash := sha256.New()
		hash.Write(temp_data.Bytes())
		hashed := hash.Sum(nil)

		// write the hash into data
		data.Write(hashed)
	}

	// get the signature
	signature := zip_reader.Comment

	type signatureData struct {
		Signature string `json:"signature"`
		Time      int64  `json:"time"`
	}

	signature_data, err := parser.UnmarshalJson[signatureData](signature)
	if err != nil {
		return err
	}

	plugin_sig := signature_data.Signature
	plugin_time := signature_data.Time

	// convert time to bytes
	time_string := strconv.FormatInt(plugin_time, 10)

	// write the time into data
	data.Write([]byte(time_string))

	sig_bytes, err := base64.StdEncoding.DecodeString(plugin_sig)
	if err != nil {
		return err
	}

	// verify signature
	err = encryption.VerifySign(public_key, data.Bytes(), sig_bytes)
	return err
}
