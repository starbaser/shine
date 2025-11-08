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

// prismState represents the current state of a prism instance
type prismState int

const (
	prismForeground prismState = iota // Currently visible
	prismBackground                    // Suspended (SIGSTOP)
)

// prismInstance represents a single prism process
type prismInstance struct {
	name  string
	pid   int
	state prismState
}

// supervisor manages the lifecycle of child prism processes
type supervisor struct {
	mu          sync.Mutex
	termState   *terminalState
	prismList   []prismInstance // MRU list: [0] = foreground, [1] = most recent background, etc.
	shutdownCh  chan struct{}
	childExitCh chan childExit
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
		prismList:   make([]prismInstance, 0),
		shutdownCh:  make(chan struct{}),
		childExitCh: make(chan childExit, 1),
	}
}

// findPrism returns the index of a prism by name, or -1 if not found
func (s *supervisor) findPrism(name string) int {
	for i, p := range s.prismList {
		if p.name == name {
			return i
		}
	}
	return -1
}

// startPrism launches the initial prism (wrapper for compatibility)
func (s *supervisor) startPrism(prismName string) error {
	return s.start(prismName)
}

// start implements idempotent launch/resume with three cases
func (s *supervisor) start(prismName string) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	// Case 1: Prism doesn't exist → launch fresh
	targetIdx := s.findPrism(prismName)
	if targetIdx == -1 {
		return s.launchAndForeground(prismName)
	}

	// Case 2: Already foreground → no-op
	if targetIdx == 0 {
		log.Printf("Prism %s already in foreground", prismName)
		return nil
	}

	// Case 3: In background → resume to foreground
	return s.resumeToForeground(targetIdx)
}

// launchAndForeground launches a new prism and brings it to foreground
func (s *supervisor) launchAndForeground(prismName string) error {
	// Resolve binary path
	binaryPath, err := exec.LookPath(prismName)
	if err != nil {
		return fmt.Errorf("prism not found in PATH: %s (%w)", prismName, err)
	}

	log.Printf("Launching new prism: %s (resolved to %s)", prismName, binaryPath)

	// Suspend current foreground if exists
	if len(s.prismList) > 0 {
		foregroundPid := s.prismList[0].pid
		log.Printf("Suspending current foreground (PID %d)", foregroundPid)
		if err := unix.Kill(foregroundPid, unix.SIGSTOP); err != nil {
			log.Printf("Warning: failed to suspend foreground: %v", err)
		}
		s.prismList[0].state = prismBackground
	}

	// Reset terminal state (CRITICAL!)
	log.Printf("Resetting terminal state")
	if err := s.termState.resetTerminalState(); err != nil {
		log.Printf("Warning: failed to reset terminal state: %v", err)
	}

	// Stabilization delay
	time.Sleep(10 * time.Millisecond)

	// Fork/exec new prism
	cmd := exec.Command(binaryPath)
	cmd.Stdin = os.Stdin
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr

	if err := cmd.Start(); err != nil {
		return fmt.Errorf("failed to start prism: %w", err)
	}

	pid := cmd.Process.Pid
	log.Printf("Prism started: %s (PID %d)", prismName, pid)

	// Add to front of MRU list
	newInstance := prismInstance{
		name:  prismName,
		pid:   pid,
		state: prismForeground,
	}
	s.prismList = append([]prismInstance{newInstance}, s.prismList...)

	return nil
}

// resumeToForeground resumes a background prism to foreground
func (s *supervisor) resumeToForeground(targetIdx int) error {
	target := s.prismList[targetIdx]
	log.Printf("Resuming prism %s (PID %d) from background", target.name, target.pid)

	// Suspend current foreground
	if len(s.prismList) > 0 && targetIdx != 0 {
		foregroundPid := s.prismList[0].pid
		log.Printf("Suspending current foreground (PID %d)", foregroundPid)
		if err := unix.Kill(foregroundPid, unix.SIGSTOP); err != nil {
			log.Printf("Warning: failed to suspend foreground: %v", err)
		}
		s.prismList[0].state = prismBackground
	}

	// Reset terminal state (CRITICAL!)
	log.Printf("Resetting terminal state")
	if err := s.termState.resetTerminalState(); err != nil {
		log.Printf("Warning: failed to reset terminal state: %v", err)
	}

	// Stabilization delay
	time.Sleep(10 * time.Millisecond)

	// Resume target prism with SIGCONT
	if err := unix.Kill(target.pid, unix.SIGCONT); err != nil {
		return fmt.Errorf("failed to resume prism: %w", err)
	}

	// Send SIGWINCH to trigger full redraw after resume
	// No delay needed - kernel queues signals properly
	if err := unix.Kill(target.pid, unix.SIGWINCH); err != nil {
		log.Printf("Warning: failed to send SIGWINCH for redraw: %v", err)
	}

	// Move target to position [0] and reorder MRU list
	// Remove from current position
	s.prismList = append(s.prismList[:targetIdx], s.prismList[targetIdx+1:]...)

	// Update state and add to front
	target.state = prismForeground
	s.prismList = append([]prismInstance{target}, s.prismList...)

	log.Printf("Prism %s resumed to foreground", target.name)

	return nil
}

