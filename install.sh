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

# Check if we're on macOS or Linux
detect_os() {
    if [[ "$OSTYPE" == "darwin"* ]]; then
        echo "macos"
    elif [[ "$OSTYPE" == "linux-gnu"* ]]; then
        echo "linux"
    else
        echo "unknown"
    fi
}

# Check if Go is installed
check_go() {
    if ! command -v go &> /dev/null; then
        print_error "Go is not installed. Please install Go first."
        print_status "Visit https://golang.org/dl/ to download Go"
        exit 1
    fi
    print_success "Go is installed: $(go version)"
}

# Check if zsh is the current shell
check_zsh() {
    if [[ "$SHELL" != *"zsh"* ]]; then
        print_warning "Current shell is not zsh. This tool is designed for zsh."
        print_status "Current shell: $SHELL"
        read -p "Continue anyway? (y/N) " -n 1 -r
        echo
        if [[ ! $REPLY =~ ^[Yy]$ ]]; then
            exit 1
        fi
    fi
    print_success "zsh detected"
}

# Build the binary
build_binary() {
    print_status "Building shortcutter binary..."
    
    if ! go build -o shortcutter; then
        print_error "Failed to build shortcutter"
        exit 1
    fi
    
    print_success "Binary built successfully"
}

# Install binary to user's local bin
install_binary() {
    local bin_dir="$HOME/.local/bin"
    
    print_status "Installing binary to $bin_dir..."
    
    # Create bin directory if it doesn't exist
    mkdir -p "$bin_dir"
    
    # Copy binary
    cp shortcutter "$bin_dir/"
    chmod +x "$bin_dir/shortcutter"
    
    # Check if ~/.local/bin is in PATH
    if [[ ":$PATH:" != *":$bin_dir:"* ]]; then
        print_warning "~/.local/bin is not in your PATH"
        print_status "Adding export to your .zshrc"
        echo 'export PATH="$HOME/.local/bin:$PATH"' >> "$HOME/.zshrc"
        print_status "Please restart your shell or run: source ~/.zshrc"
    fi
    
    print_success "Binary installed to $bin_dir/shortcutter"
}

# Add zsh integration
install_zsh_integration() {
    print_status "Installing zsh integration..."
    
    local zshrc="$HOME/.zshrc"
    local integration_marker="# Shortcutter integration"
    
    # Check if already installed
    if grep -q "$integration_marker" "$zshrc" 2>/dev/null; then
        print_warning "Shortcutter integration already exists in .zshrc"
        read -p "Replace existing integration? (y/N) " -n 1 -r
        echo
        if [[ $REPLY =~ ^[Yy]$ ]]; then
            # Remove existing integration
            sed -i.bak "/$integration_marker/,/# End shortcutter integration/d" "$zshrc"
            print_status "Removed existing integration"
        else
            print_status "Skipping zsh integration"
            return
        fi
    fi
    
    # Add integration
    echo "" >> "$zshrc"
    echo "$integration_marker" >> "$zshrc"
    cat shortcutter.zsh >> "$zshrc"
    echo "# End shortcutter integration" >> "$zshrc"
    
    print_success "zsh integration added to .zshrc"
    print_status "Key bindings:"
    print_status "  Ctrl+/ - Open shortcutter"
    print_status "  Ctrl+X Ctrl+S - Alternative binding"
}

# Main installation function
main() {
    print_status "🚀 Installing Shortcutter..."
    echo
    
    # Check system requirements
    check_go
    check_zsh
    
    # Build and install
    build_binary
    install_binary
    install_zsh_integration
    
    echo
    print_success "🎉 Shortcutter installed successfully!"
    print_status "Please restart your shell or run: source ~/.zshrc"
    print_status "Then press Ctrl+/ anywhere in your shell to open shortcutter"
    echo
}

# Run installation
main "$@"