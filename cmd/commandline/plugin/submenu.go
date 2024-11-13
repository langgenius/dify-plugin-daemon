package plugin

import tea "github.com/charmbracelet/bubbletea"

type subMenuEvent string

const (
	SUB_MENU_EVENT_NEXT subMenuEvent = "next"
	SUB_MENU_EVENT_PREV subMenuEvent = "prev"
	SUB_MENU_EVENT_NONE subMenuEvent = "none"
)

type subMenu interface {
	Init() tea.Cmd

	View() string

	Update(msg tea.Msg) (subMenu, subMenuEvent, tea.Cmd)
}
