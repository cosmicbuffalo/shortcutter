#!/bin/zsh

# Shortcutter zsh integration
# This file should be sourced in your .zshrc

shortcutter_widget() {
    # Save the current command line state
    local saved_buffer="$BUFFER"
    local saved_cursor="$CURSOR"
    
    # Move to next line (fzf pattern - don't clear current line)
    echo
    
    # Run shortcutter directly - let it take over the terminal
    shortcutter
    
    # Restore the original command line
    BUFFER="$saved_buffer"
    CURSOR="$saved_cursor"
    
    # Reset prompt
    zle reset-prompt
}

# Create the widget
zle -N shortcutter_widget

# Bind Ctrl+/ to the widget
# Note: Ctrl+/ is represented as "^_" in zsh
bindkey "^_" shortcutter_widget