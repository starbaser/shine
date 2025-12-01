package main

import (
	"encoding/json"
	"fmt"
	"log"
	"os"
	"os/exec"
	"strconv"
	"strings"
	"time"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/starbased-co/shine/cmd/prisms/internal/theme"
)

func main() {
	// Set window title using ANSI escape sequence
	fmt.Print("\033]0;shine-bar\007")

	// Note: Don't use tea.WithAltScreen() for thin status bars
	// Alt screen mode is for full-screen TUIs, causes rendering issues in panels
	p := tea.NewProgram(initialModel())

	if _, err := p.Run(); err != nil {
		log.Fatal(err)
	}
}

type workspace struct {
	ID     int    `json:"id"`
	Name   string `json:"name"`
	Active bool   `json:"active"`
}

type activeWorkspace struct {
	ID   int    `json:"id"`
	Name string `json:"name"`
}

type activeWindow struct {
	Class string `json:"class"`
	Title string `json:"title"`
}

type batteryStatus struct {
	percent  int
	charging bool
	exists   bool
}

type networkStatus struct {
	connected bool
}

type tickMsg time.Time

type model struct {
	workspaces        []workspace
	activeWorkspaceID int
	activeWindow      activeWindow
	battery           batteryStatus
	network           networkStatus
	currentTime       time.Time
	width             int
	err               error
}

func initialModel() model {
	workspaces, activeID := getWorkspaces()
	return model{
		workspaces:        workspaces,
		activeWorkspaceID: activeID,
		activeWindow:      getActiveWindow(),
		battery:           getBatteryStatus(),
		network:           getNetworkStatus(),
		currentTime:       time.Now(),
		width:             80,
	}
}

func (m model) Init() tea.Cmd {
	return tea.Batch(
		tickCmd(),
		refreshStatusCmd(),
	)
}

func tickCmd() tea.Cmd {
	return tea.Tick(time.Second, func(t time.Time) tea.Msg {
		return tickMsg(t)
	})
}

type statusUpdateMsg struct {
	workspaces   []workspace
	activeID     int
	activeWindow activeWindow
	battery      batteryStatus
	network      networkStatus
}

func refreshStatusCmd() tea.Cmd {
	return func() tea.Msg {
		workspaces, activeID := getWorkspaces()
		return statusUpdateMsg{
			workspaces:   workspaces,
			activeID:     activeID,
			activeWindow: getActiveWindow(),
			battery:      getBatteryStatus(),
			network:      getNetworkStatus(),
		}
	}
}

func (m model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		m.width = msg.Width
		return m, nil

	case tea.KeyMsg:
		switch msg.Type {
		case tea.KeyEsc:
			return m, tea.Quit
		}

	case tickMsg:
		m.currentTime = time.Time(msg)
		return m, tea.Batch(
			tickCmd(),
			refreshStatusCmd(),
		)

	case statusUpdateMsg:
		m.workspaces = msg.workspaces
		m.activeWorkspaceID = msg.activeID
		m.activeWindow = msg.activeWindow
		m.battery = msg.battery
		m.network = msg.network
		return m, nil
	}

	return m, nil
}

func (m model) View() string {
	t := theme.Current()

	// Workspace styles
	activeStyle := lipgloss.NewStyle().
		Foreground(t.Accent()).
		Background(t.Background()).
		Bold(true).
		Padding(0, 1)

	inactiveStyle := lipgloss.NewStyle().
		Foreground(t.TextMuted()).
		Background(t.Background()).
		Padding(0, 1)

	// Window title style
	windowTitleStyle := lipgloss.NewStyle().
		Foreground(t.TextSecondary()).
		Background(t.Background()).
		Padding(0, 1)

	// Clock style
	clockStyle := lipgloss.NewStyle().
		Foreground(t.Accent()).
		Background(t.Background()).
		Bold(true).
		Padding(0, 1)

	// Build workspaces section (left)
	var workspaceStrs []string
	for _, ws := range m.workspaces {
		wsLabel := fmt.Sprintf("%d", ws.ID)
		if ws.ID == m.activeWorkspaceID {
			workspaceStrs = append(workspaceStrs, activeStyle.Render(wsLabel))
		} else {
			workspaceStrs = append(workspaceStrs, inactiveStyle.Render(wsLabel))
		}
	}
	workspacesView := strings.Join(workspaceStrs, "")

	// Build center section (window title)
	windowTitle := formatWindowTitle(m.activeWindow, m.width/3)
	centerView := windowTitleStyle.Render(windowTitle)

	// Build right section (network, battery, clock)
	var rightParts []string

	// Network indicator
	if m.network.connected {
		netStyle := lipgloss.NewStyle().
			Foreground(t.Success()).
			Background(t.Background()).
			Padding(0, 1)
		rightParts = append(rightParts, netStyle.Render(theme.IconWifi))
	} else {
		netStyle := lipgloss.NewStyle().
			Foreground(t.Error()).
			Background(t.Background()).
			Padding(0, 1)
		rightParts = append(rightParts, netStyle.Render(theme.IconWifiOff))
	}

	// Battery indicator (only if exists)
	if m.battery.exists {
		batteryColor := getBatteryColor(t, m.battery.percent, m.battery.charging)
		batteryIcon := theme.GetBatteryIcon(m.battery.percent, m.battery.charging)
		batteryText := fmt.Sprintf("%s %d%%", batteryIcon, m.battery.percent)

		batteryStyle := lipgloss.NewStyle().
			Foreground(batteryColor).
			Background(t.Background()).
			Padding(0, 1)
		rightParts = append(rightParts, batteryStyle.Render(batteryText))
	}

	// Clock
	rightParts = append(rightParts, clockStyle.Render(m.currentTime.Format("15:04:05")))
	rightView := strings.Join(rightParts, "")

	// Calculate spacing
	leftWidth := lipgloss.Width(workspacesView)
	centerWidth := lipgloss.Width(centerView)
	rightWidth := lipgloss.Width(rightView)

	// Calculate padding to center the window title
	availableSpace := m.width - leftWidth - centerWidth - rightWidth
	if availableSpace < 0 {
		availableSpace = 0
	}

	leftPadding := availableSpace / 2
	rightPadding := availableSpace - leftPadding

	return lipgloss.JoinHorizontal(
		lipgloss.Top,
		workspacesView,
		strings.Repeat(" ", leftPadding),
		centerView,
		strings.Repeat(" ", rightPadding),
		rightView,
	)
}

