package internal

import (
	"encoding/json"
	"os"
	"path/filepath"
	"testing"
	"time"
)

func TestNewCacheManager(t *testing.T) {
	// Set HOME to temp directory
	tempDir, err := os.MkdirTemp("", "shortcutter-cache-test")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tempDir)

	originalHome := os.Getenv("HOME")
	defer os.Setenv("HOME", originalHome)
	os.Setenv("HOME", tempDir)

	cm, err := NewCacheManager()
	if err != nil {
		t.Fatalf("NewCacheManager() error: %v", err)
	}

	expectedCacheDir := filepath.Join(tempDir, ".config", "shortcutter", "cache")
	if cm.cacheDir != expectedCacheDir {
		t.Errorf("cacheDir = %q, want %q", cm.cacheDir, expectedCacheDir)
	}

	expectedCacheFile := filepath.Join(expectedCacheDir, "shortcuts.json")
	if cm.cacheFile != expectedCacheFile {
		t.Errorf("cacheFile = %q, want %q", cm.cacheFile, expectedCacheFile)
	}

	// Check that cache directory was created
	if _, err := os.Stat(cm.cacheDir); os.IsNotExist(err) {
		t.Error("Cache directory was not created")
	}
}

func TestGetCacheDir(t *testing.T) {
	originalHome := os.Getenv("HOME")
	defer os.Setenv("HOME", originalHome)

	tests := []struct {
		name        string
		homeDir     string
		expected    string
	}{
		{
			name:     "normal home directory",
			homeDir:  "/home/user",
			expected: "/home/user/.config/shortcutter/cache",
		},
		{
			name:     "different home directory",
			homeDir:  "/Users/testuser",
			expected: "/Users/testuser/.config/shortcutter/cache",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			os.Setenv("HOME", tt.homeDir)

			cacheDir, err := getCacheDir()
			if err != nil {
				t.Fatalf("getCacheDir() error: %v", err)
			}

			if cacheDir != tt.expected {
				t.Errorf("getCacheDir() = %q, want %q", cacheDir, tt.expected)
			}
		})
	}
}

func TestLoadCacheNoFile(t *testing.T) {
	// Create temporary directory for testing
	tempDir, err := os.MkdirTemp("", "shortcutter-cache-test")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tempDir)

	cm := &CacheManager{
		cacheDir:  tempDir,
		cacheFile: filepath.Join(tempDir, "shortcuts.json"),
	}

	// Try to load cache when no file exists
	cacheData, err := cm.LoadCache()
	if err != nil {
		t.Fatalf("LoadCache() error: %v", err)
	}

	if cacheData != nil {
		t.Error("LoadCache() should return nil when no cache file exists")
	}
}

