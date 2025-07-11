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
	selectedKey  string // "enter" or "tab"
	quitting     bool
	scrollOffset int
	maxVisible   int
	styles       ThemeStyles
	// Expanded mode fields
	expandedMode         bool
	expandedScrollOffset int
	expandedText         []string // lines of the expanded description
}

type tickMsg struct{}

func InitialModel(shortcuts []Shortcut, styles ThemeStyles) model {
	return model{
		shortcuts:            shortcuts,
		filtered:             shortcuts,
		cursor:               0,
		query:                "",
		scrollOffset:         0,
		maxVisible:           10,
		styles:               styles,
		expandedMode:         false,
		expandedScrollOffset: 0,
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
				m.selectedKey = "enter"
				m.quitting = true
				return m, tea.Quit
			}

		case "tab":
			if len(m.filtered) > 0 && m.cursor < len(m.filtered) {
				m.selected = &m.filtered[m.cursor]
				m.selectedKey = "tab"
				m.quitting = true
				return m, tea.Quit
			}

		case "up":
			if m.cursor > 0 {
				m.cursor--
				if m.cursor-m.scrollOffset < 0 {
					m.scrollOffset--
				}
				if m.expandedMode {
					m.prepareExpandedText()
				}
			}

		case "down":
			if m.cursor < len(m.filtered)-1 {
				m.cursor++
				if m.cursor-m.scrollOffset > 9 {
					m.scrollOffset++
				}
				if m.expandedMode {
					m.prepareExpandedText()
				}
			}

		case "left":
			if m.expandedMode {
				// Exit expanded mode
				m.expandedMode = false
				m.expandedScrollOffset = 0
				m.expandedText = nil
			}

		case "right":
			if !m.expandedMode && len(m.filtered) > 0 && m.cursor < len(m.filtered) {
				// Enter expanded mode
				m.expandedMode = true
				m.expandedScrollOffset = 0
				m.prepareExpandedText()
			}

		case "ctrl+d":
			if m.expandedMode {
				maxScroll := len(m.expandedText) - m.getExpandedVisibleLines()
				if m.expandedScrollOffset < maxScroll {
					m.expandedScrollOffset++
				}
			}

		case "ctrl+u":
			if m.expandedMode {
				if m.expandedScrollOffset > 0 {
					m.expandedScrollOffset--
				}
			}

		case "backspace":
			if len(m.query) > 0 {
				m.query = m.query[:len(m.query)-1]
				m.filtered = m.filterShortcuts()
				m.cursor = 0
				m.updateExpandedMode()
			}

		default:
			// Handle printable characters (including rapid typing)
			for _, r := range msg.Runes {
				if r >= 32 && r < 127 { // printable ASCII range
					m.query += string(r)
				}
			}
			if len(msg.Runes) > 0 {
				m.filtered = m.filterShortcuts()
				m.cursor = 0
				m.scrollOffset = 0
				m.updateExpandedMode()
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
		targets[i] = shortcut.Display + " " + shortcut.Description
	}

	matches := fuzzy.Find(m.query, targets)

	filtered := make([]Shortcut, len(matches))
	for i, match := range matches {
		filtered[i] = m.shortcuts[match.Index]
	}

	return filtered
}

func (m *model) updateExpandedMode() {
	if !m.expandedMode {
		return
	}

	m.prepareExpandedText()
	m.expandedScrollOffset = 0
}

// prepareExpandedText splits the full description into lines for display
func (m *model) prepareExpandedText() {
	if len(m.filtered) == 0 || m.cursor >= len(m.filtered) {
		m.expandedText = []string{"No description available"}
		return
	}

	shortcut := m.filtered[m.cursor]
	fullDesc := shortcut.FullDescription
	if fullDesc == "" {
		fullDesc = shortcut.Description
	}
	if fullDesc == "" {
		fullDesc = "No description available"
	}

	// Use right column width for wrapping
	rightWidth := m.width / 2
	maxWidth := rightWidth - 4 // Account for padding
	if maxWidth < 20 {
		maxWidth = 20
	}

	m.expandedText = m.wrapText(fullDesc, maxWidth)
}

