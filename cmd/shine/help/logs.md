# shine logs

View log files from the shine service and panels.

## Usage

```bash
# List all available log files
shine logs

# View specific log file (last 50 lines)
shine logs <filename>
```

## Description

The `logs` command provides access to shine's logging system. Logs are stored in `~/.local/share/shine/logs/` and include:
- `shinectl.log` - Service manager logs (panel spawning, config reload)
- `prismctl-{component}.log` - Panel supervisor logs (process management)
- Individual prism logs (if configured)

Without arguments, the command lists all available log files. With a filename (or partial name), it displays the last 50 lines of that log.

## Options

This command takes no options but accepts an optional log filename argument.

## Examples

### List all log files

```bash
shine logs
```

**Output**:
```
ℹ Log directory: /home/user/.local/share/shine/logs
╭──────────────────────┬────────────╮
│ Log File             │ Size       │
├──────────────────────┼────────────┤
│ shinectl.log         │ 4523 bytes │
│ prismctl-panel-0.log │ 2891 bytes │
│ prismctl-panel-1.log │ 1204 bytes │
╰──────────────────────┴────────────╯

ℹ View a log with: shine logs <filename>
```

### View shinectl logs

```bash
shine logs shinectl
```

**Output** (last 50 lines):
```
2025-11-09 14:23:45 INFO Starting shinectl service
2025-11-09 14:23:45 INFO Config loaded from /home/user/.config/shine/prism.toml
2025-11-09 14:23:45 INFO Found 3 enabled prisms
2025-11-09 14:23:46 INFO Spawned panel-0 for prisms: clock, chat, weather
2025-11-09 14:23:46 INFO Spawned panel-1 for prism: bar
2025-11-09 14:23:50 INFO Received SIGHUP, reloading configuration...
2025-11-09 14:23:50 INFO Configuration reloaded successfully
```

### View panel supervisor logs

```bash
shine logs prismctl-panel-0
```

**Output**:
```
2025-11-09 14:23:46 INFO prismctl starting for panel-0
2025-11-09 14:23:46 INFO IPC socket: /run/user/1000/shine/prism-panel-0.12345.sock
2025-11-09 14:23:46 INFO Launching prism: shine-clock
2025-11-09 14:23:46 INFO Launching prism: shine-chat
2025-11-09 14:23:46 INFO Launching prism: shine-weather
2025-11-09 14:23:46 INFO Foreground: shine-clock
2025-11-09 14:23:46 INFO Background: shine-chat, shine-weather
```

### Filename shortcuts

You can omit the `.log` extension:
```bash
# These are equivalent
shine logs shinectl.log
shine logs shinectl
```

## Log File Details

### shinectl.log

**Contents**: Service manager lifecycle events
- Startup and shutdown
- Configuration loading and reloading
- Panel spawning and monitoring
- Error conditions

**Common entries**:
```
INFO Starting shinectl service
INFO Config loaded from ...
INFO Found N enabled prisms
INFO Spawned panel-X for prisms: ...
ERROR Failed to spawn panel: ...
INFO Received SIGHUP, reloading configuration...
```

### prismctl-{component}.log

**Contents**: Panel supervisor lifecycle events
- Prism launching and termination
- Foreground/background switching
- IPC command handling
- Process monitoring and restart

**Common entries**:
```
INFO prismctl starting for panel-X
INFO IPC socket: ...
INFO Launching prism: shine-clock
INFO Suspended: shine-chat (PID: 12345)
INFO Resumed: shine-clock (PID: 12346)
INFO IPC command: start prism=shine-weather
ERROR Prism crashed: shine-chat (exit code: 1)
INFO Restarting prism: shine-chat (attempt 1/3)
```

## Troubleshooting with Logs

### Panel won't start

Check shinectl logs:
```bash
shine logs shinectl
```

Look for:
- `ERROR Failed to spawn panel: ...` - Configuration or permission issue
- `ERROR Kitty remote control not available` - Kitty setup problem
- `ERROR Prism binary not found: ...` - Missing prism executable

### Prism keeps crashing

Check panel supervisor logs:
```bash
shine logs prismctl-panel-0
```

Look for:
- `ERROR Prism crashed: ... (exit code: 1)` - Prism error
- `INFO Restarting prism: ...` - Restart loop
- `ERROR Max restarts exceeded` - Hit restart limit

Then check the prism's own logs (if it produces any) or run standalone:
```bash
./shine-clock  # Run prism directly to see errors
```

### Configuration not reloading

Check shinectl logs:
```bash
shine logs shinectl
```

Look for:
- `INFO Received SIGHUP, reloading configuration...` - Reload triggered
- `ERROR Failed to load config: ...` - Syntax error in prism.toml
- `INFO Configuration reloaded successfully` - Should appear after SIGHUP

### IPC connection failures

Check panel supervisor logs:
```bash
shine logs prismctl-panel-0
```

Look for:
- `INFO IPC socket: ...` - Verify socket path
- `ERROR Failed to create IPC socket: ...` - Permission or path issue

Verify socket exists:
```bash
ls -la /run/user/$(id -u)/shine/
```

## Advanced Usage

### Follow logs in real-time

Use standard Unix tools:
```bash
# Follow shinectl logs
tail -f ~/.local/share/shine/logs/shinectl.log

# Follow with filtering
tail -f ~/.local/share/shine/logs/shinectl.log | grep ERROR
```

### Search logs

```bash
# Search for errors
grep ERROR ~/.local/share/shine/logs/*.log

# Search for specific prism
grep "shine-clock" ~/.local/share/shine/logs/*.log

# Search with context
grep -C 3 "crashed" ~/.local/share/shine/logs/*.log
```

### Rotate logs

Logs are not automatically rotated. To clear old logs:
```bash
# Clear all logs
rm ~/.local/share/shine/logs/*.log

# Or archive them
tar czf shine-logs-$(date +%Y%m%d).tar.gz ~/.local/share/shine/logs/
rm ~/.local/share/shine/logs/*.log
```

### Debug mode

For more verbose logging, set log level in config:
```toml
# ~/.config/shine/prism.toml
[core]
log_level = "debug"  # Options: debug, info, warning, error
```

Then reload:
```bash
pkill -HUP shinectl
```

## Related Commands

- `shine status` - View current panel state (real-time)
- `shine start` - Start service (check logs if it fails)
- `shine stop` - Stop service (logs will show shutdown sequence)

## Technical Details

The `logs` command:
1. Resolves log directory: `~/.local/share/shine/logs/`
2. Lists files with `os.ReadDir()` if no argument provided
3. If filename provided:
   - Appends `.log` if extension missing
   - Uses `tail -n 50` to show last 50 lines
   - Pipes output directly to stdout

Log locations:
- **Config**: Defined in `core.data_dir` (default: `~/.local/share/shine`)
- **Actual logs**: `{data_dir}/logs/`
- **Permissions**: Created with mode 0755 (user read/write/execute)

## See Also

- Log configuration: `~/.config/shine/prism.toml`
- Data directory: `~/.local/share/shine/`
- Debugging guide: docs/TROUBLESHOOTING.md (if available)
