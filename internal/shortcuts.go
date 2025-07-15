package internal

import (
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"sort"
	"strings"

	"github.com/BurntSushi/toml"
)

type Shortcut struct {
	Display         string // What to show in UI (e.g., "Ctrl+A", "gs")
	Description     string // Human-readable short description
	FullDescription string // Complete description from manual
	Type            string // "widget", "command", or "sequence"
	Target          string // What to execute (widget name, command, or key sequence)
	IsCustom        bool   // True if added/modified by user config
}

type Config struct {
	Shortcuts map[string]interface{} `toml:"shortcuts"`
	Theme     ThemeConfig            `toml:"theme"`
}

type ThemeConfig struct {
	Name string `toml:"name"`
}

func LoadShortcuts() ([]Shortcut, error) {
	shell, err := detectShell()
	if err != nil {
		return nil, err
	}

	// Get hardcoded shortcuts for the detected shell
	builtinShortcuts, err := getBuiltinShortcuts(shell)
	if err != nil {
		return nil, fmt.Errorf("failed to load builtin shortcuts: %w", err)
	}

	// Get man page descriptions to enhance the shortcuts
	manDescriptions, err := getWidgetDescriptions()
	if err != nil {
		// Don't fail if we can't get man page descriptions, just use hardcoded ones
		fmt.Fprintf(os.Stderr, "Warning: Failed to get widget descriptions: %v\n", err)
		manDescriptions = make(map[string]WidgetDescription)
	}

	// Enhance shortcuts with man page descriptions
	enhancedShortcuts := enhanceShortcutsWithManPages(builtinShortcuts, manDescriptions)

	// Load user config
	config, err := loadConfig()
	if err != nil {
		return nil, err
	}

	// Merge with user config
	shortcuts := mergeShortcuts(enhancedShortcuts, config)

	return shortcuts, nil
}

// enhanceShortcutsWithManPages enhances hardcoded shortcuts with man page descriptions
func enhanceShortcutsWithManPages(shortcuts []Shortcut, manDescriptions map[string]WidgetDescription) []Shortcut {
	enhanced := make([]Shortcut, len(shortcuts))
	
	for i, shortcut := range shortcuts {
		enhanced[i] = shortcut
		
		// Only enhance widget type shortcuts
		if shortcut.Type == "widget" {
			if manDesc, exists := manDescriptions[shortcut.Target]; exists {
				// Replace with man page descriptions
				enhanced[i].Description = manDesc.ShortDescription
				enhanced[i].FullDescription = manDesc.FullDescription
			}
		}
	}
	
	return enhanced
}

func LoadShortcutsAndTheme() ([]Shortcut, ThemeStyles, error) {
	shortcuts, err := LoadShortcuts()
	if err != nil {
		return nil, ThemeStyles{}, err
	}

	config, err := loadConfig()
	if err != nil {
		defaultTheme := GetDefaultTheme()
		styles := CreateThemeStyles(defaultTheme)
		return shortcuts, styles, nil
	}

	themeName := config.Theme.Name
	if themeName == "" {
		themeName = "default"
	}

	theme, err := LoadTheme(themeName)
	if err != nil {
		theme = GetDefaultTheme()
	}

	styles := CreateThemeStyles(theme)

	return shortcuts, styles, nil
}

func detectShell() (string, error) {
	shell := getShellEnv()
	if shell == "" {
		return "", fmt.Errorf("SHELL environment variable not set")
	}

	shellName := filepath.Base(shell)

	switch shellName {
	case "zsh":
		return "zsh", nil
	case "bash":
		return "", fmt.Errorf("bash support not implemented yet - please use zsh")
	case "fish":
		return "", fmt.Errorf("fish support not implemented yet - please use zsh")
	default:
		return "", fmt.Errorf("unsupported shell '%s' - only zsh is supported", shellName)
	}
}

func getBuiltinShortcuts(shell string) ([]Shortcut, error) {
	switch shell {
	case "zsh":
		return getZshBuiltinShortcuts(), nil
	case "bash":
		return getBashBuiltinShortcuts(), nil
	default:
		return getGenericBuiltinShortcuts(), nil
	}
}


