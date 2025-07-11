package internal

import (
	"strings"
	"unicode"
)

// normalizeEscapeSequence converts zsh escape sequences to human-readable display names
func normalizeEscapeSequence(seq string) string {
	if seq == "" {
		return ""
	}

	// Handle simple printable characters first
	if len(seq) == 1 && unicode.IsPrint(rune(seq[0])) && seq[0] != '^' {
		return seq
	}

	// Handle control sequences starting with ^
	if strings.HasPrefix(seq, "^") {
		return normalizeControlSequence(seq)
	}

	// Handle quoted sequences (remove quotes)
	if len(seq) >= 2 && seq[0] == '"' && seq[len(seq)-1] == '"' {
		return normalizeEscapeSequence(seq[1 : len(seq)-1])
	}

	return seq
}

func normalizeControlSequence(seq string) string {
	if len(seq) < 2 {
		return seq
	}

	// Handle basic control characters
	if len(seq) == 2 && seq[0] == '^' {
		char := seq[1]
		switch char {
		case '@':
			return "Ctrl+@"
		case '[':
			return "Esc"
		case '\\':
			return "Ctrl+\\"
		case ']':
			return "Ctrl+]"
		case '^':
			return "Ctrl+^"
		case '_':
			return "Ctrl+_"
		case '?':
			return "Backspace"
		default:
			if char >= 'A' && char <= 'Z' {
				return "Ctrl+" + string(char)
			}
			if char >= 'a' && char <= 'z' {
				return "Ctrl+" + strings.ToUpper(string(char))
			}
		}
	}

	// Handle Alt/Meta sequences ^[x
	if len(seq) >= 3 && seq[:2] == "^[" {
		rest := seq[2:]
		
		// Handle Alt + Control combinations like ^[^H
		if len(rest) >= 2 && rest[0] == '^' {
			ctrlPart := normalizeControlSequence(rest)
			if strings.HasPrefix(ctrlPart, "Ctrl+") {
				return "Alt+" + ctrlPart
			}
			return "Alt+Ctrl+" + rest[1:]
		}

		// Handle special arrow key sequences
		if rest == "[A" {
			return "↑"
		}
		if rest == "[B" {
			return "↓"
		}
		if rest == "[C" {
			return "→"
		}
		if rest == "[D" {
			return "←"
		}

		// Handle other bracket sequences like ^[[1~
		if strings.HasPrefix(rest, "[") {
			return normalizeSpecialSequence(rest)
		}

		// Handle single character Alt sequences
		if len(rest) == 1 {
			char := rest[0]
			if char >= 'a' && char <= 'z' {
				return "Alt+" + strings.ToUpper(string(char))
			}
			if char >= 'A' && char <= 'Z' {
				return "Alt+" + string(char)
			}
			// Handle special Alt characters
			switch char {
			case ' ':
				return "Alt+Space"
			case '.':
				return "Alt+."
			case ',':
				return "Alt+,"
			case '/':
				return "Alt+/"
			case '!':
				return "Alt+!"
			case '"':
				return "Alt+\""
			case '#':
				return "Alt+#"
			case '$':
				return "Alt+$"
			case '&':
				return "Alt+&"
			case '\'':
				return "Alt+'"
			case '-':
				return "Alt+-"
			case '<':
				return "Alt+<"
			case '>':
				return "Alt+>"
			case '?':
				return "Alt+?"
			case '|':
				return "Alt+|"
			case '~':
				return "Alt+~"
			default:
				return "Alt+" + string(char)
			}
		}

		// Handle multi-character sequences after ^[
		return "Alt+" + rest
	}

	// Handle multi-character control sequences like ^X^E
	if strings.Contains(seq, "^") && len(seq) > 2 {
		parts := strings.Split(seq, "^")
		result := ""
		for i, part := range parts {
			if part == "" && i > 0 {
				continue
			}
			if i == 0 && part == "" {
				continue
			}
			if i > 0 {
				if result != "" {
					result += " "
				}
				// Handle single character control sequences directly to avoid recursion
				if len(part) == 1 {
					char := strings.ToUpper(part)
					switch char {
					case "A":
						result += "Ctrl+A"
					case "B":
						result += "Ctrl+B"
					case "C":
						result += "Ctrl+C"
					case "D":
						result += "Ctrl+D"
					case "E":
						result += "Ctrl+E"
					case "F":
						result += "Ctrl+F"
					case "G":
						result += "Ctrl+G"
					case "H":
						result += "Ctrl+H"
					case "I":
						result += "Ctrl+I"
					case "J":
						result += "Ctrl+J"
					case "K":
						result += "Ctrl+K"
					case "L":
						result += "Ctrl+L"
					case "M":
						result += "Ctrl+M"
					case "N":
						result += "Ctrl+N"
					case "O":
						result += "Ctrl+O"
					case "P":
						result += "Ctrl+P"
					case "Q":
						result += "Ctrl+Q"
					case "R":
						result += "Ctrl+R"
					case "S":
						result += "Ctrl+S"
					case "T":
						result += "Ctrl+T"
					case "U":
						result += "Ctrl+U"
					case "V":
						result += "Ctrl+V"
					case "W":
						result += "Ctrl+W"
					case "X":
						result += "Ctrl+X"
					case "Y":
						result += "Ctrl+Y"
					case "Z":
						result += "Ctrl+Z"
					case "@":
						result += "Ctrl+@"
					case "_":
						result += "Ctrl+_"
					case "\\":
						result += "Ctrl+\\"
					case "]":
						result += "Ctrl+]"
					default:
						result += "Ctrl+" + char
					}
				} else {
					// For multi-character parts, just add as-is
					result += "^" + part
				}
			}
		}
		return result
	}

	return seq
}

func normalizeSpecialSequence(seq string) string {
	// Handle bracket sequences like [1~, [A, etc.
	if !strings.HasPrefix(seq, "[") {
		return seq
	}

	switch seq {
	case "[A":
		return "↑"
	case "[B":
		return "↓"
	case "[C":
		return "→"
	case "[D":
		return "←"
	case "[H":
		return "Home"
	case "[F":
		return "End"
	case "[1~":
		return "Home"
	case "[2~":
		return "Insert"
	case "[3~":
		return "Delete"
	case "[4~":
		return "End"
	case "[5~":
		return "Page Up"
	case "[6~":
		return "Page Down"
	default:
		return seq
	}
}