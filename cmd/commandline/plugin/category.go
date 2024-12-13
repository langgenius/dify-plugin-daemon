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

const PLUGIN_GUIDE = `But before starting, you need some basic knowledge about the Plugin types, Plugin supports to extend the following abilities in Dify:
` + "\n" + BOLD + `- Tool` + RESET + `: ` + GREEN + `Tool Providers like Google Search, Stable Diffusion, etc. it can be used to perform a specific task.` + RESET + `
` + BOLD + `- Model` + RESET + `: ` + GREEN + `Model Providers like OpenAI, Anthropic, etc. you can use their models to enhance the AI capabilities.` + RESET + `
` + BOLD + `- Endpoint` + RESET + `: ` + GREEN + `Like Service API in Dify and Ingress in Kubernetes, you can extend a http service as an endpoint and control its logics using your own code.` + RESET + `
` + BOLD + `- Agent Strategy` + RESET + `: ` + GREEN + `You can implement your own agent strategy like Function Calling, ReAct, ToT, Cot, etc. anyway you want.` + RESET + `

Based on the ability you want to extend, we have divided the Plugin into four types: ` + BOLD + `Tool` + RESET + `, ` + BOLD + `Model` + RESET + `, ` + BOLD + `Extension` + RESET + `, and ` + BOLD + `Agent Strategy` + RESET + `.

` + BOLD + `- Tool` + RESET + `: ` + YELLOW + `It's a tool provider, but not only limited to tools, you can implement an endpoint there, for example, you need both ` + BLUE + `Sending Message` + RESET + YELLOW + ` and ` + BLUE + `Receiving Message` + RESET + YELLOW + ` if you are building a Discord Bot, ` + BOLD + `Tool` + RESET + YELLOW + ` and ` + BOLD + `Endpoint` + RESET + YELLOW + ` are both required.` + RESET + `
` + BOLD + `- Model` + RESET + `: ` + YELLOW + `Just a model provider, extending others is not allowed.` + RESET + `
` + BOLD + `- Extension` + RESET + `: ` + YELLOW + `Other times, you may only need a simple http service to extend the functionalities, ` + BOLD + `Extension` + RESET + YELLOW + ` is the right choice for you.` + RESET + `
` + BOLD + `- Agent Strategy` + RESET + `: ` + YELLOW + `Implement your own logics here, just by focusing on Agent itself` + RESET + `

What's more, we have provided the template for you, you can choose one of them below:
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
