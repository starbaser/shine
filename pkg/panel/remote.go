package panel

import (
	"encoding/json"
	"fmt"
	"net"
	"os/exec"
)

type RemoteControl struct {
	socketPath string
}

func NewRemoteControl(socketPath string) *RemoteControl {
	return &RemoteControl{
		socketPath: socketPath,
	}
}

func (rc *RemoteControl) ToggleVisibility() error {
	conn, err := net.Dial("unix", rc.socketPath)
	if err != nil {
		return fmt.Errorf("failed to connect to %s: %w", rc.socketPath, err)
	}
	defer conn.Close()

	payload := map[string]interface{}{
		"cmd":    "resize-os-window",
		"action": "toggle-visibility",
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

type WindowInfo struct {
	ID    int    `json:"id"`
	Title string `json:"title"`
}

func (rc *RemoteControl) CloseWindow(windowTitle string) error {
	cmd := exec.Command("kitty", "@", "--to", "unix:"+rc.socketPath,
		"close-window", "--match", "title:"+windowTitle)

	if err := cmd.Run(); err != nil {
		return fmt.Errorf("failed to close window %s: %w", windowTitle, err)
	}

	return nil
}

func (rc *RemoteControl) FocusWindow(windowTitle string) error {
	cmd := exec.Command("kitty", "@", "--to", "unix:"+rc.socketPath,
		"focus-window", "--match", "title:"+windowTitle)

	if err := cmd.Run(); err != nil {
		return fmt.Errorf("failed to focus window %s: %w", windowTitle, err)
	}

	return nil
}

func (rc *RemoteControl) ListWindows() ([]WindowInfo, error) {
	cmd := exec.Command("kitty", "@", "--to", "unix:"+rc.socketPath, "ls")
	output, err := cmd.Output()
	if err != nil {
		return nil, fmt.Errorf("failed to list windows: %w", err)
	}

	var osWindows []struct {
		Tabs []struct {
			Windows []WindowInfo `json:"windows"`
		} `json:"tabs"`
	}

	if err := json.Unmarshal(output, &osWindows); err != nil {
		return nil, fmt.Errorf("failed to parse window list: %w", err)
	}

	var windows []WindowInfo
	for _, osWin := range osWindows {
		for _, tab := range osWin.Tabs {
			windows = append(windows, tab.Windows...)
		}
	}

	return windows, nil
}