// killPrism terminates a prism by name with auto-resume
func (s *supervisor) killPrism(prismName string) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	targetIdx := s.findPrism(prismName)
	if targetIdx == -1 {
		return fmt.Errorf("prism not found: %s", prismName)
	}

	target := s.prismList[targetIdx]
	pid := target.pid

	log.Printf("Killing prism %s (PID %d)", prismName, pid)

	// Send SIGTERM
	if err := unix.Kill(pid, unix.SIGTERM); err != nil {
		log.Printf("Warning: failed to send SIGTERM: %v", err)
	}

	// Wait for clean exit with timeout (release mutex to avoid deadlock)
	s.mu.Unlock()

	exitCh := make(chan bool, 1)
	go func() {
		select {
		case <-s.childExitCh:
			exitCh <- true
		case <-time.After(5 * time.Second):
			exitCh <- false
		}
	}()

	cleanExit := <-exitCh

	// Re-acquire mutex
	s.mu.Lock()

	if !cleanExit {
		// Timeout - force kill
		log.Printf("Timeout waiting for clean exit, sending SIGKILL to PID %d", pid)
		if err := unix.Kill(pid, unix.SIGKILL); err != nil {
			log.Printf("Warning: failed to send SIGKILL: %v", err)
		}
		// Wait for SIGCHLD after SIGKILL
		s.mu.Unlock()
		<-s.childExitCh
		s.mu.Lock()
	}

	log.Printf("Prism %s terminated", prismName)

	// Remove from MRU list
	s.prismList = append(s.prismList[:targetIdx], s.prismList[targetIdx+1:]...)

	// If we killed foreground AND others exist, auto-resume next
	if targetIdx == 0 && len(s.prismList) > 0 {
		s.termState.resetTerminalState()
		time.Sleep(10 * time.Millisecond)

		nextPid := s.prismList[0].pid
		nextName := s.prismList[0].name

		if err := unix.Kill(nextPid, unix.SIGCONT); err != nil {
			log.Printf("Warning: failed to resume next prism: %v", err)
		}

		// Send SIGWINCH to trigger redraw
		unix.Kill(nextPid, unix.SIGWINCH)

		s.prismList[0].state = prismForeground

		log.Printf("Auto-resumed: %s (PID %d)", nextName, nextPid)
	}

	return nil
}

// handleChildExit processes child exit events
func (s *supervisor) handleChildExit(pid, exitCode int) {
	s.mu.Lock()
	defer s.mu.Unlock()

	// Find the exited prism in MRU list
	exitedIdx := -1
	for i, p := range s.prismList {
		if p.pid == pid {
			exitedIdx = i
			break
		}
	}

	if exitedIdx == -1 {
		log.Printf("Warning: received exit for unknown PID %d", pid)
		return
	}

	exitedName := s.prismList[exitedIdx].name
	log.Printf("Child exited: %s (PID %d, code %d)", exitedName, pid, exitCode)

	// Notify any waiting kill operation
	select {
	case s.childExitCh <- childExit{pid: pid, exitCode: exitCode}:
	default:
	}

	// If foreground exited, reset terminal state
	if exitedIdx == 0 {
		if err := s.termState.resetTerminalState(); err != nil {
			log.Printf("Error resetting terminal state after child exit: %v", err)
		}
	}

	// Remove from MRU list
	s.prismList = append(s.prismList[:exitedIdx], s.prismList[exitedIdx+1:]...)

	// If foreground crashed AND others exist, auto-resume next
	if exitedIdx == 0 && len(s.prismList) > 0 {
		time.Sleep(10 * time.Millisecond)

		nextPid := s.prismList[0].pid
		nextName := s.prismList[0].name

		if err := unix.Kill(nextPid, unix.SIGCONT); err != nil {
			log.Printf("Warning: failed to resume next prism after crash: %v", err)
		}

		// Send SIGWINCH to trigger redraw
		unix.Kill(nextPid, unix.SIGWINCH)

		s.prismList[0].state = prismForeground

		log.Printf("Auto-resumed after crash: %s (PID %d)", nextName, nextPid)
	}
}

// forwardResize forwards SIGWINCH to the foreground process
func (s *supervisor) forwardResize() {
	s.mu.Lock()
	defer s.mu.Unlock()

	// Forward only to foreground prism (position [0])
	if len(s.prismList) > 0 {
		pid := s.prismList[0].pid
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

	// Kill all prisms in the list
	for _, prism := range s.prismList {
		log.Printf("Terminating prism %s (PID %d)", prism.name, prism.pid)

		// Try graceful shutdown first
		if err := unix.Kill(prism.pid, unix.SIGTERM); err != nil {
			log.Printf("Warning: failed to send SIGTERM to %s: %v", prism.name, err)
		}
	}

	// Wait briefly for graceful exits (match Kitty's signal delivery timing)
	time.Sleep(20 * time.Millisecond)

	// Force kill any remaining prisms
	for _, prism := range s.prismList {
		if err := unix.Kill(prism.pid, unix.SIGKILL); err == nil {
			log.Printf("Sent SIGKILL to %s (PID %d)", prism.name, prism.pid)
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