// wrapText breaks text into lines that fit within the specified width
func (m *model) wrapText(text string, maxWidth int) []string {
	words := strings.Fields(text)
	if len(words) == 0 {
		return []string{text}
	}

	lines := []string{}
	currentLine := ""

	for _, word := range words {
		if currentLine == "" {
			currentLine = word
		} else if len(currentLine)+1+len(word) <= maxWidth {
			currentLine += " " + word
		} else {
			lines = append(lines, currentLine)
			currentLine = word
		}
	}

	if currentLine != "" {
		lines = append(lines, currentLine)
	}

	return lines
}

// getExpandedVisibleLines calculates how many lines can be shown in expanded mode
func (m *model) getExpandedVisibleLines() int {
	return m.maxVisible
	// // Reserve space for query line, widget name, borders, and help
	// availableHeight := m.height - 6
	// if availableHeight < 3 {
	// 	availableHeight = 3
	// }
	// return availableHeight
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

	// Always render the main view, but with split layout if expanded
	return m.renderSplitView()
}

// renderSplitView renders the main view with optional right pane
func (m model) renderSplitView() string {
	var result strings.Builder

	// Calculate column widths dynamically based on terminal width
	// Use 20% for commands, 80% for descriptions with minimum widths
	minLeftWidth := 20
	minRightWidth := 30
	
	leftWidth := int(float64(m.width) * 0.2)
	if leftWidth < minLeftWidth {
		leftWidth = minLeftWidth
	}
	
	rightWidth := m.width - leftWidth
	if rightWidth < minRightWidth {
		// If terminal is too narrow, prioritize description column
		rightWidth = minRightWidth
		leftWidth = m.width - rightWidth
		if leftWidth < minLeftWidth {
			leftWidth = minLeftWidth
		}
	}

	// Query line (spans full width)
	result.WriteString(m.styles.Query.Render("❯ "))
	result.WriteString(m.styles.Query.Render(m.query))
	result.WriteString("\n")

	// Render each line with left and optionally right content
	lines := m.renderContentLines(leftWidth, rightWidth)
	for _, line := range lines {
		result.WriteString(line)
		result.WriteString("\n")
	}

	return result.String()
}

// renderContentLines generates all content lines for the split layout
func (m model) renderContentLines(leftWidth int, rightWidth int) []string {
	var lines []string

	// Status line
	totalCount := len(m.shortcuts)
	filteredCount := len(m.filtered)
	status := fmt.Sprintf("  %d/%d ", filteredCount, totalCount)
	statusLine := m.styles.Status.Render(status)

	// Add right pane header if in expanded mode, otherwise fill with separator
	if m.expandedMode {
		separatorLength := leftWidth - len(status) - 2
		if separatorLength > 0 {
			statusLine += m.styles.Separator.Render(strings.Repeat("─", separatorLength))
		}
		widgetName := ""
		if len(m.filtered) > 0 && m.cursor < len(m.filtered) {
			widgetName = m.filtered[m.cursor].Target
		}
		rightHeader := fmt.Sprintf(" %s ", widgetName)
		if len(rightHeader) > rightWidth {
			rightHeader = rightHeader[:rightWidth-4] + "... "
		}
		remainingWidth := rightWidth - len(rightHeader)
		if remainingWidth/2 > 0 {
			statusLine += m.styles.Separator.Render(strings.Repeat("─", remainingWidth/2))
			rightHeader += m.styles.Separator.Render(strings.Repeat("─", remainingWidth-(remainingWidth/2)))
		}
		statusLine += m.styles.Command.Render(rightHeader)
	} else {
		// Not in expanded mode - fill the entire width with separator
		separatorLength := (leftWidth + rightWidth) - len(status) - 2
		if separatorLength > 0 {
			statusLine += m.styles.Separator.Render(strings.Repeat("─", separatorLength))
		}
	}

	lines = append(lines, statusLine)

	start := m.scrollOffset
	end := start + m.maxVisible
	if end > len(m.filtered) {
		end = len(m.filtered)
	}

	// Prepare expanded text if needed
	var expandedLines []string
	if m.expandedMode {
		expandedLines = m.getExpandedDisplayLines(rightWidth)
	}

	// Render shortcut list lines with optional right pane
	for lineIdx := 0; lineIdx < m.maxVisible; lineIdx++ {
		listItemIdx := start + lineIdx
		var leftContent string

		// Render left column (shortcut list)
		if listItemIdx < end {
			leftContent = m.renderShortcut(m.filtered[listItemIdx], listItemIdx == m.cursor, leftWidth)
		} else {
			// Empty line for left column
			leftContent = strings.Repeat(" ", leftWidth)
		}

		// Render right column (expanded description) if needed
		var rightContent string
		if m.expandedMode {
			if lineIdx < len(expandedLines) {
				rightContent = m.formatExpandedLine(expandedLines[lineIdx], rightWidth)
			} else {
				rightContent = strings.Repeat(" ", rightWidth)
			}
		} else {
			if listItemIdx < end {
				rightContent = m.renderDescription(m.filtered[listItemIdx], rightWidth)
			} else {
				rightContent = strings.Repeat(" ", rightWidth)
			}
		}

		// Combine left and right content
		fullLine := leftContent + rightContent

		lines = append(lines, fullLine)
	}

	// Add empty line before help
	emptyLine := strings.Repeat(" ", leftWidth+rightWidth)
	lines = append(lines, emptyLine)

	// Help text (always at bottom)
	helpText := " ↑/↓ navigate  |  → expand"
	if m.expandedMode {
		helpText = " ↑/↓ navigate  |  ← collapse  |  ^D/^U scroll"
	}
	helpLine := m.styles.Help.Render(helpText)
	// Pad help line to full width
	if len(helpLine) < m.width {
		helpLine += strings.Repeat(" ", m.width-len(helpLine))
	}
	lines = append(lines, helpLine)

	return lines
}

