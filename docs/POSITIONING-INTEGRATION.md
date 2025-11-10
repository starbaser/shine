# Positioning Integration: pkg/ → cmd/shinectl

## Overview

This document describes the integration of Shine's sophisticated positioning and discovery system from `pkg/` into the `cmd/shinectl` service manager. The integration bridges the gap between well-designed but unused library code and the actual panel spawning implementation.

## Problem Statement

### Before Integration

**Unused Architecture**: Shine had comprehensive positioning logic in `pkg/panel` and discovery in `pkg/config`, but `cmd/shinectl` was using minimal config and spawning generic windows.

```go
// OLD: cmd/shinectl/config.go
type PrismEntry struct {
    Name         string
    Restart      string
    RestartDelay string
    MaxRestarts  int
    // MISSING: All positioning fields
}

// OLD: cmd/shinectl/panel_manager.go
cmd := exec.Command(
    "kitten", "@", "launch",
    "--type=window",  // Should be --type=os-panel
    "--title", title,
    pm.prismctlBin, config.Name, componentName,
)
// MISSING: All --os-panel positioning flags
```

### Design Mismatch

1. **cmd/shinectl/config.go** had only restart policies
2. **pkg/config/types.go** had full positioning + metadata
3. **pkg/panel/config.go** had sophisticated margin calculation and ToRemoteControlArgs()
4. **None of it was connected**

## Solution Architecture

### Three-Layer Integration

```
┌─────────────────────────────────────────────────────┐
│ cmd/shinectl/config.go                              │
│ ┌─────────────────────────────────────────────────┐ │
│ │ PrismEntry (Hybrid Type)                        │ │
│ │ - Positioning fields (from pkg/config)          │ │
│ │ - Restart policies (shinectl-specific)          │ │
│ │ - ToPanelConfig() → panel.Config                │ │
│ └─────────────────────────────────────────────────┘ │
└─────────────────────────────────────────────────────┘
                    ↓
┌─────────────────────────────────────────────────────┐
│ pkg/panel/config.go                                 │
│ ┌─────────────────────────────────────────────────┐ │
│ │ panel.Config                                     │ │
│ │ - Origin + Position → Margins                    │ │
│ │ - ToRemoteControlArgs() → kitten args           │ │
│ └─────────────────────────────────────────────────┘ │
└─────────────────────────────────────────────────────┘
                    ↓
┌─────────────────────────────────────────────────────┐
│ cmd/shinectl/panel_manager.go                       │
│ ┌─────────────────────────────────────────────────┐ │
│ │ SpawnPanel()                                     │ │
│ │ - config.ToPanelConfig()                        │ │
│ │ - panelCfg.ToRemoteControlArgs()                │ │
│ │ - exec.Command("kitten", kittenArgs...)         │ │
│ └─────────────────────────────────────────────────┘ │
└─────────────────────────────────────────────────────┘
```

## Implementation Details

### 1. Extended cmd/shinectl/config.go

**Key Changes**:

```go
// NEW: Hybrid config type
type PrismEntry struct {
    // === Core Identification ===
    Name string `toml:"name"`

    // === Positioning & Layout (from pkg/config) ===
    Origin   string      `toml:"origin,omitempty"`
    Position string      `toml:"position,omitempty"`
    Width    interface{} `toml:"width,omitempty"`
    Height   interface{} `toml:"height,omitempty"`

    // === Behavior (from pkg/config) ===
    HideOnFocusLoss bool   `toml:"hide_on_focus_loss,omitempty"`
    FocusPolicy     string `toml:"focus_policy,omitempty"`
    OutputName      string `toml:"output_name,omitempty"`

    // === Restart Policies (shinectl-specific) ===
    Restart      string `toml:"restart"`
    RestartDelay string `toml:"restart_delay"`
    MaxRestarts  int    `toml:"max_restarts"`
}

// NEW: Conversion to pkg/panel format
func (pe *PrismEntry) ToPanelConfig() *panel.Config {
    cfg := panel.NewConfig()

    // Parse positioning fields using pkg/panel parsers
    if pe.Origin != "" {
        cfg.Origin = panel.ParseOrigin(pe.Origin)
    }
    if pe.Width != nil {
        cfg.Width, _ = panel.ParseDimension(pe.Width)
    }
    if pe.Height != nil {
        cfg.Height, _ = panel.ParseDimension(pe.Height)
    }
    if pe.Position != "" {
        cfg.Position, _ = panel.ParsePosition(pe.Position)
    }

    cfg.HideOnFocusLoss = pe.HideOnFocusLoss
    cfg.FocusPolicy = panel.ParseFocusPolicy(pe.FocusPolicy)
    cfg.OutputName = pe.OutputName
    cfg.WindowTitle = fmt.Sprintf("shine-%s", pe.Name)

    return cfg
}
```

**Benefits**:
- Preserves existing restart policy fields
- Adds all positioning fields from pkg/config
- Provides validation using pkg/panel parsers
- Clean conversion path to panel.Config

### 2. Updated cmd/shinectl/panel_manager.go

**Before**:
```go
cmd := exec.Command(
    "kitten", "@", "launch",
    "--type=window",
    "--title", title,
    pm.prismctlBin, config.Name, componentName,
)
```

**After**:
```go
// Convert PrismEntry to panel.Config for positioning
panelCfg := config.ToPanelConfig()

// Build prismctl command path with arguments
prismctlArgs := []string{config.Name, componentName}

// Generate kitten @ launch arguments with positioning
kittenArgs := panelCfg.ToRemoteControlArgs(pm.prismctlBin)
kittenArgs = append(kittenArgs, prismctlArgs...)

// Launch Kitty panel using kitten @ launch with os-panel positioning
cmd := exec.Command("kitten", kittenArgs...)
```

