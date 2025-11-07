package prism

import (
	"fmt"
	"os"
	"os/exec"
	"time"

	"github.com/starbased-co/shine/pkg/config"
	"github.com/starbased-co/shine/pkg/panel"
)

// Manager handles prism lifecycle and tracking
type Manager struct {
	panelMgr *panel.Manager
	prisms   map[string]*Instance
}

// Instance tracks a running prism
type Instance struct {
	Name       string
	BinaryPath string
	Config     *config.PrismConfig
	Panel      *panel.Instance
	StartTime  time.Time
}

// Health represents prism health status
type Health struct {
	Name      string
	Running   bool
	Uptime    time.Duration
	WindowID  string
	ProcessID int
}

// NewManager creates a new prism manager
func NewManager(panelMgr *panel.Manager) *Manager {
	return &Manager{
		panelMgr: panelMgr,
		prisms:   make(map[string]*Instance),
	}
}

// Launch starts a prism with the given configuration
// Binary path must already be resolved in cfg.ResolvedPath
func (m *Manager) Launch(name string, cfg *config.PrismConfig) error {
	// Check if already running
	if _, exists := m.prisms[name]; exists {
		return fmt.Errorf("prism %s is already running", name)
	}

	// Get binary path from config (already resolved by config discovery)
	binaryPath := cfg.ResolvedPath
	if binaryPath == "" {
		// Fallback: try PATH lookup for inline configs (Type 1)
		binaryName := cfg.Path
		if binaryName == "" {
			binaryName = "shine-" + name
		}

		path, err := exec.LookPath(binaryName)
		if err != nil {
			return fmt.Errorf("binary not found: %s (use config discovery or ensure binary is in PATH)", binaryName)
		}
		binaryPath = path
	}

	// Convert to panel config
	panelCfg := cfg.ToPanelConfig()

	// Launch via panel manager
	panelInstance, err := m.panelMgr.Launch(name, panelCfg, binaryPath)
	if err != nil {
		return fmt.Errorf("failed to launch: %w", err)
	}

	// Track instance
	m.prisms[name] = &Instance{
		Name:       name,
		BinaryPath: binaryPath,
		Config:     cfg,
		Panel:      panelInstance,
		StartTime:  time.Now(),
	}

	return nil
}

// Stop terminates a prism
func (m *Manager) Stop(name string) error {
	_, exists := m.prisms[name]
	if !exists {
		return fmt.Errorf("prism %s is not running", name)
	}

	if err := m.panelMgr.Stop(name); err != nil {
		return err
	}

	delete(m.prisms, name)
	return nil
}

// Reload restarts a prism with potentially updated config/binary
func (m *Manager) Reload(name string, cfg *config.PrismConfig) error {
	// Stop existing instance if running
	if _, exists := m.prisms[name]; exists {
		if err := m.Stop(name); err != nil {
			return fmt.Errorf("failed to stop: %w", err)
		}
		// Small delay for clean shutdown
		time.Sleep(200 * time.Millisecond)
	}

	// Launch with new config
	return m.Launch(name, cfg)
}

// ReloadAll reloads all prisms based on new configuration
func (m *Manager) ReloadAll(cfg *config.Config) error {
	// Determine which prisms should be running
	shouldRun := make(map[string]*config.PrismConfig)
	for name, prismCfg := range cfg.Prisms {
		if prismCfg.Enabled {
			shouldRun[name] = prismCfg
		}
	}

	// Stop prisms that are no longer enabled
	for name := range m.prisms {
		if _, enabled := shouldRun[name]; !enabled {
			if err := m.Stop(name); err != nil {
				return fmt.Errorf("failed to stop %s: %w", name, err)
			}
		}
	}

	// Reload or launch enabled prisms
	for name, prismCfg := range shouldRun {
		if err := m.Reload(name, prismCfg); err != nil {
			return fmt.Errorf("failed to reload %s: %w", name, err)
		}
		// Delay between launches for single-instance mode
		time.Sleep(300 * time.Millisecond)
	}

	return nil
}

// Health returns health status for a prism
func (m *Manager) Health(name string) (*Health, error) {
	instance, exists := m.prisms[name]
	if !exists {
		return &Health{
			Name:    name,
			Running: false,
		}, nil
	}

	health := &Health{
		Name:      name,
		Running:   true,
		Uptime:    time.Since(instance.StartTime),
		WindowID:  instance.Panel.WindowID,
		ProcessID: 0,
	}

	if instance.Panel.Command != nil && instance.Panel.Command.Process != nil {
		health.ProcessID = instance.Panel.Command.Process.Pid
	}

	return health, nil
}

// List returns names of all running prisms
func (m *Manager) List() []string {
	names := make([]string, 0, len(m.prisms))
	for name := range m.prisms {
		names = append(names, name)
	}
	return names
}

// Get returns the instance for a running prism
func (m *Manager) Get(name string) (*Instance, bool) {
	instance, ok := m.prisms[name]
	return instance, ok
}

// isExecutable checks if a file exists and is executable
func isExecutable(path string) bool {
	info, err := os.Stat(path)
	if err != nil {
		return false
	}
	return !info.IsDir() && info.Mode()&0111 != 0
}
