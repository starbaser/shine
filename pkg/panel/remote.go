package panel

import (
	"encoding/json"
	"fmt"
	"net"
	"os/exec"
)

// RemoteControl represents a kitty remote control client
type RemoteControl struct {
	socketPath string
}

// NewRemoteControl creates a new remote control client
func NewRemoteControl(socketPath string) *RemoteControl {
	return &RemoteControl{
		socketPath: socketPath,
	}
}

// ToggleVisibility toggles panel visibility via remote control
func (rc *RemoteControl) ToggleVisibility() error {
	// Connect to Unix socket
	conn, err := net.Dial("unix", rc.socketPath)
	if err != nil {
		return fmt.Errorf("failed to connect to %s: %w", rc.socketPath, err)
	}
	defer conn.Close()

	// Build remote command payload
	// Format: {"cmd": "resize-os-window", "action": "toggle-visibility"}
	payload := map[string]interface{}{
		"cmd":    "resize-os-window",
		"action": "toggle-visibility",
	}

	// Encode as JSON
	data, err := json.Marshal(payload)
	if err != nil {
		return fmt.Errorf("failed to encode command: %w", err)
	}

	// Send command
	_, err = conn.Write(data)
	if err != nil {
		return fmt.Errorf("failed to send command: %w", err)
	}

	return nil
}

// Show explicitly shows the panel
func (rc *RemoteControl) Show() error {
	conn, err := net.Dial("unix", rc.socketPath)
	if err != nil {
		return fmt.Errorf("failed to connect to %s: %w", rc.socketPath, err)
	}
	defer conn.Close()

	payload := map[string]interface{}{
		"cmd":    "resize-os-window",
		"action": "show",
	}

	data, err := json.Marshal(payload)
	if err != nil {
		return fmt.Errorf("failed to encode command: %w", err)
	}

	_, err = conn.Write(data)
	if err != nil {
		return fmt.Errorf("failed to send command: %w", err)
	}

	return nil
}

// Hide explicitly hides the panel
func (rc *RemoteControl) Hide() error {
	conn, err := net.Dial("unix", rc.socketPath)
	if err != nil {
		return fmt.Errorf("failed to connect to %s: %w", rc.socketPath, err)
	}
	defer conn.Close()

	payload := map[string]interface{}{
		"cmd":    "resize-os-window",
		"action": "hide",
	}

	data, err := json.Marshal(payload)
	if err != nil {
		return fmt.Errorf("failed to encode command: %w", err)
	}

	_, err = conn.Write(data)
	if err != nil {
		return fmt.Errorf("failed to send command: %w", err)
	}

	return nil
}

// WindowInfo represents information about a Kitty window
type WindowInfo struct {
	ID    int    `json:"id"`
	Title string `json:"title"`
}

// CloseWindow closes a specific window by title using kitty @ close-window
func (rc *RemoteControl) CloseWindow(windowTitle string) error {
	cmd := exec.Command("kitty", "@", "--to", "unix:"+rc.socketPath,
		"close-window", "--match", "title:"+windowTitle)

	if err := cmd.Run(); err != nil {
		return fmt.Errorf("failed to close window %s: %w", windowTitle, err)
	}

	return nil
}

// FocusWindow focuses a specific window by title
func (rc *RemoteControl) FocusWindow(windowTitle string) error {
	cmd := exec.Command("kitty", "@", "--to", "unix:"+rc.socketPath,
		"focus-window", "--match", "title:"+windowTitle)

	if err := cmd.Run(); err != nil {
		return fmt.Errorf("failed to focus window %s: %w", windowTitle, err)
	}

	return nil
}

// ListWindows lists all windows in the Kitty instance
func (rc *RemoteControl) ListWindows() ([]WindowInfo, error) {
	cmd := exec.Command("kitty", "@", "--to", "unix:"+rc.socketPath, "ls")
	output, err := cmd.Output()
	if err != nil {
		return nil, fmt.Errorf("failed to list windows: %w", err)
	}

	// Parse JSON output
	// Kitty ls returns an array of OS windows, each containing tabs with windows
	var osWindows []struct {
		Tabs []struct {
			Windows []WindowInfo `json:"windows"`
		} `json:"tabs"`
	}

	if err := json.Unmarshal(output, &osWindows); err != nil {
		return nil, fmt.Errorf("failed to parse window list: %w", err)
	}

	// Flatten to list of windows
	var windows []WindowInfo
	for _, osWin := range osWindows {
		for _, tab := range osWin.Tabs {
			windows = append(windows, tab.Windows...)
		}
	}

	return windows, nil
}
