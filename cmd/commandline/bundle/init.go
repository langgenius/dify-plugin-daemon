package bundle

import (
	_ "embed"
	"errors"
	"os"
	"path"

	"github.com/langgenius/dify-plugin-daemon/internal/types/entities/bundle_entities"
	"github.com/langgenius/dify-plugin-daemon/internal/types/entities/manifest_entities"
	"github.com/langgenius/dify-plugin-daemon/internal/types/entities/plugin_entities"
	"github.com/langgenius/dify-plugin-daemon/internal/utils/log"
	"github.com/langgenius/dify-plugin-daemon/internal/utils/parser"

	tea "github.com/charmbracelet/bubbletea"
)

//go:embed templates/icon.svg
var BUNDLE_ICON []byte

func generateNewBundle() (*bundle_entities.Bundle, error) {
	m := newProfile()
	p := tea.NewProgram(m)
	if result, err := p.Run(); err != nil {
		return nil, err
	} else {
		if _, ok := result.(profile); ok {
			author := m.inputs[1].Value()
			name := m.inputs[0].Value()
			description := m.inputs[2].Value()

			bundle := &bundle_entities.Bundle{
				Name:         name,
				Icon:         "icon.svg",
				Labels:       plugin_entities.NewI18nObject(name),
				Description:  plugin_entities.NewI18nObject(description),
				Version:      "0.0.1",
				Author:       author,
				Type:         manifest_entities.BundleType,
				Dependencies: []bundle_entities.Dependency{},
			}

			return bundle, nil
		} else {
			return nil, errors.New("invalid profile")
		}
	}
}

func InitBundle() {
	bundle, err := generateNewBundle()
	if err != nil {
		log.Error("Failed to generate new bundle: %v", err)
		return
	}

	// create bundle directory
	cwd, err := os.Getwd()
	if err != nil {
		log.Error("Error getting current directory: %v", err)
		return
	}

	bundleDir := path.Join(cwd, bundle.Name)
	if err := os.MkdirAll(bundleDir, 0755); err != nil {
		log.Error("Error creating bundle directory: %v", err)
		return
	}

	success := false
	defer func() {
		if !success {
			os.RemoveAll(bundleDir)
		}
	}()

	// save
	bundleYaml := parser.MarshalYamlBytes(bundle)
	if err := os.WriteFile(path.Join(bundleDir, "manifest.yaml"), bundleYaml, 0644); err != nil {
		log.Error("Error saving manifest.yaml: %v", err)
		return
	}

	// create _assets directory
	if err := os.MkdirAll(path.Join(bundleDir, "_assets"), 0755); err != nil {
		log.Error("Error creating _assets directory: %v", err)
		return
	}

	// create _assets/icon.svg
	if err := os.WriteFile(path.Join(bundleDir, "_assets", "icon.svg"), BUNDLE_ICON, 0644); err != nil {
		log.Error("Error saving icon.svg: %v", err)
		return
	}

	success = true

	log.Info("Bundle created successfully: %s", bundleDir)
}
