# Shine Widget Specification Form

## Overview

Use this form to document all desktop widgets, panels, and UI elements in the Shine project.

## Grid System Reference

```
Display: 2560x1440 (16×9 cells, each cell 160×160 pixels)

Position Calculation (0-indexed):
- Cell notation: cell(column, row) where column: 0-15, row: 0-8
- Base position: cell(3, 5) = (480px, 800px)
- Fine position: cell(3, 5) + offset(80, 40) = (560px, 840px)
- Formula: (column * 160 + offset_x, row * 160 + offset_y)

Grid Layout:
0    1    2    3    4    5    6    7    8    9    10   11   12   13   14   15
/----+----+----+----+----+----+----+----+----+----+----+----+----+----+----+---\
╭────┬────┬────┬────┬────┬────┬────┬────┬────┬────┬────┬────┬────┬────┬────┬────╮ 0
│0,0 │1,0 │2,0 │3,0 │4,0 │5,0 │6,0 │7,0 │8,0 │9,0 │10,0│11,0│12,0│13,0│14,0│15,0│
├────┼────┼────┼────┼────┼────┼────┼────┼────┼────┼────┼────┼────┼────┼────┼────┤ 1
│0,1 │1,1 │2,1 │3,1 │4,1 │5,1 │6,1 │7,1 │8,1 │9,1 │10,1│11,1│12,1│13,1│14,1│15,1│
├────┼────┼────┼────┼────┼────┼────┼────┼────┼────┼────┼────┼────┼────┼────┼────┤ 2
│0,2 │1,2 │2,2 │3,2 │4,2 │5,2 │6,2 │7,2 │8,2 │9,2 │10,2│11,2│12,2│13,2│14,2│15,2│
├────┼────┼────┼────┼────┼────┼────┼────┼────┼────┼────┼────┼────┼────┼────┼────┤ 3
│0,3 │1,3 │2,3 │3,3 │4,3 │5,3 │6,3 │7,3 │8,3 │9,3 │10,3│11,3│12,3│13,3│14,3│15,3│
├────┼────┼────┼────┼────┼────┼────┼────┼────┼────┼────┼────┼────┼────┼────┼────┤ 4
│0,4 │1,4 │2,4 │3,4 │4,4 │5,4 │6,4 │7,4 │8,4 │9,4 │10,4│11,4│12,4│13,4│14,4│15,4│
├────┼────┼────┼────┼────┼────┼────┼────┼────┼────┼────┼────┼────┼────┼────┼────┤ 5
│0,5 │1,5 │2,5 │3,5 │4,5 │5,5 │6,5 │7,5 │8,5 │9,5 │10,5│11,5│12,5│13,5│14,5│15,5│
├────┼────┼────┼────┼────┼────┼────┼────┼────┼────┼────┼────┼────┼────┼────┼────┤ 6
│0,6 │1,6 │2,6 │3,6 │4,6 │5,6 │6,6 │7,6 │8,6 │9,6 │10,6│11,6│12,6│13,6│14,6│15,6│
├────┼────┼────┼────┼────┼────┼────┼────┼────┼────┼────┼────┼────┼────┼────┼────┤ 7
│0,7 │1,7 │2,7 │3,7 │4,7 │5,7 │6,7 │7,7 │8,7 │9,7 │10,7│11,7│12,7│13,7│14,7│15,7│
├────┼────┼────┼────┼────┼────┼────┼────┼────┼────┼────┼────┼────┼────┼────┼────┤ 8
│0,8 │1,8 │2,8 │3,8 │4,8 │5,8 │6,8 │7,8 │8,8 │9,8 │10,8│11,8│12,8│13,8│14,8│15,8│
╰────┴────┴────┴────┴────┴────┴────┴────┴────┴────┴────┴────┴────┴────┴────┴────╯
```

---

## Widget Specification Template

Copy this template for each widget/element you're documenting:

