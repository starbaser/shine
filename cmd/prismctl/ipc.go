package main

import (
	"bufio"
	"encoding/json"
	"fmt"
	"log"
	"net"
	"os"
	"path/filepath"
	"strings"
	"sync"
)

// ipcCommand represents a command sent via IPC
type ipcCommand struct {
	Action string `json:"action"` // "swap", "status", "stop"
	Prism  string `json:"prism"`  // Prism name for swap action
}

// ipcResponse represents a response from IPC
type ipcResponse struct {
	Success bool   `json:"success"`
	Message string `json:"message"`
	Data    any    `json:"data,omitempty"`
}

// ipcServer manages the Unix socket IPC server
type ipcServer struct {
	socketPath string
	listener   net.Listener
	supervisor *supervisor
	wg         sync.WaitGroup
	stopCh     chan struct{}
}

// newIPCServer creates a new IPC server
func newIPCServer(component string, supervisor *supervisor) (*ipcServer, error) {
	// Use XDG runtime directory with PID suffix
	uid := os.Getuid()
	runtimeDir := fmt.Sprintf("/run/user/%d/shine", uid)

	// Create directory if needed
	if err := os.MkdirAll(runtimeDir, 0700); err != nil {
		return nil, fmt.Errorf("failed to create runtime directory: %w", err)
	}

	// Socket path with PID to prevent conflicts on restart
	socketPath := filepath.Join(runtimeDir, fmt.Sprintf("prism-%s.%d.sock", component, os.Getpid()))

	// Remove stale socket if exists
	_ = os.Remove(socketPath)

	// Create Unix socket listener
	listener, err := net.Listen("unix", socketPath)
	if err != nil {
		return nil, fmt.Errorf("failed to create Unix socket: %w", err)
	}

	// Set socket permissions to user-only
	if err := os.Chmod(socketPath, 0600); err != nil {
		listener.Close()
		return nil, fmt.Errorf("failed to set socket permissions: %w", err)
	}

	log.Printf("IPC server listening on: %s", socketPath)

	return &ipcServer{
		socketPath: socketPath,
		listener:   listener,
		supervisor: supervisor,
		stopCh:     make(chan struct{}),
	}, nil
}

// serve starts accepting IPC connections
func (s *ipcServer) serve() {
	s.wg.Add(1)
	defer s.wg.Done()

	for {
		select {
		case <-s.stopCh:
			return
		default:
		}

		conn, err := s.listener.Accept()
		if err != nil {
			select {
			case <-s.stopCh:
				return
			default:
				log.Printf("Error accepting connection: %v", err)
				continue
			}
		}

		// Handle connection in goroutine
		s.wg.Add(1)
		go s.handleConnection(conn)
	}
}

// handleConnection processes a single IPC connection
func (s *ipcServer) handleConnection(conn net.Conn) {
	defer s.wg.Done()
	defer conn.Close()

	reader := bufio.NewReader(conn)
	line, err := reader.ReadString('\n')
	if err != nil {
		s.sendError(conn, "failed to read command")
		return
	}

	// Parse command
	var cmd ipcCommand
	if err := json.Unmarshal([]byte(strings.TrimSpace(line)), &cmd); err != nil {
		s.sendError(conn, fmt.Sprintf("invalid JSON: %v", err))
		return
	}

	// Process command
	switch cmd.Action {
	case "swap":
		s.handleSwap(conn, cmd)
	case "status":
		s.handleStatus(conn)
	case "stop":
		s.handleStop(conn)
	default:
		s.sendError(conn, fmt.Sprintf("unknown action: %s", cmd.Action))
	}
}

// handleSwap processes a hot-swap command
func (s *ipcServer) handleSwap(conn net.Conn, cmd ipcCommand) {
	if cmd.Prism == "" {
		s.sendError(conn, "prism name required for swap action")
		return
	}

	log.Printf("IPC: Received swap request to %s", cmd.Prism)

	if err := s.supervisor.hotSwap(cmd.Prism); err != nil {
		s.sendError(conn, fmt.Sprintf("swap failed: %v", err))
		return
	}

	s.sendSuccess(conn, "swap initiated", nil)
}

// handleStatus processes a status query
func (s *ipcServer) handleStatus(conn net.Conn) {
	s.supervisor.mu.Lock()
	currentPrism := s.supervisor.currentPrism
	currentPid := s.supervisor.currentPid
	s.supervisor.mu.Unlock()

	data := map[string]any{
		"prism": currentPrism,
		"pid":   currentPid,
	}

	s.sendSuccess(conn, "status ok", data)
}

// handleStop processes a stop command
func (s *ipcServer) handleStop(conn net.Conn) {
	log.Printf("IPC: Received stop request")
	s.sendSuccess(conn, "stopping", nil)

	// Trigger shutdown
	go s.supervisor.shutdown()
}

// sendSuccess sends a success response
func (s *ipcServer) sendSuccess(conn net.Conn, message string, data any) {
	resp := ipcResponse{
		Success: true,
		Message: message,
		Data:    data,
	}
	s.sendResponse(conn, resp)
}

// sendError sends an error response
func (s *ipcServer) sendError(conn net.Conn, message string) {
	resp := ipcResponse{
		Success: false,
		Message: message,
	}
	s.sendResponse(conn, resp)
}

// sendResponse sends a JSON response
func (s *ipcServer) sendResponse(conn net.Conn, resp ipcResponse) {
	data, err := json.Marshal(resp)
	if err != nil {
		log.Printf("Error marshaling response: %v", err)
		return
	}

	data = append(data, '\n')
	if _, err := conn.Write(data); err != nil {
		log.Printf("Error writing response: %v", err)
	}
}

// stop stops the IPC server
func (s *ipcServer) stop() {
	close(s.stopCh)
	s.listener.Close()
	s.wg.Wait()

	// Clean up socket file
	_ = os.Remove(s.socketPath)

	log.Printf("IPC server stopped")
}
