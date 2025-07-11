package internal

import (
	"testing"
)

func TestNormalizeEscapeSequence(t *testing.T) {
	tests := []struct {
		input    string
		expected string
		desc     string
	}{
		// Basic control characters
		{"^A", "Ctrl+A", "Ctrl+A"},
		{"^B", "Ctrl+B", "Ctrl+B"},
		{"^C", "Ctrl+C", "Ctrl+C"},
		{"^D", "Ctrl+D", "Ctrl+D"},
		{"^E", "Ctrl+E", "Ctrl+E"},
		{"^F", "Ctrl+F", "Ctrl+F"},
		{"^G", "Ctrl+G", "Ctrl+G"},
		{"^H", "Ctrl+H", "Ctrl+H"},
		{"^I", "Ctrl+I", "Ctrl+I (Tab)"},
		{"^J", "Ctrl+J", "Ctrl+J (Enter)"},
		{"^K", "Ctrl+K", "Ctrl+K"},
		{"^L", "Ctrl+L", "Ctrl+L"},
		{"^M", "Ctrl+M", "Ctrl+M (Enter)"},
		{"^N", "Ctrl+N", "Ctrl+N"},
		{"^O", "Ctrl+O", "Ctrl+O"},
		{"^P", "Ctrl+P", "Ctrl+P"},
		{"^Q", "Ctrl+Q", "Ctrl+Q"},
		{"^R", "Ctrl+R", "Ctrl+R"},
		{"^S", "Ctrl+S", "Ctrl+S"},
		{"^T", "Ctrl+T", "Ctrl+T"},
		{"^U", "Ctrl+U", "Ctrl+U"},
		{"^V", "Ctrl+V", "Ctrl+V"},
		{"^W", "Ctrl+W", "Ctrl+W"},
		{"^X", "Ctrl+X", "Ctrl+X"},
		{"^Y", "Ctrl+Y", "Ctrl+Y"},
		{"^Z", "Ctrl+Z", "Ctrl+Z"},

		// Special control characters
		{"^@", "Ctrl+@", "Ctrl+@ (null)"},
		{"^[", "Esc", "Escape"},
		{"^\\", "Ctrl+\\", "Ctrl+\\"},
		{"^]", "Ctrl+]", "Ctrl+]"},
		{"^^", "Ctrl+^", "Ctrl+^"},
		{"^_", "Ctrl+_", "Ctrl+_"},
		{"^?", "Backspace", "Backspace"},

		// Lowercase control characters (should be normalized to uppercase)
		{"^a", "Ctrl+A", "lowercase ctrl+a"},
		{"^b", "Ctrl+B", "lowercase ctrl+b"},
		{"^z", "Ctrl+Z", "lowercase ctrl+z"},

		// Alt/Meta sequences
		{"^[a", "Alt+A", "Alt+A"},
		{"^[b", "Alt+B", "Alt+B"},
		{"^[f", "Alt+F", "Alt+F"},
		{"^[t", "Alt+T", "Alt+T"},
		{"^[A", "Alt+A", "Alt+A uppercase"},
		{"^[F", "Alt+F", "Alt+F uppercase"},
		{"^[T", "Alt+T", "Alt+T uppercase"},

		// Alt + special characters
		{"^[ ", "Alt+Space", "Alt+Space"},
		{"^[.", "Alt+.", "Alt+."},
		{"^[,", "Alt+,", "Alt+,"},
		{"^[/", "Alt+/", "Alt+/"},
		{"^[!", "Alt+!", "Alt+!"},
		{"^[\"", "Alt+\"", "Alt+\""},
		{"^[#", "Alt+#", "Alt+#"},
		{"^[$", "Alt+$", "Alt+$"},
		{"^[&", "Alt+&", "Alt+&"},
		{"^['", "Alt+'", "Alt+'"},
		{"^[-", "Alt+-", "Alt+-"},
		{"^[<", "Alt+<", "Alt+<"},
		{"^[>", "Alt+>", "Alt+>"},
		{"^[?", "Alt+?", "Alt+?"},
		{"^[|", "Alt+|", "Alt+|"},
		{"^[~", "Alt+~", "Alt+~"},

		// Alt + Control combinations
		{"^[^H", "Alt+Ctrl+H", "Alt+Ctrl+H"},
		{"^[^D", "Alt+Ctrl+D", "Alt+Ctrl+D"},
		{"^[^G", "Alt+Ctrl+G", "Alt+Ctrl+G"},
		{"^[^I", "Alt+Ctrl+I", "Alt+Ctrl+I"},
		{"^[^J", "Alt+Ctrl+J", "Alt+Ctrl+J"},
		{"^[^M", "Alt+Ctrl+M", "Alt+Ctrl+M"},

		// Arrow keys
		{"^[[A", "↑", "Up arrow"},
		{"^[[B", "↓", "Down arrow"},
		{"^[[C", "→", "Right arrow"},
		{"^[[D", "←", "Left arrow"},
		{"^[OA", "Alt+OA", "Alt+OA (some terminals)"},
		{"^[OB", "Alt+OB", "Alt+OB (some terminals)"},
		{"^[OC", "Alt+OC", "Alt+OC (some terminals)"},
		{"^[OD", "Alt+OD", "Alt+OD (some terminals)"},

		// Function and special keys
		{"^[[H", "Home", "Home key"},
		{"^[[F", "End", "End key"},
		{"^[[1~", "Home", "Home key variant"},
		{"^[[2~", "Insert", "Insert key"},
		{"^[[3~", "Delete", "Delete key"},
		{"^[[4~", "End", "End key variant"},
		{"^[[5~", "Page Up", "Page Up key"},
		{"^[[6~", "Page Down", "Page Down key"},

		// Multi-character control sequences
		{"^X^E", "Ctrl+X Ctrl+E", "Ctrl+X Ctrl+E"},
		{"^X^R", "Ctrl+X Ctrl+R", "Ctrl+X Ctrl+R"},

		// Quoted sequences (from bindkey output)
		{"\"^A\"", "Ctrl+A", "quoted Ctrl+A"},
		{"\"^[f\"", "Alt+F", "quoted Alt+F"},
		{"\"^[[A\"", "↑", "quoted up arrow"},

		// Simple printable characters
		{"a", "a", "simple character"},
		{"A", "A", "simple uppercase"},
		{"1", "1", "digit"},
		{" ", " ", "space"},
		{"!", "!", "exclamation"},

		// Edge cases
		{"", "", "empty string"},
		{"^", "^", "lone caret"},
		{"[A", "[A", "bracket sequence without ^["},
		{"xyz", "xyz", "multi-char string"},
	}

	for _, test := range tests {
		result := normalizeEscapeSequence(test.input)
		if result != test.expected {
			t.Errorf("normalizeEscapeSequence(%q) = %q, want %q (%s)", 
				test.input, result, test.expected, test.desc)
		}
	}
}

