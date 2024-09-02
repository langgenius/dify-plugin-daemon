package main

import (
	"fmt"

	ti "github.com/charmbracelet/bubbles/textinput"
	tea "github.com/charmbracelet/bubbletea"
)

type profile struct {
	cursor int
	inputs []ti.Model
}

func newProfile() profile {
	name := ti.New()
	name.Placeholder = "Plugin name, a directory will be created with this name"
	name.CharLimit = 128
	name.Prompt = "Plugin name (press Enter to next step): "
	name.Focus()

	author := ti.New()
	author.Placeholder = "Author name"
	author.CharLimit = 128
	author.Prompt = "Author (press Enter to next step): "

	return profile{
		inputs: []ti.Model{name, author},
	}
}

func (p profile) Name() string {
	return p.inputs[0].Value()
}

func (p profile) Author() string {
	return p.inputs[1].Value()
}

func (p profile) View() string {
	return fmt.Sprintf("Edit profile of the plugin\n%s\n%s\n", p.inputs[0].View(), p.inputs[1].View())
}

func (p profile) Update(msg tea.Msg) (subMenu, subMenuEvent, tea.Cmd) {
	var cmds []tea.Cmd

	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.String() {
		case "ctrl+c", "q":
			return p, SUB_MENU_EVENT_NONE, tea.Quit
		case "down":
			// focus next
			p.cursor++
			if p.cursor >= len(p.inputs) {
				p.cursor = 0
			}
		case "up":
			// focus previous
			p.cursor--
			if p.cursor < 0 {
				p.cursor = len(p.inputs) - 1
			}
		case "enter":
			// submit
			if p.cursor == len(p.inputs)-1 {
				return p, SUB_MENU_EVENT_NEXT, nil
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

	return p, SUB_MENU_EVENT_NONE, tea.Batch(cmds...)
}

func (p profile) Init() tea.Cmd {
	return nil
}
