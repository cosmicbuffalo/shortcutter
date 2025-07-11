package internal

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"time"
)

// CacheData represents cached dynamic shortcut data
type CacheData struct {
	BindkeyEntries   []BindkeyEntry                  `json:"bindkey_entries"`
	ManDescriptions  map[string]WidgetDescription    `json:"man_descriptions"`
	CacheVersion     string                          `json:"cache_version"`
	Timestamp        time.Time                       `json:"timestamp"`
	ZshBinaryHash    string                          `json:"zsh_binary_hash"`
	ZshrcHash        string                          `json:"zshrc_hash"`
}

// CacheManager handles caching of dynamic shortcut data
type CacheManager struct {
	cacheDir  string
	cacheFile string
}

// NewCacheManager creates a new cache manager
func NewCacheManager() (*CacheManager, error) {
	cacheDir, err := getCacheDir()
	if err != nil {
		return nil, fmt.Errorf("failed to get cache directory: %w", err)
	}

	// Ensure cache directory exists
	if err := os.MkdirAll(cacheDir, 0755); err != nil {
		return nil, fmt.Errorf("failed to create cache directory: %w", err)
	}

	cm := &CacheManager{
		cacheDir:  cacheDir,
		cacheFile: filepath.Join(cacheDir, "shortcuts.json"),
	}

	return cm, nil
}

// getCacheDir returns the appropriate cache directory
func getCacheDir() (string, error) {
	// Use ~/.config/shortcutter/cache/ for better organization
	homeDir, err := os.UserHomeDir()
	if err != nil {
		return "", fmt.Errorf("failed to get user home directory: %w", err)
	}

	return filepath.Join(homeDir, ".config", "shortcutter", "cache"), nil
}

// LoadCache loads cached data if it exists, returns nil if missing
func (cm *CacheManager) LoadCache() (*CacheData, error) {
	// Check if cache file exists
	if _, err := os.Stat(cm.cacheFile); os.IsNotExist(err) {
		return nil, nil // No cache file
	}

	// Read cache file
	data, err := os.ReadFile(cm.cacheFile)
	if err != nil {
		return nil, fmt.Errorf("failed to read cache file: %w", err)
	}

	// Parse cache data
	var cacheData CacheData
	if err := json.Unmarshal(data, &cacheData); err != nil {
		return nil, fmt.Errorf("failed to parse cache data: %w", err)
	}

	// Trust the cache - no expensive validation during runtime
	return &cacheData, nil
}

// SaveCache saves data to cache
func (cm *CacheManager) SaveCache(bindkeyEntries []BindkeyEntry, manDescriptions map[string]WidgetDescription) error {
	cacheData := CacheData{
		BindkeyEntries:  bindkeyEntries,
		ManDescriptions: manDescriptions,
		CacheVersion:    "1.0",
		Timestamp:       time.Now(),
		ZshBinaryHash:   "", // Not used anymore
		ZshrcHash:       "", // Not used anymore
	}

	// Marshal to JSON
	data, err := json.MarshalIndent(cacheData, "", "  ")
	if err != nil {
		return fmt.Errorf("failed to marshal cache data: %w", err)
	}

	// Write to cache file
	if err := os.WriteFile(cm.cacheFile, data, 0644); err != nil {
		return fmt.Errorf("failed to write cache file: %w", err)
	}

	return nil
}

// ClearCache removes the cache file (called during install)
func (cm *CacheManager) ClearCache() error {
	if err := os.Remove(cm.cacheFile); err != nil && !os.IsNotExist(err) {
		return fmt.Errorf("failed to remove cache file: %w", err)
	}
	return nil
}

