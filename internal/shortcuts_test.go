package internal

import (
	"testing"
)

func TestNormalizeKey(t *testing.T) {
	tests := []struct {
		input    string
		expected string
	}{
		// Caret notation
		{"^A", "Ctrl+A"},
		{"^a", "Ctrl+A"},
		{"^E", "Ctrl+E"},
		{"^[", "Esc"},
		{"^I", "Tab"},
		{"^M", "Enter"},
		{"^H", "Backspace"},
		{"^@", "Ctrl+@"},
		{"^_", "Ctrl+_"},

		// C- notation
		{"C-a", "Ctrl+A"},
		{"c-a", "Ctrl+A"},
		{"C-E", "Ctrl+E"},

		// M- notation
		{"M-a", "Alt+A"},
		{"m-f", "Alt+F"},
		{"M-F", "Alt+F"},

		// Already normalized
		{"Ctrl+A", "Ctrl+A"},
		{"Alt+F", "Alt+F"},
		{"Shift+Tab", "Shift+Tab"},

		// Case variations
		{"ctrl+a", "Ctrl+A"},
		{"alt+f", "Alt+F"},
		{"CTRL+E", "Ctrl+E"},
		{"ALT+B", "Alt+B"},
		{"meta+a", "Alt+A"},

		// Special keys
		{"Tab", "Tab"},
		{"Enter", "Enter"},
		{"Home", "Home"},
		{"End", "End"},
		{"Delete", "Delete"},
		{"Backspace", "Backspace"},

		// Arrow keys
		{"↑", "↑"},
		{"↓", "↓"},
		{"←", "←"},
		{"→", "→"},

		// Whitespace handling
		{" Ctrl+A ", "Ctrl+A"},
		{" ^A ", "Ctrl+A"},
	}

	for _, test := range tests {
		result := normalizeKey(test.input)
		if result != test.expected {
			t.Errorf("normalizeKey(%q) = %q, want %q", test.input, result, test.expected)
		}
	}
}

func TestDetectShell(t *testing.T) {
	tests := []struct {
		shellEnv    string
		expected    string
		shouldError bool
	}{
		{"/bin/zsh", "zsh", false},
		{"/usr/bin/zsh", "zsh", false},
		{"/bin/bash", "", true},
		{"/usr/bin/fish", "", true},
		{"/bin/unknown", "", true},
		{"", "", true},
	}

	for _, test := range tests {
		originalGetShellEnv := getShellEnv
		getShellEnv = func() string { return test.shellEnv }

		result, err := detectShell()

		getShellEnv = originalGetShellEnv

		if test.shouldError {
			if err == nil {
				t.Errorf("detectShell() with SHELL=%q should have returned error", test.shellEnv)
			}
		} else {
			if err != nil {
				t.Errorf("detectShell() with SHELL=%q returned unexpected error: %v", test.shellEnv, err)
			}
			if result != test.expected {
				t.Errorf("detectShell() with SHELL=%q = %q, want %q", test.shellEnv, result, test.expected)
			}
		}
	}
}

func TestGetBuiltinShortcuts(t *testing.T) {
	// All shells should return error since hardcoded shortcuts are removed
	shells := []string{"zsh", "bash", "fish", "unknown"}
	
	for _, shell := range shells {
		_, err := getBuiltinShortcuts(shell)
		if err == nil {
			t.Errorf("getBuiltinShortcuts(%q) should return error since hardcoded shortcuts are removed", shell)
		}
	}
}

