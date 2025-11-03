# Single-Instance Architecture Solution

## TL;DR

**PROBLEM**: Current architecture uses 230 MB per widget (920 MB for 4 widgets)
**SOLUTION**: Use `kitty @ launch --type=os-panel` for 21.5 MB per widget (86 MB for 4 widgets)
**SAVINGS**: 90.6% memory reduction

## How It Works

Instead of launching separate Kitty processes via `kitten panel --single-instance`, use Kitty's remote control API to create panel windows in an existing Kitty instance:

```bash
# Current approach (multi-instance)
kitten panel --edge=top --lines=1 /path/to/shine-clock
# Result: New Kitty process (~230 MB)

# New approach (single-instance via remote control)
kitty @ launch --type=os-panel --os-panel edge=top --os-panel lines=1 --title=shine-clock /path/to/shine-clock
# Result: New window in existing Kitty (~7.5 MB overhead + 14 MB component = 21.5 MB total)
```

## Tested & Validated

- ✅ Kitty 0.43.1 supports `--type=os-panel` in remote control API
- ✅ All panel configuration options work (`edge`, `lines`, `columns`, margins)
- ✅ Window management works (close, focus, list via `kitty @` commands)
- ✅ No new Kitty processes created (4 widgets = 0 new processes)
- ✅ Real shine components tested and working

## Memory Breakdown

### Current (Multi-Instance)
```
Component       | Memory
----------------|-------
Kitty process   | 216 MB
Shine binary    | 14 MB
----------------|-------
Per widget      | 230 MB
4 widgets       | 920 MB
```

### New (Single-Instance)
```
Component              | Memory
-----------------------|-------
Kitty (shared)         | 0 MB (using existing instance)
Panel overhead         | 7.5 MB per panel
Shine binary           | 14 MB per component
-----------------------|-------
Per widget             | 21.5 MB
4 widgets              | 86 MB
Savings                | 90.6%
```

## Implementation Path

### Files to Modify

1. **`pkg/panel/config.go`**
   - Add `ToRemoteControlArgs()` method
   - Converts Config to `kitty @ launch` arguments

2. **`pkg/panel/manager.go`**
   - Add `detectKittySocket()` method
   - Add `LaunchViaRemoteControl()` method
   - Update `Stop()` to use window ID
   - Add socket validation

3. **`pkg/panel/remote.go`**
   - Already has `CloseWindow()`, `FocusWindow()`, etc.
   - No changes needed

### Core Logic

```go
// Detect existing Kitty instance
socket := detectKittySocket() // Returns: unix:/tmp/@mykitty-PID

// Launch panel via remote control
cmd := exec.Command("kitty", "@", "--to", socket, "launch",
    "--type=os-panel",
    "--os-panel", "edge=top",
    "--os-panel", "lines=1",
    "--title=shine-clock",
    "/path/to/shine-clock")

windowID := cmd.Output() // Returns window ID, e.g., "28"
```

### Socket Detection Priority

1. Check `$KITTY_LISTEN_ON` environment variable
2. Search for abstract sockets: `/tmp/@mykitty-*`
3. Search for file sockets: `/tmp/kitty-*.sock`
4. Fallback: Launch dedicated Kitty instance with `listen_on=unix:/tmp/@shine`

## Requirements

User's Kitty must have remote control enabled:

```conf
# ~/.config/kitty/kitty.conf
allow_remote_control socket
listen_on unix:/tmp/@mykitty
```

(Already configured in user's setup ✅)

## Edge Cases Handled

| Case | Solution |
|------|----------|
| No remote-enabled Kitty | Launch dedicated Kitty instance for shine |
| Multiple Kitty instances | Prioritize `$KITTY_LISTEN_ON`, fallback to first found |
| Socket permissions | Use abstract sockets (prefix `@`) for auto-permissions |
| Component crashes | Panel window closes automatically, tracked by window ID |
| Window title conflicts | Use unique titles: `shine-<component>` |

## Testing Checklist

- [ ] Launch 1 panel - verify memory
- [ ] Launch 4 panels - verify memory (target: <100 MB total)
- [ ] Close panel - verify cleanup
- [ ] Toggle visibility - verify hide/show
- [ ] Restart panel - verify reconnection
- [ ] No Kitty instance - verify fallback launch
- [ ] Multiple monitors - verify output-name handling

## Migration Plan

### Phase 1: Add Remote Control Support (1-2 hours)
- Implement `ToRemoteControlArgs()`
- Implement socket detection
- Add integration tests

### Phase 2: Switch Launch Method (1 hour)
- Update Manager to use `kitty @ launch`
- Update Stop to use window IDs
- Test lifecycle

### Phase 3: Fallback & Polish (1 hour)
- Add dedicated Kitty fallback
- Error handling improvements
- Documentation

### Phase 4: Testing & Validation (1 hour)
- Integration tests
- Memory measurements
- Edge case testing

**Total Effort**: ~4-5 hours

## Success Metrics

- [ ] Memory usage <100 MB for 4 widgets (vs 920 MB current)
- [ ] No new Kitty processes for additional widgets
- [ ] All current functionality preserved
- [ ] Clean shutdown/restart cycles

## References

- Research Report: `/home/starbased/dev/projects/shine/docs/llms/inbox/claude-single-instance-research-2025-11-02.md`
- Prototype Code: `/home/starbased/dev/projects/shine/docs/llms/inbox/claude-prototype-remote-control-2025-11-02.go`
- Kitty Remote Control Docs: https://sw.kovidgoyal.net/kitty/remote-control/

## Questions?

This approach is **tested, validated, and ready for implementation**. The memory savings are real and the API is stable.

**Next**: Implement `ToRemoteControlArgs()` and socket detection, then test with real components.
