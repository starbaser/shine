package prism

import (
	"fmt"
	"time"

	"github.com/starbased-co/shine/pkg/config"
	"github.com/starbased-co/shine/pkg/panel"
)

// LifecycleManager handles prism lifecycle operations
type LifecycleManager struct {
	prismMgr *Manager
	panelMgr *panel.Manager
	prisms   map[string]*PrismInstance
}

// PrismInstance tracks a running prism
type PrismInstance struct {
	Name       string
	BinaryPath string
	Config     *config.PrismConfig
	Panel      *panel.Instance
	StartTime  time.Time
}

// HealthStatus represents prism health information
type HealthStatus struct {
	Name      string
	Running   bool
	Uptime    time.Duration
	WindowID  string
	ProcessID int
}

// NewLifecycleManager creates a new lifecycle manager
func NewLifecycleManager(prismMgr *Manager, panelMgr *panel.Manager) *LifecycleManager {
	return &LifecycleManager{
		prismMgr: prismMgr,
		panelMgr: panelMgr,
		prisms:   make(map[string]*PrismInstance),
	}
}

// Launch starts a new prism instance
func (lm *LifecycleManager) Launch(name string, cfg *config.PrismConfig) error {
	// Check if already running
	if _, exists := lm.prisms[name]; exists {
		return fmt.Errorf("prism %s is already running", name)
	}

	// Find prism binary
	binaryPath, err := lm.prismMgr.FindPrism(name, cfg)
	if err != nil {
		return fmt.Errorf("failed to find prism: %w", err)
	}

	// Convert to panel config
	panelCfg := cfg.ToPanelConfig()

	// Launch via panel manager
	instance, err := lm.panelMgr.Launch(name, panelCfg, binaryPath)
	if err != nil {
		return fmt.Errorf("failed to launch: %w", err)
	}

	// Track instance
	lm.prisms[name] = &PrismInstance{
		Name:       name,
		BinaryPath: binaryPath,
		Config:     cfg,
		Panel:      instance,
		StartTime:  time.Now(),
	}

	return nil
}

// Stop terminates a prism instance
func (lm *LifecycleManager) Stop(name string) error {
	_, exists := lm.prisms[name]
	if !exists {
		return fmt.Errorf("prism %s is not running", name)
	}

	if err := lm.panelMgr.Stop(name); err != nil {
		return err
	}

	delete(lm.prisms, name)
	return nil
}

// Reload restarts a prism with potentially updated binary/config
func (lm *LifecycleManager) Reload(name string, cfg *config.PrismConfig) error {
	// Stop existing instance if running
	if _, exists := lm.prisms[name]; exists {
		if err := lm.Stop(name); err != nil {
			return fmt.Errorf("failed to stop existing instance: %w", err)
		}
		// Small delay to ensure clean shutdown
		time.Sleep(200 * time.Millisecond)
	}

	// Launch new instance
	return lm.Launch(name, cfg)
}

// ReloadAll reloads all running prisms
func (lm *LifecycleManager) ReloadAll(newConfig *config.Config) error {
	// Build list of prisms to reload
	toReload := make(map[string]*config.PrismConfig)

	for name, cfg := range newConfig.Prisms {
		if cfg.Enabled {
			toReload[name] = cfg
		}
	}

	// Stop prisms that are no longer enabled
	for name := range lm.prisms {
		if _, shouldRun := toReload[name]; !shouldRun {
			if err := lm.Stop(name); err != nil {
				return fmt.Errorf("failed to stop %s: %w", name, err)
			}
		}
	}

	// Reload or launch prisms
	for name, cfg := range toReload {
		if err := lm.Reload(name, cfg); err != nil {
			return fmt.Errorf("failed to reload %s: %w", name, err)
		}
		// Delay between launches for single-instance mode
		time.Sleep(300 * time.Millisecond)
	}

	return nil
}

// Health checks prism health status
func (lm *LifecycleManager) Health(name string) (*HealthStatus, error) {
	instance, exists := lm.prisms[name]
	if !exists {
		return &HealthStatus{
			Name:    name,
			Running: false,
		}, nil
	}

	status := &HealthStatus{
		Name:      name,
		Running:   true,
		Uptime:    time.Since(instance.StartTime),
		WindowID:  instance.Panel.WindowID,
		ProcessID: 0,
	}

	if instance.Panel.Command != nil && instance.Panel.Command.Process != nil {
		status.ProcessID = instance.Panel.Command.Process.Pid
	}

	return status, nil
}

// List returns names of all running prisms
func (lm *LifecycleManager) List() []string {
	names := make([]string, 0, len(lm.prisms))
	for name := range lm.prisms {
		names = append(names, name)
	}
	return names
}