func TestMergeShortcuts(t *testing.T) {
	builtins := []Shortcut{
		{Display: "Ctrl+A", Description: "Beginning of line", Type: "widget", Target: "beginning-of-line", IsCustom: false},
		{Display: "Ctrl+E", Description: "End of line", Type: "widget", Target: "end-of-line", IsCustom: false},
		{Display: "Alt+F", Description: "Forward word", Type: "widget", Target: "forward-word", IsCustom: false},
	}

	emptyConfig := &Config{Shortcuts: make(map[string]interface{})}
	result := mergeShortcuts(builtins, emptyConfig)
	if len(result) != len(builtins) {
		t.Errorf("mergeShortcuts with empty config: got %d shortcuts, want %d", len(result), len(builtins))
	}

	removeConfig := &Config{
		Shortcuts: map[string]interface{}{
			"Ctrl+A": false,
		},
	}
	result = mergeShortcuts(builtins, removeConfig)
	if len(result) != len(builtins)-1 {
		t.Errorf("mergeShortcuts with removal: got %d shortcuts, want %d", len(result), len(builtins)-1)
	}

	for _, shortcut := range result {
		if shortcut.Display == "Ctrl+A" {
			t.Error("Ctrl+A should have been removed")
		}
	}

	overrideConfig := &Config{
		Shortcuts: map[string]interface{}{
			"Ctrl+E": "Custom end of line",
		},
	}
	result = mergeShortcuts(builtins, overrideConfig)
	if len(result) != len(builtins) {
		t.Errorf("mergeShortcuts with override: got %d shortcuts, want %d", len(result), len(builtins))
	}

	foundCustom := false
	for _, shortcut := range result {
		if shortcut.Display == "Ctrl+E" {
			if shortcut.Description != "Custom end of line" {
				t.Errorf("Ctrl+E description: got %q, want %q", shortcut.Description, "Custom end of line")
			}
			if !shortcut.IsCustom {
				t.Error("Ctrl+E should be marked as custom")
			}
			foundCustom = true
		}
	}
	if !foundCustom {
		t.Error("Overridden Ctrl+E not found")
	}

	addConfig := &Config{
		Shortcuts: map[string]interface{}{
			"Ctrl+X": "Custom shortcut",
		},
	}
	result = mergeShortcuts(builtins, addConfig)
	if len(result) != len(builtins)+1 {
		t.Errorf("mergeShortcuts with addition: got %d shortcuts, want %d", len(result), len(builtins)+1)
	}

	normalizeConfig := &Config{
		Shortcuts: map[string]interface{}{
			"ctrl+a": "Normalized override",
			"^E":     false,
		},
	}
	result = mergeShortcuts(builtins, normalizeConfig)

	ctrlAFound := false
	ctrlEFound := false
	for _, shortcut := range result {
		if shortcut.Display == "Ctrl+A" {
			ctrlAFound = true
			if shortcut.Description != "Normalized override" {
				t.Errorf("Normalized Ctrl+A: got %q, want %q", shortcut.Description, "Normalized override")
			}
		}
		if shortcut.Display == "Ctrl+E" {
			ctrlEFound = true
		}
	}

	if !ctrlAFound {
		t.Error("Normalized ctrl+a override not found")
	}
	if ctrlEFound {
		t.Error("^E should have been removed")
	}
}

func TestLoadShortcuts(t *testing.T) {
	originalGetShellEnv := getShellEnv
	defer func() { getShellEnv = originalGetShellEnv }()

	// Test with zsh - should succeed now that we have proper zsh integration
	getShellEnv = func() string { return "/bin/zsh" }
	shortcuts, err := LoadShortcuts()
	if err != nil {
		t.Errorf("LoadShortcuts() should succeed with zsh: %v", err)
	}
	if len(shortcuts) == 0 {
		t.Error("LoadShortcuts() should return shortcuts for zsh")
	}

	// Test with bash - should fail because only zsh is supported
	getShellEnv = func() string { return "/bin/bash" }
	_, err = LoadShortcuts()
	if err == nil {
		t.Error("LoadShortcuts() should return error for non-zsh shell")
	}

	getShellEnv = func() string { return "" }

	_, err = LoadShortcuts()
	if err == nil {
		t.Error("LoadShortcuts() should return error when SHELL not set")
	}
}

func TestDetectShortcuts(t *testing.T) {
	originalGetShellEnv := getShellEnv
	defer func() { getShellEnv = originalGetShellEnv }()

	getShellEnv = func() string { return "/bin/zsh" }

	shortcuts, err := DetectShortcuts()
	if err != nil {
		t.Errorf("DetectShortcuts() should succeed with zsh: %v", err)
	}
	if len(shortcuts) == 0 {
		t.Error("DetectShortcuts() should return shortcuts for zsh")
	}
}

func TestLoadConfig(t *testing.T) {
	config, err := loadConfig()
	if err != nil {
		t.Errorf("loadConfig() returned error: %v", err)
	}

	if config == nil {
		t.Error("loadConfig() returned nil config")
	}

	if config.Shortcuts == nil {
		t.Error("loadConfig() returned config with nil Shortcuts map")
	}
}

