// System Monitor Prism - Display system resource usage
//
// This example demonstrates:
// - Real-time system metrics
// - Vertical layout with tables
// - Progress bars for resource usage
// - Fast refresh rate (2 seconds)

package main

import (
	"fmt"
	"log"
	"math/rand"
	"runtime"
	"time"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

func main() {
	// Set window title for tracking (REQUIRED)
	fmt.Print("\033]0;shine-sysmonitor\007")

	// Create Bubble Tea program without alt screen
	p := tea.NewProgram(initialModel())

	if _, err := p.Run(); err != nil {
		log.Fatal(err)
	}
}

// tickMsg is sent every 2 seconds
type tickMsg time.Time

// systemMetrics holds system resource usage
type systemMetrics struct {
	cpuPercent float64
	memUsed    uint64
	memTotal   uint64
	diskUsed   uint64
	diskTotal  uint64
	netRxBytes uint64
	netTxBytes uint64
	goroutines int
	uptime     time.Duration
}

// model holds the application state
type model struct {
	metrics    systemMetrics
	lastUpdate time.Time
	width      int
	height     int
	startTime  time.Time
}

// initialModel creates the initial application state
func initialModel() model {
	return model{
		metrics:    getSystemMetrics(),
		lastUpdate: time.Now(),
		startTime:  time.Now(),
		width:      80,
		height:     24,
	}
}

// Init returns the initial command
func (m model) Init() tea.Cmd {
	return tickCmd()
}

// tickCmd creates a command that ticks every 2 seconds
func tickCmd() tea.Cmd {
	return tea.Tick(2*time.Second, func(t time.Time) tea.Msg {
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
		// Update system metrics
		m.metrics = getSystemMetrics()
		m.metrics.uptime = time.Since(m.startTime)
		m.lastUpdate = time.Time(msg)
		return m, tickCmd()
	}

	return m, nil
}

// getSystemMetrics fetches current system metrics
// This is a simplified version - use proper system libraries for production
func getSystemMetrics() systemMetrics {
	var mem runtime.MemStats
	runtime.ReadMemStats(&mem)

	// Simulate CPU and other metrics (replace with real system calls)
	return systemMetrics{
		cpuPercent: float64(rand.Intn(100)),    // Mock CPU
		memUsed:    mem.Alloc,                  // Real Go heap
		memTotal:   16 * 1024 * 1024 * 1024,    // Mock: 16GB
		diskUsed:   500 * 1024 * 1024 * 1024,   // Mock: 500GB
		diskTotal:  1000 * 1024 * 1024 * 1024,  // Mock: 1TB
		netRxBytes: uint64(rand.Intn(1000000)), // Mock: network RX
		netTxBytes: uint64(rand.Intn(500000)),  // Mock: network TX
		goroutines: runtime.NumGoroutine(),     // Real goroutines
	}
}

// View renders the UI
func (m model) View() string {
	// Styles
	headerStyle := lipgloss.NewStyle().
		Foreground(lipgloss.Color("12")). // Bright blue
		Bold(true).
		Padding(0, 1)

	labelStyle := lipgloss.NewStyle().
		Foreground(lipgloss.Color("14")). // Bright cyan
		Width(12)

	valueStyle := lipgloss.NewStyle().
		Foreground(lipgloss.Color("10")) // Bright green

	barStyle := lipgloss.NewStyle().
		Foreground(lipgloss.Color("13")) // Bright magenta

	// Build display sections
	sections := []string{
		headerStyle.Render("System Monitor"),
	}

	// CPU
	cpuBar := renderBar(m.metrics.cpuPercent, 100, 20)
	cpuLine := labelStyle.Render("CPU:") + " " +
		valueStyle.Render(fmt.Sprintf("%5.1f%%", m.metrics.cpuPercent)) + " " +
		barStyle.Render(cpuBar)
	sections = append(sections, cpuLine)

	// Memory
	memPercent := float64(m.metrics.memUsed) / float64(m.metrics.memTotal) * 100
	memBar := renderBar(memPercent, 100, 20)
	memLine := labelStyle.Render("Memory:") + " " +
		valueStyle.Render(fmt.Sprintf("%5.1f%%", memPercent)) + " " +
		barStyle.Render(memBar) + " " +
		valueStyle.Render(fmt.Sprintf("%s / %s",
			formatBytes(m.metrics.memUsed),
			formatBytes(m.metrics.memTotal)))
	sections = append(sections, memLine)

	// Disk
	diskPercent := float64(m.metrics.diskUsed) / float64(m.metrics.diskTotal) * 100
	diskBar := renderBar(diskPercent, 100, 20)
	diskLine := labelStyle.Render("Disk:") + " " +
		valueStyle.Render(fmt.Sprintf("%5.1f%%", diskPercent)) + " " +
		barStyle.Render(diskBar) + " " +
		valueStyle.Render(fmt.Sprintf("%s / %s",
			formatBytes(m.metrics.diskUsed),
			formatBytes(m.metrics.diskTotal)))
	sections = append(sections, diskLine)

	// Network
	netLine := labelStyle.Render("Network:") + " " +
		valueStyle.Render(fmt.Sprintf("↓ %s  ↑ %s",
			formatBytes(m.metrics.netRxBytes),
			formatBytes(m.metrics.netTxBytes)))
	sections = append(sections, netLine)

	// Uptime & Goroutines
	infoLine := labelStyle.Render("Info:") + " " +
		valueStyle.Render(fmt.Sprintf("Uptime: %s  Goroutines: %d",
			formatDuration(m.metrics.uptime),
			m.metrics.goroutines))
	sections = append(sections, infoLine)

	// Join all sections vertically
	return lipgloss.JoinVertical(lipgloss.Left, sections...)
}

// renderBar creates a text-based progress bar
func renderBar(value, max float64, width int) string {
	if max == 0 {
		return ""
	}

	percent := value / max
	filled := int(percent * float64(width))
	if filled > width {
		filled = width
	}

	bar := ""
	for i := 0; i < width; i++ {
		if i < filled {
			bar += "█"
		} else {
			bar += "░"
		}
	}

	return bar
}

// formatBytes formats bytes as human-readable size
func formatBytes(bytes uint64) string {
	const unit = 1024
	if bytes < unit {
		return fmt.Sprintf("%d B", bytes)
	}
	div, exp := uint64(unit), 0
	for n := bytes / unit; n >= unit; n /= unit {
		div *= unit
		exp++
	}
	sizes := []string{"KB", "MB", "GB", "TB"}
	return fmt.Sprintf("%.1f %s", float64(bytes)/float64(div), sizes[exp])
}

// formatDuration formats a duration as human-readable time
func formatDuration(d time.Duration) string {
	hours := int(d.Hours())
	minutes := int(d.Minutes()) % 60
	seconds := int(d.Seconds()) % 60

	if hours > 0 {
		return fmt.Sprintf("%dh%dm%ds", hours, minutes, seconds)
	} else if minutes > 0 {
		return fmt.Sprintf("%dm%ds", minutes, seconds)
	}
	return fmt.Sprintf("%ds", seconds)
}