func getZshBuiltinShortcuts() []Shortcut {
	return []Shortcut{
		{Display: "Ctrl+@", Description: "Set Mark", FullDescription: "Set the mark at the cursor position.", Type: "widget", Target: "set-mark-command", IsCustom: false},
		{Display: "Ctrl+A", Description: "Beginning of line", FullDescription: "Move to the beginning of the line.", Type: "widget", Target: "beginning-of-line", IsCustom: false},
		{Display: "Ctrl+B", Description: "Back one character", FullDescription: "Move backward one character.", Type: "widget", Target: "backward-char", IsCustom: false},
		{Display: "Ctrl+D", Description: "Delete character or list", FullDescription: "Delete the character under the cursor, or list possible completions.", Type: "widget", Target: "delete-char-or-list", IsCustom: false},
		{Display: "Ctrl+E", Description: "End of line", FullDescription: "Move to the end of the line.", Type: "widget", Target: "end-of-line", IsCustom: false},
		{Display: "Ctrl+F", Description: "Forward one character", FullDescription: "Move forward one character.", Type: "widget", Target: "forward-char", IsCustom: false},
		{Display: "Ctrl+G", Description: "Send break", FullDescription: "Abort the current editor function.", Type: "widget", Target: "send-break", IsCustom: false},
		{Display: "Ctrl+H", Description: "Delete character backward", FullDescription: "Delete the character behind the cursor.", Type: "widget", Target: "backward-delete-char", IsCustom: false},
		{Display: "Ctrl+I", Description: "Expand or complete prefix", FullDescription: "Attempt shell expansion on the current word. If that fails, attempt completion.", Type: "widget", Target: "expand-or-complete-prefix", IsCustom: false},
		{Display: "Ctrl+J", Description: "Accept line", FullDescription: "Finish editing the buffer. Normally, this will accept the line and execute it.", Type: "widget", Target: "accept-line", IsCustom: false},
		{Display: "Ctrl+K", Description: "Kill line", FullDescription: "Kill from the cursor to the end of the line.", Type: "widget", Target: "kill-line", IsCustom: false},
		{Display: "Ctrl+L", Description: "Clear screen", FullDescription: "Clear the screen leaving the current line at the top of the screen.", Type: "widget", Target: "clear-screen", IsCustom: false},
		{Display: "Ctrl+M", Description: "Accept line", FullDescription: "Finish editing the buffer. Normally, this will accept the line and execute it.", Type: "widget", Target: "accept-line", IsCustom: false},
		{Display: "Ctrl+N", Description: "Down line or history", FullDescription: "Move down a line in the buffer, or if already at the bottom line, move to the next event in the history list.", Type: "widget", Target: "down-line-or-history", IsCustom: false},
		{Display: "Ctrl+O", Description: "Accept line and down history", FullDescription: "Execute the current line, and push the next history event on the editing buffer stack.", Type: "widget", Target: "accept-line-and-down-history", IsCustom: false},
		{Display: "Ctrl+P", Description: "Up line or history", FullDescription: "Move up a line in the buffer, or if already at the top line, move to the previous event in the history list.", Type: "widget", Target: "up-line-or-history", IsCustom: false},
		{Display: "Ctrl+Q", Description: "Push line", FullDescription: "Push the current line onto the buffer stack and clear the line.", Type: "widget", Target: "push-line", IsCustom: false},
		{Display: "Ctrl+R", Description: "Atuin search", FullDescription: "Search backward incrementally for a specified string using atuin.", Type: "widget", Target: "atuin-search", IsCustom: false},
		{Display: "Ctrl+S", Description: "History incremental search forward", FullDescription: "Search forward incrementally for a specified string.", Type: "widget", Target: "history-incremental-search-forward", IsCustom: false},
		{Display: "Ctrl+T", Description: "Atuin search", FullDescription: "Search backward incrementally for a specified string using atuin.", Type: "widget", Target: "atuin-search", IsCustom: false},
		{Display: "Ctrl+U", Description: "Backward kill line", FullDescription: "Kill from the beginning of the line to the cursor position.", Type: "widget", Target: "backward-kill-line", IsCustom: false},
		{Display: "Ctrl+V", Description: "Quoted insert", FullDescription: "Insert the next character typed, even if it is a special character.", Type: "widget", Target: "quoted-insert", IsCustom: false},
		{Display: "Ctrl+W", Description: "Backward kill word", FullDescription: "Kill the word behind the cursor.", Type: "widget", Target: "_backward-kill-word", IsCustom: false},
		{Display: "Ctrl+X Ctrl+B", Description: "Vi match bracket", FullDescription: "Move to the bracket character that matches the one under the cursor.", Type: "widget", Target: "vi-match-bracket", IsCustom: false},
		{Display: "Ctrl+X Ctrl+E", Description: "Edit command line", FullDescription: "Edit the command line using the editor specified by the EDITOR or VISUAL environment variable.", Type: "widget", Target: "edit-command-line", IsCustom: false},
		{Display: "Ctrl+X Ctrl+F", Description: "Vi find next char", FullDescription: "Read a character and move to the next occurrence of it on the line.", Type: "widget", Target: "vi-find-next-char", IsCustom: false},
		{Display: "Ctrl+X Ctrl+J", Description: "Vi join", FullDescription: "Join the current line with the next line.", Type: "widget", Target: "vi-join", IsCustom: false},
		{Display: "Ctrl+X Ctrl+K", Description: "Kill buffer", FullDescription: "Kill the entire buffer.", Type: "widget", Target: "kill-buffer", IsCustom: false},
		{Display: "Ctrl+X Ctrl+N", Description: "Infer next history", FullDescription: "Infer the next history event.", Type: "widget", Target: "infer-next-history", IsCustom: false},
		{Display: "Ctrl+X Ctrl+O", Description: "Overwrite mode", FullDescription: "Toggle overwrite mode.", Type: "widget", Target: "overwrite-mode", IsCustom: false},
		{Display: "Ctrl+X Ctrl+R", Description: "Read comp", FullDescription: "Read a completion specification.", Type: "widget", Target: "_read_comp", IsCustom: false},
		{Display: "Ctrl+X Ctrl+U", Description: "Undo", FullDescription: "Undo the last change to the line.", Type: "widget", Target: "undo", IsCustom: false},
		{Display: "Ctrl+X Ctrl+V", Description: "Vi command mode", FullDescription: "Enter vi command mode.", Type: "widget", Target: "vi-cmd-mode", IsCustom: false},
		{Display: "Ctrl+X Ctrl+X", Description: "Exchange point and mark", FullDescription: "Exchange the cursor position with the mark.", Type: "widget", Target: "exchange-point-and-mark", IsCustom: false},
		{Display: "Ctrl+X *", Description: "Expand word", FullDescription: "Expand the current word.", Type: "widget", Target: "expand-word", IsCustom: false},
		{Display: "Ctrl+X =", Description: "What cursor position", FullDescription: "Display information about the cursor position.", Type: "widget", Target: "what-cursor-position", IsCustom: false},
		{Display: "Ctrl+X ?", Description: "Complete debug", FullDescription: "Display completion debugging information.", Type: "widget", Target: "_complete_debug", IsCustom: false},
		{Display: "Ctrl+X C", Description: "Correct filename", FullDescription: "Correct the current filename.", Type: "widget", Target: "_correct_filename", IsCustom: false},
		{Display: "Ctrl+X G", Description: "List expand", FullDescription: "List possible expansions.", Type: "widget", Target: "list-expand", IsCustom: false},
		{Display: "Ctrl+X a", Description: "Expand alias", FullDescription: "Expand the current alias.", Type: "widget", Target: "_expand_alias", IsCustom: false},
		{Display: "Ctrl+X c", Description: "Correct word", FullDescription: "Correct the current word.", Type: "widget", Target: "_correct_word", IsCustom: false},
		{Display: "Ctrl+X d", Description: "List expansions", FullDescription: "List possible expansions.", Type: "widget", Target: "_list_expansions", IsCustom: false},
		{Display: "Ctrl+X e", Description: "Expand word", FullDescription: "Expand the current word.", Type: "widget", Target: "_expand_word", IsCustom: false},
		{Display: "Ctrl+X g", Description: "List expand", FullDescription: "List possible expansions.", Type: "widget", Target: "list-expand", IsCustom: false},
		{Display: "Ctrl+X h", Description: "Complete help", FullDescription: "Display help for completion.", Type: "widget", Target: "_complete_help", IsCustom: false},
		{Display: "Ctrl+X m", Description: "Most recent file", FullDescription: "Complete with the most recent file.", Type: "widget", Target: "_most_recent_file", IsCustom: false},
		{Display: "Ctrl+X n", Description: "Next tags", FullDescription: "Move to the next completion tag.", Type: "widget", Target: "_next_tags", IsCustom: false},
		{Display: "Ctrl+X r", Description: "History incremental search backward", FullDescription: "Search backward incrementally for a specified string.", Type: "widget", Target: "history-incremental-search-backward", IsCustom: false},
		{Display: "Ctrl+X s", Description: "History incremental search forward", FullDescription: "Search forward incrementally for a specified string.", Type: "widget", Target: "history-incremental-search-forward", IsCustom: false},
		{Display: "Ctrl+X t", Description: "Complete tag", FullDescription: "Complete with a tag.", Type: "widget", Target: "_complete_tag", IsCustom: false},
		{Display: "Ctrl+X u", Description: "Undo", FullDescription: "Undo the last change to the line.", Type: "widget", Target: "undo", IsCustom: false},
		{Display: "Ctrl+X ~", Description: "Bash list choices", FullDescription: "List possible completions using bash-style completion.", Type: "widget", Target: "_bash_list-choices", IsCustom: false},
		{Display: "Ctrl+Y", Description: "Yank", FullDescription: "Insert the most recently killed text at the cursor position.", Type: "widget", Target: "yank", IsCustom: false},
		{Display: "Alt+Ctrl+D", Description: "List choices", FullDescription: "List possible completions for the current word.", Type: "widget", Target: "list-choices", IsCustom: false},
		{Display: "Alt+Ctrl+G", Description: "Send break", FullDescription: "Abort the current editor function.", Type: "widget", Target: "send-break", IsCustom: false},
		{Display: "Alt+Ctrl+H", Description: "Backward kill word", FullDescription: "Kill the word behind the cursor.", Type: "widget", Target: "backward-kill-word", IsCustom: false},
		{Display: "Alt+Ctrl+I", Description: "Self insert unmeta", FullDescription: "Insert the character without the meta bit.", Type: "widget", Target: "self-insert-unmeta", IsCustom: false},
		{Display: "Alt+Ctrl+J", Description: "Self insert unmeta", FullDescription: "Insert the character without the meta bit.", Type: "widget", Target: "self-insert-unmeta", IsCustom: false},
		{Display: "Alt+Ctrl+L", Description: "Clear screen", FullDescription: "Clear the screen leaving the current line at the top of the screen.", Type: "widget", Target: "clear-screen", IsCustom: false},
		{Display: "Alt+Ctrl+M", Description: "Self insert unmeta", FullDescription: "Insert the character without the meta bit.", Type: "widget", Target: "self-insert-unmeta", IsCustom: false},
		{Display: "Alt+Ctrl+_", Description: "Copy prev word", FullDescription: "Copy the previous word to the cursor position.", Type: "widget", Target: "copy-prev-word", IsCustom: false},
		{Display: "Alt+Space", Description: "Expand history", FullDescription: "Expand history substitutions in the current line.", Type: "widget", Target: "expand-history", IsCustom: false},
		{Display: "Alt+!", Description: "Expand history", FullDescription: "Expand history substitutions in the current line.", Type: "widget", Target: "expand-history", IsCustom: false},
		{Display: "Alt+\"", Description: "Quote region", FullDescription: "Quote the region between the cursor and the mark.", Type: "widget", Target: "quote-region", IsCustom: false},
		{Display: "Alt+$", Description: "Spell word", FullDescription: "Attempt spelling correction on the current word.", Type: "widget", Target: "spell-word", IsCustom: false},
		{Display: "Alt+'", Description: "Quote line", FullDescription: "Quote the current line.", Type: "widget", Target: "quote-line", IsCustom: false},
		{Display: "Alt+,", Description: "History complete newer", FullDescription: "Complete from newer history entries.", Type: "widget", Target: "_history-complete-newer", IsCustom: false},
		{Display: "Alt+-", Description: "Neg argument", FullDescription: "Begin a negative numeric argument.", Type: "widget", Target: "neg-argument", IsCustom: false},
		{Display: "Alt+.", Description: "Insert last word", FullDescription: "Insert the last word from the previous history event at the cursor position.", Type: "widget", Target: "insert-last-word", IsCustom: false},
		{Display: "Alt+/", Description: "History complete older", FullDescription: "Complete from older history entries.", Type: "widget", Target: "_history-complete-older", IsCustom: false},
		{Display: "Alt+0", Description: "Digit argument", FullDescription: "Begin a numeric argument with the given digit.", Type: "widget", Target: "digit-argument", IsCustom: false},
		{Display: "Alt+1", Description: "Digit argument", FullDescription: "Begin a numeric argument with the given digit.", Type: "widget", Target: "digit-argument", IsCustom: false},
		{Display: "Alt+2", Description: "Digit argument", FullDescription: "Begin a numeric argument with the given digit.", Type: "widget", Target: "digit-argument", IsCustom: false},
		{Display: "Alt+3", Description: "Digit argument", FullDescription: "Begin a numeric argument with the given digit.", Type: "widget", Target: "digit-argument", IsCustom: false},
		{Display: "Alt+4", Description: "Digit argument", FullDescription: "Begin a numeric argument with the given digit.", Type: "widget", Target: "digit-argument", IsCustom: false},
		{Display: "Alt+5", Description: "Digit argument", FullDescription: "Begin a numeric argument with the given digit.", Type: "widget", Target: "digit-argument", IsCustom: false},
		{Display: "Alt+6", Description: "Digit argument", FullDescription: "Begin a numeric argument with the given digit.", Type: "widget", Target: "digit-argument", IsCustom: false},
		{Display: "Alt+7", Description: "Digit argument", FullDescription: "Begin a numeric argument with the given digit.", Type: "widget", Target: "digit-argument", IsCustom: false},
		{Display: "Alt+8", Description: "Digit argument", FullDescription: "Begin a numeric argument with the given digit.", Type: "widget", Target: "digit-argument", IsCustom: false},
		{Display: "Alt+9", Description: "Digit argument", FullDescription: "Begin a numeric argument with the given digit.", Type: "widget", Target: "digit-argument", IsCustom: false},
		{Display: "Alt+<", Description: "Beginning of buffer or history", FullDescription: "Move to the beginning of the buffer or history.", Type: "widget", Target: "beginning-of-buffer-or-history", IsCustom: false},
		{Display: "Alt+>", Description: "End of buffer or history", FullDescription: "Move to the end of the buffer or history.", Type: "widget", Target: "end-of-buffer-or-history", IsCustom: false},
		{Display: "Alt+?", Description: "Which command", FullDescription: "Display information about the command.", Type: "widget", Target: "which-command", IsCustom: false},
		{Display: "Alt+A", Description: "Accept and hold", FullDescription: "Accept the current line and push it onto the history stack.", Type: "widget", Target: "accept-and-hold", IsCustom: false},
		{Display: "Alt+B", Description: "Backward word", FullDescription: "Move to the beginning of the previous word.", Type: "widget", Target: "backward-word", IsCustom: false},
		{Display: "Alt+C", Description: "Capitalize word", FullDescription: "Capitalize the current word.", Type: "widget", Target: "capitalize-word", IsCustom: false},
		{Display: "Alt+D", Description: "Kill word", FullDescription: "Kill the current word.", Type: "widget", Target: "kill-word", IsCustom: false},
		{Display: "Alt+F", Description: "Forward word", FullDescription: "Move to the beginning of the next word.", Type: "widget", Target: "forward-word", IsCustom: false},
		{Display: "Alt+G", Description: "Get line", FullDescription: "Get a line from the history.", Type: "widget", Target: "get-line", IsCustom: false},
		{Display: "Alt+H", Description: "Run help", FullDescription: "Run help on the current command.", Type: "widget", Target: "run-help", IsCustom: false},
		{Display: "Alt+L", Description: "Down case word", FullDescription: "Convert the current word to lowercase.", Type: "widget", Target: "down-case-word", IsCustom: false},
		{Display: "Alt+N", Description: "History search forward", FullDescription: "Search forward in history for a line beginning with the current word.", Type: "widget", Target: "history-search-forward", IsCustom: false},
		{Display: "Alt+P", Description: "History search backward", FullDescription: "Search backward in history for a line beginning with the current word.", Type: "widget", Target: "history-search-backward", IsCustom: false},
		{Display: "Alt+Q", Description: "Push line", FullDescription: "Push the current line onto the buffer stack and clear the line.", Type: "widget", Target: "push-line", IsCustom: false},
		{Display: "Alt+S", Description: "Spell word", FullDescription: "Attempt spelling correction on the current word.", Type: "widget", Target: "spell-word", IsCustom: false},
		{Display: "Alt+T", Description: "Transpose words", FullDescription: "Exchange the current word with the one before it.", Type: "widget", Target: "transpose-words", IsCustom: false},
		{Display: "Alt+U", Description: "Up case word", FullDescription: "Convert the current word to uppercase.", Type: "widget", Target: "up-case-word", IsCustom: false},
		{Display: "Alt+W", Description: "Copy region as kill", FullDescription: "Copy the region between the cursor and the mark to the kill ring.", Type: "widget", Target: "copy-region-as-kill", IsCustom: false},
		{Display: "Alt+a", Description: "Accept and hold", FullDescription: "Accept the current line and push it onto the history stack.", Type: "widget", Target: "accept-and-hold", IsCustom: false},
		{Display: "Alt+b", Description: "Backward word", FullDescription: "Move to the beginning of the previous word.", Type: "widget", Target: "backward-word", IsCustom: false},
		{Display: "Alt+c", Description: "Capitalize word", FullDescription: "Capitalize the current word.", Type: "widget", Target: "capitalize-word", IsCustom: false},
		{Display: "Alt+d", Description: "Kill word", FullDescription: "Kill the current word.", Type: "widget", Target: "kill-word", IsCustom: false},
		{Display: "Alt+f", Description: "Forward word", FullDescription: "Move to the beginning of the next word.", Type: "widget", Target: "forward-word", IsCustom: false},
		{Display: "Alt+g", Description: "Get line", FullDescription: "Get a line from the history.", Type: "widget", Target: "get-line", IsCustom: false},
		{Display: "Alt+h", Description: "Run help", FullDescription: "Run help on the current command.", Type: "widget", Target: "run-help", IsCustom: false},
		{Display: "Alt+l", Description: "Down case word", FullDescription: "Convert the current word to lowercase.", Type: "widget", Target: "down-case-word", IsCustom: false},
		{Display: "Alt+n", Description: "History search forward", FullDescription: "Search forward in history for a line beginning with the current word.", Type: "widget", Target: "history-search-forward", IsCustom: false},
		{Display: "Alt+p", Description: "History search backward", FullDescription: "Search backward in history for a line beginning with the current word.", Type: "widget", Target: "history-search-backward", IsCustom: false},
		{Display: "Alt+q", Description: "Push line", FullDescription: "Push the current line onto the buffer stack and clear the line.", Type: "widget", Target: "push-line", IsCustom: false},
		{Display: "Alt+s", Description: "Spell word", FullDescription: "Attempt spelling correction on the current word.", Type: "widget", Target: "spell-word", IsCustom: false},
		{Display: "Alt+t", Description: "Transpose words", FullDescription: "Exchange the current word with the one before it.", Type: "widget", Target: "transpose-words", IsCustom: false},
		{Display: "Alt+u", Description: "Up case word", FullDescription: "Convert the current word to uppercase.", Type: "widget", Target: "up-case-word", IsCustom: false},
		{Display: "Alt+w", Description: "Copy region as kill", FullDescription: "Copy the region between the cursor and the mark to the kill ring.", Type: "widget", Target: "copy-region-as-kill", IsCustom: false},
		{Display: "Alt+x", Description: "Execute named cmd", FullDescription: "Execute a named command.", Type: "widget", Target: "execute-named-cmd", IsCustom: false},
		{Display: "Alt+y", Description: "Yank pop", FullDescription: "Replace the just-yanked text with the next item from the kill ring.", Type: "widget", Target: "yank-pop", IsCustom: false},
		{Display: "Alt+z", Description: "Execute last named cmd", FullDescription: "Execute the last named command.", Type: "widget", Target: "execute-last-named-cmd", IsCustom: false},
		{Display: "Alt+|", Description: "Vi goto column", FullDescription: "Go to the specified column.", Type: "widget", Target: "vi-goto-column", IsCustom: false},
		{Display: "Alt+~", Description: "Bash complete word", FullDescription: "Complete the current word using bash-style completion.", Type: "widget", Target: "_bash_complete-word", IsCustom: false},
		{Display: "Alt+Backspace", Description: "Backward kill word", FullDescription: "Kill the word behind the cursor.", Type: "widget", Target: "backward-kill-word", IsCustom: false},
		{Display: "Ctrl+_", Description: "Shortcutter widget", FullDescription: "Open the shortcutter widget.", Type: "widget", Target: "shortcutter_widget", IsCustom: false},
		{Display: "Backspace", Description: "Backward delete char", FullDescription: "Delete the character behind the cursor.", Type: "widget", Target: "backward-delete-char", IsCustom: false},
		{Display: "↑", Description: "Up line or history", FullDescription: "Move up a line in the buffer, or if already at the top line, move to the previous event in the history list.", Type: "widget", Target: "up-line-or-history", IsCustom: false},
		{Display: "↓", Description: "Down line or history", FullDescription: "Move down a line in the buffer, or if already at the bottom line, move to the next event in the history list.", Type: "widget", Target: "down-line-or-history", IsCustom: false},
		{Display: "→", Description: "Forward char", FullDescription: "Move forward one character.", Type: "widget", Target: "forward-char", IsCustom: false},
		{Display: "←", Description: "Backward char", FullDescription: "Move backward one character.", Type: "widget", Target: "backward-char", IsCustom: false},
		{Display: "End", Description: "End of line", FullDescription: "Move to the end of the line.", Type: "widget", Target: "end-of-line", IsCustom: false},
		{Display: "Home", Description: "Beginning of line", FullDescription: "Move to the beginning of the line.", Type: "widget", Target: "beginning-of-line", IsCustom: false},
		{Display: "Delete", Description: "Delete char", FullDescription: "Delete the character under the cursor.", Type: "widget", Target: "delete-char", IsCustom: false},
		{Display: "Insert", Description: "Overwrite mode", FullDescription: "Toggle overwrite mode.", Type: "widget", Target: "overwrite-mode", IsCustom: false},
		{Display: "Ctrl+Shift+5", Description: "Delete char", FullDescription: "Delete the character under the cursor.", Type: "widget", Target: "delete-char", IsCustom: false},
	}
}

