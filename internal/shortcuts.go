package internal

import (
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"sort"
	"strings"

	"github.com/BurntSushi/toml"
)

type Shortcut struct {
	Display         string // What to show in UI (e.g., "Ctrl+A", "gs")
	Description     string // Human-readable short description
	FullDescription string // Complete description from manual
	Type            string // "widget", "command", or "sequence"
	Target          string // What to execute (widget name, command, or key sequence)
	IsCustom        bool   // True if added/modified by user config
}

type Config struct {
	Shortcuts map[string]interface{} `toml:"shortcuts"`
	Theme     ThemeConfig            `toml:"theme"`
}

type ThemeConfig struct {
	Name string `toml:"name"`
}

func LoadShortcuts() ([]Shortcut, error) {
	shell, err := detectShell()
	if err != nil {
		return nil, err
	}

	// Get dynamic shortcuts from bindkey and man pages
	dynamicShortcuts, err := loadDynamicShortcuts(shell)
	if err != nil {
		return nil, fmt.Errorf("failed to load dynamic shortcuts: %w", err)
	}

	// Load user config
	config, err := loadConfig()
	if err != nil {
		return nil, err
	}

	// Merge with user config
	shortcuts := mergeShortcuts(dynamicShortcuts, config)

	return shortcuts, nil
}

// loadDynamicShortcuts loads shortcuts dynamically from bindkey and man pages with caching
func loadDynamicShortcuts(shell string) ([]Shortcut, error) {
	if shell != "zsh" {
		// For non-zsh shells, return error since we don't support them dynamically
		return nil, fmt.Errorf("dynamic shortcut loading only supported for zsh")
	}

	// Initialize cache manager
	cacheManager, err := NewCacheManager()
	if err != nil {
		return nil, fmt.Errorf("failed to initialize cache manager: %w", err)
	}

	// Try to load from cache first
	if cacheData, err := cacheManager.LoadCache(); err == nil && cacheData != nil {
		return convertCacheToShortcuts(cacheData), nil
	}

	// Cache miss - generate fresh data

	bindkeyEntries, err := parseBindkeyOutput("")
	if err != nil {
		return nil, fmt.Errorf("failed to parse bindkey output: %w", err)
	}

	manDescriptions, err := getWidgetDescriptions()
	if err != nil {
		return nil, fmt.Errorf("failed to get widget descriptions: %w", err)
	}

	// Save to cache for next time
	if err := cacheManager.SaveCache(bindkeyEntries, manDescriptions); err != nil {
		// Don't fail if we can't save cache, just log warning
		fmt.Fprintf(os.Stderr, "Warning: Failed to save cache: %v\n", err)
	}

	// Convert to shortcuts
	return convertBindkeyToShortcuts(bindkeyEntries, manDescriptions), nil
}

// convertCacheToShortcuts converts cached data back to shortcuts
func convertCacheToShortcuts(cacheData *CacheData) []Shortcut {
	return convertBindkeyToShortcuts(cacheData.BindkeyEntries, cacheData.ManDescriptions)
}

// convertBindkeyToShortcuts converts bindkey entries and descriptions to shortcuts
func convertBindkeyToShortcuts(bindkeyEntries []BindkeyEntry, manDescriptions map[string]WidgetDescription) []Shortcut {
	shortcuts := make([]Shortcut, 0, len(bindkeyEntries))

	for _, entry := range bindkeyEntries {
		// Get description from man page, fall back to widget name
		widgetDesc, exists := manDescriptions[entry.WidgetName]
		shortDescription := entry.WidgetName
		fullDescription := entry.WidgetName
		if exists {
			shortDescription = widgetDesc.ShortDescription
			fullDescription = widgetDesc.FullDescription
		}

		shortcut := Shortcut{
			Display:         entry.DisplayName,
			Description:     shortDescription,
			FullDescription: fullDescription,
			Type:            "widget",
			Target:          entry.WidgetName,
			IsCustom:        false,
		}

		shortcuts = append(shortcuts, shortcut)
	}

	// Sort by display name for consistent ordering
	sort.Slice(shortcuts, func(i, j int) bool {
		return shortcuts[i].Display < shortcuts[j].Display
	})

	return shortcuts
}

