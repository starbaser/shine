package config

import "github.com/starbased-co/shine/pkg/panel"

// Config represents the main shine configuration
type Config struct {
	Core   *CoreConfig               `toml:"core"`
	Prisms map[string]*PrismConfig   `toml:"prisms"`

	// Deprecated: Use Prisms["chat"] instead
	Chat    *ChatConfig    `toml:"chat"`
	// Deprecated: Use Prisms["bar"] instead
	Bar     *BarConfig     `toml:"bar"`
	// Deprecated: Use Prisms["clock"] instead
	Clock   *ClockConfig   `toml:"clock"`
	// Deprecated: Use Prisms["sysinfo"] instead
	SysInfo *SysInfoConfig `toml:"sysinfo"`
}

// CoreConfig holds global shine settings
type CoreConfig struct {
	// PrismDirs specifies directories to search for prism binaries (in priority order)
	PrismDirs []string `toml:"prism_dirs"`

	// AutoPath automatically adds prism directories to PATH for discovery
	AutoPath bool `toml:"auto_path"`

	// DiscoveryMode determines how prisms are discovered
	// Options: "convention" (shine-* naming), "manifest" (prism.toml), "auto" (try both)
	DiscoveryMode string `toml:"discovery_mode"`
}

// PrismConfig is the unified configuration for ALL prisms (built-in and user)
type PrismConfig struct {
	// Name is the prism identifier
	Name string `toml:"name"`

	// Binary specifies a custom binary name or path (optional)
	// If empty, defaults to "shine-{name}"
	// Can be a simple name (e.g., "shine-weather") or a path (e.g., "/usr/bin/shine-weather")
	Binary string `toml:"binary"`

	// Enabled controls whether this prism should be launched
	Enabled bool `toml:"enabled"`

	// Panel configuration
	Edge            string `toml:"edge"`
	Lines           int    `toml:"lines"`
	Columns         int    `toml:"columns"`
	LinesPixels     int    `toml:"lines_pixels"`
	ColumnsPixels   int    `toml:"columns_pixels"`
	MarginTop       int    `toml:"margin_top"`
	MarginLeft      int    `toml:"margin_left"`
	MarginBottom    int    `toml:"margin_bottom"`
	MarginRight     int    `toml:"margin_right"`
	HideOnFocusLoss bool   `toml:"hide_on_focus_loss"`
	FocusPolicy     string `toml:"focus_policy"`
	OutputName      string `toml:"output_name"`
}