func TestMergeShortcutsWithObjectConfig(t *testing.T) {
	builtins := []Shortcut{
		{Display: "Ctrl+A", Description: "Beginning of line", Type: "widget", Target: "beginning-of-line", IsCustom: false},
	}

	// Test full object configuration
	objectConfig := &Config{
		Shortcuts: map[string]interface{}{
			"git-status": map[string]interface{}{
				"display":     "gs",
				"description": "Git status",
				"type":        "command",
				"target":      "git status",
			},
		},
	}
	result := mergeShortcuts(builtins, objectConfig)
	
	foundCustom := false
	for _, shortcut := range result {
		if shortcut.Display == "gs" {
			foundCustom = true
			if shortcut.Description != "Git status" {
				t.Errorf("Object config description: got %q, want %q", shortcut.Description, "Git status")
			}
			if shortcut.Type != "command" {
				t.Errorf("Object config type: got %q, want %q", shortcut.Type, "command")
			}
			if shortcut.Target != "git status" {
				t.Errorf("Object config target: got %q, want %q", shortcut.Target, "git status")
			}
			if !shortcut.IsCustom {
				t.Error("Object config shortcut should be marked as custom")
			}
		}
	}
	if !foundCustom {
		t.Error("Object config shortcut not found")
	}

	// Test partial override of built-in
	partialConfig := &Config{
		Shortcuts: map[string]interface{}{
			"Ctrl+A": map[string]interface{}{
				"description": "Start of line",
			},
		},
	}
	result = mergeShortcuts(builtins, partialConfig)
	
	foundPartial := false
	for _, shortcut := range result {
		if shortcut.Display == "Ctrl+A" {
			foundPartial = true
			if shortcut.Description != "Start of line" {
				t.Errorf("Partial override description: got %q, want %q", shortcut.Description, "Start of line")
			}
			if shortcut.Type != "widget" {
				t.Errorf("Partial override should inherit type: got %q, want %q", shortcut.Type, "widget")
			}
			if shortcut.Target != "beginning-of-line" {
				t.Errorf("Partial override should inherit target: got %q, want %q", shortcut.Target, "beginning-of-line")
			}
			if !shortcut.IsCustom {
				t.Error("Partial override should be marked as custom")
			}
		}
	}
	if !foundPartial {
		t.Error("Partial override shortcut not found")
	}
}

func TestNormalizeKeyEdgeCases(t *testing.T) {
	tests := []struct {
		input    string
		expected string
	}{
		// Edge cases and complex scenarios
		{"", ""},
		{"   ", ""},
		{"Ctrl+X Ctrl+E", "Ctrl+X Ctrl+E"},
		{"ctrl+shift+a", "Ctrl+Shift+A"},
		{"CTRL+ALT+F", "Ctrl+Alt+F"},
		{"meta+shift+b", "Alt+Shift+B"},
		{"shift+tab", "Shift+Tab"},
		{"F1", "F1"},
		{"Space", "Space"},
		{"PageUp", "PageUp"},
		{"PageDown", "PageDown"},

		// Invalid/malformed inputs
		{"^", "^"},
		{"C-", "C-"},
		{"M-", "M-"},
		{"ctrl+", "Ctrl+"},
		{"alt+", "Alt+"},

		// Multi-character keys
		{"Ctrl+Home", "Ctrl+Home"},
		{"Alt+End", "Alt+End"},
		{"Ctrl+Delete", "Ctrl+Delete"},
	}

	for _, test := range tests {
		result := normalizeKey(test.input)
		if result != test.expected {
			t.Errorf("normalizeKey(%q) = %q, want %q", test.input, result, test.expected)
		}
	}
}

func TestLoadDynamicShortcuts(t *testing.T) {
	tests := []struct {
		name      string
		shell     string
		shouldErr bool
	}{
		{"bash shell", "bash", true},  // Should fail - dynamic loading only for zsh
		{"fish shell", "fish", true},  // Should fail - dynamic loading only for zsh
		{"zsh shell", "zsh", false},   // Should succeed with zsh integration
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			shortcuts, err := loadDynamicShortcuts(tt.shell)
			
			if tt.shouldErr {
				if err == nil {
					t.Errorf("loadDynamicShortcuts(%q) should have failed", tt.shell)
				}
				return
			}
			
			if err != nil {
				t.Fatalf("loadDynamicShortcuts(%q) error: %v", tt.shell, err)
			}

			// Verify shortcuts have required fields
			for i, shortcut := range shortcuts {
				if shortcut.Display == "" {
					t.Errorf("shortcuts[%d].Display is empty", i)
				}
				if shortcut.Type == "" {
					t.Errorf("shortcuts[%d].Type is empty", i)
				}
				if shortcut.Target == "" {
					t.Errorf("shortcuts[%d].Target is empty", i)
				}
			}
		})
	}
}

