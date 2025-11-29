package main

import (
	"fmt"
	"log"
	"os"
	"path/filepath"

	"github.com/starbased-co/shine/pkg/paths"
)

// setupLogging configures logging to file instead of stderr
func setupLogging() error {
	// Create log directory
	logDir := filepath.Join(os.Getenv("HOME"), ".local", "share", "shine", "logs")
	if err := os.MkdirAll(logDir, 0755); err != nil {
		return fmt.Errorf("failed to create log directory: %w", err)
	}

	// Open log file (append mode)
	logPath := filepath.Join(logDir, "prismctl.log")
	logFile, err := os.OpenFile(logPath, os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0644)
	if err != nil {
		return fmt.Errorf("failed to open log file: %w", err)
	}

	// Set log output to file
	log.SetOutput(logFile)
	log.SetFlags(log.Ldate | log.Ltime | log.Lmicroseconds)
	log.SetPrefix("[prismctl] ")

	return nil
}


func main() {
	// Handle help requests before anything else
	if len(os.Args) >= 2 {
		arg := os.Args[1]
		if arg == "-h" || arg == "--help" {
			showHelp("")
			os.Exit(0)
		}
		if arg == "help" {
			topic := ""
			if len(os.Args) >= 3 {
				topic = os.Args[2]
			}
			showHelp(topic)
			os.Exit(0)
		}
		if arg == "--json" && len(os.Args) >= 3 && os.Args[2] == "help" {
			topic := ""
			if len(os.Args) >= 4 {
				topic = os.Args[3]
			}
			if err := helpJSON(topic); err != nil {
				fmt.Fprintf(os.Stderr, "Error: %v\n", err)
				os.Exit(1)
			}
			os.Exit(0)
		}
	}

	// Setup logging to file
	if err := setupLogging(); err != nil {
		fmt.Fprintf(os.Stderr, "Failed to setup logging: %v\n", err)
		os.Exit(1)
	}

	// Parse arguments - only expect instance name
	if len(os.Args) < 2 {
		showHelp("")
		os.Exit(1)
	}

	instanceName := os.Args[1]
	log.Printf("prismctl starting (instance: %s)", instanceName)

	// Initialize terminal state management
	termState, err := newTerminalState()
	if err != nil {
		log.Fatalf("Failed to initialize terminal state: %v", err)
	}
	log.Printf("Terminal state saved")

	// Create state manager
	statePath := paths.PrismState(instanceName)
	stateMgr, err := newStateManager(statePath, instanceName)
	if err != nil {
		log.Fatalf("Failed to create state manager: %v", err)
	}
	defer stateMgr.Remove() // Clean up state file on exit
	log.Printf("State file created: %s", statePath)

	// Create notification manager
	notifyMgr := newNotificationManager(instanceName)
	defer notifyMgr.Close()
	log.Printf("Notification manager started")

	// Create supervisor with state manager and notification manager
	sup := newSupervisor(termState, stateMgr, notifyMgr)

	// Setup signal handling
	sigHandler := newSignalHandler(sup)
	defer sigHandler.stop()

	// Start RPC server
	rpcServer, err := startRPCServer(instanceName, sup, stateMgr)
	if err != nil {
		log.Fatalf("Failed to start RPC server: %v", err)
	}
	defer stopRPCServer(rpcServer)

	// Wait for configuration via RPC (no initial prism launch)
	log.Printf("prismctl running (PID %d), awaiting configuration via RPC", os.Getpid())
	sigHandler.run()

	log.Printf("prismctl exiting")
}
