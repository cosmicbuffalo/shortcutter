package internal

import (
	"fmt"
	"os/exec"
	"strings"
)

type Shortcut struct {
	Command     string
	Description string
	Action      string // What to do: "execute" or "populate"
	Type        string // "alias", "function", "builtin", "keybinding"
}

func DetectShortcuts() ([]Shortcut, error) {
	var shortcuts []Shortcut

	// Detect aliases
	aliases, err := detectAliases()
	if err != nil {
		return nil, fmt.Errorf("failed to detect aliases: %w", err)
	}
	shortcuts = append(shortcuts, aliases...)

	// Detect functions
	functions, err := detectFunctions()
	if err != nil {
		return nil, fmt.Errorf("failed to detect functions: %w", err)
	}
	shortcuts = append(shortcuts, functions...)

	// Detect key bindings
	keybindings, err := detectKeybindings()
	if err != nil {
		return nil, fmt.Errorf("failed to detect keybindings: %w", err)
	}
	shortcuts = append(shortcuts, keybindings...)

	// Add common builtins
	builtins := getCommonBuiltins()
	shortcuts = append(shortcuts, builtins...)

	return shortcuts, nil
}

func detectAliases() ([]Shortcut, error) {
	cmd := exec.Command("zsh", "-c", "alias")
	output, err := cmd.Output()
	if err != nil {
		return nil, err
	}

	var shortcuts []Shortcut
	lines := strings.Split(string(output), "\n")
	for _, line := range lines {
		line = strings.TrimSpace(line)
		if line == "" {
			continue
		}

		// Parse alias format: alias_name='command'
		if strings.Contains(line, "=") {
			parts := strings.SplitN(line, "=", 2)
			if len(parts) == 2 {
				name := strings.TrimSpace(parts[0])
				command := strings.Trim(parts[1], "'\"")
				shortcuts = append(shortcuts, Shortcut{
					Command:     name,
					Description: fmt.Sprintf("Alias for: %s", command),
					Action:      "execute",
					Type:        "alias",
				})
			}
		}
	}

	return shortcuts, nil
}

func detectFunctions() ([]Shortcut, error) {
	cmd := exec.Command("zsh", "-c", "functions | grep -v '^_'")
	output, err := cmd.Output()
	if err != nil {
		return nil, err
	}

	var shortcuts []Shortcut
	lines := strings.Split(string(output), "\n")
	for _, line := range lines {
		line = strings.TrimSpace(line)
		if line == "" {
			continue
		}

		shortcuts = append(shortcuts, Shortcut{
			Command:     line,
			Description: "User-defined function",
			Action:      "populate",
			Type:        "function",
		})
	}

	return shortcuts, nil
}

func detectKeybindings() ([]Shortcut, error) {
	cmd := exec.Command("zsh", "-c", "bindkey")
	output, err := cmd.Output()
	if err != nil {
		return nil, err
	}

	var shortcuts []Shortcut
	lines := strings.Split(string(output), "\n")
	
	// Map of common key bindings to descriptions
	keyDescriptions := map[string]string{
		"^A":     "Beginning of line",
		"^E":     "End of line",
		"^K":     "Kill line (cut to end)",
		"^U":     "Kill line (cut to beginning)",
		"^W":     "Kill word backwards",
		"^Y":     "Yank (paste)",
		"^F":     "Forward character",
		"^B":     "Backward character",
		"^N":     "Next line",
		"^P":     "Previous line",
		"^D":     "Delete character",
		"^H":     "Backspace",
		"^L":     "Clear screen",
		"^R":     "Reverse search",
		"^S":     "Forward search",
		"^T":     "Transpose characters",
		"^Z":     "Suspend process",
		"^C":     "Cancel/interrupt",
		"^G":     "Abort",
		"^V":     "Literal next character",
		"^X^E":   "Edit command in editor",
		"^X^R":   "Read file",
		"^X^U":   "Undo",
		"^X^X":   "Exchange point and mark",
		"\\e[A":  "Up arrow (previous command)",
		"\\e[B":  "Down arrow (next command)",
		"\\e[C":  "Right arrow (forward char)",
		"\\e[D":  "Left arrow (backward char)",
		"\\e[3~": "Delete key",
		"\\e[H":  "Home key",
		"\\e[F":  "End key",
		"\\eb":   "Alt+B (backward word)",
		"\\ef":   "Alt+F (forward word)",
		"\\ed":   "Alt+D (delete word)",
		"\\e\\177": "Alt+Backspace (delete word backwards)",
	}

	for _, line := range lines {
		line = strings.TrimSpace(line)
		if line == "" {
			continue
		}

		// Parse bindkey format: "key" function
		if strings.Contains(line, "\"") {
			parts := strings.Fields(line)
			if len(parts) >= 2 {
				key := strings.Trim(parts[0], "\"")
				if description, exists := keyDescriptions[key]; exists {
					shortcuts = append(shortcuts, Shortcut{
						Command:     formatKeyForDisplay(key),
						Description: description,
						Action:      "info",
						Type:        "keybinding",
					})
				}
			}
		}
	}

	return shortcuts, nil
}

