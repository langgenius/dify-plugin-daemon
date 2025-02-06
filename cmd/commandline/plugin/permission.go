package plugin

import (
	"fmt"
	"os"
	"path/filepath"
	"strconv"
	"strings"

	ti "github.com/charmbracelet/bubbles/textinput"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/langgenius/dify-plugin-daemon/internal/utils/log"
	"github.com/langgenius/dify-plugin-daemon/pkg/entities/plugin_entities"
	"github.com/langgenius/dify-plugin-daemon/pkg/plugin_packager/decoder"
)

var permissionKeySeq = []string{
	"tool.enabled",
	"model.enabled",
	"model.llm",
	"model.text_embedding",
	"model.rerank",
	"model.tts",
	"model.speech2text",
	"model.moderation",
	"app.enabled",
	"storage.enabled",
	"storage.size",
	"endpoint.enabled",
}

type permission struct {
	cursor string

	permission plugin_entities.PluginPermissionRequirement

	storageSizeEditor ti.Model
}

func newPermission(defaultPermission plugin_entities.PluginPermissionRequirement) permission {
	return permission{
		cursor:            permissionKeySeq[0],
		permission:        defaultPermission,
		storageSizeEditor: ti.New(),
	}
}

func (p permission) Permission() plugin_entities.PluginPermissionRequirement {
	return p.permission
}

func (p permission) View() string {
	cursor := func(key string) string {
		if p.cursor == key {
			return "→ "
		}
		return "  "
	}

	checked := func(enabled bool) string {
		if enabled {
			return fmt.Sprintf("\033[32m%s\033[0m", "[✔]")
		}
		return fmt.Sprintf("\033[31m%s\033[0m", "[✘]")
	}

	s := "Configure the permissions of the plugin, use \033[32mup\033[0m and \033[32mdown\033[0m to navigate, \033[32mtab\033[0m to select, after selection, press \033[32menter\033[0m to finish\n"
	s += "Backwards Invocation:\n"
	s += "Tools:\n"
	s += fmt.Sprintf("  %sEnabled: %v %s You can invoke tools inside Dify if it's enabled %s\n", cursor("tool.enabled"), checked(p.permission.AllowInvokeTool()), YELLOW, RESET)
	s += "Models:\n"
	s += fmt.Sprintf("  %sEnabled: %v %s You can invoke models inside Dify if it's enabled %s\n", cursor("model.enabled"), checked(p.permission.AllowInvokeModel()), YELLOW, RESET)
	s += fmt.Sprintf("  %sLLM: %v %s You can invoke LLM models inside Dify if it's enabled %s\n", cursor("model.llm"), checked(p.permission.AllowInvokeLLM()), YELLOW, RESET)
	s += fmt.Sprintf("  %sText Embedding: %v %s You can invoke text embedding models inside Dify if it's enabled %s\n", cursor("model.text_embedding"), checked(p.permission.AllowInvokeTextEmbedding()), YELLOW, RESET)
	s += fmt.Sprintf("  %sRerank: %v %s You can invoke rerank models inside Dify if it's enabled %s\n", cursor("model.rerank"), checked(p.permission.AllowInvokeRerank()), YELLOW, RESET)
	s += fmt.Sprintf("  %sTTS: %v %s You can invoke TTS models inside Dify if it's enabled %s\n", cursor("model.tts"), checked(p.permission.AllowInvokeTTS()), YELLOW, RESET)
	s += fmt.Sprintf("  %sSpeech2Text: %v %s You can invoke speech2text models inside Dify if it's enabled %s\n", cursor("model.speech2text"), checked(p.permission.AllowInvokeSpeech2Text()), YELLOW, RESET)
	s += fmt.Sprintf("  %sModeration: %v %s You can invoke moderation models inside Dify if it's enabled %s\n", cursor("model.moderation"), checked(p.permission.AllowInvokeModeration()), YELLOW, RESET)
	s += "Apps:\n"
	s += fmt.Sprintf("  %sEnabled: %v %s Ability to invoke apps like BasicChat/ChatFlow/Agent/Workflow etc. %s\n", cursor("app.enabled"), checked(p.permission.AllowInvokeApp()), YELLOW, RESET)
	s += "Resources:\n"
	s += "Storage:\n"
	s += fmt.Sprintf("  %sEnabled: %v %s Persistence storage for the plugin %s\n", cursor("storage.enabled"), checked(p.permission.AllowInvokeStorage()), YELLOW, RESET)

	if p.permission.AllowInvokeStorage() {
		s += fmt.Sprintf("  %sSize: %v\n", cursor("storage.size"), p.storageSizeEditor.View())
	} else {
		s += fmt.Sprintf("  %sSize: %v %s The maximum size of the storage %s\n", cursor("storage.size"), "N/A", YELLOW, RESET)
	}

	s += "Endpoints:\n"
	s += fmt.Sprintf("  %sEnabled: %v %s Ability to register endpoints %s\n", cursor("endpoint.enabled"), checked(p.permission.AllowRegisterEndpoint()), YELLOW, RESET)
	return s
}

