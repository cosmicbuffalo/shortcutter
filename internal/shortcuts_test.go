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
	shortcuts, err := getBuiltinShortcuts("zsh")
	if err != nil {
		t.Errorf("getBuiltinShortcuts(\"zsh\") returned error: %v", err)
	}

	if len(shortcuts) == 0 {
		t.Error("getBuiltinShortcuts(\"zsh\") returned empty slice")
	}

	for _, shortcut := range shortcuts {
		if shortcut.IsCustom {
			t.Errorf("Built-in shortcut %q should not be marked as custom", shortcut.Command)
		}
		if shortcut.Type != "keybinding" {
			t.Errorf("Built-in shortcut %q should have type 'keybinding', got %q", shortcut.Command, shortcut.Type)
		}
	}

	expectedShortcuts := []string{"Ctrl+A", "Ctrl+E", "Alt+F", "Alt+B", "Ctrl+R"}
	shortcutMap := make(map[string]bool)
	for _, shortcut := range shortcuts {
		shortcutMap[shortcut.Command] = true
	}

	for _, expected := range expectedShortcuts {
		if !shortcutMap[expected] {
			t.Errorf("Expected shortcut %q not found in built-in shortcuts", expected)
		}
	}

	_, err = getBuiltinShortcuts("bash")
	if err == nil {
		t.Error("getBuiltinShortcuts(\"bash\") should return error")
	}
}

func TestMergeShortcuts(t *testing.T) {
	builtins := []Shortcut{
		{Command: "Ctrl+A", Description: "Beginning of line", Action: "info", Type: "keybinding", IsCustom: false},
		{Command: "Ctrl+E", Description: "End of line", Action: "info", Type: "keybinding", IsCustom: false},
		{Command: "Alt+F", Description: "Forward word", Action: "info", Type: "keybinding", IsCustom: false},
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
		if shortcut.Command == "Ctrl+A" {
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
		if shortcut.Command == "Ctrl+E" {
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
		if shortcut.Command == "Ctrl+A" {
			ctrlAFound = true
			if shortcut.Description != "Normalized override" {
				t.Errorf("Normalized Ctrl+A: got %q, want %q", shortcut.Description, "Normalized override")
			}
		}
		if shortcut.Command == "Ctrl+E" {
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

	getShellEnv = func() string { return "/bin/zsh" }

	shortcuts, err := LoadShortcuts()
	if err != nil {
		t.Errorf("LoadShortcuts() returned error: %v", err)
	}

	if len(shortcuts) == 0 {
		t.Error("LoadShortcuts() returned empty slice")
	}

	getShellEnv = func() string { return "/bin/bash" }

	_, err = LoadShortcuts()
	if err == nil {
		t.Error("LoadShortcuts() should return error for unsupported shell")
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
		t.Errorf("DetectShortcuts() returned error: %v", err)
	}

	if len(shortcuts) == 0 {
		t.Error("DetectShortcuts() returned empty slice")
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
