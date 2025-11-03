package main

import (
	"fmt"
	"log"
	"os"
	"os/exec"
	"os/signal"
	"path/filepath"
	"syscall"
	"time"

	"github.com/starbased-co/shine/pkg/config"
	"github.com/starbased-co/shine/pkg/panel"
)

// findComponentBinary finds the full path to a component binary
func findComponentBinary(name string) (string, error) {
	// First, check if it's in PATH
	path, err := exec.LookPath(name)
	if err == nil {
		return path, nil
	}

	// If not in PATH, try relative to the shine binary
	exePath, err := os.Executable()
	if err != nil {
		return "", fmt.Errorf("failed to get executable path: %w", err)
	}

	// Try in the same directory as shine
	binDir := filepath.Dir(exePath)
	componentPath := filepath.Join(binDir, name)
	if _, err := os.Stat(componentPath); err == nil {
		return componentPath, nil
	}

	return "", fmt.Errorf("component binary %s not found in PATH or %s", name, binDir)
}

func main() {
	// Load configuration
	configPath := config.DefaultConfigPath()
	cfg := config.LoadOrDefault(configPath)

	fmt.Printf("✨ Shine - Hyprland Layer Shell TUI Toolkit\n")
	fmt.Printf("Configuration: %s\n\n", configPath)

	// Create panel manager
	mgr := panel.NewManager()

	// Launch enabled components
	if cfg.Chat != nil && cfg.Chat.Enabled {
		fmt.Println("Launching chat panel...")
		panelCfg := cfg.Chat.ToPanelConfig()

		// Find the shine-chat binary
		chatBinary, err := findComponentBinary("shine-chat")
		if err != nil {
			log.Fatalf("Failed to find shine-chat binary: %v", err)
		}

		instance, err := mgr.Launch("chat", panelCfg, chatBinary)
		if err != nil {
			log.Fatalf("Failed to launch chat panel: %v", err)
		}

		if instance.WindowID != "" {
			fmt.Printf("  ✓ Chat panel launched (Window ID: %s)\n", instance.WindowID)
		} else if instance.Command != nil && instance.Command.Process != nil {
			fmt.Printf("  ✓ Chat panel launched (PID: %d)\n", instance.Command.Process.Pid)
		} else {
			fmt.Printf("  ✓ Chat panel launched\n")
		}
		if panelCfg.ListenSocket != "" {
			fmt.Printf("  ✓ Remote control: %s\n", panelCfg.ListenSocket)
		}

		// Wait for the first instance to fully start before launching the second
		// This ensures single-instance mode can detect the existing instance
		time.Sleep(500 * time.Millisecond)
	}

	if cfg.Bar != nil && cfg.Bar.Enabled {
		fmt.Println("Launching status bar...")
		panelCfg := cfg.Bar.ToPanelConfig()

		// Find the shine-bar binary
		barBinary, err := findComponentBinary("shine-bar")
		if err != nil {
			log.Fatalf("Failed to find shine-bar binary: %v", err)
		}

		instance, err := mgr.Launch("bar", panelCfg, barBinary)
		if err != nil {
			log.Fatalf("Failed to launch status bar: %v", err)
		}

		if instance.WindowID != "" {
			fmt.Printf("  ✓ Status bar launched (Window ID: %s)\n", instance.WindowID)
		} else if instance.Command != nil && instance.Command.Process != nil {
			fmt.Printf("  ✓ Status bar launched (PID: %d)\n", instance.Command.Process.Pid)
		} else {
			fmt.Printf("  ✓ Status bar launched\n")
		}
		if panelCfg.ListenSocket != "" {
			fmt.Printf("  ✓ Remote control: %s\n", panelCfg.ListenSocket)
		}

		time.Sleep(500 * time.Millisecond)
	}

	if cfg.Clock != nil && cfg.Clock.Enabled {
		fmt.Println("Launching clock...")
		panelCfg := cfg.Clock.ToPanelConfig()

		// Find the shine-clock binary
		clockBinary, err := findComponentBinary("shine-clock")
		if err != nil {
			log.Fatalf("Failed to find shine-clock binary: %v", err)
		}

		instance, err := mgr.Launch("clock", panelCfg, clockBinary)
		if err != nil {
			log.Fatalf("Failed to launch clock: %v", err)
		}

		if instance.WindowID != "" {
			fmt.Printf("  ✓ Clock launched (Window ID: %s)\n", instance.WindowID)
		} else if instance.Command != nil && instance.Command.Process != nil {
			fmt.Printf("  ✓ Clock launched (PID: %d)\n", instance.Command.Process.Pid)
		} else {
			fmt.Printf("  ✓ Clock launched\n")
		}
		if panelCfg.ListenSocket != "" {
			fmt.Printf("  ✓ Remote control: %s\n", panelCfg.ListenSocket)
		}

		time.Sleep(500 * time.Millisecond)
	}

	if cfg.SysInfo != nil && cfg.SysInfo.Enabled {
		fmt.Println("Launching system info...")
		panelCfg := cfg.SysInfo.ToPanelConfig()

		// Find the shine-sysinfo binary
		sysinfoBinary, err := findComponentBinary("shine-sysinfo")
		if err != nil {
			log.Fatalf("Failed to find shine-sysinfo binary: %v", err)
		}

		instance, err := mgr.Launch("sysinfo", panelCfg, sysinfoBinary)
		if err != nil {
			log.Fatalf("Failed to launch system info: %v", err)
		}

		if instance.WindowID != "" {
			fmt.Printf("  ✓ System info launched (Window ID: %s)\n", instance.WindowID)
		} else if instance.Command != nil && instance.Command.Process != nil {
			fmt.Printf("  ✓ System info launched (PID: %d)\n", instance.Command.Process.Pid)
		} else {
			fmt.Printf("  ✓ System info launched\n")
		}
		if panelCfg.ListenSocket != "" {
			fmt.Printf("  ✓ Remote control: %s\n", panelCfg.ListenSocket)
		}
	}

	// List running panels
	panels := mgr.List()
	if len(panels) == 0 {
		fmt.Println("\nNo panels enabled. Edit your config to enable components.")
		fmt.Printf("Config location: %s\n", configPath)
		os.Exit(0)
	}

	fmt.Printf("\nRunning %d panel(s): %v\n", len(panels), panels)
	fmt.Println("Press Ctrl+C to stop all panels")

	// Setup signal handling
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, os.Interrupt, syscall.SIGTERM)

	// Wait for signal or panels to exit
	go mgr.Wait()

	<-sigChan
	fmt.Println("\n\nShutting down...")

	// Stop all panels
	for _, name := range mgr.List() {
		fmt.Printf("Stopping %s...\n", name)
		if err := mgr.Stop(name); err != nil {
			log.Printf("Error stopping %s: %v", name, err)
		}
	}

	fmt.Println("Goodbye!")
}