func LoadShortcutsAndTheme() ([]Shortcut, ThemeStyles, error) {
	shortcuts, err := LoadShortcuts()
	if err != nil {
		return nil, ThemeStyles{}, err
	}

	config, err := loadConfig()
	if err != nil {
		defaultTheme := GetDefaultTheme()
		styles := CreateThemeStyles(defaultTheme)
		return shortcuts, styles, nil
	}

	themeName := config.Theme.Name
	if themeName == "" {
		themeName = "default"
	}

	theme, err := LoadTheme(themeName)
	if err != nil {
		theme = GetDefaultTheme()
	}

	styles := CreateThemeStyles(theme)

	return shortcuts, styles, nil
}

func detectShell() (string, error) {
	shell := getShellEnv()
	if shell == "" {
		return "", fmt.Errorf("SHELL environment variable not set")
	}

	shellName := filepath.Base(shell)

	switch shellName {
	case "zsh":
		return "zsh", nil
	case "bash":
		return "", fmt.Errorf("bash support not implemented yet - please use zsh")
	case "fish":
		return "", fmt.Errorf("fish support not implemented yet - please use zsh")
	default:
		return "", fmt.Errorf("unsupported shell '%s' - only zsh is supported", shellName)
	}
}

func getBuiltinShortcuts(shell string) ([]Shortcut, error) {
	// No hardcoded shortcuts - always return error as per user requirements
	return nil, fmt.Errorf("hardcoded shortcuts have been removed - dynamic loading required")
}