func TestNormalizeControlSequence(t *testing.T) {
	tests := []struct {
		input    string
		expected string
	}{
		{"^A", "Ctrl+A"},
		{"^a", "Ctrl+A"},
		{"^@", "Ctrl+@"},
		{"^[", "Esc"},
		{"^?", "Backspace"},
		{"^[f", "Alt+F"},
		{"^[^H", "Alt+Ctrl+H"},
		{"^X^E", "Ctrl+X Ctrl+E"},
		{"^", "^"},
		{"", ""},
	}

	for _, test := range tests {
		result := normalizeControlSequence(test.input)
		if result != test.expected {
			t.Errorf("normalizeControlSequence(%q) = %q, want %q", 
				test.input, result, test.expected)
		}
	}
}

func TestNormalizeSpecialSequence(t *testing.T) {
	tests := []struct {
		input    string
		expected string
	}{
		{"[A", "↑"},
		{"[B", "↓"},
		{"[C", "→"},
		{"[D", "←"},
		{"[H", "Home"},
		{"[F", "End"},
		{"[1~", "Home"},
		{"[2~", "Insert"},
		{"[3~", "Delete"},
		{"[4~", "End"},
		{"[5~", "Page Up"},
		{"[6~", "Page Down"},
		{"[99~", "[99~"}, // Unknown sequence
		{"A", "A"},       // No bracket prefix
		{"", ""},         // Empty
	}

	for _, test := range tests {
		result := normalizeSpecialSequence(test.input)
		if result != test.expected {
			t.Errorf("normalizeSpecialSequence(%q) = %q, want %q", 
				test.input, result, test.expected)
		}
	}
}

// Benchmark tests
func BenchmarkNormalizeEscapeSequence(b *testing.B) {
	testCases := []string{
		"^A", "^[f", "^[[A", "^X^E", "^[^H", "\"^A\"", "a",
	}
	
	for i := 0; i < b.N; i++ {
		for _, tc := range testCases {
			normalizeEscapeSequence(tc)
		}
	}
}

func BenchmarkNormalizeControlSequence(b *testing.B) {
	testCases := []string{
		"^A", "^[f", "^X^E", "^[^H",
	}
	
	for i := 0; i < b.N; i++ {
		for _, tc := range testCases {
			normalizeControlSequence(tc)
		}
	}
}