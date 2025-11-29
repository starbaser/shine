package main

import (
	"context"
	"fmt"
	"log"
	"os"
	"time"

	"github.com/creachadair/jrpc2"
	"github.com/starbased-co/shine/pkg/paths"
	"github.com/starbased-co/shine/pkg/rpc"
)

// startRPCServer creates and starts the JSON-RPC 2.0 server
func startRPCServer(instance string, supervisor *supervisor, stateMgr *StateManager) (*rpc.Server, error) {
	// Create runtime directory
	runtimeDir := paths.RuntimeDir()
	if err := os.MkdirAll(runtimeDir, 0700); err != nil {
		return nil, fmt.Errorf("failed to create runtime directory: %w", err)
	}

	// Get socket path
	socketPath := paths.PrismSocket(instance)

	// Create handler map
	handlers := newRPCHandlers(supervisor, stateMgr)

	// Server options with logging
	opts := &jrpc2.ServerOptions{
		Logger: func(text string) {
			log.Printf("RPC: %s", text)
		},
	}

	// Create RPC server
	server := rpc.NewServer(socketPath, handlers, opts)

	// Start server
	if err := server.Start(); err != nil {
		return nil, fmt.Errorf("failed to start RPC server: %w", err)
	}

	log.Printf("RPC server listening on: %s", socketPath)

	return server, nil
}

// stopRPCServer gracefully stops the RPC server
func stopRPCServer(server *rpc.Server) {
	if server == nil {
		return
	}

	log.Printf("Stopping RPC server...")

	// Create context with timeout for graceful shutdown
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	if err := server.Stop(ctx); err != nil {
		log.Printf("Warning: error stopping RPC server: %v", err)
	}

	log.Printf("RPC server stopped")
}
