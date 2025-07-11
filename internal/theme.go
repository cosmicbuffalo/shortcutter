package internal

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/BurntSushi/toml"
	"github.com/charmbracelet/lipgloss"
)

type Theme struct {
	Name            string `toml:"name"`
	Primary         string `toml:"primary"`
	Secondary       string `toml:"secondary"`
	Query           string `toml:"query"`
	Accent          string `toml:"accent"`
	SelectedBg      string `toml:"selected_bg"`
	AppBg           string `toml:"app_bg"`
	Muted           string `toml:"muted"`
	Help            string `toml:"help"`
	CustomIndicator string `toml:"custom_indicator"`
	Border          string `toml:"border"`
}

type ThemeStyles struct {
	Title           lipgloss.Style
	SelectedBar     lipgloss.Style
	UnselectedBar   lipgloss.Style
	SelectedLine    lipgloss.Style
	Status          lipgloss.Style
	Separator       lipgloss.Style
	Match           lipgloss.Style
	Command         lipgloss.Style
	Description     lipgloss.Style
	Query           lipgloss.Style
	Help            lipgloss.Style
	CustomIndicator lipgloss.Style
	AppBackground   lipgloss.Style
}

func GetDefaultTheme() Theme {
	return Theme{
		Name:            "default",
		Primary:         "#10B981",
		Secondary:       "#3B82F6",
		Query:           "#FFFFFF",
		Accent:          "#F97316",
		SelectedBg:      "#3F3F3F",
		AppBg:           "transparent",
		Muted:           "#6B7280",
		Help:            "#9CA3AF",
		CustomIndicator: "#9333EA",
		Border:          "#6B7280",
	}
}

// LoadTheme loads a theme from ~/.config/shortcutter/themes/<name>.toml
func LoadTheme(name string) (Theme, error) {
	if name == "" {
		return GetDefaultTheme(), nil
	}

	if name == "default" {
		return GetDefaultTheme(), nil
	}

	homeDir, err := os.UserHomeDir()
	if err != nil {
		return GetDefaultTheme(), fmt.Errorf("could not get home directory: %w", err)
	}

	themePath := filepath.Join(homeDir, ".config", "shortcutter", "themes", name+".toml")

	if _, err := os.Stat(themePath); os.IsNotExist(err) {
		return GetDefaultTheme(), fmt.Errorf("theme '%s' not found at %s", name, themePath)
	}

	var theme Theme
	if _, err := toml.DecodeFile(themePath, &theme); err != nil {
		return GetDefaultTheme(), fmt.Errorf("failed to parse theme file %s: %w", themePath, err)
	}

	if theme.Name == "" {
		theme.Name = name
	}
	defaultTheme := GetDefaultTheme()
	if theme.Primary == "" {
		theme.Primary = defaultTheme.Primary
	}
	if theme.Secondary == "" {
		theme.Secondary = defaultTheme.Secondary
	}
	if theme.Accent == "" {
		theme.Accent = defaultTheme.Accent
	}
	if theme.SelectedBg == "" {
		theme.SelectedBg = defaultTheme.SelectedBg
	}
	if theme.AppBg == "" {
		theme.AppBg = defaultTheme.AppBg
	}
	if theme.Muted == "" {
		theme.Muted = defaultTheme.Muted
	}
	if theme.Help == "" {
		theme.Help = defaultTheme.Help
	}
	if theme.CustomIndicator == "" {
		theme.CustomIndicator = defaultTheme.CustomIndicator
	}
	if theme.Border == "" {
		theme.Border = defaultTheme.Border
	}

	return theme, nil
}

// CreateThemeStyles converts a Theme to ThemeStyles for use in the UI
func CreateThemeStyles(theme Theme) ThemeStyles {
	styles := ThemeStyles{
		Title: lipgloss.NewStyle().
			Bold(true).
			Foreground(lipgloss.Color(theme.Primary)).
			Background(lipgloss.Color(theme.AppBg)),

		SelectedBar: lipgloss.NewStyle().
			Foreground(lipgloss.Color(theme.Accent)).
			Background(lipgloss.Color(theme.SelectedBg)),

		UnselectedBar: lipgloss.NewStyle().
			Foreground(lipgloss.Color(theme.SelectedBg)).
			Background(lipgloss.Color(theme.AppBg)),

		SelectedLine: lipgloss.NewStyle().
			Background(lipgloss.Color(theme.SelectedBg)),

		Status: lipgloss.NewStyle().
			Foreground(lipgloss.Color(theme.Muted)).
			Background(lipgloss.Color(theme.AppBg)),

		Separator: lipgloss.NewStyle().
			Foreground(lipgloss.Color(theme.Border)).
			Background(lipgloss.Color(theme.AppBg)),

		Match: lipgloss.NewStyle().
			Foreground(lipgloss.Color(theme.Secondary)).
			Background(lipgloss.Color(theme.AppBg)),

		Command: lipgloss.NewStyle().
			Bold(true).
			Foreground(lipgloss.Color(theme.Primary)).
			Background(lipgloss.Color(theme.AppBg)),

		Description: lipgloss.NewStyle().
			Foreground(lipgloss.Color(theme.Muted)).
			Background(lipgloss.Color(theme.AppBg)),

		Query: lipgloss.NewStyle().
			Bold(true).
			Foreground(lipgloss.Color(theme.Query)).
		  Background(lipgloss.Color("transparent")),

		Help: lipgloss.NewStyle().
			Foreground(lipgloss.Color(theme.Help)).
			Background(lipgloss.Color(theme.AppBg)),

		CustomIndicator: lipgloss.NewStyle().
			Foreground(lipgloss.Color(theme.CustomIndicator)).
			Background(lipgloss.Color(theme.AppBg)),
	}

	if theme.AppBg != "transparent" && theme.AppBg != "default" && theme.AppBg != "" {
		styles.AppBackground = lipgloss.NewStyle().
			Background(lipgloss.Color(theme.AppBg))
	} else {
		styles.AppBackground = lipgloss.NewStyle()
	}

	return styles
}

func EnsureThemeDirectory() error {
	homeDir, err := os.UserHomeDir()
	if err != nil {
		return err
	}

	themesDir := filepath.Join(homeDir, ".config", "shortcutter", "themes")
	return os.MkdirAll(themesDir, 0755)
}

func ListAvailableThemes() ([]string, error) {
	themes := []string{"default"}

	homeDir, err := os.UserHomeDir()
	if err != nil {
		return themes, nil
	}

	themesDir := filepath.Join(homeDir, ".config", "shortcutter", "themes")

	if _, err := os.Stat(themesDir); os.IsNotExist(err) {
		return themes, nil
	}
	files, err := os.ReadDir(themesDir)
	if err != nil {
		return themes, nil // Return just default if can't read directory
	}

	for _, file := range files {
		if !file.IsDir() && filepath.Ext(file.Name()) == ".toml" {
			themeName := file.Name()[:len(file.Name())-5] // Remove .toml extension
			if themeName != "default" {                   // Don't duplicate built-in default
				themes = append(themes, themeName)
			}
		}
	}

	return themes, nil
}
