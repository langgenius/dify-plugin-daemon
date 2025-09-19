package decoder

import (
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"strings"

	"github.com/langgenius/dify-plugin-daemon/internal/utils/parser"
	"github.com/langgenius/dify-plugin-daemon/pkg/entities/plugin_entities"
)

type PluginDecoderHelper struct {
	pluginDeclaration *plugin_entities.PluginDeclaration
	checksum          string

	verifiedFlag *bool // used to store the verified flag, avoid calling verified function multiple times
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
		pluginYaml, err := decoder.ReadFile(tool)
		if err != nil {
			return plugin_entities.PluginDeclaration{}, errors.Join(err, fmt.Errorf("failed to read tool file: %s", tool))
		}

		pluginDec, err := parser.UnmarshalYamlBytes[plugin_entities.ToolProviderDeclaration](pluginYaml)
		if err != nil {
			return plugin_entities.PluginDeclaration{}, errors.Join(err, fmt.Errorf("failed to unmarshal plugin file: %s", tool))
		}

		// read tools
		for _, tool_file := range pluginDec.ToolFiles {
			toolFileContent, err := decoder.ReadFile(tool_file)
			if err != nil {
				return plugin_entities.PluginDeclaration{}, errors.Join(err, fmt.Errorf("failed to read tool file: %s", tool_file))
			}

			toolFileDec, err := parser.UnmarshalYamlBytes[plugin_entities.ToolDeclaration](toolFileContent)
			if err != nil {
				return plugin_entities.PluginDeclaration{}, errors.Join(err, fmt.Errorf("failed to unmarshal tool file: %s", tool_file))
			}

			pluginDec.Tools = append(pluginDec.Tools, toolFileDec)
		}

		dec.Tool = &pluginDec
	}

	for _, endpoint := range plugins.Endpoints {
		// read yaml
		pluginYaml, err := decoder.ReadFile(endpoint)
		if err != nil {
			return plugin_entities.PluginDeclaration{}, errors.Join(err, fmt.Errorf("failed to read endpoint file: %s", endpoint))
		}

		pluginDec, err := parser.UnmarshalYamlBytes[plugin_entities.EndpointProviderDeclaration](pluginYaml)
		if err != nil {
			return plugin_entities.PluginDeclaration{}, errors.Join(err, fmt.Errorf("failed to unmarshal plugin file: %s", endpoint))
		}

		// read detailed endpoints
		endpointsFiles := pluginDec.EndpointFiles

		for _, endpoint_file := range endpointsFiles {
			endpointFileContent, err := decoder.ReadFile(endpoint_file)
			if err != nil {
				return plugin_entities.PluginDeclaration{}, errors.Join(err, fmt.Errorf("failed to read endpoint file: %s", endpoint_file))
			}

			endpointFileDec, err := parser.UnmarshalYamlBytes[plugin_entities.EndpointDeclaration](endpointFileContent)
			if err != nil {
				return plugin_entities.PluginDeclaration{}, errors.Join(err, fmt.Errorf("failed to unmarshal endpoint file: %s", endpoint_file))
			}

			pluginDec.Endpoints = append(pluginDec.Endpoints, endpointFileDec)
		}

		dec.Endpoint = &pluginDec
	}

	for _, model := range plugins.Models {
		// read yaml
		pluginYaml, err := decoder.ReadFile(model)
		if err != nil {
			return plugin_entities.PluginDeclaration{}, errors.Join(err, fmt.Errorf("failed to read model file: %s", model))
		}

		pluginDec, err := parser.UnmarshalYamlBytes[plugin_entities.ModelProviderDeclaration](pluginYaml)
		if err != nil {
			return plugin_entities.PluginDeclaration{}, errors.Join(err, fmt.Errorf("failed to unmarshal plugin file: %s", model))
		}

		// read model position file
		if pluginDec.PositionFiles != nil {
			pluginDec.Position = &plugin_entities.ModelPosition{}

			llmFileName, ok := pluginDec.PositionFiles["llm"]
			if ok {
				llmFile, err := decoder.ReadFile(llmFileName)
				if err != nil {
					return plugin_entities.PluginDeclaration{}, errors.Join(err, fmt.Errorf("failed to read llm position file: %s", llmFileName))
				}

				position, err := parser.UnmarshalYamlBytes[[]string](llmFile)
				if err != nil {
					return plugin_entities.PluginDeclaration{}, errors.Join(err, fmt.Errorf("failed to unmarshal llm position file: %s", llmFileName))
				}

				pluginDec.Position.LLM = &position
			}

			textEmbeddingFileName, ok := pluginDec.PositionFiles["text_embedding"]
			if ok {
				textEmbeddingFile, err := decoder.ReadFile(textEmbeddingFileName)
				if err != nil {
					return plugin_entities.PluginDeclaration{}, errors.Join(err, fmt.Errorf("failed to read text embedding position file: %s", textEmbeddingFileName))
				}

				position, err := parser.UnmarshalYamlBytes[[]string](textEmbeddingFile)
				if err != nil {
					return plugin_entities.PluginDeclaration{}, errors.Join(err, fmt.Errorf("failed to unmarshal text embedding position file: %s", textEmbeddingFileName))
				}

				pluginDec.Position.TextEmbedding = &position
			}

			rerankFileName, ok := pluginDec.PositionFiles["rerank"]
			if ok {
				rerankFile, err := decoder.ReadFile(rerankFileName)
				if err != nil {
					return plugin_entities.PluginDeclaration{}, errors.Join(err, fmt.Errorf("failed to read rerank position file: %s", rerankFileName))
				}

				position, err := parser.UnmarshalYamlBytes[[]string](rerankFile)
				if err != nil {
					return plugin_entities.PluginDeclaration{}, errors.Join(err, fmt.Errorf("failed to unmarshal rerank position file: %s", rerankFileName))
				}

				pluginDec.Position.Rerank = &position
			}

			ttsFileName, ok := pluginDec.PositionFiles["tts"]
			if ok {
				ttsFile, err := decoder.ReadFile(ttsFileName)
				if err != nil {
					return plugin_entities.PluginDeclaration{}, errors.Join(err, fmt.Errorf("failed to read tts position file: %s", ttsFileName))
				}

				position, err := parser.UnmarshalYamlBytes[[]string](ttsFile)
				if err != nil {
					return plugin_entities.PluginDeclaration{}, errors.Join(err, fmt.Errorf("failed to unmarshal tts position file: %s", ttsFileName))
				}

				pluginDec.Position.TTS = &position
			}

			speech2textFileName, ok := pluginDec.PositionFiles["speech2text"]
			if ok {
				speech2textFile, err := decoder.ReadFile(speech2textFileName)
				if err != nil {
					return plugin_entities.PluginDeclaration{}, errors.Join(err, fmt.Errorf("failed to read speech2text position file: %s", speech2textFileName))
				}

				position, err := parser.UnmarshalYamlBytes[[]string](speech2textFile)
				if err != nil {
					return plugin_entities.PluginDeclaration{}, errors.Join(err, fmt.Errorf("failed to unmarshal speech2text position file: %s", speech2textFileName))
				}

				pluginDec.Position.Speech2text = &position
			}

			moderationFileName, ok := pluginDec.PositionFiles["moderation"]
			if ok {
				moderationFile, err := decoder.ReadFile(moderationFileName)
				if err != nil {
					return plugin_entities.PluginDeclaration{}, errors.Join(err, fmt.Errorf("failed to read moderation position file: %s", moderationFileName))
				}

				position, err := parser.UnmarshalYamlBytes[[]string](moderationFile)
				if err != nil {
					return plugin_entities.PluginDeclaration{}, errors.Join(err, fmt.Errorf("failed to unmarshal moderation position file: %s", moderationFileName))
				}

				pluginDec.Position.Moderation = &position
			}
		}

		// read models
		if err := decoder.Walk(func(filename, dir string) error {
			modelPatterns := pluginDec.ModelFiles
			// using glob to match if dir/filename is in models
			modelFileName := filepath.Join(dir, filename)
			if strings.HasSuffix(modelFileName, "_position.yaml") {
				return nil
			}

			for _, model_pattern := range modelPatterns {
				matched, err := filepath.Match(model_pattern, modelFileName)
				if err != nil {
					return err
				}
				if matched {
					// read model file
					modelFile, err := decoder.ReadFile(modelFileName)
					if err != nil {
						return err
					}

					modelDec, err := parser.UnmarshalYamlBytes[plugin_entities.ModelDeclaration](modelFile)
					if err != nil {
						return err
					}

					pluginDec.Models = append(pluginDec.Models, modelDec)
				}
			}

			return nil
		}); err != nil {
			return plugin_entities.PluginDeclaration{}, err
		}

		dec.Model = &pluginDec
	}

	for _, agentStrategy := range plugins.AgentStrategies {
		// read yaml
		pluginYaml, err := decoder.ReadFile(agentStrategy)
		if err != nil {
			return plugin_entities.PluginDeclaration{}, errors.Join(err, fmt.Errorf("failed to read agent strategy file: %s", agentStrategy))
		}

		pluginDec, err := parser.UnmarshalYamlBytes[plugin_entities.AgentStrategyProviderDeclaration](pluginYaml)
		if err != nil {
			return plugin_entities.PluginDeclaration{}, errors.Join(err, fmt.Errorf("failed to unmarshal plugin file: %s", agentStrategy))
		}

		for _, strategyFile := range pluginDec.StrategyFiles {
			strategyFileContent, err := decoder.ReadFile(strategyFile)
			if err != nil {
				return plugin_entities.PluginDeclaration{}, errors.Join(err, fmt.Errorf("failed to read agent strategy file: %s", strategyFile))
			}

			strategyDec, err := parser.UnmarshalYamlBytes[plugin_entities.AgentStrategyDeclaration](strategyFileContent)
			if err != nil {
				return plugin_entities.PluginDeclaration{}, errors.Join(err, fmt.Errorf("failed to unmarshal agent strategy file: %s", strategyFile))
			}

			pluginDec.Strategies = append(pluginDec.Strategies, strategyDec)
		}

		dec.AgentStrategy = &pluginDec
	}

	for _, datasource := range plugins.Datasources {
		// read yaml
		pluginYaml, err := decoder.ReadFile(datasource)
		if err != nil {
			return plugin_entities.PluginDeclaration{}, errors.Join(err, fmt.Errorf("failed to read datasource file: %s", datasource))
		}

		pluginDec, err := parser.UnmarshalYamlBytes[plugin_entities.DatasourceProviderDeclaration](pluginYaml)
		if err != nil {
			return plugin_entities.PluginDeclaration{}, errors.Join(err, fmt.Errorf("failed to unmarshal plugin file: %s", datasource))
		}

		for _, datasourceFile := range pluginDec.DatasourceFiles {
			datasourceFileContent, err := decoder.ReadFile(datasourceFile)
			if err != nil {
				return plugin_entities.PluginDeclaration{}, errors.Join(err, fmt.Errorf("failed to read datasource file: %s", datasourceFile))
			}

			datasourceDec, err := parser.UnmarshalYamlBytes[plugin_entities.DatasourceDeclaration](datasourceFileContent)
			if err != nil {
				return plugin_entities.PluginDeclaration{}, errors.Join(err, fmt.Errorf("failed to unmarshal datasource file: %s", datasourceFile))
			}

			pluginDec.Datasources = append(pluginDec.Datasources, datasourceDec)
		}

		dec.Datasource = &pluginDec
	}

	dec.FillInDefaultValues()

	dec.Verified = p.verified(decoder)

	p.pluginDeclaration = &dec
	return dec, nil
}

