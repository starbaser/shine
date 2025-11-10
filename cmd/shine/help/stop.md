# shine stop

Gracefully stop all panel supervisors and prisms.

## Usage

```bash
shine stop
```

## Description

The `stop` command sends a shutdown signal to all running `prismctl` panel supervisors via IPC. Each supervisor will:
1. Send SIGTERM to all managed prisms (foreground and background)
2. Wait 20ms for graceful shutdown
3. Send SIGKILL if prisms don't exit cleanly
4. Clean up IPC sockets and exit

**Note**: This command stops the panels but not the `shinectl` service manager itself. The service manager will remain running to handle future start requests.

## Options

This command takes no options.

## Examples

### Stop all running panels

```bash
shine stop
```

**Output**:
```
ℹ Stopping shine service...
  Stopping panel-0...
  Stopping panel-1...
✓ Stopped 2 panel(s)
```

### Stop when no panels are running

```bash
shine stop
```

**Output**:
```
⚠ No panels running
```

## Behavior

### Graceful Shutdown

Each prism receives SIGTERM first, allowing it to:
- Save state
- Close file handles
- Flush buffers
- Clean up resources

If a prism doesn't exit within 20ms, prismctl sends SIGKILL to force termination.

### Socket Cleanup

After stopping all prisms, each prismctl supervisor:
- Removes its IPC socket from `/run/user/{uid}/shine/`
- Terminates cleanly

### Service Manager Persistence

The `shinectl` service manager remains running after `shine stop`. This allows for quick restart via `shine start` without re-initializing the service.

To fully shut down shine:
```bash
shine stop          # Stop panels
pkill shinectl      # Stop service manager
```

## Troubleshooting

### Panel won't stop

**Problem**: Panel process remains after `shine stop`

**Solution**: Manually kill the prismctl supervisor:
```bash
# Find the PID
ps aux | grep prismctl

# Kill it
kill <PID>

# Force kill if necessary
kill -9 <PID>
```

### Socket remains after stop

**Problem**: IPC sockets persist in `/run/user/{uid}/shine/`

**Solution**: Clean up manually:
```bash
rm /run/user/$(id -u)/shine/prism-*.sock
```

### Partial shutdown

**Problem**: Some panels stop but others remain running

**Solution**: Check logs for individual panel errors:
```bash
shine logs
```

Then manually stop problematic panels:
```bash
# Find the socket
ls /run/user/$(id -u)/shine/

# Send stop command manually
echo '{"action":"stop"}' | nc -U /run/user/$(id -u)/shine/prism-panel-0.*.sock
```

## Related Commands

- `shine start` - Start the service and panels
- `shine status` - Check which panels are running
- `shine logs` - View logs for debugging shutdown issues

## Technical Details

The `stop` command:
1. Scans `/run/user/{uid}/shine/` for `prism-*.sock` files
2. Connects to each socket via Unix domain socket
3. Sends JSON IPC command: `{"action":"stop"}`
4. Waits for confirmation response
5. Reports any errors encountered

IPC protocol:
```json
// Request
{"action": "stop"}

// Response
{"success": true, "message": "stopped"}
```

The prismctl supervisor handles the "stop" action by:
1. Sending SIGTERM to all prisms (foreground + background)
2. Waiting 20ms grace period
3. Sending SIGKILL to stragglers
4. Removing IPC socket
5. Exiting with status 0

## See Also

- Configuration reference: `~/.config/shine/prism.toml`
- IPC protocol: `/run/user/{uid}/shine/`
- Logs: `~/.local/share/shine/logs/`
