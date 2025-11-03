package panel

import (
	"bufio"
	"fmt"
	"log"
	"os"
	"os/exec"
	"strconv"
	"strings"
	"sync"
)

// Instance represents a running panel instance
type Instance struct {
	Name        string
	Command     *exec.Cmd
	Config      *Config
	Remote      *RemoteControl
	WindowID    string // Window ID from kitty @ launch
	WindowTitle string // For window matching in shared instance mode
}

// kittyInstance tracks the shared Kitty instance
type kittyInstance struct {
	socketPath string
	pid        int
}

// Manager manages panel instances
type Manager struct {
	instances      map[string]*Instance
	kittyInstance  *kittyInstance // Detected Kitty instance with remote control
	sharedInstance *exec.Cmd      // The shared Kitty process (for kitten panel mode)
	mu             sync.RWMutex
}

// NewManager creates a new panel manager
func NewManager() *Manager {
	return &Manager{
		instances: make(map[string]*Instance),
	}
}

// testSocket tests if a socket accepts remote control connections
func (m *Manager) testSocket(socketPath string) bool {
	var testCmd *exec.Cmd
	if socketPath != "" {
		testCmd = exec.Command("kitty", "@", "--to", socketPath, "ls")
	} else {
		// Use default socket
		testCmd = exec.Command("kitty", "@", "ls")
	}
	return testCmd.Run() == nil
}

// detectKittySocket finds a Kitty instance with remote control enabled
func (m *Manager) detectKittySocket() (string, error) {
	// Check if we already have an instance
	if m.kittyInstance != nil {
		// Verify it's still running
		socketPath := m.kittyInstance.socketPath
		if m.testSocket(socketPath) {
			return socketPath, nil
		}
		// Stale instance, clear it
		m.kittyInstance = nil
	}

	// Method 1: Check if we're already inside Kitty (has window ID)
	if kittyWindowID := os.Getenv("KITTY_WINDOW_ID"); kittyWindowID != "" {
		log.Printf("[kitty detection] Running inside Kitty window ID: %s", kittyWindowID)
		// When inside Kitty, try default socket first (no --to needed)
		testCmd := exec.Command("kitty", "@", "ls")
		if err := testCmd.Run(); err == nil {
			log.Printf("[kitty detection] Using default socket (no --to required)")
			// Use empty string to signal "use default"
			m.kittyInstance = &kittyInstance{
				socketPath: "", // Empty means use default
				pid:        0,
			}
			return "", nil
		}
	}

	// Method 2: Check KITTY_LISTEN_ON environment variable
	if listenOn := os.Getenv("KITTY_LISTEN_ON"); listenOn != "" {
		log.Printf("[kitty detection] Checking KITTY_LISTEN_ON: %s", listenOn)
		if m.testSocket(listenOn) {
			log.Printf("[kitty detection] Using KITTY_LISTEN_ON socket")
			m.kittyInstance = &kittyInstance{
				socketPath: listenOn,
				pid:        0,
			}
			return listenOn, nil
		}
	}

	// Method 3: Check for Kitty processes with PID-based sockets
	// Use "pgrep kitty" without -x to match Nix wrappers
	log.Printf("[kitty detection] Searching for Kitty processes...")
	cmd := exec.Command("pgrep", "kitty")
	output, err := cmd.Output()
	if err != nil {
		kittyWindowID := os.Getenv("KITTY_WINDOW_ID")
		return "", fmt.Errorf(`no kitty processes found

Troubleshooting:
1. Ensure Kitty is running
2. If you want single-instance mode, enable remote control:
   Add to ~/.config/kitty/kitty.conf:
     allow_remote_control yes
     listen_on unix:/tmp/@mykitty
3. Restart Kitty
4. Or use multi-panel mode: set single_instance=false in config

Environment check:
  KITTY_WINDOW_ID: %s
  KITTY_LISTEN_ON: %s
  Running in Kitty: %v`, kittyWindowID, os.Getenv("KITTY_LISTEN_ON"), kittyWindowID != "")
	}

	pids := strings.Split(strings.TrimSpace(string(output)), "\n")
	log.Printf("[kitty detection] Found %d Kitty processes", len(pids))

	// Try each PID's socket (common pattern: /tmp/@mykitty-<PID>)
	for _, pidStr := range pids {
		pidStr = strings.TrimSpace(pidStr)
		if pidStr == "" {
			continue
		}
		socketPath := fmt.Sprintf("unix:/tmp/@mykitty-%s", pidStr)

		log.Printf("[kitty detection] Testing socket: %s", socketPath)
		if m.testSocket(socketPath) {
			log.Printf("[kitty detection] Successfully connected to socket")
			pid, _ := strconv.Atoi(pidStr)
			m.kittyInstance = &kittyInstance{
				socketPath: socketPath,
				pid:        pid,
			}
			return socketPath, nil
		}
	}

	kittyWindowID := os.Getenv("KITTY_WINDOW_ID")
	return "", fmt.Errorf(`no kitty instance with remote control enabled found

Checked %d Kitty processes but none accepted remote control connections.

To enable remote control, add to ~/.config/kitty/kitty.conf:
  allow_remote_control yes
  listen_on unix:/tmp/@mykitty

Then restart Kitty.

Alternatively, set single_instance=false in your Shine config to use
multi-panel mode (doesn't require remote control).

Environment:
  KITTY_WINDOW_ID: %s
  KITTY_LISTEN_ON: %s`, len(pids), kittyWindowID, os.Getenv("KITTY_LISTEN_ON"))
}

