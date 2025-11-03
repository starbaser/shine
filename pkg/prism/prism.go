package prism

import (
	"os/exec"

	"github.com/starbased-co/shine/pkg/config"
)

// Prism represents a self-contained widget (bar, clock, weather, etc.)
// All widgets are treated as prisms - no distinction between built-in and user prisms
type Prism struct {
	Name     string             // Prism name (e.g., "bar", "weather")
	Binary   string             // Full path to executable
	Config   *config.PrismConfig // Configuration from TOML
	Process  *exec.Cmd          // Runtime process handle
	WindowID string             // Hyprland window ID (if available)
}

// Launch starts the prism process
func (p *Prism) Launch() error {
	if p.Process != nil && p.Process.Process != nil {
		return nil // Already running
	}

	cmd := exec.Command(p.Binary)
	if err := cmd.Start(); err != nil {
		return err
	}

	p.Process = cmd
	return nil
}

// Stop terminates the prism process
func (p *Prism) Stop() error {
	if p.Process == nil || p.Process.Process == nil {
		return nil // Not running
	}

	return p.Process.Process.Kill()
}

// IsRunning checks if the prism process is running
func (p *Prism) IsRunning() bool {
	if p.Process == nil || p.Process.Process == nil {
		return false
	}

	// Check if process still exists
	return p.Process.ProcessState == nil || !p.Process.ProcessState.Exited()
}
