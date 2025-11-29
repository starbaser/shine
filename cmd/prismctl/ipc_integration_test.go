package main

import (
	"context"
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/creachadair/jrpc2/handler"
	"github.com/starbased-co/shine/pkg/rpc"
)

// TestPrismctlIPC_PrismLifecycle tests full prism up → list → down flow
func TestPrismctlIPC_PrismLifecycle(t *testing.T) {
	tmpDir := t.TempDir()
	sockPath := filepath.Join(tmpDir, "prism.sock")

	// Create mock supervisor
	sup := &mockSupervisor{
		prisms: make(map[string]*mockPrism),
	}

	// Create handlers
	h := &mockRPCHandlers{supervisor: sup}

	mux := handler.Map{
		"prism/up":   handler.New(h.handleUp),
		"prism/down": handler.New(h.handleDown),
		"prism/list": handler.New(h.handleList),
		"prism/fg":   handler.New(h.handleFg),
		"prism/bg":   handler.New(h.handleBg),
	}

	// Start server
	srv := rpc.NewServer(sockPath, mux, nil)
	if err := srv.Start(); err != nil {
		t.Fatalf("Start() error: %v", err)
	}
	defer srv.Stop(context.Background())

	time.Sleep(10 * time.Millisecond)

	// Create client
	client, err := rpc.NewPrismClient(sockPath)
	if err != nil {
		t.Fatalf("NewPrismClient() error: %v", err)
	}
	defer client.Close()

	ctx := context.Background()

	// Test 1: List prisms (should be empty)
	listResult, err := client.List(ctx)
	if err != nil {
		t.Fatalf("List() error: %v", err)
	}
	if len(listResult.Prisms) != 0 {
		t.Errorf("initial List() = %d prisms, want 0", len(listResult.Prisms))
	}

	// Test 2: Start prism
	upResult, err := client.Up(ctx, "clock")
	if err != nil {
		t.Fatalf("Up() error: %v", err)
	}
	if upResult.PID == 0 {
		t.Error("Up() returned PID=0")
	}
	if upResult.State != "fg" {
		t.Errorf("Up() State = %q, want fg", upResult.State)
	}

	// Test 3: List prisms (should have one)
	listResult, err = client.List(ctx)
	if err != nil {
		t.Fatalf("List() after up error: %v", err)
	}
	if len(listResult.Prisms) != 1 {
		t.Errorf("List() after up = %d prisms, want 1", len(listResult.Prisms))
	}
	if listResult.Prisms[0].Name != "clock" {
		t.Errorf("prism name = %q, want clock", listResult.Prisms[0].Name)
	}
	if listResult.Prisms[0].State != "fg" {
		t.Errorf("prism state = %q, want fg", listResult.Prisms[0].State)
	}

	// Test 4: Start second prism
	upResult2, err := client.Up(ctx, "bar")
	if err != nil {
		t.Fatalf("Up(bar) error: %v", err)
	}
	if upResult2.State != "fg" {
		t.Errorf("Up(bar) State = %q, want fg (should be new foreground)", upResult2.State)
	}

	// Test 5: List prisms (should have two, bar in fg)
	listResult, err = client.List(ctx)
	if err != nil {
		t.Fatalf("List() after second up error: %v", err)
	}
	if len(listResult.Prisms) != 2 {
		t.Errorf("List() after second up = %d prisms, want 2", len(listResult.Prisms))
	}

	// Verify MRU ordering: bar should be fg, clock should be bg
	var barFound, clockFound bool
	for _, p := range listResult.Prisms {
		if p.Name == "bar" {
			barFound = true
			if p.State != "fg" {
				t.Errorf("bar state = %q, want fg", p.State)
			}
		}
		if p.Name == "clock" {
			clockFound = true
			if p.State != "bg" {
				t.Errorf("clock state = %q, want bg", p.State)
			}
		}
	}
	if !barFound || !clockFound {
		t.Error("expected both bar and clock in list")
	}

	// Test 6: Stop prism
	downResult, err := client.Down(ctx, "clock")
	if err != nil {
		t.Fatalf("Down() error: %v", err)
	}
	if !downResult.Stopped {
		t.Error("Down() returned Stopped=false")
	}

	// Test 7: List prisms (should have one)
	listResult, err = client.List(ctx)
	if err != nil {
		t.Fatalf("List() after down error: %v", err)
	}
	if len(listResult.Prisms) != 1 {
		t.Errorf("List() after down = %d prisms, want 1", len(listResult.Prisms))
	}
	if listResult.Prisms[0].Name != "bar" {
		t.Errorf("remaining prism = %q, want bar", listResult.Prisms[0].Name)
	}
}

