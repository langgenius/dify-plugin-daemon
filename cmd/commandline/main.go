package main

import (
	"fmt"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/langgenius/dify-plugin-daemon/internal/utils/log"
)

func main() {
	m := initialize()
	p := tea.NewProgram(m)
	if result, err := p.Run(); err != nil {
		fmt.Println("Error running program:", err)
	} else {
		if m, ok := result.(model); ok {
			m.createPlugin()
		} else {
			log.Error("Error running program:", err)
			return
		}
	}
}
