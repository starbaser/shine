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

func (s *Server) Start() error {
	if err := os.Remove(s.sockPath); err != nil && !os.IsNotExist(err) {
		return fmt.Errorf("failed to remove stale socket: %w", err)
	}

	listener, err := net.Listen("unix", s.sockPath)
	if err != nil {
		return fmt.Errorf("failed to listen on %s: %w", s.sockPath, err)
	}

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
		if err.Error() != "EOF" {
			log.Printf("server error: %v", err)
		}
	}
}

func (s *Server) Stop(ctx context.Context) error {
	s.mu.Lock()
	if !s.running {
		s.mu.Unlock()
		return nil
	}
	s.running = false
	close(s.shutdown)
	s.mu.Unlock()

	if s.listener != nil {
		s.listener.Close()
	}

	s.mu.Lock()
	for _, srv := range s.servers {
		srv.Stop()
	}
	servers := make([]*jrpc2.Server, 0, len(s.servers))
	for _, srv := range s.servers {
		servers = append(servers, srv)
	}
	s.mu.Unlock()

	for _, srv := range servers {
		srv.Wait()
	}

	os.Remove(s.sockPath)
	return nil
}

func (s *Server) SocketPath() string {
	return s.sockPath
}

func (s *Server) Running() bool {
	s.mu.Lock()
	defer s.mu.Unlock()
	return s.running
}

func Handler[P, R any](fn func(context.Context, *P) (R, error)) handler.Func {
	return handler.New(fn)
}

func HandlerFunc[R any](fn func(context.Context) (R, error)) handler.Func {
	return handler.New(func(ctx context.Context) (R, error) {
		return fn(ctx)
	})
}
