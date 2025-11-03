# Single-Instance Architecture Investigation Report

**Date**: 2025-11-02
**Goal**: Reduce memory usage by using one Kitty process for all widgets
**Status**: BLOCKED - Kitty's `--single-instance` mode not working as expected

## Problem Statement

Current architecture uses separate Kitty processes per widget:
- Memory usage: ~257MB per widget
- With 2 widgets: 514MB total
- With 10 widgets: 2,570MB total (unacceptable!)

**Target**: Single Kitty process shared by all widgets (~270MB total)

## Implementation Attempt

### Changes Made

1. **Config Updates** (`pkg/config/types.go`):
   - Changed default `SingleInstance: true`
   - Shared socket path: `/tmp/shine.sock`

2. **Panel Config** (`pkg/panel/config.go`):
   - Added `--single-instance` flag
   - Added `--instance-group=shine` flag
   - Added `WindowTitle` field for window matching

3. **Panel Manager** (`pkg/panel/manager.go`):
   - Track shared Kitty instance
   - First launch creates process, subsequent launches should join
   - Window-based remote control via titles

4. **Remote Control** (`pkg/panel/remote.go`):
   - Added `CloseWindow(title)` method
   - Added `FocusWindow(title)` method
   - Added `ListWindows()` method

5. **Component Apps**:
   - Set window titles via ANSI escape: `\033]0;shine-chat\007`

6. **shinectl**:
   - Use fixed socket path `/tmp/shine.sock`
   - Window-based commands instead of PID-based

### Test Results

**FAILED**: Still creating 2 separate Kitty processes

```bash
Kitty processes: 2 (expected: 1)
Total memory: 462-467 MB (expected: ~270 MB)
```

### Root Cause Analysis

The `--single-instance` flag with `--instance-group=shine` is **not working**.

**Observations**:
1. Each `kitten panel` invocation creates a new Kitty process
2. Both processes have different PIDs (546592, 546593)
3. Adding 500ms delay between launches: no effect
4. Removing `listen_on` option: no effect
5. Kitty version 0.43.1 (recent, should support the feature)

**Possible Reasons**:
1. `--single-instance` for `kitten panel` works differently than documented
2. Bug in Kitty's single-instance implementation for panels
3. Layer-shell panels might not support single-instance mode
4. Instance detection mechanism incompatible with our launch method

### Commands Generated

```bash
# First panel (chat)
kitten panel \
  --edge=bottom \
  --lines=10 \
  --single-instance \
  --instance-group=shine \
  /path/to/shine-chat

# Second panel (bar) - should join first instance
kitten panel \
  --edge=top \
  --lines=1 \
  --single-instance \
  --instance-group=shine \
  /path/to/shine-bar
```

**Expected**: Second command creates window in first Kitty process
**Actual**: Second command creates entirely new Kitty process

## Alternative Approaches Considered

### Option 1: Direct Kitty Management
- Launch `kitty` directly (not `kitten panel`)
- Use `kitty @ launch` for additional windows
- **Problem**: `kitty @ launch` doesn't support layer-shell properties

### Option 2: Kitty Config File
- Pre-configure Kitty with layer-shell settings per window
- **Problem**: Layer-shell requires runtime window properties

### Option 3: Manual Process Management
- Create wrapper that launches one Kitty with special config
- Manually manage window creation via remote control
- **Problem**: Complex, fragile, loses `kitten panel` benefits

### Option 4: Accept Multiple Processes, Optimize Elsewhere
- Keep current architecture
- Optimize memory within each process
- Share more resources (fonts, libs) via system config
- **Impact**: Limited savings, doesn't solve core issue

## Recommended Next Steps

### Immediate: Report to Kitty Project

File issue on Kitty GitHub with:
- Minimal reproduction case
- Expected vs actual behavior
- Version info
- Ask if `--single-instance` is supposed to work with `kitten panel`

### Short-term: Hybrid Approach

1. **Keep improvements made**:
   - Shared socket path (when it works)
   - Window title management
   - Remote control via window matching

2. **Document limitation**:
   - Current memory usage per widget
   - Recommend limiting widget count
   - Note expected future improvement when Kitty fixes issue

3. **Add configuration option**:
   - Let users choose single vs multi-instance
   - Default to current behavior (works reliably)
   - Experimental single-instance mode for future

### Long-term: Custom Compositor Integration

If Kitty doesn't support this use case:
- Consider Wayland-native TUI framework
- Or lightweight terminal emulator (foot, alacritty) with layer-shell patches
- Or custom layer-shell implementation

## Files Modified

- `/home/starbased/dev/projects/shine/pkg/config/types.go`
- `/home/starbased/dev/projects/shine/pkg/panel/config.go`
- `/home/starbased/dev/projects/shine/pkg/panel/manager.go`
- `/home/starbased/dev/projects/shine/pkg/panel/remote.go`
- `/home/starbased/dev/projects/shine/cmd/shine/main.go`
- `/home/starbased/dev/projects/shine/cmd/shinectl/main.go`
- `/home/starbased/dev/projects/shine/cmd/shine-chat/main.go`
- `/home/starbased/dev/projects/shine/cmd/shine-bar/main.go`

## Code Status

All changes compile and run correctly, but don't achieve the memory savings goal because Kitty's single-instance mode doesn't work as expected.

**The code is ready for when Kitty's single-instance mode works properly.**

## Conclusion

The refactor is **technically complete** but **functionally blocked** by Kitty's behavior.

The implementation correctly:
- Sets single-instance flags
- Manages shared socket paths
- Implements window-based control
- Sets window titles

But Kitty still creates separate processes.

**Action Required**: Determine if this is:
1. A Kitty bug to report
2. A misunderstanding of the feature
3. An unsupported use case

Until resolved, the memory usage remains at ~460MB for 2 widgets instead of the target ~270MB.
