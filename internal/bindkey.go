package internal

import (
	"bufio"
	"fmt"
	"os/exec"
	"regexp"
	"strings"
)

// BindkeyEntry represents a single key binding from bindkey output
type BindkeyEntry struct {
	EscapeSequence string // Raw escape sequence like "^A" or "^[f"
	WidgetName     string // Widget name like "beginning-of-line"
	DisplayName    string // Human-readable name like "Ctrl+A"
}

// getZshBindings executes bindkey command and parses the output
func getZshBindings() ([]BindkeyEntry, error) {
	// Use an interactive zsh session that loads the user's config
	// The -i flag makes it interactive, which loads .zshrc
	// Redirect stderr to suppress configuration warnings
	cmd := exec.Command("zsh", "-i", "-c", "bindkey 2>/dev/null")
	output, err := cmd.Output()
	if err != nil {
		return nil, fmt.Errorf("failed to execute bindkey command: %w", err)
	}

	return parseBindkeyOutput(string(output))
}

// parseBindkeyOutput parses the bindkey command output
// If output is empty, it executes the bindkey command with filtering
func parseBindkeyOutput(output string) ([]BindkeyEntry, error) {
	if output == "" {
		return getFilteredZshBindings()
	}
	entries := make([]BindkeyEntry, 0)
	scanner := bufio.NewScanner(strings.NewReader(output))
	
	// Regular expression to match bindkey output format: "key" widget-name
	// Skip range entries like "^A"-"^C" and only process single key entries
	bindkeyRegex := regexp.MustCompile(`^"([^"]*)" +([a-zA-Z0-9_.-]+)$`)
	rangeRegex := regexp.MustCompile(`^"[^"]*"-"[^"]*"`)  // Skip range entries
	
	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())
		if line == "" {
			continue
		}

		// Skip range entries like "^A"-"^C" self-insert
		if rangeRegex.MatchString(line) {
			continue
		}

		matches := bindkeyRegex.FindStringSubmatch(line)
		if len(matches) != 3 {
			// Skip lines that don't match expected format
			continue
		}

		escapeSeq := matches[1]
		widgetName := matches[2]

		// Filter out self-insert and other non-useful widgets
		if shouldSkipWidget(widgetName) {
			continue
		}

		displayName := normalizeEscapeSequence(escapeSeq)
		
		entry := BindkeyEntry{
			EscapeSequence: escapeSeq,
			WidgetName:     widgetName,
			DisplayName:    displayName,
		}
		
		entries = append(entries, entry)
	}

	if err := scanner.Err(); err != nil {
		return nil, fmt.Errorf("error reading bindkey output: %w", err)
	}

	return entries, nil
}

// shouldSkipWidget returns true if the widget should be filtered out
func shouldSkipWidget(widgetName string) bool {
	skipWidgets := map[string]bool{
		"self-insert":          true,
		"self-insert-unmeta":   true,
		"undefined-key":        true,
		"digit-argument":       true,
		"neg-argument":         true,
		"universal-argument":   true,
		"bracketed-paste":      true,
		"bracketed-paste-magic": true,
		"beep":                 true,
		"noop":                 true,
		// "which-command":        true,
	}

	// Skip self-insert and related widgets
	if skipWidgets[widgetName] {
		return true
	}

	// Skip internal/private widgets (start with _)
	if strings.HasPrefix(widgetName, "_") {
		return true
	}

	// Skip vi-mode specific widgets if they contain vi- prefix
	// (we might want to make this configurable later)
	if strings.HasPrefix(widgetName, "vi-") {
		return true
	}

	return false
}

// filterBindkeyEntries filters entries based on various criteria
func filterBindkeyEntries(entries []BindkeyEntry) []BindkeyEntry {
	var filtered []BindkeyEntry
	seen := make(map[string]bool)

	for _, entry := range entries {
		// Skip entries with empty display names
		if entry.DisplayName == "" {
			continue
		}

		// Skip duplicate display names (prefer first occurrence)
		if seen[entry.DisplayName] {
			continue
		}
		seen[entry.DisplayName] = true

		// Skip entries that are just single printable characters
		// (these are usually self-insert and not useful as shortcuts)
		if len(entry.DisplayName) == 1 && 
		   entry.DisplayName[0] >= ' ' && 
		   entry.DisplayName[0] <= '~' && 
		   entry.DisplayName[0] != '^' {
			continue
		}

		filtered = append(filtered, entry)
	}

	return filtered
}

// getFilteredZshBindings is the main function to get useful zsh bindings
func getFilteredZshBindings() ([]BindkeyEntry, error) {
	entries, err := getZshBindings()
	if err != nil {
		return nil, err
	}

	return filterBindkeyEntries(entries), nil
}