```yaml
---
# Basic Information
widget_id: "" # Unique identifier (e.g., "status-bar-main", "clock-widget")
widget_name: "" # Human-readable name
component: "" # Component name (e.g., "shine-bar", "shine-clock")
developer: "" # Your name/handle
date: "" # YYYY-MM-DD
status: "" # planned | in-development | implemented | deprecated

# Position & Dimensions
position:
  grid_cell:
    column: 0 # 0-15
    row: 0 # 0-8
  offset: # Fine-tune position within cell
    x: 0 # 0-159 pixels within cell
    y: 0 # 0-159 pixels within cell
  absolute: # Calculated absolute position
    x: 0 # column*160 + offset.x
    y: 0 # row*160 + offset.y

size:
  width: 0 # Width in pixels
  height: 0 # Height in pixels
  grid_span: # Optional: span across cells
    columns: 0 # Number of columns
    rows: 0 # Number of rows

# Anchoring & Alignment
anchor:
  "" # top-left | top-center | top-right | center-left | center |
  # center-right | bottom-left | bottom-center | bottom-right
alignment: "" # left | center | right | stretch

# Visual Properties
colors:
  background: "" # Hex, RGB, or named color
  foreground: "" # Text/icon color
  border: "" # Border color (if applicable)
  accent: "" # Accent/highlight color

transparency:
  opacity: 1.0 # 0.0 (transparent) to 1.0 (opaque)
  blur: false # Enable background blur

borders:
  enabled: false
  width: 0 # Border width in pixels
  radius: 0 # Border radius for rounded corners
  style: "" # solid | dashed | dotted

# Typography (if applicable)
font:
  family: "" # Font family
  size: 0 # Font size in pixels
  weight: "" # normal | bold | light | etc.
  style: "" # normal | italic

# Functionality
functionality:
  primary_purpose: "" # Brief description of main function
  interactions: [] # List of user interactions (click, hover, drag, etc.)
  updates: "" # static | dynamic | real-time
  update_interval: "" # e.g., "1s", "30s", "on-event"

# Content
content:
  type: "" # text | icon | image | mixed | custom
  source: "" # Where content comes from (config, system, API, etc.)
  format: "" # Format/template for content display
  example: "" # Example of actual content shown

# Technical Details
implementation:
  manager: "" # Which manager handles this (panel, remote, standalone)
  config_key: "" # Configuration key in config.toml
  dependencies: [] # Other components this depends on
  file_location: "" # Primary source file path

# Layer & Z-Index
layer:
  z_index: 0 # Stacking order (higher = on top)
  layer_name: "" # Logical layer (background, base, overlay, etc.)

# Behavior
behavior:
  always_visible: true # Always shown or conditional
  conditions: [] # Conditions for visibility
  animations: [] # List of animations (fade, slide, etc.)
  responsive: false # Adapts to screen size changes

# Multi-Monitor
multi_monitor:
  display_mode: "" # all | primary | specific | per-monitor-config
  positioning: "" # absolute | relative-to-monitor

# Notes
notes: "" # Additional notes, TODOs, or considerations

# Related Widgets
related: [] # IDs of related widgets
---
```

---

## Example: Status Bar Widget

```yaml
---
widget_id: "status-bar-main"
widget_name: "Main Status Bar"
component: "shine-bar"
developer: "starbased"
date: "2025-11-02"
status: "implemented"

position:
  grid_cell:
    column: 0
    row: 8
  offset:
    x: 0
    y: 0
  absolute:
    x: 0
    y: 1280

size:
  width: 2560
  height: 160
  grid_span:
    columns: 16
    rows: 1

anchor: "bottom-left"
alignment: "stretch"

colors:
  background: "#1e1e2e"
  foreground: "#cdd6f4"
  border: "#313244"
  accent: "#89b4fa"

transparency:
  opacity: 0.95
  blur: true

borders:
  enabled: true
  width: 1
  radius: 0
  style: "solid"

font:
  family: "JetBrains Mono"
  size: 12
  weight: "normal"
  style: "normal"

functionality:
  primary_purpose: "Display workspace info, clock, and system status"
  interactions:
    - "click: switch workspace"
    - "hover: show tooltip"
  updates: "real-time"
  update_interval: "1s"

content:
  type: "mixed"
  source: "system + config"
  format: "[workspace] | [clock] | [system-info]"
  example: "workspace-1 | 14:23:45 | CPU: 45% RAM: 8.2GB"

implementation:
  manager: "panel"
  config_key: "panel.components.status_bar"
  dependencies:
    - "shine-clock"
    - "workspace-manager"
  file_location: "cmd/shine-bar/main.go"

layer:
  z_index: 100
  layer_name: "overlay"

behavior:
  always_visible: true
  conditions: []
  animations:
    - "slide-up on show"
  responsive: true

multi_monitor:
  display_mode: "all"
  positioning: "relative-to-monitor"

notes: |
  Currently experiencing text visibility issues in certain lighting conditions.
  See issue #phase-2-statusbar for ongoing fixes.

related:
  - "clock-widget"
  - "workspace-indicator"
---
```

---

## Example: Clock Widget

