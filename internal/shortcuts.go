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
	Display     string // What to show in UI (e.g., "Ctrl+A", "gs")
	Description string // Human-readable description
	Type        string // "widget", "command", or "sequence"
	Target      string // What to execute (widget name, command, or key sequence)
	IsCustom    bool   // True if added/modified by user config
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

	builtins, err := getBuiltinShortcuts(shell)
	if err != nil {
		return nil, err
	}

	config, err := loadConfig()
	if err != nil {
		return nil, err
	}

	shortcuts := mergeShortcuts(builtins, config)

	return shortcuts, nil
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
	switch shell {
	case "zsh":
		return getZshBuiltinShortcuts(), nil
	default:
		return nil, fmt.Errorf("no built-in shortcuts available for shell: %s", shell)
	}
}

func getZshBuiltinShortcuts() []Shortcut {
	return []Shortcut{
		{Display: "Ctrl+A", Description: "Beginning of the line", Type: "widget", Target: "beginning-of-line", IsCustom: false},
		{Display: "Ctrl+E", Description: "End of the line", Type: "widget", Target: "end-of-line", IsCustom: false},
		{Display: "Ctrl+F", Description: "Forward one character", Type: "widget", Target: "forward-char", IsCustom: false},
		{Display: "Ctrl+B", Description: "Back one character", Type: "widget", Target: "backward-char", IsCustom: false},
		{Display: "Alt+F", Description: "Forward one word", Type: "widget", Target: "forward-word", IsCustom: false},
		{Display: "Alt+B", Description: "Back one word", Type: "widget", Target: "backward-word", IsCustom: false},
		{Display: "Ctrl+T", Description: "Swap cursor with prev character", Type: "widget", Target: "transpose-chars", IsCustom: false},
		{Display: "Alt+T", Description: "Swap cursor with prev word", Type: "widget", Target: "transpose-words", IsCustom: false},
		{Display: "Ctrl+U", Description: "Clear to beginning of line", Type: "widget", Target: "backward-kill-line", IsCustom: false},
		{Display: "Ctrl+K", Description: "Kill to end of line", Type: "widget", Target: "kill-line", IsCustom: false},
		{Display: "Ctrl+H", Description: "Kill one character backward", Type: "widget", Target: "backward-delete-char", IsCustom: false},
		{Display: "Ctrl+W", Description: "Kill word back (if no Mark)", Type: "widget", Target: "backward-kill-word", IsCustom: false},
		{Display: "Ctrl+@", Description: "Set Mark", Type: "widget", Target: "set-mark-command", IsCustom: false},
		{Display: "Ctrl+Y", Description: "Paste from Kill Ring", Type: "widget", Target: "yank", IsCustom: false},
		{Display: "Ctrl+V", Description: "Quoted insert", Type: "widget", Target: "quoted-insert", IsCustom: false},
		{Display: "Ctrl+Q", Description: "Push line to be used again", Type: "widget", Target: "push-line", IsCustom: false},
		{Display: "Ctrl+_", Description: "Undo", Type: "widget", Target: "undo", IsCustom: false},
		{Display: "Ctrl+P", Description: "Prev line", Type: "widget", Target: "up-line-or-history", IsCustom: false},
		{Display: "Ctrl+N", Description: "Next Line", Type: "widget", Target: "down-line-or-history", IsCustom: false},
		{Display: "Ctrl+R", Description: "Search", Type: "widget", Target: "history-incremental-search-backward", IsCustom: false},
		{Display: "Alt+P", Description: "Match word on line", Type: "widget", Target: "history-search-backward", IsCustom: false},
		{Display: "Alt+.", Description: "Extract last word", Type: "widget", Target: "insert-last-word", IsCustom: false},
		{Display: "Ctrl+L", Description: "Clear screen", Type: "widget", Target: "clear-screen", IsCustom: false},
		{Display: "Ctrl+S", Description: "Stop screen output", Type: "sequence", Target: "C-s", IsCustom: false},
		{Display: "Ctrl+C", Description: "Kill proc", Type: "sequence", Target: "C-c", IsCustom: false},
		{Display: "Ctrl+Z", Description: "Suspend proc", Type: "sequence", Target: "C-z", IsCustom: false},
		{Display: "Ctrl+O", Description: "Exec cmd but keep line", Type: "widget", Target: "accept-line-and-down-history", IsCustom: false},
		{Display: "Tab", Description: "Complete command/filename", Type: "widget", Target: "expand-or-complete", IsCustom: false},
		{Display: "Enter", Description: "Execute command", Type: "widget", Target: "accept-line", IsCustom: false},
		{Display: "Ctrl+D", Description: "Delete character or EOF", Type: "widget", Target: "delete-char-or-list", IsCustom: false},
		{Display: "Ctrl+G", Description: "Abort current operation", Type: "widget", Target: "send-break", IsCustom: false},
		{Display: "Ctrl+X Ctrl+E", Description: "Edit command in editor", Type: "widget", Target: "edit-command-line", IsCustom: false},
		{Display: "↑", Description: "Previous command in history", Type: "widget", Target: "up-line-or-history", IsCustom: false},
		{Display: "↓", Description: "Next command in history", Type: "widget", Target: "down-line-or-history", IsCustom: false},
		{Display: "←", Description: "Move cursor left", Type: "widget", Target: "backward-char", IsCustom: false},
		{Display: "→", Description: "Move cursor right", Type: "widget", Target: "forward-char", IsCustom: false},
		{Display: "Home", Description: "Beginning of line", Type: "widget", Target: "beginning-of-line", IsCustom: false},
		{Display: "End", Description: "End of line", Type: "widget", Target: "end-of-line", IsCustom: false},
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
