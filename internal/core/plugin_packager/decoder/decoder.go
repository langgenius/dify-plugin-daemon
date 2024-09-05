package decoder

import (
	"errors"
	"fmt"
	"io"
	"io/fs"
	"strings"

	"github.com/langgenius/dify-plugin-daemon/internal/types/entities/plugin_entities"
	"github.com/langgenius/dify-plugin-daemon/internal/utils/parser"
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

	// CreateTime returns the creation time of the plugin as a Unix timestamp
	CreateTime() (int64, error)

	// Manifest returns the manifest of the plugin
	Manifest() (plugin_entities.PluginDeclaration, error)

	// Assets returns a map of assets, the key is the filename, the value is the content
	Assets() (map[string][]byte, error)
}

type PluginDecoderHelper struct {
	pluginDeclaration *plugin_entities.PluginDeclaration
}

func (p *PluginDecoderHelper) Manifest(decoder PluginDecoder) (plugin_entities.PluginDeclaration, error) {
	if p.pluginDeclaration != nil {
		return *p.pluginDeclaration, nil
	}

	// read the manifest file
	manifest, err := decoder.ReadFile("manifest.yaml")
	if err != nil {
		return plugin_entities.PluginDeclaration{}, err
	}

	dec, err := parser.UnmarshalYamlBytes[plugin_entities.PluginDeclaration](manifest)
	if err != nil {
		return plugin_entities.PluginDeclaration{}, err
	}

	// try to load plugins
	plugins := dec.Plugins
	for _, plugin := range plugins {
		// read yaml
		plugin_yaml, err := decoder.ReadFile(plugin)
		if err != nil {
			return plugin_entities.PluginDeclaration{}, errors.Join(err, fmt.Errorf("failed to read plugin file: %s", plugin))
		}

		plugin_dec, err := parser.UnmarshalYamlBytes[plugin_entities.GenericProviderDeclaration](plugin_yaml)
		if err != nil {
			return plugin_entities.PluginDeclaration{}, errors.Join(err, fmt.Errorf("failed to unmarshal plugin file: %s", plugin))
		}

		switch plugin_dec.Type {
		case plugin_entities.PROVIDER_TYPE_ENDPOINT:
			dec.Endpoint, err = parser.MapToStruct[plugin_entities.EndpointProviderDeclaration](plugin_dec.Provider)
			if err != nil {
				return plugin_entities.PluginDeclaration{}, errors.Join(err, fmt.Errorf("failed to convert endpoint to struct: %s", plugin))
			}
		case plugin_entities.PROVIDER_TYPE_TOOL:
			dec.Tool, err = parser.MapToStruct[plugin_entities.ToolProviderDeclaration](plugin_dec.Provider)
			if err != nil {
				return plugin_entities.PluginDeclaration{}, errors.Join(err, fmt.Errorf("failed to convert tool to struct: %s", plugin))
			}
		case plugin_entities.PROVIDER_TYPE_MODEL:
			dec.Model, err = parser.MapToStruct[plugin_entities.ModelProviderDeclaration](plugin_dec.Provider)
			if err != nil {
				return plugin_entities.PluginDeclaration{}, errors.Join(err, fmt.Errorf("failed to convert model to struct: %s", plugin))
			}
		default:
			return plugin_entities.PluginDeclaration{}, fmt.Errorf("unknown provider type: %s", plugin_dec.Type)
		}
	}

	if err := dec.ManifestValidate(); err != nil {
		return plugin_entities.PluginDeclaration{}, err
	}

	p.pluginDeclaration = &dec
	return dec, nil
}

func (p *PluginDecoderHelper) Assets(decoder PluginDecoder) (map[string][]byte, error) {
	files, err := decoder.ReadDir("_assets")
	if err != nil {
		return nil, err
	}

	assets := make(map[string][]byte)
	for _, file := range files {
		content, err := decoder.ReadFile(file)
		if err != nil {
			return nil, err
		}
		// trim _assets
		file, _ = strings.CutPrefix(file, "_assets/")
		assets[file] = content
	}

	return assets, nil
}