func getZshBuiltinShortcuts() []Shortcut {
	return []Shortcut{
		{Display: "Ctrl+A", Description: "Beginning of the line", FullDescription: "Move to the beginning of the line. If already at the beginning of the line, move to the beginning of the previous line, if any.", Type: "widget", Target: "beginning-of-line", IsCustom: false},
		{Display: "Ctrl+E", Description: "End of the line", FullDescription: "Move to the end of the line. If already at the end of the line, move to the end of the next line, if any.", Type: "widget", Target: "end-of-line", IsCustom: false},
		{Display: "Ctrl+F", Description: "Forward one character", FullDescription: "Move forward one character.", Type: "widget", Target: "forward-char", IsCustom: false},
		{Display: "Ctrl+B", Description: "Back one character", FullDescription: "Move backward one character.", Type: "widget", Target: "backward-char", IsCustom: false},
		{Display: "Alt+F", Description: "Forward one word", FullDescription: "Move to the beginning of the next word.", Type: "widget", Target: "forward-word", IsCustom: false},
		{Display: "Alt+B", Description: "Back one word", FullDescription: "Move to the beginning of the previous word.", Type: "widget", Target: "backward-word", IsCustom: false},
		{Display: "Ctrl+T", Description: "Swap cursor with prev character", FullDescription: "Transpose the characters at the cursor position with the character before it.", Type: "widget", Target: "transpose-chars", IsCustom: false},
		{Display: "Alt+T", Description: "Swap cursor with prev word", FullDescription: "Exchange the current word with the one before it.", Type: "widget", Target: "transpose-words", IsCustom: false},
		{Display: "Ctrl+U", Description: "Clear to beginning of line", FullDescription: "Kill from the beginning of the line to the cursor position.", Type: "widget", Target: "backward-kill-line", IsCustom: false},
		{Display: "Ctrl+K", Description: "Kill to end of line", FullDescription: "Kill from the cursor to the end of the line.", Type: "widget", Target: "kill-line", IsCustom: false},
		{Display: "Ctrl+H", Description: "Kill one character backward", FullDescription: "Delete the character behind the cursor.", Type: "widget", Target: "backward-delete-char", IsCustom: false},
		{Display: "Ctrl+W", Description: "Kill word back (if no Mark)", FullDescription: "Kill the word behind the cursor.", Type: "widget", Target: "backward-kill-word", IsCustom: false},
		{Display: "Ctrl+@", Description: "Set Mark", FullDescription: "Set the mark at the cursor position.", Type: "widget", Target: "set-mark-command", IsCustom: false},
		{Display: "Ctrl+Y", Description: "Paste from Kill Ring", FullDescription: "Insert the most recently killed text at the cursor position.", Type: "widget", Target: "yank", IsCustom: false},
		{Display: "Ctrl+V", Description: "Quoted insert", FullDescription: "Insert the next character typed, even if it is a special character.", Type: "widget", Target: "quoted-insert", IsCustom: false},
		{Display: "Ctrl+Q", Description: "Push line to be used again", FullDescription: "Push the current line onto the buffer stack and clear the line.", Type: "widget", Target: "push-line", IsCustom: false},
		{Display: "Ctrl+_", Description: "Undo", FullDescription: "Undo the last change to the line.", Type: "widget", Target: "undo", IsCustom: false},
		{Display: "Ctrl+P", Description: "Prev line", FullDescription: "Move up a line in the buffer, or if already at the top line, move to the previous event in the history list.", Type: "widget", Target: "up-line-or-history", IsCustom: false},
		{Display: "Ctrl+N", Description: "Next Line", FullDescription: "Move down a line in the buffer, or if already at the bottom line, move to the next event in the history list.", Type: "widget", Target: "down-line-or-history", IsCustom: false},
		{Display: "Ctrl+R", Description: "Search", FullDescription: "Search backward incrementally for a specified string. The search is case-insensitive if the search string does not have uppercase letters.", Type: "widget", Target: "history-incremental-search-backward", IsCustom: false},
		{Display: "Alt+P", Description: "Match word on line", FullDescription: "Search backward in history for a line beginning with the current word.", Type: "widget", Target: "history-search-backward", IsCustom: false},
		{Display: "Alt+.", Description: "Extract last word", FullDescription: "Insert the last word from the previous history event at the cursor position.", Type: "widget", Target: "insert-last-word", IsCustom: false},
		{Display: "Ctrl+L", Description: "Clear screen", FullDescription: "Clear the screen leaving the current line at the top of the screen.", Type: "widget", Target: "clear-screen", IsCustom: false},
		{Display: "Ctrl+S", Description: "Stop screen output", FullDescription: "Stop output to the screen (flow control).", Type: "sequence", Target: "C-s", IsCustom: false},
		{Display: "Ctrl+C", Description: "Kill proc", FullDescription: "Send SIGINT signal to the current process.", Type: "sequence", Target: "C-c", IsCustom: false},
		{Display: "Ctrl+Z", Description: "Suspend proc", FullDescription: "Send SIGTSTP signal to suspend the current process.", Type: "sequence", Target: "C-z", IsCustom: false},
		{Display: "Ctrl+O", Description: "Exec cmd but keep line", FullDescription: "Execute the current line, and push the next history event on the editing buffer stack.", Type: "widget", Target: "accept-line-and-down-history", IsCustom: false},
		{Display: "Tab", Description: "Complete command/filename", FullDescription: "Attempt shell expansion on the current word. If that fails, attempt completion.", Type: "widget", Target: "expand-or-complete", IsCustom: false},
		{Display: "Enter", Description: "Execute command", FullDescription: "Finish editing the buffer. Normally, this will accept the line and execute it.", Type: "widget", Target: "accept-line", IsCustom: false},
		{Display: "Ctrl+D", Description: "Delete character or EOF", FullDescription: "Delete the character under the cursor, or, if the cursor is at the end of the line, list possible completions for the current word.", Type: "widget", Target: "delete-char-or-list", IsCustom: false},
		{Display: "Ctrl+G", Description: "Abort current operation", FullDescription: "Abort the current editor function.", Type: "widget", Target: "send-break", IsCustom: false},
		{Display: "Ctrl+X Ctrl+E", Description: "Edit command in editor", FullDescription: "Edit the command line using the editor specified by the EDITOR or VISUAL environment variable.", Type: "widget", Target: "edit-command-line", IsCustom: false},
		{Display: "↑", Description: "Previous command in history", FullDescription: "Move up a line in the buffer, or if already at the top line, move to the previous event in the history list.", Type: "widget", Target: "up-line-or-history", IsCustom: false},
		{Display: "↓", Description: "Next command in history", FullDescription: "Move down a line in the buffer, or if already at the bottom line, move to the next event in the history list.", Type: "widget", Target: "down-line-or-history", IsCustom: false},
		{Display: "←", Description: "Move cursor left", FullDescription: "Move backward one character.", Type: "widget", Target: "backward-char", IsCustom: false},
		{Display: "→", Description: "Move cursor right", FullDescription: "Move forward one character.", Type: "widget", Target: "forward-char", IsCustom: false},
		{Display: "Home", Description: "Beginning of line", FullDescription: "Move to the beginning of the line. If already at the beginning of the line, move to the beginning of the previous line, if any.", Type: "widget", Target: "beginning-of-line", IsCustom: false},
		{Display: "End", Description: "End of line", FullDescription: "Move to the end of the line. If already at the end of the line, move to the end of the next line, if any.", Type: "widget", Target: "end-of-line", IsCustom: false},
	}
}

