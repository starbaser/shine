# Weather Prism Example

Simple weather widget prism for Shine demonstrating the prism interface.

**What is a Prism?** A self-contained executable that refracts light (Shine) to display information. This weather prism shows location, condition, and temperature.

## Features

- Displays location, weather condition, and temperature
- Updates every 15 minutes
- Minimal resource usage
- Clean, readable styling

## Building

```bash
cd docs/examples/prism-weather

# Initialize Go module
go mod init shine-weather

# Add dependencies
go get github.com/charmbracelet/bubbletea
go get github.com/charmbracelet/lipgloss

# Build prism binary
go build -o shine-weather main.go
```

## Installation

### Option 1: User Prism Directory

```bash
# Create prism directory
mkdir -p ~/.config/shine/prisms

# Copy binary
cp shine-weather ~/.config/shine/prisms/

# Configure Shine
```

Add to `~/.config/shine/shine.toml`:

```toml
[core]
prism_dirs = ["~/.config/shine/prisms"]
auto_path = true

[prisms.weather]
enabled = true
name = "weather"
edge = "top-right"
columns_pixels = 300
lines_pixels = 30
margin_top = 10
margin_right = 10
focus_policy = "not-allowed"
output_name = "DP-2"  # Adjust to your monitor
```

### Option 2: System-Wide Installation

```bash
# Copy to user bin directory
cp shine-weather ~/.local/bin/

# Ensure ~/.local/bin is in PATH
export PATH="$HOME/.local/bin:$PATH"
```

## Configuration

All standard Shine panel configuration options are supported:

| Option | Description | Example |
|--------|-------------|---------|
| `edge` | Panel placement | `"top-right"`, `"bottom"`, `"left"` |
| `lines_pixels` | Height in pixels | `30` |
| `columns_pixels` | Width in pixels | `300` |
| `margin_top` | Top margin | `10` |
| `margin_right` | Right margin | `10` |
| `focus_policy` | Keyboard focus | `"not-allowed"` |
| `output_name` | Target monitor | `"DP-2"` |

## Usage

```bash
# Start Shine (will automatically load weather prism if enabled)
shine
```

## Customization

### Change Update Interval

Edit `tickCmd()` in `main.go`:

```go
func tickCmd() tea.Cmd {
    // Change from 15 minutes to 5 minutes
    return tea.Tick(5*time.Minute, func(t time.Time) tea.Msg {
        return tickMsg(t)
    })
}
```

### Add Real Weather API

Replace the mock data in `Update()` with actual API calls:

```go
case tickMsg:
    m.lastUpdate = time.Time(msg)

    // Fetch real weather data
    weather, err := fetchWeather(m.location)
    if err != nil {
        m.err = err
        return m, tickCmd()
    }

    m.temperature = weather.Temperature
    m.condition = weather.Condition
    return m, tickCmd()
```

### Style Customization

Edit the styles in `View()` to match your theme:

```go
mainStyle := lipgloss.NewStyle().
    Foreground(lipgloss.Color("14")). // Your color choice
    Background(lipgloss.Color("0")).
    Bold(true).
    Padding(0, 1)
```

## Prism Interface Compliance

This prism demonstrates all required interface elements:

1. **Window Title**: Set via ANSI escape sequence for tracking
2. **No Alt Screen**: Omits `tea.WithAltScreen()` for panel compatibility
3. **Bubble Tea Program**: Standard Model-Update-View pattern
4. **Window Resize Handling**: Responds to `tea.WindowSizeMsg`
5. **Clean Exit**: Quits on Esc/Ctrl+C
6. **Standard Streams**: Uses stdin/stdout (Bubble Tea default)

## Troubleshooting

### Prism Not Discovered

Check prism search paths:

```bash
# Verify binary exists
ls -la ~/.config/shine/prisms/shine-weather

# Verify it's executable
chmod +x ~/.config/shine/prisms/shine-weather

# Check Shine config
cat ~/.config/shine/shine.toml
```

### Prism Crashes on Launch

Run standalone to see errors:

```bash
~/.config/shine/prisms/shine-weather
```

### Not Rendering in Panel

Ensure you're NOT using alt screen mode:

```go
// WRONG - breaks panel rendering
p := tea.NewProgram(initialModel(), tea.WithAltScreen())

// CORRECT - works in panels
p := tea.NewProgram(initialModel())
```

## Development Tips

1. **Test Standalone First**: Run prism directly before integrating with Shine
2. **Use High Contrast Colors**: Panels need bright colors for visibility
3. **Keep It Small**: Panels are small, design for compact space
4. **Minimize Updates**: Reduce CPU usage with longer tick intervals
5. **Handle Resize**: Always respond to `tea.WindowSizeMsg`

## Understanding Prisms

**The Prism Metaphor**:
- Light (Shine) passes through prisms
- Each prism refracts differently to show different information
- Prisms are self-contained: binary + configuration + runtime management
- No distinction between "built-in" and "user" prisms

**Why "Prism"?**
- Evokes the idea of light refraction (displaying information)
- Self-contained, independent units
- Each unique but part of a unified system
- More elegant than "plugin" or "component"

## Further Reading

- [Bubble Tea Documentation](https://github.com/charmbracelet/bubbletea)
- [Lip Gloss Styling](https://github.com/charmbracelet/lipgloss)
- [Shine Prism Development Guide](../../PRISM_DEVELOPMENT.md)
- [Shine Prism System Design](../../PRISM_SYSTEM_DESIGN.md)