// LaunchViaRemoteControl launches a panel using Kitty's remote control API
func (m *Manager) LaunchViaRemoteControl(name string, config *Config, component string) (*Instance, error) {
	m.mu.Lock()
	defer m.mu.Unlock()

	// Check if already running
	if _, exists := m.instances[name]; exists {
		return nil, fmt.Errorf("panel %s already running", name)
	}

	// Find Kitty socket
	socketPath, err := m.detectKittySocket()
	if err != nil {
		return nil, fmt.Errorf("failed to find kitty instance: %w (ensure Kitty is running with allow_remote_control=yes)", err)
	}

	// Set window title for tracking
	config.WindowTitle = fmt.Sprintf("shine-%s", name)

	// Build remote control args
	args := config.ToRemoteControlArgs(component)

	// Build full args
	var fullArgs []string
	if socketPath != "" {
		// Use specific socket
		fullArgs = append([]string{"--to", socketPath}, args...)
	} else {
		// Use default socket (when running inside Kitty)
		fullArgs = args
	}

	// Launch via remote control
	cmd := exec.Command("kitty", fullArgs...)
	output, err := cmd.Output()
	if err != nil {
		return nil, fmt.Errorf("failed to launch via remote control: %w", err)
	}

	// Parse window ID from output (kitty @ launch returns window ID)
	windowID := strings.TrimSpace(string(output))

	// Create remote control client
	remote := NewRemoteControl(strings.TrimPrefix(socketPath, "unix:"))

	// Create instance
	instance := &Instance{
		Name:        name,
		Command:     nil, // No command to track
		Config:      config,
		Remote:      remote,
		WindowID:    windowID,
		WindowTitle: config.WindowTitle,
	}

	// Store instance
	m.instances[name] = instance

	return instance, nil
}

