package main

import (
	"fmt"
	"os"
	"path/filepath"
	"time"

	"github.com/BurntSushi/toml"
	"github.com/starbased-co/shine/pkg/config"
	"github.com/starbased-co/shine/pkg/panel"
)

// Config represents the shinectl configuration
type Config struct {
	Prisms []PrismEntry `toml:"prism"`
}

// PrismEntry represents a single prism configuration entry in prism.toml
// Combines positioning from pkg/config.PrismConfig with restart policies
type PrismEntry struct {
	// === Core Identification ===
	Name string `toml:"name"`

	// === Positioning & Layout (from pkg/config) ===
	Origin   string      `toml:"origin,omitempty"`   // top-left, top-center, top-right, etc.
	Position string      `toml:"position,omitempty"` // "x,y" offset from origin in pixels
	Width    interface{} `toml:"width,omitempty"`    // int or string (with "px" or "%")
	Height   interface{} `toml:"height,omitempty"`   // int or string (with "px" or "%")

	// === Behavior (from pkg/config) ===
	HideOnFocusLoss bool   `toml:"hide_on_focus_loss,omitempty"`
	FocusPolicy     string `toml:"focus_policy,omitempty"`
	OutputName      string `toml:"output_name,omitempty"`

	// === Restart Policies (shinectl-specific) ===
	Restart      string `toml:"restart"`       // always | on-failure | unless-stopped | no
	RestartDelay string `toml:"restart_delay"` // Duration string (e.g., "5s")
	MaxRestarts  int    `toml:"max_restarts"`  // Max restarts per hour (0 = unlimited)
}

// RestartPolicy represents the restart behavior
type RestartPolicy int

const (
	RestartNo RestartPolicy = iota
	RestartOnFailure
	RestartUnlessStopped
	RestartAlways
)

// GetRestartPolicy converts string to RestartPolicy enum
func (pe *PrismEntry) GetRestartPolicy() RestartPolicy {
	switch pe.Restart {
	case "always":
		return RestartAlways
	case "on-failure":
		return RestartOnFailure
	case "unless-stopped":
		return RestartUnlessStopped
	case "no", "":
		return RestartNo
	default:
		return RestartNo
	}
}

// GetRestartDelay parses the restart_delay string into a Duration
func (pe *PrismEntry) GetRestartDelay() time.Duration {
	if pe.RestartDelay == "" {
		return 1 * time.Second // Default
	}
	d, err := time.ParseDuration(pe.RestartDelay)
	if err != nil {
		return 1 * time.Second // Fallback
	}
	return d
}

// ToPanelConfig converts PrismEntry to panel.Config for positioning
func (pe *PrismEntry) ToPanelConfig() *panel.Config {
	cfg := panel.NewConfig()

	// Origin
	if pe.Origin != "" {
		cfg.Origin = panel.ParseOrigin(pe.Origin)
	}

	// Size
	var err error
	if pe.Width != nil {
		cfg.Width, err = panel.ParseDimension(pe.Width)
		if err != nil {
			// Keep default on error
		}
	}

	if pe.Height != nil {
		cfg.Height, err = panel.ParseDimension(pe.Height)
		if err != nil {
			// Keep default on error
		}
	}

	// Position offset from origin
	if pe.Position != "" {
		cfg.Position, err = panel.ParsePosition(pe.Position)
		// Silently ignore position parse errors, keep default
	}

	// Behavior
	cfg.HideOnFocusLoss = pe.HideOnFocusLoss

	// Focus policy
	cfg.FocusPolicy = panel.ParseFocusPolicy(pe.FocusPolicy)

	// If hide_on_focus_loss is enabled, ensure focus policy is on-demand
	if cfg.HideOnFocusLoss {
		cfg.FocusPolicy = panel.FocusOnDemand
	}

	// Output
	cfg.OutputName = pe.OutputName

	// Window title for targeting
	cfg.WindowTitle = fmt.Sprintf("shine-%s", pe.Name)

	return cfg
}

// DefaultConfigPath returns the default prism.toml location
func DefaultConfigPath() string {
	home, err := os.UserHomeDir()
	if err != nil {
		return ""
	}
	return filepath.Join(home, ".config", "shine", "prism.toml")
}

