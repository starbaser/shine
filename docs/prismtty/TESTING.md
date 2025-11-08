# prismctl Phase 1 Testing Guide

## Prerequisites

- Build all binaries: `make build` or manually:
  ```bash
  go build -o bin/prismctl ./cmd/prismctl
  go build -o bin/shine-clock ./cmd/shine-clock
  go build -o bin/shine-sysinfo ./cmd/shine-sysinfo
  go build -o bin/shine-bar ./cmd/shine-bar
  ```

- Add bin/ to PATH for testing:
  ```bash
  export PATH="$PWD/bin:$PATH"
  ```

## Test 1: Basic Launch ✓

**Goal**: Verify prismctl can launch a prism and render correctly without corruption.

**Steps**:
1. Open a new terminal or Kitty panel
2. Run: `prismctl shine-clock`
3. Observe the clock rendering

**Expected Behavior**:
- Clock displays correctly with time updating
- No visual corruption
- Terminal in raw mode (no echo when typing)

**Pass Criteria**:
- ✓ Clock renders cleanly
- ✓ Updates every second
- ✓ No garbled output

**To Stop**: Press `Ctrl+C` or send SIGTERM

---

## Test 2: Hot-Swap Between Prisms via IPC ✓

**Goal**: Verify hot-swap works cleanly without terminal corruption.

**Steps**:

1. **Terminal 1**: Start prismctl with shine-clock
   ```bash
   prismctl shine-clock test-panel
   # Note the PID from output
   ```

2. **Terminal 2**: Find the socket and send swap command
   ```bash
   # List sockets
   ls -la /run/user/$(id -u)/shine/

   # Hot-swap to sysinfo (replace <pid> with actual PID)
   ./scripts/prismctl-ipc.sh prism-test-panel.<pid>.sock swap shine-sysinfo
   ```

3. **Terminal 1**: Observe the swap
   - shine-clock should exit cleanly
   - Screen should clear properly
   - shine-sysinfo should render correctly

4. **Terminal 2**: Swap back to clock
   ```bash
   ./scripts/prismctl-ipc.sh prism-test-panel.<pid>.sock swap shine-clock
   ```

**Expected Behavior**:
- Old prism exits gracefully
- Terminal resets (no leftover artifacts)
- New prism starts immediately
- No visual corruption between swaps
- Swap latency < 100ms

**Pass Criteria**:
- ✓ Both swaps complete without corruption
- ✓ No leftover visual artifacts
- ✓ Terminal state properly reset
- ✓ New prism renders correctly

---

## Test 3: Rapid Swaps (Stress Test) ✓

**Goal**: Verify system remains stable under rapid hot-swaps.

**Steps**:

1. **Terminal 1**: Start prismctl
   ```bash
   prismctl shine-clock rapid-test
   ```

2. **Terminal 2**: Run rapid swap script
   ```bash
   # Create rapid swap test script
   cat > /tmp/rapid-swap.sh << 'EOF'
   #!/bin/bash
   SOCKET=$(ls /run/user/$(id -u)/shine/prism-rapid-test.*.sock | head -1)
   for i in {1..10}; do
       echo "Swap $i..."
       if [ $((i % 2)) -eq 0 ]; then
           ./scripts/prismctl-ipc.sh $(basename $SOCKET) swap shine-clock
       else
           ./scripts/prismctl-ipc.sh $(basename $SOCKET) swap shine-sysinfo
       fi
       sleep 1
   done
   EOF
   chmod +x /tmp/rapid-swap.sh

   # Run the test
   /tmp/rapid-swap.sh
   ```

**Expected Behavior**:
- All 10 swaps complete successfully
- No crashes or hangs
- Terminal remains stable
- No memory leaks
- No zombie processes

**Pass Criteria**:
- ✓ 10/10 swaps succeed
- ✓ System remains responsive
- ✓ No visual corruption
- ✓ prismctl still responsive to IPC

**Verification**:
```bash
# Check for zombies
ps aux | grep defunct

# Check prismctl is still running
ps aux | grep prismctl
```

---

## Test 4: Crash Recovery (SIGKILL) ✓

**Goal**: Verify prismctl recovers gracefully when child is killed forcefully.

**Steps**:

1. **Terminal 1**: Start prismctl
   ```bash
   prismctl shine-clock crash-test
   ```

2. **Terminal 2**: Find and kill child process
   ```bash
   # Find the shine-clock PID (not prismctl)
   ps aux | grep shine-clock

   # Kill it with SIGKILL
   kill -9 <shine-clock-pid>
   ```

3. **Terminal 1**: Observe recovery
   - Terminal should reset cleanly
   - No corruption
   - prismctl should remain running
   - Ready for new swap command

4. **Terminal 2**: Verify prismctl still works
   ```bash
   # Send status command
   SOCKET=$(ls /run/user/$(id -u)/shine/prism-crash-test.*.sock | head -1)
   ./scripts/prismctl-ipc.sh $(basename $SOCKET) status

   # Should show no current prism

   # Swap to new prism
   ./scripts/prismctl-ipc.sh $(basename $SOCKET) swap shine-sysinfo
   ```

**Expected Behavior**:
- prismctl detects child death (SIGCHLD)
- Terminal state is reset immediately
- No leftover corruption
- prismctl remains operational
- Can start new prism via IPC

