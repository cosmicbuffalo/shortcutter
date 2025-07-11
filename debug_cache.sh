#!/bin/bash

# Debug script to check cache status
echo "=== Shortcutter Cache Debug ==="

# Check cache directory
if [[ -n "$XDG_CACHE_HOME" ]]; then
    CACHE_DIR="$XDG_CACHE_HOME/shortcutter"
else
    CACHE_DIR="$HOME/.cache/shortcutter"
fi

echo "Cache directory: $CACHE_DIR"

if [[ -d "$CACHE_DIR" ]]; then
    echo "Cache directory exists"
    ls -la "$CACHE_DIR"
    
    if [[ -f "$CACHE_DIR/shortcuts.json" ]]; then
        echo ""
        echo "Cache file exists, size: $(stat -c%s "$CACHE_DIR/shortcuts.json") bytes"
        echo "Modified: $(stat -c%y "$CACHE_DIR/shortcuts.json")"
        echo ""
        echo "Cache file preview:"
        head -n 10 "$CACHE_DIR/shortcuts.json"
    else
        echo "No cache file found"
    fi
else
    echo "Cache directory does not exist"
fi

echo ""
echo "=== Performance Test ==="
echo "Testing shortcut loading speed..."
time shortcutter --help >/dev/null 2>&1 || time ./shortcutter --help >/dev/null 2>&1 || echo "Could not run shortcutter"