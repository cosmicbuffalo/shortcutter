package internal

import (
	"os"
	"path/filepath"
	"testing"
)

func TestParseManPageDescriptions(t *testing.T) {
	sampleManPage := `ZSHZLE(1)                                                            ZSHZLE(1)

NAME
       zsh line editor

DESCRIPTION
       ZLE (Zsh Line Editor) is the command line editor for zsh.

       accept-line (^J ^M) (^J ^M) (^J ^M)
              Finish editing the buffer.  Normally, this will accept the line
              and execute it.

       accept-line-and-down-history (^O) (unbound) (unbound)
              Execute the current line, and push the next history event on the
              editing buffer stack.

       backward-char (^B) (unbound) (^B)
              Move backward one character.

       backward-delete-char (^H ^?) (unbound) (unbound)
              Delete the character behind the cursor.

       beginning-of-line (^A) (unbound) (unbound)
              Move to the beginning of the line.  If already at the beginning
              of the line, move to the beginning of the previous line, if any.

       clear-screen (^L) (unbound) (unbound)
              Clear the screen leaving the current line at the top of the
              screen.  On terminals without termcap, the terminal is cleared.

       transpose-words (^[T ^[t) (unbound) (unbound)
              Exchange the current word with the one before it.

   Movement
       These widgets move the cursor around the line.

       forward-word (^[F ^[f) (W) (^[F ^[f)
              Move to the beginning of the next word.

SEE ALSO
       zsh(1), zshcontrib(1)

AUTHOR
       This manual page was written by the Zsh development team.`

	descriptions, err := ParseManPageDescriptions(sampleManPage)
	if err != nil {
		t.Fatalf("parseManPageDescriptions() returned error: %v", err)
	}

	// Test expected descriptions
	expectedDescriptions := map[string]WidgetDescription{
		"accept-line": {
			WidgetName:       "accept-line",
			ShortDescription: "Finish editing the buffer.",
			FullDescription:  "Finish editing the buffer.  Normally, this will accept the line and execute it.",
		},
		"accept-line-and-down-history": {
			WidgetName:       "accept-line-and-down-history",
			ShortDescription: "Execute the current line, and push the next history event on the editing buffer stack.",
			FullDescription:  "Execute the current line, and push the next history event on the editing buffer stack.",
		},
		"backward-char": {
			WidgetName:       "backward-char",
			ShortDescription: "Move backward one character.",
			FullDescription:  "Move backward one character.",
		},
		"backward-delete-char": {
			WidgetName:       "backward-delete-char",
			ShortDescription: "Delete the character behind the cursor.",
			FullDescription:  "Delete the character behind the cursor.",
		},
		"beginning-of-line": {
			WidgetName:       "beginning-of-line",
			ShortDescription: "Move to the beginning of the line.",
			FullDescription:  "Move to the beginning of the line.  If already at the beginning of the line, move to the beginning of the previous line, if any.",
		},
		"clear-screen": {
			WidgetName:       "clear-screen",
			ShortDescription: "Clear the screen leaving the current line at the top of the screen.",
			FullDescription:  "Clear the screen leaving the current line at the top of the screen.  On terminals without termcap, the terminal is cleared.",
		},
		"transpose-words": {
			WidgetName:       "transpose-words",
			ShortDescription: "Exchange the current word with the one before it.",
			FullDescription:  "Exchange the current word with the one before it.",
		},
		"forward-word": {
			WidgetName:       "forward-word",
			ShortDescription: "Move to the beginning of the next word.",
			FullDescription:  "Move to the beginning of the next word.",
		},
	}

	for widget, expectedDesc := range expectedDescriptions {
		if desc, exists := descriptions[widget]; !exists {
			t.Errorf("Expected description for widget %q not found", widget)
		} else if desc.ShortDescription != expectedDesc.ShortDescription {
			t.Errorf("Short description for %q:\nGot:  %q\nWant: %q", widget, desc.ShortDescription, expectedDesc.ShortDescription)
		} else if desc.FullDescription != expectedDesc.FullDescription {
			t.Errorf("Full description for %q:\nGot:  %q\nWant: %q", widget, desc.FullDescription, expectedDesc.FullDescription)
		}
	}

	// Check that we didn't pick up non-widget entries
	nonWidgets := []string{"zsh", "ZSHZLE", "NAME", "DESCRIPTION", "Movement", "SEE ALSO", "AUTHOR"}
	for _, nonWidget := range nonWidgets {
		if _, exists := descriptions[nonWidget]; exists {
			t.Errorf("Non-widget %q should not be in descriptions", nonWidget)
		}
	}
}

