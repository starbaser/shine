package main

import (
	"fmt"
	"os"

	"github.com/starbased-co/shine/cmd/prisms/internal/theme"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

// model holds the Bubble Tea application state
type model struct {
	width  int
	height int
}

func initialModel() model {
	return model{}
}

func (m model) Init() tea.Cmd {
	return nil
}

func (m model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		if msg.String() == "q" || msg.String() == "ctrl+c" {
			return m, tea.Quit
		}

	case tea.WindowSizeMsg:
		m.width = msg.Width
		m.height = msg.Height
	}

	return m, nil
}

func (m model) View() string {
	t := theme.Current()

	// Header
	header := lipgloss.NewStyle().
		Foreground(t.Primary()).
		Background(t.Surface1()).
		Bold(true).
		Width(m.width).
		Align(lipgloss.Center).
		Padding(1).
		Render(fmt.Sprintf("%s System Monitor", theme.IconDashboard))

	// System stats panel
	stats := []string{
		fmt.Sprintf("%s CPU: 45%%", theme.IconCPU),
		fmt.Sprintf("%s Memory: 8.2 GB / 16 GB", theme.IconMemory),
		fmt.Sprintf("%s Disk: 120 GB free", theme.IconDisk),
		fmt.Sprintf("%s Network: Connected", theme.IconNetwork),
	}

	statsPanel := theme.CardStyle().
		Width(m.width - 4).
		Render(lipgloss.JoinVertical(
			lipgloss.Left,
			theme.SubtitleStyle().Render("System Status"),
			"",
			lipgloss.JoinVertical(lipgloss.Left, stats...),
		))

	// Battery status with dynamic icon
	batteryPercent := 75
	batteryIcon := theme.GetBatteryIcon(batteryPercent, false)
	batteryStyle := lipgloss.NewStyle().Foreground(t.Success())
	if batteryPercent < 20 {
		batteryStyle = lipgloss.NewStyle().Foreground(t.Error())
	} else if batteryPercent < 50 {
		batteryStyle = lipgloss.NewStyle().Foreground(t.Warning())
	}

	battery := batteryStyle.Render(fmt.Sprintf("%s %d%%", batteryIcon, batteryPercent))

	// Volume status with dynamic icon
	volumePercent := 80
	volumeMuted := false
	volumeIcon := theme.GetVolumeIcon(volumePercent, volumeMuted)
	volume := lipgloss.NewStyle().
		Foreground(t.Info()).
		Render(fmt.Sprintf("%s %d%%", volumeIcon, volumePercent))

	// Status bar
	statusLeft := lipgloss.NewStyle().
		Foreground(t.TextPrimary()).
		Render(fmt.Sprintf("%s 15:30", theme.IconClock))

	statusRight := lipgloss.JoinHorizontal(
		lipgloss.Left,
		battery,
		"  ",
		volume,
	)

	statusBar := theme.StatusBarStyle().
		Width(m.width).
		Render(lipgloss.JoinHorizontal(
			lipgloss.Left,
			statusLeft,
			lipgloss.NewStyle().
				Width(m.width-lipgloss.Width(statusLeft)-lipgloss.Width(statusRight)).
				Render(""),
			statusRight,
		))

	// Help text
	help := theme.MutedTextStyle().
		Width(m.width).
		Align(lipgloss.Center).
		Render("Press q to quit")

	// Compose the view
	content := lipgloss.JoinVertical(
		lipgloss.Left,
		header,
		"",
		statsPanel,
		"",
		help,
	)

	// Add padding to ensure status bar is at bottom
	availableHeight := m.height - lipgloss.Height(content) - lipgloss.Height(statusBar)
	if availableHeight > 0 {
		content += lipgloss.NewStyle().Height(availableHeight).Render("")
	}

	return lipgloss.JoinVertical(
		lipgloss.Left,
		content,
		statusBar,
	)
}

func main() {
	p := tea.NewProgram(
		initialModel(),
		tea.WithAltScreen(),
		tea.WithMouseCellMotion(),
	)

	if _, err := p.Run(); err != nil {
		fmt.Printf("Error: %v\n", err)
		os.Exit(1)
	}
}
