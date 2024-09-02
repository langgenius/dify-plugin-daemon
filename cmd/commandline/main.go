package main

import (
	"fmt"

	tea "github.com/charmbracelet/bubbletea"
)

type subMenuKey string

const (
	SUB_MENU_KEY_PROFILE    subMenuKey = "profile"
	SUB_MENU_KEY_LANGUAGE   subMenuKey = "language"
	SUB_MENU_KEY_PERMISSION subMenuKey = "permission"
)

type model struct {
	subMenus       map[subMenuKey]subMenu
	subMenuSeq     []subMenuKey
	currentSubMenu subMenuKey
}

func initialize() model {
	m := model{}
	m.subMenus = map[subMenuKey]subMenu{
		SUB_MENU_KEY_PROFILE:    newProfile(),
		SUB_MENU_KEY_LANGUAGE:   newLanguage(),
		SUB_MENU_KEY_PERMISSION: newPermission(),
	}
	m.currentSubMenu = SUB_MENU_KEY_PROFILE

	m.subMenuSeq = []subMenuKey{
		SUB_MENU_KEY_PROFILE,
		SUB_MENU_KEY_LANGUAGE,
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

func main() {
	p := tea.NewProgram(initialize())
	if _, err := p.Run(); err != nil {
		fmt.Println("Error running program:", err)
	}
}