func getBashBuiltinShortcuts() []Shortcut {
	return []Shortcut{
		{Display: "Ctrl+A", Description: "Beginning of line", FullDescription: "Beginning of line", Type: "sequence", Target: "C-a", IsCustom: false},
		{Display: "Ctrl+E", Description: "End of line", FullDescription: "End of line", Type: "sequence", Target: "C-e", IsCustom: false},
		{Display: "Ctrl+F", Description: "Forward one character", FullDescription: "Forward one character", Type: "sequence", Target: "C-f", IsCustom: false},
		{Display: "Ctrl+B", Description: "Back one character", Type: "sequence", Target: "C-b", IsCustom: false},
		{Display: "Alt+F", Description: "Forward one word", Type: "sequence", Target: "M-f", IsCustom: false},
		{Display: "Alt+B", Description: "Back one word", Type: "sequence", Target: "M-b", IsCustom: false},
		{Display: "Ctrl+T", Description: "Transpose characters", Type: "sequence", Target: "C-t", IsCustom: false},
		{Display: "Alt+T", Description: "Transpose words", Type: "sequence", Target: "M-t", IsCustom: false},
		{Display: "Ctrl+U", Description: "Kill line backward", Type: "sequence", Target: "C-u", IsCustom: false},
		{Display: "Ctrl+K", Description: "Kill line forward", Type: "sequence", Target: "C-k", IsCustom: false},
		{Display: "Ctrl+H", Description: "Delete character backward", Type: "sequence", Target: "C-h", IsCustom: false},
		{Display: "Ctrl+W", Description: "Kill word backward", Type: "sequence", Target: "C-w", IsCustom: false},
		{Display: "Ctrl+Y", Description: "Yank", Type: "sequence", Target: "C-y", IsCustom: false},
		{Display: "Ctrl+L", Description: "Clear screen", Type: "sequence", Target: "C-l", IsCustom: false},
		{Display: "Ctrl+R", Description: "Reverse search", Type: "sequence", Target: "C-r", IsCustom: false},
		{Display: "Ctrl+S", Description: "Forward search", Type: "sequence", Target: "C-s", IsCustom: false},
		{Display: "Ctrl+P", Description: "Previous line", Type: "sequence", Target: "C-p", IsCustom: false},
		{Display: "Ctrl+N", Description: "Next line", Type: "sequence", Target: "C-n", IsCustom: false},
		{Display: "Ctrl+D", Description: "Delete character or EOF", Type: "sequence", Target: "C-d", IsCustom: false},
		{Display: "Ctrl+C", Description: "Interrupt", Type: "sequence", Target: "C-c", IsCustom: false},
		{Display: "Ctrl+Z", Description: "Suspend", Type: "sequence", Target: "C-z", IsCustom: false},
		{Display: "Tab", Description: "Complete", Type: "sequence", Target: "Tab", IsCustom: false},
		{Display: "Enter", Description: "Execute command", Type: "sequence", Target: "Enter", IsCustom: false},
		{Display: "↑", Description: "Previous command", Type: "sequence", Target: "Up", IsCustom: false},
		{Display: "↓", Description: "Next command", Type: "sequence", Target: "Down", IsCustom: false},
		{Display: "←", Description: "Move cursor left", Type: "sequence", Target: "Left", IsCustom: false},
		{Display: "→", Description: "Move cursor right", Type: "sequence", Target: "Right", IsCustom: false},
		{Display: "Home", Description: "Beginning of line", Type: "sequence", Target: "Home", IsCustom: false},
		{Display: "End", Description: "End of line", Type: "sequence", Target: "End", IsCustom: false},
		{Display: "Delete", Description: "Delete character", Type: "sequence", Target: "Delete", IsCustom: false},
		{Display: "Backspace", Description: "Delete character backward", Type: "sequence", Target: "Backspace", IsCustom: false},
		{Display: "Page Up", Description: "Page up", Type: "sequence", Target: "Page_Up", IsCustom: false},
		{Display: "Page Down", Description: "Page down", Type: "sequence", Target: "Page_Down", IsCustom: false},
	}
}

