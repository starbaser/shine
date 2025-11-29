package main

import (
	"context"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"time"

	"github.com/starbased-co/shine/pkg/paths"
	"github.com/starbased-co/shine/pkg/rpc"
	"github.com/starbased-co/shine/pkg/state"
)

func connectShinectl() (*rpc.ShinectlClient, error) {
	sockPath := paths.ShinectlSocket()
	return rpc.NewShinectlClient(sockPath, rpc.WithTimeout(3*time.Second))
}

func isShinectlRunning() bool {
	_, err := os.Stat(paths.ShinectlSocket())
	return err == nil
}

// discovers running prismctl instances by scanning for runtime files (sockets/mmap)
func discoverPrismInstances() ([]string, error) {
	socketsDir := paths.RuntimeDir()

	// Try sockets first (they're authoritative for running instances)
	pattern := filepath.Join(socketsDir, "prism-*.sock")
	matches, err := filepath.Glob(pattern)
	if err != nil {
		return nil, fmt.Errorf("failed to search for sockets: %w", err)
	}

	instances := make([]string, 0, len(matches))
	for _, socket := range matches {
		instances = append(instances, extractInstanceName(socket))
	}

	return instances, nil
}

// findPrismctlSockets finds all prismctl sockets (legacy helper)
func findPrismctlSockets() ([]string, error) {
	instances, err := discoverPrismInstances()
	if err != nil {
		return nil, err
	}

	sockets := make([]string, len(instances))
	for i, instance := range instances {
		sockets[i] = paths.PrismSocket(instance)
	}

	return sockets, nil
}


// cmdStart starts or resumes the shinectl service
func cmdStart() error {
	// Check if shinectl is already running
	if isShinectlRunning() {
		Success("shinectl is already running")
		return nil
	}

	Info("Starting shinectl service...")

	// Find shinectl binary
	shinectlBin, err := exec.LookPath("shinectl")
	if err != nil {
		return fmt.Errorf("shinectl not found in PATH: %w", err)
	}

	// Start shinectl in background
	cmd := exec.Command(shinectlBin)
	cmd.Stdout = nil
	cmd.Stderr = nil
	cmd.Stdin = nil

	if err := cmd.Start(); err != nil {
		return fmt.Errorf("failed to start shinectl: %w", err)
	}

	// Wait for socket to appear
	for i := 0; i < 50; i++ {
		if isShinectlRunning() {
			Success(fmt.Sprintf("shinectl started (PID: %d)", cmd.Process.Pid))
			return nil
		}
		time.Sleep(100 * time.Millisecond)
	}

	return fmt.Errorf("shinectl started but socket not created within timeout")
}

// cmdStop gracefully stops all panels and shinectl
func cmdStop() error {
	Info("Stopping shine service...")

	ctx := context.Background()

	// Try shinectl first
	if isShinectlRunning() {
		client, err := connectShinectl()
		if err != nil {
			Warning(fmt.Sprintf("Failed to connect to shinectl: %v, falling back to direct shutdown", err))
		} else {
			defer client.Close()

			// Get panel list
			result, err := client.Status(ctx)
			if err != nil {
				Warning(fmt.Sprintf("Failed to query shinectl status: %v", err))
			} else {
				// Shutdown each panel via shinectl
				for _, panel := range result.Panels {
					Muted(fmt.Sprintf("Stopping %s...", panel.Instance))
					_, err := client.KillPanel(ctx, panel.Instance)
					if err != nil {
						Warning(fmt.Sprintf("Failed to stop %s: %v", panel.Instance, err))
					}
				}
				Success(fmt.Sprintf("Stopped %d panel(s)", len(result.Panels)))
				return nil
			}
		}
	}

	// Fallback: discover and stop prismctl instances directly
	instances, err := discoverPrismInstances()
	if err != nil {
		return err
	}

	if len(instances) == 0 {
		Warning("No panels running")
		return nil
	}

	// Send shutdown command to each panel
	var stopped int
	for _, instance := range instances {
		Muted(fmt.Sprintf("Stopping %s...", instance))

		client, err := rpc.NewPrismClient(paths.PrismSocket(instance))
		if err != nil {
			Warning(fmt.Sprintf("Failed to connect to %s: %v", instance, err))
			continue
		}

		_, err = client.Shutdown(ctx, true)
		client.Close()

		if err != nil {
			Warning(fmt.Sprintf("Failed to stop %s: %v", instance, err))
		} else {
			stopped++
		}
	}

	Success(fmt.Sprintf("Stopped %d panel(s)", stopped))
	return nil
}