func TestSaveAndLoadCache(t *testing.T) {
	// Create temporary directory for testing
	tempDir, err := os.MkdirTemp("", "shortcutter-cache-test")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tempDir)

	cm := &CacheManager{
		cacheDir:  tempDir,
		cacheFile: filepath.Join(tempDir, "shortcuts.json"),
	}

	// Create test data
	bindkeyEntries := []BindkeyEntry{
		{EscapeSequence: "^A", WidgetName: "beginning-of-line", DisplayName: "Ctrl+A"},
		{EscapeSequence: "^E", WidgetName: "end-of-line", DisplayName: "Ctrl+E"},
	}

	manDescriptions := map[string]WidgetDescription{
		"beginning-of-line": {
			WidgetName:       "beginning-of-line",
			ShortDescription: "Move to the beginning of the line.",
			FullDescription:  "Move to the beginning of the line.",
		},
		"end-of-line": {
			WidgetName:       "end-of-line",
			ShortDescription: "Move to the end of the line.",
			FullDescription:  "Move to the end of the line.",
		},
	}

	// No need for hash functions in optimized version

	// Save cache
	err = cm.SaveCache(bindkeyEntries, manDescriptions)
	if err != nil {
		t.Fatalf("SaveCache() error: %v", err)
	}

	// Check that cache file was created
	if _, err := os.Stat(cm.cacheFile); os.IsNotExist(err) {
		t.Fatal("Cache file was not created")
	}

	// Load cache
	loadedCache, err := cm.LoadCache()
	if err != nil {
		t.Fatalf("LoadCache() error: %v", err)
	}

	if loadedCache == nil {
		t.Fatal("LoadCache() returned nil")
	}

	// Verify loaded data
	if len(loadedCache.BindkeyEntries) != len(bindkeyEntries) {
		t.Errorf("Loaded bindkey entries count = %d, want %d", len(loadedCache.BindkeyEntries), len(bindkeyEntries))
	}

	for i, entry := range loadedCache.BindkeyEntries {
		if entry.WidgetName != bindkeyEntries[i].WidgetName {
			t.Errorf("BindkeyEntry[%d].WidgetName = %q, want %q", i, entry.WidgetName, bindkeyEntries[i].WidgetName)
		}
	}

	if len(loadedCache.ManDescriptions) != len(manDescriptions) {
		t.Errorf("Loaded man descriptions count = %d, want %d", len(loadedCache.ManDescriptions), len(manDescriptions))
	}

	for widget, desc := range manDescriptions {
		loadedDesc := loadedCache.ManDescriptions[widget]
		if loadedDesc.ShortDescription != desc.ShortDescription {
			t.Errorf("ManDescriptions[%q].ShortDescription = %q, want %q", widget, loadedDesc.ShortDescription, desc.ShortDescription)
		}
		if loadedDesc.FullDescription != desc.FullDescription {
			t.Errorf("ManDescriptions[%q].FullDescription = %q, want %q", widget, loadedDesc.FullDescription, desc.FullDescription)
		}
	}

	// Verify metadata
	if loadedCache.CacheVersion != "1.0" {
		t.Errorf("CacheVersion = %q, want %q", loadedCache.CacheVersion, "1.0")
	}
}

// TestIsCacheValid removed - cache validation is now done at install time

func TestClearCache(t *testing.T) {
	// Create temporary directory for testing
	tempDir, err := os.MkdirTemp("", "shortcutter-cache-test")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tempDir)

	cm := &CacheManager{
		cacheDir:  tempDir,
		cacheFile: filepath.Join(tempDir, "shortcuts.json"),
	}

	// Create a cache file
	testData := CacheData{
		CacheVersion: "1.0",
		Timestamp:    time.Now(),
	}
	data, err := json.Marshal(testData)
	if err != nil {
		t.Fatalf("Failed to marshal test data: %v", err)
	}

	err = os.WriteFile(cm.cacheFile, data, 0644)
	if err != nil {
		t.Fatalf("Failed to create test cache file: %v", err)
	}

	// Verify file exists
	if _, err := os.Stat(cm.cacheFile); os.IsNotExist(err) {
		t.Fatal("Test cache file was not created")
	}

	// Clear cache
	err = cm.ClearCache()
	if err != nil {
		t.Fatalf("ClearCache() error: %v", err)
	}

	// Verify file is gone
	if _, err := os.Stat(cm.cacheFile); !os.IsNotExist(err) {
		t.Error("Cache file still exists after ClearCache()")
	}

	// Test clearing non-existent cache (should not error)
	err = cm.ClearCache()
	if err != nil {
		t.Fatalf("ClearCache() on non-existent file error: %v", err)
	}
}

func TestLoadCacheInvalidJSON(t *testing.T) {
	// Create temporary directory for testing
	tempDir, err := os.MkdirTemp("", "shortcutter-cache-test")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tempDir)

	cm := &CacheManager{
		cacheDir:  tempDir,
		cacheFile: filepath.Join(tempDir, "shortcuts.json"),
	}

	// Create invalid JSON file
	err = os.WriteFile(cm.cacheFile, []byte("invalid json"), 0644)
	if err != nil {
		t.Fatalf("Failed to create invalid cache file: %v", err)
	}

	// Try to load cache
	_, err = cm.LoadCache()
	if err == nil {
		t.Error("LoadCache() should return error for invalid JSON")
	}
}

// TestGetZshPath removed - hash functions no longer used

// TestGetZshrcPath removed - hash functions no longer used

// TestHashFunctions removed - hash functions no longer used