func getGenericBuiltinShortcuts() []Shortcut {
	return []Shortcut{
		{Display: "Ctrl+A", Description: "Beginning of line", Type: "sequence", Target: "C-a", IsCustom: false},
		{Display: "Ctrl+E", Description: "End of line", Type: "sequence", Target: "C-e", IsCustom: false},
		{Display: "Ctrl+F", Description: "Forward one character", Type: "sequence", Target: "C-f", IsCustom: false},
		{Display: "Ctrl+B", Description: "Back one character", Type: "sequence", Target: "C-b", IsCustom: false},
		{Display: "Ctrl+U", Description: "Kill line backward", Type: "sequence", Target: "C-u", IsCustom: false},
		{Display: "Ctrl+K", Description: "Kill line forward", Type: "sequence", Target: "C-k", IsCustom: false},
		{Display: "Ctrl+L", Description: "Clear screen", Type: "sequence", Target: "C-l", IsCustom: false},
		{Display: "Ctrl+C", Description: "Interrupt", Type: "sequence", Target: "C-c", IsCustom: false},
		{Display: "Ctrl+Z", Description: "Suspend", Type: "sequence", Target: "C-z", IsCustom: false},
		{Display: "Tab", Description: "Complete", Type: "sequence", Target: "Tab", IsCustom: false},
		{Display: "Enter", Description: "Execute command", Type: "sequence", Target: "Enter", IsCustom: false},
		{Display: "↑", Description: "Previous command", Type: "sequence", Target: "Up", IsCustom: false},
		{Display: "↓", Description: "Next command", Type: "sequence", Target: "Down", IsCustom: false},
		{Display: "←", Description: "Move cursor left", Type: "sequence", Target: "Left", IsCustom: false},
		{Display: "→", Description: "Move cursor right", Type: "sequence", Target: "Right", IsCustom: false},
		{Display: "Home", Description: "Beginning of line", Type: "sequence", Target: "Home", IsCustom: false},
		{Display: "End", Description: "End of line", Type: "sequence", Target: "End", IsCustom: false},
		{Display: "Delete", Description: "Delete character", Type: "sequence", Target: "Delete", IsCustom: false},
		{Display: "Backspace", Description: "Delete character backward", Type: "sequence", Target: "Backspace", IsCustom: false},
	}
}

func normalizeKey(key string) string {
	key = strings.TrimSpace(key)
	if matched, _ := regexp.MatchString(`^\^[A-Za-z@_\[\]\\]$`, key); matched {
		char := strings.ToUpper(string(key[1]))
		switch char {
		case "[":
			return "Esc"
		case "I":
			return "Tab"
		case "M":
			return "Enter"
		case "H":
			return "Backspace"
		case "@":
			return "Ctrl+@"
		case "_":
			return "Ctrl+_"
		case "\\":
			return "Ctrl+\\"
		case "]":
			return "Ctrl+]"
		default:
			return "Ctrl+" + char
		}
	}
	if matched, _ := regexp.MatchString(`^[Cc]-[a-zA-Z@_\[\]\\]$`, key); matched {
		char := strings.ToUpper(string(key[2]))
		return "Ctrl+" + char
	}
	if matched, _ := regexp.MatchString(`^[Mm]-[a-zA-Z]$`, key); matched {
		char := strings.ToUpper(string(key[2]))
		return "Alt+" + char
	}
	key = regexp.MustCompile(`(?i)ctrl\+`).ReplaceAllString(key, "Ctrl+")
	key = regexp.MustCompile(`(?i)alt\+`).ReplaceAllString(key, "Alt+")
	key = regexp.MustCompile(`(?i)shift\+`).ReplaceAllString(key, "Shift+")
	key = regexp.MustCompile(`(?i)meta\+`).ReplaceAllString(key, "Alt+")

	parts := strings.Split(key, "+")
	if len(parts) > 1 {
		lastPart := parts[len(parts)-1]
		if len(lastPart) == 1 && lastPart >= "a" && lastPart <= "z" {
			parts[len(parts)-1] = strings.ToUpper(lastPart)
		} else if strings.ToLower(lastPart) == "tab" {
			parts[len(parts)-1] = "Tab"
		}
		key = strings.Join(parts, "+")
	}

	return key
}