func formatKeyForDisplay(key string) string {
	// Convert control characters to readable format
	replacements := map[string]string{
		"^A": "Ctrl+A",
		"^B": "Ctrl+B",
		"^C": "Ctrl+C",
		"^D": "Ctrl+D",
		"^E": "Ctrl+E",
		"^F": "Ctrl+F",
		"^G": "Ctrl+G",
		"^H": "Ctrl+H",
		"^K": "Ctrl+K",
		"^L": "Ctrl+L",
		"^N": "Ctrl+N",
		"^P": "Ctrl+P",
		"^R": "Ctrl+R",
		"^S": "Ctrl+S",
		"^T": "Ctrl+T",
		"^U": "Ctrl+U",
		"^V": "Ctrl+V",
		"^W": "Ctrl+W",
		"^Y": "Ctrl+Y",
		"^Z": "Ctrl+Z",
		"^X^E": "Ctrl+X Ctrl+E",
		"^X^R": "Ctrl+X Ctrl+R",
		"^X^U": "Ctrl+X Ctrl+U",
		"^X^X": "Ctrl+X Ctrl+X",
		"\\e[A": "↑",
		"\\e[B": "↓",
		"\\e[C": "→",
		"\\e[D": "←",
		"\\e[3~": "Delete",
		"\\e[H": "Home",
		"\\e[F": "End",
		"\\eb": "Alt+B",
		"\\ef": "Alt+F",
		"\\ed": "Alt+D",
		"\\e\\177": "Alt+Backspace",
	}

	if display, exists := replacements[key]; exists {
		return display
	}
	return key
}

func getCommonBuiltins() []Shortcut {
	return []Shortcut{
		{Command: "cd", Description: "Change directory", Action: "populate", Type: "builtin"},
		{Command: "ls", Description: "List directory contents", Action: "populate", Type: "builtin"},
		{Command: "pwd", Description: "Print working directory", Action: "execute", Type: "builtin"},
		{Command: "echo", Description: "Display text", Action: "populate", Type: "builtin"},
		{Command: "cat", Description: "Display file contents", Action: "populate", Type: "builtin"},
		{Command: "grep", Description: "Search text patterns", Action: "populate", Type: "builtin"},
		{Command: "find", Description: "Find files and directories", Action: "populate", Type: "builtin"},
		{Command: "ps", Description: "Show running processes", Action: "populate", Type: "builtin"},
		{Command: "kill", Description: "Terminate processes", Action: "populate", Type: "builtin"},
		{Command: "top", Description: "Display running processes", Action: "execute", Type: "builtin"},
		{Command: "history", Description: "Show command history", Action: "execute", Type: "builtin"},
		{Command: "which", Description: "Locate command", Action: "populate", Type: "builtin"},
		{Command: "man", Description: "Show manual page", Action: "populate", Type: "builtin"},
		{Command: "chmod", Description: "Change file permissions", Action: "populate", Type: "builtin"},
		{Command: "chown", Description: "Change file ownership", Action: "populate", Type: "builtin"},
		{Command: "cp", Description: "Copy files", Action: "populate", Type: "builtin"},
		{Command: "mv", Description: "Move/rename files", Action: "populate", Type: "builtin"},
		{Command: "rm", Description: "Remove files", Action: "populate", Type: "builtin"},
		{Command: "mkdir", Description: "Create directory", Action: "populate", Type: "builtin"},
		{Command: "rmdir", Description: "Remove directory", Action: "populate", Type: "builtin"},
		{Command: "touch", Description: "Create empty file", Action: "populate", Type: "builtin"},
		{Command: "head", Description: "Show first lines of file", Action: "populate", Type: "builtin"},
		{Command: "tail", Description: "Show last lines of file", Action: "populate", Type: "builtin"},
		{Command: "less", Description: "View file contents (paginated)", Action: "populate", Type: "builtin"},
		{Command: "more", Description: "View file contents (paginated)", Action: "populate", Type: "builtin"},
		{Command: "sort", Description: "Sort lines of text", Action: "populate", Type: "builtin"},
		{Command: "uniq", Description: "Remove duplicate lines", Action: "populate", Type: "builtin"},
		{Command: "wc", Description: "Word, line, character count", Action: "populate", Type: "builtin"},
		{Command: "du", Description: "Show disk usage", Action: "populate", Type: "builtin"},
		{Command: "df", Description: "Show filesystem disk space", Action: "execute", Type: "builtin"},
		{Command: "mount", Description: "Mount filesystem", Action: "populate", Type: "builtin"},
		{Command: "umount", Description: "Unmount filesystem", Action: "populate", Type: "builtin"},
		{Command: "tar", Description: "Archive files", Action: "populate", Type: "builtin"},
		{Command: "gzip", Description: "Compress files", Action: "populate", Type: "builtin"},
		{Command: "gunzip", Description: "Decompress files", Action: "populate", Type: "builtin"},
		{Command: "zip", Description: "Create zip archive", Action: "populate", Type: "builtin"},
		{Command: "unzip", Description: "Extract zip archive", Action: "populate", Type: "builtin"},
		{Command: "wget", Description: "Download files from web", Action: "populate", Type: "builtin"},
		{Command: "curl", Description: "Transfer data from servers", Action: "populate", Type: "builtin"},
		{Command: "ssh", Description: "Secure shell remote access", Action: "populate", Type: "builtin"},
		{Command: "scp", Description: "Secure copy files", Action: "populate", Type: "builtin"},
		{Command: "rsync", Description: "Synchronize files", Action: "populate", Type: "builtin"},
		{Command: "git", Description: "Version control system", Action: "populate", Type: "builtin"},
		{Command: "vim", Description: "Text editor", Action: "populate", Type: "builtin"},
		{Command: "nano", Description: "Simple text editor", Action: "populate", Type: "builtin"},
		{Command: "emacs", Description: "Text editor", Action: "populate", Type: "builtin"},
		{Command: "code", Description: "VS Code editor", Action: "populate", Type: "builtin"},
	}
}