func TestParseManPageDescriptionsFromFile(t *testing.T) {
	// Read test data from file
	testDataPath := filepath.Join("..", "testdata", "sample_manpage.txt")
	content, err := os.ReadFile(testDataPath)
	if err != nil {
		t.Skipf("Skipping test: could not read test data file: %v", err)
	}

	descriptions, err := ParseManPageDescriptions(string(content))
	if err != nil {
		t.Fatalf("parseManPageDescriptions() returned error: %v", err)
	}

	if len(descriptions) == 0 {
		t.Fatal("parseManPageDescriptions() returned no descriptions")
	}

	// Test some essential widgets
	essentialWidgets := []string{
		"accept-line", "beginning-of-line", "end-of-line", "forward-char", "backward-char",
		"kill-line", "backward-kill-line", "forward-word", "backward-word", "transpose-words",
	}

	for _, widget := range essentialWidgets {
		if desc, exists := descriptions[widget]; !exists {
			t.Errorf("Essential widget %q not found in descriptions", widget)
		} else if desc.ShortDescription == "" {
			t.Errorf("Widget %q has empty short description", widget)
		} else if len(desc.ShortDescription) < 10 {
			t.Errorf("Widget %q has suspiciously short description: %q", widget, desc.ShortDescription)
		}
	}
}

func TestExtractFirstSentence(t *testing.T) {
	tests := []struct {
		input    string
		expected string
		desc     string
	}{
		{
			"Finish editing the buffer. Normally, this will accept the line and execute it.",
			"Finish editing the buffer.",
			"basic sentence",
		},
		{
			"Move to the beginning of the line.  If already at the beginning of the line, move to the beginning of the previous line, if any.",
			"Move to the beginning of the line.",
			"sentence with extra spaces",
		},
		{
			"Delete the character behind the cursor",
			"Delete the character behind the cursor.",
			"no ending punctuation",
		},
		{
			"Clear the screen! This is important.",
			"Clear the screen!",
			"exclamation mark",
		},
		{
			"Why would you do this? Because it's useful.",
			"Why would you do this?",
			"question mark",
		},
		{
			"Short text",
			"Short text.",
			"short text without punctuation",
		},
		{
			"",
			"",
			"empty string",
		},
		{
			"   \n\t   ",
			"",
			"whitespace only",
		},
		{
			"This is a very long description that goes on and on and on and should be truncated because it's way too long for a short description and would be annoying to read",
			"This is a very long description that goes on and on and on and should be truncated because it's w...",
			"very long text without sentences",
		},
	}

	for _, test := range tests {
		result := extractFirstSentence(test.input)
		if result != test.expected {
			t.Errorf("extractFirstSentence(%q) = %q, want %q (%s)",
				test.input, result, test.expected, test.desc)
		}
	}
}

func TestIsNewSection(t *testing.T) {
	tests := []struct {
		line     string
		expected bool
		desc     string
	}{
		{"NAME", true, "NAME section"},
		{"DESCRIPTION", true, "DESCRIPTION section"},
		{"BUILTIN WIDGETS", true, "BUILTIN WIDGETS section"},
		{"   Movement", true, "Movement subsection"},
		{"   History Control", true, "History subsection"},
		{"accept-line (^J ^M) (^J ^M) (^J ^M)", false, "widget header"},
		{"              Finish editing the buffer.", false, "widget description"},
		{"ZSHZLE(1)", true, "man page header"},
		{"SEE ALSO", true, "SEE ALSO section"},
		{"", false, "empty line"},
		{"       random text", false, "indented text"},
	}

	for _, test := range tests {
		result := isNewSection(test.line)
		if result != test.expected {
			t.Errorf("isNewSection(%q) = %v, want %v (%s)",
				test.line, result, test.expected, test.desc)
		}
	}
}

