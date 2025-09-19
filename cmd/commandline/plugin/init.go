package plugin

import (
	"fmt"
	"os"
	"path/filepath"
	"time"

	_ "embed"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/langgenius/dify-plugin-daemon/internal/utils/log"
	"github.com/langgenius/dify-plugin-daemon/internal/utils/parser"
	"github.com/langgenius/dify-plugin-daemon/pkg/entities/constants"
	"github.com/langgenius/dify-plugin-daemon/pkg/entities/manifest_entities"
	"github.com/langgenius/dify-plugin-daemon/pkg/entities/plugin_entities"
)

var (
	//go:embed templates/icons/agent_light.svg
	agentLight []byte
	//go:embed templates/icons/agent_dark.svg
	agentDark []byte
	//go:embed templates/icons/datasource_light.svg
	datasourceLight []byte
	//go:embed templates/icons/datasource_dark.svg
	datasourceDark []byte
	//go:embed templates/icons/extension_light.svg
	extensionLight []byte
	//go:embed templates/icons/extension_dark.svg
	extensionDark []byte
	//go:embed templates/icons/model_light.svg
	modelLight []byte
	//go:embed templates/icons/model_dark.svg
	modelDark []byte
	//go:embed templates/icons/tool_light.svg
	toolLight []byte
	//go:embed templates/icons/tool_dark.svg
	toolDark []byte
	//go:embed templates/icons/trigger_light.svg
	triggerLight []byte
	//go:embed templates/icons/trigger_dark.svg
	triggerDark []byte

	//go:embed templates/readme/zh_Hans.md
	zhHansReadme []byte
	//go:embed templates/readme/ja_JP.md
	jaJpReadme []byte
	//go:embed templates/readme/pt_BR.md
	ptBrReadme []byte
)

var icon = map[string]map[string][]byte{
	"light": {
		"agent-strategy": agentLight,
		"datasource":     datasourceLight,
		"extension":      extensionLight,
		"model":          modelLight,
		"tool":           toolLight,
		"trigger":        triggerLight,
	},
	"dark": {
		"agent-strategy": agentDark,
		"datasource":     datasourceDark,
		"extension":      extensionDark,
		"model":          modelDark,
		"tool":           toolDark,
		"trigger":        triggerDark,
	},
}

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