```yaml
---
widget_id: "clock-widget"
widget_name: "Digital Clock"
component: "shine-clock"
developer: "starbased"
date: "2025-11-02"
status: "implemented"

position:
  grid_cell:
    column: 7
    row: 8
  offset:
    x: 0
    y: 20
  absolute:
    x: 1120
    y: 1300

size:
  width: 320
  height: 120
  grid_span:
    columns: 2
    rows: 1

anchor: "bottom-center"
alignment: "center"

colors:
  background: "transparent"
  foreground: "#89b4fa"
  border: "none"
  accent: "#cba6f7"

transparency:
  opacity: 1.0
  blur: false

borders:
  enabled: false
  width: 0
  radius: 0
  style: ""

font:
  family: "JetBrains Mono"
  size: 48
  weight: "bold"
  style: "normal"

functionality:
  primary_purpose: "Display current time"
  interactions:
    - "click: show calendar"
    - "right-click: time settings"
  updates: "real-time"
  update_interval: "1s"

content:
  type: "text"
  source: "system-time"
  format: "HH:MM:SS"
  example: "14:23:45"

implementation:
  manager: "standalone"
  config_key: "clock.format"
  dependencies: []
  file_location: "cmd/shine-clock/main.go"

layer:
  z_index: 50
  layer_name: "base"

behavior:
  always_visible: true
  conditions: []
  animations: []
  responsive: false

multi_monitor:
  display_mode: "primary"
  positioning: "absolute"

notes: |
  Part of status bar component system.
  Can run standalone or embedded.

related:
  - "status-bar-main"
---
```

---

## Quick Reference: Common Positions

### Top Bar

```yaml
position:
  grid_cell: { column: 0, row: 0 }
  offset: { x: 0, y: 0 }
  absolute: { x: 0, y: 0 }
size:
  width: 2560
  height: 160
```

### Bottom Bar

```yaml
position:
  grid_cell: { column: 0, row: 8 }
  offset: { x: 0, y: 0 }
  absolute: { x: 0, y: 1280 }
size:
  width: 2560
  height: 160
```

### Top-Right Corner

```yaml
position:
  grid_cell: { column: 15, row: 0 }
  offset: { x: 0, y: 0 }
  absolute: { x: 2400, y: 0 }
anchor: "top-right"
```

### Center Screen

```yaml
position:
  grid_cell: { column: 7, row: 4 }
  offset: { x: 80, y: 80 }
  absolute: { x: 1200, y: 720 }
anchor: "center"
```

### Left Panel

```yaml
position:
  grid_cell: { column: 0, row: 0 }
  offset: { x: 0, y: 0 }
  absolute: { x: 0, y: 0 }
size:
  width: 320
  height: 1440
```

---

## Submission

1. Create a new file in `docs/widgets/` with format: `{widget-id}.yaml`
2. Fill out the template above
3. Include visual mockup if available (place in `docs/widgets/mockups/`)
4. Submit PR with tag: `widget-spec`

## Validation Checklist

- [ ] Unique widget_id
- [ ] Valid grid coordinates (column: 0-15, row: 0-8)
- [ ] Offset within bounds (0-159 pixels)
- [ ] Absolute position calculated correctly (column*160 + offset_x, row*160 + offset_y)
- [ ] All required fields completed
- [ ] Colors specified in valid format
- [ ] File location exists or is planned
- [ ] Related widgets referenced correctly

---

## Tools

### Position Calculator

```python
def calculate_absolute_position(column: int, row: int, offset_x: int = 0, offset_y: int = 0):
    """Calculate absolute pixel position from 0-indexed grid coordinates.

    Args:
        column: Grid column (0-15)
        row: Grid row (0-8)
        offset_x: Pixel offset within cell (0-159)
        offset_y: Pixel offset within cell (0-159)

    Returns:
        Tuple of (x, y) absolute pixel position
    """
    x = column * 160 + offset_x
    y = row * 160 + offset_y
    return (x, y)

# Examples
pos1 = calculate_absolute_position(column=3, row=5, offset_x=80, offset_y=40)
print(f"Position: {pos1}")  # Position: (560, 840)

pos2 = calculate_absolute_position(column=0, row=8)
print(f"Bottom-left: {pos2}")  # Bottom-left: (0, 1280)

pos3 = calculate_absolute_position(column=7, row=4, offset_x=80, offset_y=80)
print(f"Center: {pos3}")  # Center: (1200, 720)
```

### Grid Visualizer

```bash
# Generate visual grid with widget positions
cd /home/starbased/dev/projects/shine
./scripts/visualize_widgets.sh docs/widgets/*.yaml
```

---

## Questions?

See `docs/WIDGET_SYSTEM.md` for architecture details or contact the Shine development team.
