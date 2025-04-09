package plugin

import (
	"fmt"
	"html/template"
	"os"
	"path/filepath"

	"github.com/langgenius/dify-plugin-daemon/internal/utils/log"
	"github.com/langgenius/dify-plugin-daemon/pkg/entities/constants"
	"github.com/langgenius/dify-plugin-daemon/pkg/plugin_packager/decoder"
)

const TOOL_MODULE_TEMPLATE = `
========== {{.Identity.Name}} ==========
Author: {{.Identity.Author}}
Label: {{.Identity.Label.EnUS}}
Description: {{.Description.Human.EnUS}}
Parameters:
{{- range .Parameters}}
  - Name: {{.Name}}
    Type: {{.Type}}
    Required: {{.Required}}
    Description: {{.HumanDescription.EnUS}}
    {{- if .Default}}
    Default: {{.Default}}
    {{- end}}
    {{- if .Options}}
    Options:
      {{- range .Options}}
      - Value: {{.Value}}
        Label: {{.Label.EnUS}}
      {{- end}}
    {{- end}}
{{- end}}
`

const AGENT_MODULE_TEMPLATE = `
========== {{.Identity.Name}} ==========
Author: {{.Identity.Author}}
Label: {{.Identity.Label.EnUS}}
Description: {{.Description.EnUS}}
Parameters:
{{- range .Parameters}}
  - Name: {{.Name}}
    Type: {{.Type}}
    Required: {{.Required}}
    {{- if .Default}}
    Default: {{.Default}}
    {{- end}}
{{- end}}
`

const MODEL_MODULE_TEMPLATE = `
========== {{.Model}} ==========
Name: {{.Model}}
Type: {{.ModelType}}
Label: {{.Label.EnUS}}
Parameters:
{{- range .ParameterRules}}
  - Name: {{.Name}}
    Type: {{.Type}}
    Required: {{.Required}}
    Description: {{.Help.EnUS}}
    {{- if .Default}}
    Default: {{.Default}}
    {{- end}}
    {{- if .Min}}
    Min: {{.Min}}
    {{- end}}
    {{- if .Max}}
    Max: {{.Max}}
    {{- end}}
    {{- if .Options}}
    Options: {{range .Options}}{{.}}, {{end}}
    {{- end}}
{{- end}}
`

const ENDPOINT_MODULE_TEMPLATE = `
========== Endpoints ==========
Path: {{.Path}}
Method: {{.Method}}
`

const PLUGIN_MODULE_TEMPLATE = `
========== Plugin ==========
Name: {{.Name}}
Version: {{.Version}}
Description: {{.Description.EnUS}}
Author: {{.Author}}
Icon: {{.Icon}}
Tags: {{range .Tags}}{{.}}, {{end}}
Category: {{.Category}}
Resource:
  Memory: {{.Resource.Memory}} bytes
Permissions:
  {{- if .Resource.Permission.Tool}}
  Tool: {{.Resource.Permission.Tool.Enabled}}
  {{- end}}
  {{- if .Resource.Permission.Model}}
  Model:
    Enabled: {{.Resource.Permission.Model.Enabled}}
    LLM: {{.Resource.Permission.Model.LLM}}
    TextEmbedding: {{.Resource.Permission.Model.TextEmbedding}}
    Rerank: {{.Resource.Permission.Model.Rerank}}
    TTS: {{.Resource.Permission.Model.TTS}}
    Speech2text: {{.Resource.Permission.Model.Speech2text}}
    Moderation: {{.Resource.Permission.Model.Moderation}}
  {{- end}}
  {{- if .Resource.Permission.Node}}
  Node: {{.Resource.Permission.Node.Enabled}}
  {{- end}}
  {{- if .Resource.Permission.Endpoint}}
  Endpoint: {{.Resource.Permission.Endpoint.Enabled}}
  {{- end}}
  {{- if .Resource.Permission.App}}
  App: {{.Resource.Permission.App.Enabled}}
  {{- end}}
  {{- if .Resource.Permission.Storage}}
  Storage:
    Enabled: {{.Resource.Permission.Storage.Enabled}}
    Size: {{.Resource.Permission.Storage.Size}} bytes
  {{- end}}
`

