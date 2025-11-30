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

// DiscoverPrisms searches for prisms in the configured prism directories.
// extraPaths are additional directories to search for binaries.
func DiscoverPrisms(prismDirs []string, extraPaths []string) (map[string]*DiscoveredPrism, error) {
	discovered := make(map[string]*DiscoveredPrism)

	for _, baseDir := range prismDirs {
		expandedDir := paths.ExpandHome(baseDir)

		if _, err := os.Stat(expandedDir); os.IsNotExist(err) {
			continue
		}

		entries, err := os.ReadDir(expandedDir)
		if err != nil {
			continue
		}

		for _, entry := range entries {
			if entry.IsDir() {
				prism, err := discoverDirectoryPrism(expandedDir, entry.Name(), extraPaths)
				if err == nil && prism != nil {
					if _, exists := discovered[prism.Config.Name]; !exists {
						discovered[prism.Config.Name] = prism
					}
				}
				continue
			}

			if strings.HasSuffix(entry.Name(), ".toml") && entry.Name() != "prism.toml" {
				prism, err := discoverStandalonePrism(expandedDir, entry.Name(), extraPaths)
				if err == nil && prism != nil {
					if _, exists := discovered[prism.Config.Name]; !exists {
						discovered[prism.Config.Name] = prism
					}
				}
			}
		}
	}

	return discovered, nil
}

// discoverDirectoryPrism handles directory with prism.toml.
// Binary resolution: checks prism directory first, then falls back to PATH.
func discoverDirectoryPrism(baseDir, dirName string, extraPaths []string) (*DiscoveredPrism, error) {
	prismDir := filepath.Join(baseDir, dirName)
	manifestPath := filepath.Join(prismDir, "prism.toml")

	if _, err := os.Stat(manifestPath); os.IsNotExist(err) {
		return nil, fmt.Errorf("no prism.toml found in %s", prismDir)
	}

	config, err := loadPrismConfig(manifestPath)
	if err != nil {
		return nil, fmt.Errorf("failed to load prism.toml: %w", err)
	}

	if config.Name == "" {
		return nil, fmt.Errorf("prism.toml missing required 'name' field")
	}

	resolveAppPaths(config, prismDir, extraPaths)

	var resolvedPath string
	if config.Path != "" {
		binaryPath := filepath.Join(prismDir, config.Path)
		if isExecutable(binaryPath) {
			resolvedPath = binaryPath
		} else {
			resolvedPath = findInPATH(config.Path, extraPaths)
		}
	} else {
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

func discoverStandalonePrism(baseDir, filename string, extraPaths []string) (*DiscoveredPrism, error) {
	configPath := filepath.Join(baseDir, filename)

	config, err := loadPrismConfig(configPath)
	if err != nil {
		return nil, fmt.Errorf("failed to load %s: %w", filename, err)
	}

	if config.Name == "" {
		return nil, fmt.Errorf("%s missing required 'name' field", filename)
	}

	resolveAppPaths(config, "", extraPaths)

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

func MergePrismConfigs(prismSource, userConfig *PrismConfig) *PrismConfig {
	merged := &PrismConfig{}

	merged.Name = prismSource.Name
	if userConfig.Name != "" {
		merged.Name = userConfig.Name
	}

	merged.Version = prismSource.Version

	if userConfig.Path != "" {
		merged.Path = userConfig.Path
	} else {
		merged.Path = prismSource.Path
	}

	if len(userConfig.Apps) > 0 {
		merged.Apps = userConfig.Apps
	} else {
		merged.Apps = prismSource.Apps
	}

	merged.Enabled = userConfig.Enabled || prismSource.Enabled

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

	// Metadata from user config is intentionally skipped
	merged.Metadata = prismSource.Metadata
	merged.ResolvedPath = prismSource.ResolvedPath

	return merged
}

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
			binaryName = appName
		}

		if prismDir != "" {
			binaryPath := filepath.Join(prismDir, binaryName)
			if isExecutable(binaryPath) {
				app.ResolvedPath = binaryPath
				continue
			}
		}

		app.ResolvedPath = findInPATH(binaryName, extraPaths)
	}
}

func findInPATH(name string, extraPaths []string) string {
	for _, extraPath := range extraPaths {
		expandedPath := paths.ExpandHome(extraPath)
		candidatePath := filepath.Join(expandedPath, name)
		if isExecutable(candidatePath) {
			return candidatePath
		}
	}

	path, err := exec.LookPath(name)
	if err != nil {
		return ""
	}
	return path
}

func isExecutable(path string) bool {
	info, err := os.Stat(path)
	if err != nil {
		return false
	}
	return !info.IsDir() && (info.Mode()&0111 != 0)
}