// TestPrismctlIPC_ForegroundBackgroundSwitch tests fg/bg operations
func TestPrismctlIPC_ForegroundBackgroundSwitch(t *testing.T) {
	tmpDir := t.TempDir()
	sockPath := filepath.Join(tmpDir, "prism.sock")

	sup := &mockSupervisor{
		prisms: make(map[string]*mockPrism),
	}

	h := &mockRPCHandlers{supervisor: sup}

	mux := handler.Map{
		"prism/up":   handler.New(h.handleUp),
		"prism/fg":   handler.New(h.handleFg),
		"prism/bg":   handler.New(h.handleBg),
		"prism/list": handler.New(h.handleList),
	}

	srv := rpc.NewServer(sockPath, mux, nil)
	if err := srv.Start(); err != nil {
		t.Fatalf("Start() error: %v", err)
	}
	defer srv.Stop(context.Background())

	time.Sleep(10 * time.Millisecond)

	client, err := rpc.NewPrismClient(sockPath)
	if err != nil {
		t.Fatalf("NewPrismClient() error: %v", err)
	}
	defer client.Close()

	ctx := context.Background()

	// Start two prisms
	client.Up(ctx, "clock")
	client.Up(ctx, "bar")

	// bar should be fg now
	listResult, _ := client.List(ctx)
	if listResult.Prisms[0].Name != "bar" || listResult.Prisms[0].State != "fg" {
		t.Error("bar should be foreground")
	}

	// Switch clock to foreground
	fgResult, err := client.Fg(ctx, "clock")
	if err != nil {
		t.Fatalf("Fg(clock) error: %v", err)
	}
	if !fgResult.OK {
		t.Error("Fg() returned OK=false")
	}

	// Verify clock is now fg
	listResult, _ = client.List(ctx)
	if listResult.Prisms[0].Name != "clock" || listResult.Prisms[0].State != "fg" {
		t.Error("clock should be foreground after fg")
	}
	if listResult.Prisms[1].Name != "bar" || listResult.Prisms[1].State != "bg" {
		t.Error("bar should be background after fg")
	}

	// Test bg operation (should be no-op for already bg prism)
	bgResult, err := client.Bg(ctx, "bar")
	if err != nil {
		t.Fatalf("Bg(bar) error: %v", err)
	}
	if !bgResult.OK {
		t.Error("Bg() returned OK=false")
	}
}

// TestPrismctlIPC_ServiceHealth tests service health endpoint
func TestPrismctlIPC_ServiceHealth(t *testing.T) {
	tmpDir := t.TempDir()
	sockPath := filepath.Join(tmpDir, "prism.sock")

	sup := &mockSupervisor{
		prisms: make(map[string]*mockPrism),
	}

	h := &mockRPCHandlers{supervisor: sup}

	mux := handler.Map{
		"service/health": handler.New(h.handleHealth),
		"prism/up":       handler.New(h.handleUp),
	}

	srv := rpc.NewServer(sockPath, mux, nil)
	if err := srv.Start(); err != nil {
		t.Fatalf("Start() error: %v", err)
	}
	defer srv.Stop(context.Background())

	time.Sleep(10 * time.Millisecond)

	client, err := rpc.NewPrismClient(sockPath)
	if err != nil {
		t.Fatalf("NewPrismClient() error: %v", err)
	}
	defer client.Close()

	ctx := context.Background()

	// Get health with no prisms
	health, err := client.Health(ctx)
	if err != nil {
		t.Fatalf("Health() error: %v", err)
	}
	if !health.Healthy {
		t.Error("Health().Healthy = false, want true")
	}
	if health.PrismCount != 0 {
		t.Errorf("Health().PrismCount = %d, want 0", health.PrismCount)
	}

	// Start some prisms
	client.Up(ctx, "clock")
	client.Up(ctx, "bar")

	// Get health again
	health, err = client.Health(ctx)
	if err != nil {
		t.Fatalf("Health() after up error: %v", err)
	}
	if health.PrismCount != 2 {
		t.Errorf("Health().PrismCount = %d, want 2", health.PrismCount)
	}
}

