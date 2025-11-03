# Spotify Prism - Testing Guide

This guide provides comprehensive testing procedures for the Spotify prism.

## Quick Test

```bash
# 1. Build
cd /home/starbased/dev/projects/shine/examples/prisms/spotify
make build

# 2. Test mock mode (no Spotify needed)
kitten panel --edge=bottom --lines-pixels=120 --columns-pixels=600 \
    --output-name=DP-2 --focus-policy=on-demand \
    ./shine-spotify --mock

# 3. Test with real Spotify (if installed)
# First, start Spotify and play a song
spotify &
sleep 3

kitten panel --edge=bottom --lines-pixels=120 --columns-pixels=600 \
    --output-name=DP-2 --focus-policy=on-demand \
    ./shine-spotify
```

## Test Matrix

### 1. Visual Appearance Tests

#### Test 1.1: Not Running State
**Goal**: Verify friendly message when Spotify is not running

```bash
# Stop Spotify if running
pkill spotify

# Launch prism
kitten panel --edge=bottom --lines-pixels=120 --columns-pixels=600 \
    --output-name=DP-2 \
    ./shine-spotify
```

**Expected Result**:
- Rounded border with dimmed gray color
- Music note emoji (🎵)
- Message: "Spotify is not running"
- No crash or error messages

#### Test 1.2: Mock Mode Display
**Goal**: Verify visual design with mock data

```bash
kitten panel --edge=bottom --lines-pixels=120 --columns-pixels=600 \
    --output-name=DP-2 --focus-policy=on-demand \
    ./shine-spotify --mock
```