func getBashBuiltinShortcuts() []Shortcut {
	return []Shortcut{
		{Display: "Ctrl+A", Description: "Beginning of line", FullDescription: "Beginning of line", Type: "sequence", Target: "C-a", IsCustom: false},
		{Display: "Ctrl+E", Description: "End of line", FullDescription: "End of line", Type: "sequence", Target: "C-e", IsCustom: false},
		{Display: "Ctrl+F", Description: "Forward one character", FullDescription: "Forward one character", Type: "sequence", Target: "C-f", IsCustom: false},
		{Display: "Ctrl+B", Description: "Back one character", Type: "sequence", Target: "C-b", IsCustom: false},
		{Display: "Alt+F", Description: "Forward one word", Type: "sequence", Target: "M-f", IsCustom: false},
		{Display: "Alt+B", Description: "Back one word", Type: "sequence", Target: "M-b", IsCustom: false},
		{Display: "Ctrl+T", Description: "Transpose characters", Type: "sequence", Target: "C-t", IsCustom: false},
		{Display: "Alt+T", Description: "Transpose words", Type: "sequence", Target: "M-t", IsCustom: false},
		{Display: "Ctrl+U", Description: "Kill line backward", Type: "sequence", Target: "C-u", IsCustom: false},
		{Display: "Ctrl+K", Description: "Kill line forward", Type: "sequence", Target: "C-k", IsCustom: false},
		{Display: "Ctrl+H", Description: "Delete character backward", Type: "sequence", Target: "C-h", IsCustom: false},
		{Display: "Ctrl+W", Description: "Kill word backward", Type: "sequence", Target: "C-w", IsCustom: false},
		{Display: "Ctrl+Y", Description: "Yank", Type: "sequence", Target: "C-y", IsCustom: false},
		{Display: "Ctrl+L", Description: "Clear screen", Type: "sequence", Target: "C-l", IsCustom: false},
		{Display: "Ctrl+R", Description: "Reverse search", Type: "sequence", Target: "C-r", IsCustom: false},
		{Display: "Ctrl+S", Description: "Forward search", Type: "sequence", Target: "C-s", IsCustom: false},
		{Display: "Ctrl+P", Description: "Previous line", Type: "sequence", Target: "C-p", IsCustom: false},
		{Display: "Ctrl+N", Description: "Next line", Type: "sequence", Target: "C-n", IsCustom: false},
		{Display: "Ctrl+D", Description: "Delete character or EOF", Type: "sequence", Target: "C-d", IsCustom: false},
		{Display: "Ctrl+C", Description: "Interrupt", Type: "sequence", Target: "C-c", IsCustom: false},
		{Display: "Ctrl+Z", Description: "Suspend", Type: "sequence", Target: "C-z", IsCustom: false},
		{Display: "Tab", Description: "Complete", Type: "sequence", Target: "Tab", IsCustom: false},
		{Display: "Enter", Description: "Execute command", Type: "sequence", Target: "Enter", IsCustom: false},
		{Display: "↑", Description: "Previous command", Type: "sequence", Target: "Up", IsCustom: false},
		{Display: "↓", Description: "Next command", Type: "sequence", Target: "Down", IsCustom: false},
		{Display: "←", Description: "Move cursor left", Type: "sequence", Target: "Left", IsCustom: false},
		{Display: "→", Description: "Move cursor right", Type: "sequence", Target: "Right", IsCustom: false},
		{Display: "Home", Description: "Beginning of line", Type: "sequence", Target: "Home", IsCustom: false},
		{Display: "End", Description: "End of line", Type: "sequence", Target: "End", IsCustom: false},
		{Display: "Delete", Description: "Delete character", Type: "sequence", Target: "Delete", IsCustom: false},
		{Display: "Backspace", Description: "Delete character backward", Type: "sequence", Target: "Backspace", IsCustom: false},
		{Display: "Page Up", Description: "Page up", Type: "sequence", Target: "Page_Up", IsCustom: false},
		{Display: "Page Down", Description: "Page down", Type: "sequence", Target: "Page_Down", IsCustom: false},
	}
}