func (m model) renderShortcut(shortcut Shortcut, isSelected bool, maxWidth int) string {
	// Reserve space for bar (1) + space (1) + padding (2) = 4 chars
	commandWidth := maxWidth - 4
	command := shortcut.Display
	
	// Truncate command text if too long (before styling)
	if len(command) > commandWidth {
		command = command[:commandWidth-3] + "..."
	}
	
	// Pad command to exact width (before styling)
	paddedCommand := fmt.Sprintf("%-*s", commandWidth, command)
	
	// Apply highlighting to the padded command
	highlightedCommand := m.highlightMatches(paddedCommand, m.query, m.styles.Command, isSelected, m.styles)

	// Build the line with proper components
	var barChar, spaceBg, columnBg string
	if isSelected {
		barChar = m.styles.SelectedBar.Render("▌")
		spaceBg = m.styles.SelectedLine.Render(" ")
		columnBg = m.styles.SelectedLine.Render("  ")
	} else {
		barChar = m.styles.UnselectedBar.Render("█")
		spaceBg = m.styles.AppBackground.Render(" ")
		columnBg = m.styles.AppBackground.Render("  ")
	}
	
	// Combine components
	line := barChar + spaceBg + highlightedCommand + columnBg

	return line
}

func (m model) renderDescription(shortcut Shortcut, maxWidth int) string {
	description := shortcut.Description
	descWidth := maxWidth - 2
	
	// Truncate description text if too long (before styling)
	if len(description) > descWidth {
		description = description[:descWidth-3] + "..."
	}
	
	// Pad description to exact width (before styling)
	paddedDesc := fmt.Sprintf("%-*s", descWidth, description)
	
	// Apply highlighting to the padded description
	highlightedDesc := m.highlightMatches(paddedDesc, m.query, m.styles.Description, false, m.styles)

	// Add padding spaces around the description
	line := "  " + highlightedDesc

	return line
}

// getExpandedDisplayLines prepares the expanded description lines for display
func (m model) getExpandedDisplayLines(maxWidth int) []string {
	if len(m.filtered) == 0 || m.cursor >= len(m.filtered) {
		return []string{"No description available"}
	}

	shortcut := m.filtered[m.cursor]
	fullDesc := shortcut.FullDescription
	if fullDesc == "" {
		fullDesc = shortcut.Description
	}
	if fullDesc == "" {
		fullDesc = "No description available"
	}

	// Wrap text to fit in the right column
	wrappedLines := m.wrapText(fullDesc, maxWidth-4) // Account for padding

	// Apply scroll offset
	visibleLines := m.maxVisible
	start := m.expandedScrollOffset
	end := start + visibleLines
	if end > len(wrappedLines) {
		end = len(wrappedLines)
	}

	if start >= len(wrappedLines) {
		return []string{}
	}

	return wrappedLines[start:end]
}

