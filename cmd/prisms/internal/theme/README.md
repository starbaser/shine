# Theme Package

Generic, reusable theme system for Bubble Tea TUI applications in Shine.

## Overview

The `theme` package provides a flexible theming system with:

- **Generic Theme interface** - Not tied to any specific colorscheme
- **Theme registry** - Runtime theme switching
- **Pre-built style builders** - Reusable Lipgloss styles
- **Nerd Font icons** - Common system widget icons
- **Custom borders** - Various border styles for panels

## Usage

### Basic Setup

```go
package main

import (
    "github.com/charmbracelet/lipgloss"
    "shine/cmd/prisms/internal/theme"
)

func main() {
    // Srcery is auto-registered via init()
    // Use current theme
    t := theme.Current()

    // Create styled text
    style := lipgloss.NewStyle().
        Foreground(t.Primary()).
        Background(t.Surface0())

    // Or use pre-built styles
    title := theme.TitleStyle().Render("My App")
    panel := theme.PanelStyle().Render("Content")
}
```

### Using Pre-Built Styles

```go
// Panel with theme colors
panel := theme.PanelStyle().
    Width(40).
    Height(20).
    Render("Panel content")

// Active panel with highlighted border
activePanel := theme.ActivePanelStyle().
    Width(40).
    Height(20).
    Render("Active panel")

// Card with raised surface
card := theme.CardStyle().Render("Card content")

// Status messages
success := theme.SuccessStyle().Render("✓ Operation complete")
error := theme.ErrorStyle().Render("✗ Error occurred")
warning := theme.WarningStyle().Render("⚠ Warning")
```

### Icons

```go
import "shine/cmd/prisms/internal/theme"

// System status icons
cpuIcon := theme.IconCPU
memIcon := theme.IconMemory

// Battery icon based on state
batteryIcon := theme.GetBatteryIcon(75, false) // 75%, not charging
chargingIcon := theme.GetBatteryIcon(50, true) // 50%, charging

// Volume icon based on level
volumeIcon := theme.GetVolumeIcon(80, false) // 80%, not muted
mutedIcon := theme.GetVolumeIcon(0, true)    // muted
```

### Custom Borders

```go
// Glass panel with rounded corners
glassBorder := theme.GlassBorder()
glassPanel := theme.GlassPanel().Render("Glass effect")

// Thick border for emphasis
thickStyle := lipgloss.NewStyle().
    Border(theme.ThickBorder()).
    BorderForeground(theme.Current().Primary())

// Dashed border for subtle separation
dashedStyle := lipgloss.NewStyle().
    Border(theme.DashedBorder()).
    BorderForeground(theme.Current().BorderSubtle())
```

### Runtime Theme Switching

```go
// List available themes
themes := theme.List()
// ["srcery"]

// Switch theme (when more themes are added)
err := theme.SetCurrent("catppuccin")
if err != nil {
    // Theme not found
}

// Get specific theme without activating
srcery := theme.Get("srcery")
```

## Theme Interface

Implement the `Theme` interface to add new themes:

```go
type MyTheme struct{}

func (t MyTheme) Name() string { return "mytheme" }

// Base colors
func (t MyTheme) Background() lipgloss.TerminalColor { ... }
func (t MyTheme) Foreground() lipgloss.TerminalColor { ... }

// Semantic colors
func (t MyTheme) Primary() lipgloss.TerminalColor { ... }
func (t MyTheme) Secondary() lipgloss.TerminalColor { ... }
func (t MyTheme) Accent() lipgloss.TerminalColor { ... }
func (t MyTheme) Muted() lipgloss.TerminalColor { ... }

// Status colors
func (t MyTheme) Success() lipgloss.TerminalColor { ... }
func (t MyTheme) Warning() lipgloss.TerminalColor { ... }
func (t MyTheme) Error() lipgloss.TerminalColor { ... }
func (t MyTheme) Info() lipgloss.TerminalColor { ... }

// Surface colors (layered backgrounds)
func (t MyTheme) Surface0() lipgloss.TerminalColor { ... }
func (t MyTheme) Surface1() lipgloss.TerminalColor { ... }
func (t MyTheme) Surface2() lipgloss.TerminalColor { ... }

// Border colors
func (t MyTheme) BorderSubtle() lipgloss.TerminalColor { ... }
func (t MyTheme) BorderActive() lipgloss.TerminalColor { ... }

// Text hierarchy
func (t MyTheme) TextPrimary() lipgloss.TerminalColor { ... }
func (t MyTheme) TextSecondary() lipgloss.TerminalColor { ... }
func (t MyTheme) TextMuted() lipgloss.TerminalColor { ... }

func init() {
    theme.Register(MyTheme{})
}
```

## Color Semantics

### Base Colors
- `Background()`: Base background color
- `Foreground()`: Default text color

### Semantic Colors
- `Primary()`: Main brand/accent color for interactive elements
- `Secondary()`: Supporting accent color
- `Accent()`: Highlight color for emphasis
- `Muted()`: Subdued color for less important elements

### Status Colors
- `Success()`: Positive/success state (green)
- `Warning()`: Caution/warning state (yellow/orange)
- `Error()`: Error/danger state (red)
- `Info()`: Informational state (blue/cyan)

### Surface Colors
- `Surface0()`: Base surface (usually same as background)
- `Surface1()`: Raised surface (cards, modals)
- `Surface2()`: Highest surface (tooltips, popovers)

### Border Colors
- `BorderSubtle()`: Subtle/inactive borders
- `BorderActive()`: Active/focused borders

### Text Hierarchy
- `TextPrimary()`: Primary text (headings, important content)
- `TextSecondary()`: Secondary text (descriptions)
- `TextMuted()`: Muted text (metadata, timestamps)

## Included Themes

### Srcery
Dark colorscheme based on srcery.sh vim theme.

**Palette highlights**:
- Primary: Bright Blue (#68A8E4)
- Success: Bright Green (#98BC37)
- Warning: Bright Yellow (#FED06E)
- Error: Bright Red (#F75341)

**Extended methods**:
- Vim highlight group mappings (Comment, String, Number, etc.)
- Search highlighting colors
- Diff colors (Add, Change, Delete)

## Future Themes

The interface is designed to support additional themes:

- **Catppuccin** (Mocha, Latte, Frappe, Macchiato)
- **Nord** (Arctic-inspired)
- **Gruvbox** (Retro groove)
- **Tokyo Night** (Night-inspired)
- **One Dark** (Atom-inspired)

## Requirements

- Nerd Font for icons (recommended: JetBrainsMono Nerd Font)
- Terminal with true color support (24-bit)
- Charm Bracelet Lipgloss library

## Examples

See `examples/theme_demo.go` for a complete demonstration.
