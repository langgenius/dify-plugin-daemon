package serverless

import (
	"archive/tar"
	"compress/gzip"
	"errors"
	"io"
	"io/fs"
	"os"
	"path"
	"strings"
	"time"

	"github.com/langgenius/dify-plugin-daemon/internal/core/plugin_manager/aws_manager/dockerfile"
	"github.com/langgenius/dify-plugin-daemon/internal/core/plugin_packager/decoder"
	"github.com/langgenius/dify-plugin-daemon/internal/utils/tmpfile"
)

type Packager struct {
	decoder decoder.PluginDecoder
}

func NewPackager(decoder decoder.PluginDecoder) *Packager {
	return &Packager{
		decoder: decoder,
	}
}

type dockerFileInfo struct {
	fs.FileInfo

	size int64
}

func (d *dockerFileInfo) Size() int64 {
	return d.size
}

func (d *dockerFileInfo) Name() string {
	return "Dockerfile"
}

func (d *dockerFileInfo) Mode() os.FileMode {
	return 0644
}

func (d *dockerFileInfo) ModTime() time.Time {
	return time.Now()
}

func (d *dockerFileInfo) IsDir() bool {
	return false
}

func (d *dockerFileInfo) Sys() any {
	return nil
}

// Pack takes a plugin and packs it into a tar file with dockerfile inside
// returns a *os.File with the tar file
func (p *Packager) Pack() (*os.File, error) {
	declaration, err := p.decoder.Manifest()
	if err != nil {
		return nil, err
	}

	// walk through the plugin directory and add it to a tar file
	// create a tmpfile
	tmpfile, cleanup, err := tmpfile.CreateTempFile("plugin-aws-tar-*")
	if err != nil {
		return nil, err
	}
	success := false

	defer func() {
		if !success {
			cleanup()
		}
	}()

	gzipWriter, err := gzip.NewWriterLevel(tmpfile, gzip.BestCompression)
	if err != nil {
		return nil, err
	}
	defer gzipWriter.Close()

	tarWriter := tar.NewWriter(gzipWriter)
	defer tarWriter.Close()

	if err := p.decoder.Walk(func(filename, dir string) error {
		if strings.ToLower(filename) == "dockerfile" {
			return errors.New("dockerfile is not allowed to be in the plugin directory")
		}

		fullFilename := path.Join(dir, filename)

		state, err := p.decoder.Stat(fullFilename)
		if err != nil {
			return err
		}

		tarHeader, err := tar.FileInfoHeader(state, fullFilename)
		if err != nil {
			return err
		}
		tarHeader.Name = fullFilename

		// write tar header
		if err := tarWriter.WriteHeader(tarHeader); err != nil {
			return err
		}

		// write file content
		fileReader, err := p.decoder.FileReader(fullFilename)
		if err != nil {
			return err
		}
		if _, err := io.Copy(tarWriter, fileReader); err != nil {
			fileReader.Close()
			return err
		}
		// release resources
		fileReader.Close()

		return nil
	}); err != nil {
		return nil, err
	}

	// add dockerfile
	dockerfile, err := dockerfile.GenerateDockerfile(&declaration)
	if err != nil {
		return nil, err
	}

	tarHeader, err := tar.FileInfoHeader(&dockerFileInfo{
		size: int64(len(dockerfile)),
	}, "Dockerfile")
	if err != nil {
		return nil, err
	}

	// create a fake dockerfile stat
	if err := tarWriter.WriteHeader(tarHeader); err != nil {
		return nil, err
	}

	if _, err := tarWriter.Write([]byte(dockerfile)); err != nil {
		return nil, err
	}

	// close writers to flush data
	tarWriter.Close()
	gzipWriter.Close()

	tmpfile.Seek(0, io.SeekStart)

	success = true

	return tmpfile, nil
}