// TestPrismctlIPC_ServiceShutdown tests service shutdown endpoint
func TestPrismctlIPC_ServiceShutdown(t *testing.T) {
	tmpDir := t.TempDir()
	sockPath := filepath.Join(tmpDir, "prism.sock")

	sup := &mockSupervisor{
		prisms: make(map[string]*mockPrism),
	}

	shutdownCalled := false
	mux := handler.Map{
		"service/shutdown": handler.New(func(ctx context.Context, req *rpc.ShutdownRequest) (*rpc.ShutdownResult, error) {
			shutdownCalled = true
			sup.shuttingDown = true
			return &rpc.ShutdownResult{ShuttingDown: true}, nil
		}),
	}

	srv := rpc.NewServer(sockPath, mux, nil)
	if err := srv.Start(); err != nil {
		t.Fatalf("Start() error: %v", err)
	}
	defer srv.Stop(context.Background())

	time.Sleep(10 * time.Millisecond)

	client, err := rpc.NewPrismClient(sockPath)
	if err != nil {
		t.Fatalf("NewPrismClient() error: %v", err)
	}
	defer client.Close()

	ctx := context.Background()

	// Shutdown with graceful=true
	result, err := client.Shutdown(ctx, true)
	if err != nil {
		t.Fatalf("Shutdown() error: %v", err)
	}
	if !result.ShuttingDown {
		t.Error("Shutdown() returned ShuttingDown=false")
	}
	if !shutdownCalled {
		t.Error("shutdown handler not called")
	}
}

// TestPrismctlIPC_ErrorHandling tests error handling for invalid requests
func TestPrismctlIPC_ErrorHandling(t *testing.T) {
	tmpDir := t.TempDir()
	sockPath := filepath.Join(tmpDir, "prism.sock")

	sup := &mockSupervisor{
		prisms: make(map[string]*mockPrism),
	}

	handlers := &mockRPCHandlers{supervisor: sup}

	mux := handler.Map{
		"prism/up":   handler.New(handlers.handleUp),
		"prism/down": handler.New(handlers.handleDown),
		"prism/fg":   handler.New(handlers.handleFg),
		"prism/bg":   handler.New(handlers.handleBg),
	}

	srv := rpc.NewServer(sockPath, mux, nil)
	if err := srv.Start(); err != nil {
		t.Fatalf("Start() error: %v", err)
	}
	defer srv.Stop(context.Background())

	time.Sleep(10 * time.Millisecond)

	client, err := rpc.NewPrismClient(sockPath)
	if err != nil {
		t.Fatalf("NewPrismClient() error: %v", err)
	}
	defer client.Close()

	ctx := context.Background()

	tests := []struct {
		name    string
		testFn  func() error
		wantErr bool
	}{
		{
			name: "down nonexistent prism",
			testFn: func() error {
				_, err := client.Down(ctx, "nonexistent")
				return err
			},
			wantErr: true,
		},
		{
			name: "fg nonexistent prism",
			testFn: func() error {
				_, err := client.Fg(ctx, "nonexistent")
				return err
			},
			wantErr: true,
		},
		{
			name: "bg nonexistent prism",
			testFn: func() error {
				_, err := client.Bg(ctx, "nonexistent")
				return err
			},
			wantErr: true,
		},
		{
			name: "up already running prism",
			testFn: func() error {
				// Start prism
				_, err := client.Up(ctx, "clock")
				if err != nil {
					return err
				}

				// Try to start again (should be idempotent)
				_, err = client.Up(ctx, "clock")
				return err
			},
			wantErr: false, // Should be idempotent
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := tt.testFn()
			if (err != nil) != tt.wantErr {
				t.Errorf("test error = %v, wantErr = %v", err, tt.wantErr)
			}
		})
	}
}

