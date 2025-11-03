# Shine Status Bar - Visual Testing Guide

**Reference Document:** [VISUAL_STATE_SPECIFICATION.md](./VISUAL_STATE_SPECIFICATION.md)
**Branch:** phase-2-statusbar
**Date:** 2025-11-02

---

## Quick Testing Workflow

### 1. Pre-Test Setup

```bash
# Ensure you're on the correct branch
cd /home/starbased/dev/projects/shine
git status  # Should show phase-2-statusbar

# Build the project
go build -o shine-bar ./cmd/shine-bar
go build -o shine ./cmd/shine

# Create test configuration
mkdir -p ~/.config/shine
cp examples/shine.toml ~/.config/shine/shine.toml

# Verify Kitty has remote control enabled
grep -E "allow_remote_control|listen_on" ~/.config/kitty/kitty.conf
# Should show:
#   allow_remote_control yes
#   listen_on unix:/tmp/@mykitty
```

### 2. Launch Status Bar

**Method A: Via Shine Launcher (Recommended)**
```bash
./shine
# Should output:
#   ✨ Shine - Hyprland Layer Shell TUI Toolkit
#   Configuration: /home/starbased/.config/shine/shine.toml
#   Launching status bar...
#   ✓ Status bar launched (Window ID: ...)
```

**Method B: Direct Launch (For Debugging)**
```bash
./shine-bar
# Note: Will not position correctly without panel wrapper
# Only use for testing TUI logic, not visual appearance
```

### 3. Visual Inspection Checklist

Use this checklist while viewing the running status bar:

**Basic Appearance:**
- [ ] Bar visible at top of screen
- [ ] Full width (edge to edge)
- [ ] Height approximately 30 pixels (thin bar)
- [ ] Black background
- [ ] No gaps or margins at screen edges

**Workspace Section (Left):**
- [ ] Workspace numbers visible (e.g., "1 2 3 4")
- [ ] Numbers have white text (inactive)
- [ ] Current workspace in bright cyan
- [ ] Current workspace text is bold
- [ ] Proper spacing between numbers

**Clock Section (Right):**
- [ ] Time displayed in HH:MM:SS format
- [ ] Clock text in bright magenta
- [ ] Clock text is bold
- [ ] Clock aligned to right edge
- [ ] Seconds update every second

**Color Verification:**
```
Expected colors (may vary by terminal theme):
- Active workspace: #00FFFF or similar (bright cyan)
- Inactive workspace: #FFFFFF (white)
- Clock: #FF00FF or similar (bright magenta)
- Background: #000000 (black)
```

### 4. Functional Testing

**Clock Updates:**
```bash
# Watch for 10 seconds
# Verify seconds increment correctly
# Verify time matches system clock: date +%H:%M:%S
```

**Workspace Switching:**
```bash
# Switch to different workspace
hyprctl dispatch workspace 2

# Verify:
# - Workspace "2" now bright cyan and bold
# - Previously active workspace now white
```

**Keyboard Input:**
```bash
# If you can access the bar's terminal:
# Press ESC → bar should close

# Relaunch if needed: ./shine
```

**Resize Handling:**
```bash
# Resize terminal window (if possible in panel mode)
# Verify:
# - Clock remains right-aligned
# - Bar adapts to new width
# - No content clipping
```

### 5. Screenshot Capture

Capture screenshots for documentation and comparison:

```bash
# Create screenshots directory
mkdir -p ~/Pictures/Screenshots/shine-bar

# Capture full screen (includes status bar)
grim ~/Pictures/Screenshots/shine-bar/full-screen-$(date +%Y%m%d-%H%M%S).png

# Capture specific region (status bar only)
# First, get bar position (typically 0,0 to screen_width,30)
# Then capture:
grim -g "0,0 1920x30" ~/Pictures/Screenshots/shine-bar/bar-only-$(date +%Y%m%d-%H%M%S).png

# For comparison, capture with different workspace states:
# 1. Single workspace active
grim ~/Pictures/Screenshots/shine-bar/single-workspace.png

# 2. Multiple workspaces, workspace 2 active
hyprctl dispatch workspace 2
sleep 0.5
grim ~/Pictures/Screenshots/shine-bar/multi-workspace-active-2.png

# 3. Many workspaces (create 10)
for i in {1..10}; do hyprctl dispatch workspace $i; done
hyprctl dispatch workspace 5
sleep 0.5
grim ~/Pictures/Screenshots/shine-bar/many-workspaces.png
```

### 6. Troubleshooting Common Issues

**Issue: Bar not visible**
```bash
# Check if bar is running
pgrep shine-bar
# If not found: check shine launcher output for errors

# Check Kitty windows
kitty @ ls
# Should show window with title "shine-bar"
```

**Issue: Wrong position/size**
```bash
# Verify configuration
cat ~/.config/shine/shine.toml | grep -A 10 "\[bar\]"

# Check what arguments were passed to Kitty
# Look for launch logs in shine output
```

**Issue: Wrong colors**
```bash
# Check terminal color scheme
kitty @ get-colors

# Test ANSI color output directly
echo -e "\033[38;5;14mBright Cyan\033[0m"
echo -e "\033[38;5;13mBright Magenta\033[0m"
echo -e "\033[38;5;15mWhite\033[0m"
```