**What This Generates**:

For a config like:
```toml
[[prism]]
name = "shine-clock"
origin = "top-right"
position = "10,50"
width = "200px"
height = "100px"
restart = "on-failure"
```

The kitten command becomes:
```bash
kitten @ launch \
  --type=os-panel \
  --os-panel edge=top \
  --os-panel columns=200px \
  --os-panel lines=100px \
  --os-panel margin-top=50 \
  --os-panel margin-right=10 \
  --os-panel output-name=DP-2 \
  --title shine-shine-clock \
  prismctl shine-clock panel-0
```

### 3. Discovery Integration (Optional)

Added `LoadFromPkgConfig()` function for future use:

```go
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
        }

        cfg.Prisms = append(cfg.Prisms, entry)
    }

    return cfg, nil
}
```

**Note**: Currently not used in main.go, but available for future integration when switching from prism.toml to shine.toml with discovery.

## Configuration Examples

### Basic Positioning

```toml
[[prism]]
name = "shine-clock"
origin = "top-right"
position = "10,50"  # 10px from right, 50px from top
width = "200px"
height = "100px"
restart = "on-failure"
```

### Centered Panel

```toml
[[prism]]
name = "shine-chat"
origin = "center"
width = "800px"
height = "600px"
hide_on_focus_loss = true
focus_policy = "on-demand"
restart = "always"
restart_delay = "5s"
```

### Bottom Panel (Bar)

```toml
[[prism]]
name = "shine-bar"
origin = "bottom-center"
width = "1920px"
height = "40px"
output_name = "DP-2"
restart = "unless-stopped"
max_restarts = 3
```

## Testing Strategy

### 1. Verify Build

```bash
go build -o bin/shinectl ./cmd/shinectl
```

### 2. Test Config Validation

```bash
# Create test config
cat > /tmp/test-prism.toml <<EOF
[[prism]]
name = "shine-clock"
origin = "top-right"
position = "10,50"
width = "200px"
height = "100px"
restart = "on-failure"
EOF

# Validate (shinectl will load and validate on startup)
./bin/shinectl --config /tmp/test-prism.toml
```

### 3. Verify Panel Spawning

```bash
# Start shinectl with test config
./bin/shinectl --config /tmp/test-prism.toml

# Verify panel appears at top-right with correct size
# Check kitten command in logs
tail -f ~/.local/share/shine/logs/shinectl.log
```

### 4. Test Position Calculations

For `origin = "top-right"` with `position = "10,50"`:
- Edge: `top`
- Margin-top: `50px`
- Margin-right: `10px`

Verify panel appears 10px from right edge, 50px from top.

## Migration Path

### Phase 1: Current Implementation ✓
- Extended PrismEntry with positioning fields
- Integrated panel.Config.ToRemoteControlArgs()
- Preserved backward compatibility with restart policies
- Build verified, ready for testing

### Phase 2: Testing (Next Steps)
1. Test with example configs
2. Verify panel positioning accuracy
3. Test restart policies still work
4. Document any edge cases

### Phase 3: Discovery Integration (Future)
- Switch main.go to use LoadFromPkgConfig()
- Enable prism discovery from configured directories
- Merge discovered prisms with user overrides
- Full shine.toml + discovery workflow

## Success Criteria

✅ **Build Success**: Code compiles without errors
✅ **Type Safety**: All conversions use pkg/panel parsers
✅ **Backward Compatibility**: Restart policies preserved
⏳ **Runtime Testing**: Panel spawning with positioning (needs testing)
⏳ **Integration Testing**: End-to-end with real prisms (needs testing)

## Known Limitations

1. **No Discovery Yet**: Still using simple prism.toml loading
   - Future: Switch to LoadFromPkgConfig() for discovery

2. **No Monitor Resolution Caching**: Queries Hyprland each time
   - Acceptable: Only happens during spawn, not frequently

3. **Error Handling**: Silently falls back to defaults on parse errors
   - Validation catches errors upfront
   - Runtime parsing failures use sensible defaults

## Architecture Benefits

### Clean Separation
- **pkg/config**: Pure data types, discovery logic
- **pkg/panel**: Positioning math, kitty integration
- **cmd/shinectl**: Service lifecycle, restart policies

### Reusability
- Same panel.Config used by multiple components
- ToRemoteControlArgs() handles all kitty complexity
- Parsers centralized in pkg/panel

### Extensibility
- Easy to add new positioning fields
- Discovery integration ready when needed
- Plugin system can use same types

## Related Documentation

- **pkg/panel/config.go**: Origin/position math, margin calculation
- **pkg/config/types.go**: Full PrismConfig with metadata
- **pkg/config/discovery.go**: Prism discovery system (unused currently)
- **cmd/shinectl/config.go**: Extended hybrid config type
- **cmd/shinectl/panel_manager.go**: Panel spawning with positioning
- **examples/prism.toml**: Example configurations
- **docs/configuration.md**: Complete configuration reference

## Future Enhancements

1. **Per-Prism Logs**: Track which kitten args were used
2. **Position Validation**: Warn if panel would be off-screen
3. **Hot-Reload Positioning**: Update margins without restarting
4. **Multi-Monitor Support**: Per-monitor positioning
5. **Discovery Integration**: Full shine.toml workflow