// TestPrismctlIPC_StateFilePersistence tests state file updates during operations
func TestPrismctlIPC_StateFilePersistence(t *testing.T) {
	tmpDir := t.TempDir()
	sockPath := filepath.Join(tmpDir, "prism.sock")
	statePath := filepath.Join(tmpDir, "test.state")

	// Create state writer
	stateWriter, err := newMockStateWriter(statePath)
	if err != nil {
		t.Fatalf("newMockStateWriter() error: %v", err)
	}
	defer stateWriter.cleanup()

	sup := &mockSupervisor{
		prisms:       make(map[string]*mockPrism),
		stateWriter:  stateWriter,
	}

	h := &mockRPCHandlers{supervisor: sup}

	mux := handler.Map{
		"prism/up":   handler.New(h.handleUp),
		"prism/down": handler.New(h.handleDown),
		"prism/fg":   handler.New(h.handleFg),
	}

	srv := rpc.NewServer(sockPath, mux, nil)
	if err := srv.Start(); err != nil {
		t.Fatalf("Start() error: %v", err)
	}
	defer srv.Stop(context.Background())

	time.Sleep(10 * time.Millisecond)

	client, err := rpc.NewPrismClient(sockPath)
	if err != nil {
		t.Fatalf("NewPrismClient() error: %v", err)
	}
	defer client.Close()

	ctx := context.Background()

	// Start prism - should update state
	_, err = client.Up(ctx, "clock")
	if err != nil {
		t.Fatalf("Up() error: %v", err)
	}

	// Verify state was written
	if len(stateWriter.prisms) != 1 {
		t.Errorf("state has %d prisms, want 1", len(stateWriter.prisms))
	}
	if _, ok := stateWriter.prisms["clock"]; !ok {
		t.Error("clock not in state")
	}
	if stateWriter.fgPrism != "clock" {
		t.Errorf("fg prism = %q, want clock", stateWriter.fgPrism)
	}

	// Start second prism
	_, err = client.Up(ctx, "bar")
	if err != nil {
		t.Fatalf("Up(bar) error: %v", err)
	}

	// Verify state updated
	if len(stateWriter.prisms) != 2 {
		t.Errorf("state has %d prisms, want 2", len(stateWriter.prisms))
	}
	if stateWriter.fgPrism != "bar" {
		t.Errorf("fg prism = %q, want bar", stateWriter.fgPrism)
	}

	// Switch foreground
	_, err = client.Fg(ctx, "clock")
	if err != nil {
		t.Fatalf("Fg(clock) error: %v", err)
	}

	// Verify fg switched in state
	if stateWriter.fgPrism != "clock" {
		t.Errorf("fg prism after switch = %q, want clock", stateWriter.fgPrism)
	}

	// Stop prism
	_, err = client.Down(ctx, "clock")
	if err != nil {
		t.Fatalf("Down() error: %v", err)
	}

	// Verify state updated
	if len(stateWriter.prisms) != 1 {
		t.Errorf("state after down has %d prisms, want 1", len(stateWriter.prisms))
	}
	if _, ok := stateWriter.prisms["clock"]; ok {
		t.Error("clock should be removed from state")
	}
}

// Mock implementations

type mockSupervisor struct {
	prisms       map[string]*mockPrism
	fgPrism      string
	shuttingDown bool
	stateWriter  *mockStateWriter
}

type mockPrism struct {
	name  string
	pid   int
	state string // "fg" or "bg"
}

func (s *mockSupervisor) start(name string) error {
	if _, exists := s.prisms[name]; exists {
		// Idempotent - already running
		return nil
	}

	prism := &mockPrism{
		name:  name,
		pid:   1000 + len(s.prisms),
		state: "fg",
	}

	// Previous fg becomes bg
	if s.fgPrism != "" {
		if p, ok := s.prisms[s.fgPrism]; ok {
			p.state = "bg"
		}
	}

	s.prisms[name] = prism
	s.fgPrism = name

	if s.stateWriter != nil {
		s.stateWriter.addPrism(name, prism.pid, true)
	}

	return nil
}

func (s *mockSupervisor) killPrism(name string) error {
	if _, exists := s.prisms[name]; !exists {
		return rpc.ErrPrismNotFound(name)
	}

	delete(s.prisms, name)

	if s.fgPrism == name {
		s.fgPrism = ""
	}

	if s.stateWriter != nil {
		s.stateWriter.removePrism(name)
	}

	return nil
}

func (s *mockSupervisor) findPrism(name string) int {
	if _, ok := s.prisms[name]; ok {
		return 0
	}
	return -1
}