// formatExpandedLine formats a line for the right column
func (m model) formatExpandedLine(text string, maxWidth int) string {
	// Add padding and ensure exact width
	padded := fmt.Sprintf("  %s  ", text)
	if len(padded) > maxWidth {
		padded = padded[:maxWidth]
	} else if len(padded) < maxWidth {
		padded += strings.Repeat(" ", maxWidth-len(padded))
	}
	return m.styles.Description.Render(padded)
}

// renderExpandedView renders the expanded description view
func (m model) renderExpandedView() string {
	var a strings.Builder
	var b strings.Builder

	// Render query line
	a.WriteString(m.styles.Query.Render("❯ "))
	a.WriteString(m.styles.Query.Render(m.query))
	a.WriteString("\n")

	// Widget name header
	if len(m.filtered) > 0 && m.cursor < len(m.filtered) {
		widgetName := m.filtered[m.cursor].Target
		b.WriteString(m.styles.Command.Render("  " + widgetName))
		b.WriteString("\n")
	}

	// Description box
	visibleLines := m.getExpandedVisibleLines()
	start := m.expandedScrollOffset
	end := start + visibleLines
	if end > len(m.expandedText) {
		end = len(m.expandedText)
	}

	// Top border
	borderWidth := m.width - 4
	if borderWidth < 10 {
		borderWidth = 10
	}
	b.WriteString(m.styles.Separator.Render("┌" + strings.Repeat("─", borderWidth-2) + "┐"))
	b.WriteString("\n")

	// Content lines
	for i := start; i < end; i++ {
		line := m.expandedText[i]
		padding := borderWidth - 2 - len(line)
		if padding < 0 {
			padding = 0
		}
		b.WriteString(m.styles.Separator.Render("│"))
		b.WriteString(m.styles.Description.Render(" " + line + strings.Repeat(" ", padding)))
		b.WriteString(m.styles.Separator.Render("│"))
		b.WriteString("\n")
	}

	// Fill empty lines if needed
	for i := end - start; i < visibleLines; i++ {
		padding := borderWidth - 2
		b.WriteString(m.styles.Separator.Render("│"))
		b.WriteString(m.styles.AppBackground.Render(strings.Repeat(" ", padding)))
		b.WriteString(m.styles.Separator.Render("│"))
		b.WriteString("\n")
	}

	// Bottom border
	b.WriteString(m.styles.Separator.Render("└" + strings.Repeat("─", borderWidth-2) + "┘"))
	b.WriteString("\n")

	// Fill remaining space to push help to bottom
	usedLines := 2 + 1 + visibleLines + 2 + 2 // query + widget + borders + help
	remainingLines := m.height - usedLines
	for i := 0; i < remainingLines; i++ {
		b.WriteString("\n")
	}

	// Help text
	helpText := "← - back to list"
	if len(m.expandedText) > visibleLines {
		helpText = "^D/^U - scroll • ← - back to list"
	}
	b.WriteString(m.styles.Help.Render(helpText))

	a.WriteString(m.styles.AppBackground.Render(b.String()))
	return a.String()
}

func ShowUI(shortcuts []Shortcut, styles ThemeStyles) (*Shortcut, string, error) {
	// Force true color support
	lipgloss.SetColorProfile(termenv.TrueColor)

	m := InitialModel(shortcuts, styles)

	tty, err := os.OpenFile("/dev/tty", os.O_RDWR, 0)
	if err != nil {
		p := tea.NewProgram(m, tea.WithMouseAllMotion())
		finalModel, err := p.Run()
		if err != nil {
			return nil, "", err
		}

		if finalModel, ok := finalModel.(model); ok {
			return finalModel.selected, finalModel.selectedKey, nil
		}
		return nil, "", nil
	}
	defer tty.Close()

	p := tea.NewProgram(m, tea.WithMouseAllMotion(), tea.WithInput(tty), tea.WithOutput(tty))

	finalModel, err := p.Run()
	if err != nil {
		return nil, "", err
	}

	if finalModel, ok := finalModel.(model); ok {
		return finalModel.selected, finalModel.selectedKey, nil
	}

	return nil, "", nil
}
