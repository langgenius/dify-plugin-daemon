package aws_manager

import (
	"archive/tar"
	"compress/gzip"
	"errors"
	"fmt"
	"io"
	"io/fs"
	"os"
	"path"
	"strings"
	"time"

	"github.com/langgenius/dify-plugin-daemon/internal/core/plugin_manager/aws_manager/dockerfile"
	"github.com/langgenius/dify-plugin-daemon/internal/core/plugin_packager/decoder"
	"github.com/langgenius/dify-plugin-daemon/internal/types/entities/plugin_entities"
	"github.com/langgenius/dify-plugin-daemon/internal/utils/tmpfile"
)

type Packager struct {
	runtime plugin_entities.PluginRuntimeInterface
	decoder decoder.PluginDecoder
}

func NewPackager(runtime plugin_entities.PluginRuntimeInterface, decoder decoder.PluginDecoder) *Packager {
	return &Packager{
		runtime: runtime,
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

	gzip_writer, err := gzip.NewWriterLevel(tmpfile, gzip.BestCompression)
	if err != nil {
		return nil, err
	}
	defer gzip_writer.Close()

	tar_writer := tar.NewWriter(gzip_writer)
	defer tar_writer.Close()

	if err := p.decoder.Walk(func(filename, dir string) error {
		if strings.ToLower(filename) == "dockerfile" {
			return errors.New("dockerfile is not allowed to be in the plugin directory")
		}

		full_filename := path.Join(dir, filename)

		state, err := p.decoder.Stat(full_filename)
		if err != nil {
			return err
		}

		if state.Size() > 1024*1024*10 {
			// 10MB, 1 single file is too large
			return fmt.Errorf("file size is too large: %s, max 10MB", full_filename)
		}

		tar_header, err := tar.FileInfoHeader(state, full_filename)
		if err != nil {
			return err
		}
		tar_header.Name = filename

		// write tar header
		if err := tar_writer.WriteHeader(tar_header); err != nil {
			return err
		}

		// write file content
		file_reader, err := p.decoder.FileReader(full_filename)
		if err != nil {
			return err
		}
		if _, err := io.Copy(tar_writer, file_reader); err != nil {
			file_reader.Close()
			return err
		}
		// release resources
		file_reader.Close()

		return nil
	}); err != nil {
		return nil, err
	}

	// add dockerfile
	dockerfile, err := dockerfile.GenerateDockerfile(p.runtime.Configuration())
	if err != nil {
		return nil, err
	}

	tar_header, err := tar.FileInfoHeader(&dockerFileInfo{
		size: int64(len(dockerfile)),
	}, "Dockerfile")
	if err != nil {
		return nil, err
	}

	// create a fake dockerfile stat
	if err := tar_writer.WriteHeader(tar_header); err != nil {
		return nil, err
	}

	if _, err := tar_writer.Write([]byte(dockerfile)); err != nil {
		return nil, err
	}

	// close writers to flush data
	tar_writer.Close()
	gzip_writer.Close()

	tmpfile.Seek(0, io.SeekStart)

	success = true

	return tmpfile, nil
}
