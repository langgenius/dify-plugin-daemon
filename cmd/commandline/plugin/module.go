package plugin

import (
	"html/template"
	"os"

	"github.com/langgenius/dify-plugin-daemon/internal/core/plugin_packager/decoder"
	"github.com/langgenius/dify-plugin-daemon/internal/utils/log"
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

}

func ModuleAppendEndpoints(pluginPath string) {

}
