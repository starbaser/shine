package rpc

import (
	"context"
	"fmt"
	"log"
	"net"
	"os"
	"sync"

	"github.com/creachadair/jrpc2"
	"github.com/creachadair/jrpc2/channel"
	"github.com/creachadair/jrpc2/handler"
)

// Server wraps a JSON-RPC 2.0 server over a Unix socket
type Server struct {
	sockPath string
	listener net.Listener
	mux      handler.Map
	opts     *jrpc2.ServerOptions

	mu       sync.Mutex
	running  bool
	servers  map[net.Conn]*jrpc2.Server
	shutdown chan struct{}
}

// NewServer creates a new RPC server
func NewServer(sockPath string, mux handler.Map, opts *jrpc2.ServerOptions) *Server {
	if opts == nil {
		opts = &jrpc2.ServerOptions{}
	}
	return &Server{
		sockPath: sockPath,
		mux:      mux,
		opts:     opts,
		servers:  make(map[net.Conn]*jrpc2.Server),
		shutdown: make(chan struct{}),
	}
}

// Start begins listening and accepting connections
func (s *Server) Start() error {
	// Remove stale socket if it exists
	if err := os.Remove(s.sockPath); err != nil && !os.IsNotExist(err) {
		return fmt.Errorf("failed to remove stale socket: %w", err)
	}

	listener, err := net.Listen("unix", s.sockPath)
	if err != nil {
		return fmt.Errorf("failed to listen on %s: %w", s.sockPath, err)
	}

	// Set socket permissions (user only by default)
	if err := os.Chmod(s.sockPath, 0600); err != nil {
		listener.Close()
		return fmt.Errorf("failed to set socket permissions: %w", err)
	}

	s.mu.Lock()
	s.listener = listener
	s.running = true
	s.mu.Unlock()

	go s.acceptLoop()
	return nil
}

func (s *Server) acceptLoop() {
	for {
		conn, err := s.listener.Accept()
		if err != nil {
			select {
			case <-s.shutdown:
				return
			default:
				log.Printf("accept error: %v", err)
				continue
			}
		}
		go s.serveConn(conn)
	}
}

func (s *Server) serveConn(conn net.Conn) {
	defer conn.Close()

	// Create channel from connection (newline-delimited JSON)
	ch := channel.Line(conn, conn)

	srv := jrpc2.NewServer(s.mux, s.opts)

	s.mu.Lock()
	s.servers[conn] = srv
	s.mu.Unlock()

	defer func() {
		s.mu.Lock()
		delete(s.servers, conn)
		s.mu.Unlock()
	}()

	srv.Start(ch)
	if err := srv.Wait(); err != nil {
		// Ignore EOF errors from client disconnect
		if err.Error() != "EOF" {
			log.Printf("server error: %v", err)
		}
	}
}

// Stop gracefully shuts down the server
func (s *Server) Stop(ctx context.Context) error {
	s.mu.Lock()
	if !s.running {
		s.mu.Unlock()
		return nil
	}
	s.running = false
	close(s.shutdown)
	s.mu.Unlock()

	// Close listener first to stop accepting
	if s.listener != nil {
		s.listener.Close()
	}

	// Stop all active servers
	s.mu.Lock()
	for _, srv := range s.servers {
		srv.Stop()
	}
	servers := make([]*jrpc2.Server, 0, len(s.servers))
	for _, srv := range s.servers {
		servers = append(servers, srv)
	}
	s.mu.Unlock()

	// Wait for all servers to finish
	for _, srv := range servers {
		srv.Wait()
	}

	// Remove socket file
	os.Remove(s.sockPath)
	return nil
}

// SocketPath returns the socket path
func (s *Server) SocketPath() string {
	return s.sockPath
}

// Running returns whether the server is running
func (s *Server) Running() bool {
	s.mu.Lock()
	defer s.mu.Unlock()
	return s.running
}

// Handler is a convenience function for creating a handler from a function
func Handler[P, R any](fn func(context.Context, *P) (R, error)) handler.Func {
	return handler.New(fn)
}

// HandlerFunc is a convenience type for handlers without params
func HandlerFunc[R any](fn func(context.Context) (R, error)) handler.Func {
	return handler.New(func(ctx context.Context) (R, error) {
		return fn(ctx)
	})
}
