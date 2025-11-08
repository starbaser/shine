package main

import (
	"log"
	"os"
	"os/signal"

	"golang.org/x/sys/unix"
)

// signalHandler manages all signal handling for prismctl
type signalHandler struct {
	sigCh      chan os.Signal
	supervisor *supervisor
}

// newSignalHandler creates and configures signal handling
func newSignalHandler(sup *supervisor) *signalHandler {
	sh := &signalHandler{
		sigCh:      make(chan os.Signal, 10),
		supervisor: sup,
	}

	// Register for all signals we need to handle
	signal.Notify(sh.sigCh,
		unix.SIGCHLD,  // Child process state change
		unix.SIGTERM,  // Termination request
		unix.SIGINT,   // Interrupt (Ctrl+C)
		unix.SIGHUP,   // Hangup (Kitty panel close)
		unix.SIGWINCH, // Window resize
	)

	return sh
}

// run processes signals in a loop until shutdown
func (sh *signalHandler) run() {
	for sig := range sh.sigCh {
		switch sig {
		case unix.SIGCHLD:
			sh.handleSIGCHLD()
		case unix.SIGTERM, unix.SIGINT, unix.SIGHUP:
			sh.handleShutdown(sig)
			return
		case unix.SIGWINCH:
			sh.handleSIGWINCH()
		}
	}
}

// handleSIGCHLD reaps zombie processes and handles child exits
func (sh *signalHandler) handleSIGCHLD() {
	// Reap all exited children
	for {
		var status unix.WaitStatus
		pid, err := unix.Wait4(-1, &status, unix.WNOHANG, nil)
		if err != nil || pid <= 0 {
			// No more children to reap
			break
		}

		// Notify supervisor of child exit
		exitCode := 0
		if status.Exited() {
			exitCode = status.ExitStatus()
			log.Printf("Child %d exited with code %d", pid, exitCode)
		} else if status.Signaled() {
			exitCode = 128 + int(status.Signal())
			log.Printf("Child %d terminated by signal %s", pid, status.Signal())
		}

		sh.supervisor.handleChildExit(pid, exitCode)
	}
}

// handleShutdown performs graceful shutdown
func (sh *signalHandler) handleShutdown(sig os.Signal) {
	log.Printf("Received %s, shutting down gracefully", sig)
	sh.supervisor.shutdown()
}

// handleSIGWINCH forwards window resize to child process
func (sh *signalHandler) handleSIGWINCH() {
	sh.supervisor.forwardResize()
}

// stop stops signal handling
func (sh *signalHandler) stop() {
	signal.Stop(sh.sigCh)
	close(sh.sigCh)
}