// ToPanelConfig converts PrismConfig to panel.Config
func (pc *PrismConfig) ToPanelConfig() *panel.Config {
	cfg := panel.NewConfig()

	// Edge placement
	cfg.Edge = panel.ParseEdge(pc.Edge)

	// Handle background edge special case
	if cfg.Edge == panel.EdgeBackground {
		cfg.Type = panel.LayerShellBackground
	}

	// Size
	cfg.Lines = pc.Lines
	cfg.Columns = pc.Columns
	cfg.LinesPixels = pc.LinesPixels
	cfg.ColumnsPixels = pc.ColumnsPixels

	// Margins
	cfg.MarginTop = pc.MarginTop
	cfg.MarginLeft = pc.MarginLeft
	cfg.MarginBottom = pc.MarginBottom
	cfg.MarginRight = pc.MarginRight

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

// ChatConfig represents chat component configuration
type ChatConfig struct {
	Enabled         bool   `toml:"enabled"`
	Edge            string `toml:"edge"`
	Lines           int    `toml:"lines"`
	Columns         int    `toml:"columns"`
	LinesPixels     int    `toml:"lines_pixels"`
	ColumnsPixels   int    `toml:"columns_pixels"`
	MarginTop       int    `toml:"margin_top"`
	MarginLeft      int    `toml:"margin_left"`
	MarginBottom    int    `toml:"margin_bottom"`
	MarginRight     int    `toml:"margin_right"`
	HideOnFocusLoss bool   `toml:"hide_on_focus_loss"`
	FocusPolicy     string `toml:"focus_policy"`
	OutputName      string `toml:"output_name"`
}

// ToPanelConfig converts ChatConfig to panel.Config
func (cc *ChatConfig) ToPanelConfig() *panel.Config {
	cfg := panel.NewConfig()

	// Edge placement
	cfg.Edge = panel.ParseEdge(cc.Edge)

	// Handle background edge special case
	if cfg.Edge == panel.EdgeBackground {
		cfg.Type = panel.LayerShellBackground
	}

	// Size
	cfg.Lines = cc.Lines
	cfg.Columns = cc.Columns
	cfg.LinesPixels = cc.LinesPixels
	cfg.ColumnsPixels = cc.ColumnsPixels

	// Margins
	cfg.MarginTop = cc.MarginTop
	cfg.MarginLeft = cc.MarginLeft
	cfg.MarginBottom = cc.MarginBottom
	cfg.MarginRight = cc.MarginRight

	// Behavior
	cfg.HideOnFocusLoss = cc.HideOnFocusLoss

	// Focus policy
	cfg.FocusPolicy = panel.ParseFocusPolicy(cc.FocusPolicy)

	// If hide_on_focus_loss is enabled, ensure focus policy is on-demand
	if cfg.HideOnFocusLoss {
		cfg.FocusPolicy = panel.FocusOnDemand
	}

	// Output
	cfg.OutputName = cc.OutputName

	// Remote control socket (shared across all components)
	cfg.ListenSocket = "/tmp/shine.sock"

	return cfg
}

// BarConfig represents status bar component configuration
type BarConfig struct {
	Enabled         bool   `toml:"enabled"`
	Edge            string `toml:"edge"`
	Lines           int    `toml:"lines"`
	Columns         int    `toml:"columns"`
	LinesPixels     int    `toml:"lines_pixels"`
	ColumnsPixels   int    `toml:"columns_pixels"`
	MarginTop       int    `toml:"margin_top"`
	MarginLeft      int    `toml:"margin_left"`
	MarginBottom    int    `toml:"margin_bottom"`
	MarginRight     int    `toml:"margin_right"`
	HideOnFocusLoss bool   `toml:"hide_on_focus_loss"`
	FocusPolicy     string `toml:"focus_policy"`
	OutputName      string `toml:"output_name"`
}

// ToPanelConfig converts BarConfig to panel.Config
func (bc *BarConfig) ToPanelConfig() *panel.Config {
	cfg := panel.NewConfig()

	// Edge placement
	cfg.Edge = panel.ParseEdge(bc.Edge)

	// Handle background edge special case
	if cfg.Edge == panel.EdgeBackground {
		cfg.Type = panel.LayerShellBackground
	}

	// Size
	cfg.Lines = bc.Lines
	cfg.Columns = bc.Columns
	cfg.LinesPixels = bc.LinesPixels
	cfg.ColumnsPixels = bc.ColumnsPixels

	// Margins
	cfg.MarginTop = bc.MarginTop
	cfg.MarginLeft = bc.MarginLeft
	cfg.MarginBottom = bc.MarginBottom
	cfg.MarginRight = bc.MarginRight

	// Behavior
	cfg.HideOnFocusLoss = bc.HideOnFocusLoss

	// Focus policy
	cfg.FocusPolicy = panel.ParseFocusPolicy(bc.FocusPolicy)

	// If hide_on_focus_loss is enabled, ensure focus policy is on-demand
	if cfg.HideOnFocusLoss {
		cfg.FocusPolicy = panel.FocusOnDemand
	}

	// Output
	cfg.OutputName = bc.OutputName

	// Remote control socket (shared across all components)
	cfg.ListenSocket = "/tmp/shine.sock"

	return cfg
}

// ClockConfig represents clock component configuration
type ClockConfig struct {
	Enabled         bool   `toml:"enabled"`
	Edge            string `toml:"edge"`
	Lines           int    `toml:"lines"`
	Columns         int    `toml:"columns"`
	LinesPixels     int    `toml:"lines_pixels"`
	ColumnsPixels   int    `toml:"columns_pixels"`
	MarginTop       int    `toml:"margin_top"`
	MarginLeft      int    `toml:"margin_left"`
	MarginBottom    int    `toml:"margin_bottom"`
	MarginRight     int    `toml:"margin_right"`
	HideOnFocusLoss bool   `toml:"hide_on_focus_loss"`
	FocusPolicy     string `toml:"focus_policy"`
	OutputName      string `toml:"output_name"`
}

// ToPanelConfig converts ClockConfig to panel.Config
func (clc *ClockConfig) ToPanelConfig() *panel.Config {
	cfg := panel.NewConfig()

	// Edge placement
	cfg.Edge = panel.ParseEdge(clc.Edge)

	// Handle background edge special case
	if cfg.Edge == panel.EdgeBackground {
		cfg.Type = panel.LayerShellBackground
	}

	// Size
	cfg.Lines = clc.Lines
	cfg.Columns = clc.Columns
	cfg.LinesPixels = clc.LinesPixels
	cfg.ColumnsPixels = clc.ColumnsPixels

	// Margins
	cfg.MarginTop = clc.MarginTop
	cfg.MarginLeft = clc.MarginLeft
	cfg.MarginBottom = clc.MarginBottom
	cfg.MarginRight = clc.MarginRight

	// Behavior
	cfg.HideOnFocusLoss = clc.HideOnFocusLoss

	// Focus policy
	cfg.FocusPolicy = panel.ParseFocusPolicy(clc.FocusPolicy)

	// If hide_on_focus_loss is enabled, ensure focus policy is on-demand
	if cfg.HideOnFocusLoss {
		cfg.FocusPolicy = panel.FocusOnDemand
	}

	// Output
	cfg.OutputName = clc.OutputName

	// Remote control socket (shared across all components)
	cfg.ListenSocket = "/tmp/shine.sock"

	return cfg
}

// SysInfoConfig represents system info component configuration
type SysInfoConfig struct {
	Enabled         bool   `toml:"enabled"`
	Edge            string `toml:"edge"`
	Lines           int    `toml:"lines"`
	Columns         int    `toml:"columns"`
	LinesPixels     int    `toml:"lines_pixels"`
	ColumnsPixels   int    `toml:"columns_pixels"`
	MarginTop       int    `toml:"margin_top"`
	MarginLeft      int    `toml:"margin_left"`
	MarginBottom    int    `toml:"margin_bottom"`
	MarginRight     int    `toml:"margin_right"`
	HideOnFocusLoss bool   `toml:"hide_on_focus_loss"`
	FocusPolicy     string `toml:"focus_policy"`
	OutputName      string `toml:"output_name"`
}

// ToPanelConfig converts SysInfoConfig to panel.Config
func (sic *SysInfoConfig) ToPanelConfig() *panel.Config {
	cfg := panel.NewConfig()

	// Edge placement
	cfg.Edge = panel.ParseEdge(sic.Edge)

	// Handle background edge special case
	if cfg.Edge == panel.EdgeBackground {
		cfg.Type = panel.LayerShellBackground
	}

	// Size
	cfg.Lines = sic.Lines
	cfg.Columns = sic.Columns
	cfg.LinesPixels = sic.LinesPixels
	cfg.ColumnsPixels = sic.ColumnsPixels

	// Margins
	cfg.MarginTop = sic.MarginTop
	cfg.MarginLeft = sic.MarginLeft
	cfg.MarginBottom = sic.MarginBottom
	cfg.MarginRight = sic.MarginRight

	// Behavior
	cfg.HideOnFocusLoss = sic.HideOnFocusLoss

	// Focus policy
	cfg.FocusPolicy = panel.ParseFocusPolicy(sic.FocusPolicy)

	// If hide_on_focus_loss is enabled, ensure focus policy is on-demand
	if cfg.HideOnFocusLoss {
		cfg.FocusPolicy = panel.FocusOnDemand
	}

	// Output
	cfg.OutputName = sic.OutputName

	// Remote control socket (shared across all components)
	cfg.ListenSocket = "/tmp/shine.sock"

	return cfg
}

// NewDefaultConfig creates a default configuration
func NewDefaultConfig() *Config {
	return &Config{
		Core: &CoreConfig{
			PrismDirs: []string{
				"/usr/lib/shine/prisms",
				"~/.config/shine/prisms",
				"~/.local/share/shine/prisms",
			},
			AutoPath: true,
		},
		Prisms: map[string]*PrismConfig{
			"bar": {
				Name:        "bar",
				Enabled:     true,
				Edge:        "top",
				LinesPixels: 30,
				FocusPolicy: "not-allowed",
			},
		},
		// Backward compatibility: old config format
		Chat: &ChatConfig{
			Enabled:         false,
			Edge:            "bottom",
			Lines:           10,
			MarginLeft:      10,
			MarginRight:     10,
			MarginBottom:    10,
			HideOnFocusLoss: true,
			FocusPolicy:     "on-demand",
		},
	}
}