func InitPluginWithFlags(
	author string,
	name string,
	repo string,
	description string,
	allowRegisterEndpoint bool,
	allowInvokeTool bool,
	allowInvokeModel bool,
	allowInvokeLLM bool,
	allowInvokeTextEmbedding bool,
	allowInvokeRerank bool,
	allowInvokeTTS bool,
	allowInvokeSpeech2Text bool,
	allowInvokeModeration bool,
	allowInvokeNode bool,
	allowInvokeApp bool,
	allowUseStorage bool,
	storageSize uint64,
	categoryStr string,
	languageStr string,
	minDifyVersion string,
	quick bool,
) {
	if quick {
		// Validate name and author
		if !plugin_entities.PluginNameRegex.MatchString(name) {
			log.Error("Plugin name must be 1-128 characters long, and can only contain lowercase letters, numbers, dashes and underscores")
			return
		}
		if !plugin_entities.AuthorRegex.MatchString(author) {
			log.Error("Author name must be 1-64 characters long, and can only contain lowercase letters, numbers, dashes and underscores")
			return
		}
		if description == "" {
			log.Error("Description cannot be empty")
			return
		}
	}

	// Validate language
	if languageStr != "" {
		validLanguages := []string{
			string(constants.Python),
			// Add more languages here if supported
		}
		valid := false
		for _, lang := range validLanguages {
			if languageStr == lang {
				valid = true
				break
			}
		}
		if !valid {
			log.Error("Invalid language. Supported languages are: %v", validLanguages)
			return
		}
	}

	// Validate category
	if categoryStr != "" {
		validCategories := []string{
			"tool",
			"llm",
			"text-embedding",
			"speech2text",
			"moderation",
			"rerank",
			"tts",
			"extension",
			"agent-strategy",
			"datasource",
		}
		valid := false
		for _, cat := range validCategories {
			if categoryStr == cat {
				valid = true
				break
			}
		}
		if !valid {
			log.Error("Invalid category. Supported categories are: %v", validCategories)
			return
		}
	}

	m := newModel()

	// Set profile information
	profile := m.subMenus[SUB_MENU_KEY_PROFILE].(profile)
	profile.SetAuthor(author)
	profile.SetName(name)
	profile.inputs[2].SetValue(description)
	m.subMenus[SUB_MENU_KEY_PROFILE] = profile

	// Set category if provided
	if categoryStr != "" {
		cat := m.subMenus[SUB_MENU_KEY_CATEGORY].(category)
		cat.SetCategory(categoryStr)
		m.subMenus[SUB_MENU_KEY_CATEGORY] = cat
	}

	// Set language if provided
	if languageStr != "" {
		lang := m.subMenus[SUB_MENU_KEY_LANGUAGE].(language)
		lang.SetLanguage(languageStr)
		m.subMenus[SUB_MENU_KEY_LANGUAGE] = lang
	}

	// Set minimal Dify version if provided
	if minDifyVersion != "" {
		ver := m.subMenus[SUB_MENU_KEY_VERSION_REQUIRE].(versionRequire)
		ver.SetMinimalDifyVersion(minDifyVersion)
		m.subMenus[SUB_MENU_KEY_VERSION_REQUIRE] = ver
	}

	// Update permissions
	perm := m.subMenus[SUB_MENU_KEY_PERMISSION].(permission)
	permissionRequirement := &plugin_entities.PluginPermissionRequirement{}

	if allowRegisterEndpoint {
		permissionRequirement.Endpoint = &plugin_entities.PluginPermissionEndpointRequirement{
			Enabled: allowRegisterEndpoint,
		}
	}

	if allowInvokeTool {
		permissionRequirement.Tool = &plugin_entities.PluginPermissionToolRequirement{
			Enabled: allowInvokeTool,
		}
	}

	if allowInvokeModel {
		permissionRequirement.Model = &plugin_entities.PluginPermissionModelRequirement{
			Enabled:       allowInvokeModel,
			LLM:           allowInvokeLLM,
			TextEmbedding: allowInvokeTextEmbedding,
			Rerank:        allowInvokeRerank,
			TTS:           allowInvokeTTS,
			Speech2text:   allowInvokeSpeech2Text,
			Moderation:    allowInvokeModeration,
		}
	}

	if allowInvokeNode {
		permissionRequirement.Node = &plugin_entities.PluginPermissionNodeRequirement{
			Enabled: allowInvokeNode,
		}
	}

	if allowInvokeApp {
		permissionRequirement.App = &plugin_entities.PluginPermissionAppRequirement{
			Enabled: allowInvokeApp,
		}
	}

	if allowUseStorage {
		permissionRequirement.Storage = &plugin_entities.PluginPermissionStorageRequirement{
			Enabled: allowUseStorage,
			Size:    storageSize,
		}
	}

	perm.UpdatePermission(*permissionRequirement)
	m.subMenus[SUB_MENU_KEY_PERMISSION] = perm

	// If quick mode is enabled, skip interactive mode
	if quick {
		m.createPlugin()
		return
	}

	// Otherwise, start interactive mode
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
	SUB_MENU_KEY_PROFILE         subMenuKey = "profile"
	SUB_MENU_KEY_LANGUAGE        subMenuKey = "language"
	SUB_MENU_KEY_CATEGORY        subMenuKey = "category"
	SUB_MENU_KEY_PERMISSION      subMenuKey = "permission"
	SUB_MENU_KEY_VERSION_REQUIRE subMenuKey = "version_require"
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
		SUB_MENU_KEY_PROFILE:         newProfile(),
		SUB_MENU_KEY_LANGUAGE:        newLanguage(),
		SUB_MENU_KEY_CATEGORY:        newCategory(),
		SUB_MENU_KEY_PERMISSION:      newPermission(plugin_entities.PluginPermissionRequirement{}),
		SUB_MENU_KEY_VERSION_REQUIRE: newVersionRequire(),
	}
	m.currentSubMenu = SUB_MENU_KEY_PROFILE

	m.subMenuSeq = []subMenuKey{
		SUB_MENU_KEY_PROFILE,
		SUB_MENU_KEY_LANGUAGE,
		SUB_MENU_KEY_CATEGORY,
		SUB_MENU_KEY_PERMISSION,
		SUB_MENU_KEY_VERSION_REQUIRE,
	}

	return m
}

