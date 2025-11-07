# Prism System Migration Guide

## Overview

Shine has transitioned to a unified "prism" system where ALL widgets are treated identically. This guide explains the changes and how to migrate your configuration.

## What Changed?

### Old System (Deprecated)
- Widgets configured in separate sections: `[bar]`, `[chat]`, `[clock]`, `[sysinfo]`
- No distinction between built-in and user widgets
- Hardcoded component discovery

### New System (Prism)
- ALL widgets are "prisms" configured under `[prisms.*]`
- Unified discovery mechanism
- Configurable prism directories
- Support for user-created prisms

## Migration Steps

### Option 1: No Action (Automatic Migration)

Your old config will continue working! Shine automatically migrates old-format configs internally.

**What you'll see:**
```
⚠️  Warning: Detected deprecated config format ([bar], [chat], etc.)
   Consider migrating to new [prisms.*] format.
```

**No action required** - everything will work as before.

### Option 2: Manual Migration

Update your `~/.config/shine/shine.toml`:

**Before (Old Format):**
```toml
[bar]
enabled = true
edge = "top"
lines_pixels = 30
focus_policy = "not-allowed"

[chat]
enabled = false
edge = "bottom"
lines = 10
```

**After (New Format):**
```toml
# Add core configuration
[core]
prism_dirs = [
    "/usr/lib/shine/prisms",
    "~/.config/shine/prisms",
    "~/.local/share/shine/prisms",
]
auto_path = true

# Convert [bar] → [prisms.bar]
[prisms.bar]
enabled = true
edge = "top"
lines_pixels = 30
focus_policy = "not-allowed"

# Convert [chat] → [prisms.chat]
[prisms.chat]
enabled = false
edge = "bottom"
lines = 10
```

### Option 3: Use Example Config

```bash
# Backup old config
cp ~/.config/shine/shine.toml ~/.config/shine/shine.toml.backup

# Copy new example
cp examples/shine.toml ~/.config/shine/shine.toml

# Edit to match your preferences
vim ~/.config/shine/shine.toml
```

## Configuration Reference

### Core Section

```toml
[core]
# Directories to search for prism binaries (in priority order)
prism_dirs = [
    "/usr/lib/shine/prisms",      # System-wide prisms
    "~/.config/shine/prisms",      # User prisms
    "~/.local/share/shine/prisms", # Alternative user location
]

# Automatically add prism directories to PATH for discovery
auto_path = true
```

### Prism Sections

```toml
[prisms.{name}]
# Enable/disable this prism
enabled = true

# Optional: Override binary name or path
# If not set, defaults to "shine-{name}"
path = "shine-custom-name"

# Panel configuration (same as before)
edge = "top"
lines_pixels = 30
margin_top = 0
focus_policy = "not-allowed"
output_name = "DP-2"
```

## Complete Migration Example

**Old Config:**
```toml
[bar]
enabled = true
edge = "top"
lines_pixels = 30
margin_top = 0
focus_policy = "not-allowed"
output_name = "DP-2"

[chat]
enabled = false
edge = "bottom"
lines = 10
margin_left = 10
hide_on_focus_loss = true

[clock]
enabled = true
edge = "top-right"
columns_pixels = 150

[sysinfo]
enabled = false
```

**New Config:**
```toml
[core]
prism_dirs = [
    "/usr/lib/shine/prisms",
    "~/.config/shine/prisms",
    "~/.local/share/shine/prisms",
]
auto_path = true

[prisms.bar]
enabled = true
edge = "top"
lines_pixels = 30
margin_top = 0
focus_policy = "not-allowed"
output_name = "DP-2"

[prisms.chat]
enabled = false
edge = "bottom"
lines = 10
margin_left = 10
hide_on_focus_loss = true

[prisms.clock]
enabled = true
edge = "top-right"
columns_pixels = 150

[prisms.sysinfo]
enabled = false
```

## Benefits of New Format

### 1. User Prisms
Add custom widgets without modifying Shine:

```toml
[prisms.weather]
enabled = true
path = "shine-weather"  # Your custom path
edge = "top-right"
columns_pixels = 200
```

