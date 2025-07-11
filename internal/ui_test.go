package internal

import (
	"strings"
	"testing"

	tea "github.com/charmbracelet/bubbletea"
)

func createTestModel(shortcuts []Shortcut) model {
	theme := GetDefaultTheme()
	styles := CreateThemeStyles(theme)
	return InitialModel(shortcuts, styles)
}

func TestInitialModel(t *testing.T) {
	shortcuts := []Shortcut{
		{Display: "Ctrl+A", Description: "Beginning of line", Type: "widget", Target: "beginning-of-line", IsCustom: false},
		{Display: "Ctrl+E", Description: "End of line", Type: "widget", Target: "end-of-line", IsCustom: false},
		{Display: "Alt+F", Description: "Forward word", Type: "widget", Target: "forward-word", IsCustom: true},
	}

	model := createTestModel(shortcuts)

	if len(model.shortcuts) != len(shortcuts) {
		t.Errorf("InitialModel shortcuts: got %d, want %d", len(model.shortcuts), len(shortcuts))
	}

	if len(model.filtered) != len(shortcuts) {
		t.Errorf("InitialModel filtered: got %d, want %d", len(model.filtered), len(shortcuts))
	}

	if model.cursor != 0 {
		t.Errorf("InitialModel cursor: got %d, want 0", model.cursor)
	}

	if model.query != "" {
		t.Errorf("InitialModel query: got %q, want empty string", model.query)
	}

	if model.maxVisible != 10 {
		t.Errorf("InitialModel maxVisible: got %d, want 10", model.maxVisible)
	}

	if model.selected != nil {
		t.Error("InitialModel selected should be nil")
	}

	if model.quitting {
		t.Error("InitialModel quitting should be false")
	}
}

func TestFilterShortcuts(t *testing.T) {
	shortcuts := []Shortcut{
		{Display: "Ctrl+A", Description: "Beginning of line", Type: "widget", Target: "beginning-of-line"},
		{Display: "Ctrl+E", Description: "End of line", Type: "widget", Target: "end-of-line"},
		{Display: "Alt+F", Description: "Forward word", Type: "widget", Target: "forward-word"},
		{Display: "Tab", Description: "Complete command", Type: "widget", Target: "expand-or-complete"},
	}

	model := createTestModel(shortcuts)

	model.query = ""
	filtered := model.filterShortcuts()
	if len(filtered) != len(shortcuts) {
		t.Errorf("Empty query filter: got %d shortcuts, want %d", len(filtered), len(shortcuts))
	}

	model.query = "Ctrl"
	filtered = model.filterShortcuts()
	if len(filtered) == 0 {
		t.Error("Query 'Ctrl' should match some shortcuts")
	}

	foundMatch := false
	for _, shortcut := range filtered {
		if shortcut.Display == "Ctrl+A" || shortcut.Display == "Ctrl+E" {
			foundMatch = true
			break
		}
	}
	if !foundMatch {
		t.Error("Query 'Ctrl' should match Ctrl+A or Ctrl+E")
	}

	model.query = "word"
	filtered = model.filterShortcuts()
	foundMatch = false
	for _, shortcut := range filtered {
		if shortcut.Display == "Alt+F" {
			foundMatch = true
			break
		}
	}
	if !foundMatch {
		t.Error("Query 'word' should match Alt+F (Forward word)")
	}

	model.query = "zzzzz"
	filtered = model.filterShortcuts()
	if len(filtered) != 0 {
		t.Errorf("Query 'zzzzz' should match no shortcuts, got %d", len(filtered))
	}
}

func TestHighlightMatches(t *testing.T) {
	shortcuts := []Shortcut{
		{Display: "Ctrl+A", Description: "Beginning of line", Type: "widget", Target: "beginning-of-line"},
	}

	model := createTestModel(shortcuts)

	result := model.highlightMatches("Ctrl+A", "", model.styles.Command, false, model.styles)
	if result == "" {
		t.Error("highlightMatches should not return empty string")
	}

	result = model.highlightMatches("Ctrl+A", "Ctrl", model.styles.Command, false, model.styles)
	if result == "" {
		t.Error("highlightMatches should not return empty string")
	}

	result = model.highlightMatches("Ctrl+A", "Ctrl", model.styles.Command, true, model.styles)
	if result == "" {
		t.Error("highlightMatches should not return empty string")
	}
}

func TestModelView(t *testing.T) {
	shortcuts := []Shortcut{
		{Display: "Ctrl+A", Description: "Beginning of line", Type: "widget", Target: "beginning-of-line", IsCustom: false},
		{Display: "Ctrl+E", Description: "End of line", Type: "widget", Target: "end-of-line", IsCustom: true},
	}

	model := createTestModel(shortcuts)
	model.width = 80
	model.height = 25

	view := model.View()
	if view == "" {
		t.Error("View should not return empty string")
	}

	if !strings.Contains(view, "❯") {
		t.Error("View should contain query prompt '❯'")
	}

	if !strings.Contains(view, "Ctrl+A") {
		t.Error("View should contain Ctrl+A shortcut")
	}

	if !strings.Contains(view, "2/2") {
		t.Error("View should contain status '2/2'")
	}

	if !strings.Contains(view, "navigate") {
		t.Error("View should contain help text")
	}

	model.query = "Ctrl"
	model.filtered = model.filterShortcuts()
	view = model.View()
	if !strings.Contains(view, "Ctrl") {
		t.Error("View should contain query in prompt")
	}

	model.quitting = true
	view = model.View()
	if view != "" {
		t.Error("View should return empty string when quitting")
	}
}

func TestModelInit(t *testing.T) {
	shortcuts := []Shortcut{
		{Display: "Ctrl+A", Description: "Beginning of line", Type: "widget", Target: "beginning-of-line"},
	}

	model := createTestModel(shortcuts)
	cmd := model.Init()
	if cmd != nil {
		t.Error("Init should return nil command")
	}
}

func TestModelUpdate(t *testing.T) {
	shortcuts := []Shortcut{
		{Display: "Ctrl+A", Description: "Beginning of line", Type: "widget", Target: "beginning-of-line"},
		{Display: "Ctrl+E", Description: "End of line", Type: "widget", Target: "end-of-line"},
		{Display: "Alt+F", Description: "Forward word", Type: "widget", Target: "forward-word"},
	}

	model := createTestModel(shortcuts)

	_, _ = model.Update(tea.WindowSizeMsg{Width: 100, Height: 30})
	_, cmd := model.Update(tea.KeyMsg{Type: tea.KeyEsc})

	if cmd == nil {
		t.Error("Escape key should return a command")
	}
}