// LaunchViaKittenPanel starts a panel via kitten panel (legacy mode)
func (m *Manager) LaunchViaKittenPanel(name string, config *Config, component string) (*Instance, error) {
	m.mu.Lock()
	defer m.mu.Unlock()

	// Check if already running
	if _, exists := m.instances[name]; exists {
		return nil, fmt.Errorf("panel %s already running", name)
	}

	// Set window title for this component
	windowTitle := fmt.Sprintf("shine-%s", name)
	config.WindowTitle = windowTitle

	// Build kitten args
	args := config.ToKittenArgs(component)

	// Create command
	cmd := exec.Command("kitten", args...)

	// In single-instance mode:
	// - First launch: kitten panel starts and stays running
	// - Subsequent launches: kitten panel connects, creates window, exits
	isFirstInstance := m.sharedInstance == nil

	var stderr *bufio.Scanner
	if isFirstInstance {
		// Only capture stderr for the first instance (the actual Kitty process)
		stderrPipe, err := cmd.StderrPipe()
		if err != nil {
			return nil, fmt.Errorf("failed to create stderr pipe for panel %s: %w", name, err)
		}
		stderr = bufio.NewScanner(stderrPipe)
	}

	// Start process
	if err := cmd.Start(); err != nil {
		return nil, fmt.Errorf("failed to start panel %s: %w", name, err)
	}

	// Track the shared Kitty instance (only for first launch)
	if isFirstInstance {
		m.sharedInstance = cmd

		// Read stderr in background (for debugging)
		go func() {
			for stderr.Scan() {
				fmt.Printf("[kitty stderr] %s\n", stderr.Text())
			}
		}()
	} else {
		// For subsequent launches, wait for the command to exit
		// (it exits immediately after creating the window)
		go cmd.Wait()
	}

	// Create remote control client using shared socket
	var remote *RemoteControl
	if config.ListenSocket != "" {
		remote = NewRemoteControl(config.ListenSocket)
	}

	// Create instance
	// Store the shared instance command for all panels
	instanceCmd := m.sharedInstance
	instance := &Instance{
		Name:        name,
		Command:     instanceCmd,
		Config:      config,
		Remote:      remote,
		WindowTitle: windowTitle,
	}

	// Store instance
	m.instances[name] = instance

	return instance, nil
}

// Launch starts a panel (via remote control or kitten panel)
func (m *Manager) Launch(name string, config *Config, component string) (*Instance, error) {
	// Route to appropriate launch method based on SingleInstance flag
	if config.SingleInstance {
		instance, err := m.LaunchViaRemoteControl(name, config, component)
		if err != nil {
			// If remote control fails, fall back to kitten panel mode
			log.Printf("Warning: Remote control launch failed (%v), falling back to kitten panel mode", err)
			log.Printf("Note: Set single_instance=false in config to suppress this warning")

			// Create a copy of config with SingleInstance disabled to avoid recursion
			fallbackConfig := *config
			fallbackConfig.SingleInstance = false

			return m.LaunchViaKittenPanel(name, &fallbackConfig, component)
		}
		return instance, nil
	}

	// Otherwise, use kitten panel approach
	return m.LaunchViaKittenPanel(name, config, component)
}

// Get retrieves an instance by name
func (m *Manager) Get(name string) (*Instance, bool) {
	m.mu.RLock()
	defer m.mu.RUnlock()

	instance, exists := m.instances[name]
	return instance, exists
}

// Stop closes a panel window (in shared instance mode) or kills process
func (m *Manager) Stop(name string) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	instance, exists := m.instances[name]
	if !exists {
		return fmt.Errorf("panel %s not found", name)
	}

	// If launched via remote control (no command), close window
	if instance.WindowID != "" && instance.Remote != nil {
		if err := instance.Remote.CloseWindow(instance.WindowTitle); err != nil {
			return fmt.Errorf("failed to close window: %w", err)
		}
	} else if instance.Command != nil {
		// Otherwise, kill process (old kitten panel method or multi-instance mode)
		if instance.Command.Process != nil {
			if err := instance.Command.Process.Kill(); err != nil {
				return fmt.Errorf("failed to kill panel %s: %w", name, err)
			}
			_ = instance.Command.Wait()
		}
	}

	// Remove from instances
	delete(m.instances, name)

	return nil
}

// List returns all running panel names
func (m *Manager) List() []string {
	m.mu.RLock()
	defer m.mu.RUnlock()

	names := make([]string, 0, len(m.instances))
	for name := range m.instances {
		names = append(names, name)
	}
	return names
}

// Wait waits for all panels to exit
func (m *Manager) Wait() {
	m.mu.RLock()
	instances := make([]*Instance, 0, len(m.instances))
	for _, instance := range m.instances {
		instances = append(instances, instance)
	}
	m.mu.RUnlock()

	// Wait for each instance
	for _, instance := range instances {
		// Skip if launched via remote control (no Command to wait on)
		if instance.Command != nil {
			_ = instance.Command.Wait()
		}
	}
}
