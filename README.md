# Shortcutter 🚀

A terminal shortcut reference tool for zsh that provides a fuzzy-searchable interface for your shell commands, aliases, functions, and key bindings.

## Features

- **Auto-detection**: Automatically detects your zsh aliases, functions, and key bindings
- **Fuzzy search**: Search through shortcuts by command name or description
- **Built-in commands**: Includes common shell commands and utilities
- **Interactive UI**: Clean, responsive interface with keyboard navigation
- **Smart execution**: Execute commands directly or pre-populate them for editing
- **Global access**: Trigger from anywhere in your shell with Ctrl+/

## Installation

1. Clone or download this repository
2. Navigate to the shortcutter directory
3. Run the installation script:

```bash
./install.sh
```

4. Restart your shell or run:

```bash
source ~/.zshrc
```

## Usage

### Opening Shortcutter

Press **Ctrl+/** anywhere in your zsh shell to open the shortcut reference.

Alternative binding: **Ctrl+X Ctrl+S**

### Navigation

- **Type** to search through shortcuts
- **↑/↓** or **j/k** to navigate through results
- **Enter** to select a shortcut
- **Esc** to quit

### Shortcut Types

- **Aliases**: Your custom shell aliases
- **Functions**: User-defined zsh functions
- **Key bindings**: Terminal key combinations (Ctrl+A, Ctrl+E, etc.)
- **Built-ins**: Common shell commands and utilities

### Actions

- **Execute**: Run the command immediately
- **Populate**: Place the command in your shell prompt for editing
- **Info**: Display information about key bindings

## Requirements

- Go 1.19 or later
- zsh shell
- Terminal with 256 color support (recommended)

## File Structure

```
shortcutter/
├── main.go              # Entry point
├── internal/
│   ├── shortcuts.go     # Shortcut detection logic
│   └── ui.go           # Fuzzy search interface
├── install.sh          # Installation script
├── shortcutter.zsh     # zsh integration
└── README.md           # This file
```

## Uninstallation

To remove shortcutter:

1. Remove the binary:
   ```bash
   rm ~/.local/bin/shortcutter
   ```

2. Remove the zsh integration from your `.zshrc`:
   ```bash
   # Remove lines between "# Shortcutter integration" and "# End shortcutter integration"
   ```

## Contributing

Feel free to submit issues and enhancement requests!

## License

This project is open source and available under the [MIT License](LICENSE).