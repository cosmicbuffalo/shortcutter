package internal

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestParseBindkeyOutput(t *testing.T) {
	// Test with sample bindkey output
	sampleOutput := `"^@" set-mark-command
"^A" beginning-of-line
"^B" backward-char
"^D" delete-char-or-list
"^E" end-of-line
"^F" forward-char
"^G" send-break
"^H" backward-delete-char
"^I" expand-or-complete
"^J" accept-line
"^K" kill-line
"^L" clear-screen
"^M" accept-line
"^N" down-line-or-history
"^O" accept-line-and-down-history
"^P" up-line-or-history
"^Q" push-line
"^R" history-incremental-search-backward
"^S" history-incremental-search-forward
"^T" transpose-chars
"^U" backward-kill-line
"^V" quoted-insert
"^W" backward-kill-word
"^X^E" edit-command-line
"^Y" yank
"^Z" undo
"^[" vi-cmd-mode
"^[^H" backward-kill-word
"^[." insert-last-word
"^[B" backward-word
"^[F" forward-word
"^[T" transpose-words
"^[[A" up-line-or-history
"^[[B" down-line-or-history
"^[[C" forward-char
"^[[D" backward-char
" " self-insert
"!" self-insert
"a" self-insert
"A" self-insert
"_read_comp" _read_comp`

	entries, err := parseBindkeyOutput(sampleOutput)
	if err != nil {
		t.Fatalf("parseBindkeyOutput() returned error: %v", err)
	}

	// Check that we got some entries
	if len(entries) == 0 {
		t.Fatal("parseBindkeyOutput() returned no entries")
	}

	// Test specific entries
	expectedEntries := map[string]string{
		"Ctrl+A": "beginning-of-line",
		"Ctrl+B": "backward-char",
		"Ctrl+K": "kill-line",
		"Alt+F":  "forward-word",
		"Alt+T":  "transpose-words",
		"↑":      "up-line-or-history",
		"↓":      "down-line-or-history",
	}

	entryMap := make(map[string]string)
	for _, entry := range entries {
		entryMap[entry.DisplayName] = entry.WidgetName
	}

	for displayName, expectedWidget := range expectedEntries {
		if widget, exists := entryMap[displayName]; !exists {
			t.Errorf("Expected entry for %q not found", displayName)
		} else if widget != expectedWidget {
			t.Errorf("Entry for %q: got widget %q, want %q", displayName, widget, expectedWidget)
		}
	}

	// Check that self-insert entries are filtered out
	for _, entry := range entries {
		if entry.WidgetName == "self-insert" {
			t.Errorf("self-insert widget should be filtered out, but found entry: %+v", entry)
		}
	}

	// Check that internal widgets are filtered out
	for _, entry := range entries {
		if entry.WidgetName == "_read_comp" {
			t.Errorf("internal widget should be filtered out, but found entry: %+v", entry)
		}
	}
}

func TestParseBindkeyOutputFromFile(t *testing.T) {
	// Read test data from file
	testDataPath := filepath.Join("..", "testdata", "sample_bindkey.txt")
	content, err := os.ReadFile(testDataPath)
	if err != nil {
		t.Skipf("Skipping test: could not read test data file: %v", err)
	}

	entries, err := parseBindkeyOutput(string(content))
	if err != nil {
		t.Fatalf("parseBindkeyOutput() returned error: %v", err)
	}

	if len(entries) == 0 {
		t.Fatal("parseBindkeyOutput() returned no entries")
	}

	// Test that we have expected essential entries
	essentialBindings := []string{
		"Ctrl+A", "Ctrl+E", "Ctrl+F", "Ctrl+B", "Ctrl+K", "Ctrl+U",
		"Alt+F", "Alt+B", "Alt+T", "↑", "↓", "→", "←",
	}

	entryMap := make(map[string]bool)
	for _, entry := range entries {
		entryMap[entry.DisplayName] = true
	}

	for _, essential := range essentialBindings {
		if !entryMap[essential] {
			t.Errorf("Essential binding %q not found in parsed entries", essential)
		}
	}
}

func TestShouldSkipWidget(t *testing.T) {
	tests := []struct {
		widget   string
		expected bool
		desc     string
	}{
		{"self-insert", true, "self-insert should be skipped"},
		{"self-insert-unmeta", true, "self-insert-unmeta should be skipped"},
		{"undefined-key", true, "undefined-key should be skipped"},
		{"digit-argument", true, "digit-argument should be skipped"},
		{"neg-argument", true, "neg-argument should be skipped"},
		{"universal-argument", true, "universal-argument should be skipped"},
		{"_read_comp", true, "internal widgets starting with _ should be skipped"},
		{"_history-complete-newer", true, "internal widgets starting with _ should be skipped"},
		{"vi-cmd-mode", true, "vi-mode widgets should be skipped"},
		{"vi-beginning-of-line", true, "vi-mode widgets should be skipped"},
		{"beginning-of-line", false, "normal widgets should not be skipped"},
		{"forward-word", false, "normal widgets should not be skipped"},
		{"transpose-chars", false, "normal widgets should not be skipped"},
		{"accept-line", false, "normal widgets should not be skipped"},
	}

	for _, test := range tests {
		result := shouldSkipWidget(test.widget)
		if result != test.expected {
			t.Errorf("shouldSkipWidget(%q) = %v, want %v (%s)", 
				test.widget, result, test.expected, test.desc)
		}
	}
}

