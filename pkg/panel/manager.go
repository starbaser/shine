package panel

import (
	"bufio"
	"fmt"
	"os/exec"
	"sync"
)

// Instance represents a running panel instance
type Instance struct {
	Name    string
	Command *exec.Cmd
	Config  *Config
	Remote  *RemoteControl
}

// Manager manages panel instances
type Manager struct {
	instances map[string]*Instance
	mu        sync.RWMutex
}

// NewManager creates a new panel manager
func NewManager() *Manager {
	return &Manager{
		instances: make(map[string]*Instance),
	}
}

// Launch starts a panel via kitten panel
func (m *Manager) Launch(name string, config *Config, component string) (*Instance, error) {
	m.mu.Lock()
	defer m.mu.Unlock()

	// Check if already running
	if _, exists := m.instances[name]; exists {
		return nil, fmt.Errorf("panel %s already running", name)
	}

	// Build kitten args
	args := config.ToKittenArgs(component)

	// Create command
	cmd := exec.Command("kitten", args...)

	// Capture stderr for debugging
	stderr, err := cmd.StderrPipe()
	if err != nil {
		return nil, fmt.Errorf("failed to create stderr pipe for panel %s: %w", name, err)
	}

	// Start process
	if err := cmd.Start(); err != nil {
		return nil, fmt.Errorf("failed to start panel %s: %w", name, err)
	}

	// Read stderr in background (for debugging)
	go func() {
		scanner := bufio.NewScanner(stderr)
		for scanner.Scan() {
			fmt.Printf("[%s stderr] %s\n", name, scanner.Text())
		}
	}()

	// Create remote control client
	// NOTE: Kitty appends the PID to the socket path when using panels
	// So /tmp/shine-chat.sock becomes /tmp/shine-chat.sock-PID
	var remote *RemoteControl
	if config.ListenSocket != "" {
		actualSocketPath := fmt.Sprintf("%s-%d", config.ListenSocket, cmd.Process.Pid)
		remote = NewRemoteControl(actualSocketPath)
	}

	// Create instance
	instance := &Instance{
		Name:    name,
		Command: cmd,
		Config:  config,
		Remote:  remote,
	}

	// Store instance
	m.instances[name] = instance

	return instance, nil
}

// Get retrieves an instance by name
func (m *Manager) Get(name string) (*Instance, bool) {
	m.mu.RLock()
	defer m.mu.RUnlock()

	instance, exists := m.instances[name]
	return instance, exists
}

// Stop kills a panel process
func (m *Manager) Stop(name string) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	instance, exists := m.instances[name]
	if !exists {
		return fmt.Errorf("panel %s not found", name)
	}

	// Kill process
	if err := instance.Command.Process.Kill(); err != nil {
		return fmt.Errorf("failed to kill panel %s: %w", name, err)
	}

	// Wait for process to exit
	_ = instance.Command.Wait()

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
		_ = instance.Command.Wait()
	}
}
