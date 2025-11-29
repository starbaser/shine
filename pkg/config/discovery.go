package config

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"

	"github.com/BurntSushi/toml"
	"github.com/starbased-co/shine/pkg/paths"
)

// PrismSource represents where a prism configuration was discovered
type PrismSource int

const (
	SourceUnknown        PrismSource = iota
	SourceShineToml                  // From shine.toml [prisms.*]
	SourcePrismDir                   // From directory with prism.toml
	SourceStandaloneTOML             // From standalone .toml file in prisms/
	SourceBinaryOnly                 // Binary found but no config
)

// DiscoveredPrism represents a discovered prism with its source
type DiscoveredPrism struct {
	Config *PrismConfig
	Source PrismSource
	Path   string // Path to config file or directory
}

// DiscoverPrisms searches for prisms in the configured prism directories
// Returns a map of prism name -> discovered prism information
// extraPaths are additional directories to search for binaries
func DiscoverPrisms(prismDirs []string, extraPaths []string) (map[string]*DiscoveredPrism, error) {
	discovered := make(map[string]*DiscoveredPrism)

	for _, baseDir := range prismDirs {
		// Expand home directory
		expandedDir := paths.ExpandHome(baseDir)

		// Check if directory exists
		if _, err := os.Stat(expandedDir); os.IsNotExist(err) {
			continue
		}

		// Read directory entries
		entries, err := os.ReadDir(expandedDir)
		if err != nil {
			continue
		}

		for _, entry := range entries {
			// Type 2: Directory with prism.toml
			if entry.IsDir() {
				prism, err := discoverDirectoryPrism(expandedDir, entry.Name(), extraPaths)
				if err == nil && prism != nil {
					// Only add if not already discovered (first match wins)
					if _, exists := discovered[prism.Config.Name]; !exists {
						discovered[prism.Config.Name] = prism
					}
				}
				continue
			}

			// Type 3: Standalone .toml files
			if strings.HasSuffix(entry.Name(), ".toml") && entry.Name() != "prism.toml" {
				prism, err := discoverStandalonePrism(expandedDir, entry.Name(), extraPaths)
				if err == nil && prism != nil {
					// Only add if not already discovered (first match wins)
					if _, exists := discovered[prism.Config.Name]; !exists {
						discovered[prism.Config.Name] = prism
					}
				}
			}
		}
	}

	return discovered, nil
}

// discoverDirectoryPrism handles Type 2: directory with prism.toml
// Binary resolution: checks prism directory first, then falls back to PATH
func discoverDirectoryPrism(baseDir, dirName string, extraPaths []string) (*DiscoveredPrism, error) {
	prismDir := filepath.Join(baseDir, dirName)
	manifestPath := filepath.Join(prismDir, "prism.toml")

	// Check for prism.toml
	if _, err := os.Stat(manifestPath); os.IsNotExist(err) {
		return nil, fmt.Errorf("no prism.toml found in %s", prismDir)
	}

	// Load prism.toml
	config, err := loadPrismConfig(manifestPath)
	if err != nil {
		return nil, fmt.Errorf("failed to load prism.toml: %w", err)
	}

	// Validate name is set
	if config.Name == "" {
		return nil, fmt.Errorf("prism.toml missing required 'name' field")
	}

	// Resolve paths for multi-app prisms
	resolveAppPaths(config, prismDir, extraPaths)

	// Resolve single-app path (for backward compatibility)
	var resolvedPath string
	if config.Path != "" {
		// Check for bundled binary in prism directory
		binaryPath := filepath.Join(prismDir, config.Path)
		if isExecutable(binaryPath) {
			resolvedPath = binaryPath
		} else {
			// Fall back to PATH lookup
			resolvedPath = findInPATH(config.Path, extraPaths)
		}
	} else {
		// Default binary name
		defaultName := "shine-" + config.Name
		binaryPath := filepath.Join(prismDir, defaultName)
		if isExecutable(binaryPath) {
			resolvedPath = binaryPath
		} else {
			resolvedPath = findInPATH(defaultName, extraPaths)
		}
	}

	config.ResolvedPath = resolvedPath

	return &DiscoveredPrism{
		Config: config,
		Source: SourcePrismDir,
		Path:   manifestPath,
	}, nil
}

// discoverStandalonePrism handles Type 3 (standalone .toml files)
func discoverStandalonePrism(baseDir, filename string, extraPaths []string) (*DiscoveredPrism, error) {
	configPath := filepath.Join(baseDir, filename)

	// Load configuration
	config, err := loadPrismConfig(configPath)
	if err != nil {
		return nil, fmt.Errorf("failed to load %s: %w", filename, err)
	}

	// Validate name is set
	if config.Name == "" {
		return nil, fmt.Errorf("%s missing required 'name' field", filename)
	}

	// Resolve paths for multi-app prisms (no prismDir for standalone)
	resolveAppPaths(config, "", extraPaths)

	// Resolve single-app path (always via PATH for standalone configs)
	var binaryName string
	if config.Path != "" {
		binaryName = config.Path
	} else {
		binaryName = "shine-" + config.Name
	}

	config.ResolvedPath = findInPATH(binaryName, extraPaths)

	return &DiscoveredPrism{
		Config: config,
		Source: SourceStandaloneTOML,
		Path:   configPath,
	}, nil
}