func loadConfig() (*Config, error) {
	homeDir, err := os.UserHomeDir()
	if err != nil {
		return &Config{Shortcuts: make(map[string]interface{})}, nil
	}

	configPath := filepath.Join(homeDir, ".config", "shortcutter", "config.toml")

	if _, err := os.Stat(configPath); os.IsNotExist(err) {
		return &Config{Shortcuts: make(map[string]interface{})}, nil
	}

	var config Config
	if _, err := toml.DecodeFile(configPath, &config); err != nil {
		return nil, fmt.Errorf("failed to parse config file: %w", err)
	}

	if config.Shortcuts == nil {
		config.Shortcuts = make(map[string]interface{})
	}

	return &config, nil
}

func mergeShortcuts(builtins []Shortcut, config *Config) []Shortcut {
	shortcutMap := make(map[string]Shortcut)
	
	// Index built-ins by their display name
	for _, shortcut := range builtins {
		normalizedKey := normalizeKey(shortcut.Display)
		shortcutMap[normalizedKey] = shortcut
	}

	for configKey, configValue := range config.Shortcuts {
		normalizedKey := normalizeKey(configKey)

		switch v := configValue.(type) {
		case bool:
			// Disable shortcut
			if !v {
				delete(shortcutMap, normalizedKey)
			}
		case string:
			// Simple override - just change description, inherit everything else from built-in
			if v != "" {
				if existing, exists := shortcutMap[normalizedKey]; exists {
					// Override description but keep other fields
					existing.Description = v
					existing.IsCustom = true
					shortcutMap[normalizedKey] = existing
				} else {
					// New shortcut with just description - assume it's a command
					shortcut := Shortcut{
						Display:     normalizedKey,
						Description: v,
						Type:        "command",
						Target:      v, // Use description as command for simple cases
						IsCustom:    true,
					}
					shortcutMap[normalizedKey] = shortcut
				}
			}
		case map[string]interface{}:
			// Full object configuration
			shortcut := Shortcut{
				Display:  normalizedKey,
				IsCustom: true,
			}
			
			// Start with existing built-in if it exists
			if existing, exists := shortcutMap[normalizedKey]; exists {
				shortcut = existing
				shortcut.IsCustom = true
			}
			
			// Override with config values
			if display, ok := v["display"].(string); ok {
				shortcut.Display = display
			}
			if description, ok := v["description"].(string); ok {
				shortcut.Description = description
			}
			if shortcutType, ok := v["type"].(string); ok {
				shortcut.Type = shortcutType
			}
			if target, ok := v["target"].(string); ok {
				shortcut.Target = target
			}
			
			shortcutMap[normalizedKey] = shortcut
		}
	}

	result := make([]Shortcut, 0, len(shortcutMap))
	for _, shortcut := range shortcutMap {
		result = append(result, shortcut)
	}

	sort.Slice(result, func(i, j int) bool {
		return result[i].Display < result[j].Display
	})

	return result
}

func DetectShortcuts() ([]Shortcut, error) {
	return LoadShortcuts()
}

func NormalizeKeyForTesting(key string) string {
	return normalizeKey(key)
}

var getShellEnv = func() string {
	return os.Getenv("SHELL")
}
