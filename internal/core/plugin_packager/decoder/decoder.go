package decoder

import (
	"io"
	"io/fs"

	"github.com/langgenius/dify-plugin-daemon/internal/types/entities/plugin_entities"
)

// PluginDecoder is an interface for decoding and interacting with plugin files
type PluginDecoder interface {
	// Open initializes the decoder and prepares it for use
	Open() error

	// Walk traverses the plugin files and calls the provided function for each file
	// The function is called with the filename and directory of each file
	Walk(fn func(filename string, dir string) error) error

	// ReadFile reads the entire contents of a file and returns it as a byte slice
	ReadFile(filename string) ([]byte, error)

	// Close releases any resources used by the decoder
	Close() error

	// Stat returns file info for the specified filename
	Stat(filename string) (fs.FileInfo, error)

	// FileReader returns an io.ReadCloser for reading the contents of a file
	// Remember to close the reader when done using it
	FileReader(filename string) (io.ReadCloser, error)

	// Signature returns the signature of the plugin, if available
	Signature() (string, error)

	// CreateTime returns the creation time of the plugin as a Unix timestamp
	CreateTime() (int64, error)

	// Manifest returns the manifest of the plugin
	Manifest() (plugin_entities.PluginDeclaration, error)
}
