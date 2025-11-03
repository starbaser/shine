package main

import (
	"fmt"
	"os"

	"github.com/starbased-co/shine/pkg/panel"
)

func usage() {
	fmt.Println("Usage: shinectl <command> [arguments]")
	fmt.Println()
	fmt.Println("Commands:")
	fmt.Println("  toggle <panel>      Toggle panel visibility")
	fmt.Println("  show <panel>        Show panel")
	fmt.Println("  hide <panel>        Hide panel")
	fmt.Println("  new-prism <name>    Create new prism from template")
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
	fmt.Println("  shinectl new-prism weather")
}

func getSocketPath() string {
	// In single-instance mode, all components share the same socket
	return "/tmp/shine.sock"
}

func getWindowTitle(panelName string) string {
	// Generate window title for targeting
	return fmt.Sprintf("shine-%s", panelName)
}

func main() {
	if len(os.Args) < 2 {
		usage()
		os.Exit(1)
	}

	command := os.Args[1]

	// Handle new-prism command (doesn't need panel operations)
	if command == "new-prism" {
		if len(os.Args) < 3 {
			fmt.Println("Error: prism name required")
			fmt.Println("Usage: shinectl new-prism <name>")
			os.Exit(1)
		}
		prismName := os.Args[2]
		if err := newPrism(prismName); err != nil {
			fmt.Printf("Error creating prism: %v\n", err)
			os.Exit(1)
		}
		return
	}

	// All other commands require panel name
	if len(os.Args) < 3 {
		usage()
		os.Exit(1)
	}

	panelName := os.Args[2]

	// Get shared socket path
	socketPath := getSocketPath()
	windowTitle := getWindowTitle(panelName)

	// Create remote control client
	rc := panel.NewRemoteControl(socketPath)

	// Execute command
	var err error
	switch command {
	case "toggle":
		fmt.Printf("Toggling %s panel (window: %s)...\n", panelName, windowTitle)
		// List windows to check if this panel exists
		windows, listErr := rc.ListWindows()
		if listErr != nil {
			fmt.Printf("Error listing windows: %v\n", listErr)
			os.Exit(1)
		}

		// Check if window exists
		windowExists := false
		for _, win := range windows {
			if win.Title == windowTitle {
				windowExists = true
				break
			}
		}

		if windowExists {
			// Window exists, close it
			err = rc.CloseWindow(windowTitle)
		} else {
			// Window doesn't exist, would need to launch it
			fmt.Printf("Panel %s is not running. Use 'shine' to launch panels.\n", panelName)
			os.Exit(1)
		}

	case "show":
		fmt.Printf("Showing %s panel (window: %s)...\n", panelName, windowTitle)
		err = rc.FocusWindow(windowTitle)

	case "hide":
		fmt.Printf("Hiding %s panel (window: %s)...\n", panelName, windowTitle)
		err = rc.CloseWindow(windowTitle)

	default:
		fmt.Printf("Error: unknown command '%s'\n\n", command)
		usage()
		os.Exit(1)
	}

	if err != nil {
		fmt.Printf("Error: %v\n", err)
		fmt.Printf("\nTroubleshooting:\n")
		fmt.Printf("  - Is shine running? (check with: ps aux | grep 'kitten panel')\n")
		fmt.Printf("  - Is the socket accessible? (check: ls -la %s)\n", socketPath)
		fmt.Printf("  - List windows with: kitty @ --to unix:%s ls\n", socketPath)
		os.Exit(1)
	}

	fmt.Printf("✓ Command sent successfully\n")
}
