package plugin

import (
	"fmt"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/langgenius/dify-plugin-daemon/pkg/entities/constants"
)

var languages = []constants.Language{
	constants.Python,
	constants.Go + " (not supported yet)",
}

type language struct {
	cursor int
}

func newLanguage() language {
	return language{
		// default language is python
		cursor: 0,
	}
}

func (l language) Language() constants.Language {
	return languages[l.cursor]
}

func (l language) View() string {
	s := `Select the language you want to use for plugin development, and press ` + GREEN + `Enter` + RESET + ` to continue, 
BTW, you need Python 3.12+ to develop the Plugin if you choose Python.
`
	for i, language := range languages {
		if i == l.cursor {
			s += fmt.Sprintf("\033[32m-> %s\033[0m\n", language)
		} else {
			s += fmt.Sprintf("  %s\n", language)
		}
	}
	return s
}

func (l language) Update(msg tea.Msg) (subMenu, subMenuEvent, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.String() {
		case "ctrl+c", "q":
			return l, SUB_MENU_EVENT_NONE, tea.Quit
		case "j", "down":
			l.cursor++
			if l.cursor >= len(languages) {
				l.cursor = len(languages) - 1
			}
		case "k", "up":
			l.cursor--
			if l.cursor < 0 {
				l.cursor = 0
			}
		case "enter":
			if l.cursor != 0 {
				l.cursor = 0
				return l, SUB_MENU_EVENT_NONE, nil
			}

			return l, SUB_MENU_EVENT_NEXT, nil
		}
	}

	return l, SUB_MENU_EVENT_NONE, nil
}

func (l language) Init() tea.Cmd {
	return nil
}

func (l *language) SetLanguage(lang string) {
	for i, v := range languages {
		if string(v) == lang {
			l.cursor = i
			return
		}
	}
}
