package decoder

import (
	"io"
	"io/fs"

	"github.com/langgenius/dify-plugin-daemon/pkg/entities/plugin_entities"
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

	// ReadDir reads the contents of a directory and returns a slice of strings
	// The strings are the filenames, it's a full path and directory will not be included
	// It executes recursively
	// Example:
	// - dirname: "config"
	// - return: ["config/settings.yaml", "config/default.yaml"]
	ReadDir(dirname string) ([]string, error)

	// Close releases any resources used by the decoder
	Close() error

	// Stat returns file info for the specified filename
	Stat(filename string) (fs.FileInfo, error)

	// FileReader returns an io.ReadCloser for reading the contents of a file
	// Remember to close the reader when done using it
	FileReader(filename string) (io.ReadCloser, error)

	// Signature returns the signature of the plugin, if available
	Signature() (string, error)

	// Verified returns true if the plugin is verified
	Verified() bool

	// Verification returns the verification of the plugin, if available
	// Error will only returns if the plugin is not verified
	Verification() (*Verification, error)

	// CreateTime returns the creation time of the plugin as a Unix timestamp
	CreateTime() (int64, error)

	// Manifest returns the manifest of the plugin
	Manifest() (plugin_entities.PluginDeclaration, error)

	// Assets returns a map of assets, the key is the filename, the value is the content
	Assets() (map[string][]byte, error)

	// UniqueIdentity returns the unique identity of the plugin
	UniqueIdentity() (plugin_entities.PluginUniqueIdentifier, error)

	// Checksum returns the checksum of the plugin
	Checksum() (string, error)

	// Check Assets valid
	CheckAssetsValid() error

	// AvailableI18nReadme returns a map of available readme i18n, the key is the language, the value is the readme.
	// The language code is in the format of IETF BCP 47 language tag
	AvailableI18nReadme() (map[string]string, error)
}