**Issue: Clock not updating**
```bash
# Check if bar process is running
pgrep -a shine-bar

# Check for errors in stderr
# (shine launcher shows stderr output)
```

**Issue: Workspaces not showing**
```bash
# Test Hyprland connection
hyprctl workspaces -j
hyprctl activeworkspace -j

# Verify Hyprland is running
pgrep -a Hyprland
```

### 7. Comparison with Specification

After testing, compare actual behavior with specification:

**Reference:** See [VISUAL_STATE_SPECIFICATION.md](./VISUAL_STATE_SPECIFICATION.md)

**Key Sections to Verify:**
- Section 3: Visual Components
- Section 4: Color Scheme & Styling
- Section 5: Dynamic Behavior
- Section 9: Verification Checklist (page ~15)

**Create Comparison Report:**
```markdown
# Visual State Comparison Report

**Date:** [Current Date]
**Branch:** phase-2-statusbar
**Tester:** [Your Name]

## Matches Specification

- [x] Item that matches
- [x] Another matching item

## Deviations from Specification

- [ ] Item that differs: [describe difference]
- [ ] Another deviation: [describe difference]

## Screenshots

- Full screen: [path/to/screenshot]
- Bar only: [path/to/screenshot]
- Workspace switching: [path/to/screenshot]

## Notes

[Any additional observations]
```

### 8. Performance Testing

**CPU Usage:**
```bash
# Monitor CPU usage while bar is running
top -p $(pgrep shine-bar)

# Expected: <1% CPU when idle
# Expected: ~1-2% CPU during updates
```

**Memory Usage:**
```bash
# Check memory consumption
ps aux | grep shine-bar

# Expected: <50MB RSS
```

**Long-Running Stability:**
```bash
# Leave bar running for extended period
# Check for:
# - Memory leaks (increasing RSS over time)
# - CPU usage creep
# - Rendering glitches

# Monitor for 1 hour:
while true; do
  date
  ps aux | grep shine-bar | grep -v grep
  sleep 300  # Every 5 minutes
done
```

---

## Testing Matrix

### Environment Variables to Test

| Variable | Value | Expected Result |
|----------|-------|-----------------|
| Normal | Default | Bar works as specified |
| Different monitor | `output_name = "HDMI-1"` | Bar appears on specified monitor |
| No config file | (delete config) | Uses default (bar disabled) |
| Invalid config | (malformed TOML) | Error message, uses default |

### Configuration Variants

Test different configurations by modifying `~/.config/shine/shine.toml`:

**Test 1: Different Edge**
```toml
[bar]
edge = "bottom"  # Instead of "top"
```
Expected: Bar appears at bottom of screen

**Test 2: Different Height**
```toml
[bar]
lines_pixels = 50  # Instead of 30
```
Expected: Taller bar (50px height)

**Test 3: With Margins**
```toml
[bar]
margin_top = 10
margin_left = 100
margin_right = 100
```
Expected: Bar has 10px top margin, 100px side margins

**Test 4: Different Focus Policy**
```toml
[bar]
focus_policy = "on-demand"  # Instead of "not-allowed"
```
Expected: Bar can receive keyboard focus

---

## Automated Visual Comparison (Future)

For future iterations, consider implementing automated screenshot comparison:

```bash
# Capture baseline screenshot
grim -g "0,0 1920x30" baseline.png

# After code changes, capture new screenshot
grim -g "0,0 1920x30" current.png

# Compare using ImageMagick
compare baseline.png current.png diff.png

# Generate diff metrics
compare -metric RMSE baseline.png current.png null: 2>&1
# Low value = images are similar
# High value = significant differences
```

---

## Documentation Update Workflow

After testing, update documentation if needed:

1. **If specification is accurate:**
   - Mark document as "VERIFIED" with date
   - Add screenshots to documentation

2. **If deviations found:**
   - Create issue tracking deviation
   - Update specification if implementation is correct
   - Or create bug report if implementation is wrong

3. **For new features:**
   - Update specification document
   - Add new test cases to this guide
   - Capture screenshots of new features

---

## Quick Reference Commands

```bash
# Build
go build -o shine-bar ./cmd/shine-bar && go build -o shine ./cmd/shine

# Launch
./shine

# Stop
pkill shine-bar
# OR press ESC in bar window
# OR Ctrl+C in shine launcher

# View config
cat ~/.config/shine/shine.toml

# Test Hyprland
hyprctl workspaces -j
hyprctl activeworkspace -j

# Switch workspace
hyprctl dispatch workspace 2

# Screenshot
grim ~/Pictures/Screenshots/shine-test-$(date +%Y%m%d-%H%M%S).png

# Check processes
pgrep -a shine
kitty @ ls

# View colors
kitty @ get-colors
```

---

## Contact & Support

**Issues:** Create issue in Shine repository
**Questions:** Reference VISUAL_STATE_SPECIFICATION.md
**Screenshots:** Store in `~/Pictures/Screenshots/shine-bar/`
