package main

import (
	"fmt"
	"os"
	"shortcutter/internal"
)

func main() {
	shortcuts, styles, err := internal.LoadShortcutsAndTheme()
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error loading shortcuts and theme: %v\n", err)
		os.Exit(1)
	}

	selected, selectedKey, err := internal.ShowUI(shortcuts, styles)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error showing UI: %v\n", err)
		os.Exit(1)
	}

	if selected != nil {
		fmt.Printf("%s:%s:%s\n", selectedKey, selected.Type, selected.Target)
	}
}