func newModel() model {
	m := model{}
	m.subMenus = map[subMenuKey]subMenu{
		SUB_MENU_KEY_PROFILE:         newProfile(),
		SUB_MENU_KEY_LANGUAGE:        newLanguage(),
		SUB_MENU_KEY_CATEGORY:        newCategory(),
		SUB_MENU_KEY_PERMISSION:      newPermission(plugin_entities.PluginPermissionRequirement{}),
		SUB_MENU_KEY_VERSION_REQUIRE: newVersionRequire(),
	}
	m.currentSubMenu = SUB_MENU_KEY_PROFILE

	m.subMenuSeq = []subMenuKey{
		SUB_MENU_KEY_PROFILE,
		SUB_MENU_KEY_LANGUAGE,
		SUB_MENU_KEY_CATEGORY,
		SUB_MENU_KEY_PERMISSION,
		SUB_MENU_KEY_VERSION_REQUIRE,
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
					// check if the next sub menu is permission
					if key == SUB_MENU_KEY_CATEGORY {
						// get the type of current category
						category := m.subMenus[SUB_MENU_KEY_CATEGORY].(category).Category()
						if category == "agent-strategy" {
							// update the permission to add tool and model invocation
							perm := m.subMenus[SUB_MENU_KEY_PERMISSION].(permission)
							perm.UpdatePermission(plugin_entities.PluginPermissionRequirement{
								Tool: &plugin_entities.PluginPermissionToolRequirement{
									Enabled: true,
								},
								Model: &plugin_entities.PluginPermissionModelRequirement{
									Enabled: true,
									LLM:     true,
								},
							})
							m.subMenus[SUB_MENU_KEY_PERMISSION] = perm
						}
					}
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
			Version:     manifest_entities.Version("0.0.1"),
			Type:        manifest_entities.PluginType,
			Icon:        "icon.svg",
			IconDark:    "icon-dark.svg",
			Author:      m.subMenus[SUB_MENU_KEY_PROFILE].(profile).Author(),
			Name:        m.subMenus[SUB_MENU_KEY_PROFILE].(profile).Name(),
			Description: plugin_entities.NewI18nObject(m.subMenus[SUB_MENU_KEY_PROFILE].(profile).Description()),
			CreatedAt:   time.Now(),
			Resource: plugin_entities.PluginResourceRequirement{
				Memory:     1024 * 1024 * 256, // 256MB
				Permission: &permission,
			},
			Label:   plugin_entities.NewI18nObject(m.subMenus[SUB_MENU_KEY_PROFILE].(profile).Name()),
			Privacy: parser.ToPtr("PRIVACY.md"),
		},
	}

	repo := m.subMenus[SUB_MENU_KEY_PROFILE].(profile).Repo()
	if repo != "" {
		manifest.Repo = parser.ToPtr(repo)
	}

	categoryString := m.subMenus[SUB_MENU_KEY_CATEGORY].(category).Category()
	if categoryString == "tool" {
		manifest.Plugins.Tools = []string{fmt.Sprintf("provider/%s.yaml", manifest.Name)}
	}

	if categoryString == "llm" ||
		categoryString == "text-embedding" ||
		categoryString == "speech2text" ||
		categoryString == "moderation" ||
		categoryString == "rerank" ||
		categoryString == "tts" {
		manifest.Plugins.Models = []string{fmt.Sprintf("provider/%s.yaml", manifest.Name)}
	}

	if categoryString == "extension" {
		manifest.Plugins.Endpoints = []string{fmt.Sprintf("group/%s.yaml", manifest.Name)}
	}

	if categoryString == "agent-strategy" {
		manifest.Plugins.AgentStrategies = []string{fmt.Sprintf("provider/%s.yaml", manifest.Name)}
	}

	if categoryString == "datasource" {
		manifest.Plugins.Datasources = []string{fmt.Sprintf("provider/%s.yaml", manifest.Name)}
	}

	manifest.Meta = plugin_entities.PluginMeta{
		Version: "0.0.1",
		Arch: []constants.Arch{
			constants.AMD64,
			constants.ARM64,
		},
		Runner: plugin_entities.PluginRunner{},
	}

	if m.subMenus[SUB_MENU_KEY_VERSION_REQUIRE].(versionRequire).MinimalDifyVersion() != "" {
		manifest.Meta.MinimumDifyVersion = parser.ToPtr(m.subMenus[SUB_MENU_KEY_VERSION_REQUIRE].(versionRequire).MinimalDifyVersion())
	}

	switch m.subMenus[SUB_MENU_KEY_LANGUAGE].(language).Language() {
	case constants.Python:
		manifest.Meta.Runner.Entrypoint = "main"
		manifest.Meta.Runner.Language = constants.Python
		manifest.Meta.Runner.Version = "3.12"
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

	manifestFile := marshalYamlBytes(manifest)
	// create the plugin directory
	cwd, err := os.Getwd()
	if err != nil {
		log.Error("failed to get current working directory: %s", err)
		return
	}

	pluginDir := filepath.Join(cwd, manifest.Name)

	if err := writeFile(filepath.Join(pluginDir, "manifest.yaml"), string(manifestFile)); err != nil {
		log.Error("failed to write manifest file: %s", err)
		return
	}

	// get icon and icon-dark
	iconLight := icon["light"][string(manifest.Category())]
	if iconLight == nil {
		log.Error("icon not found for category: %s", manifest.Category())
		return
	}
	iconDark := icon["dark"][string(manifest.Category())]
	if iconDark == nil {
		log.Error("icon-dark not found for category: %s", manifest.Category())
		return
	}

	// create icon.svg
	if err := writeFile(filepath.Join(pluginDir, "_assets", "icon.svg"), string(iconLight)); err != nil {
		log.Error("failed to write icon file: %s", err)
		return
	}

	// create icon-dark.svg
	if err := writeFile(filepath.Join(pluginDir, "_assets", "icon-dark.svg"), string(iconDark)); err != nil {
		log.Error("failed to write icon-dark file: %s", err)
		return
	}

	// create README.md
	readme, err := renderTemplate(README, manifest, []string{})
	if err != nil {
		log.Error("failed to render README template: %s", err)
		return
	}
	if err := writeFile(filepath.Join(pluginDir, "README.md"), readme); err != nil {
		log.Error("failed to write README file: %s", err)
		return
	}

	// create multilingual README files if enabled
	profileMenu := m.subMenus[SUB_MENU_KEY_PROFILE].(profile)
	if profileMenu.EnableI18nReadme() {
		selectedLanguages := profileMenu.SelectedLanguages()

		// Define language template mapping
		languageTemplates := map[string][]byte{
			"zh_Hans": zhHansReadme,
			"ja_JP":   jaJpReadme,
			"pt_BR":   ptBrReadme,
		}

		for _, lang := range selectedLanguages {
			if lang == "en" {
				// English README is already created as README.md
				continue
			}

			if template, exists := languageTemplates[lang]; exists {
				// Render the template for this language
				langReadme, err := renderTemplate(template, manifest, []string{})
				if err != nil {
					log.Error("failed to render %s README template: %s", lang, err)
					return
				}

				// Write the language-specific README file
				readmeFilename := fmt.Sprintf("README_%s.md", lang)
				if err := writeFile(filepath.Join(pluginDir, "readme", readmeFilename), langReadme); err != nil {
					log.Error("failed to write %s README file: %s", lang, err)
					return
				}
			}
		}
	}

	// create .env.example
	if err := writeFile(filepath.Join(pluginDir, ".env.example"), string(ENV_EXAMPLE)); err != nil {
		log.Error("failed to write .env.example file: %s", err)
		return
	}

	// create PRIVACY.md
	if err := writeFile(filepath.Join(pluginDir, "PRIVACY.md"), string(PRIVACY)); err != nil {
		log.Error("failed to write PRIVACY file: %s", err)
		return
	}
	// create github CI workflow
	if err := writeFile(filepath.Join(pluginDir, ".github", "workflows", "plugin-publish.yml"), string(PLUGIN_PUBLISH_WORKFLOW)); err != nil {
		log.Error("failed to write plugin-publish workflow file: %s", err)
		return
	}

	err = createPythonEnvironment(
		pluginDir,
		manifest.Meta.Runner.Entrypoint,
		manifest,
		m.subMenus[SUB_MENU_KEY_CATEGORY].(category).Category(),
	)
	if err != nil {
		log.Error("failed to create python environment: %s", err)
		return
	}

	success = true

	log.Info("plugin %s created successfully, you can refer to `%s/GUIDE.md` for more information about how to develop it", manifest.Name, manifest.Name)
}
