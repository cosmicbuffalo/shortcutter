package internal

import (
	"fmt"
	"strings"

	"github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/sahilm/fuzzy"
)

type model struct {
	shortcuts    []Shortcut
	filtered     []Shortcut
	cursor       int
	query        string
	width        int
	height       int
	selected     *Shortcut
	quitting     bool
}

type tickMsg struct{}

var (
	titleStyle = lipgloss.NewStyle().
			Bold(true).
			Foreground(lipgloss.Color("#7C3AED"))

	selectedStyle = lipgloss.NewStyle().
			Bold(true).
			Foreground(lipgloss.Color("#FFFFFF")).
			Background(lipgloss.Color("#7C3AED"))

	commandStyle = lipgloss.NewStyle().
			Bold(true).
			Foreground(lipgloss.Color("#10B981"))

	descStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("#6B7280"))

	queryStyle = lipgloss.NewStyle().
			Bold(true).
			Foreground(lipgloss.Color("#3B82F6"))

	helpStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("#9CA3AF"))
)

func InitialModel(shortcuts []Shortcut) model {
	return model{
		shortcuts: shortcuts,
		filtered:  shortcuts,
		cursor:    0,
		query:     "",
	}
}

func (m model) Init() tea.Cmd {
	return nil
}

func (m model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		m.width = msg.Width
		m.height = msg.Height
		return m, nil

	case tea.KeyMsg:
		switch msg.String() {
		case "ctrl+c", "esc":
			m.quitting = true
			return m, tea.Quit

		case "enter":
			if len(m.filtered) > 0 && m.cursor < len(m.filtered) {
				m.selected = &m.filtered[m.cursor]
				m.quitting = true
				return m, tea.Quit
			}

		case "up", "k":
			if m.cursor > 0 {
				m.cursor--
			}

		case "down", "j":
			if m.cursor < len(m.filtered)-1 {
				m.cursor++
			}

		case "backspace":
			if len(m.query) > 0 {
				m.query = m.query[:len(m.query)-1]
				m.filtered = m.filterShortcuts()
				m.cursor = 0
			}

		default:
			// Handle regular character input
			if len(msg.String()) == 1 {
				m.query += msg.String()
				m.filtered = m.filterShortcuts()
				m.cursor = 0
			}
		}
	}

	return m, nil
}

func (m model) filterShortcuts() []Shortcut {
	if m.query == "" {
		return m.shortcuts
	}

	// Create search targets for fuzzy matching
	targets := make([]string, len(m.shortcuts))
	for i, shortcut := range m.shortcuts {
		targets[i] = shortcut.Command + " " + shortcut.Description
	}

	// Perform fuzzy search
	matches := fuzzy.Find(m.query, targets)

	// Build filtered results
	filtered := make([]Shortcut, len(matches))
	for i, match := range matches {
		filtered[i] = m.shortcuts[match.Index]
	}

	return filtered
}

func (m model) View() string {
	if m.quitting {
		return ""
	}

	var b strings.Builder

	// Title
	b.WriteString(titleStyle.Render("🚀 Shortcutter"))
	b.WriteString("\n\n")

	// Query line
	b.WriteString("Search: ")
	b.WriteString(queryStyle.Render(m.query))
	if len(m.filtered) > 0 {
		b.WriteString(fmt.Sprintf(" (%d matches)", len(m.filtered)))
	}
	b.WriteString("\n\n")

	// Shortcuts list
	maxVisible := m.height - 8 // Leave space for header, query, and help
	if maxVisible < 5 {
		maxVisible = 5
	}

	start := 0
	if m.cursor >= maxVisible {
		start = m.cursor - maxVisible + 1
	}

	end := start + maxVisible
	if end > len(m.filtered) {
		end = len(m.filtered)
	}

	for i := start; i < end; i++ {
		shortcut := m.filtered[i]
		
		// Calculate column widths
		commandWidth := 25
		if m.width > 80 {
			commandWidth = 35
		}
		
		command := shortcut.Command
		if len(command) > commandWidth {
			command = command[:commandWidth-3] + "..."
		} else {
			command = fmt.Sprintf("%-*s", commandWidth, command)
		}

		description := shortcut.Description
		maxDescWidth := m.width - commandWidth - 10
		if maxDescWidth > 0 && len(description) > maxDescWidth {
			description = description[:maxDescWidth-3] + "..."
		}

		line := fmt.Sprintf("%s  %s", commandStyle.Render(command), descStyle.Render(description))
		
		if i == m.cursor {
			line = selectedStyle.Render(fmt.Sprintf(" %s ", line))
		} else {
			line = fmt.Sprintf(" %s ", line)
		}

		b.WriteString(line)
		b.WriteString("\n")
	}

	// Help text
	b.WriteString("\n")
	b.WriteString(helpStyle.Render("↑/↓ or j/k: navigate • Enter: select • Esc: quit"))

	return b.String()
}

func ShowUI(shortcuts []Shortcut) (*Shortcut, error) {
	m := InitialModel(shortcuts)
	p := tea.NewProgram(m, tea.WithAltScreen())
	
	finalModel, err := p.Run()
	if err != nil {
		return nil, err
	}

	if finalModel, ok := finalModel.(model); ok {
		return finalModel.selected, nil
	}

	return nil, nil
}