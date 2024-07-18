package decoder

import (
	"archive/zip"
	"bytes"
	"path"
)

type ZipPluginDecoder struct {
	PluginDecoder

	reader *zip.Reader
	err    error
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

	return decoder, nil
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