func ModuleList(pluginPath string) {
	var pluginDecoder decoder.PluginDecoder
	var err error

	stat, err := os.Stat(pluginPath)
	if err != nil {
		log.Error("failed to get plugin file stat: %s", err)
		return
	}

	if stat.IsDir() {
		pluginDecoder, err = decoder.NewFSPluginDecoder(pluginPath)
	} else {
		fileContent, err := os.ReadFile(pluginPath)
		if err != nil {
			log.Error("failed to read plugin file: %s", err)
			return
		}
		pluginDecoder, err = decoder.NewZipPluginDecoder(fileContent)
		if err != nil {
			log.Error("failed to create zip plugin decoder: %s", err)
			return
		}
	}
	if err != nil {
		log.Error("your plugin is not a valid plugin: %s", err)
		return
	}

	manifest, err := pluginDecoder.Manifest()
	if err != nil {
		log.Error("failed to get manifest: %s", err)
		return
	}

	if manifest.Tool != nil {
		for _, tool := range manifest.Tool.Tools {
			tmpl, err := template.New("tool").Parse(TOOL_MODULE_TEMPLATE)
			if err != nil {
				log.Error("failed to parse template: %s", err)
				return
			}

			err = tmpl.Execute(os.Stdout, tool)
			if err != nil {
				log.Error("failed to execute template: %s", err)
				return
			}
		}
	}

	if manifest.AgentStrategy != nil {
		for _, strategy := range manifest.AgentStrategy.Strategies {
			tmpl, err := template.New("agent").Parse(AGENT_MODULE_TEMPLATE)
			if err != nil {
				log.Error("failed to parse template: %s", err)
				return
			}

			err = tmpl.Execute(os.Stdout, strategy)
			if err != nil {
				log.Error("failed to execute template: %s", err)
				return
			}
		}
	}

	if manifest.Model != nil {
		for _, model := range manifest.Model.Models {
			tmpl, err := template.New("model").Parse(MODEL_MODULE_TEMPLATE)
			if err != nil {
				log.Error("failed to parse template: %s", err)
				return
			}

			err = tmpl.Execute(os.Stdout, model)
			if err != nil {
				log.Error("failed to execute template: %s", err)
				return
			}
		}
	}

	if manifest.Endpoint != nil {
		for _, endpoint := range manifest.Endpoint.Endpoints {
			tmpl, err := template.New("endpoint").Parse(ENDPOINT_MODULE_TEMPLATE)
			if err != nil {
				log.Error("failed to parse template: %s", err)
				return
			}

			err = tmpl.Execute(os.Stdout, endpoint)
			if err != nil {
				log.Error("failed to execute template: %s", err)
				return
			}
		}
	}
}

func ModuleAppendTools(pluginPath string) {
	decoder, err := decoder.NewFSPluginDecoder(pluginPath)
	if err != nil {
		log.Error("your plugin is not a valid plugin: %s", err)
		return
	}

	manifest, err := decoder.Manifest()
	if err != nil {
		log.Error("failed to get manifest: %s", err)
		return
	}

	if manifest.Tool != nil {
		log.Error("you have already declared tools in this plugin, " +
			"you can add new tool by modifying the `provider.yaml` file to add new tools, " +
			"this command is used to create new module that never been declared in this plugin.")
		return
	}

	if manifest.Model != nil {
		log.Error("model plugin dose not support declare tools.")
		return
	}

	if manifest.Plugins.Tools == nil {
		manifest.Plugins.Tools = []string{}
	}

	manifest.Plugins.Tools = append(manifest.Plugins.Tools, fmt.Sprintf("provider/%s.yaml", manifest.Name))

	if manifest.Meta.Runner.Language == constants.Python {
		if err := createPythonTool(pluginPath, &manifest); err != nil {
			log.Error("failed to create python tool: %s", err)
			return
		}

		if err := createPythonToolProvider(pluginPath, &manifest); err != nil {
			log.Error("failed to create python tool provider: %s", err)
			return
		}
	}

	// save manifest
	manifest_file := marshalYamlBytes(manifest.PluginDeclarationWithoutAdvancedFields)
	if err := writeFile(filepath.Join(pluginPath, "manifest.yaml"), string(manifest_file)); err != nil {
		log.Error("failed to save manifest: %s", err)
		return
	}

	log.Info("created tool module successfully")
}

func ModuleAppendEndpoints(pluginPath string) {
	decoder, err := decoder.NewFSPluginDecoder(pluginPath)
	if err != nil {
		log.Error("your plugin is not a valid plugin: %s", err)
		return
	}

	manifest, err := decoder.Manifest()
	if err != nil {
		log.Error("failed to get manifest: %s", err)
		return
	}

	if manifest.Endpoint != nil {
		log.Error("you have already declared endpoints in this plugin, " +
			"you can add new endpoint by modifying the `provider.yaml` file to add new endpoints, " +
			"this command is used to create new module that never been declared in this plugin.")
		return
	}

	if manifest.Model != nil {
		log.Error("model plugin dose not support declare endpoints.")
		return
	}

	if manifest.Plugins.Endpoints == nil {
		manifest.Plugins.Endpoints = []string{}
	}

	manifest.Plugins.Endpoints = append(manifest.Plugins.Endpoints, fmt.Sprintf("group/%s.yaml", manifest.Name))

	if manifest.Meta.Runner.Language == constants.Python {
		if err := createPythonEndpoint(pluginPath, &manifest); err != nil {
			log.Error("failed to create python endpoint: %s", err)
			return
		}

		if err := createPythonEndpointGroup(pluginPath, &manifest); err != nil {
			log.Error("failed to create python group: %s", err)
			return
		}
	}

	// save manifest
	manifest_file := marshalYamlBytes(manifest.PluginDeclarationWithoutAdvancedFields)
	if err := writeFile(filepath.Join(pluginPath, "manifest.yaml"), string(manifest_file)); err != nil {
		log.Error("failed to save manifest: %s", err)
		return
	}

	log.Info("created endpoint module successfully")
}
