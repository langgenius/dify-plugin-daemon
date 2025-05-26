package withkey

import (
	"archive/zip"
	"bytes"
	"crypto/rsa"
	"crypto/sha256"
	"encoding/base64"
	"errors"
	"io"
	"path"
	"strconv"
	"time"

	"github.com/langgenius/dify-plugin-daemon/internal/utils/encryption"
	"github.com/langgenius/dify-plugin-daemon/internal/utils/parser"
	"github.com/langgenius/dify-plugin-daemon/pkg/plugin_packager/consts"
	"github.com/langgenius/dify-plugin-daemon/pkg/plugin_packager/decoder"
)

// SignPluginWithPrivateKey is a function that signs a plugin
// It takes a plugin as a stream of bytes and a private key to sign it with RSA-4096
func SignPluginWithPrivateKey(
	plugin []byte,
	verification *decoder.Verification,
	privateKey *rsa.PrivateKey,
) ([]byte, error) {
	decoder, err := decoder.NewZipPluginDecoder(plugin)
	if err != nil {
		return nil, err
	}

	if verification == nil {
		return nil, errors.New("verification cannot be nil")
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

	// write the verification into data
	// NOTE: .verification.dify.json is a special file that contains the verification information
	// and it will be placed at the end of the zip file, checksum is calculated using it also
	verificationBytes := parser.MarshalJsonBytes(verification)

	// write verification into the zip file
	fileWriter, err := zipWriter.Create(consts.VERIFICATION_FILE)
	if err != nil {
		return nil, err
	}

	if _, err := fileWriter.Write(verificationBytes); err != nil {
		return nil, err
	}

	// hash the verification
	hash := sha256.New()
	hash.Write(verificationBytes)
	hashed := hash.Sum(nil)

	// write the hash into data
	if _, err := data.Write(hashed); err != nil {
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

// Only used for testing
// WARNING: This function is deprecated, use SignPluginWithPrivateKey instead
func TraditionalSignPlugin(
	plugin []byte,
	privateKey *rsa.PrivateKey,
) ([]byte, error) {
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
