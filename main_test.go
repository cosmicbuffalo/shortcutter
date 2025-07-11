package main

import (
	"shortcutter/internal"
	"testing"
)

func TestDetectShortcuts(t *testing.T) {
	shortcuts, err := internal.DetectShortcuts()
	if err != nil {
		t.Errorf("DetectShortcuts() returned error: %v", err)
	}

	if len(shortcuts) == 0 {
		t.Error("DetectShortcuts() returned empty slice")
	}

	for i, shortcut := range shortcuts {
		if shortcut.Display == "" {
			t.Errorf("Shortcut %d has empty Display", i)
		}
		if shortcut.Description == "" {
			t.Errorf("Shortcut %d (%s) has empty Description", i, shortcut.Display)
		}
		if shortcut.Type == "" {
			t.Errorf("Shortcut %d (%s) has empty Type", i, shortcut.Display)
		}
		if shortcut.Target == "" {
			t.Errorf("Shortcut %d (%s) has empty Target", i, shortcut.Display)
		}
	}
}

func TestMainIntegration(t *testing.T) {
	shortcuts, err := internal.DetectShortcuts()
	if err != nil {
		t.Errorf("Main integration test failed: %v", err)
	}

	if len(shortcuts) > 0 {
		theme := internal.GetDefaultTheme()
		styles := internal.CreateThemeStyles(theme)
		model := internal.InitialModel(shortcuts, styles)

		if len(model.Shortcuts()) != len(shortcuts) {
			t.Errorf("Model shortcuts count: got %d, want %d", len(model.Shortcuts()), len(shortcuts))
		}

		view := model.View()
		if view == "" {
			t.Error("Model view should not be empty")
		}
	}
}
