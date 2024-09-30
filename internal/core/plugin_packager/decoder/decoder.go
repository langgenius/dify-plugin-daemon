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

	// UniqueIdentity returns the unique identity of the plugin
	UniqueIdentity() (plugin_entities.PluginUniqueIdentifier, error)

	// Checksum returns the checksum of the plugin
	Checksum() (string, error)
}

type PluginDecoderHelper struct {
	pluginDeclaration *plugin_entities.PluginDeclaration
	checksum          string
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
	for _, tool := range plugins.Tools {
		// read yaml
		plugin_yaml, err := decoder.ReadFile(tool)
		if err != nil {
			return plugin_entities.PluginDeclaration{}, errors.Join(err, fmt.Errorf("failed to read tool file: %s", tool))
		}

		// TODO

		plugin_dec, err := parser.UnmarshalYamlBytes[plugin_entities.ToolProviderDeclaration](plugin_yaml)
		if err != nil {
			return plugin_entities.PluginDeclaration{}, errors.Join(err, fmt.Errorf("failed to unmarshal plugin file: %s", tool))
		}

		dec.Tool = &plugin_dec
	}

	for _, endpoint := range plugins.Endpoints {
		// read yaml
		plugin_yaml, err := decoder.ReadFile(endpoint)
		if err != nil {
			return plugin_entities.PluginDeclaration{}, errors.Join(err, fmt.Errorf("failed to read endpoint file: %s", endpoint))
		}

		plugin_dec, err := parser.UnmarshalYamlBytes[plugin_entities.EndpointProviderDeclaration](plugin_yaml)
		if err != nil {
			return plugin_entities.PluginDeclaration{}, errors.Join(err, fmt.Errorf("failed to unmarshal plugin file: %s", endpoint))
		}

		dec.Endpoint = &plugin_dec
	}

	for _, model := range plugins.Models {
		// read yaml
		plugin_yaml, err := decoder.ReadFile(model)
		if err != nil {
			return plugin_entities.PluginDeclaration{}, errors.Join(err, fmt.Errorf("failed to read model file: %s", model))
		}

		plugin_dec, err := parser.UnmarshalYamlBytes[plugin_entities.ModelProviderDeclaration](plugin_yaml)
		if err != nil {
			return plugin_entities.PluginDeclaration{}, errors.Join(err, fmt.Errorf("failed to unmarshal plugin file: %s", model))
		}

		dec.Model = &plugin_dec
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

func (p *PluginDecoderHelper) Checksum(decoder PluginDecoder) (string, error) {
	if p.checksum != "" {
		return p.checksum, nil
	}

	var err error

	p.checksum, err = CalculateChecksum(decoder)
	if err != nil {
		return "", err
	}

	return p.checksum, nil
}

func (p *PluginDecoderHelper) UniqueIdentity(decoder PluginDecoder) (plugin_entities.PluginUniqueIdentifier, error) {
	manifest, err := decoder.Manifest()
	if err != nil {
		return plugin_entities.PluginUniqueIdentifier(""), err
	}

	identity := manifest.Identity()
	checksum, err := decoder.Checksum()
	if err != nil {
		return plugin_entities.PluginUniqueIdentifier(""), err
	}

	return plugin_entities.NewPluginUniqueIdentifier(fmt.Sprintf("%s@%s", identity, checksum))
}
