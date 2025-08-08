package decoder

import (
	"archive/zip"
	"bytes"
	"errors"
	"fmt"
	"io"
	"io/fs"
	"os"
	"path"
	"path/filepath"
	"strconv"
	"strings"

	"github.com/langgenius/dify-plugin-daemon/internal/utils/parser"
	"github.com/langgenius/dify-plugin-daemon/pkg/entities/plugin_entities"
	"github.com/langgenius/dify-plugin-daemon/pkg/plugin_packager/consts"
)

type ZipPluginDecoder struct {
	PluginDecoder
	PluginDecoderHelper

	reader *zip.Reader
	err    error

	sig          string
	createTime   int64
	verification *Verification

	thirdPartySignatureVerificationConfig *ThirdPartySignatureVerificationConfig
}

type ThirdPartySignatureVerificationConfig struct {
	Enabled        bool
	PublicKeyPaths []string
}

func newZipPluginDecoder(
	binary []byte,
	thirdPartySignatureVerificationConfig *ThirdPartySignatureVerificationConfig,
) (*ZipPluginDecoder, error) {
	reader, err := zip.NewReader(bytes.NewReader(binary), int64(len(binary)))
	if err != nil {
		return nil, errors.New(strings.ReplaceAll(err.Error(), "zip", "difypkg"))
	}

	decoder := &ZipPluginDecoder{
		reader:                                reader,
		err:                                   err,
		thirdPartySignatureVerificationConfig: thirdPartySignatureVerificationConfig,
	}

	err = decoder.Open()
	if err != nil {
		return nil, err
	}

	if _, err := decoder.Manifest(); err != nil {
		return nil, err
	}

	return decoder, nil
}

// NewZipPluginDecoder is a helper function to create ZipPluginDecoder
func NewZipPluginDecoder(binary []byte) (*ZipPluginDecoder, error) {
	return newZipPluginDecoder(binary, nil)
}

// NewZipPluginDecoderWithThirdPartySignatureVerificationConfig is a helper function
// to create a ZipPluginDecoder with a third party signature verification
func NewZipPluginDecoderWithThirdPartySignatureVerificationConfig(
	binary []byte,
	thirdPartySignatureVerificationConfig *ThirdPartySignatureVerificationConfig,
) (*ZipPluginDecoder, error) {
	return newZipPluginDecoder(binary, thirdPartySignatureVerificationConfig)
}

// NewZipPluginDecoderWithSizeLimit is a helper function to create a ZipPluginDecoder with a size limit
// It checks the total uncompressed size of the plugin package and returns an error if it exceeds the max size
func NewZipPluginDecoderWithSizeLimit(binary []byte, maxSize int64) (*ZipPluginDecoder, error) {
	reader, err := zip.NewReader(bytes.NewReader(binary), int64(len(binary)))
	if err != nil {
		return nil, errors.New(strings.ReplaceAll(err.Error(), "zip", "difypkg"))
	}

	totalSize := int64(0)
	for _, file := range reader.File {
		totalSize += int64(file.UncompressedSize64)
		if totalSize > maxSize {
			return nil, errors.New(
				"plugin package size is too large, please ensure the uncompressed size is less than " +
					strconv.FormatInt(maxSize, 10) + " bytes",
			)
		}
	}

	return newZipPluginDecoder(binary, nil)
}

func (z *ZipPluginDecoder) Stat(filename string) (fs.FileInfo, error) {
	f, err := z.reader.Open(filename)
	if err != nil {
		return nil, err
	}
	defer f.Close()

	return f.Stat()
}

func (z *ZipPluginDecoder) Open() error {
	if z.reader == nil {
		return z.err
	}

	return nil
}

func (z *ZipPluginDecoder) Walk(fn func(filename string, dir string) error) error {
	if z.reader == nil {
		return z.err
	}

	for _, file := range z.reader.File {
		// split the path into directory and filename
		dir, filename := path.Split(file.Name)
		if err := fn(filename, dir); err != nil {
			return err
		}
	}

	return nil
}

func (z *ZipPluginDecoder) Close() error {
	return nil
}

func (z *ZipPluginDecoder) ReadFile(filename string) ([]byte, error) {
	if z.reader == nil {
		return nil, z.err
	}

	file, err := z.reader.Open(filename)
	if err != nil {
		return nil, err
	}
	defer file.Close()

	data := new(bytes.Buffer)
	_, err = data.ReadFrom(file)
	if err != nil {
		return nil, err
	}

	return data.Bytes(), nil
}

