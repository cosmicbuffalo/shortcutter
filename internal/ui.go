package internal

import (
	"fmt"
	"os"
	"strings"

	tea "github.com/charmbracelet/bubbletea"
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
	scrollOffset int
	maxVisible   int
}

type tickMsg struct{}

var (
	titleStyle = lipgloss.NewStyle().
			Bold(true).
			Foreground(lipgloss.Color("#10B981"))

	selectedBarStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("#F97316")).
			Background(lipgloss.Color("#2D2D2D"))

	unselectedBarStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("#2D2D2D"))

	selectedLineStyle = lipgloss.NewStyle().
			Background(lipgloss.Color("#2D2D2D"))

	statusStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("#6B7280"))

	separatorStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("#6B7280"))

	matchStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("#3B82F6"))

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
		shortcuts:  shortcuts,
		filtered:   shortcuts,
		cursor:     0,
		query:      "",
		scrollOffset: 0,
		maxVisible: 10,
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

		case "up":
			if m.cursor > 0 {
				m.cursor--
				if m.cursor - m.scrollOffset < 0 {
					m.scrollOffset--
				}

			}

		case "down":
			if m.cursor < len(m.filtered)-1 {
				m.cursor++
				if m.cursor - m.scrollOffset > 9 {
					m.scrollOffset++
				}

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

	case tea.MouseMsg:
		if msg.Type == tea.MouseLeft {
			displayLine := msg.Y - (m.height - 14)
			item := displayLine - 2

			if item >= 0 && item < 10 {
				// If the click is on a valid line, set the cursor to that item
				m.cursor = item + m.scrollOffset
			}

		}
		if msg.Type == tea.MouseWheelUp {
			if m.cursor > 0 {
				m.cursor--
				if m.cursor - m.scrollOffset < 0 {
					m.scrollOffset--
				}
			}
		}
		if msg.Type == tea.MouseWheelDown {
			if m.cursor < len(m.filtered)-1 {
				m.cursor++
				if m.cursor - m.scrollOffset > m.maxVisible - 1 {
					m.scrollOffset++
				}
			}
		}
	}

	return m, nil
}

func (m model) filterShortcuts() []Shortcut {
	if m.query == "" {
		return m.shortcuts
	}

	targets := make([]string, len(m.shortcuts))
	for i, shortcut := range m.shortcuts {
		targets[i] = shortcut.Command + " " + shortcut.Description
	}

	matches := fuzzy.Find(m.query, targets)

	filtered := make([]Shortcut, len(matches))
	for i, match := range matches {
		filtered[i] = m.shortcuts[match.Index]
	}

	return filtered
}

func (m model) highlightMatches(text string, query string, baseStyle lipgloss.Style, isSelected bool) string {
	if query == "" {
		if isSelected {
			return baseStyle.Copy().Background(lipgloss.Color("#2D2D2D")).Render(text)
		}
		return baseStyle.Render(text)
	}

	highlighted := ""
	queryLower := strings.ToLower(query)
	queryIndex := 0

	for _, char := range text {
		charStyle := baseStyle.Copy()
		if isSelected {
			charStyle = charStyle.Background(lipgloss.Color("#2D2D2D"))
		}

		if queryIndex < len(queryLower) && strings.ToLower(string(char)) == string(queryLower[queryIndex]) {
			// This character matches the query - combine base style with match highlighting
			matchChar := charStyle.Foreground(lipgloss.Color("#3B82F6")).Render(string(char))
			highlighted += matchChar
			queryIndex++
		} else {
			highlighted += charStyle.Render(string(char))
		}
	}

	return highlighted
}

func (m model) View() string {
	if m.quitting {
		return ""
	}

	var b strings.Builder

	// Query line (fzf style with >)
	b.WriteString("> ")
	if m.query == "" {
		b.WriteString(queryStyle.Render(""))
	} else {
		b.WriteString(queryStyle.Render(m.query))
	}
	b.WriteString("\n")

	// Status line (fzf style) - no parentheses since we're single-select
	totalCount := len(m.shortcuts)
	filteredCount := len(m.filtered)
	status := fmt.Sprintf("  %d/%d", filteredCount, totalCount)
	b.WriteString(statusStyle.Render(status))
	b.WriteString(" ")

	// Separator line
	separatorLength := m.width - len(status) - 2
	if separatorLength > 0 {
		b.WriteString(separatorStyle.Render(strings.Repeat("─", separatorLength)))
	}
	b.WriteString("\n")

	if m.height > 0 && m.height < 15 {
		m.maxVisible = m.height - 5 // Leave space for query and help
	}
	if m.maxVisible < 5 {
		m.maxVisible = 5
	}

	start := m.scrollOffset

	end := start + m.maxVisible
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

		isSelected := i == m.cursor

		highlightedCommand := m.highlightMatches(command, m.query, commandStyle, isSelected)
		highlightedDesc := m.highlightMatches(description, m.query, descStyle, false)

		// Add fzf-style highlighting and block character
		if isSelected {
			// Selected line: orange left half block + darker gray background only for command area
			barChar := selectedBarStyle.Render("▌")
			spaceBg := lipgloss.NewStyle().Background(lipgloss.Color("#2D2D2D")).Render(" ")
			line := fmt.Sprintf("%s%s%s  %s", barChar, spaceBg, highlightedCommand, highlightedDesc)
			b.WriteString(line)
		} else {
			// Unselected line: gray full block + normal background
			barChar := unselectedBarStyle.Render("█")
			line := fmt.Sprintf("%s %s  %s", barChar, highlightedCommand, highlightedDesc)
			b.WriteString(line)
		}
		b.WriteString("\n")
	}

	// Help text (simplified like fzf)
	b.WriteString("\n")
	b.WriteString(helpStyle.Render("↑/↓: navigate • Enter: select • Esc: quit"))

	return b.String()
}

func ShowUI(shortcuts []Shortcut) (*Shortcut, error) {
	m := InitialModel(shortcuts)

	// Open TTY for input/output
	tty, err := os.OpenFile("/dev/tty", os.O_RDWR, 0)
	if err != nil {
		// Fallback to stdin/stdout if TTY not available
		p := tea.NewProgram(m, tea.WithMouseAllMotion())
		finalModel, err := p.Run()
		if err != nil {
			return nil, err
		}

		if finalModel, ok := finalModel.(model); ok {
			return finalModel.selected, nil
		}
		return nil, nil
	}
	defer tty.Close()

	p := tea.NewProgram(m, tea.WithMouseAllMotion(), tea.WithInput(tty), tea.WithOutput(tty))

	finalModel, err := p.Run()
	if err != nil {
		return nil, err
	}

	if finalModel, ok := finalModel.(model); ok {
		return finalModel.selected, nil
	}

	return nil, nil
}
