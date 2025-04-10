package plugin

import (
	"fmt"

	tea "github.com/charmbracelet/bubbletea"
)

const (
	// Colors
	RESET = "\033[0m"
	BOLD  = "\033[1m"

	// Foreground colors
	GREEN  = "\033[32m"
	YELLOW = "\033[33m"
	BLUE   = "\033[34m"
)
const PLUGIN_GUIDE = `Before starting, here's some basic knowledge about Plugin types in Dify:

` + BOLD + `- Tool` + RESET + `: ` + GREEN + `Tool Providers like Google Search, Stable Diffusion, etc. Used to perform specific tasks.` + RESET + `
` + BOLD + `- Model` + RESET + `: ` + GREEN + `Model Providers like OpenAI, Anthropic, etc. Use their models to enhance AI capabilities.` + RESET + `
` + BOLD + `- Endpoint` + RESET + `: ` + GREEN + `Similar to Service API in Dify and Ingress in Kubernetes. Extend HTTP services as endpoints with custom logic.` + RESET + `
` + BOLD + `- Agent Strategy` + RESET + `: ` + GREEN + `Implement your own agent strategies like Function Calling, ReAct, ToT, CoT, etc.` + RESET + `

Based on the ability you want to extend, Plugins are divided into four types: ` + BOLD + `Tool` + RESET + `, ` + BOLD + `Model` + RESET + `, ` + BOLD + `Extension` + RESET + `, and ` + BOLD + `Agent Strategy` + RESET + `.

` + BOLD + `- Tool` + RESET + `: ` + YELLOW + `A tool provider that can also implement endpoints. For example, building a Discord Bot requires both ` + BLUE + `Sending` + RESET + YELLOW + ` and ` + BLUE + `Receiving Messages` + RESET + YELLOW + `, so both ` + BOLD + `Tool` + RESET + YELLOW + ` and ` + BOLD + `Endpoint` + RESET + YELLOW + ` functionality.` + RESET + `
` + BOLD + `- Model` + RESET + `: ` + YELLOW + `Strictly for model providers, no other extensions allowed.` + RESET + `
` + BOLD + `- Extension` + RESET + `: ` + YELLOW + `For simple HTTP services that extend functionality.` + RESET + `
` + BOLD + `- Agent Strategy` + RESET + `: ` + YELLOW + `Implement custom agent logic with a focused approach.` + RESET + `

We've provided templates to help you get started. Choose one of the options below:
`

type category struct {
	cursor int
}

var categories = []string{
	"tool",
	"agent-strategy",
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
	s := "Select the type of plugin you want to create, and press `Enter` to continue\n"
	s += PLUGIN_GUIDE
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