func (z *ZipPluginDecoder) ReadDir(dirname string) ([]string, error) {
	if z.reader == nil {
		return nil, z.err
	}

	files := make([]string, 0)
	dirNameWithSlash := strings.TrimSuffix(dirname, "/") + "/"

	for _, file := range z.reader.File {
		if strings.HasPrefix(file.Name, dirNameWithSlash) {
			files = append(files, file.Name)
		}
	}

	return files, nil
}

func (z *ZipPluginDecoder) FileReader(filename string) (io.ReadCloser, error) {
	return z.reader.Open(filename)
}

func (z *ZipPluginDecoder) decode() error {
	if z.reader == nil {
		return z.err
	}

	signatureData, err := parser.UnmarshalJson[struct {
		Signature string `json:"signature"`
		Time      int64  `json:"time"`
	}](z.reader.Comment)

	if err != nil {
		return err
	}

	pluginSig := signatureData.Signature
	pluginTime := signatureData.Time

	var verification *Verification

	// try to read the verification file
	verificationData, err := z.ReadFile(consts.VERIFICATION_FILE)
	if err != nil {
		if !errors.Is(err, os.ErrNotExist) {
			return err
		}

		// if the verification file is not found, set the verification to nil
		verification = nil
	} else {
		// unmarshal the verification data
		verificationData, err := parser.UnmarshalJsonBytes[Verification](verificationData)
		if err != nil {
			return err
		}

		verification = &verificationData
	}

	z.sig = pluginSig
	z.createTime = pluginTime
	z.verification = verification

	return nil
}

func (z *ZipPluginDecoder) Signature() (string, error) {
	if z.sig != "" {
		return z.sig, nil
	}

	if z.reader == nil {
		return "", z.err
	}

	err := z.decode()
	if err != nil {
		return "", err
	}

	return z.sig, nil
}

func (z *ZipPluginDecoder) CreateTime() (int64, error) {
	if z.createTime != 0 {
		return z.createTime, nil
	}

	if z.reader == nil {
		return 0, z.err
	}

	err := z.decode()
	if err != nil {
		return 0, err
	}

	return z.createTime, nil
}

func (z *ZipPluginDecoder) Verification() (*Verification, error) {
	if !z.Verified() {
		return nil, errors.New("plugin is not verified")
	}

	if z.verification != nil {
		return z.verification, nil
	}

	err := z.decode()
	if err != nil {
		return nil, err
	}

	// if the plugin is verified but the verification is nil
	// it's historical reason that all plugins are signed by langgenius
	return nil, nil
}

func (z *ZipPluginDecoder) Manifest() (plugin_entities.PluginDeclaration, error) {
	return z.PluginDecoderHelper.Manifest(z)
}

func (z *ZipPluginDecoder) Assets() (map[string][]byte, error) {
	// FIXES: https://github.com/langgenius/dify-plugin-daemon/issues/166
	// zip file is os-independent, `/` is the separator
	return z.PluginDecoderHelper.Assets(z, "/")
}

func (z *ZipPluginDecoder) Checksum() (string, error) {
	return z.PluginDecoderHelper.Checksum(z)
}

func (z *ZipPluginDecoder) UniqueIdentity() (plugin_entities.PluginUniqueIdentifier, error) {
	return z.PluginDecoderHelper.UniqueIdentity(z)
}

func (z *ZipPluginDecoder) ExtractTo(dst string) error {
	// copy to working directory
	if err := z.Walk(func(filename, dir string) error {
		workingPath := path.Join(dst, dir)
		// check if directory exists
		if err := os.MkdirAll(workingPath, 0755); err != nil {
			return err
		}

		bytes, err := z.ReadFile(filepath.Join(dir, filename))
		if err != nil {
			return err
		}

		filename = filepath.Join(workingPath, filename)

		// copy file
		if err := os.WriteFile(filename, bytes, 0644); err != nil {
			return err
		}

		return nil
	}); err != nil {
		// if error, delete the working directory
		os.RemoveAll(dst)
		return errors.Join(fmt.Errorf("copy plugin to working directory error: %v", err), err)
	}

	return nil
}

func (z *ZipPluginDecoder) CheckAssetsValid() error {
	return z.PluginDecoderHelper.CheckAssetsValid(z)
}

func (z *ZipPluginDecoder) Verified() bool {
	return z.PluginDecoderHelper.verified(z)
}

func (z *ZipPluginDecoder) AvailableI18nReadme() (map[string]string, error) {
	return z.PluginDecoderHelper.AvailableI18nReadme(z, "/")
}