func TestIsDescriptionLine(t *testing.T) {
	tests := []struct {
		line     string
		expected bool
		desc     string
	}{
		{"              Finish editing the buffer.", true, "widget description"},
		{"       Move backward one character.", true, "indented description"},
		{"\t\tDelete the character behind the cursor.", true, "tab-indented description"},
		{"ZSHZLE(1)", false, "man page header"},
		{"zsh: command not found", false, "error message"},
		{"Page 1", false, "page number"},
		{"short", false, "too short"},
		{"", false, "empty line"},
		{"accept-line (^J ^M)", false, "widget header (no leading space)"},
		{"    ", false, "whitespace only"},
	}

	for _, test := range tests {
		result := isDescriptionLine(test.line)
		if result != test.expected {
			t.Errorf("isDescriptionLine(%q) = %v, want %v (%s)",
				test.line, result, test.expected, test.desc)
		}
	}
}

func TestGetWidgetDescription(t *testing.T) {
	tests := []struct {
		widget   string
		expected string
		desc     string
	}{
		{"beginning-of-line", "Move to the beginning of the line.", "existing widget"},
		{"forward-word", "Move to the beginning of the next word.", "another existing widget"},
		{"nonexistent-widget", "nonexistent-widget", "fallback to widget name"},
		{"", "", "empty widget name"},
	}

	// Convert to new format
	newDescriptions := map[string]WidgetDescription{
		"beginning-of-line": {
			WidgetName:       "beginning-of-line",
			ShortDescription: "Move to the beginning of the line.",
			FullDescription:  "Move to the beginning of the line.",
		},
		"forward-word": {
			WidgetName:       "forward-word",
			ShortDescription: "Move to the beginning of the next word.",
			FullDescription:  "Move to the beginning of the next word.",
		},
	}

	for _, test := range tests {
		result := getWidgetDescription(test.widget, newDescriptions)
		if result != test.expected {
			t.Errorf("getWidgetDescription(%q) = %q, want %q (%s)",
				test.widget, result, test.expected, test.desc)
		}
	}
}

func TestWidgetDescription(t *testing.T) {
	wd := WidgetDescription{
		WidgetName:       "test-widget",
		ShortDescription: "Test short description",
		FullDescription:  "Test full description with more details",
	}

	if wd.WidgetName != "test-widget" {
		t.Errorf("WidgetName = %q, want %q", wd.WidgetName, "test-widget")
	}
	if wd.ShortDescription != "Test short description" {
		t.Errorf("ShortDescription = %q, want %q", wd.ShortDescription, "Test short description")
	}
	if wd.FullDescription != "Test full description with more details" {
		t.Errorf("FullDescription = %q, want %q", wd.FullDescription, "Test full description with more details")
	}
}

// Benchmark tests
func BenchmarkParseManPageDescriptions(b *testing.B) {
	sampleContent := `
       accept-line (^J ^M) (^J ^M) (^J ^M)
              Finish editing the buffer.  Normally, this will accept the line
              and execute it.

       backward-char (^B) (unbound) (^B)
              Move backward one character.

       beginning-of-line (^A) (unbound) (unbound)
              Move to the beginning of the line.  If already at the beginning
              of the line, move to the beginning of the previous line, if any.

       forward-word (^[F ^[f) (W) (^[F ^[f)
              Move to the beginning of the next word.
	`

	for i := 0; i < b.N; i++ {
		_, err := ParseManPageDescriptions(sampleContent)
		if err != nil {
			b.Fatalf("parseManPageDescriptions() error: %v", err)
		}
	}
}

func BenchmarkExtractFirstSentence(b *testing.B) {
	testTexts := []string{
		"Finish editing the buffer. Normally, this will accept the line and execute it.",
		"Move to the beginning of the line.  If already at the beginning of the line, move to the beginning of the previous line, if any.",
		"Delete the character behind the cursor",
		"This is a very long description that goes on and on and should be truncated",
	}

	for i := 0; i < b.N; i++ {
		for _, text := range testTexts {
			extractFirstSentence(text)
		}
	}
}