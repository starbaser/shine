# shine status

Display the current state of all panel supervisors and their managed prisms.

## Usage

```bash
shine status
```

## Description

The `status` command queries all running `prismctl` panel supervisors via IPC to retrieve:
- Which panel is running
- Number of managed prisms
- Foreground prism (actively displayed)
- Background prisms (suspended with SIGSTOP)
- Per-prism state (PID and suspend state)

This provides a real-time snapshot of the shine panel system.

## Options

This command takes no options.

## Examples

### View status with multiple panels running

```bash
shine status
```

**Output**:
```
═══ Shine Status (2 panel(s)) ═══

Panel: panel-0
Socket: /run/user/1000/shine/prism-panel-0.12345.sock

╭─────────────────────────────────────╮
│ Foreground: shine-clock             │
│ Background: 2 prisms                │
│ Total: 3 prisms                     │
╰─────────────────────────────────────╯

╭────────────────┬───────┬────────────╮
│ Prism          │ PID   │ State      │
├────────────────┼───────┼────────────┤
│ shine-clock    │ 12346 │ foreground │
│ shine-chat     │ 12347 │ background │
│ shine-weather  │ 12348 │ background │
╰────────────────┴───────┴────────────╯

Panel: panel-1
Socket: /run/user/1000/shine/prism-panel-1.12350.sock

╭─────────────────────────────────────╮
│ Foreground: shine-bar               │
│ Background: 0 prisms                │
│ Total: 1 prism                      │
╰─────────────────────────────────────╯

╭──────────────┬───────┬────────────╮
│ Prism        │ PID   │ State      │
├──────────────┼───────┼────────────┤
│ shine-bar    │ 12351 │ foreground │
╰──────────────┴───────┴────────────╯
```

### View status when no panels are running

```bash
shine status
```

**Output**:
```
⚠ No panels running
ℹ Start panels with: shine start
```

## Understanding the Output

### Panel Information

- **Panel**: Component name (e.g., "panel-0", "panel-1")
- **Socket**: IPC socket path for direct communication

### Status Box

- **Foreground**: Currently visible prism in the panel
- **Background**: Number of suspended prisms (hidden but running)
- **Total**: Sum of foreground (1) + background prisms

### Prisms Table

Each row shows:
- **Prism**: Binary name (e.g., "shine-clock", "shine-chat")
- **PID**: Process ID of the prism
- **State**:
  - `foreground` - Currently displayed, receiving input
  - `background` - Suspended with SIGSTOP, not consuming CPU

## Panel State Management

### Foreground vs Background

**Foreground prism**:
- Actively running (not suspended)
- Receiving keyboard/mouse input
- Drawing to the Kitty panel terminal
- Consuming CPU resources

**Background prisms**:
- Suspended via SIGSTOP signal
- Not consuming CPU (frozen in place)
- Memory intact (state preserved)
- Can be resumed instantly with SIGCONT

### MRU Ordering

Panels use Most Recently Used (MRU) ordering:
- The last-accessed prism stays in foreground
- Others are moved to background
- No automatic eviction (unlimited background prisms)

## Troubleshooting

### Status query fails

**Problem**: `Failed to query: connection refused`

**Solution**: The panel's prismctl supervisor may have crashed. Check logs:
```bash
shine logs
```

Restart the service:
```bash
shine stop
shine start
```

### No panels shown but shinectl is running

**Problem**: `shine status` shows no panels, but `ps aux | grep shinectl` shows the process

**Solution**: The panels failed to spawn. Check shinectl logs:
```bash
shine logs shinectl
```

Common causes:
- Configuration errors in `prism.toml`
- Kitty remote control unavailable
- Prism binaries not in PATH
- Missing prism configuration files

### Prism stuck in background

**Problem**: A prism is listed as "background" but should be foreground

**Solution**: Manually switch to it via IPC:
```bash
# Find the panel socket
SOCK=$(ls -t /run/user/$(id -u)/shine/prism-*.sock | head -1)

# Send start command to bring prism to foreground
echo '{"action":"start","prism":"shine-clock"}' | nc -U $SOCK
```

### Prism PID mismatch

**Problem**: Status shows a PID but `ps` shows different/no process

**Solution**: The prism may have crashed. Check for restart loops:
```bash
# Check prism logs
shine logs

# Verify process exists
ps aux | grep shine-clock

# Check restart policy in config
cat ~/.config/shine/prism.toml
```

## Related Commands

- `shine start` - Start panels if none are running
- `shine stop` - Stop all panels
- `shine logs` - View detailed logs for debugging

## Technical Details

The `status` command:
1. Scans `/run/user/{uid}/shine/` for `prism-*.sock` files
2. Connects to each socket
3. Sends JSON IPC command: `{"action":"status"}`
4. Parses response containing panel state
5. Formats output with lipgloss styling

IPC protocol:
```json
// Request
{"action": "status"}

// Response
{
  "success": true,
  "data": {
    "foreground": "shine-clock",
    "background": ["shine-chat", "shine-weather"],
    "prisms": [
      {"name": "shine-clock", "pid": 12346, "state": "foreground"},
      {"name": "shine-chat", "pid": 12347, "state": "background"},
      {"name": "shine-weather", "pid": 12348, "state": "background"}
    ]
  }
}
```

## See Also

- Panel architecture: docs/PHASE2-3-IMPLEMENTATION.md
- Configuration reference: `~/.config/shine/prism.toml`
- IPC protocol: `/run/user/{uid}/shine/`
