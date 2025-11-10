# shine start

Start or resume the shinectl service manager and all enabled panels.

## Usage

```bash
shine start
```

## Description

The `start` command initializes the shine service by launching the `shinectl` service manager. If shinectl is already running, the command will report success without taking action.

When shinectl starts, it:
1. Reads the configuration from `~/.config/shine/prism.toml`
2. Spawns Kitty panels for each enabled prism via remote control
3. Launches `prismctl` supervisors to manage prism lifecycle
4. Creates IPC sockets in `/run/user/{uid}/shine/` for communication

## Options

This command takes no options.

## Examples

### Start the service for the first time

```bash
shine start
```

**Output**:
```
ℹ Starting shinectl service...
✓ shinectl started (PID: 12345)
```

### Start when already running

```bash
shine start
```

**Output**:
```
✓ shinectl is already running
```

## Configuration

Before starting, ensure your configuration is set up:

```toml
# ~/.config/shine/prism.toml
[prisms.clock]
enabled = true
origin = "top-right"
width = "200px"
height = "100px"

[prisms.bar]
enabled = true
origin = "bottom"
width = "100%"
height = "40px"
```

Only prisms with `enabled = true` will start automatically.

## Troubleshooting

### Service won't start

**Problem**: `shine start` fails with "shinectl not found in PATH"

**Solution**: Ensure shinectl is installed and in your PATH:
```bash
which shinectl
# If not found, install or add to PATH
export PATH="$HOME/.local/bin:$PATH"
```

### Socket creation timeout

**Problem**: `shinectl started but socket not created within timeout`

**Solution**: Check shinectl logs for errors:
```bash
shine logs shinectl
```

Common causes:
- Configuration file syntax errors
- Missing Kitty remote control socket
- Insufficient permissions on `/run/user/{uid}/shine/`

### Kitty remote control not available

**Problem**: Panels don't appear after starting

**Solution**: Verify Kitty remote control is enabled:
```bash
# Check kitty.conf
grep allow_remote_control ~/.config/kitty/kitty.conf
# Should show: allow_remote_control yes

# Verify socket exists
kitty @ ls
```

Add to `~/.config/kitty/kitty.conf`:
```
allow_remote_control yes
listen_on unix:/tmp/@mykitty
```

## Related Commands

- `shine status` - Check if service is running and view panel status
- `shine stop` - Stop the service and all panels
- `shine logs shinectl` - View service manager logs

## Technical Details

The `start` command:
1. Checks for existing shinectl socket at `/run/user/{uid}/shine/shine-service.*.sock`
2. If not found, executes `shinectl` binary with detached stdio
3. Polls for socket creation with 5-second timeout (50 attempts × 100ms)
4. Reports PID and success status

The shinectl process runs in the background and does not daemonize itself. It relies on the shell's job control for background execution.

## See Also

- Configuration reference: `~/.config/shine/prism.toml`
- Service logs: `~/.local/share/shine/logs/shinectl.log`
- IPC sockets: `/run/user/{uid}/shine/`