func TestFilterBindkeyEntries(t *testing.T) {
	entries := []BindkeyEntry{
		{EscapeSequence: "^A", WidgetName: "beginning-of-line", DisplayName: "Ctrl+A"},
		{EscapeSequence: "^B", WidgetName: "backward-char", DisplayName: "Ctrl+B"},
		{EscapeSequence: "^A", WidgetName: "different-widget", DisplayName: "Ctrl+A"}, // duplicate
		{EscapeSequence: "a", WidgetName: "self-insert", DisplayName: "a"},             // single char
		{EscapeSequence: "", WidgetName: "some-widget", DisplayName: ""},               // empty display
		{EscapeSequence: "^[f", WidgetName: "forward-word", DisplayName: "Alt+F"},
	}

	filtered := filterBindkeyEntries(entries)

	expectedCount := 3 // Ctrl+A, Ctrl+B, Alt+F
	if len(filtered) != expectedCount {
		t.Errorf("filterBindkeyEntries() returned %d entries, want %d", len(filtered), expectedCount)
	}

	// Check that we have the expected entries
	displayNames := make(map[string]bool)
	for _, entry := range filtered {
		displayNames[entry.DisplayName] = true
	}

	expected := []string{"Ctrl+A", "Ctrl+B", "Alt+F"}
	for _, exp := range expected {
		if !displayNames[exp] {
			t.Errorf("Expected entry %q not found in filtered results", exp)
		}
	}

	// Check that duplicates are removed
	seenDisplayNames := make(map[string]int)
	for _, entry := range filtered {
		seenDisplayNames[entry.DisplayName]++
	}

	for displayName, count := range seenDisplayNames {
		if count > 1 {
			t.Errorf("Duplicate display name %q found %d times, should be unique", displayName, count)
		}
	}
}

func TestParseBindkeyOutputErrorCases(t *testing.T) {
	tests := []struct {
		input    string
		desc     string
		hasError bool
	}{
		{"", "empty input", false},                    // Should call getZshBindings and succeed
		{"invalid line\n", "invalid format", false},  // Should skip invalid lines
		{"\"^A\" beginning-of-line\n", "valid line", false},
		{"malformed \"^A beginning-of-line\n", "malformed quotes", false}, // Should skip
	}

	for _, test := range tests {
		entries, err := parseBindkeyOutput(test.input)
		
		if test.hasError && err == nil {
			t.Errorf("parseBindkeyOutput(%q) expected error but got none (%s)", test.input, test.desc)
		}
		
		if !test.hasError && err != nil {
			t.Errorf("parseBindkeyOutput(%q) unexpected error: %v (%s)", test.input, err, test.desc)
		}

		// For non-error cases, entries should not be nil (but can be empty)
		if !test.hasError && entries == nil {
			t.Errorf("parseBindkeyOutput(%q) returned nil entries but no error (%s)", test.input, test.desc)
		}
		
		// For invalid input (not empty), we should get empty slice, not nil
		if test.input == "invalid line\n" || strings.Contains(test.input, "malformed") {
			if len(entries) != 0 {
				t.Errorf("parseBindkeyOutput(%q) expected empty slice but got %d entries (%s)", 
					test.input, len(entries), test.desc)
			}
		}
	}
}

func TestBindkeyEntry(t *testing.T) {
	entry := BindkeyEntry{
		EscapeSequence: "^A",
		WidgetName:     "beginning-of-line",
		DisplayName:    "Ctrl+A",
	}

	if entry.EscapeSequence != "^A" {
		t.Errorf("EscapeSequence = %q, want %q", entry.EscapeSequence, "^A")
	}
	if entry.WidgetName != "beginning-of-line" {
		t.Errorf("WidgetName = %q, want %q", entry.WidgetName, "beginning-of-line")
	}
	if entry.DisplayName != "Ctrl+A" {
		t.Errorf("DisplayName = %q, want %q", entry.DisplayName, "Ctrl+A")
	}
}

// Benchmark tests
func BenchmarkParseBindkeyOutput(b *testing.B) {
	sampleOutput := `"^A" beginning-of-line
"^B" backward-char
"^C" send-break
"^D" delete-char-or-list
"^E" end-of-line
"^F" forward-char
"^G" send-break
"^H" backward-delete-char
"^I" expand-or-complete
"^J" accept-line
"^K" kill-line
"^L" clear-screen
"^[F" forward-word
"^[B" backward-word
"^[T" transpose-words
"^[[A" up-line-or-history
"^[[B" down-line-or-history`

	for i := 0; i < b.N; i++ {
		_, err := parseBindkeyOutput(sampleOutput)
		if err != nil {
			b.Fatalf("parseBindkeyOutput() error: %v", err)
		}
	}
}

func BenchmarkFilterBindkeyEntries(b *testing.B) {
	entries := []BindkeyEntry{
		{EscapeSequence: "^A", WidgetName: "beginning-of-line", DisplayName: "Ctrl+A"},
		{EscapeSequence: "^B", WidgetName: "backward-char", DisplayName: "Ctrl+B"},
		{EscapeSequence: "^E", WidgetName: "end-of-line", DisplayName: "Ctrl+E"},
		{EscapeSequence: "^F", WidgetName: "forward-char", DisplayName: "Ctrl+F"},
		{EscapeSequence: "^[F", WidgetName: "forward-word", DisplayName: "Alt+F"},
		{EscapeSequence: "^[B", WidgetName: "backward-word", DisplayName: "Alt+B"},
		{EscapeSequence: "a", WidgetName: "self-insert", DisplayName: "a"},
		{EscapeSequence: "b", WidgetName: "self-insert", DisplayName: "b"},
	}

	for i := 0; i < b.N; i++ {
		filterBindkeyEntries(entries)
	}
}