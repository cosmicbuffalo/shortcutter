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
	Command     string
	Description string
	Action      string // What to do: "execute" or "populate" or "info"
	Type        string // "keybinding"
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
		{Command: "Ctrl+A", Description: "Beginning of the line", Action: "info", Type: "keybinding", IsCustom: false},
		{Command: "Ctrl+E", Description: "End of the line", Action: "info", Type: "keybinding", IsCustom: false},
		{Command: "Ctrl+F", Description: "Forward one character", Action: "info", Type: "keybinding", IsCustom: false},
		{Command: "Ctrl+B", Description: "Back one character", Action: "info", Type: "keybinding", IsCustom: false},
		{Command: "Alt+F", Description: "Forward one word", Action: "info", Type: "keybinding", IsCustom: false},
		{Command: "Alt+B", Description: "Back one word", Action: "info", Type: "keybinding", IsCustom: false},
		{Command: "Ctrl+T", Description: "Swap cursor with prev character", Action: "info", Type: "keybinding", IsCustom: false},
		{Command: "Alt+T", Description: "Swap cursor with prev word", Action: "info", Type: "keybinding", IsCustom: false},
		{Command: "Ctrl+U", Description: "Clear to beginning of line", Action: "info", Type: "keybinding", IsCustom: false},
		{Command: "Ctrl+K", Description: "Kill to end of line", Action: "info", Type: "keybinding", IsCustom: false},
		{Command: "Ctrl+H", Description: "Kill one character backward", Action: "info", Type: "keybinding", IsCustom: false},
		{Command: "Ctrl+W", Description: "Kill word back (if no Mark)", Action: "info", Type: "keybinding", IsCustom: false},
		{Command: "Ctrl+@", Description: "Set Mark", Action: "info", Type: "keybinding", IsCustom: false},
		{Command: "Ctrl+Y", Description: "Paste from Kill Ring", Action: "info", Type: "keybinding", IsCustom: false},
		{Command: "Ctrl+V", Description: "Quoted insert", Action: "info", Type: "keybinding", IsCustom: false},
		{Command: "Ctrl+Q", Description: "Push line to be used again", Action: "info", Type: "keybinding", IsCustom: false},
		{Command: "Ctrl+_", Description: "Undo", Action: "info", Type: "keybinding", IsCustom: false},
		{Command: "Ctrl+P", Description: "Prev line", Action: "info", Type: "keybinding", IsCustom: false},
		{Command: "Ctrl+N", Description: "Next Line", Action: "info", Type: "keybinding", IsCustom: false},
		{Command: "Ctrl+R", Description: "Search", Action: "info", Type: "keybinding", IsCustom: false},
		{Command: "Alt+P", Description: "Match word on line", Action: "info", Type: "keybinding", IsCustom: false},
		{Command: "Alt+.", Description: "Extract last word", Action: "info", Type: "keybinding", IsCustom: false},
		{Command: "Ctrl+L", Description: "Clear screen", Action: "info", Type: "keybinding", IsCustom: false},
		{Command: "Ctrl+S", Description: "Stop screen output", Action: "info", Type: "keybinding", IsCustom: false},
		{Command: "Ctrl+C", Description: "Kill proc", Action: "info", Type: "keybinding", IsCustom: false},
		{Command: "Ctrl+Z", Description: "Suspend proc", Action: "info", Type: "keybinding", IsCustom: false},
		{Command: "Ctrl+O", Description: "Exec cmd but keep line", Action: "info", Type: "keybinding", IsCustom: false},
		{Command: "Tab", Description: "Complete command/filename", Action: "info", Type: "keybinding", IsCustom: false},
		{Command: "Enter", Description: "Execute command", Action: "info", Type: "keybinding", IsCustom: false},
		{Command: "Ctrl+D", Description: "Delete character or EOF", Action: "info", Type: "keybinding", IsCustom: false},
		{Command: "Ctrl+G", Description: "Abort current operation", Action: "info", Type: "keybinding", IsCustom: false},
		{Command: "Ctrl+X Ctrl+E", Description: "Edit command in editor", Action: "info", Type: "keybinding", IsCustom: false},
		{Command: "↑", Description: "Previous command in history", Action: "info", Type: "keybinding", IsCustom: false},
		{Command: "↓", Description: "Next command in history", Action: "info", Type: "keybinding", IsCustom: false},
		{Command: "←", Description: "Move cursor left", Action: "info", Type: "keybinding", IsCustom: false},
		{Command: "→", Description: "Move cursor right", Action: "info", Type: "keybinding", IsCustom: false},
		{Command: "Home", Description: "Beginning of line", Action: "info", Type: "keybinding", IsCustom: false},
		{Command: "End", Description: "End of line", Action: "info", Type: "keybinding", IsCustom: false},
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
	for _, shortcut := range builtins {
		normalizedKey := normalizeKey(shortcut.Command)
		shortcutMap[normalizedKey] = shortcut
	}

	for configKey, configValue := range config.Shortcuts {
		normalizedKey := normalizeKey(configKey)

		switch v := configValue.(type) {
		case bool:
			if !v {
				delete(shortcutMap, normalizedKey)
			}
		case string:
			if v != "" {
				shortcut := Shortcut{
					Command:     normalizedKey,
					Description: v,
					Action:      "info",
					Type:        "keybinding",
					IsCustom:    true,
				}
				shortcutMap[normalizedKey] = shortcut
			}
		}
	}

	result := make([]Shortcut, 0, len(shortcutMap))
	for _, shortcut := range shortcutMap {
		result = append(result, shortcut)
	}

	sort.Slice(result, func(i, j int) bool {
		return result[i].Command < result[j].Command
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