func (p *permission) edit() {
	if p.cursor == "tool.enabled" {
		if p.permission.AllowInvokeTool() {
			p.permission.Tool = nil
		} else {
			p.permission.Tool = &plugin_entities.PluginPermissionToolRequirement{
				Enabled: true,
			}
		}
	}

	if strings.HasPrefix(p.cursor, "model.") {
		if p.permission.AllowInvokeModel() {
			if p.cursor == "model.enabled" {
				p.permission.Model = nil
				return
			}
		} else {
			p.permission.Model = &plugin_entities.PluginPermissionModelRequirement{
				Enabled: true,
			}
		}
	}

	if p.cursor == "model.llm" {
		if p.permission.AllowInvokeLLM() {
			p.permission.Model.LLM = false
		} else {
			p.permission.Model.LLM = true
		}
	}

	if p.cursor == "model.text_embedding" {
		if p.permission.AllowInvokeTextEmbedding() {
			p.permission.Model.TextEmbedding = false
		} else {
			p.permission.Model.TextEmbedding = true
		}
	}

	if p.cursor == "model.rerank" {
		if p.permission.AllowInvokeRerank() {
			p.permission.Model.Rerank = false
		} else {
			p.permission.Model.Rerank = true
		}
	}

	if p.cursor == "model.tts" {
		if p.permission.AllowInvokeTTS() {
			p.permission.Model.TTS = false
		} else {
			p.permission.Model.TTS = true
		}
	}

	if p.cursor == "model.speech2text" {
		if p.permission.AllowInvokeSpeech2Text() {
			p.permission.Model.Speech2text = false
		} else {
			p.permission.Model.Speech2text = true
		}
	}

	if p.cursor == "model.moderation" {
		if p.permission.AllowInvokeModeration() {
			p.permission.Model.Moderation = false
		} else {
			p.permission.Model.Moderation = true
		}
	}

	if p.cursor == "app.enabled" {
		if p.permission.AllowInvokeApp() {
			p.permission.App = nil
		} else {
			p.permission.App = &plugin_entities.PluginPermissionAppRequirement{
				Enabled: true,
			}
		}
	}

	if p.cursor == "storage.enabled" {
		if p.permission.AllowInvokeStorage() {
			p.permission.Storage = nil
		} else {
			p.permission.Storage = &plugin_entities.PluginPermissionStorageRequirement{
				Enabled: true,
				Size:    1048576,
			}
			p.storageSizeEditor.SetValue(fmt.Sprintf("%d", p.permission.Storage.Size))
		}
	}

	if p.cursor == "endpoint.enabled" {
		if p.permission.AllowRegisterEndpoint() {
			p.permission.Endpoint = nil
		} else {
			p.permission.Endpoint = &plugin_entities.PluginPermissionEndpointRequirement{
				Enabled: true,
			}
		}
	}
}

