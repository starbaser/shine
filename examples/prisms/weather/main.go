// Weather Prism - Display current weather information
//
// This example demonstrates:
// - Periodic data updates (simulated)
// - Horizontal layout with lipgloss
// - Clean styling for panel display
// - Proper window title setting

package main

import (
	"fmt"
	"log"
	"math/rand"
	"time"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

func main() {
	// Set window title for tracking (REQUIRED)
	fmt.Print("\033]0;shine-weather\007")

	// Create Bubble Tea program without alt screen
	p := tea.NewProgram(initialModel())

	if _, err := p.Run(); err != nil {
		log.Fatal(err)
	}
}

// tickMsg is sent every update interval
type tickMsg time.Time

// weatherData represents weather information
type weatherData struct {
	temperature int
	condition   string
	humidity    int
	windSpeed   int
}

// model holds the application state
type model struct {
	weather    weatherData
	location   string
	lastUpdate time.Time
	width      int
	height     int
}

// initialModel creates the initial application state
func initialModel() model {
	return model{
		weather: weatherData{
			temperature: 72,
			condition:   "Sunny",
			humidity:    45,
			windSpeed:   5,
		},
		location:   "San Francisco",
		lastUpdate: time.Now(),
		width:      80,
		height:     1,
	}
}

// Init returns the initial command
func (m model) Init() tea.Cmd {
	return tickCmd()
}

// tickCmd creates a command that ticks every 5 minutes
func tickCmd() tea.Cmd {
	return tea.Tick(5*time.Minute, func(t time.Time) tea.Msg {
		return tickMsg(t)
	})
}

// Update handles messages and updates the model
func (m model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		m.width = msg.Width
		m.height = msg.Height
		return m, nil

	case tea.KeyMsg:
		switch msg.Type {
		case tea.KeyEsc, tea.KeyCtrlC:
			return m, tea.Quit
		}

	case tickMsg:
		// Simulate weather data update
		m.weather = fetchWeatherData()
		m.lastUpdate = time.Time(msg)
		return m, tickCmd()
	}

	return m, nil
}

// fetchWeatherData simulates fetching weather data from an API
// In a real prism, replace this with actual API calls
func fetchWeatherData() weatherData {
	conditions := []string{"Sunny", "Cloudy", "Rainy", "Foggy", "Windy"}
	return weatherData{
		temperature: 65 + rand.Intn(20), // 65-84°F
		condition:   conditions[rand.Intn(len(conditions))],
		humidity:    30 + rand.Intn(40), // 30-70%
		windSpeed:   rand.Intn(15),      // 0-15 mph
	}
}

// View renders the UI
func (m model) View() string {
	// Styles with high contrast for panel visibility
	locationStyle := lipgloss.NewStyle().
		Foreground(lipgloss.Color("12")). // Bright blue
		Bold(true).
		Padding(0, 1)

	conditionStyle := lipgloss.NewStyle().
		Foreground(lipgloss.Color("11")). // Bright yellow
		Padding(0, 1)

	tempStyle := lipgloss.NewStyle().
		Foreground(lipgloss.Color("13")). // Bright magenta
		Bold(true).
		Padding(0, 1)

	detailStyle := lipgloss.NewStyle().
		Foreground(lipgloss.Color("14")). // Bright cyan
		Padding(0, 1)

	// Get weather icon based on condition
	icon := getWeatherIcon(m.weather.condition)

	// Build display components
	location := locationStyle.Render(m.location)
	condition := conditionStyle.Render(fmt.Sprintf("%s %s", icon, m.weather.condition))
	temp := tempStyle.Render(fmt.Sprintf("%d°F", m.weather.temperature))
	details := detailStyle.Render(fmt.Sprintf("💧%d%% 💨%dmph", m.weather.humidity, m.weather.windSpeed))

	// Combine all elements horizontally
	return lipgloss.JoinHorizontal(
		lipgloss.Top,
		location,
		condition,
		temp,
		details,
	)
}

// getWeatherIcon returns an emoji icon for the weather condition
func getWeatherIcon(condition string) string {
	icons := map[string]string{
		"Sunny":  "☀️",
		"Cloudy": "☁️",
		"Rainy":  "🌧️",
		"Foggy":  "🌫️",
		"Windy":  "💨",
	}
	if icon, ok := icons[condition]; ok {
		return icon
	}
	return "🌡️"
}
