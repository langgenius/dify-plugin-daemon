package init

import (
	"fmt"

	tea "github.com/charmbracelet/bubbletea"
)

type category struct {
	cursor int
}

var categories = []string{
	"tool",
	"llm",
	"text-embedding",
	"rerank",
	"tts",
	"speech2text",
	"moderation",
	"extension",
}

func newCategory() category {
	return category{
		// default category is tool
		cursor: 0,
	}
}

func (c category) Category() string {
	return categories[c.cursor]
}

func (c category) View() string {
	s := "Select the type of plugin you want to create\n"
	for i, category := range categories {
		if i == c.cursor {
			s += fmt.Sprintf("\033[32m-> %s\033[0m\n", category)
		} else {
			s += fmt.Sprintf("  %s\n", category)
		}
	}
	return s
}

func (c category) Update(msg tea.Msg) (subMenu, subMenuEvent, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.String() {
		case "ctrl+c", "q":
			return c, SUB_MENU_EVENT_NONE, tea.Quit
		case "j", "down":
			c.cursor++
			if c.cursor >= len(categories) {
				c.cursor = len(categories) - 1
			}
		case "k", "up":
			c.cursor--
			if c.cursor < 0 {
				c.cursor = 0
			}
		case "enter":
			return c, SUB_MENU_EVENT_NEXT, nil
		}
	}

	return c, SUB_MENU_EVENT_NONE, nil
}

func (c category) Init() tea.Cmd {
	return nil
}
