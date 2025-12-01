package main

import (
	"fmt"
	"log"
	"os"
	"strconv"
	"strings"
	"time"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/bubbles/progress"
	"github.com/charmbracelet/lipgloss"
	"github.com/starbased-co/shine/cmd/prisms/internal/theme"
)

func main() {
	// Set window title using ANSI escape sequence
	fmt.Print("\033]0;shine-sysinfo\007")

	// Use alt screen mode to take over the full terminal
	p := tea.NewProgram(
		initialModel(),
		tea.WithAltScreen(),
	)

	if _, err := p.Run(); err != nil {
		log.Fatal(err)
	}
}

type tickMsg time.Time

type sysInfo struct {
	hostname   string
	uptime     string
	cpuUsage   float64 // 0.0-1.0
	memUsage   float64 // 0.0-1.0
	lastUpdate time.Time
}

type model struct {
	info      sysInfo
	width     int
	height    int
	cpuBar    progress.Model
	memBar    progress.Model
	prevCPU   cpuStat
}

type cpuStat struct {
	user   uint64
	system uint64
	idle   uint64
}

func initialModel() model {
	// Initialize progress bars
	cpuBar := progress.New(
		progress.WithDefaultGradient(),
		progress.WithWidth(20),
	)
	memBar := progress.New(
		progress.WithDefaultGradient(),
		progress.WithWidth(20),
	)

	m := model{
		info:   getSysInfo(),
		width:  30,
		height: 10,
		cpuBar: cpuBar,
		memBar: memBar,
	}

	// Get initial CPU stat
	m.prevCPU = getCPUStat()

	return m
}

func (m model) Init() tea.Cmd {
	return tickCmd()
}

func tickCmd() tea.Cmd {
	return tea.Tick(2*time.Second, func(t time.Time) tea.Msg {
		return tickMsg(t)
	})
}

func (m model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		m.width = msg.Width
		m.height = msg.Height
		return m, nil

	case tea.KeyMsg:
		switch msg.Type {
		case tea.KeyEsc:
			return m, tea.Quit
		}

	case tickMsg:
		// Calculate CPU usage
		currentCPU := getCPUStat()
		m.info.cpuUsage = calculateCPUUsage(m.prevCPU, currentCPU)
		m.prevCPU = currentCPU

		// Update other info
		m.info.hostname = getHostname()
		m.info.uptime = getUptime()
		m.info.memUsage = getMemoryUsage()
		m.info.lastUpdate = time.Time(msg)

		return m, tickCmd()
	}

	return m, nil
}

func (m model) View() string {
	// Create title
	title := theme.TitleStyle().Render("sysinfo")

	// Create info rows with icons
	hostLabel := theme.TextStyle().Bold(true).Render("Host:")
	hostValue := theme.TextStyle().Render(m.info.hostname)
	hostRow := lipgloss.JoinHorizontal(lipgloss.Left, hostLabel, " ", hostValue)

	uptimeLabel := theme.TextStyle().Bold(true).Render("Uptime:")
	uptimeValue := theme.TextStyle().Render(m.info.uptime)
	uptimeRow := lipgloss.JoinHorizontal(lipgloss.Left, uptimeLabel, " ", uptimeValue)

	// CPU progress bar with dynamic color
	cpuColor := getStatusColor(m.info.cpuUsage)
	cpuBar := m.cpuBar.ViewAs(m.info.cpuUsage)
	cpuPercent := theme.TextStyle().Foreground(cpuColor).Render(fmt.Sprintf("%.0f%%", m.info.cpuUsage*100))
	cpuIcon := theme.TextStyle().Render(theme.IconCPU)
	cpuLabel := theme.TextStyle().Bold(true).Render("CPU:")
	cpuRow := lipgloss.JoinHorizontal(lipgloss.Left,
		cpuIcon, " ", cpuLabel, " ", cpuBar, " ", cpuPercent,
	)

	// Memory progress bar with dynamic color
	memColor := getStatusColor(m.info.memUsage)
	memBar := m.memBar.ViewAs(m.info.memUsage)
	memPercent := theme.TextStyle().Foreground(memColor).Render(fmt.Sprintf("%.0f%%", m.info.memUsage*100))
	memIcon := theme.TextStyle().Render(theme.IconMemory)
	memLabel := theme.TextStyle().Bold(true).Render("RAM:")
	memRow := lipgloss.JoinHorizontal(lipgloss.Left,
		memIcon, " ", memLabel, " ", memBar, " ", memPercent,
	)

	// Last update timestamp
	updateText := fmt.Sprintf("Updated %ds ago", int(time.Since(m.info.lastUpdate).Seconds()))
	updateRow := theme.MutedTextStyle().Render(updateText)

	// Compose all rows
	content := lipgloss.JoinVertical(
		lipgloss.Left,
		title,
		"",
		hostRow,
		uptimeRow,
		"",
		cpuRow,
		memRow,
		"",
		updateRow,
	)

	// Apply glass panel style
	return theme.GlassPanel().Render(content)
}

