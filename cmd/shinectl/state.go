package main

import (
	"log"
	"time"

	"github.com/starbased-co/shine/pkg/paths"
	"github.com/starbased-co/shine/pkg/state"
)

// StateManager handles shinectl state persistence
type StateManager struct {
	writer    *state.ShinectlStateWriter
	startTime time.Time
}

// newStateManager creates a new state manager
func newStateManager() (*StateManager, error) {
	writer, err := state.NewShinectlStateWriter(paths.ShinectlState())
	if err != nil {
		return nil, err
	}

	return &StateManager{
		writer:    writer,
		startTime: time.Now(),
	}, nil
}

// OnPanelSpawned is called when a panel is spawned
func (sm *StateManager) OnPanelSpawned(instance, name string, pid int, healthy bool) {
	_, err := sm.writer.AddPanel(instance, name, int32(pid), healthy)
	if err != nil {
		log.Printf("Failed to add panel to state: %v", err)
	}
}

// OnPanelKilled is called when a panel is killed
func (sm *StateManager) OnPanelKilled(instance string) {
	sm.writer.RemovePanel(instance)
}

// OnPanelHealthChanged is called when a panel's health changes
func (sm *StateManager) OnPanelHealthChanged(instance string, healthy bool) {
	sm.writer.SetPanelHealth(instance, healthy)
}

// OnPanelPrismStarted is called when a prism starts in a panel
func (sm *StateManager) OnPanelPrismStarted(panel, name string, pid int) {
	log.Printf("State: panel %s - prism started: %s (PID %d)", panel, name, pid)
	// Future: could track prism state in panel metadata
}

// OnPanelPrismStopped is called when a prism stops normally in a panel
func (sm *StateManager) OnPanelPrismStopped(panel, name string, exitCode int) {
	log.Printf("State: panel %s - prism stopped: %s (exit=%d)", panel, name, exitCode)
	// Future: could update prism state in panel metadata
}

// OnPanelPrismCrashed is called when a prism crashes in a panel
func (sm *StateManager) OnPanelPrismCrashed(panel, name string, exitCode, signal int) {
	log.Printf("State: panel %s - prism crashed: %s (exit=%d, signal=%d)", panel, name, exitCode, signal)
	// Future: could trigger restart policy or mark panel unhealthy
}

// OnPanelSurfaceSwitched is called when foreground prism changes in a panel
func (sm *StateManager) OnPanelSurfaceSwitched(panel, from, to string) {
	log.Printf("State: panel %s - surface switched: %s → %s", panel, from, to)
	// Future: could track current foreground prism in panel metadata
}

// Uptime returns the duration since shinectl started
func (sm *StateManager) Uptime() time.Duration {
	return time.Since(sm.startTime)
}

// Close closes the state manager
func (sm *StateManager) Close() error {
	return sm.writer.Close()
}

// Remove removes the state file
func (sm *StateManager) Remove() error {
	return sm.writer.Remove()
}
