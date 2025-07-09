#!/bin/zsh

# Shortcutter zsh integration
# This file should be sourced in your .zshrc

shortcutter_widget() {
    # Clear the current line
    zle kill-whole-line
    
    # Run shortcutter and capture the output
    local result
    result=$(shortcutter 2>/dev/null)
    
    # Check if we got a result
    if [[ -n "$result" ]]; then
        # Determine the action based on the result format
        # The Go program outputs the action (execute/populate/info)
        local action_type="${result%%:*}"
        local command="${result#*:}"
        
        case "$action_type" in
            "execute")
                # Execute the command directly
                eval "$command"
                ;;
            "populate")
                # Put the command in the buffer for editing
                BUFFER="$command"
                CURSOR=${#BUFFER}
                ;;
            "info")
                # Just show info, don't execute
                echo "\n$command"
                ;;
            *)
                # Fallback - treat as populate
                BUFFER="$result"
                CURSOR=${#BUFFER}
                ;;
        esac
    fi
    
    # Redraw the prompt
    zle reset-prompt
}

# Create the widget
zle -N shortcutter_widget

# Bind Ctrl+/ to the widget
# Note: Ctrl+/ is represented as "^_" in zsh
bindkey "^_" shortcutter_widget

# Alternative binding for terminals that don't support Ctrl+/
# You can use Ctrl+X Ctrl+S instead
bindkey "^X^S" shortcutter_widget