**Pass Criteria**:
- ✓ Terminal resets cleanly after SIGKILL
- ✓ No visual artifacts
- ✓ prismctl responds to IPC
- ✓ Can swap to new prism

---

## Test 5: Clean Shutdown ✓

**Goal**: Verify graceful shutdown on SIGTERM/panel close.

**Test 5a: SIGTERM**

**Steps**:

1. **Terminal 1**: Start prismctl
   ```bash
   prismctl shine-clock shutdown-test
   ```

2. **Terminal 2**: Send SIGTERM to prismctl
   ```bash
   # Find prismctl PID
   ps aux | grep "prismctl shine-clock"

   # Send SIGTERM (not to child, to prismctl)
   kill <prismctl-pid>
   ```

3. **Terminal 1**: Observe shutdown
   - prismctl receives SIGTERM
   - Forwards SIGTERM to child
   - Waits for child to exit (max 5s)
   - Restores terminal state
   - Exits cleanly

**Expected Behavior**:
- Child receives SIGTERM and exits gracefully
- Terminal restored to original state
- No leftover processes
- Socket cleaned up

**Pass Criteria**:
- ✓ Both prismctl and child exit
- ✓ Terminal restored (canonical mode, cursor visible)
- ✓ No zombie processes
- ✓ Socket file removed

**Verification**:
```bash
# No zombies
ps aux | grep defunct

# Socket removed
ls /run/user/$(id -u)/shine/prism-shutdown-test.*
```

---

**Test 5b: Kitty Panel Close**

**Steps**:

1. Open Kitty
2. Create new panel: `Ctrl+Shift+Enter`
3. In panel: `prismctl shine-clock`
4. Close panel with: `Ctrl+Shift+W`

**Expected Behavior**:
- Kitty sends SIGHUP to prismctl
- prismctl forwards signal to child
- Clean shutdown
- No orphaned processes

**Pass Criteria**:
- ✓ Panel closes immediately
- ✓ No orphaned prismctl or child processes

**Verification**:
```bash
# After closing panel
ps aux | grep prismctl  # Should be none
ps aux | grep shine-clock  # Should be none
```

---

## Test 6: Window Resize (SIGWINCH) - Bonus

**Goal**: Verify SIGWINCH is forwarded correctly to child.

**Steps**:

1. Start prismctl in a terminal
   ```bash
   prismctl shine-bar
   ```

2. Resize the terminal window
3. Observe the bar re-rendering

**Expected Behavior**:
- prismctl receives SIGWINCH
- Forwards to child process group
- Bubble Tea handles resize
- UI re-renders for new size

**Pass Criteria**:
- ✓ UI adapts to new terminal size
- ✓ No corruption during resize
- ✓ Smooth re-render

---

## Test Checklist

Copy and paste to track your testing:

```
Phase 1 MVP Testing:

[ ] Test 1: Basic launch with shine-clock
    [ ] Clock renders correctly
    [ ] No visual corruption
    [ ] Updates smoothly

[ ] Test 2: Hot-swap via IPC
    [ ] clock → sysinfo swap works
    [ ] Terminal resets cleanly
    [ ] sysinfo → clock swap works
    [ ] No leftover artifacts

[ ] Test 3: Rapid swaps (10 in 10s)
    [ ] All swaps complete
    [ ] System remains stable
    [ ] No crashes or hangs
    [ ] No zombie processes

[ ] Test 4: Crash recovery (SIGKILL child)
    [ ] Terminal resets after SIGKILL
    [ ] prismctl remains running
    [ ] IPC still works
    [ ] Can swap to new prism

[ ] Test 5: Clean shutdown
    [ ] SIGTERM to prismctl works
    [ ] Child exits gracefully
    [ ] Terminal restored
    [ ] Socket cleaned up
    [ ] Kitty panel close works
    [ ] No orphaned processes

Bonus:
[ ] Test 6: Window resize (SIGWINCH)
    [ ] Resize events forwarded
    [ ] UI adapts correctly
```

---

## Debugging Tips

**View logs**:
```bash
# prismctl logs to stderr
prismctl shine-clock 2>&1 | tee /tmp/prismctl.log
```

**Check processes**:
```bash
# Process tree
pstree -p $(pgrep prismctl)

# Detailed info
ps -ef | grep prism
```

**Check sockets**:
```bash
ls -la /run/user/$(id -u)/shine/
```

**Test IPC manually**:
```bash
echo '{"action":"status"}' | socat - UNIX-CONNECT:/run/user/$(id -u)/shine/prism-test.12345.sock
```

**Terminal state inspection**:
```bash
stty -a  # View current terminal settings
```

---

## Known Limitations (Phase 1)

- No automatic crash recovery/restart (manual swap required)
- No persistent logging to file
- No health monitoring
- Single prism at a time per prismctl instance

These are expected Phase 3 features.

---

## Success Criteria

Phase 1 is complete when **ALL 5 core tests pass reliably**:

1. ✓ Basic launch works
2. ✓ Hot-swap works without corruption
3. ✓ Rapid swaps (10x) stable
4. ✓ Crash recovery functional
5. ✓ Clean shutdown works

**Ready for Phase 2**: Integration with shinectl and config-driven panel management.