func getGenericBuiltinShortcuts() []Shortcut {
	return []Shortcut{
		{Display: "Ctrl+A", Description: "Beginning of line", Type: "sequence", Target: "C-a", IsCustom: false},
		{Display: "Ctrl+E", Description: "End of line", Type: "sequence", Target: "C-e", IsCustom: false},
		{Display: "Ctrl+F", Description: "Forward one character", Type: "sequence", Target: "C-f", IsCustom: false},
		{Display: "Ctrl+B", Description: "Back one character", Type: "sequence", Target: "C-b", IsCustom: false},
		{Display: "Ctrl+U", Description: "Kill line backward", Type: "sequence", Target: "C-u", IsCustom: false},
		{Display: "Ctrl+K", Description: "Kill line forward", Type: "sequence", Target: "C-k", IsCustom: false},
		{Display: "Ctrl+L", Description: "Clear screen", Type: "sequence", Target: "C-l", IsCustom: false},
		{Display: "Ctrl+C", Description: "Interrupt", Type: "sequence", Target: "C-c", IsCustom: false},
		{Display: "Ctrl+Z", Description: "Suspend", Type: "sequence", Target: "C-z", IsCustom: false},
		{Display: "Tab", Description: "Complete", Type: "sequence", Target: "Tab", IsCustom: false},
		{Display: "Enter", Description: "Execute command", Type: "sequence", Target: "Enter", IsCustom: false},
		{Display: "↑", Description: "Previous command", Type: "sequence", Target: "Up", IsCustom: false},
		{Display: "↓", Description: "Next command", Type: "sequence", Target: "Down", IsCustom: false},
		{Display: "←", Description: "Move cursor left", Type: "sequence", Target: "Left", IsCustom: false},
		{Display: "→", Description: "Move cursor right", Type: "sequence", Target: "Right", IsCustom: false},
		{Display: "Home", Description: "Beginning of line", Type: "sequence", Target: "Home", IsCustom: false},
		{Display: "End", Description: "End of line", Type: "sequence", Target: "End", IsCustom: false},
		{Display: "Delete", Description: "Delete character", Type: "sequence", Target: "Delete", IsCustom: false},
		{Display: "Backspace", Description: "Delete character backward", Type: "sequence", Target: "Backspace", IsCustom: false},
	}
}

