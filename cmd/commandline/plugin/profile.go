package plugin

import (
	"fmt"
	"strings"

	ti "github.com/charmbracelet/bubbles/textinput"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/langgenius/dify-plugin-daemon/pkg/entities/plugin_entities"
)

type languageChoice struct {
	code     string
	name     string
	selected bool
	required bool // English is required by default
}

type profile struct {
	cursor int
	inputs []ti.Model

	enableI18nReadme bool
	languageChoices  []languageChoice
	languageSection  bool // true when in language selection section
	warning          string
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

	description := ti.New()
	description.Placeholder = "Description"
	description.CharLimit = 1024
	description.Prompt = "Description (press Enter to next step): "

	repo := ti.New()
	repo.Placeholder = "Repository URL (Optional)"
	repo.CharLimit = 128
	repo.Prompt = "Repository URL (Optional) (press Enter to next step): "

	return profile{
		inputs:           []ti.Model{name, author, description, repo},
		enableI18nReadme: false,
		languageChoices: []languageChoice{
			{code: "en", name: "English", selected: true, required: true},
			{code: "zh_Hans", name: "简体中文 (Simplified Chinese)", selected: false, required: false},
			{code: "ja_JP", name: "日本語 (Japanese)", selected: false, required: false},
			{code: "pt_BR", name: "Português (Portuguese - Brazil)", selected: false, required: false},
		},
		languageSection: false,
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

func (p profile) Repo() string {
	return p.inputs[3].Value()
}

func (p profile) EnableI18nReadme() bool {
	return p.enableI18nReadme
}

func (p profile) SelectedLanguages() []string {
	if !p.enableI18nReadme {
		return []string{"en"} // Only English if i18n is disabled
	}

	var selected []string
	for _, choice := range p.languageChoices {
		if choice.selected {
			selected = append(selected, choice.code)
		}
	}
	return selected
}

func (p profile) View() string {
	var s strings.Builder

	s.WriteString("Edit profile of the plugin\n")
	s.WriteString(fmt.Sprintf("%s\n%s\n%s\n%s\n", p.inputs[0].View(), p.inputs[1].View(), p.inputs[2].View(), p.inputs[3].View()))

	// Cursor helper function
	cursor := func(isSelected bool) string {
		if isSelected {
			return "→ "
		}
		return "  "
	}

	// Checkbox helper function
	checked := func(enabled bool) string {
		if enabled {
			return fmt.Sprintf("\033[32m%s\033[0m", "[✔]")
		}
		return fmt.Sprintf("\033[31m%s\033[0m", "[✘]")
	}

	// Add i18n readme checkbox
	s.WriteString(fmt.Sprintf("%sEnable multilingual README: %s \033[33mEnglish is required by default\033[0m\n",
		cursor(p.cursor == 4 && !p.languageSection),
		checked(p.enableI18nReadme)))

	// Show language selection if i18n is enabled
	if p.enableI18nReadme {
		s.WriteString("\nLanguages to generate:\n")
		for i, choice := range p.languageChoices {
			isCurrentCursor := p.languageSection && p.cursor == i

			statusText := ""
			if choice.required {
				statusText = " \033[33m(required)\033[0m"
			}

			s.WriteString(fmt.Sprintf("  %s%s: %s%s%s\n",
				cursor(isCurrentCursor),
				choice.name,
				checked(choice.selected),
				statusText,
				"\033[0m"))
		}
	}

	// Add operation hints
	s.WriteString("\n\033[36mControls:\033[0m\n")
	s.WriteString("  ↑/↓ Navigate • Space/Tab Toggle selection • Enter Next step\n")

	if p.warning != "" {
		s.WriteString(fmt.Sprintf("\n\033[31m%s\033[0m\n", p.warning))
	}

	return s.String()
}

func (p *profile) checkRule() bool {
	if p.cursor >= 0 && p.cursor <= 2 && p.inputs[p.cursor].Value() == "" {
		p.warning = "Name, author and description cannot be empty"
		return false
	} else if p.cursor == 0 && !plugin_entities.PluginNameRegex.MatchString(p.inputs[p.cursor].Value()) {
		p.warning = "Plugin name must be 1-128 characters long, and can only contain lowercase letters, numbers, dashes and underscores"
		return false
	} else if p.cursor == 1 && !plugin_entities.AuthorRegex.MatchString(p.inputs[p.cursor].Value()) {
		p.warning = "Author name must be 1-64 characters long, and can only contain lowercase letters, numbers, dashes and underscores"
		return false
	} else {
		p.warning = ""
	}
	return true
}

func (p profile) Update(msg tea.Msg) (subMenu, subMenuEvent, tea.Cmd) {
	var cmds []tea.Cmd

	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.String() {
		case "ctrl+c":
			return p, SUB_MENU_EVENT_NONE, tea.Quit
		case "down":
			if p.languageSection {
				// In language selection section
				p.cursor++
				if p.cursor >= len(p.languageChoices) {
					p.cursor = len(p.languageChoices) - 1
				}
			} else {
				// In main form section
				if p.cursor <= 3 && !p.checkRule() {
					return p, SUB_MENU_EVENT_NONE, nil
				}

				p.cursor++
				if p.enableI18nReadme && p.cursor == 5 {
					// Move to language selection
					p.languageSection = true
					p.cursor = 0
				} else if p.cursor > 4 {
					p.cursor = 0
				}
			}
		case "up":
			if p.languageSection {
				// In language selection section
				p.cursor--
				if p.cursor < 0 {
					// Move back to checkbox
					p.languageSection = false
					p.cursor = 4
				}
			} else {
				// In main form section
				if p.cursor <= 3 && !p.checkRule() {
					return p, SUB_MENU_EVENT_NONE, nil
				}

				p.cursor--
				if p.cursor < 0 {
					if p.enableI18nReadme {
						// Move to last language option
						p.languageSection = true
						p.cursor = len(p.languageChoices) - 1
					} else {
						p.cursor = 4
					}
				}
			}
		case "enter":
			if p.languageSection {
				// In language selection, enter means finish
				return p, SUB_MENU_EVENT_NEXT, nil
			}

			if p.cursor == 4 {
				// Toggle checkbox for i18n readme
				p.enableI18nReadme = !p.enableI18nReadme
				if !p.enableI18nReadme {
					// Reset language selections to default when disabled
					for i := range p.languageChoices {
						p.languageChoices[i].selected = p.languageChoices[i].required
					}
				}
				return p, SUB_MENU_EVENT_NONE, nil
			}

			if p.cursor <= 3 && !p.checkRule() {
				return p, SUB_MENU_EVENT_NONE, nil
			}

			// submit
			if p.cursor == 3 && p.inputs[p.cursor].Value() == "" {
				// repo is optional, move to checkbox
				p.cursor = 4
				return p, SUB_MENU_EVENT_NONE, nil
			}

			if p.cursor == 3 {
				p.cursor = 4
				return p, SUB_MENU_EVENT_NONE, nil
			}
			// move to next
			p.cursor++
		case " ", "tab":
			if p.languageSection {
				// Toggle language selection (but not for required ones)
				if !p.languageChoices[p.cursor].required {
					p.languageChoices[p.cursor].selected = !p.languageChoices[p.cursor].selected
				}
			} else if p.cursor == 4 {
				// Toggle checkbox with space/tab when focused on it
				p.enableI18nReadme = !p.enableI18nReadme
				if !p.enableI18nReadme {
					// Reset language selections to default when disabled
					for i := range p.languageChoices {
						p.languageChoices[i].selected = p.languageChoices[i].required
					}
				}
			}
		}
	}

	// update cursor (only for input fields)
	if !p.languageSection {
		for i := 0; i < len(p.inputs); i++ {
			if i == p.cursor {
				p.inputs[i].Focus()
			} else {
				p.inputs[i].Blur()
			}
		}
	}

	// update view (only for input fields when not in language section)
	if !p.languageSection && p.cursor <= 3 {
		for i := range p.inputs {
			var cmd tea.Cmd
			p.inputs[i], cmd = p.inputs[i].Update(msg)
			if cmd != nil {
				cmds = append(cmds, cmd)
			}
		}
	}

	return p, SUB_MENU_EVENT_NONE, tea.Batch(cmds...)
}

func (p profile) Init() tea.Cmd {
	return nil
}

func (p *profile) SetAuthor(author string) {
	p.inputs[1].SetValue(author)
}

func (p *profile) SetName(name string) {
	p.inputs[0].SetValue(name)
}

func (p *profile) SetEnableI18nReadme(enable bool) {
	p.enableI18nReadme = enable
}

func (p *profile) SetSelectedLanguages(languages []string) {
	// Reset all selections except required ones
	for i := range p.languageChoices {
		if !p.languageChoices[i].required {
			p.languageChoices[i].selected = false
		}
	}

	// Set selected languages
	for _, lang := range languages {
		for i, choice := range p.languageChoices {
			if choice.code == lang {
				p.languageChoices[i].selected = true
				break
			}
		}
	}
}
