package cmd

import (
	"fmt"
	"strconv"
	"strings"

	ti "github.com/charmbracelet/bubbles/textinput"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/langgenius/dify-plugin-daemon/internal/types/entities/plugin_entities"
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

func newPermission() permission {
	return permission{
		cursor:            permissionKeySeq[0],
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

	s := "Configure the permissions of the plugin, use up and down to navigate, enter to select, after selection, press right to move to the next menu\n"
	s += "Backwards Invocation:\n"
	s += "Tools:\n"
	s += fmt.Sprintf("  %sEnabled: %v\n", cursor("tool.enabled"), checked(p.permission.AllowInvokeTool()))
	s += "Models:\n"
	s += fmt.Sprintf("  %sEnabled: %v\n", cursor("model.enabled"), checked(p.permission.AllowInvokeModel()))
	s += fmt.Sprintf("  %sLLM: %v\n", cursor("model.llm"), checked(p.permission.AllowInvokeLLM()))
	s += fmt.Sprintf("  %sText Embedding: %v\n", cursor("model.text_embedding"), checked(p.permission.AllowInvokeTextEmbedding()))
	s += fmt.Sprintf("  %sRerank: %v\n", cursor("model.rerank"), checked(p.permission.AllowInvokeRerank()))
	s += fmt.Sprintf("  %sTTS: %v\n", cursor("model.tts"), checked(p.permission.AllowInvokeTTS()))
	s += fmt.Sprintf("  %sSpeech2Text: %v\n", cursor("model.speech2text"), checked(p.permission.AllowInvokeSpeech2Text()))
	s += fmt.Sprintf("  %sModeration: %v\n", cursor("model.moderation"), checked(p.permission.AllowInvokeModeration()))
	s += "Apps:\n"
	s += fmt.Sprintf("  %sEnabled: %v\n", cursor("app.enabled"), checked(p.permission.AllowInvokeApp()))
	s += "Resources:\n"
	s += "Storage:\n"
	s += fmt.Sprintf("  %sEnabled: %v\n", cursor("storage.enabled"), checked(p.permission.AllowInvokeStorage()))

	if p.permission.AllowInvokeStorage() {
		s += fmt.Sprintf("  %sSize: %v\n", cursor("storage.size"), p.storageSizeEditor.View())
	} else {
		s += fmt.Sprintf("  %sSize: %v\n", cursor("storage.size"), "N/A")
	}

	s += "Endpoints:\n"
	s += fmt.Sprintf("  %sEnabled: %v\n", cursor("endpoint.enabled"), checked(p.permission.AllowRegistryEndpoint()))
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
		}
	}

	if p.cursor == "endpoint.enabled" {
		if p.permission.AllowRegistryEndpoint() {
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
		case "enter":
			p.edit()
		case "right":
			if p.cursor == "storage.size" {
				break
			}
			p.cursor = permissionKeySeq[0]
			p.updateStorageSize()
			return p, SUB_MENU_EVENT_NEXT, nil
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

func (p permission) Init() tea.Cmd {
	return nil
}
