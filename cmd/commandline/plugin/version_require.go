package plugin

import (
	"fmt"

	ti "github.com/charmbracelet/bubbles/textinput"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/langgenius/dify-plugin-daemon/pkg/entities/manifest_entities"
)

type versionRequire struct {
	minimalDifyVersion ti.Model

	warning string
}

func newVersionRequire() versionRequire {
	minimalDifyVersion := ti.New()
	minimalDifyVersion.Placeholder = "Minimal Dify version"
	minimalDifyVersion.CharLimit = 128
	minimalDifyVersion.Prompt = "Minimal Dify version (press Enter to next step): "
	minimalDifyVersion.Focus()

	return versionRequire{
		minimalDifyVersion: minimalDifyVersion,
	}
}

func (p versionRequire) MinimalDifyVersion() string {
	return p.minimalDifyVersion.Value()
}

func (p versionRequire) View() string {
	s := fmt.Sprintf("Edit minimal Dify version requirement, leave it blank by default\n%s\n", p.minimalDifyVersion.View())
	if p.warning != "" {
		s += fmt.Sprintf("\033[31m%s\033[0m\n", p.warning)
	}
	return s
}

func (p *versionRequire) checkRule() bool {
	if p.minimalDifyVersion.Value() == "" {
		p.warning = ""
		return true
	}

	_, err := manifest_entities.NewVersion(p.minimalDifyVersion.Value())
	if err != nil {
		p.warning = "Invalid minimal Dify version"
		return false
	}

	p.warning = ""
	return true
}

func (p versionRequire) Update(msg tea.Msg) (subMenu, subMenuEvent, tea.Cmd) {
	var cmds []tea.Cmd

	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.String() {
		case "ctrl+c":
			return p, SUB_MENU_EVENT_NONE, tea.Quit
		case "enter":
			// check if empty
			if !p.checkRule() {
				return p, SUB_MENU_EVENT_NONE, nil
			}
			return p, SUB_MENU_EVENT_NEXT, nil
		}
	}

	// update view
	var cmd tea.Cmd
	p.minimalDifyVersion, cmd = p.minimalDifyVersion.Update(msg)
	if cmd != nil {
		cmds = append(cmds, cmd)
	}

	return p, SUB_MENU_EVENT_NONE, tea.Batch(cmds...)
}

func (p versionRequire) Init() tea.Cmd {
	return nil
}

func (p *versionRequire) SetMinimalDifyVersion(version string) {
	p.minimalDifyVersion.SetValue(version)
}