func TestConvertBindkeyToShortcuts(t *testing.T) {
	bindkeyEntries := []BindkeyEntry{
		{EscapeSequence: "^A", WidgetName: "beginning-of-line", DisplayName: "Ctrl+A"},
		{EscapeSequence: "^E", WidgetName: "end-of-line", DisplayName: "Ctrl+E"},
		{EscapeSequence: "^F", WidgetName: "forward-char", DisplayName: "Ctrl+F"},
	}

	manDescriptions := map[string]WidgetDescription{
		"beginning-of-line": {
			WidgetName:       "beginning-of-line",
			ShortDescription: "Move to the beginning of the line.",
			FullDescription:  "Move to the beginning of the line.",
		},
		"end-of-line": {
			WidgetName:       "end-of-line",
			ShortDescription: "Move to the end of the line.",
			FullDescription:  "Move to the end of the line.",
		},
		// "forward-char" missing to test fallback
	}

	shortcuts := convertBindkeyToShortcuts(bindkeyEntries, manDescriptions)

	if len(shortcuts) != 3 {
		t.Fatalf("convertBindkeyToShortcuts() returned %d shortcuts, want 3", len(shortcuts))
	}

	expectedShortcuts := []struct {
		display     string
		description string
		target      string
	}{
		{"Ctrl+A", "Move to the beginning of the line.", "beginning-of-line"},
		{"Ctrl+E", "Move to the end of the line.", "end-of-line"},
		{"Ctrl+F", "forward-char", "forward-char"}, // Should fall back to widget name
	}

	for i, expected := range expectedShortcuts {
		shortcut := shortcuts[i]
		if shortcut.Display != expected.display {
			t.Errorf("shortcuts[%d].Display = %q, want %q", i, shortcut.Display, expected.display)
		}
		if shortcut.Description != expected.description {
			t.Errorf("shortcuts[%d].Description = %q, want %q", i, shortcut.Description, expected.description)
		}
		if shortcut.Target != expected.target {
			t.Errorf("shortcuts[%d].Target = %q, want %q", i, shortcut.Target, expected.target)
		}
		if shortcut.Type != "widget" {
			t.Errorf("shortcuts[%d].Type = %q, want %q", i, shortcut.Type, "widget")
		}
		if shortcut.IsCustom != false {
			t.Errorf("shortcuts[%d].IsCustom = %v, want %v", i, shortcut.IsCustom, false)
		}
	}
}

func TestConvertCacheToShortcuts(t *testing.T) {
	cacheData := &CacheData{
		BindkeyEntries: []BindkeyEntry{
			{EscapeSequence: "^A", WidgetName: "beginning-of-line", DisplayName: "Ctrl+A"},
			{EscapeSequence: "^E", WidgetName: "end-of-line", DisplayName: "Ctrl+E"},
		},
		ManDescriptions: map[string]WidgetDescription{
			"beginning-of-line": {
				WidgetName:       "beginning-of-line",
				ShortDescription: "Move to the beginning of the line.",
				FullDescription:  "Move to the beginning of the line.",
			},
			"end-of-line": {
				WidgetName:       "end-of-line",
				ShortDescription: "Move to the end of the line.",
				FullDescription:  "Move to the end of the line.",
			},
		},
	}

	shortcuts := convertCacheToShortcuts(cacheData)

	if len(shortcuts) != 2 {
		t.Fatalf("convertCacheToShortcuts() returned %d shortcuts, want 2", len(shortcuts))
	}

	// Test first shortcut
	if shortcuts[0].Display != "Ctrl+A" {
		t.Errorf("shortcuts[0].Display = %q, want %q", shortcuts[0].Display, "Ctrl+A")
	}
	if shortcuts[0].Description != "Move to the beginning of the line." {
		t.Errorf("shortcuts[0].Description = %q, want %q", shortcuts[0].Description, "Move to the beginning of the line.")
	}
	if shortcuts[0].Target != "beginning-of-line" {
		t.Errorf("shortcuts[0].Target = %q, want %q", shortcuts[0].Target, "beginning-of-line")
	}
}