// getStatusColor returns appropriate color based on usage percentage
func getStatusColor(usage float64) lipgloss.TerminalColor {
	t := theme.Current()
	switch {
	case usage >= 0.80:
		return t.Error() // Red for >80%
	case usage >= 0.50:
		return t.Warning() // Yellow for 50-80%
	default:
		return t.Success() // Green for <50%
	}
}

func getSysInfo() sysInfo {
	info := sysInfo{
		hostname:   getHostname(),
		uptime:     getUptime(),
		cpuUsage:   0.0,
		memUsage:   getMemoryUsage(),
		lastUpdate: time.Now(),
	}
	return info
}

func getHostname() string {
	hostname, err := os.Hostname()
	if err != nil {
		return "unknown"
	}
	return hostname
}

func getUptime() string {
	data, err := os.ReadFile("/proc/uptime")
	if err != nil {
		return "N/A"
	}

	fields := strings.Fields(string(data))
	if len(fields) < 1 {
		return "N/A"
	}

	seconds, err := strconv.ParseFloat(fields[0], 64)
	if err != nil {
		return "N/A"
	}

	days := int(seconds / 86400)
	hours := int((seconds - float64(days*86400)) / 3600)

	if days > 0 {
		return fmt.Sprintf("%dd %dh", days, hours)
	}
	return fmt.Sprintf("%dh", hours)
}

func getCPUStat() cpuStat {
	data, err := os.ReadFile("/proc/stat")
	if err != nil {
		return cpuStat{}
	}

	lines := strings.Split(string(data), "\n")
	if len(lines) < 1 {
		return cpuStat{}
	}

	// First line: cpu  user nice system idle iowait irq softirq ...
	fields := strings.Fields(lines[0])
	if len(fields) < 5 || fields[0] != "cpu" {
		return cpuStat{}
	}

	user, _ := strconv.ParseUint(fields[1], 10, 64)
	system, _ := strconv.ParseUint(fields[3], 10, 64)
	idle, _ := strconv.ParseUint(fields[4], 10, 64)

	return cpuStat{
		user:   user,
		system: system,
		idle:   idle,
	}
}

func calculateCPUUsage(prev, current cpuStat) float64 {
	prevTotal := prev.user + prev.system + prev.idle
	currentTotal := current.user + current.system + current.idle

	if prevTotal == 0 || currentTotal == prevTotal {
		return 0.0
	}

	totalDiff := currentTotal - prevTotal
	idleDiff := current.idle - prev.idle
	usedDiff := totalDiff - idleDiff

	return float64(usedDiff) / float64(totalDiff)
}

func getMemoryUsage() float64 {
	data, err := os.ReadFile("/proc/meminfo")
	if err != nil {
		return 0.0
	}

	var memTotal, memAvailable uint64

	lines := strings.Split(string(data), "\n")
	for _, line := range lines {
		fields := strings.Fields(line)
		if len(fields) < 2 {
			continue
		}

		value, err := strconv.ParseUint(fields[1], 10, 64)
		if err != nil {
			continue
		}

		switch fields[0] {
		case "MemTotal:":
			memTotal = value
		case "MemAvailable:":
			memAvailable = value
		}
	}

	if memTotal == 0 {
		return 0.0
	}

	memUsed := memTotal - memAvailable
	return float64(memUsed) / float64(memTotal)
}