func normalizeKey(key string) string {
	key = strings.TrimSpace(key)
	if matched, _ := regexp.MatchString(`^\^[A-Za-z@_\[\]\\]$`, key); matched {
		char := strings.ToUpper(string(key[1]))
		switch char {
		case "[":
			return "Esc"
		case "I":
			return "Tab"
		case "M":
			return "Enter"
		case "H":
			return "Backspace"
		case "@":
			return "Ctrl+@"
		case "_":
			return "Ctrl+_"
		case "\\":
			return "Ctrl+\\"
		case "]":
			return "Ctrl+]"
		default:
			return "Ctrl+" + char
		}
	}
	if matched, _ := regexp.MatchString(`^[Cc]-[a-zA-Z@_\[\]\\]$`, key); matched {
		char := strings.ToUpper(string(key[2]))
		return "Ctrl+" + char
	}
	if matched, _ := regexp.MatchString(`^[Mm]-[a-zA-Z]$`, key); matched {
		char := strings.ToUpper(string(key[2]))
		return "Alt+" + char
	}
	key = regexp.MustCompile(`(?i)ctrl\+`).ReplaceAllString(key, "Ctrl+")
	key = regexp.MustCompile(`(?i)alt\+`).ReplaceAllString(key, "Alt+")
	key = regexp.MustCompile(`(?i)shift\+`).ReplaceAllString(key, "Shift+")
	key = regexp.MustCompile(`(?i)meta\+`).ReplaceAllString(key, "Alt+")

	parts := strings.Split(key, "+")
	if len(parts) > 1 {
		lastPart := parts[len(parts)-1]
		if len(lastPart) == 1 && lastPart >= "a" && lastPart <= "z" {
			parts[len(parts)-1] = strings.ToUpper(lastPart)
		} else if strings.ToLower(lastPart) == "tab" {
			parts[len(parts)-1] = "Tab"
		}
		key = strings.Join(parts, "+")
	}

	return key
}

