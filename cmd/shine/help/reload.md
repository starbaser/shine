# shine reload

Reload the shine configuration and update panels without stopping prisms.

## Usage

```bash
shine reload
```

## Description

The `reload` command triggers a configuration refresh in the `shinectl` service manager. This allows you to apply changes from `~/.config/shine/prism.toml` without restarting panels or interrupting running prisms.

**Current Status**: ⚠️ **Not yet fully implemented** via IPC. Manual workaround available.

When reload is fully implemented, it will:
1. Re-read `~/.config/shine/prism.toml`
2. Compare current config with running state
3. Start newly enabled prisms
4. Stop disabled prisms
5. Update panel geometry (position, size) if changed
6. Preserve running prisms where config is unchanged

## Options

This command takes no options.

## Examples

### Reload configuration (manual workaround)

```bash
shine reload
```

**Output**:
```
ℹ Reloading configuration...
⚠ Config reload via IPC not yet implemented
ℹ To reload config, send SIGHUP to shinectl process:
  pkill -HUP shinectl
```

**Manual reload**:
```bash
# Send SIGHUP to shinectl
pkill -HUP shinectl

# Verify reload worked
shine status
```

### Common reload scenarios

#### Enable a new prism

Edit `~/.config/shine/prism.toml`:
```toml
[prisms.weather]
enabled = true  # Changed from false
origin = "top-left"
width = "300px"
height = "200px"
```

Reload:
```bash
pkill -HUP shinectl
shine status  # Should show weather panel
```

#### Disable a prism

Edit `~/.config/shine/prism.toml`:
```toml
[prisms.clock]
enabled = false  # Changed from true
```

Reload:
```bash
pkill -HUP shinectl
shine status  # clock panel should be gone
```

#### Change panel geometry

Edit `~/.config/shine/prism.toml`:
```toml
[prisms.bar]
enabled = true
width = "100%"
height = "50px"  # Changed from 40px
```

Reload:
```bash
pkill -HUP shinectl
# Panel should resize
```

## Configuration Hot-Reload Behavior

### What gets reloaded

- ✅ `enabled` flag (start/stop prisms)
- ✅ Panel geometry (`origin`, `width`, `height`)
- ✅ Panel behavior (`anchor`, `margin`)
- ✅ Restart policies (`restart`, `restart_delay`, `max_restarts`)
- ✅ Prism configuration files (re-read `prism.toml` in prism directories)

### What does NOT reload

- ❌ Prism binaries (must restart panels to load new binaries)
- ❌ Environment variables
- ❌ Core config (`core.config_dir`, `core.data_dir`, `core.log_level`)

### State preservation

Hot-reload preserves:
- Running prism processes (unless explicitly stopped)
- Prism application state (in-memory data)
- MRU ordering (foreground/background state)
- IPC connections

## Troubleshooting

### Reload has no effect

**Problem**: `pkill -HUP shinectl` executes but config changes don't apply

**Solution**: Verify shinectl received the signal:
```bash
# Check shinectl logs
shine logs shinectl

# Should show: "Received SIGHUP, reloading configuration..."
```

If no log entry, shinectl may not be running:
```bash
ps aux | grep shinectl
shine start
```

### New prism doesn't start after reload

**Problem**: Enabled a prism in config but it doesn't appear

**Solution**: Check configuration syntax and prism availability:
```bash
# Verify prism binary exists
which shine-weather

# Check config syntax
cat ~/.config/shine/prism.toml

# View shinectl logs for errors
shine logs shinectl
```

### Panel geometry doesn't update

**Problem**: Changed `width`/`height` but panel size unchanged

**Solution**: Kitty panel geometry updates may not take effect until restart:
```bash
shine stop
shine start
```

**Note**: This is a known limitation. Full reload should handle geometry updates in the future.

### Prism restarts unexpectedly

**Problem**: After reload, a prism restarts even though config wasn't changed

**Solution**: This may be a bug in config comparison logic. Check shinectl logs:
```bash
shine logs shinectl
```

If the issue persists, manually stop/start instead of reload:
```bash
shine stop
shine start
```

## Related Commands

- `shine start` - Start the service (reads config initially)
- `shine stop` - Stop panels (config ignored)
- `shine status` - Verify reload effects

## Technical Details (Future Implementation)

The `reload` command will:
1. Connect to shinectl IPC socket
2. Send JSON command: `{"action":"reload"}`
3. shinectl will:
   - Re-read `~/.config/shine/prism.toml`
   - Diff against in-memory config
   - Send updates to panel managers
   - Start/stop prisms as needed
4. Return success/failure response

IPC protocol (planned):
```json
// Request
{"action": "reload"}

// Response
{
  "success": true,
  "message": "configuration reloaded",
  "data": {
    "started": ["shine-weather"],
    "stopped": ["shine-clock"],
    "updated": ["shine-bar"]
  }
}
```

## Current Manual Workaround

Until IPC reload is implemented, use SIGHUP:

```bash
# 1. Edit config
vim ~/.config/shine/prism.toml

# 2. Send reload signal
pkill -HUP shinectl

# 3. Verify changes
shine status
```

Or restart the service for guaranteed config reload:
```bash
shine stop
shine start
```

## See Also

- Configuration reference: `~/.config/shine/prism.toml`
- Service logs: `~/.local/share/shine/logs/shinectl.log`
- Signal handling: docs/PHASE2-3-IMPLEMENTATION.md
