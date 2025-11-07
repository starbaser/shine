package config

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestLoad(t *testing.T) {
	// Create temp config file
	tmpDir := t.TempDir()
	configPath := filepath.Join(tmpDir, "test.toml")

	configContent := `[core]
path = ["~/.config/shine/bin"]

[prisms.chat]
name = "chat"
enabled = true
anchor = "bottom"
height = 10
width = 80
margin_left = 10
margin_right = 10
margin_bottom = 10
hide_on_focus_loss = true
focus_policy = "on-demand"
`

	if err := os.WriteFile(configPath, []byte(configContent), 0644); err != nil {
		t.Fatalf("Failed to write test config: %v", err)
	}

	// Load config
	cfg, err := Load(configPath)
	if err != nil {
		t.Fatalf("Failed to load config: %v", err)
	}

	// Verify chat prism config
	chatPrism, ok := cfg.Prisms["chat"]
	if !ok {
		t.Fatal("Chat prism not found")
	}

	if !chatPrism.Enabled {
		t.Error("Expected chat to be enabled")
	}

	if chatPrism.Anchor != "bottom" {
		t.Errorf("Expected anchor=bottom, got %s", chatPrism.Anchor)
	}

	if chatPrism.MarginLeft != 10 {
		t.Errorf("Expected margin_left=10, got %d", chatPrism.MarginLeft)
	}

	if !chatPrism.HideOnFocusLoss {
		t.Error("Expected hide_on_focus_loss=true")
	}

	if chatPrism.FocusPolicy != "on-demand" {
		t.Errorf("Expected focus_policy=on-demand, got %s", chatPrism.FocusPolicy)
	}
}

func TestLoadOrDefault(t *testing.T) {
	// Test with non-existent file
	cfg := LoadOrDefault("/tmp/nonexistent-config-file-xyz.toml")
	if cfg == nil {
		t.Fatal("Expected default config, got nil")
	}

	// Default config should have core and prisms
	if cfg.Core == nil {
		t.Fatal("Expected default core config, got nil")
	}

	if cfg.Prisms == nil {
		t.Fatal("Expected default prisms, got nil")
	}

	// Default prisms should be empty (discovery-based)
	if len(cfg.Prisms) != 0 {
		t.Errorf("Expected empty prisms map, got %d prisms", len(cfg.Prisms))
	}
}

func TestSave(t *testing.T) {
	tmpDir := t.TempDir()
	configPath := filepath.Join(tmpDir, "saved.toml")

	cfg := NewDefaultConfig()
	// Add a test prism
	cfg.Prisms["test"] = &PrismConfig{
		Name:    "test",
		Enabled: true,
		Anchor:  "top",
	}

	if err := Save(cfg, configPath); err != nil {
		t.Fatalf("Failed to save config: %v", err)
	}

	// Verify file exists
	if _, err := os.Stat(configPath); err != nil {
		t.Fatalf("Config file not created: %v", err)
	}

	// Load it back
	loaded, err := Load(configPath)
	if err != nil {
		t.Fatalf("Failed to load saved config: %v", err)
	}

	if testPrism, ok := loaded.Prisms["test"]; !ok || !testPrism.Enabled {
		t.Error("Saved and loaded config don't match")
	}
}

func TestLoad_PrismConfig(t *testing.T) {
	tmpDir := t.TempDir()
	configPath := filepath.Join(tmpDir, "test.toml")

	configContent := `[core]
path = ["/usr/lib/shine/bin", "~/.config/shine/bin"]

[prisms.bar]
enabled = true
edge = "top"
lines_pixels = 30
focus_policy = "not-allowed"

[prisms.weather]
enabled = true
name = "weather"
path = "shine-weather"
edge = "top-right"
columns_pixels = 200
`

	if err := os.WriteFile(configPath, []byte(configContent), 0644); err != nil {
		t.Fatalf("Failed to write test config: %v", err)
	}

	cfg, err := Load(configPath)
	if err != nil {
		t.Fatalf("Failed to load config: %v", err)
	}

	// Verify core config
	if cfg.Core == nil {
		t.Fatal("Core config is nil")
	}

	paths := cfg.Core.GetPaths()
	if len(paths) != 2 {
		t.Errorf("Expected 2 paths, got %d", len(paths))
	}

	// Verify prisms
	if cfg.Prisms == nil {
		t.Fatal("Prisms map is nil")
	}

	barCfg, ok := cfg.Prisms["bar"]
	if !ok {
		t.Fatal("Bar prism not found")
	}

	if !barCfg.Enabled {
		t.Error("Expected bar to be enabled")
	}

	if barCfg.Edge != "top" {
		t.Errorf("Expected edge 'top', got %s", barCfg.Edge)
	}

	weatherCfg, ok := cfg.Prisms["weather"]
	if !ok {
		t.Fatal("Weather prism not found")
	}

	if weatherCfg.Path != "shine-weather" {
		t.Errorf("Expected path 'shine-weather', got %s", weatherCfg.Path)
	}
}


func TestLoadOrDefault_InitializesDefaults(t *testing.T) {
	tmpDir := t.TempDir()
	configPath := filepath.Join(tmpDir, "test.toml")

	// Minimal config without core
	configContent := `[prisms.bar]
enabled = true
`

	if err := os.WriteFile(configPath, []byte(configContent), 0644); err != nil {
		t.Fatalf("Failed to write test config: %v", err)
	}

	cfg := LoadOrDefault(configPath)

	// Should have default core config
	if cfg.Core == nil {
		t.Fatal("Core config should be initialized")
	}

	paths := cfg.Core.GetPaths()
	if len(paths) == 0 {
		t.Error("Should have default paths")
	}
}

func TestPrismConfig_ToPanelConfig(t *testing.T) {
	prismCfg := &PrismConfig{
		Name:          "test",
		Enabled:       true,
		Edge:          "top",
		LinesPixels:   30,
		MarginTop:     10,
		MarginLeft:    20,
		FocusPolicy:   "not-allowed",
		OutputName:    "DP-2",
	}

	panelCfg := prismCfg.ToPanelConfig()

	if panelCfg == nil {
		t.Fatal("Panel config is nil")
	}

	if panelCfg.Height.Value != 30 || !panelCfg.Height.IsPixels {
		t.Errorf("Expected height 30px, got %+v", panelCfg.Height)
	}

	if panelCfg.MarginTop != 10 {
		t.Errorf("Expected margin_top 10, got %d", panelCfg.MarginTop)
	}

	if panelCfg.MarginLeft != 20 {
		t.Errorf("Expected margin_left 20, got %d", panelCfg.MarginLeft)
	}

	if panelCfg.OutputName != "DP-2" {
		t.Errorf("Expected output_name 'DP-2', got %s", panelCfg.OutputName)
	}
}

func TestNewDefaultConfig_HasPrisms(t *testing.T) {
	cfg := NewDefaultConfig()

	if cfg == nil {
		t.Fatal("Default config is nil")
	}

	if cfg.Core == nil {
		t.Fatal("Default config missing core")
	}

	if cfg.Prisms == nil {
		t.Fatal("Default config missing prisms")
	}

	// Default config should have empty prisms (populated by discovery)
	if len(cfg.Prisms) != 0 {
		t.Errorf("Expected empty prisms in default config, got %d", len(cfg.Prisms))
	}

	// Core should have prisms directory in search paths
	paths := cfg.Core.GetPaths()
	found := false
	for _, p := range paths {
		if strings.Contains(p, "prisms") {
			found = true
			break
		}
	}
	if !found {
		t.Error("Default config should include prisms directory in search paths")
	}
}
