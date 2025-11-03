// Example Shine Prism: Weather Widget
//
// This demonstrates the minimal interface required for a Shine prism.
// Prisms are standard Bubble Tea applications that follow specific conventions.
//
// A prism is a self-contained unit that refracts light (Shine) to display information.

package main

import (
	"fmt"
	"log"
	"time"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

func main() {
	// Set window title for tracking (recommended but optional)
	// This allows Shine to identify the window via Kitty remote control
	fmt.Print("\033]0;shine-weather\007")

	// Create Bubble Tea program
	// IMPORTANT: Do NOT use tea.WithAltScreen() for panel widgets
	// Alt screen mode is for full-screen TUIs and breaks panel rendering
	p := tea.NewProgram(initialModel())

	if _, err := p.Run(); err != nil {
		log.Fatal(err)
	}
}

// tickMsg is sent every update interval
type tickMsg time.Time

// model holds the application state
type model struct {
	temperature int
	location    string
	condition   string
	lastUpdate  time.Time
	width       int
	err         error
}

// initialModel creates the initial application state
func initialModel() model {
	return model{
		temperature: 72,
		location:    "San Francisco",
		condition:   "Sunny",
		lastUpdate:  time.Now(),
		width:       80,
	}
}

// Init returns the initial command (start ticker)
func (m model) Init() tea.Cmd {
	return tickCmd()
}

// tickCmd creates a command that ticks every 15 minutes
func tickCmd() tea.Cmd {
	return tea.Tick(15*time.Minute, func(t time.Time) tea.Msg {
		return tickMsg(t)
	})
}

// Update handles messages and updates the model
func (m model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		// Handle terminal resize
		m.width = msg.Width
		return m, nil

	case tea.KeyMsg:
		// Handle keyboard input
		switch msg.Type {
		case tea.KeyEsc, tea.KeyCtrlC:
			return m, tea.Quit
		}

	case tickMsg:
		// Update weather data
		m.lastUpdate = time.Time(msg)
		// In a real prism, fetch weather data from API here
		// For demo, just cycle through conditions
		conditions := []string{"Sunny", "Cloudy", "Rainy", "Foggy"}
		m.condition = conditions[int(m.lastUpdate.Unix())%len(conditions)]
		return m, tickCmd()
	}

	return m, nil
}

// View renders the UI
func (m model) View() string {
	// Styles with high contrast for visibility in thin panels
	// Use bright colors on dark background for readability
	mainStyle := lipgloss.NewStyle().
		Foreground(lipgloss.Color("14")). // Bright cyan
		Background(lipgloss.Color("0")).  // Black background
		Bold(true).
		Padding(0, 1)

	conditionStyle := lipgloss.NewStyle().
		Foreground(lipgloss.Color("11")). // Bright yellow
		Background(lipgloss.Color("0")).
		Padding(0, 1)

	tempStyle := lipgloss.NewStyle().
		Foreground(lipgloss.Color("13")). // Bright magenta
		Background(lipgloss.Color("0")).
		Bold(true).
		Padding(0, 1)

	// Compose the weather display
	locationView := mainStyle.Render(m.location)
	conditionView := conditionStyle.Render(m.condition)
	tempView := tempStyle.Render(fmt.Sprintf("%d°F", m.temperature))

	// Combine all parts horizontally
	return lipgloss.JoinHorizontal(
		lipgloss.Top,
		locationView,
		conditionView,
		tempView,
	)
}
