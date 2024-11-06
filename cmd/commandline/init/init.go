package init

import (
	"fmt"
	"os"
	"path/filepath"
	"time"

	_ "embed"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/langgenius/dify-plugin-daemon/internal/types/entities/constants"
	"github.com/langgenius/dify-plugin-daemon/internal/types/entities/plugin_entities"
	"github.com/langgenius/dify-plugin-daemon/internal/utils/log"
)

//go:embed templates/python/icon.svg
var icon []byte

func InitPlugin() {
	m := initialize()
	p := tea.NewProgram(m)
	if result, err := p.Run(); err != nil {
		fmt.Println("Error running program:", err)
	} else {
		if m, ok := result.(model); ok {
			if m.completed {
				m.createPlugin()
			}
		} else {
			log.Error("Error running program:", err)
			return
		}
	}
}

type subMenuKey string

const (
	SUB_MENU_KEY_PROFILE    subMenuKey = "profile"
	SUB_MENU_KEY_LANGUAGE   subMenuKey = "language"
	SUB_MENU_KEY_CATEGORY   subMenuKey = "category"
	SUB_MENU_KEY_PERMISSION subMenuKey = "permission"
)

type model struct {
	subMenus       map[subMenuKey]subMenu
	subMenuSeq     []subMenuKey
	currentSubMenu subMenuKey

	completed bool
}

func initialize() model {
	m := model{}
	m.subMenus = map[subMenuKey]subMenu{
		SUB_MENU_KEY_PROFILE:    newProfile(),
		SUB_MENU_KEY_LANGUAGE:   newLanguage(),
		SUB_MENU_KEY_CATEGORY:   newCategory(),
		SUB_MENU_KEY_PERMISSION: newPermission(),
	}
	m.currentSubMenu = SUB_MENU_KEY_PROFILE

	m.subMenuSeq = []subMenuKey{
		SUB_MENU_KEY_PROFILE,
		SUB_MENU_KEY_LANGUAGE,
		SUB_MENU_KEY_CATEGORY,
		SUB_MENU_KEY_PERMISSION,
	}

	return m
}

func (m model) Init() tea.Cmd {
	return nil
}

func (m model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	currentSubMenu, event, cmd := m.subMenus[m.currentSubMenu].Update(msg)
	m.subMenus[m.currentSubMenu] = currentSubMenu

	switch event {
	case SUB_MENU_EVENT_NEXT:
		if m.currentSubMenu != m.subMenuSeq[len(m.subMenuSeq)-1] {
			// move the current sub menu to the next one
			for i, key := range m.subMenuSeq {
				if key == m.currentSubMenu {
					m.currentSubMenu = m.subMenuSeq[i+1]
					break
				}
			}
		} else {
			m.completed = true
			return m, tea.Quit
		}
	case SUB_MENU_EVENT_PREV:
		if m.currentSubMenu != m.subMenuSeq[0] {
			// move the current sub menu to the previous one
			for i, key := range m.subMenuSeq {
				if key == m.currentSubMenu {
					m.currentSubMenu = m.subMenuSeq[i-1]
					break
				}
			}
		}
	}

	return m, cmd
}

func (m model) View() string {
	return m.subMenus[m.currentSubMenu].View()
}

func (m model) createPlugin() {
	permission := m.subMenus[SUB_MENU_KEY_PERMISSION].(permission).Permission()

	manifest := &plugin_entities.PluginDeclaration{
		PluginDeclarationWithoutAdvancedFields: plugin_entities.PluginDeclarationWithoutAdvancedFields{
			Version: "0.0.1",
			Type:    plugin_entities.PluginType,
			Icon:    "icon.svg",
			Author:  m.subMenus[SUB_MENU_KEY_PROFILE].(profile).Author(),
			Name:    m.subMenus[SUB_MENU_KEY_PROFILE].(profile).Name(),
			Description: plugin_entities.I18nObject{
				EnUS: m.subMenus[SUB_MENU_KEY_PROFILE].(profile).Description(),
			},
			CreatedAt: time.Now(),
			Resource: plugin_entities.PluginResourceRequirement{
				Memory:     1024 * 1024 * 256, // 256MB
				Permission: &permission,
			},
			Label: plugin_entities.I18nObject{
				EnUS: m.subMenus[SUB_MENU_KEY_PROFILE].(profile).Name(),
			},
		},
	}

	category_string := m.subMenus[SUB_MENU_KEY_CATEGORY].(category).Category()
	if category_string == "tool" {
		manifest.Plugins.Tools = []string{fmt.Sprintf("provider/%s.yaml", manifest.Name)}
	}

	manifest.Meta = plugin_entities.PluginMeta{
		Version: "0.0.1",
		Arch: []constants.Arch{
			constants.AMD64,
			constants.ARM64,
		},
		Runner: plugin_entities.PluginRunner{},
	}

	switch m.subMenus[SUB_MENU_KEY_LANGUAGE].(language).Language() {
	case constants.Python:
		manifest.Meta.Runner.Entrypoint = "main"
		manifest.Meta.Runner.Language = constants.Python
		manifest.Meta.Runner.Version = "3.10"
	default:
		log.Error("unsupported language: %s", m.subMenus[SUB_MENU_KEY_LANGUAGE].(language).Language())
		return
	}

	success := false

	clear := func() {
		if !success {
			os.RemoveAll(manifest.Name)
		}
	}
	defer clear()

	manifest_file := marshalYamlBytes(manifest)
	// create the plugin directory
	cwd, err := os.Getwd()
	if err != nil {
		log.Error("failed to get current working directory: %s", err)
		return
	}

	plugin_dir := filepath.Join(cwd, manifest.Name)
	if err := os.MkdirAll(plugin_dir, 0o755); err != nil {
		log.Error("failed to create plugin directory: %s", err)
		return
	}

	manifest_file_path := filepath.Join(plugin_dir, "manifest.yaml")
	if err := os.WriteFile(manifest_file_path, manifest_file, 0o644); err != nil {
		log.Error("failed to write manifest file: %s", err)
		return
	}

	// create _assets directory
	assets_dir := filepath.Join(plugin_dir, "_assets")
	if err := os.MkdirAll(assets_dir, 0o755); err != nil {
		log.Error("failed to create assets directory: %s", err)
		return
	}

	// create icon.svg
	icon_file_path := filepath.Join(assets_dir, "icon.svg")
	if err := os.WriteFile(icon_file_path, icon, 0o644); err != nil {
		log.Error("failed to write icon file: %s", err)
		return
	}

	err = createPythonEnvironment(
		plugin_dir,
		manifest.Meta.Runner.Entrypoint,
		manifest,
		m.subMenus[SUB_MENU_KEY_CATEGORY].(category).Category(),
	)
	if err != nil {
		log.Error("failed to create python environment: %s", err)
		return
	}

	success = true

	log.Info("plugin %s created successfully", manifest.Name)
}
