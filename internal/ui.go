package internal

import (
	"fmt"
	"os"
	"strings"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/muesli/termenv"
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
	styles       ThemeStyles
}

type tickMsg struct{}

func InitialModel(shortcuts []Shortcut, styles ThemeStyles) model {
	return model{
		shortcuts:    shortcuts,
		filtered:     shortcuts,
		cursor:       0,
		query:        "",
		scrollOffset: 0,
		maxVisible:   10,
		styles:       styles,
	}
}

func (m model) Shortcuts() []Shortcut {
	return m.shortcuts
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
				if m.cursor-m.scrollOffset < 0 {
					m.scrollOffset--
				}

			}

		case "down":
			if m.cursor < len(m.filtered)-1 {
				m.cursor++
				if m.cursor-m.scrollOffset > 9 {
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
			if len(msg.String()) == 1 {
				m.query += msg.String()
				m.filtered = m.filterShortcuts()
				m.cursor = 0
				m.scrollOffset = 0
			}
		}

	case tea.MouseMsg:
		if msg.Type == tea.MouseLeft {
			displayLine := msg.Y - (m.height - 14)
			item := displayLine - 2

			if item >= 0 && item < 10 {
				m.cursor = item + m.scrollOffset
			}

		}
		if msg.Type == tea.MouseWheelUp {
			if m.cursor > 0 {
				m.cursor--
				if m.cursor-m.scrollOffset < 0 {
					m.scrollOffset--
				}
			}
		}
		if msg.Type == tea.MouseWheelDown {
			if m.cursor < len(m.filtered)-1 {
				m.cursor++
				if m.cursor-m.scrollOffset > m.maxVisible-1 {
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

func (m model) highlightMatches(text string, query string, baseStyle lipgloss.Style, isSelected bool, styles ThemeStyles) string {
	if query == "" {
		if isSelected {
			return baseStyle.Copy().Background(styles.SelectedBar.GetBackground()).Render(text)
		}
		return baseStyle.Render(text)
	}

	highlighted := ""
	unhighlighted := ""
	queryLower := strings.ToLower(query)
	queryIndex := 0
	maxMatchLength := 0
	currentMatchLength := 0

	for _, char := range text {
		charStyle := baseStyle.Copy()
		if isSelected {
			charStyle = charStyle.Background(m.styles.SelectedLine.GetBackground())
		}

		if queryIndex < len(queryLower) && strings.ToLower(string(char)) == string(queryLower[queryIndex]) {
			matchChar := charStyle.Foreground(m.styles.Match.GetForeground()).Render(string(char))
			highlighted += matchChar
			queryIndex++
			currentMatchLength++
			if currentMatchLength > maxMatchLength {
				maxMatchLength = currentMatchLength
			}
		} else {
			currentMatchLength = 0
			highlighted += charStyle.Render(string(char))
		}
		unhighlighted += charStyle.Render(string(char))
	}

	matchDiff := len(query) - maxMatchLength
	if matchDiff < 2 {
		return highlighted
	}

	return unhighlighted
}

func (m model) View() string {
	if m.quitting {
		return ""
	}

	var a strings.Builder
	var b strings.Builder

	a.WriteString(m.styles.Query.Render("❯ "))
	a.WriteString(m.styles.Query.Render(m.query))
	a.WriteString("\n")

	totalCount := len(m.shortcuts)
	filteredCount := len(m.filtered)
	status := fmt.Sprintf("  %d/%d ", filteredCount, totalCount)
	b.WriteString(m.styles.Status.Render(status))

	separatorLength := m.width - len(status) - 2
	if separatorLength > 0 {
		b.WriteString(m.styles.Separator.Render(strings.Repeat("─", separatorLength)))
	}
	b.WriteString("\n")

	if m.height > 0 && m.height < 15 {
		m.maxVisible = m.height - 5
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
		commandWidth := 22
		indicatorWidth := 3
		if m.width > 80 {
			commandWidth = 30
		}

		command := shortcut.Command
		if len(command) > commandWidth {
			command = command[:commandWidth-3] + "..."
		} else {
			command = fmt.Sprintf("%-*s", commandWidth, command)
		}

		description := shortcut.Description
		maxDescWidth := m.width - commandWidth - indicatorWidth - 12
		if maxDescWidth > 0 && len(description) > maxDescWidth {
			description = description[:maxDescWidth-3] + "..."
		}

		customIndicator := m.styles.AppBackground.Render(" ")
		if shortcut.IsCustom {
			customIndicator = m.styles.CustomIndicator.Render("*")
		}

		isSelected := i == m.cursor

		highlightedCommand := m.highlightMatches(command, m.query, m.styles.Command, isSelected, m.styles)
		highlightedDesc := m.highlightMatches(description, m.query, m.styles.Description, false, m.styles)

		if isSelected {
			barChar := m.styles.SelectedBar.Render("▌")
			spaceBg := m.styles.SelectedLine.Render(" ")
			columnBg := m.styles.AppBackground.Render("  ")
			line := fmt.Sprintf("%s%s%s%s%s%s", barChar, spaceBg, highlightedCommand, columnBg, highlightedDesc, customIndicator)
			b.WriteString(m.styles.AppBackground.Render(line))
		} else {
			barChar := m.styles.UnselectedBar.Render("█")
			spaceBg := m.styles.AppBackground.Render(" ")
			columnBg := m.styles.AppBackground.Render("  ")
			line := fmt.Sprintf("%s%s%s%s%s%s", barChar, spaceBg, highlightedCommand, columnBg, highlightedDesc, customIndicator)
			b.WriteString(line)
		}
		b.WriteString("\n")
	}

	b.WriteString("\n")
	b.WriteString(m.styles.Help.Render("↑/↓: navigate • Enter: select • Esc: quit"))

	a.WriteString(m.styles.AppBackground.Render(b.String()))
	return a.String()
	// return content
	// return m.styles.AppBackground.Render(content)
}

func ShowUI(shortcuts []Shortcut, styles ThemeStyles) (*Shortcut, error) {
	// Force true color support
	lipgloss.SetColorProfile(termenv.TrueColor)
	
	m := InitialModel(shortcuts, styles)

	tty, err := os.OpenFile("/dev/tty", os.O_RDWR, 0)
	if err != nil {
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
