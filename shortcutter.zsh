#!/bin/zsh

# Shortcutter zsh integration
# This file should be sourced in your .zshrc

shortcutter_widget() {
    # Save the current command line state
    local saved_buffer="$BUFFER"
    local saved_cursor="$CURSOR"
    
    # Move to next line (fzf pattern - don't clear current line)
    echo
    
    # Run shortcutter and capture output
    local result=$(shortcutter 2>/dev/null)
    
    # Parse the result format: key:type:target
    if [[ -n "$result" ]]; then
        local key=$(echo "$result" | cut -d: -f1)
        local type=$(echo "$result" | cut -d: -f2)
        local target=$(echo "$result" | cut -d: -f3-)
        
        # Determine action based on key press and context
        local should_populate=false
        
        if [[ "$key" == "tab" ]]; then
            # Tab always populates (for command types only)
            should_populate=true
        elif [[ "$key" == "enter" && "$type" == "command" && -n "$saved_buffer" ]]; then
            # Enter with existing buffer content should populate
            should_populate=true
        fi
        
        # Execute based on type and action
        if [[ "$type" == "widget" ]]; then
            # Restore buffer first, then execute widget
            BUFFER="$saved_buffer"
            CURSOR="$saved_cursor"
            
            # Map widget names back to key sequences and use zle -U to queue them
            # This preserves proper zle context for all widgets
            case "$target" in
                "beginning-of-line") zle -U $'\C-a' ;;  # Ctrl+A
                "end-of-line") zle -U $'\C-e' ;;        # Ctrl+E
                "forward-char") zle -U $'\C-f' ;;       # Ctrl+F
                "backward-char") zle -U $'\C-b' ;;      # Ctrl+B
                "forward-word") zle -U $'\ef' ;;        # Alt+F (ESC+f)
                "backward-word") zle -U $'\eb' ;;       # Alt+B (ESC+b)
                "transpose-chars") zle -U $'\C-t' ;;    # Ctrl+T
                "transpose-words") zle -U $'\et' ;;     # Alt+T (ESC+t)
                "backward-kill-line") zle -U $'\C-u' ;; # Ctrl+U
                "kill-line") zle -U $'\C-k' ;;          # Ctrl+K
                "backward-delete-char") zle -U $'\C-h' ;; # Ctrl+H
                "backward-kill-word") zle -U $'\C-w' ;; # Ctrl+W
                "set-mark-command") zle -U $'\C-@' ;;   # Ctrl+@
                "yank") zle -U $'\C-y' ;;               # Ctrl+Y
                "quoted-insert") zle -U $'\C-v' ;;      # Ctrl+V
                "push-line") zle -U $'\C-q' ;;          # Ctrl+Q
                "undo") zle -U $'\C-_' ;;               # Ctrl+_
                "up-line-or-history") zle -U $'\C-p' ;; # Ctrl+P
                "down-line-or-history") zle -U $'\C-n' ;; # Ctrl+N
                "history-incremental-search-backward") zle -U $'\C-r' ;; # Ctrl+R
                "history-search-backward") zle -U $'\ep' ;; # Alt+P
                "insert-last-word") zle -U $'\e.' ;;    # Alt+.
                "clear-screen") zle -U $'\C-l' ;;       # Ctrl+L
                "accept-line-and-down-history") zle -U $'\C-o' ;; # Ctrl+O
                "expand-or-complete") zle -U $'\t' ;;   # Tab
                "accept-line") zle -U $'\n' ;;          # Enter
                "delete-char-or-list") zle -U $'\C-d' ;; # Ctrl+D
                "send-break") zle -U $'\C-g' ;;         # Ctrl+G
                "edit-command-line") zle -U $'\C-x\C-e' ;; # Ctrl+X Ctrl+E
                *) 
                    # Fallback to direct widget execution for unmapped widgets
                    zle "$target" 
                    ;;
            esac
        elif [[ "$type" == "sequence" ]]; then
            # Sequences always execute (send raw keystrokes)
            # Restore buffer first
            BUFFER="$saved_buffer"
            CURSOR="$saved_cursor"
            # Convert target format (e.g., "C-c" to actual key sequence)
            case "$target" in
                "C-c") printf "\003" ;;
                "C-z") printf "\032" ;;
                "C-s") printf "\023" ;;
                *) ;;
            esac
        elif [[ "$type" == "command" ]]; then
            if [[ "$should_populate" == "true" ]]; then
                # Populate command into buffer
                if [[ -n "$saved_buffer" ]]; then
                    # Add space if buffer doesn't end with one
                    if [[ "$saved_buffer" != *" " ]]; then
                        BUFFER="$saved_buffer $target"
                    else
                        BUFFER="$saved_buffer$target"
                    fi
                    CURSOR=${#BUFFER}
                else
                    # Empty buffer, just set the command
                    BUFFER="$target"
                    CURSOR=${#BUFFER}
                fi
            else
                # Execute command immediately, then restore state
                eval "$target"
                BUFFER="$saved_buffer"
                CURSOR="$saved_cursor"
            fi
        fi
    else
        # No selection made, restore original state
        BUFFER="$saved_buffer"
        CURSOR="$saved_cursor"
    fi
    
    # Always reset the prompt at the end
    zle reset-prompt
}

# Create the widget
zle -N shortcutter_widget

# Bind Ctrl+/ to the widget
# Note: Ctrl+/ is represented as "^_" in zsh
bindkey "^_" shortcutter_widget
