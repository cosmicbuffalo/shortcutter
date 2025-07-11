#!/bin/bash

set -e

# Colors for output
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
NC='\033[0m' # No Color

# Helper functions
print_status() {
    echo -e "${BLUE}[INFO]${NC} $1"
}

print_success() {
    echo -e "${GREEN}[SUCCESS]${NC} $1"
}

print_warning() {
    echo -e "${YELLOW}[WARNING]${NC} $1"
}

print_error() {
    echo -e "${RED}[ERROR]${NC} $1"
}

# Remove binary from user's local bin
remove_binary() {
    local bin_dir="$HOME/.local/bin"
    local binary_path="$bin_dir/shortcutter"
    
    if [[ -f "$binary_path" ]]; then
        print_status "Removing binary from $bin_dir..."
        rm -f "$binary_path"
        print_success "Binary removed from $bin_dir/shortcutter"
    else
        print_status "Binary not found at $binary_path"
    fi
}

# Remove zsh integration
remove_zsh_integration() {
    print_status "Removing zsh integration..."
    
    local zshrc="$HOME/.zshrc"
    local integration_marker="# Shortcutter integration"
    
    if [[ -f "$zshrc" ]] && grep -q "$integration_marker" "$zshrc"; then
        print_status "Found shortcutter integration in .zshrc"
        
        # Create backup
        cp "$zshrc" "$zshrc.shortcutter-backup-$(date +%Y%m%d-%H%M%S)"
        print_status "Created backup: $zshrc.shortcutter-backup-$(date +%Y%m%d-%H%M%S)"
        
        # Remove integration between markers
        sed -i.tmp "/$integration_marker/,/# End shortcutter integration/d" "$zshrc"
        rm -f "$zshrc.tmp"
        
        print_success "Shortcutter integration removed from .zshrc"
    else
        print_status "No shortcutter integration found in .zshrc"
    fi
}

# Clear cache and config
clear_cache_and_config() {
    print_status "Clearing shortcutter cache and configuration..."
    
    local locations=(
        "$HOME/.config/shortcutter"
        "$HOME/.cache/shortcutter"
    )
    
    if [[ -n "$XDG_CACHE_HOME" ]]; then
        locations+=("$XDG_CACHE_HOME/shortcutter")
    fi
    
    if [[ -n "$XDG_CONFIG_HOME" ]]; then
        locations+=("$XDG_CONFIG_HOME/shortcutter")
    fi
    
    local removed=false
    for location in "${locations[@]}"; do
        if [[ -d "$location" ]]; then
            print_status "Removing: $location"
            rm -rf "$location"
            removed=true
        fi
    done
    
    if [[ "$removed" == "true" ]]; then
        print_success "Cache and configuration cleared"
    else
        print_status "No cache or configuration found"
    fi
}

# Unload any active zle widgets
unload_zle_widgets() {
    print_status "Attempting to unload shortcutter zle widgets..."
    
    # If running in zsh, try to unload the widget
    if [[ -n "$ZSH_VERSION" ]]; then
        # Check if widget exists and remove it
        if zle -la | grep -q "shortcutter_widget"; then
            print_status "Removing shortcutter_widget from current session..."
            # Unbind the key first
            bindkey -r "^_" 2>/dev/null || true
            # Delete the widget
            zle -D shortcutter_widget 2>/dev/null || true
            print_success "Widget unloaded from current session"
        else
            print_status "No active shortcutter widget found in current session"
        fi
    else
        print_warning "Not running in zsh - widgets will be removed on next shell restart"
    fi
}

# Main uninstallation function
main() {
    print_status "🗑️  Uninstalling Shortcutter..."
    echo
    
    unload_zle_widgets
    remove_binary
    remove_zsh_integration
    clear_cache_and_config
    
    echo
    print_success "🎉 Shortcutter uninstalled successfully!"
    print_warning "Please restart your shell or run: source ~/.zshrc"
    print_status "This will ensure all zsh widgets are properly removed"
    echo
}

# Run uninstallation
main "$@"