func (s *mockSupervisor) setForeground(name string) error {
	if _, exists := s.prisms[name]; !exists {
		return rpc.ErrPrismNotFound(name)
	}

	// Previous fg becomes bg
	if s.fgPrism != "" && s.fgPrism != name {
		if p, ok := s.prisms[s.fgPrism]; ok {
			p.state = "bg"
		}
	}

	s.prisms[name].state = "fg"
	s.fgPrism = name

	if s.stateWriter != nil {
		s.stateWriter.setForeground(name)
	}

	return nil
}

type mockRPCHandlers struct {
	supervisor *mockSupervisor
}

func (h *mockRPCHandlers) handleUp(ctx context.Context, req *rpc.UpRequest) (*rpc.UpResult, error) {
	if req.Name == "" {
		return nil, rpc.ErrInvalidParams("name is required")
	}

	if err := h.supervisor.start(req.Name); err != nil {
		return nil, err
	}

	prism := h.supervisor.prisms[req.Name]
	return &rpc.UpResult{
		PID:   prism.pid,
		State: prism.state,
	}, nil
}

func (h *mockRPCHandlers) handleDown(ctx context.Context, req *rpc.DownRequest) (*rpc.DownResult, error) {
	if req.Name == "" {
		return nil, rpc.ErrInvalidParams("name is required")
	}

	if err := h.supervisor.killPrism(req.Name); err != nil {
		return nil, err
	}

	return &rpc.DownResult{Stopped: true}, nil
}

func (h *mockRPCHandlers) handleFg(ctx context.Context, req *rpc.FgRequest) (*rpc.FgResult, error) {
	if req.Name == "" {
		return nil, rpc.ErrInvalidParams("name is required")
	}

	wasFg := h.supervisor.fgPrism == req.Name

	if err := h.supervisor.setForeground(req.Name); err != nil {
		return nil, err
	}

	return &rpc.FgResult{
		OK:    true,
		WasFg: wasFg,
	}, nil
}

func (h *mockRPCHandlers) handleBg(ctx context.Context, req *rpc.BgRequest) (*rpc.BgResult, error) {
	if req.Name == "" {
		return nil, rpc.ErrInvalidParams("name is required")
	}

	if _, exists := h.supervisor.prisms[req.Name]; !exists {
		return nil, rpc.ErrPrismNotFound(req.Name)
	}

	wasBg := h.supervisor.prisms[req.Name].state == "bg"

	return &rpc.BgResult{
		OK:    true,
		WasBg: wasBg,
	}, nil
}

func (h *mockRPCHandlers) handleList(ctx context.Context) (*rpc.ListResult, error) {
	result := &rpc.ListResult{
		Prisms: make([]rpc.PrismInfo, 0, len(h.supervisor.prisms)),
	}

	// Add fg prism first
	if h.supervisor.fgPrism != "" {
		if p, ok := h.supervisor.prisms[h.supervisor.fgPrism]; ok {
			result.Prisms = append(result.Prisms, rpc.PrismInfo{
				Name:  p.name,
				PID:   p.pid,
				State: p.state,
			})
		}
	}

	// Add bg prisms
	for name, p := range h.supervisor.prisms {
		if name != h.supervisor.fgPrism {
			result.Prisms = append(result.Prisms, rpc.PrismInfo{
				Name:  p.name,
				PID:   p.pid,
				State: p.state,
			})
		}
	}

	return result, nil
}

func (h *mockRPCHandlers) handleHealth(ctx context.Context) (*rpc.HealthResult, error) {
	return &rpc.HealthResult{
		Healthy:    !h.supervisor.shuttingDown,
		PrismCount: len(h.supervisor.prisms),
	}, nil
}

type mockStateWriter struct {
	prisms   map[string]int // name -> pid
	fgPrism  string
	filePath string
}

func newMockStateWriter(path string) (*mockStateWriter, error) {
	f, err := os.Create(path)
	if err != nil {
		return nil, err
	}
	f.Close()

	return &mockStateWriter{
		prisms:   make(map[string]int),
		filePath: path,
	}, nil
}

func (w *mockStateWriter) addPrism(name string, pid int, fg bool) {
	w.prisms[name] = pid
	if fg {
		w.fgPrism = name
	}
}

func (w *mockStateWriter) removePrism(name string) {
	delete(w.prisms, name)
	if w.fgPrism == name {
		w.fgPrism = ""
	}
}

func (w *mockStateWriter) setForeground(name string) {
	w.fgPrism = name
}

func (w *mockStateWriter) cleanup() {
	os.Remove(w.filePath)
}