### 2. Flexible Discovery
Specify where to look for prism binaries:

```toml
[core]
prism_dirs = [
    "~/my-prisms",           # Custom location
    "~/.config/shine/prisms"
]
```

### 3. Unified Interface
All prisms configured the same way - no special cases.

## Creating User Prisms

Want to create your own prism? It's simple!

### 1. Create Binary

Any Bubble Tea program can be a prism:

```bash
# Your prism in any language
mkdir -p ~/.config/shine/prisms/
cd ~/.config/shine/prisms/

# Example: Go-based weather prism
mkdir shine-weather
cd shine-weather
go mod init shine-weather
go get github.com/charmbracelet/bubbletea
# ... write your code ...
go build -o shine-weather
```

### 2. Add to Config

```toml
[prisms.weather]
enabled = true
edge = "top-right"
columns_pixels = 200
```

### 3. Launch

```bash
shine  # Your weather prism will be discovered and launched!
```

## Discovery Priority

When looking for a prism binary, Shine searches in order:

1. **Explicit path** from config (`binary` field)
2. **Cache** (previous discoveries)
3. **PATH** (includes prism dirs if `auto_path = true`)
4. **Prism directories** (from `prism_dirs`)
5. **Shine executable directory** (backward compatibility)

## Naming Convention

Prisms follow the `shine-{name}` convention:

- Prism name: `bar` → Binary: `shine-bar`
- Prism name: `weather` → Binary: `shine-weather`

Override with `binary` field:
```toml
[prisms.mywidget]
path = "custom-widget-name"
```

## Troubleshooting

### "Prism not found" Error

```
Failed to launch prism weather: failed to find prism binary: prism weather not found
```

**Solutions:**
1. Check binary exists: `ls ~/.config/shine/prisms/shine-weather`
2. Check binary is executable: `chmod +x ~/.config/shine/prisms/shine-weather`
3. Check prism_dirs includes the directory
4. Try explicit path: `path = "/full/path/to/binary"`

### Deprecation Warning

```
⚠️  Warning: Detected deprecated config format ([bar], [chat], etc.)
```

**Solution:** This is just a warning. Your config still works! Migrate when convenient.

### Wrong Binary Launched

If wrong prism launches, check discovery priority. Use explicit path:

```toml
[prisms.bar]
path = "/usr/local/bin/shine-bar"  # Explicit path to binary
```

## Migration Checklist

- [ ] Backup current config: `cp ~/.config/shine/shine.toml ~/.config/shine/shine.toml.backup`
- [ ] Add `[core]` section with `prism_dirs` and `auto_path`
- [ ] Rename `[bar]` → `[prisms.bar]`
- [ ] Rename `[chat]` → `[prisms.chat]`
- [ ] Rename `[clock]` → `[prisms.clock]`
- [ ] Rename `[sysinfo]` → `[prisms.sysinfo]`
- [ ] Test: `./bin/shine` (verify all prisms launch)
- [ ] Remove old `[bar]`, `[chat]` sections
- [ ] No more deprecation warning!

## Future Features

Coming in Phase 2:
- `shine new-prism <name>` - Template generator for creating prisms
- Example prisms: weather, spotify, system monitor
- Prism development guide

Coming in Phase 3:
- Manifest-based discovery (`prism.toml`)
- Hot reload for config changes
- Prism validation and sandboxing
- Prism marketplace

## Getting Help

- Design documentation: `docs/PRISM_SYSTEM_DESIGN.md`
- Implementation summary: `docs/PHASE_1_IMPLEMENTATION_SUMMARY.md`
- Example config: `examples/shine.toml`
- GitHub issues: https://github.com/starbased-co/shine/issues

## Summary

- **Old config still works** - automatic migration with warning
- **New format is simple** - just rename sections and add `[core]`
- **Benefits are clear** - user prisms, flexible discovery, unified interface
- **Migration is optional** - do it when convenient
- **No breaking changes** - backward compatibility maintained

Happy prisming! ✨
