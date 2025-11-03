# Shine Status Bar - Visual State Specification

**Document Version:** 1.0
**Branch:** phase-2-statusbar
**Date:** 2025-11-02
**Purpose:** Reference document for visual comparison testing and debugging

---

## Table of Contents

1. [Executive Summary](#executive-summary)
2. [Overall Layout & Position](#overall-layout--position)
3. [Visual Components](#visual-components)
4. [Color Scheme & Styling](#color-scheme--styling)
5. [Dynamic Behavior](#dynamic-behavior)
6. [Technical Implementation Details](#technical-implementation-details)
7. [Configuration Options](#configuration-options)
8. [Expected Visual Appearance](#expected-visual-appearance)
9. [Verification Checklist](#verification-checklist)

---

## Executive Summary

The Shine status bar (`shine-bar`) is a minimal, single-line panel displaying Hyprland workspace indicators and a real-time clock. It uses Kitty's layer shell integration to render as a persistent overlay at the top of the screen.

**Key Characteristics:**
- Horizontal bar spanning full screen width
- 30 pixels in height (default configuration)
- Positioned at top edge with no margins
- Black background with high-contrast colored text
- Updates every second for clock and workspace state
- No keyboard focus (FocusNotAllowed policy)
- Persistent visibility (does not hide on focus loss)

---

## Overall Layout & Position

### Window Dimensions

**Default Configuration (from `examples/shine.toml`):**
```toml
[bar]
edge = "top"
lines_pixels = 30
margin_top = 0
margin_left = 0
margin_right = 0
```

**Calculated Dimensions:**
- **Height:** 30 pixels (fixed, via `lines_pixels`)
- **Width:** Full screen width (spans entire monitor)
- **Position:** Top edge of screen, aligned to left edge

**Fallback Configuration (when no config file exists):**
```go
// From pkg/config/types.go NewDefaultConfig()
// Bar component is NOT enabled by default
// Only chat panel is enabled in default config
```

### Anchor Points and Screen Placement

**Edge Placement:** `EdgeTop` (value 0 in Edge enum)
```go
// From pkg/panel/config.go
type Edge int
const (
    EdgeTop Edge = iota  // = 0
    // ...
)
```

**Layer Shell Type:** `LayerShellPanel` (bottom layer)
```go
// From pkg/panel/config.go
const (
    LayerShellNone LayerType = iota
    LayerShellBackground  // "background"
    LayerShellPanel       // "bottom" - DEFAULT for bar
    LayerShellTop         // "top"
    LayerShellOverlay     // "overlay"
)
```

**Margins:** All set to 0 (flush with screen edges)
- `margin_top: 0`
- `margin_left: 0`
- `margin_right: 0`
- `margin_bottom: 0` (N/A for top edge)

### Layer/Stacking Behavior

**Wayland Layer:** Panel layer (renders above background, below overlays)

**Exclusive Zone:** -1 (auto-calculated)
```go
// From pkg/panel/config.go NewConfig()
ExclusiveZone: -1  // Auto
```

**Focus Policy:** `FocusNotAllowed` (status bar doesn't accept keyboard input)
```toml
# From examples/shine.toml
focus_policy = "not-allowed"
```

**Visibility:** Persistent
```toml
hide_on_focus_loss = false
```

**Z-Order:** Above desktop background, below focused windows and overlays

---

## Visual Components

### Component Layout (Horizontal)

The status bar has a three-part horizontal layout:

```
┌─────────────────────────────────────────────────────────────┐
│ [Workspaces]          [Spacer]                    [Clock]   │
└─────────────────────────────────────────────────────────────┘
```

**Layout Code:**
```go
// From cmd/shine-bar/main.go View()
return lipgloss.JoinHorizontal(
    lipgloss.Top,
    workspacesView,    // Left: workspace indicators
    spacer,            // Middle: flexible spacing
    clockView,         // Right: time display
)
```

### Left Section: Workspaces

**Data Source:** Hyprland workspace information via `hyprctl`
```bash
hyprctl workspaces -j      # All workspaces
hyprctl activeworkspace -j # Active workspace
```

**Visual Elements:**
- Each workspace rendered as its numeric ID (e.g., "1", "2", "3")
- Workspaces displayed in order received from Hyprland
- No separator between workspace indicators
- Active workspace has distinct styling (bright cyan on black)
- Inactive workspaces have muted styling (white on black)

**Workspace Indicator Structure:**
```go
// From cmd/shine-bar/main.go
for _, ws := range m.workspaces {
    wsLabel := fmt.Sprintf("%d", ws.ID)
    if ws.ID == m.activeWorkspaceID {
        workspaceStrs = append(workspaceStrs, activeStyle.Render(wsLabel))
    } else {
        workspaceStrs = append(workspaceStrs, inactiveStyle.Render(wsLabel))
    }
}
workspacesView := strings.Join(workspaceStrs, "")
```

**Padding:** Each workspace indicator has 1 character padding on left and right
```go
Padding(0, 1)  // top/bottom=0, left/right=1
```

**Example rendering (4 workspaces, workspace 2 active):**
```
 1  2  3  4
    ^^^ active (bright cyan, bold)
 ^  ^  ^  inactive (white)
```

### Middle Section: Spacer

**Purpose:** Push clock to right edge while filling background
```go
// From cmd/shine-bar/main.go
contentWidth := lipgloss.Width(workspacesView) + lipgloss.Width(clockView)
spacerWidth := m.width - contentWidth
if spacerWidth < 0 {
    spacerWidth = 0
}
spacer := strings.Repeat(" ", spacerWidth)
```

**Behavior:**
- Dynamically calculated based on terminal width
- Fills with spaces (inherits background color from terminal)
- Adjusts on window resize events

### Right Section: Clock

**Format:** 24-hour time with seconds
```go
// From cmd/shine-bar/main.go
clockView := clockStyle.Render(m.currentTime.Format("15:04:05"))
```

**Example:** `14:23:47`

**Padding:** 1 character padding on left and right
```go
Padding(0, 1)  // top/bottom=0, left/right=1
```

**Update Frequency:** 1 second
```go
// From cmd/shine-bar/main.go tickCmd()
tea.Tick(time.Second, func(t time.Time) tea.Msg {
    return tickMsg(t)
})
```

---

## Color Scheme & Styling

### Design Philosophy

**High Contrast for Thin Panels:**
```go
// From cmd/shine-bar/main.go View()
// Styles with high contrast for visibility in thin panels
// Use bright colors on dark/transparent background
```

### Color Palette

Using ANSI 256-color codes:

| Element | Foreground | Background | Attributes |
|---------|------------|------------|------------|
| Active Workspace | Color 14 (Bright Cyan) | Color 0 (Black) | Bold |
| Inactive Workspace | Color 15 (White) | Color 0 (Black) | Normal |
| Clock | Color 13 (Bright Magenta) | Color 0 (Black) | Bold |

**Color Definitions:**
```go
activeStyle := lipgloss.NewStyle().
    Foreground(lipgloss.Color("14")).  // Bright cyan
    Background(lipgloss.Color("0")).   // Black background
    Bold(true).
    Padding(0, 1)

inactiveStyle := lipgloss.NewStyle().
    Foreground(lipgloss.Color("15")).  // White
    Background(lipgloss.Color("0")).   // Black background
    Padding(0, 1)

clockStyle := lipgloss.NewStyle().
    Foreground(lipgloss.Color("13")).  // Bright magenta
    Background(lipgloss.Color("0")).   // Black background
    Bold(true).
    Padding(0, 1)
```

### ANSI Color Reference

**Standard 16-Color Palette (assuming typical terminal colors):**
- Color 0: Black (#000000)
- Color 13: Bright Magenta (#FF00FF or similar)
- Color 14: Bright Cyan (#00FFFF or similar)
- Color 15: White (#FFFFFF)

**Note:** Actual RGB values depend on terminal color scheme configuration.

### Typography

**Font:** Inherited from Kitty terminal configuration
**Weight:** Bold for active workspace and clock, normal for inactive workspaces
**Size:** Determined by Kitty's font size and 30px panel height

### Spacing and Padding

**Element Padding:** 1 character horizontal, 0 vertical
```go
Padding(0, 1)  // (top/bottom, left/right)
```

**Visual Spacing Breakdown:**
```
 1  2  3  14:23:47
^  ^  ^  ^        ^
│  │  │  │        └─ Right padding (1 char)
│  │  │  └────────── Left padding (1 char)
│  │  └───────────── Right padding + Left padding (2 chars between)
│  └──────────────── Right padding (1 char)
└─────────────────── Left padding (1 char)
```

---

## Dynamic Behavior

### Update Intervals

**Clock Updates:** Every 1 second
```go
// From cmd/shine-bar/main.go
func tickCmd() tea.Cmd {
    return tea.Tick(time.Second, func(t time.Time) tea.Msg {
        return tickMsg(t)
    })
}
```

**Workspace Updates:** Every 1 second (refreshed with clock)
```go
// From cmd/shine-bar/main.go Update()
case tickMsg:
    m.currentTime = time.Time(msg)
    return m, tea.Batch(
        tickCmd(),
        refreshWorkspacesCmd(),  // Query Hyprland
    )
```

**Initial State:**
```go
// From cmd/shine-bar/main.go initialModel()
width: 80,  // Initial width (updated on first WindowSizeMsg)
```

### Interactive Elements

**Keyboard Input:**
- **ESC key:** Quit application (exits status bar)
```go
case tea.KeyMsg:
    switch msg.Type {
    case tea.KeyEsc:
        return m, tea.Quit
    }
```

**Mouse Input:** No mouse handling implemented

**Focus Behavior:**
- Focus policy: `FocusNotAllowed` (cannot receive keyboard focus)
- Does not hide on focus loss (persistent visibility)

### Responsive Behavior

**Terminal Resize Handling:**
```go
case tea.WindowSizeMsg:
    m.width = msg.Width
    return m, nil
```

**Layout Adaptation:**
- Spacer width recalculates to maintain right-aligned clock
- Minimum spacer width is 0 (prevents negative values)
- Content may overlap if terminal width is insufficient

### Error Handling

**Hyprland Query Failures:**
```go
// From cmd/shine-bar/main.go getWorkspaces()
// Fallback to single workspace on error
return []workspace{{ID: 1, Name: "1"}}, 1
```

**Expected Behavior:**
- On `hyprctl` failure: Display workspace "1" as active
- Clock continues to update regardless of workspace query status

---

## Technical Implementation Details

### Positioning Calculation

**Remote Control Launch (Preferred Method):**
```go
// From pkg/panel/config.go ToRemoteControlArgs()
panelProps := []string{}
panelProps = append(panelProps, fmt.Sprintf("edge=%s", c.Edge.String()))  // "top"
panelProps = append(panelProps, fmt.Sprintf("lines=%dpx", c.LinesPixels)) // "30px"
```

**Kitty Command:**
```bash
kitty @ --to <socket> launch --type=os-panel \
  --os-panel edge=top \
  --os-panel lines=30px \
  --title shine-bar \
  shine-bar
```

**Kitten Panel Launch (Fallback Method):**
```go
// From pkg/panel/config.go ToKittenArgs()
args = append(args, "--edge="+edgeStr)  // "--edge=top"
args = append(args, "--lines="+strconv.Itoa(c.LinesPixels)+"px")  // "--lines=30px"
```

**Kitten Command:**
```bash
kitten panel --edge=top --lines=30px \
  --focus-policy=not-allowed \
  --single-instance --instance-group=shine \
  -o allow_remote_control=socket-only \
  -o listen_on=unix:/tmp/shine.sock \
  shine-bar
```

### Color Rendering

**Lip Gloss Color Application:**
```go
// Colors applied via ANSI escape sequences
lipgloss.Color("14")  // \033[38;5;14m (256-color foreground)
lipgloss.Color("0")   // \033[48;5;0m  (256-color background)
```

**Bold Attribute:**
```go
Bold(true)  // \033[1m
```

### Font and Text Rendering

**Font Source:** Kitty terminal configuration
**Text Rendering:** Via Kitty's GPU-accelerated text engine
**Character Encoding:** UTF-8 (supports Unicode)

**No Custom Fonts:** Inherits from:
```bash
~/.config/kitty/kitty.conf
# font_family, bold_font, italic_font, bold_italic_font
```

### Wayland/Compositor Integration

**Layer Shell Protocol:** Implemented by Kitty
```go
// From pkg/panel/config.go
Type: LayerShellPanel  // Maps to "bottom" layer in Kitty
Edge: EdgeTop          // Maps to "top" anchor
```

**Exclusive Zone Behavior:**
```go
ExclusiveZone: -1  // Auto-calculated by Kitty
// Kitty reserves 30px at top of screen
// Other windows respect this reservation
```

**Compositor-Specific Notes:**
- Requires Hyprland compositor
- Layer shell support required
- Kitty must be compiled with Wayland support

---

## Configuration Options

### Default Configuration

**From `pkg/config/types.go`:**
```go
func NewDefaultConfig() *Config {
    return &Config{
        Chat: &ChatConfig{
            Enabled: true,
            // ... chat config ...
        },
        // Bar: nil,  // Bar NOT enabled by default!
    }
}
```

**Important:** Status bar is NOT enabled in default configuration. Must be explicitly enabled via config file.

### Example Configuration

**From `examples/shine.toml`:**
```toml
[bar]
enabled = true
edge = "top"
lines_pixels = 30
margin_top = 0
margin_left = 0
margin_right = 0
single_instance = true
hide_on_focus_loss = false
focus_policy = "not-allowed"
# output_name = "DP-1"  # Optional: target specific monitor
```

### Configuration Options Reference

| Option | Type | Default | Description |
|--------|------|---------|-------------|
| `enabled` | bool | false | Enable status bar component |
| `edge` | string | "top" | Screen edge: top, bottom, left, right |
| `lines` | int | 1 | Height in terminal lines |
| `lines_pixels` | int | 30 | Height in pixels (overrides lines) |
| `columns` | int | - | Width in terminal columns (unused, auto-spans) |
| `columns_pixels` | int | - | Width in pixels (unused, auto-spans) |
| `margin_top` | int | 0 | Top margin in pixels |
| `margin_left` | int | 0 | Left margin in pixels |
| `margin_bottom` | int | 0 | Bottom margin in pixels |
| `margin_right` | int | 0 | Right margin in pixels |
| `single_instance` | bool | true | Only one instance allowed |
| `hide_on_focus_loss` | bool | false | Hide when losing focus |
| `focus_policy` | string | "not-allowed" | Focus behavior: not-allowed, exclusive, on-demand |
| `output_name` | string | "" | Target monitor name (empty = primary) |

### Edge Options

Available edge values (from `pkg/panel/config.go`):
- `top` - Top edge (default for bar)
- `bottom` - Bottom edge
- `left` - Left edge
- `right` - Right edge
- `center` - Centered
- `center-sized` - Centered with specific size
- `background` - Background layer
- `top-left` - Top-left corner
- `top-right` - Top-right corner
- `bottom-left` - Bottom-left corner
- `bottom-right` - Bottom-right corner

### Focus Policy Options

- `not-allowed` - No keyboard focus (default for bar)
- `exclusive` - Exclusive keyboard focus
- `on-demand` - Focus on demand

---

## Expected Visual Appearance

### ASCII Representation

**Full Width Display (assuming 1920px monitor @ ~10px per char = 192 chars):**
```
┌────────────────────────────────────────────────────────────────────────────────────────────────────────────────────────────────────────────────────────────────────────────────────────────────┐
│ 1  2  3  4                                                                                                                                                                          14:23:47 │
└────────────────────────────────────────────────────────────────────────────────────────────────────────────────────────────────────────────────────────────────────────────────────────────────┘
 ^^^^^^^^^^                                                                                                                                                                          ^^^^^^^^^^
 Workspaces (left-aligned)                                                                                                                                                          Clock (right-aligned)
```

**Color Visualization:**
```
 1  2  3  4                                                      14:23:47
 ┬  ┬  ┬  ┬                                                      ┬
 │  │  │  │                                                      └─ Bright Magenta (bold)
 │  │  │  └─ White
 │  │  └─ White
 │  └─ Bright Cyan (bold) ← ACTIVE
 └─ White
```

**Character-Level Breakdown:**
```
Position: 0    3    6    9    12                          ~185
Content:  [ 1 ][ 2 ][ 3 ][ 4 ][     spacer     ][14:23:47]
          │ ▲ ││ ▲ ││ ▲ ││ ▲ │                  │ ▲     ▲│
          │ │ ││ │ ││ │ ││ │ │                  │ │     ││
          └─┴─┘└─┴─┘└─┴─┘└─┴─┘                  └─┴─────┴┘
           pad  pad  pad  pad                     pad    pad
```

### Minimal Example (3 workspaces, workspace 2 active)

**Visual:**
```
 1  2  3                                                  14:23:47
    ^^^ Bright cyan, bold
 ^  ^   White
```

**ANSI Escape Sequence Representation:**
```
\033[48;5;0m\033[38;5;15m 1 \033[0m\033[48;5;0m\033[38;5;14m\033[1m 2 \033[0m\033[48;5;0m\033[38;5;15m 3 \033[0m                    \033[48;5;0m\033[38;5;13m\033[1m 14:23:47 \033[0m
│           │          │ │   │ │           │         │          │ │   │ │           │          │ │   │                        │           │         │          │ │     │
│           │          │ │   │ │           │         │          │ │   │ │           │          │ │   │                        │           │         │          │ │     │
│           │          └─┴───┴─┘           │         │          └─┴───┴─┘           │          └─┴───┴─┘                        │           │         │          └─┴─────┘
│           │            " 1 "             │         │            " 2 "             │            " 3 "                          │           │           " 14:23:47 "
│           └─ FG:White                    │         └─ FG:Cyan (bold)              └─ FG:White                                 │           └─ FG:Magenta (bold)
└─ BG:Black                                └─ BG:Black                                  BG:Black                                └─ BG:Black
```

### Real-World Scenarios

**Scenario 1: Many Workspaces (10 workspaces)**
```
 1  2  3  4  5  6  7  8  9  10                            14:23:47
    ^^^ active
```

**Scenario 2: Single Workspace**
```
 1                                                         14:23:47
 ^^^ active
```

**Scenario 3: No Hyprland Connection (Fallback)**
```
 1                                                         14:23:47
 ^^^ active (fallback)
```

**Scenario 4: Narrow Terminal (80 chars)**
```
 1  2  3  4                    14:23:47
```

**Scenario 5: Very Narrow Terminal (30 chars, overlap)**
```
 1  2  3  414:23:47
          ^^^^^^^^^ overlapping content
```

---

## Verification Checklist

### Visual Verification

- [ ] Bar appears at top edge of screen
- [ ] Bar height is exactly 30 pixels
- [ ] Bar spans full width of monitor
- [ ] No margins (flush with screen edges)
- [ ] Black background visible
- [ ] Workspace indicators on left side
- [ ] Clock on right side
- [ ] Workspace numbers match Hyprland workspaces
- [ ] Active workspace highlighted in bright cyan
- [ ] Active workspace text is bold
- [ ] Inactive workspaces shown in white
- [ ] Clock displayed in bright magenta
- [ ] Clock text is bold
- [ ] Clock format is HH:MM:SS (24-hour)
- [ ] Proper spacing between elements
- [ ] No text clipping or truncation

### Functional Verification

- [ ] Clock updates every second
- [ ] Clock shows current time accurately
- [ ] Workspace indicators update when switching workspaces
- [ ] Active workspace indicator follows current workspace
- [ ] Pressing ESC closes the status bar
- [ ] Bar does not disappear when clicking away
- [ ] Bar persists across workspace switches
- [ ] Bar does not accept keyboard focus
- [ ] Terminal resize adjusts layout correctly
- [ ] Clock remains right-aligned on resize

### Technical Verification

- [ ] Window title is "shine-bar"
- [ ] Layer shell type is "panel" (bottom layer)
- [ ] Edge placement is "top"
- [ ] Focus policy is "not-allowed"
- [ ] Single instance mode active (no duplicate bars)
- [ ] Remote control socket at `/tmp/shine.sock` (if applicable)
- [ ] Hyprland queries execute without errors
- [ ] No memory leaks during continuous operation
- [ ] CPU usage minimal (<1% idle)

### Configuration Verification

- [ ] Config loaded from `~/.config/shine/shine.toml`
- [ ] Bar enabled in configuration (`enabled = true`)
- [ ] Edge setting matches configuration
- [ ] Height setting matches configuration (30px)
- [ ] Margins match configuration (all 0)
- [ ] Single instance setting respected
- [ ] Focus policy setting respected
- [ ] Hide on focus loss setting respected (false)

### Error Handling Verification

- [ ] Graceful fallback if Hyprland unavailable
- [ ] Fallback to workspace "1" on query error
- [ ] Clock continues if workspace query fails
- [ ] No crashes on terminal resize
- [ ] Clean exit on ESC key
- [ ] Clean exit on Ctrl+C

---

## Appendix: Code References

### Key Source Files

1. **Main Status Bar Logic:** `/home/starbased/dev/projects/shine/cmd/shine-bar/main.go`
   - Model definition (lines 42-48)
   - View rendering (lines 116-167)
   - Workspace querying (lines 170-202)
   - Color styling (lines 119-134)

2. **Panel Configuration:** `/home/starbased/dev/projects/shine/pkg/panel/config.go`
   - Edge types (lines 36-115)
   - Layer types (lines 10-34)
   - Focus policies (lines 117-151)
   - Remote control args (lines 234-302)
   - Kitten args (lines 304-428)

3. **Panel Manager:** `/home/starbased/dev/projects/shine/pkg/panel/manager.go`
   - Launch methods (lines 165-326)
   - Kitty socket detection (lines 57-163)

4. **Configuration Types:** `/home/starbased/dev/projects/shine/pkg/config/types.go`
   - BarConfig definition (lines 76-92)
   - ToPanelConfig conversion (lines 94-137)
   - Default config (lines 266-280)

5. **Main Launcher:** `/home/starbased/dev/projects/shine/cmd/shine/main.go`
   - Bar launch logic (lines 84-111)
   - Component discovery (lines 17-39)

### Configuration File

**Example Configuration:** `/home/starbased/dev/projects/shine/examples/shine.toml`
- Bar section (lines 32-57)

---

## Revision History

| Version | Date | Changes |
|---------|------|---------|
| 1.0 | 2025-11-02 | Initial specification document created |

---

**Document Status:** APPROVED for Phase 2 Testing
**Next Phase:** Visual comparison testing with screenshots
