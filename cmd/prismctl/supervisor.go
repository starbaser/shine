package main

import (
	"fmt"
	"log"
	"os"
	"os/exec"
	"sync"
	"time"

	"golang.org/x/sys/unix"
)

// supervisor manages the lifecycle of child prism processes
type supervisor struct {
	mu             sync.Mutex
	termState      *terminalState
	currentPrism   string
	currentPid     int
	shutdownCh     chan struct{}
	childExitCh    chan childExit
	swapping       bool
	pendingSwap    string
}

// childExit represents a child process exit event
type childExit struct {
	pid      int
	exitCode int
}

// newSupervisor creates a new supervisor instance
func newSupervisor(termState *terminalState) *supervisor {
	return &supervisor{
		termState:   termState,
		shutdownCh:  make(chan struct{}),
		childExitCh: make(chan childExit, 1),
	}
}

// startPrism launches the initial prism
func (s *supervisor) startPrism(prismName string) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	// Resolve binary path
	binaryPath, err := exec.LookPath(prismName)
	if err != nil {
		return fmt.Errorf("prism not found in PATH: %s (%w)", prismName, err)
	}

	log.Printf("Starting prism: %s (resolved to %s)", prismName, binaryPath)

	return s.forkExec(binaryPath, prismName)
}

// forkExec spawns a new child process
func (s *supervisor) forkExec(binaryPath, prismName string) error {
	// Create command with inherited stdin/stdout/stderr
	cmd := exec.Command(binaryPath)
	cmd.Stdin = os.Stdin
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr

	// Don't use Setpgid - child needs to be in same process group
	// to have terminal control for TUI applications

	if err := cmd.Start(); err != nil {
		return fmt.Errorf("failed to start prism: %w", err)
	}

	s.currentPrism = prismName
	s.currentPid = cmd.Process.Pid

	log.Printf("Prism started: %s (PID %d)", prismName, s.currentPid)

	return nil
}

// hotSwap performs a sequential hot-swap to a new prism
func (s *supervisor) hotSwap(newPrismName string) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	// Prevent concurrent swaps
	if s.swapping {
		s.pendingSwap = newPrismName
		log.Printf("Swap already in progress, queuing: %s", newPrismName)
		return nil
	}

	s.swapping = true
	defer func() { s.swapping = false }()

	log.Printf("Hot-swap initiated: %s -> %s", s.currentPrism, newPrismName)

	// Resolve new binary path before killing old child
	binaryPath, err := exec.LookPath(newPrismName)
	if err != nil {
		return fmt.Errorf("new prism not found in PATH: %s (%w)", newPrismName, err)
	}

	// Step 1: SIGTERM old child
	if s.currentPid > 0 {
		oldPid := s.currentPid
		log.Printf("Sending SIGTERM to PID %d", oldPid)
		if err := unix.Kill(oldPid, unix.SIGTERM); err != nil {
			log.Printf("Warning: failed to send SIGTERM: %v", err)
		}

		// Step 2: Wait for clean exit with timeout
		// CRITICAL: Release mutex while waiting to avoid deadlock with handleChildExit
		s.mu.Unlock()

		exitCh := make(chan bool, 1)
		go func() {
			// Wait for SIGCHLD to notify us of exit
			select {
			case <-s.childExitCh:
				exitCh <- true
			case <-time.After(5 * time.Second):
				exitCh <- false
			}
		}()

		cleanExit := <-exitCh

		// Re-acquire mutex before continuing
		s.mu.Lock()

		if !cleanExit {
			// Timeout - force kill
			log.Printf("Timeout waiting for clean exit, sending SIGKILL to PID %d", oldPid)
			if err := unix.Kill(oldPid, unix.SIGKILL); err != nil {
				log.Printf("Warning: failed to send SIGKILL: %v", err)
			}
			// Wait for SIGCHLD after SIGKILL (with mutex unlocked)
			s.mu.Unlock()
			<-s.childExitCh
			s.mu.Lock()
		}

		log.Printf("Old child exited")
	}

	// Step 3: CRITICAL - Reset terminal state
	log.Printf("Resetting terminal state")
	if err := s.termState.resetTerminalState(); err != nil {
		log.Printf("Warning: failed to reset terminal state: %v", err)
	}

	// Step 4: Stabilization delay
	time.Sleep(10 * time.Millisecond)

	// Step 5: Launch new child
	log.Printf("Launching new prism: %s", newPrismName)
	if err := s.forkExec(binaryPath, newPrismName); err != nil {
		return fmt.Errorf("failed to launch new prism: %w", err)
	}

	log.Printf("Hot-swap completed successfully")

	// Process pending swap if any
	if s.pendingSwap != "" {
		pending := s.pendingSwap
		s.pendingSwap = ""
		go func() {
			if err := s.hotSwap(pending); err != nil {
				log.Printf("Error processing pending swap: %v", err)
			}
		}()
	}

	return nil
}

// handleChildExit processes child exit events
func (s *supervisor) handleChildExit(pid, exitCode int) {
	s.mu.Lock()
	defer s.mu.Unlock()

	if pid != s.currentPid {
		log.Printf("Warning: received exit for unknown PID %d (current: %d)", pid, s.currentPid)
		return
	}

	log.Printf("Current child exited (PID %d, code %d)", pid, exitCode)

	// Notify any waiting hot-swap
	select {
	case s.childExitCh <- childExit{pid: pid, exitCode: exitCode}:
	default:
	}

	// Reset terminal state after ANY exit
	if err := s.termState.resetTerminalState(); err != nil {
		log.Printf("Error resetting terminal state after child exit: %v", err)
	}

	// Clear current child
	s.currentPid = 0
	s.currentPrism = ""
}

// forwardResize forwards SIGWINCH to the child process
func (s *supervisor) forwardResize() {
	s.mu.Lock()
	pid := s.currentPid
	s.mu.Unlock()

	if pid > 0 {
		// Send directly to child PID (not process group, since child is in same pgrp)
		if err := unix.Kill(pid, unix.SIGWINCH); err != nil {
			log.Printf("Warning: failed to forward SIGWINCH to PID %d: %v", pid, err)
		}
	}
}

// shutdown performs graceful shutdown
func (s *supervisor) shutdown() {
	s.mu.Lock()
	defer s.mu.Unlock()

	log.Printf("Supervisor shutdown initiated")

	// Signal shutdown
	close(s.shutdownCh)

	// Kill current child if running
	if s.currentPid > 0 {
		log.Printf("Terminating child PID %d", s.currentPid)

		// Try graceful shutdown first
		if err := unix.Kill(s.currentPid, unix.SIGTERM); err != nil {
			log.Printf("Warning: failed to send SIGTERM: %v", err)
		}

		// Wait briefly for exit
		time.Sleep(1 * time.Second)

		// Force kill if still running
		if err := unix.Kill(s.currentPid, unix.SIGKILL); err == nil {
			log.Printf("Sent SIGKILL to child")
		}
	}

	// Restore terminal to original state
	if err := s.termState.restoreTerminalState(); err != nil {
		log.Printf("Warning: failed to restore terminal state: %v", err)
	}

	log.Printf("Supervisor shutdown complete")
}

// isShuttingDown checks if shutdown is in progress
func (s *supervisor) isShuttingDown() bool {
	select {
	case <-s.shutdownCh:
		return true
	default:
		return false
	}
}
