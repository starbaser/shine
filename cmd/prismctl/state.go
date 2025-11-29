package main

import (
	"log"
	"time"

	"github.com/starbased-co/shine/pkg/state"
)

// StateManager wraps state.PrismStateWriter and provides methods
// that the supervisor calls on state changes
type StateManager struct {
	writer   *state.PrismStateWriter
	instance string
}

// newStateManager creates a new state manager
func newStateManager(statePath, instance string) (*StateManager, error) {
	writer, err := state.NewPrismStateWriter(statePath)
	if err != nil {
		return nil, err
	}

	// Initialize instance name
	writer.SetInstance(instance)

	return &StateManager{
		writer:   writer,
		instance: instance,
	}, nil
}

// OnPrismStarted is called when a prism starts or launches
func (s *StateManager) OnPrismStarted(name string, pid int, fg bool) {
	log.Printf("State: prism started %s (PID %d, fg=%v)", name, pid, fg)

	// Add prism to state
	if _, err := s.writer.AddPrism(name, int32(pid), fg); err != nil {
		log.Printf("Warning: failed to add prism to state: %v", err)
	}
}

// OnPrismStopped is called when a prism stops or is killed
func (s *StateManager) OnPrismStopped(name string) {
	log.Printf("State: prism stopped %s", name)

	// Remove prism from state
	s.writer.RemovePrism(name)
}

// OnForegroundChanged is called when foreground prism changes
func (s *StateManager) OnForegroundChanged(name string) {
	log.Printf("State: foreground changed to %s", name)

	// Update foreground prism
	s.writer.SetForeground(name)
}

// OnPrismResumed is called when a background prism is resumed to foreground
// This is handled by OnForegroundChanged
func (s *StateManager) OnPrismResumed(name string) {
	s.OnForegroundChanged(name)
}

// UpdatePrism updates a prism's state in the mmap file
func (s *StateManager) UpdatePrism(index int, name string, pid int, fg bool, restarts uint8) {
	stateVal := state.PrismStateBg
	if fg {
		stateVal = state.PrismStateFg
	}

	if err := s.writer.SetPrism(index, name, int32(pid), stateVal, restarts, time.Now().UnixMilli()); err != nil {
		log.Printf("Warning: failed to update prism in state: %v", err)
	}
}

// Sync forces a sync to disk
func (s *StateManager) Sync() error {
	return s.writer.Sync()
}

// Close closes the state writer
func (s *StateManager) Close() error {
	return s.writer.Close()
}

// Remove closes and removes the state file
func (s *StateManager) Remove() error {
	log.Printf("State: removing state file at %s", s.writer.Path())
	return s.writer.Remove()
}

// Path returns the state file path
func (s *StateManager) Path() string {
	return s.writer.Path()
}
