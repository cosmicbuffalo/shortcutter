package internal

import (
	"os"
	"path/filepath"
	"testing"
)

func TestGetDefaultTheme(t *testing.T) {
	theme := GetDefaultTheme()

	if theme.Primary == "" {
		t.Error("Default theme missing primary color")
	}
	if theme.Secondary == "" {
		t.Error("Default theme missing secondary color")
	}
	if theme.Accent == "" {
		t.Error("Default theme missing accent color")
	}
	if theme.SelectedBg == "" {
		t.Error("Default theme missing selected background color")
	}
	if theme.AppBg == "" {
		t.Error("Default theme missing app background color")
	}
	if theme.Muted == "" {
		t.Error("Default theme missing muted color")
	}
	if theme.Help == "" {
		t.Error("Default theme missing help color")
	}
	if theme.CustomIndicator == "" {
		t.Error("Default theme missing custom indicator color")
	}
	if theme.Border == "" {
		t.Error("Default theme missing border color")
	}
}

func TestCreateThemeStyles(t *testing.T) {
	theme := GetDefaultTheme()
	styles := CreateThemeStyles(theme)

	testText := "test"
	if styles.Command.Render(testText) == "" {
		t.Error("Command style should render text")
	}
	if styles.Query.Render(testText) == "" {
		t.Error("Query style should render text")
	}
	if styles.Status.Render(testText) == "" {
		t.Error("Status style should render text")
	}
	if styles.Description.Render(testText) == "" {
		t.Error("Description style should render text")
	}
	if styles.Help.Render(testText) == "" {
		t.Error("Help style should render text")
	}
	if styles.Match.Render(testText) == "" {
		t.Error("Match style should render text")
	}
	if styles.SelectedBar.Render(testText) == "" {
		t.Error("SelectedBar style should render text")
	}
	if styles.UnselectedBar.Render(testText) == "" {
		t.Error("UnselectedBar style should render text")
	}
	if styles.CustomIndicator.Render(testText) == "" {
		t.Error("CustomIndicator style should render text")
	}
	if styles.Separator.Render(testText) == "" {
		t.Error("Separator style should render text")
	}
	if styles.SelectedLine.Render(testText) == "" {
		t.Error("SelectedLine style should render text")
	}
}

func TestLoadTheme(t *testing.T) {
	theme, err := LoadTheme("default")
	if err != nil {
		t.Errorf("LoadTheme('default') should not fail: %v", err)
	}
	if theme.Primary == "" {
		t.Error("Loaded default theme should have primary color")
	}

	theme, err = LoadTheme("nonexistent")
	if err == nil {
		t.Error("LoadTheme('nonexistent') should return an error")
	}
	if theme.Primary == "" {
		t.Error("Fallback theme should have primary color even when error occurs")
	}
}

func TestLoadThemeFromFile(t *testing.T) {
	// Create a temporary theme file for testing
	tempDir := t.TempDir()
	themeDir := filepath.Join(tempDir, ".config", "shortcutter", "themes")
	err := os.MkdirAll(themeDir, 0755)
	if err != nil {
		t.Fatalf("Failed to create temp theme directory: %v", err)
	}

	// Create test theme file
	testThemeContent := `name = "test"
primary = "#FF0000"
secondary = "#00FF00"
accent = "#0000FF"
selected_bg = "#333333"
app_bg = "black"
muted = "#666666"
help = "#999999"
custom_indicator = "#FFFF00"
border = "#CCCCCC"
`

	testThemePath := filepath.Join(themeDir, "test.toml")
	err = os.WriteFile(testThemePath, []byte(testThemeContent), 0644)
	if err != nil {
		t.Fatalf("Failed to write test theme file: %v", err)
	}

	// Mock the home directory for this test
	originalHome := os.Getenv("HOME")
	os.Setenv("HOME", tempDir)
	defer os.Setenv("HOME", originalHome)

	// Test loading the theme
	theme, err := LoadTheme("test")
	if err != nil {
		t.Errorf("LoadTheme('test') failed: %v", err)
	}

	// Verify theme values
	if theme.Primary != "#FF0000" {
		t.Errorf("Expected primary color '#FF0000', got '%s'", theme.Primary)
	}
	if theme.Secondary != "#00FF00" {
		t.Errorf("Expected secondary color '#00FF00', got '%s'", theme.Secondary)
	}
	if theme.Accent != "#0000FF" {
		t.Errorf("Expected accent color '#0000FF', got '%s'", theme.Accent)
	}
}

func TestLoadShortcutsAndTheme(t *testing.T) {
	// Test that LoadShortcutsAndTheme returns both shortcuts and styles
	shortcuts, styles, err := LoadShortcutsAndTheme()
	if err != nil {
		t.Errorf("LoadShortcutsAndTheme() failed: %v", err)
	}

	if len(shortcuts) == 0 {
		t.Error("LoadShortcutsAndTheme() returned empty shortcuts")
	}

	// Test that styles were created
	if styles.Command.Render("test") == "" {
		t.Error("LoadShortcutsAndTheme() should return valid styles")
	}
}

func TestThemeWithTransparentBackground(t *testing.T) {
	// Create a theme with transparent background
	theme := Theme{
		Primary:         "#10B981",
		Secondary:       "#3B82F6",
		Accent:          "#F97316",
		SelectedBg:      "#2D2D2D",
		AppBg:           "transparent", // Test transparent background
		Muted:           "#6B7280",
		Help:            "#9CA3AF",
		CustomIndicator: "#9333EA",
		Border:          "#6B7280",
	}

	styles := CreateThemeStyles(theme)

	// Test that transparent background is handled (should not error)
	if styles.Command.Render("test") == "" {
		t.Error("Command style should render text even with transparent background")
	}
}