func (p *permission) updateStorageSize() {
	if p.cursor == "storage.size" {
		// set the storage size editor to the current storage size
		if p.permission.AllowInvokeStorage() {
			p.storageSizeEditor.SetValue(fmt.Sprintf("%d", p.permission.Storage.Size))
			p.storageSizeEditor.Focus()
		}
	} else {
		p.storageSizeEditor.Blur()
		// get the storage size from the editor
		if p.permission.AllowInvokeStorage() {
			p.permission.Storage.Size, _ = strconv.ParseUint(p.storageSizeEditor.Value(), 10, 64)
		}
	}
}

func (p permission) Update(msg tea.Msg) (subMenu, subMenuEvent, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.String() {
		case "ctrl+c":
			return p, SUB_MENU_EVENT_NONE, tea.Quit
		case "down":
			// find the next key in the permissionKeySeq
			for i, key := range permissionKeySeq {
				if key == p.cursor {
					if i == len(permissionKeySeq)-1 {
						p.cursor = permissionKeySeq[0]
					} else {
						p.cursor = permissionKeySeq[i+1]
					}

					p.updateStorageSize()
					break
				}
			}
		case "up":
			// find the previous key in the permissionKeySeq
			for i, key := range permissionKeySeq {
				if key == p.cursor {
					if i == 0 {
						p.cursor = permissionKeySeq[len(permissionKeySeq)-1]
					} else {
						p.cursor = permissionKeySeq[i-1]
					}

					p.updateStorageSize()
					break
				}
			}
		case "tab":
			p.edit()
		case "enter":
			if p.cursor == "endpoint.enabled" {
				p.cursor = permissionKeySeq[0]
				p.updateStorageSize()
				return p, SUB_MENU_EVENT_NEXT, nil
			} else {
				// find the next key in the permissionKeySeq
				for i, key := range permissionKeySeq {
					if key == p.cursor {
						if i == len(permissionKeySeq)-1 {
							p.cursor = permissionKeySeq[0]
						} else {
							p.cursor = permissionKeySeq[i+1]
						}

						p.updateStorageSize()
						break
					}
				}
			}
		}
	}

	// update storage size editor
	if p.cursor == "storage.size" {
		if p.storageSizeEditor.Focused() {
			// check if msg is a number
			model, cmd := p.storageSizeEditor.Update(msg)
			p.storageSizeEditor = model
			return p, SUB_MENU_EVENT_NONE, cmd
		}
	}

	return p, SUB_MENU_EVENT_NONE, nil
}

func (p *permission) UpdatePermission(permission plugin_entities.PluginPermissionRequirement) {
	p.permission = permission
}

func (p permission) Init() tea.Cmd {
	return nil
}

// TODO: optimize implementation
type permissionModel struct {
	permission
}

func (p permissionModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	m, subMenuEvent, cmd := p.permission.Update(msg)
	p.permission = m.(permission)
	if subMenuEvent == SUB_MENU_EVENT_NEXT {
		return p, tea.Quit
	}
	return p, cmd
}

func (p permissionModel) View() string {
	return p.permission.View()
}

func EditPermission(pluginPath string) {
	plugin, err := decoder.NewFSPluginDecoder(pluginPath)
	if err != nil {
		log.Error("decode plugin failed, error: %v", err)
		os.Exit(1)
		return
	}

	manifest, err := plugin.Manifest()
	if err != nil {
		log.Error("get manifest failed, error: %v", err)
		os.Exit(1)
		return
	}

	if manifest.Resource.Permission == nil {
		manifest.Resource.Permission = &plugin_entities.PluginPermissionRequirement{}
	}

	// create a new permission
	m := permissionModel{
		permission: newPermission(*manifest.Resource.Permission),
	}

	p := tea.NewProgram(m)
	if result, err := p.Run(); err != nil {
		fmt.Println("Error running program:", err)
	} else {
		if m, ok := result.(permissionModel); ok {
			// save the manifest
			manifestPath := filepath.Join(pluginPath, "manifest.yaml")
			manifest.Resource.Permission = &m.permission.permission
			if err := writeFile(
				manifestPath,
				string(marshalYamlBytes(manifest.PluginDeclarationWithoutAdvancedFields)),
			); err != nil {
				log.Error("write manifest failed, error: %v", err)
				os.Exit(1)
				return
			}
		} else {
			log.Error("Error running program:", err)
			return
		}
	}
}
