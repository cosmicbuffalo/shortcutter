package main

import (
	"fmt"
	"os"

	"shortcutter/internal"
)

func main() {
	shortcuts, err := internal.DetectShortcuts()
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error detecting shortcuts: %v\n", err)
		os.Exit(1)
	}

	selected, err := internal.ShowUI(shortcuts)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error showing UI: %v\n", err)
		os.Exit(1)
	}

	if selected != nil {
		fmt.Printf("%s:%s\n", selected.Action, selected.Command)
	}
}