# Shine Grid System

## Overview

The Shine grid system provides a standardized 16×9 grid overlay for precise widget positioning on a 2560×1440 display. Each cell is 160×160 pixels, with 0-indexed coordinates starting at `(0,0)` in the top-left corner.

## Grid Specifications

```
Display:     2560×1440 pixels
Grid:        16 columns × 9 rows
Cell Size:   160×160 pixels
Coordinates: 0-indexed, (column, row) format
Origin:      (0,0) at top-left
```

## Coordinate System

### Basic Cell Addressing

```
(0,0)  = Top-left corner     = (0px, 0px)
(15,0) = Top-right corner    = (2400px, 0px)
(0,8)  = Bottom-left corner  = (0px, 1280px)
(15,8) = Bottom-right corner = (2400px, 1280px)
(7,4)  = Near center         = (1120px, 640px)
```

### Position Calculation

**Formula:**
```
absolute_x = column * 160 + offset_x
absolute_y = row * 160 + offset_y
```

**Examples:**
```python
# Cell (3, 5) with offset (80, 40)
x = 3 * 160 + 80 = 560px
y = 5 * 160 + 40 = 840px
# Result: (560px, 840px)

# Cell (0, 8) - bottom-left, no offset
x = 0 * 160 + 0 = 0px
y = 8 * 160 + 0 = 1280px
# Result: (0px, 1280px)
```

## Grid Background Image

### Generation

Generate the grid background for visual reference:

```bash
# Default output to ~/Pictures/shine-grid-background.png
uv run scripts/generate_grid_background.py

# Custom output location
uv run scripts/generate_grid_background.py -o /path/to/output.png

# Larger font for better visibility
uv run scripts/generate_grid_background.py --font-size 28

# Without corner pixel markers
uv run scripts/generate_grid_background.py --no-corners
```

### Setting as Wallpaper

**Hyprland:**
```bash
# Set directly
hyprctl hyprpaper wallpaper "DP-2,./scripts/.png"

# Or add to hyprpaper.conf
preload = ./scripts/.png
wallpaper = DP-2,./scripts/.png
```

**Other methods:**
```bash
# Using feh
feh --bg-scale ~/Pictures/shine-grid-background.png

# Using nitrogen
nitrogen --set-scaled ~/Pictures/shine-grid-background.png
```

## Widget Positioning

### Widget Specification Form

Use `docs/WIDGET_SPEC_FORM.md` to document widget positions. Example:

```yaml
position:
  grid_cell:
    column: 7      # 0-15
    row: 8         # 0-8
  offset:
    x: 0           # 0-159 pixels within cell
    y: 20          # 0-159 pixels within cell
  absolute:
    x: 1120        # column*160 + offset.x
    y: 1300        # row*160 + offset.y
```

### Common Positions

**Status Bar (Full Width, Bottom):**
```yaml
grid_cell: { column: 0, row: 8 }
offset: { x: 0, y: 0 }
absolute: { x: 0, y: 1280 }
size: { width: 2560, height: 160 }
```

**Clock Widget (Center-Bottom):**
```yaml
grid_cell: { column: 7, row: 8 }
offset: { x: 0, y: 20 }
absolute: { x: 1120, y: 1300 }
size: { width: 320, height: 120 }
```

**System Info (Top-Right):**
```yaml
grid_cell: { column: 14, row: 0 }
offset: { x: 0, y: 10 }
absolute: { x: 2240, y: 10 }
size: { width: 320, height: 140 }
```

## Automated Testing

### Visual Verification

1. Set grid background as wallpaper
2. Launch Shine components
3. Verify widget positions match grid coordinates
4. Take screenshots for comparison

### Position Testing

```go
// Example Go test using grid system
func TestWidgetPosition(t *testing.T) {
    // Widget at cell (7, 8) with offset (0, 20)
    expectedX := 7*160 + 0   // 1120
    expectedY := 8*160 + 20  // 1300

    pos := widget.GetPosition()
    assert.Equal(t, expectedX, pos.X)
    assert.Equal(t, expectedY, pos.Y)
}
```

### Integration Testing Workflow

1. **Setup**: Set grid background as wallpaper
2. **Launch**: Start Shine components with test configuration
3. **Capture**: Take screenshot of desktop
4. **Verify**: Compare widget positions against grid
5. **Report**: Generate visual diff with annotations

## Multi-Monitor Support

### Scaling Calculations

For different resolutions, scale the grid proportionally:

```python
def scale_position(cell_col, cell_row, offset_x, offset_y, target_width, target_height):
    """Scale grid position to different resolution."""
    scale_x = target_width / 2560
    scale_y = target_height / 1440

    base_x = cell_col * 160 * scale_x
    base_y = cell_row * 160 * scale_y

    return (
        int(base_x + offset_x * scale_x),
        int(base_y + offset_y * scale_y)
    )

# Example: 1920×1080 display
x, y = scale_position(7, 4, 80, 80, 1920, 1080)
# Result: (900, 540) - center of 1920×1080
```

### Per-Monitor Configuration

```toml
# config.toml
[display.DP-2]
resolution = "2560x1440"
grid_enabled = true

[display.DP-3]
resolution = "1920x1080"
grid_enabled = false
```

## Utilities

### Position Calculator

```bash
# Calculate absolute position from grid coordinates
scripts/grid_calc.py 7 8 0 20
# Output: (1120, 1300)
```

### Grid Overlay Tool (Planned)

Interactive overlay for development:

```bash
# Show grid overlay with coordinates (planned)
shine-grid --overlay --show-coordinates

# Toggle grid display
shine-grid --toggle
```

## File Locations

- **Form Template**: `docs/WIDGET_SPEC_FORM.md`
- **Widget Specs**: `docs/widgets/*.yaml`
- **Generator Script**: `scripts/generate_grid_background.py`
- **Default Output**: `~/Pictures/shine-grid-background.png`

## Development Guidelines

1. **Always use grid coordinates** for widget positioning in specs
2. **Document offsets** when fine-tuning position within cells
3. **Include absolute coordinates** for verification
4. **Test on grid background** before committing widget positions
5. **Update widget specs** when changing positions

## See Also

- [Widget Specification Form](WIDGET_SPEC_FORM.md)
- [Shine Configuration Guide](../README.md)
- [Integration Testing Guide](llms/reports/integration-test-complete.md)
