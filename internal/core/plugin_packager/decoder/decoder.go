package decoder

import (
	"errors"
	"fmt"
	"io"
	"io/fs"
	"path/filepath"
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

		plugin_dec, err := parser.UnmarshalYamlBytes[plugin_entities.ToolProviderDeclaration](plugin_yaml)
		if err != nil {
			return plugin_entities.PluginDeclaration{}, errors.Join(err, fmt.Errorf("failed to unmarshal plugin file: %s", tool))
		}

		// read tools
		for _, tool_file := range plugin_dec.ToolFiles {
			tool_file_content, err := decoder.ReadFile(tool_file)
			if err != nil {
				return plugin_entities.PluginDeclaration{}, errors.Join(err, fmt.Errorf("failed to read tool file: %s", tool_file))
			}

			tool_file_dec, err := parser.UnmarshalYamlBytes[plugin_entities.ToolDeclaration](tool_file_content)
			if err != nil {
				return plugin_entities.PluginDeclaration{}, errors.Join(err, fmt.Errorf("failed to unmarshal tool file: %s", tool_file))
			}

			plugin_dec.Tools = append(plugin_dec.Tools, tool_file_dec)
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

		// read detailed endpoints
		endpoints_files := plugin_dec.EndpointFiles

		for _, endpoint_file := range endpoints_files {
			endpoint_file_content, err := decoder.ReadFile(endpoint_file)
			if err != nil {
				return plugin_entities.PluginDeclaration{}, errors.Join(err, fmt.Errorf("failed to read endpoint file: %s", endpoint_file))
			}

			endpoint_file_dec, err := parser.UnmarshalYamlBytes[plugin_entities.EndpointDeclaration](endpoint_file_content)
			if err != nil {
				return plugin_entities.PluginDeclaration{}, errors.Join(err, fmt.Errorf("failed to unmarshal endpoint file: %s", endpoint_file))
			}

			plugin_dec.Endpoints = append(plugin_dec.Endpoints, endpoint_file_dec)
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

		// read model position file
		if plugin_dec.PositionFiles != nil {
			plugin_dec.Position = &plugin_entities.ModelPosition{}

			llm_file_name, ok := plugin_dec.PositionFiles["llm"]
			if ok {
				llm_file, err := decoder.ReadFile(llm_file_name)
				if err != nil {
					return plugin_entities.PluginDeclaration{}, errors.Join(err, fmt.Errorf("failed to read llm position file: %s", llm_file_name))
				}

				position, err := parser.UnmarshalYamlBytes[[]string](llm_file)
				if err != nil {
					return plugin_entities.PluginDeclaration{}, errors.Join(err, fmt.Errorf("failed to unmarshal llm position file: %s", llm_file_name))
				}

				plugin_dec.Position.LLM = &position
			}

			text_embedding_file_name, ok := plugin_dec.PositionFiles["text_embedding"]
			if ok {
				text_embedding_file, err := decoder.ReadFile(text_embedding_file_name)
				if err != nil {
					return plugin_entities.PluginDeclaration{}, errors.Join(err, fmt.Errorf("failed to read text embedding position file: %s", text_embedding_file_name))
				}

				position, err := parser.UnmarshalYamlBytes[[]string](text_embedding_file)
				if err != nil {
					return plugin_entities.PluginDeclaration{}, errors.Join(err, fmt.Errorf("failed to unmarshal text embedding position file: %s", text_embedding_file_name))
				}

				plugin_dec.Position.TextEmbedding = &position
			}

			rerank_file_name, ok := plugin_dec.PositionFiles["rerank"]
			if ok {
				rerank_file, err := decoder.ReadFile(rerank_file_name)
				if err != nil {
					return plugin_entities.PluginDeclaration{}, errors.Join(err, fmt.Errorf("failed to read rerank position file: %s", rerank_file_name))
				}

				position, err := parser.UnmarshalYamlBytes[[]string](rerank_file)
				if err != nil {
					return plugin_entities.PluginDeclaration{}, errors.Join(err, fmt.Errorf("failed to unmarshal rerank position file: %s", rerank_file_name))
				}

				plugin_dec.Position.Rerank = &position
			}

			tts_file_name, ok := plugin_dec.PositionFiles["tts"]
			if ok {
				tts_file, err := decoder.ReadFile(tts_file_name)
				if err != nil {
					return plugin_entities.PluginDeclaration{}, errors.Join(err, fmt.Errorf("failed to read tts position file: %s", tts_file_name))
				}

				position, err := parser.UnmarshalYamlBytes[[]string](tts_file)
				if err != nil {
					return plugin_entities.PluginDeclaration{}, errors.Join(err, fmt.Errorf("failed to unmarshal tts position file: %s", tts_file_name))
				}

				plugin_dec.Position.TTS = &position
			}

			speech2text_file_name, ok := plugin_dec.PositionFiles["speech2text"]
			if ok {
				speech2text_file, err := decoder.ReadFile(speech2text_file_name)
				if err != nil {
					return plugin_entities.PluginDeclaration{}, errors.Join(err, fmt.Errorf("failed to read speech2text position file: %s", speech2text_file_name))
				}

				position, err := parser.UnmarshalYamlBytes[[]string](speech2text_file)
				if err != nil {
					return plugin_entities.PluginDeclaration{}, errors.Join(err, fmt.Errorf("failed to unmarshal speech2text position file: %s", speech2text_file_name))
				}

				plugin_dec.Position.Speech2text = &position
			}

			moderation_file_name, ok := plugin_dec.PositionFiles["moderation"]
			if ok {
				moderation_file, err := decoder.ReadFile(moderation_file_name)
				if err != nil {
					return plugin_entities.PluginDeclaration{}, errors.Join(err, fmt.Errorf("failed to read moderation position file: %s", moderation_file_name))
				}

				position, err := parser.UnmarshalYamlBytes[[]string](moderation_file)
				if err != nil {
					return plugin_entities.PluginDeclaration{}, errors.Join(err, fmt.Errorf("failed to unmarshal moderation position file: %s", moderation_file_name))
				}

				plugin_dec.Position.Moderation = &position
			}
		}

		// read models
		if err := decoder.Walk(func(filename, dir string) error {
			model_patterns := plugin_dec.ModelFiles
			// using glob to match if dir/filename is in models
			model_file_name := filepath.Join(dir, filename)
			if strings.HasSuffix(model_file_name, "_position.yaml") {
				return nil
			}

			for _, model_pattern := range model_patterns {
				matched, err := filepath.Match(model_pattern, model_file_name)
				if err != nil {
					return err
				}
				if matched {
					// read model file
					model_file, err := decoder.ReadFile(model_file_name)
					if err != nil {
						return err
					}

					model_dec, err := parser.UnmarshalYamlBytes[plugin_entities.ModelDeclaration](model_file)
					if err != nil {
						return err
					}

					plugin_dec.Models = append(plugin_dec.Models, model_dec)
				}
			}

			return nil
		}); err != nil {
			return plugin_entities.PluginDeclaration{}, err
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
