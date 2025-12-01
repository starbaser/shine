# Theme Package Quick Start

## Import

```go
import "github.com/starbased-co/shine/cmd/prisms/internal/theme"
```

## Basic Usage

```go
// Get current theme (Srcery auto-registered)
t := theme.Current()

// Use theme colors
primary := t.Primary()
bg := t.Background()
success := t.Success()
```

## Common Patterns

### Panel with Theme

```go
panel := theme.PanelStyle().
    Width(40).
    Height(20).
    Render("Content")
```

### Status Messages

```go
success := theme.SuccessStyle().Render(fmt.Sprintf("%s Success", theme.IconSuccess))
error := theme.ErrorStyle().Render(fmt.Sprintf("%s Error", theme.IconError))
warning := theme.WarningStyle().Render(fmt.Sprintf("%s Warning", theme.IconWarning))
```

### System Icons

```go
cpu := fmt.Sprintf("%s CPU: 45%%", theme.IconCPU)
memory := fmt.Sprintf("%s Memory: 8GB", theme.IconMemory)
battery := theme.GetBatteryIcon(75, false)  // 75%, not charging
volume := theme.GetVolumeIcon(80, false)    // 80%, not muted
```

### Custom Borders

```go
style := lipgloss.NewStyle().
    Border(theme.GlassBorder()).
    BorderForeground(theme.Current().BorderSubtle()).
    Padding(1, 2)
```

### List Items

```go
for i, item := range items {
    styled := theme.ListItemStyle(i == selected).Render(item)
    fmt.Println(styled)
}
```

### Cards

```go
card := theme.CardStyle().Render(lipgloss.JoinVertical(
    lipgloss.Left,
    theme.TitleStyle().Render("Title"),
    "",
    "Card content here",
))
```

### Status Bar

```go
left := "Left content"
right := "Right content"

statusBar := theme.StatusBarStyle().
    Width(totalWidth).
    Render(lipgloss.JoinHorizontal(
        lipgloss.Left,
        left,
        lipgloss.NewStyle().Width(gap).Render(""),
        right,
    ))
```

## All Style Builders

| Function | Purpose |
|----------|---------|
| `PanelStyle()` | Base panel with border |
| `ActivePanelStyle()` | Panel with active border |
| `CardStyle()` | Raised surface card |
| `TitleStyle()` | Primary heading |
| `SubtitleStyle()` | Secondary heading |
| `TextStyle()` | Base text |
| `MutedTextStyle()` | Muted/subtle text |
| `ErrorStyle()` | Error messages |
| `WarningStyle()` | Warning messages |
| `SuccessStyle()` | Success messages |
| `InfoStyle()` | Info messages |
| `StatusBarStyle()` | Status bar |
| `ButtonStyle(active)` | Buttons |
| `BadgeStyle(color)` | Small badges/tags |
| `ListItemStyle(selected)` | List items |
| `SeparatorStyle()` | Separators |
| `CodeBlockStyle()` | Code blocks |
| `ProgressBarStyle()` | Progress bars |
| `GlassPanel()` | Translucent panels |
| `HighlightStyle()` | Highlighted text |

## All Border Styles

| Function | Style |
|----------|-------|
| `GlassBorder()` | Rounded corners (╭─╮) |
| `ThickBorder()` | Heavy lines (┏━┓) |
| `DoubleBorder()` | Double lines (╔═╗) |
| `DashedBorder()` | Dashed lines (┄┆) |
| `MinimalBorder()` | Square corners (┌─┐) |
| `BracketBorder()` | Bracket corners (⎾─⏋) |

## Common Icons

### System
- `IconCPU` -
- `IconMemory` - 󰍛
- `IconDisk` - 󰋊
- `IconNetwork` - 󰖩
- `IconBattery` - 󰁹

### Status
- `IconSuccess` - 󰄬
- `IconError` - 󰅖
- `IconWarning` -
- `IconInfo` - 󰋽

### UI
- `IconDashboard` - 󰕮
- `IconSettings` -
- `IconSearch` -
- `IconMenu` -
- `IconClock` - 󰥔

See `icons.go` for complete list (100+ icons).

## Theme Interface

All colors available via `theme.Current()`:

```go
t := theme.Current()

// Base
t.Background()
t.Foreground()

// Semantic
t.Primary()      // Main accent
t.Secondary()    // Supporting accent
t.Accent()       // Highlight
t.Muted()        // Subdued

// Status
t.Success()      // Green
t.Warning()      // Yellow
t.Error()        // Red
t.Info()         // Blue

// Surfaces (layered)
t.Surface0()     // Base
t.Surface1()     // Raised
t.Surface2()     // Highest

// Borders
t.BorderSubtle() // Inactive
t.BorderActive() // Active

// Text
t.TextPrimary()
t.TextSecondary()
t.TextMuted()
```

## Demo Applications

```bash
# Comprehensive showcase
./bin/theme-demo

# Interactive panel example
./bin/simple-panel
```

## Full Documentation

- `README.md` - Complete usage guide
- `SUMMARY.md` - Implementation summary
- `doc.go` - Go package docs (godoc)
- `example_test.go` - Runnable examples
