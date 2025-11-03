package config

import (
	"os"
	"path/filepath"
	"testing"
	"time"
)

func TestWatcherCreation(t *testing.T) {
	tmpDir := t.TempDir()
	configPath := filepath.Join(tmpDir, "test.toml")

	// Create test config file
	if err := os.WriteFile(configPath, []byte("[core]\nauto_path = true\n"), 0644); err != nil {
		t.Fatalf("Failed to create config file: %v", err)
	}

	changeDetected := false
	watcher, err := NewWatcher(configPath, func(cfg *Config) {
		changeDetected = true
	})

	if err != nil {
		t.Fatalf("Failed to create watcher: %v", err)
	}
	defer watcher.Stop()

	if watcher.configPath != configPath {
		t.Errorf("Expected path '%s', got '%s'", configPath, watcher.configPath)
	}

	if changeDetected {
		t.Error("Change callback should not be called during creation")
	}
}

func TestWatcherDetectsChanges(t *testing.T) {
	tmpDir := t.TempDir()
	configPath := filepath.Join(tmpDir, "test.toml")

	// Create initial config
	initialConfig := "[core]\nauto_path = true\n"
	if err := os.WriteFile(configPath, []byte(initialConfig), 0644); err != nil {
		t.Fatalf("Failed to create config file: %v", err)
	}

	changeDetected := false
	var detectedConfig *Config

	watcher, err := NewWatcher(configPath, func(cfg *Config) {
		changeDetected = true
		detectedConfig = cfg
	})

	if err != nil {
		t.Fatalf("Failed to create watcher: %v", err)
	}
	defer watcher.Stop()

	watcher.Start()

	// Wait a bit to ensure watcher is running
	time.Sleep(100 * time.Millisecond)

	// Modify the config file
	updatedConfig := "[core]\nauto_path = false\n"
	if err := os.WriteFile(configPath, []byte(updatedConfig), 0644); err != nil {
		t.Fatalf("Failed to update config file: %v", err)
	}

	// Wait for change detection
	time.Sleep(1500 * time.Millisecond)

	if !changeDetected {
		t.Error("Config change was not detected")
	}

	if detectedConfig != nil && detectedConfig.Core != nil {
		if detectedConfig.Core.AutoPath {
			t.Error("Expected AutoPath to be false after update")
		}
	}
}

func TestWatcherStop(t *testing.T) {
	tmpDir := t.TempDir()
	configPath := filepath.Join(tmpDir, "test.toml")

	if err := os.WriteFile(configPath, []byte("[core]\n"), 0644); err != nil {
		t.Fatalf("Failed to create config file: %v", err)
	}

	watcher, err := NewWatcher(configPath, func(cfg *Config) {})
	if err != nil {
		t.Fatalf("Failed to create watcher: %v", err)
	}

	watcher.Start()
	time.Sleep(100 * time.Millisecond)
	watcher.Stop()

	// Modify file after stopping
	if err := os.WriteFile(configPath, []byte("[core]\nauto_path = false\n"), 0644); err != nil {
		t.Fatalf("Failed to update config: %v", err)
	}

	// Change should not be detected
	time.Sleep(1500 * time.Millisecond)
}