// cmdReload reloads the configuration and updates panels
func cmdReload() error {
	Info("Reloading configuration...")

	if !isShinectlRunning() {
		return fmt.Errorf("shinectl is not running")
	}

	ctx := context.Background()
	client, err := connectShinectl()
	if err != nil {
		return fmt.Errorf("failed to connect to shinectl: %w", err)
	}
	defer client.Close()

	result, err := client.Reload(ctx)
	if err != nil {
		return fmt.Errorf("reload request failed: %w", err)
	}

	if !result.Reloaded {
		if len(result.Errors) > 0 {
			Error("Configuration reload failed:")
			for _, errMsg := range result.Errors {
				Muted(fmt.Sprintf("  - %s", errMsg))
			}
			return fmt.Errorf("reload completed with errors")
		}
		return fmt.Errorf("reload failed with no error details")
	}

	Success("Configuration reloaded successfully")
	return nil
}

// displayStateFromMmap formats and displays state read from mmap
func displayStateFromMmap(instance string, s *state.PrismRuntimeState) {
	fmt.Println()
	fmt.Printf("%s %s\n", styleBold.Render("Panel:"), instance)
	fmt.Printf("%s %s\n", styleMuted.Render("Source:"), "mmap")

	fgName := s.GetFgPrism()
	activePrisms := s.ActivePrisms()
	bgCount := len(activePrisms) - 1
	if fgName == "" {
		bgCount = len(activePrisms)
	}

	// Display status box
	fmt.Println(StatusBox(fgName, bgCount, len(activePrisms)))

	// Show prisms table
	if len(activePrisms) > 0 {
		table := NewTable("Prism", "PID", "State", "Uptime")
		for _, prism := range activePrisms {
			name := prism.GetName()
			stateStr := "background"
			if name == fgName {
				stateStr = styleSuccess.Render("foreground")
			} else {
				stateStr = styleMuted.Render("background")
			}
			uptime := prism.Uptime()
			uptimeStr := fmt.Sprintf("%v", uptime.Truncate(time.Second))
			table.AddRow(name, fmt.Sprintf("%d", prism.PID), stateStr, uptimeStr)
		}
		fmt.Println()
		table.Print()
	}
}

// displayStateFromRPC formats and displays state read from RPC
func displayStateFromRPC(instance string, prisms []rpc.PrismInfo) {
	fmt.Println()
	fmt.Printf("%s %s\n", styleBold.Render("Panel:"), instance)
	fmt.Printf("%s %s\n", styleMuted.Render("Source:"), "rpc")

	fgName := ""
	bgCount := 0
	for _, p := range prisms {
		if p.State == "fg" {
			fgName = p.Name
		} else {
			bgCount++
		}
	}

	// Display status box
	fmt.Println(StatusBox(fgName, bgCount, len(prisms)))

	// Show prisms table
	if len(prisms) > 0 {
		table := NewTable("Prism", "PID", "State", "Uptime")
		for _, prism := range prisms {
			stateStr := prism.State
			if prism.State == "fg" {
				stateStr = styleSuccess.Render("foreground")
			} else {
				stateStr = styleMuted.Render("background")
			}
			uptime := time.Duration(prism.UptimeMs) * time.Millisecond
			uptimeStr := fmt.Sprintf("%v", uptime.Truncate(time.Second))
			table.AddRow(prism.Name, fmt.Sprintf("%d", prism.PID), stateStr, uptimeStr)
		}
		fmt.Println()
		table.Print()
	}
}