func (p *PluginDecoderHelper) Assets(decoder PluginDecoder, separator string) (map[string][]byte, error) {
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
		file, _ = strings.CutPrefix(file, "_assets"+separator)
		assets[file] = content
	}

	return assets, nil
}

func (p *PluginDecoderHelper) Checksum(decoder_instance PluginDecoder) (string, error) {
	if p.checksum != "" {
		return p.checksum, nil
	}

	var err error

	p.checksum, err = CalculateChecksum(decoder_instance)
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

func (p *PluginDecoderHelper) CheckAssetsValid(decoder PluginDecoder) error {
	declaration, err := decoder.Manifest()
	if err != nil {
		return errors.Join(err, fmt.Errorf("failed to get manifest"))
	}

	assets, err := decoder.Assets()
	if err != nil {
		return errors.Join(err, fmt.Errorf("failed to get assets"))
	}

	if declaration.Model != nil {
		if declaration.Model.IconSmall != nil {
			if declaration.Model.IconSmall.EnUS != "" {
				if _, ok := assets[declaration.Model.IconSmall.EnUS]; !ok {
					return errors.Join(err, fmt.Errorf("model icon small en_US not found"))
				}
			}

			if declaration.Model.IconSmall.ZhHans != "" {
				if _, ok := assets[declaration.Model.IconSmall.ZhHans]; !ok {
					return errors.Join(err, fmt.Errorf("model icon small zh_Hans not found"))
				}
			}

			if declaration.Model.IconSmall.JaJp != "" {
				if _, ok := assets[declaration.Model.IconSmall.JaJp]; !ok {
					return errors.Join(err, fmt.Errorf("model icon small ja_JP not found"))
				}
			}

			if declaration.Model.IconSmall.PtBr != "" {
				if _, ok := assets[declaration.Model.IconSmall.PtBr]; !ok {
					return errors.Join(err, fmt.Errorf("model icon small pt_BR not found"))
				}
			}
		}

		if declaration.Model.IconLarge != nil {
			if declaration.Model.IconLarge.EnUS != "" {
				if _, ok := assets[declaration.Model.IconLarge.EnUS]; !ok {
					return errors.Join(err, fmt.Errorf("model icon large en_US not found"))
				}
			}

			if declaration.Model.IconLarge.ZhHans != "" {
				if _, ok := assets[declaration.Model.IconLarge.ZhHans]; !ok {
					return errors.Join(err, fmt.Errorf("model icon large zh_Hans not found"))
				}
			}

			if declaration.Model.IconLarge.JaJp != "" {
				if _, ok := assets[declaration.Model.IconLarge.JaJp]; !ok {
					return errors.Join(err, fmt.Errorf("model icon large ja_JP not found"))
				}
			}

			if declaration.Model.IconLarge.PtBr != "" {
				if _, ok := assets[declaration.Model.IconLarge.PtBr]; !ok {
					return errors.Join(err, fmt.Errorf("model icon large pt_BR not found"))
				}
			}
		}
	}

	if declaration.Tool != nil {
		if declaration.Tool.Identity.Icon != "" {
			if _, ok := assets[declaration.Tool.Identity.Icon]; !ok {
				return errors.Join(err, fmt.Errorf("tool icon not found"))
			}
		}
	}

	if declaration.Icon != "" {
		if _, ok := assets[declaration.Icon]; !ok {
			return errors.Join(err, fmt.Errorf("plugin icon not found"))
		}
	}

	if declaration.IconDark != "" {
		if _, ok := assets[declaration.IconDark]; !ok {
			return errors.Join(err, fmt.Errorf("plugin dark icon not found"))
		}
	}

	return nil
}

func (p *PluginDecoderHelper) verified(decoder PluginDecoder) bool {
	if p.verifiedFlag != nil {
		return *p.verifiedFlag
	}

	// verify signature
	// for ZipPluginDecoder, use the third party signature verification if it is enabled
	if zipDecoder, ok := decoder.(*ZipPluginDecoder); ok {
		config := zipDecoder.thirdPartySignatureVerificationConfig
		if config != nil && config.Enabled && len(config.PublicKeyPaths) > 0 {
			verified := VerifyPluginWithPublicKeyPaths(decoder, config.PublicKeyPaths) == nil
			p.verifiedFlag = &verified
			return verified
		} else {
			verified := VerifyPlugin(decoder) == nil
			p.verifiedFlag = &verified
			return verified
		}
	} else {
		verified := VerifyPlugin(decoder) == nil
		p.verifiedFlag = &verified
		return verified
	}
}

var (
	readmeRegexp = regexp.MustCompile(`^README_([a-z]{2}_[A-Za-z]{2,})\.md$`)
)

// Only the en_US readme should be at the root as README.md;
// all other readmes should be placed in the readme folder and named in the format README_$language_code.md.
// The separator is the separator of the file path, it's "/" for zip plugin and os.Separator for fs plugin.
func (p *PluginDecoderHelper) AvailableI18nReadme(decoder PluginDecoder, separator string) (map[string]string, error) {
	readmes := make(map[string]string)
	// read the en_US readme
	enUSReadme, err := decoder.ReadFile("README.md")
	if err != nil {
		// this file must exist or it's not a valid plugin
		return nil, errors.Join(err, fmt.Errorf("en_US readme not found"))
	}
	readmes["en_US"] = string(enUSReadme)

	readmeFiles, err := decoder.ReadDir("readme")
	if errors.Is(err, os.ErrNotExist) {
		return readmes, nil
	} else if err != nil {
		return nil, errors.Join(err, fmt.Errorf("an unexpected error occurred while reading readme folder"))
	}

	for _, file := range readmeFiles {
		// trim the readme folder prefix
		file, _ = strings.CutPrefix(file, "readme"+separator)
		// using regexp to match the file name
		match := readmeRegexp.FindStringSubmatch(file)
		if len(match) == 0 {
			continue
		}
		language := match[1]
		readme, err := decoder.ReadFile(filepath.Join("readme", file))
		if err != nil {
			return nil, errors.Join(err, fmt.Errorf("failed to read readme file: %s", file))
		}
		readmes[language] = string(readme)
	}

	return readmes, nil
}