func loadConfig() (*Config, error) {
	homeDir, err := os.UserHomeDir()
	if err != nil {
		return &Config{Shortcuts: make(map[string]interface{})}, nil
	}

	configPath := filepath.Join(homeDir, ".config", "shortcutter", "config.toml")

	if _, err := os.Stat(configPath); os.IsNotExist(err) {
		return &Config{Shortcuts: make(map[string]interface{})}, nil
	}

	var config Config
	if _, err := toml.DecodeFile(configPath, &config); err != nil {
		return nil, fmt.Errorf("failed to parse config file: %w", err)
	}

	if config.Shortcuts == nil {
		config.Shortcuts = make(map[string]interface{})
	}

	return &config, nil
}

func mergeShortcuts(builtins []Shortcut, config *Config) []Shortcut {
	shortcutMap := make(map[string]Shortcut)
	
	// Index built-ins by their display name
	for _, shortcut := range builtins {
		normalizedKey := normalizeKey(shortcut.Display)
		shortcutMap[normalizedKey] = shortcut
	}

	for configKey, configValue := range config.Shortcuts {
		normalizedKey := normalizeKey(configKey)

		switch v := configValue.(type) {
		case bool:
			// Disable shortcut
			if !v {
				delete(shortcutMap, normalizedKey)
			}
		case string:
			// Simple override - just change description, inherit everything else from built-in
			if v != "" {
				if existing, exists := shortcutMap[normalizedKey]; exists {
					// Override description but keep other fields
					existing.Description = v
					existing.IsCustom = true
					shortcutMap[normalizedKey] = existing
				} else {
					// New shortcut with just description - assume it's a command
					shortcut := Shortcut{
						Display:     normalizedKey,
						Description: v,
						Type:        "command",
						Target:      v, // Use description as command for simple cases
						IsCustom:    true,
					}
					shortcutMap[normalizedKey] = shortcut
				}
			}
		case map[string]interface{}:
			// Full object configuration
			shortcut := Shortcut{
				Display:  normalizedKey,
				IsCustom: true,
			}
			
			// Start with existing built-in if it exists
			if existing, exists := shortcutMap[normalizedKey]; exists {
				shortcut = existing
				shortcut.IsCustom = true
			}
			
			// Override with config values
			if display, ok := v["display"].(string); ok {
				shortcut.Display = display
			}
			if description, ok := v["description"].(string); ok {
				shortcut.Description = description
			}
			if shortcutType, ok := v["type"].(string); ok {
				shortcut.Type = shortcutType
			}
			if target, ok := v["target"].(string); ok {
				shortcut.Target = target
			}
			
			shortcutMap[normalizedKey] = shortcut
		}
	}

	result := make([]Shortcut, 0, len(shortcutMap))
	for _, shortcut := range shortcutMap {
		result = append(result, shortcut)
	}

	sort.Slice(result, func(i, j int) bool {
		return result[i].Display < result[j].Display
	})

	return result
}

func DetectShortcuts() ([]Shortcut, error) {
	return LoadShortcuts()
}

func NormalizeKeyForTesting(key string) string {
	return normalizeKey(key)
}

var getShellEnv = func() string {
	return os.Getenv("SHELL")
}
