package main

import (
	"fmt"
	"os"
	"os/exec"
	"strings"

	"github.com/starbased-co/shine/pkg/panel"
)

func usage() {
	fmt.Println("Usage: shinectl <command> <panel>")
	fmt.Println()
	fmt.Println("Commands:")
	fmt.Println("  toggle <panel>  Toggle panel visibility")
	fmt.Println("  show <panel>    Show panel")
	fmt.Println("  hide <panel>    Hide panel")
	fmt.Println()
	fmt.Println("Panels:")
	fmt.Println("  chat           Chat component")
	fmt.Println("  bar            Status bar component")
	fmt.Println()
	fmt.Println("Examples:")
	fmt.Println("  shinectl toggle chat")
	fmt.Println("  shinectl show chat")
	fmt.Println("  shinectl hide chat")
	fmt.Println("  shinectl toggle bar")
}

func getSocketPath(panelName string) (string, error) {
	// Kitty appends the PID to socket paths when using panels
	// Find the actual socket by looking for the pattern: /tmp/shine-{name}.sock-PID

	// First, try to find the process
	cmd := exec.Command("pgrep", "-f", fmt.Sprintf("shine-%s", panelName))
	output, err := cmd.Output()
	if err != nil {
		// Process not found, return the base path and let the error happen
		return fmt.Sprintf("/tmp/shine-%s.sock", panelName), fmt.Errorf("panel %s not running (process not found)", panelName)
	}

	// Get the PID (first line of output)
	pidStr := strings.TrimSpace(strings.Split(string(output), "\n")[0])

	// Construct the actual socket path with PID
	return fmt.Sprintf("/tmp/shine-%s.sock-%s", panelName, pidStr), nil
}

func main() {
	if len(os.Args) < 3 {
		usage()
		os.Exit(1)
	}

	command := os.Args[1]
	panelName := os.Args[2]

	// Get socket path for panel
	socketPath, pathErr := getSocketPath(panelName)
	if pathErr != nil {
		fmt.Printf("Warning: %v\n", pathErr)
		fmt.Printf("Attempting to use socket path: %s\n", socketPath)
	}

	// Create remote control client
	rc := panel.NewRemoteControl(socketPath)

	// Execute command
	var err error
	switch command {
	case "toggle":
		fmt.Printf("Toggling %s panel...\n", panelName)
		err = rc.ToggleVisibility()
	case "show":
		fmt.Printf("Showing %s panel...\n", panelName)
		err = rc.Show()
	case "hide":
		fmt.Printf("Hiding %s panel...\n", panelName)
		err = rc.Hide()
	default:
		fmt.Printf("Error: unknown command '%s'\n\n", command)
		usage()
		os.Exit(1)
	}

	if err != nil {
		fmt.Printf("Error: %v\n", err)
		fmt.Printf("\nTroubleshooting:\n")
		fmt.Printf("  - Is the %s panel running? (check with: ps aux | grep shine-%s)\n", panelName, panelName)
		fmt.Printf("  - Is the socket accessible? (check: ls -la %s)\n", socketPath)
		os.Exit(1)
	}

	fmt.Printf("✓ Command sent successfully\n")
}
