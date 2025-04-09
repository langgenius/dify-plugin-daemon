package bundle

import (
	"fmt"

	ti "github.com/charmbracelet/bubbles/textinput"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/langgenius/dify-plugin-daemon/pkg/entities/plugin_entities"
)

type profile struct {
	cursor int
	inputs []ti.Model

	warning string
}

func newProfile() profile {
	name := ti.New()
	name.Placeholder = "Bundle name, a directory will be created with this name"
	name.CharLimit = 128
	name.Prompt = "Bundle name (press Enter to next step): "
	name.Focus()

	author := ti.New()
	author.Placeholder = "Author name"
	author.CharLimit = 128
	author.Prompt = "Author (press Enter to next step): "

	description := ti.New()
	description.Placeholder = "Description"
	description.CharLimit = 1024
	description.Prompt = "Description (press Enter to next step): "

	return profile{
		inputs: []ti.Model{name, author, description},
	}
}

func (p profile) Name() string {
	return p.inputs[0].Value()
}

func (p profile) Author() string {
	return p.inputs[1].Value()
}

func (p profile) Description() string {
	return p.inputs[2].Value()
}

func (p profile) View() string {
	s := fmt.Sprintf("Edit profile of the bundle\n%s\n%s\n%s\n", p.inputs[0].View(), p.inputs[1].View(), p.inputs[2].View())
	if p.warning != "" {
		s += fmt.Sprintf("\033[31m%s\033[0m\n", p.warning)
	}
	return s
}

func (p *profile) checkRule() bool {
	if p.inputs[p.cursor].Value() == "" {
		p.warning = "Name, author and description cannot be empty"
		return false
	} else if p.cursor == 0 && !plugin_entities.PluginNameRegex.MatchString(p.inputs[p.cursor].Value()) {
		p.warning = "Bundle name must be 1-128 characters long, and can only contain lowercase letters, numbers, dashes and underscores"
		return false
	} else if p.cursor == 1 && !plugin_entities.AuthorRegex.MatchString(p.inputs[p.cursor].Value()) {
		p.warning = "Author name must be 1-64 characters long, and can only contain lowercase letters, numbers, dashes and underscores"
		return false
	} else {
		p.warning = ""
	}
	return true
}

func (p profile) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	var cmds []tea.Cmd

	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.String() {
		case "ctrl+c":
			return p, tea.Quit
		case "down":
			// check if empty
			if !p.checkRule() {
				return p, nil
			}

			// focus next
			p.cursor++
			if p.cursor >= len(p.inputs) {
				p.cursor = 0
			}
		case "up":
			if !p.checkRule() {
				return p, nil
			}

			p.cursor--
			if p.cursor < 0 {
				p.cursor = len(p.inputs) - 1
			}
		case "enter":
			if !p.checkRule() {
				return p, nil
			}

			// submit
			if p.cursor == len(p.inputs)-1 {
				return p, tea.Quit
			}
			// move to next
			p.cursor++
		}
	}

	// update cursor
	for i := 0; i < len(p.inputs); i++ {
		if i == p.cursor {
			p.inputs[i].Focus()
		} else {
			p.inputs[i].Blur()
		}
	}

	// update view
	for i := range p.inputs {
		var cmd tea.Cmd
		p.inputs[i], cmd = p.inputs[i].Update(msg)
		if cmd != nil {
			cmds = append(cmds, cmd)
		}
	}

	return p, tea.Batch(cmds...)
}

func (p profile) Init() tea.Cmd {
	return nil
}
