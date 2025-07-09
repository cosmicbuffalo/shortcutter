#!/bin/zsh

# Shortcutter zsh integration
# This file should be sourced in your .zshrc

shortcutter_widget() {
    # Run shortcutter directly - let it take over the terminal
    shortcutter
    
    # Reset prompt
    zle reset-prompt
}

# Create the widget
zle -N shortcutter_widget

# Bind Ctrl+/ to the widget
# Note: Ctrl+/ is represented as "^_" in zsh
bindkey "^_" shortcutter_widget