// cmdStatus shows the status of all panels
func cmdStatus() error {
	ctx := context.Background()

	// Try shinectl first for aggregated status
	if isShinectlRunning() {
		client, err := connectShinectl()
		if err == nil {
			defer client.Close()

			result, err := client.Status(ctx)
			if err == nil {
				// Display shinectl-level status
				uptime := time.Duration(result.Uptime) * time.Millisecond
				uptimeStr := uptime.Truncate(time.Second).String()

				Header(fmt.Sprintf("Shine Status (v%s, uptime: %s)", result.Version, uptimeStr))

				if len(result.Panels) == 0 {
					Warning("No panels running")
					Info("Start panels with: shine start")
					return nil
				}

				// Query each panel for detailed status
				for _, panel := range result.Panels {
					displayPanelStatus(ctx, panel.Instance)
				}
				return nil
			}
			// If shinectl query fails, fall through to discovery
			Warning(fmt.Sprintf("Failed to query shinectl: %v, falling back to discovery", err))
		}
	}

	// Fallback: discover all running prism instances directly
	instances, err := discoverPrismInstances()
	if err != nil {
		return err
	}

	if len(instances) == 0 {
		Warning("No panels running")
		Info("Start panels with: shine start")
		return nil
	}

	Header(fmt.Sprintf("Shine Status (%d panel(s))", len(instances)))

	// Query each panel
	for _, instance := range instances {
		displayPanelStatus(ctx, instance)
	}

	return nil
}

// displayPanelStatus queries a single panel and displays its status
func displayPanelStatus(ctx context.Context, instance string) {
	// Try mmap first (instant, no connection needed)
	reader, err := state.OpenPrismStateReader(paths.PrismState(instance))
	if err == nil {
		s, readErr := reader.Read()
		reader.Close()
		if readErr == nil {
			displayStateFromMmap(instance, s)
			return
		}
	}

	// Fallback to RPC
	client, err := rpc.NewPrismClient(paths.PrismSocket(instance))
	if err != nil {
		fmt.Println()
		fmt.Printf("%s %s\n", styleBold.Render("Panel:"), instance)
		Error(fmt.Sprintf("Failed to connect: %v", err))
		return
	}

	result, err := client.List(ctx)
	client.Close()

	if err != nil {
		fmt.Println()
		fmt.Printf("%s %s\n", styleBold.Render("Panel:"), instance)
		Error(fmt.Sprintf("Failed to query: %v", err))
		return
	}

	displayStateFromRPC(instance, result.Prisms)
}

// cmdLogs shows logs for a specific panel or all panels
func cmdLogs(panelID string) error {
	home, err := os.UserHomeDir()
	if err != nil {
		return fmt.Errorf("failed to get home directory: %w", err)
	}

	logDir := filepath.Join(home, ".local", "share", "shine", "logs")

	if panelID == "" {
		// Show all logs
		Info(fmt.Sprintf("Log directory: %s", logDir))

		files, err := os.ReadDir(logDir)
		if err != nil {
			return fmt.Errorf("failed to read log directory: %w", err)
		}

		if len(files) == 0 {
			Warning("No log files found")
			return nil
		}

		table := NewTable("Log File", "Size")
		for _, file := range files {
			if file.IsDir() {
				continue
			}
			info, _ := file.Info()
			size := "?"
			if info != nil {
				size = fmt.Sprintf("%d bytes", info.Size())
			}
			table.AddRow(file.Name(), size)
		}

		table.Print()
		fmt.Println()
		Info("View a log with: shine logs <filename>")
		return nil
	}

	// Show specific log file
	logPath := filepath.Join(logDir, panelID)
	if !strings.HasSuffix(logPath, ".log") {
		logPath += ".log"
	}

	// Check if file exists
	if _, err := os.Stat(logPath); os.IsNotExist(err) {
		return fmt.Errorf("log file not found: %s", logPath)
	}

	// Tail the log file (last 50 lines)
	cmd := exec.Command("tail", "-n", "50", logPath)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr

	if err := cmd.Run(); err != nil {
		return fmt.Errorf("failed to read log: %w", err)
	}

	return nil
}

// extractInstanceName extracts the instance name from a socket path
// e.g., "/run/user/1000/shine/prism-clock.sock" -> "clock"
func extractInstanceName(socketPath string) string {
	base := filepath.Base(socketPath)
	// Remove "prism-" prefix and ".sock" suffix
	name := strings.TrimPrefix(base, "prism-")
	name = strings.TrimSuffix(name, ".sock")
	return name
}
