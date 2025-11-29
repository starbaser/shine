package config

import "github.com/starbased-co/shine/pkg/panel"

// Config represents the main shine configuration
type Config struct {
	Core   *CoreConfig             `toml:"core"`
	Prisms map[string]*PrismConfig `toml:"prisms"`
}

// CoreConfig holds global shine settings
type CoreConfig struct {
	// Path specifies directories to prepend to PATH for prism binary discovery
	// Can be a single string or array of strings
	// Example: "~/.local/share/shine/bin" or ["~/.local/share/shine/bin", "~/.config/shine/bin"]
	Path interface{} `toml:"path"`
}

// GetPaths normalizes the Path field to []string
// Handles both string and []string types
func (cc *CoreConfig) GetPaths() []string {
	if cc.Path == nil {
		return []string{}
	}

	switch v := cc.Path.(type) {
	case string:
		return []string{v}
	case []interface{}:
		paths := make([]string, 0, len(v))
		for _, item := range v {
			if str, ok := item.(string); ok {
				paths = append(paths, str)
			}
		}
		return paths
	case []string:
		return v
	default:
		return []string{}
	}
}

// PrismConfig is the unified configuration for ALL prisms (built-in and user)
// Used in both shine.toml [prisms.*] sections and prism.toml files
type PrismConfig struct {
	// === Core Identification ===
	// Name is the prism identifier
	Name string `toml:"name"`

	// Version using semantic versioning (optional, primarily for prism.toml)
	Version string `toml:"version,omitempty"`

	// Path specifies a custom binary name or path (optional)
	// If empty, defaults to "shine-{name}"
	// Can be a simple name (e.g., "shine-weather") or a path (e.g., "/usr/bin/shine-weather")
	Path string `toml:"path,omitempty"`

	// === Runtime State ===
	// Enabled controls whether this prism should be launched
	Enabled bool `toml:"enabled"`

	// === Positioning & Layout ===
	Origin   string      `toml:"origin,omitempty"`   // top-left, top-center, top-right, etc.
	Position string      `toml:"position,omitempty"` // "x,y" offset from origin in pixels
	Width    interface{} `toml:"width,omitempty"`    // int or string (with "px" or "%")
	Height   interface{} `toml:"height,omitempty"`   // int or string (with "px" or "%")

	// === Behavior ===
	HideOnFocusLoss bool   `toml:"hide_on_focus_loss,omitempty"`
	FocusPolicy     string `toml:"focus_policy,omitempty"`
	OutputName      string `toml:"output_name,omitempty"`

	// === Metadata (ONLY meaningful in prism sources) ===
	// Metadata contains prism-specific information like description, author, license, etc.
	// During merge, metadata ALWAYS comes from prism source (prism.toml, standalone .toml).
	// Any metadata in shine.toml [prisms.*] is ignored during merge.
	Metadata map[string]interface{} `toml:"metadata,omitempty"`

	// === Internal fields (not from TOML) ===
	// ResolvedPath is the actual path to the binary after discovery
	// This is set during discovery and not read from configuration files
	ResolvedPath string `toml:"-"`
}

// ToPanelConfig converts PrismConfig to panel.Config
func (pc *PrismConfig) ToPanelConfig() *panel.Config {
	cfg := panel.NewConfig()

	// Origin
	if pc.Origin != "" {
		cfg.Origin = panel.ParseOrigin(pc.Origin)
	}

	// Size
	var err error
	if pc.Width != nil {
		cfg.Width, err = panel.ParseDimension(pc.Width)
		if err != nil {
			// Keep default on error
		}
	}

	if pc.Height != nil {
		cfg.Height, err = panel.ParseDimension(pc.Height)
		if err != nil {
			// Keep default on error
		}
	}

	// Position offset from origin
	if pc.Position != "" {
		cfg.Position, err = panel.ParsePosition(pc.Position)
		// Silently ignore position parse errors, keep default
	}

	// Behavior
	cfg.HideOnFocusLoss = pc.HideOnFocusLoss

	// Focus policy
	cfg.FocusPolicy = panel.ParseFocusPolicy(pc.FocusPolicy)

	// If hide_on_focus_loss is enabled, ensure focus policy is on-demand
	if cfg.HideOnFocusLoss {
		cfg.FocusPolicy = panel.FocusOnDemand
	}

	// Output
	cfg.OutputName = pc.OutputName

	// Remote control socket (shared across all components)
	cfg.ListenSocket = "/tmp/shine.sock"

	return cfg
}

// NewDefaultConfig creates a default configuration
func NewDefaultConfig() *Config {
	return &Config{
		Core: &CoreConfig{
			Path: []string{
				"~/.config/shine/prisms",
			},
		},
		Prisms: map[string]*PrismConfig{},
	}
}