// loadPrismConfig loads a PrismConfig from a TOML file
func loadPrismConfig(path string) (*PrismConfig, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}

	var config PrismConfig
	if err := toml.Unmarshal(data, &config); err != nil {
		return nil, err
	}

	return &config, nil
}

// loadShineConfig loads the main shine.toml configuration
func loadShineConfig(path string) (*Config, error) {
	data, err := os.ReadFile(paths.ExpandHome(path))
	if err != nil {
		return nil, err
	}

	var cfg Config
	if err := toml.Unmarshal(data, &cfg); err != nil {
		return nil, err
	}

	return &cfg, nil
}

// MergePrismConfigs merges user config from shine.toml over defaults from prism sources
func MergePrismConfigs(prismSource, userConfig *PrismConfig) *PrismConfig {
	merged := &PrismConfig{}

	// === Core Identification ===
	merged.Name = prismSource.Name
	if userConfig.Name != "" {
		merged.Name = userConfig.Name
	}

	merged.Version = prismSource.Version // Version always from source

	if userConfig.Path != "" {
		merged.Path = userConfig.Path
	} else {
		merged.Path = prismSource.Path
	}

	// === Multi-App Configuration ===
	// If user config has Apps, use it entirely (override, don't merge individual apps)
	// Otherwise use prism source Apps
	if len(userConfig.Apps) > 0 {
		merged.Apps = userConfig.Apps
	} else {
		merged.Apps = prismSource.Apps
	}

	// === Runtime State ===
	// User's enabled setting takes priority
	merged.Enabled = userConfig.Enabled || prismSource.Enabled

	// === Positioning & Layout ===
	merged.Origin = prismSource.Origin
	if userConfig.Origin != "" {
		merged.Origin = userConfig.Origin
	}

	merged.Position = prismSource.Position
	if userConfig.Position != "" {
		merged.Position = userConfig.Position
	}

	merged.Width = prismSource.Width
	if userConfig.Width != nil {
		merged.Width = userConfig.Width
	}

	merged.Height = prismSource.Height
	if userConfig.Height != nil {
		merged.Height = userConfig.Height
	}

	// === Behavior ===
	merged.HideOnFocusLoss = prismSource.HideOnFocusLoss
	if userConfig.HideOnFocusLoss {
		merged.HideOnFocusLoss = userConfig.HideOnFocusLoss
	}

	merged.FocusPolicy = prismSource.FocusPolicy
	if userConfig.FocusPolicy != "" {
		merged.FocusPolicy = userConfig.FocusPolicy
	}

	merged.OutputName = prismSource.OutputName
	if userConfig.OutputName != "" {
		merged.OutputName = userConfig.OutputName
	}

	// === Metadata ===
	// Note: Metadata from user config is intentionally skipped
	merged.Metadata = prismSource.Metadata

	// === Internal fields ===
	merged.ResolvedPath = prismSource.ResolvedPath

	return merged
}

// resolveAppPaths resolves binary paths for all apps in a multi-app prism
func resolveAppPaths(config *PrismConfig, prismDir string, extraPaths []string) {
	if !config.IsMultiApp() {
		return
	}

	for appName, app := range config.Apps {
		if !app.Enabled {
			continue
		}

		var binaryName string
		if app.Path != "" {
			binaryName = app.Path
		} else {
			binaryName = appName // default to app key name
		}

		// Try local directory first (if prismDir provided)
		if prismDir != "" {
			binaryPath := filepath.Join(prismDir, binaryName)
			if isExecutable(binaryPath) {
				app.ResolvedPath = binaryPath
				continue
			}
		}

		// Search in extra paths and system PATH
		app.ResolvedPath = findInPATH(binaryName, extraPaths)
	}
}

// findInPATH searches for an executable in extra paths and system PATH
func findInPATH(name string, extraPaths []string) string {
	// Try extra paths first
	for _, extraPath := range extraPaths {
		expandedPath := paths.ExpandHome(extraPath)
		candidatePath := filepath.Join(expandedPath, name)
		if isExecutable(candidatePath) {
			return candidatePath
		}
	}

	// Fall back to system PATH
	path, err := exec.LookPath(name)
	if err != nil {
		return ""
	}
	return path
}

// isExecutable checks if a file exists and is executable
func isExecutable(path string) bool {
	info, err := os.Stat(path)
	if err != nil {
		return false
	}
	// Check if file and has execute permission
	return !info.IsDir() && (info.Mode()&0111 != 0)
}

