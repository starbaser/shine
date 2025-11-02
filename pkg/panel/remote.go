package panel

import (
	"encoding/json"
	"fmt"
	"net"
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
