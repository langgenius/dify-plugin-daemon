package bundle_packager

import (
	"archive/zip"
	"bytes"
	"errors"
	"io"
	"os"
	"strconv"
	"strings"
)

type ZipBundlePackager struct {
	*MemoryZipBundlePackager

	path string
}

func NewZipBundlePackager(path string) (BundlePackager, error) {
	zipFile, err := os.Open(path)
	if err != nil {
		return nil, err
	}
	defer zipFile.Close()

	zipBytes, err := io.ReadAll(zipFile)
	if err != nil {
		return nil, err
	}

	memoryPackager, err := NewMemoryZipBundlePackager(zipBytes)
	if err != nil {
		return nil, err
	}

	zipBundlePackager := &ZipBundlePackager{
		MemoryZipBundlePackager: memoryPackager,
		path:                    path,
	}

	return zipBundlePackager, nil
}

func NewZipBundlePackagerWithSizeLimit(path string, maxSize int64) (BundlePackager, error) {
	zipFile, err := os.Open(path)
	if err != nil {
		return nil, err
	}
	defer zipFile.Close()

	zipBytes, err := io.ReadAll(zipFile)
	if err != nil {
		return nil, err
	}

	reader, err := zip.NewReader(bytes.NewReader(zipBytes), int64(len(zipBytes)))
	if err != nil {
		return nil, errors.New(strings.ReplaceAll(err.Error(), "zip", "difypkg"))
	}

	totalSize := int64(0)
	for _, file := range reader.File {
		totalSize += int64(file.UncompressedSize64)
		if totalSize > maxSize {
			return nil, errors.New(
				"bundle package size is too large, please ensure the uncompressed size is less than " +
					strconv.FormatInt(maxSize, 10) + " bytes",
			)
		}
	}

	memoryPackager, err := NewMemoryZipBundlePackager(zipBytes)
	if err != nil {
		return nil, err
	}

	zipBundlePackager := &ZipBundlePackager{
		MemoryZipBundlePackager: memoryPackager,
		path:                    path,
	}

	return zipBundlePackager, nil
}

func (p *ZipBundlePackager) Save() error {
	// export the bundle to a zip file
	zipBytes, err := p.Export()
	if err != nil {
		return err
	}

	// save the zip file
	err = os.WriteFile(p.path, zipBytes, 0644)
	if err != nil {
		return err
	}

	// reload zip reader
	zipFile, err := os.Open(p.path)
	if err != nil {
		return err
	}
	defer zipFile.Close()

	zipFileInfo, err := zipFile.Stat()
	if err != nil {
		return err
	}

	p.zipReader, err = zip.NewReader(zipFile, zipFileInfo.Size())
	if err != nil {
		return err
	}

	return nil
}