func getWorkspaces() ([]workspace, int) {
	cmd := exec.Command("hyprctl", "workspaces", "-j")
	output, err := cmd.Output()
	if err != nil {
		return []workspace{{ID: 1, Name: "1"}}, 1
	}

	var workspaces []workspace
	if err := json.Unmarshal(output, &workspaces); err != nil {
		return []workspace{{ID: 1, Name: "1"}}, 1
	}

	cmd = exec.Command("hyprctl", "activeworkspace", "-j")
	output, err = cmd.Output()
	if err != nil {
		if len(workspaces) > 0 {
			return workspaces, workspaces[0].ID
		}
		return []workspace{{ID: 1, Name: "1"}}, 1
	}

	var active activeWorkspace
	if err := json.Unmarshal(output, &active); err != nil {
		if len(workspaces) > 0 {
			return workspaces, workspaces[0].ID
		}
		return []workspace{{ID: 1, Name: "1"}}, 1
	}

	return workspaces, active.ID
}

func getActiveWindow() activeWindow {
	cmd := exec.Command("hyprctl", "activewindow", "-j")
	output, err := cmd.Output()
	if err != nil {
		return activeWindow{}
	}

	var window activeWindow
	if err := json.Unmarshal(output, &window); err != nil {
		return activeWindow{}
	}

	return window
}

func formatWindowTitle(window activeWindow, maxWidth int) string {
	if window.Title == "" && window.Class == "" {
		return ""
	}

	var title string
	if window.Class != "" && window.Title != "" {
		title = fmt.Sprintf("%s - %s", window.Class, window.Title)
	} else if window.Title != "" {
		title = window.Title
	} else {
		title = window.Class
	}

	// Truncate if too long
	if len(title) > maxWidth && maxWidth > 3 {
		title = title[:maxWidth-3] + "..."
	}

	return title
}

func getBatteryStatus() batteryStatus {
	// Check if battery exists
	if _, err := os.Stat("/sys/class/power_supply/BAT0"); os.IsNotExist(err) {
		return batteryStatus{exists: false}
	}

	// Read capacity
	capacityData, err := os.ReadFile("/sys/class/power_supply/BAT0/capacity")
	if err != nil {
		return batteryStatus{exists: false}
	}

	percent, err := strconv.Atoi(strings.TrimSpace(string(capacityData)))
	if err != nil {
		return batteryStatus{exists: false}
	}

	// Read status
	statusData, err := os.ReadFile("/sys/class/power_supply/BAT0/status")
	if err != nil {
		return batteryStatus{exists: false}
	}

	status := strings.TrimSpace(string(statusData))
	charging := status == "Charging"

	return batteryStatus{
		percent:  percent,
		charging: charging,
		exists:   true,
	}
}

func getBatteryColor(t theme.Theme, percent int, charging bool) lipgloss.TerminalColor {
	if charging {
		return t.Info()
	}

	switch {
	case percent > 50:
		return t.Success()
	case percent >= 20:
		return t.Warning()
	default:
		return t.Error()
	}
}

func getNetworkStatus() networkStatus {
	// Simple check: try nmcli first
	cmd := exec.Command("nmcli", "-t", "-f", "STATE", "general")
	output, err := cmd.Output()
	if err == nil {
		state := strings.TrimSpace(string(output))
		return networkStatus{connected: state == "connected"}
	}

	// Fallback: check if any network interface is up (excluding loopback)
	cmd = exec.Command("ip", "link", "show")
	output, err = cmd.Output()
	if err != nil {
		return networkStatus{connected: false}
	}

	lines := strings.Split(string(output), "\n")
	for _, line := range lines {
		if strings.Contains(line, "state UP") && !strings.Contains(line, "lo:") {
			return networkStatus{connected: true}
		}
	}

	return networkStatus{connected: false}
}