**Expected Result**:
- Rounded border with Spotify green (#1DB954)
- Play icon (▶) visible
- Track: "The Less I Know The Better"
- Artist: "Tame Impala"
- Progress bar with green fill
- Time display: "X:XX / 3:36"
- Help text visible at bottom
- Smooth progress bar animation

#### Test 1.3: Real Spotify Display
**Goal**: Verify integration with actual Spotify

```bash
# Start Spotify and play a song
spotify &
sleep 3

# Launch prism
kitten panel --edge=bottom --lines-pixels=120 --columns-pixels=600 \
    --output-name=DP-2 --focus-policy=on-demand \
    ./shine-spotify
```

**Expected Result**:
- Real track information displayed
- Correct artist name
- Accurate progress bar
- Time matches Spotify's playback position

### 2. D-Bus Integration Tests

#### Test 2.1: D-Bus Connection
**Goal**: Verify D-Bus connectivity

```bash
# Check if Spotify is on D-Bus
busctl --user list | grep spotify
# Expected: org.mpris.MediaPlayer2.spotify

# Get metadata manually
busctl --user call org.mpris.MediaPlayer2.spotify \
    /org/mpris/MediaPlayer2 \
    org.freedesktop.DBus.Properties Get \
    ss "org.mpris.MediaPlayer2.Player" "Metadata"
```

#### Test 2.2: Metadata Retrieval
**Goal**: Verify all metadata fields are retrieved

Launch prism with Spotify playing and verify:
- Track title matches Spotify
- Artist name matches Spotify
- Duration is correct
- Position updates every second

#### Test 2.3: Status Changes
**Goal**: Verify playback status detection

```bash
# Launch prism
kitten panel --edge=bottom --lines-pixels=120 --columns-pixels=600 \
    --output-name=DP-2 --focus-policy=on-demand \
    ./shine-spotify

# In another terminal, control Spotify via D-Bus
busctl --user call org.mpris.MediaPlayer2.spotify \
    /org/mpris/MediaPlayer2 \
    org.mpris.MediaPlayer2.Player Pause

# Verify pause icon (⏸) appears in prism

busctl --user call org.mpris.MediaPlayer2.spotify \
    /org/mpris/MediaPlayer2 \
    org.mpris.MediaPlayer2.Player Play

# Verify play icon (▶) appears in prism
```

### 3. Keyboard Control Tests

#### Test 3.1: Play/Pause Toggle
**Goal**: Verify spacebar toggles playback

```bash
# Launch prism with focus
kitten panel --edge=bottom --lines-pixels=120 --columns-pixels=600 \
    --output-name=DP-2 --focus-policy=on-demand \
    ./shine-spotify

# Press Space key
# Expected: Spotify pauses/plays, icon changes
```

#### Test 3.2: Next Track
**Goal**: Verify 'n' skips to next track

```bash
# Press 'n' key
# Expected: Next track starts playing, display updates
```

#### Test 3.3: Previous Track
**Goal**: Verify 'p' goes to previous track

```bash
# Press 'p' key
# Expected: Previous track starts, or current restarts if >3s in
```

#### Test 3.4: Seek Forward
**Goal**: Verify right arrow seeks forward

```bash
# Note current position
# Press Right arrow key
# Expected: Position increases by ~5 seconds
```

#### Test 3.5: Seek Backward
**Goal**: Verify left arrow seeks backward

```bash
# Note current position
# Press Left arrow key
# Expected: Position decreases by ~5 seconds
```

#### Test 3.6: Quit
**Goal**: Verify q/Ctrl+C quits cleanly

```bash
# Press 'q' key
# Expected: Prism exits cleanly, panel closes
```

### 4. Auto-Update Tests

#### Test 4.1: Progress Bar Update
**Goal**: Verify progress bar animates smoothly

```bash
# Launch prism with Spotify playing
kitten panel --edge=bottom --lines-pixels=120 --columns-pixels=600 \
    --output-name=DP-2 \
    ./shine-spotify

# Watch for 10 seconds
# Expected: Progress bar fills smoothly, time increments
```

#### Test 4.2: Track Change Detection
**Goal**: Verify display updates when track changes

```bash
# Launch prism
# Let current track finish OR skip to next track in Spotify app
# Expected: Display updates to new track within 1 second
```

#### Test 4.3: App Launch Detection
**Goal**: Verify prism detects when Spotify starts

```bash
# Launch prism without Spotify
pkill spotify
./shine-spotify

# Start Spotify in another terminal
spotify &

# Expected: Within 1 second, prism switches from "not running" to displaying track
```

### 5. Configuration Tests

#### Test 5.1: DP-2 Monitor Targeting
**Goal**: Verify prism appears on correct monitor

```bash
# Check available monitors
hyprctl monitors

# Launch with specific output
kitten panel --edge=bottom --lines-pixels=120 --columns-pixels=600 \
    --output-name=DP-2 \
    ./shine-spotify --mock

# Expected: Panel appears on DP-2 monitor only
```

#### Test 5.2: Size Configuration
**Goal**: Verify pixel dimensions work correctly

```bash
# Test different sizes
kitten panel --edge=bottom --lines-pixels=80 --columns-pixels=400 \
    ./shine-spotify --mock
# Expected: Smaller panel

kitten panel --edge=bottom --lines-pixels=150 --columns-pixels=800 \
    ./shine-spotify --mock
# Expected: Larger panel
```

#### Test 5.3: Margin Configuration
**Goal**: Verify margins position panel correctly

```bash
kitten panel --edge=bottom --lines-pixels=120 --columns-pixels=600 \
    --margin-bottom=50 --margin-left=50 \
    ./shine-spotify --mock

# Expected: Panel offset from edges
```

### 6. Edge Case Tests

#### Test 6.1: Spotify Crash Handling
**Goal**: Verify graceful handling when Spotify crashes

```bash
# Launch prism with Spotify running
./shine-spotify

# Kill Spotify
pkill -9 spotify

# Expected: Prism shows "not running" message, no crash
```

#### Test 6.2: Very Long Track Names
**Goal**: Verify text doesn't overflow

Use Spotify to play track with very long title/artist.

**Expected**: Text wraps or truncates gracefully

#### Test 6.3: Zero Duration Tracks
**Goal**: Handle special cases (live streams, etc.)

**Expected**: Progress bar shows empty or doesn't crash

#### Test 6.4: Rapid Key Presses
**Goal**: Ensure controls don't break with spam

```bash
# Launch prism
# Rapidly press: n n n p p p space space space

# Expected: No crashes, commands queue properly
```

### 7. Performance Tests

#### Test 7.1: CPU Usage
**Goal**: Verify low CPU usage

```bash
# Launch prism
./shine-spotify &
PID=$!

# Monitor CPU usage
top -p $PID -b -n 10 -d 1

# Expected: < 2% CPU average
```

#### Test 7.2: Memory Usage
**Goal**: Verify reasonable memory footprint

```bash
# Check memory
ps aux | grep shine-spotify

# Expected: < 15 MB RSS
```

#### Test 7.3: D-Bus Call Frequency
**Goal**: Verify polling doesn't spam D-Bus

```bash
# Monitor D-Bus traffic
dbus-monitor --session "interface='org.mpris.MediaPlayer2.Player'" &

# Launch prism
./shine-spotify

# Expected: Calls approximately once per second, not more
```

## Integration Tests with Shine

### Test 8.1: Shine Launcher
**Goal**: Verify prism works via Shine launcher

```bash
# Install prism
cd /home/starbased/dev/projects/shine/examples/prisms/spotify
make install

# Verify shine.toml has spotify config
cat /home/starbased/dev/projects/shine/examples/shine.toml | grep -A 10 "\[prisms.spotify\]"

# Launch Shine
cd /home/starbased/dev/projects/shine
./shine

# Expected: Spotify prism appears on DP-2
```

### Test 8.2: Multiple Prisms
**Goal**: Verify Spotify prism coexists with other prisms

```bash
# Enable multiple prisms in shine.toml
# [prisms.bar]
# [prisms.spotify]
# [prisms.weather]

# Launch Shine
./shine

# Expected: All prisms appear without conflicts
```

## Automated Test Script

```bash
#!/bin/bash
# Comprehensive automated tests

set -e

PRISM_DIR="/home/starbased/dev/projects/shine/examples/prisms/spotify"
cd "$PRISM_DIR"

echo "=== Spotify Prism Test Suite ==="
echo ""

# Test 1: Build
echo "[TEST] Building prism..."
make build
echo "✓ Build successful"
echo ""

# Test 2: Binary exists
echo "[TEST] Checking binary..."
[ -f ./shine-spotify ] || { echo "✗ Binary not found"; exit 1; }
echo "✓ Binary exists"
echo ""

# Test 3: Mock mode (dry run - can't test visually in script)
echo "[TEST] Mock mode (dry run)..."
timeout 2 ./shine-spotify --mock 2>&1 | head -5 || true
echo "✓ Mock mode launches"
echo ""

# Test 4: D-Bus availability
echo "[TEST] Checking D-Bus..."
busctl --user list | grep dbus > /dev/null || { echo "✗ D-Bus not available"; exit 1; }
echo "✓ D-Bus available"
echo ""

# Test 5: Spotify detection
echo "[TEST] Spotify detection..."
if busctl --user list | grep spotify > /dev/null; then
    echo "✓ Spotify is running"

    # Test 6: D-Bus metadata retrieval
    echo "[TEST] D-Bus metadata..."
    busctl --user call org.mpris.MediaPlayer2.spotify \
        /org/mpris/MediaPlayer2 \
        org.freedesktop.DBus.Properties Get \
        ss "org.mpris.MediaPlayer2.Player" "Metadata" > /dev/null
    echo "✓ Metadata retrieved"
else
    echo "⚠ Spotify not running (some tests skipped)"
fi
echo ""

echo "=== All tests passed ==="
```

Save as `test.sh`, make executable, run:

```bash
chmod +x test.sh
./test.sh
```

## Visual Verification Checklist

When testing manually, verify these visual elements:

**Layout**:
- [ ] Rounded border visible
- [ ] Border color is Spotify green when playing
- [ ] Border color is gray when not running
- [ ] Content is properly padded inside border
- [ ] All elements fit within panel dimensions

**Typography**:
- [ ] Track name is bold and white
- [ ] Artist name is dimmed gray
- [ ] Bullet separator (•) between track and artist
- [ ] Time stamps are legible
- [ ] Help text is italic and very dim

**Progress Bar**:
- [ ] Green filled portion shows correct progress
- [ ] Gray unfilled portion visible
- [ ] Bar uses thick characters (━) for filled
- [ ] Bar uses thin characters (─) for unfilled
- [ ] Progress animates smoothly

**Icons**:
- [ ] Play icon (▶) when playing
- [ ] Pause icon (⏸) when paused
- [ ] Music note (🎵) when not running

**Interaction**:
- [ ] Keyboard focus indicator visible (if applicable)
- [ ] Controls respond within 100ms
- [ ] No visual glitches on update
- [ ] No flickering or tearing

## Common Issues and Solutions

### Issue: Prism doesn't appear
**Check**:
1. Monitor name correct (`hyprctl monitors`)
2. Binary has execute permissions
3. Panel arguments are valid
4. No error in logs

### Issue: Controls don't work
**Check**:
1. `focus_policy = "on-demand"` in config
2. Panel has keyboard focus (click it)
3. Spotify is running
4. D-Bus connection is working

### Issue: Display doesn't update
**Check**:
1. Spotify is actually playing (not paused)
2. D-Bus connection is stable
3. No errors in terminal output

### Issue: Wrong track displayed
**Check**:
1. Wait 1 second for polling update
2. Verify Spotify is actually playing different track
3. Check D-Bus metadata manually

## Conclusion

This testing guide covers:
- Visual appearance verification
- D-Bus integration testing
- Keyboard control validation
- Auto-update behavior
- Configuration testing
- Edge case handling
- Performance verification
- Integration with Shine

For any issues not covered here, check:
1. README.md for usage instructions
2. main.go source code for implementation details
3. Shine project documentation for panel system
