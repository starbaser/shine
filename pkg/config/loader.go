package config

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/BurntSushi/toml"
	"github.com/starbased-co/shine/pkg/paths"
)

func DefaultConfigPath() string {
	return paths.DefaultConfigPath()
}

func Load(path string) (*Config, error) {
	cfg, err := loadShineConfig(path)
	if err != nil {
		return nil, fmt.Errorf("failed to load config file %s: %w", path, err)
	}

	if cfg.Prisms == nil {
		cfg.Prisms = make(map[string]*PrismConfig)
	}

	if cfg.Core != nil && cfg.Core.Path != nil {
		prismDirs := cfg.Core.GetPaths()
		discovered, err := DiscoverPrisms(prismDirs, prismDirs)
		if err == nil {
			for name, discoveredPrism := range discovered {
				if userConfig, exists := cfg.Prisms[name]; exists {
					cfg.Prisms[name] = MergePrismConfigs(discoveredPrism.Config, userConfig)
				} else {
					cfg.Prisms[name] = discoveredPrism.Config
				}
			}
		}
	}

	return cfg, nil
}

func LoadOrDefault(path string) *Config {
	cfg, err := Load(path)
	if err != nil {
		return NewDefaultConfig()
	}

	if cfg.Core == nil {
		cfg.Core = &CoreConfig{
			Path: []string{
				"~/.config/shine/prisms",
			},
		}
	} else if cfg.Core.Path == nil {
		cfg.Core.Path = []string{
			"~/.config/shine/prisms",
		}
	}

	if cfg.Prisms == nil {
		cfg.Prisms = make(map[string]*PrismConfig)
	}

	return cfg
}

func Save(cfg *Config, path string) error {
	if len(path) > 0 && path[0] == '~' {
		home, err := os.UserHomeDir()
		if err != nil {
			return fmt.Errorf("failed to get home directory: %w", err)
		}
		path = filepath.Join(home, path[1:])
	}

	dir := filepath.Dir(path)
	if err := os.MkdirAll(dir, 0755); err != nil {
		return fmt.Errorf("failed to create config directory: %w", err)
	}

	f, err := os.Create(path)
	if err != nil {
		return fmt.Errorf("failed to create config file: %w", err)
	}
	defer f.Close()

	encoder := toml.NewEncoder(f)
	if err := encoder.Encode(cfg); err != nil {
		return fmt.Errorf("failed to encode config: %w", err)
	}

	return nil
}
