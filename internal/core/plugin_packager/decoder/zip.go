package decoder

import (
	"archive/zip"
	"bytes"
	"io"
	"io/fs"
	"path"

	"github.com/langgenius/dify-plugin-daemon/internal/types/entities/plugin_entities"
	"github.com/langgenius/dify-plugin-daemon/internal/utils/parser"
)

type ZipPluginDecoder struct {
	PluginDecoder
	PluginDecoderHelper

	reader *zip.Reader
	err    error

	sig         string
	create_time int64

	pluginDeclaration *plugin_entities.PluginDeclaration
}

func NewZipPluginDecoder(binary []byte) (*ZipPluginDecoder, error) {
	reader, err := zip.NewReader(bytes.NewReader(binary), int64(len(binary)))

	decoder := &ZipPluginDecoder{
		reader: reader,
		err:    err,
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

func (z *ZipPluginDecoder) FileReader(filename string) (io.ReadCloser, error) {
	return z.reader.Open(filename)
}

func (z *ZipPluginDecoder) decode() error {
	if z.reader == nil {
		return z.err
	}

	type signatureData struct {
		Signature string `json:"signature"`
		Time      int64  `json:"time"`
	}

	signature_data, err := parser.UnmarshalJson[signatureData](z.reader.Comment)
	if err != nil {
		return err
	}

	plugin_sig := signature_data.Signature
	plugin_time := signature_data.Time

	z.sig = plugin_sig
	z.create_time = plugin_time

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
	if z.create_time != 0 {
		return z.create_time, nil
	}

	if z.reader == nil {
		return 0, z.err
	}

	err := z.decode()
	if err != nil {
		return 0, err
	}

	return z.create_time, nil
}

func (z *ZipPluginDecoder) Manifest() (plugin_entities.PluginDeclaration, error) {
	return z.PluginDecoderHelper.Manifest(z)
}
