package model

import (
	"embed"
	"fmt"
	"strings"

	"github.com/langgenius/dify-plugin-daemon/internal/utils/log"
)

//go:embed templates
var templates embed.FS

// provider_templates is a map of provider type to the template name
var provider_templates map[string]string

// model_templates is a map of model type to a map of template name to the template content
var model_templates map[string]map[string]string

func init() {
	provider_templates = make(map[string]string)
	model_templates = make(map[string]map[string]string)

	files, err := templates.ReadDir("templates")
	if err != nil {
		log.Error("Failed to read templates: %v", err)
		return
	}

	for _, file := range files {
		if file.IsDir() {
			continue
		}

		// get the file name
		filename := file.Name()
		// read the file content
		file_content, err := templates.ReadFile("templates/" + filename)
		if err != nil {
			log.Error("Failed to read template: %v", err)
			continue
		}
		filenames := strings.Split(filename, "_")
		// check the first element is a provider
		if filenames[0] == "provider" {
			if len(filenames) != 2 {
				log.Error("Invalid provider template: %s", filename)
				continue
			}
			provider_templates[filenames[1]] = string(file_content)
		} else if filenames[0] == "model" {
			if len(filenames) != 3 {
				log.Error("Invalid model template: %s", filename)
				continue
			}
			if _, ok := model_templates[filenames[1]]; !ok {
				model_templates[filenames[1]] = make(map[string]string)
			}

			model_templates[filenames[1]][filenames[2]] = string(file_content)
		}
	}
}

func ListTemplates(typ string, model_type string, name string) {
	color_reset := "\033[0m"
	color_cyan := "\033[36m"
	color_yellow := "\033[33m"
	color_green := "\033[32m"

	if typ == "provider" || typ == "" {
		fmt.Printf("%sProvider Templates:%s\n", color_cyan, color_reset)
		for template := range provider_templates {
			if name == "" || strings.Contains(template, name) {
				fmt.Printf("  %s%s%s\n", color_yellow, template, color_reset)
			}
		}
		fmt.Println()
	}

	if typ == "model" || typ == "" {
		fmt.Printf("%sModel Templates:%s\n", color_cyan, color_reset)
		if model_type == "" {
			for model_type, templates := range model_templates {
				fmt.Printf("%s%s:%s\n", color_yellow, model_type, color_reset)
				for template := range templates {
					if name == "" || strings.Contains(template, name) {
						fmt.Printf("  %s%s%s\n", color_green, template, color_reset)
					}
				}
				fmt.Println()
			}
		} else {
			if templates, ok := model_templates[model_type]; ok {
				fmt.Printf("%s%s:%s\n", color_yellow, model_type, color_reset)
				for template := range templates {
					if name == "" || strings.Contains(template, name) {
						fmt.Printf("  %s%s%s\n", color_green, template, color_reset)
					}
				}
				fmt.Println()
			}
		}
	}
}

func GetTemplate(typ string, name string) {

}

func CreateFromTemplate(root string, typ string, name string) {

}
