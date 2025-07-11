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
			t.Errorf("Built-in shortcut %q should not be marked as custom", shortcut.Display)
		}
		if shortcut.Type != "widget" && shortcut.Type != "sequence" {
			t.Errorf("Built-in shortcut %q should have type 'widget' or 'sequence', got %q", shortcut.Display, shortcut.Type)
		}
		if shortcut.Target == "" {
			t.Errorf("Built-in shortcut %q should have a target", shortcut.Display)
		}
	}

	expectedShortcuts := []string{"Ctrl+A", "Ctrl+E", "Alt+F", "Alt+B", "Ctrl+R"}
	shortcutMap := make(map[string]bool)
	for _, shortcut := range shortcuts {
		shortcutMap[shortcut.Display] = true
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
