# Weather Prism Example

Displays current weather information with periodic updates.

## Features

- Simulated weather data (temperature, condition, humidity, wind speed)
- Auto-refresh every 5 minutes
- Weather icons (emoji)
- High-contrast styling for panel visibility
- Horizontal layout optimized for top/bottom edges

## Building

```bash
cd examples/prisms/weather
make build
```

## Installation

```bash
make install
```

This installs `shine-weather` to `~/.local/bin/`.

## Configuration

Add to `~/.config/shine/shine.toml`:

```toml
[prisms.weather]
enabled = true
edge = "top-right"
columns_pixels = 400
lines_pixels = 30
margin_top = 0
margin_right = 0
focus_policy = "not-allowed"
```

## Real Weather Data

To fetch real weather data, you can integrate with weather APIs:

### Option 1: OpenWeatherMap API

```go
import (
	"encoding/json"
	"net/http"
)

func fetchWeatherData() weatherData {
	apiKey := os.Getenv("OPENWEATHER_API_KEY")
	city := "San Francisco"
	url := fmt.Sprintf("https://api.openweathermap.org/data/2.5/weather?q=%s&appid=%s&units=imperial", city, apiKey)

	resp, err := http.Get(url)
	if err != nil {
		// Handle error
		return weatherData{}
	}
	defer resp.Body.Close()

	var result struct {
		Main struct {
			Temp     float64 `json:"temp"`
			Humidity int     `json:"humidity"`
		} `json:"main"`
		Weather []struct {
			Main string `json:"main"`
		} `json:"weather"`
		Wind struct {
			Speed float64 `json:"speed"`
		} `json:"wind"`
	}

	json.NewDecoder(resp.Body).Decode(&result)

	return weatherData{
		temperature: int(result.Main.Temp),
		condition:   result.Weather[0].Main,
		humidity:    result.Main.Humidity,
		windSpeed:   int(result.Wind.Speed),
	}
}
```

### Option 2: Weather Command (Linux)

```go
import "os/exec"

func fetchWeatherData() weatherData {
	// Use wttr.in service
	cmd := exec.Command("curl", "-s", "wttr.in/San+Francisco?format=j1")
	output, _ := cmd.Output()
	// Parse JSON output
	// ...
}
```

## Testing

```bash
# Run standalone
make run

# Test with kitty panel
kitten panel --edge=top-right --columns=400px --lines=30px ./shine-weather
```

## Customization

### Change Update Frequency

Edit `tickCmd()` in `main.go`:

```go
func tickCmd() tea.Cmd {
	return tea.Tick(10*time.Minute, func(t time.Time) tea.Msg {
		return tickMsg(t)
	})
}
```

### Add More Weather Details

Extend the `weatherData` struct:

```go
type weatherData struct {
	temperature int
	condition   string
	humidity    int
	windSpeed   int
	pressure    int    // Add barometric pressure
	visibility  int    // Add visibility distance
	uvIndex     int    // Add UV index
}
```

### Change Styling

Customize colors in the `View()` function:

```go
tempStyle := lipgloss.NewStyle().
	Foreground(lipgloss.Color("196")). // Bright red
	Bold(true).
	Padding(0, 1)
```

See [Lip Gloss colors](https://github.com/charmbracelet/lipgloss#colors) for color options.