// LoadConfig loads prism.toml from the given path
func LoadConfig(path string) (*Config, error) {
	// Expand ~ to home directory
	if len(path) > 0 && path[0] == '~' {
		home, err := os.UserHomeDir()
		if err != nil {
			return nil, fmt.Errorf("failed to get home directory: %w", err)
		}
		path = filepath.Join(home, path[1:])
	}

	var cfg Config
	_, err := toml.DecodeFile(path, &cfg)
	if err != nil {
		return nil, fmt.Errorf("failed to parse config: %w", err)
	}

	return &cfg, nil
}

// LoadConfigOrDefault loads config from path, returns empty config if not found
func LoadConfigOrDefault(path string) *Config {
	cfg, err := LoadConfig(path)
	if err != nil {
		// Return empty config - no prisms to spawn
		return &Config{Prisms: []PrismEntry{}}
	}
	return cfg
}

// LoadFromPkgConfig loads configuration using pkg/config system with discovery
// This provides full prism discovery and merging capabilities
func LoadFromPkgConfig(path string) (*Config, error) {
	// Use pkg/config.Load for discovery and merging
	pkgCfg, err := config.Load(path)
	if err != nil {
		return nil, fmt.Errorf("failed to load config: %w", err)
	}

	// Convert pkg/config.Config to shinectl Config
	cfg := &Config{
		Prisms: make([]PrismEntry, 0, len(pkgCfg.Prisms)),
	}

	for name, prismCfg := range pkgCfg.Prisms {
		// Only include enabled prisms with resolved paths
		if !prismCfg.Enabled || prismCfg.ResolvedPath == "" {
			continue
		}

		entry := PrismEntry{
			Name:            name,
			Origin:          prismCfg.Origin,
			Position:        prismCfg.Position,
			Width:           prismCfg.Width,
			Height:          prismCfg.Height,
			HideOnFocusLoss: prismCfg.HideOnFocusLoss,
			FocusPolicy:     prismCfg.FocusPolicy,
			OutputName:      prismCfg.OutputName,
			// Note: Restart policies remain at defaults unless explicitly set in prism.toml
		}

		cfg.Prisms = append(cfg.Prisms, entry)
	}

	return cfg, nil
}

// Validate checks if the config is valid
func (c *Config) Validate() error {
	seen := make(map[string]bool)
	for i, prism := range c.Prisms {
		if prism.Name == "" {
			return fmt.Errorf("prism[%d]: name is required", i)
		}
		if seen[prism.Name] {
			return fmt.Errorf("prism[%d]: duplicate name %q", i, prism.Name)
		}
		seen[prism.Name] = true

		// Validate restart policy
		switch prism.Restart {
		case "", "no", "on-failure", "unless-stopped", "always":
			// Valid
		default:
			return fmt.Errorf("prism[%d] %q: invalid restart policy %q", i, prism.Name, prism.Restart)
		}

		// Validate restart_delay if present
		if prism.RestartDelay != "" {
			if _, err := time.ParseDuration(prism.RestartDelay); err != nil {
				return fmt.Errorf("prism[%d] %q: invalid restart_delay %q: %w", i, prism.Name, prism.RestartDelay, err)
			}
		}

		// Validate max_restarts
		if prism.MaxRestarts < 0 {
			return fmt.Errorf("prism[%d] %q: max_restarts must be >= 0", i, prism.Name)
		}

		// Validate positioning fields using pkg/panel parsers
		if prism.Origin != "" {
			// Validate origin is recognized
			_ = panel.ParseOrigin(prism.Origin)
		}

		if prism.Position != "" {
			if _, err := panel.ParsePosition(prism.Position); err != nil {
				return fmt.Errorf("prism[%d] %q: invalid position %q: %w", i, prism.Name, prism.Position, err)
			}
		}

		if prism.Width != nil {
			if _, err := panel.ParseDimension(prism.Width); err != nil {
				return fmt.Errorf("prism[%d] %q: invalid width %v: %w", i, prism.Name, prism.Width, err)
			}
		}

		if prism.Height != nil {
			if _, err := panel.ParseDimension(prism.Height); err != nil {
				return fmt.Errorf("prism[%d] %q: invalid height %v: %w", i, prism.Name, prism.Height, err)
			}
		}

		if prism.FocusPolicy != "" {
			// Validate focus policy is recognized
			_ = panel.ParseFocusPolicy(prism.FocusPolicy)
		}
	}
	return nil
}
