# System Monitor Prism Example

Displays real-time system resource usage with progress bars.

## Features

- CPU usage percentage with visual bar
- Memory usage (percentage and bytes)
- Disk usage (percentage and bytes)
- Network traffic (RX/TX bytes)
- System uptime
- Goroutine count
- Updates every 2 seconds
- Vertical layout optimized for side panels

## Building

```bash
cd examples/prisms/sysmonitor
make build
```

## Installation

```bash
make install
```

This installs `shine-sysmonitor` to `~/.local/bin/`.

## Configuration

Add to `~/.config/shine/shine.toml`:

```toml
[prisms.sysmonitor]
enabled = true
edge = "top-left"
columns_pixels = 300
lines_pixels = 150
margin_top = 40
margin_left = 10
focus_policy = "not-allowed"
```

## Real System Metrics

The example uses mock data for some metrics. For real system monitoring, integrate with system libraries:

### Option 1: gopsutil Library

Install the cross-platform system library:

```bash
go get github.com/shirou/gopsutil/v3
```

Replace `getSystemMetrics()` with real data:

```go
import (
	"github.com/shirou/gopsutil/v3/cpu"
	"github.com/shirou/gopsutil/v3/mem"
	"github.com/shirou/gopsutil/v3/disk"
	"github.com/shirou/gopsutil/v3/net"
	"github.com/shirou/gopsutil/v3/host"
)

func getSystemMetrics() systemMetrics {
	// CPU
	cpuPercents, _ := cpu.Percent(0, false)
	cpuPercent := 0.0
	if len(cpuPercents) > 0 {
		cpuPercent = cpuPercents[0]
	}

	// Memory
	memInfo, _ := mem.VirtualMemory()

	// Disk
	diskInfo, _ := disk.Usage("/")

	// Network
	netStats, _ := net.IOCounters(false)
	netRx := uint64(0)
	netTx := uint64(0)
	if len(netStats) > 0 {
		netRx = netStats[0].BytesRecv
		netTx = netStats[0].BytesSent
	}

	// Host info
	hostInfo, _ := host.Info()

	return systemMetrics{
		cpuPercent:  cpuPercent,
		memUsed:     memInfo.Used,
		memTotal:    memInfo.Total,
		diskUsed:    diskInfo.Used,
		diskTotal:   diskInfo.Total,
		netRxBytes:  netRx,
		netTxBytes:  netTx,
		goroutines:  runtime.NumGoroutine(),
		uptime:      time.Duration(hostInfo.Uptime) * time.Second,
	}
}
```

### Option 2: Direct System Calls (Linux)

Read from `/proc` filesystem:

```go
import (
	"os"
	"strconv"
	"strings"
)

func getCPUPercent() float64 {
	data, _ := os.ReadFile("/proc/stat")
	lines := strings.Split(string(data), "\n")
	for _, line := range lines {
		if strings.HasPrefix(line, "cpu ") {
			fields := strings.Fields(line)
			// Parse CPU times and calculate percentage
			// ...
		}
	}
	return 0
}

func getMemoryInfo() (used, total uint64) {
	data, _ := os.ReadFile("/proc/meminfo")
	lines := strings.Split(string(data), "\n")
	var memTotal, memAvailable uint64
	for _, line := range lines {
		if strings.HasPrefix(line, "MemTotal:") {
			fields := strings.Fields(line)
			memTotal, _ = strconv.ParseUint(fields[1], 10, 64)
			memTotal *= 1024 // Convert to bytes
		} else if strings.HasPrefix(line, "MemAvailable:") {
			fields := strings.Fields(line)
			memAvailable, _ = strconv.ParseUint(fields[1], 10, 64)
			memAvailable *= 1024
		}
	}
	return memTotal - memAvailable, memTotal
}
```

## Testing

```bash
# Run standalone
make run

# Test with kitty panel
kitten panel --edge=top-left --columns=300px --lines=150px \
	--margin-top=40 --margin-left=10 \
	./shine-sysmonitor
```

## Customization

### Change Update Frequency

Edit `tickCmd()` in `main.go`:

```go
func tickCmd() tea.Cmd {
	return tea.Tick(5*time.Second, func(t time.Time) tea.Msg {
		return tickMsg(t)
	})
}
```

### Add More Metrics

Extend the `systemMetrics` struct:

```go
type systemMetrics struct {
	cpuPercent    float64
	memUsed       uint64
	memTotal      uint64
	diskUsed      uint64
	diskTotal     uint64
	netRxBytes    uint64
	netTxBytes    uint64
	goroutines    int
	uptime        time.Duration
	// Add new metrics
	loadAvg1      float64   // 1-minute load average
	loadAvg5      float64   // 5-minute load average
	loadAvg15     float64   // 15-minute load average
	temperature   float64   // CPU temperature
	batteryPct    float64   // Battery percentage (laptops)
}
```

Then update the `View()` function to display them.

### Change Layout

For horizontal layout (better for top/bottom edges):

```go
func (m model) View() string {
	// ... build sections as before ...

	// Join horizontally instead
	return lipgloss.JoinHorizontal(lipgloss.Top, sections...)
}
```

### Customize Colors

Change the color scheme:

```go
labelStyle := lipgloss.NewStyle().
	Foreground(lipgloss.Color("220")). // Orange
	Width(12)

valueStyle := lipgloss.NewStyle().
	Foreground(lipgloss.Color("118")). // Light green
```

See [Lip Gloss colors](https://github.com/charmbracelet/lipgloss#colors) for available colors.

### Add Alerts

Add visual alerts for high resource usage:

```go
func (m model) View() string {
	// ...

	// Warn if CPU is high
	if m.metrics.cpuPercent > 80 {
		warningStyle := lipgloss.NewStyle().
			Foreground(lipgloss.Color("196")). // Red
			Bold(true)
		sections = append(sections, warningStyle.Render("⚠ High CPU Usage!"))
	}

	// ...
}
```
