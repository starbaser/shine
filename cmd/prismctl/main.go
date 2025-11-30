// main.go is the prismctl entrypoint. Initializes state, instantiates
// configuration, and awaits prisms to start via RPC.

package main

import (
	"fmt"
	"log"
	"os"
	"path/filepath"

	"github.com/starbased-co/shine/pkg/paths"
)

func setupLogging() error {
	logDir := filepath.Join(os.Getenv("HOME"), ".local", "share", "shine", "logs")
	if err := os.MkdirAll(logDir, 0755); err != nil {
		return fmt.Errorf("failed to create log directory: %w", err)
	}

	logPath := filepath.Join(logDir, "prismctl.log")
	logFile, err := os.OpenFile(logPath, os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0644)
	if err != nil {
		return fmt.Errorf("failed to open log file: %w", err)
	}

	log.SetOutput(logFile)
	log.SetFlags(log.Ldate | log.Ltime | log.Lmicroseconds)
	log.SetPrefix("[prismctl] ")

	return nil
}


func main() {
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
	}

	if err := setupLogging(); err != nil {
		fmt.Fprintf(os.Stderr, "Failed to setup logging: %v\n", err)
		os.Exit(1)
	}

	if len(os.Args) < 2 {
		showHelp("")
		os.Exit(1)
	}

	instanceName := os.Args[1]
	log.Printf("prismctl starting (instance: %s)", instanceName)

	termState, err := newTerminalState()
	if err != nil {
		log.Fatalf("Failed to initialize terminal state: %v", err)
	}
	log.Printf("Terminal state saved")

	statePath := paths.PrismState(instanceName)
	stateMgr, err := newStateManager(statePath, instanceName)
	if err != nil {
		log.Fatalf("Failed to create state manager: %v", err)
	}
	defer stateMgr.Remove()
	log.Printf("State file created: %s", statePath)

	notifyMgr := newNotificationManager(instanceName)
	defer notifyMgr.Close()
	log.Printf("Notification manager started")

	sup := newSupervisor(termState, stateMgr, notifyMgr)

	sigHandler := newSignalHandler(sup)
	defer sigHandler.stop()

	rpcServer, err := startRPCServer(instanceName, sup, stateMgr)
	if err != nil {
		log.Fatalf("Failed to start RPC server: %v", err)
	}
	defer stopRPCServer(rpcServer)

	log.Printf("prismctl running (PID %d), awaiting configuration via RPC", os.Getpid())
	sigHandler.run()

	log.Printf("prismctl exiting")
}
