# Code Tour: Shine Desktop Shell

## Overview

**Shine** is a three-tier process supervision system for running Bubble Tea TUIs in Kitty layer shell panels.

```
shine (CLI) → shinectl (service manager) → prismctl (panel supervisor) → TUI prisms
```

---

## Part 1: `cmd/shine` - User CLI

### Entry Point

[`cmd/shine/main.go#L1-40`](cmd/shine/main.go#L1-40)
```go
const version = "0.2.0"

func main() {
    if len(os.Args) < 2 {
        showUsage()
        os.Exit(1)
    }

    // Command routing
    switch os.Args[1] {
    case "start":   cmdStart()
    case "stop":    cmdStop()
    case "status":  cmdStatus()
    case "reload":  cmdReload()
    case "logs":    cmdLogs(os.Args[2:])
    case "help":    cmdHelp(os.Args[2:])
    }
}
```

### Core Commands

[`cmd/shine/commands.go#L30-80`](cmd/shine/commands.go#L30-80) - **start** spawns shinectl in background:
```go
func cmdStart() {
    // Check if already running
    if _, err := os.Stat(paths.ShinectlSocket()); err == nil {
        client, err := rpc.NewShinectlClient(paths.ShinectlSocket())
        if err == nil {
            defer client.Close()
            printError("Shine is already running")
            os.Exit(1)
        }
    }

    // Spawn shinectl in background
    cmd := exec.Command("shinectl")
    cmd.Start()
}
```

[`cmd/shine/commands.go#L337-370`](cmd/shine/commands.go#L337-370) - **status** uses mmap fast-path:
```go
func cmdStatus() {
    // Try mmap state first (100x faster than RPC!)
    if reader, err := state.OpenPrismStateReader(statePath); err == nil {
        defer reader.Close()
        if state, err := reader.Read(); err == nil {
            displayStatus(state)
            return
        }
    }
    // Fallback to RPC
    client, _ := rpc.NewShinectlClient(paths.ShinectlSocket())
}
```

### Output Formatting

[`cmd/shine/output.go#L15-40`](cmd/shine/output.go#L15-40) - Lipgloss styled terminal output:
```go
var (
    successStyle = lipgloss.NewStyle().Foreground(lipgloss.Color("10")) // green
    errorStyle   = lipgloss.NewStyle().Foreground(lipgloss.Color("9"))  // red
    dimStyle     = lipgloss.NewStyle().Foreground(lipgloss.Color("8"))  // gray
)

func printSuccess(msg string) {
    fmt.Println(successStyle.Render("✓ " + msg))
}
```

---

## Part 2: `cmd/shinectl` - Service Manager

### Main Loop

[`cmd/shinectl/main.go#L35-90`](cmd/shinectl/main.go#L35-90):
```go
func main() {
    // Load config
    cfg, _ := config.LoadOrDefault(paths.DefaultConfigPath())

    // Create panel manager
    pm := NewPanelManager()

    // Create state manager (mmap)
    stateMgr, _ := newStateManager()

    // Start RPC server
    startRPCServer(pm, stateMgr, cfgPath)

    // Spawn panels from config
    for name, prism := range cfg.Prisms {
        if prism.Enabled {
            pm.SpawnPanel(prism, name)
        }
    }

    // Handle signals (SIGHUP = reload, SIGTERM = shutdown)
    handleSignals(pm, cfgPath)
}
```

### Panel Manager

[`cmd/shinectl/panel_manager.go#L20-60`](cmd/shinectl/panel_manager.go#L20-60) - Core panel tracking:
```go
type Panel struct {
    Name       string
    Instance   string
    WindowID   string
    PID        int           // prismctl process PID (retrieved via kitten @ ls)
    SocketPath string
    RPCClient  *rpc.PrismClient
}

type PanelManager struct {
    mu           sync.Mutex
    panels       map[string]*Panel
    kittySocket  string
    restartState map[string]map[string]*PrismRestartState // restart tracking
}
```

[`cmd/shinectl/panel_manager.go#L100-150`](cmd/shinectl/panel_manager.go#L100-150) - **Spawning a panel** via `kitten @`:
```go
func (pm *PanelManager) SpawnPanel(entry *PrismEntry, instanceName string) (*Panel, error) {
    panelCfg := entry.PrismConfig.ToPanelConfig()

    // Build kitten panel args
    args := panelCfg.ToRemoteControlArgs("prismctl " + entry.PrismConfig.Name + " " + instanceName)

    // Execute: kitty @ launch --type=os-panel prismctl clock panel-0
    cmd := exec.Command("kitty", args...)
    output, _ := cmd.Output()

    // Parse window ID from output
    windowID := strings.TrimSpace(string(output))

    // Get PID via kitten @ ls
    pid := pm.getPIDFromWindowID(windowID)

    return &Panel{
        Instance:   instanceName,
        WindowID:   windowID,
        PID:        pid,
        SocketPath: paths.PrismSocket(instanceName),
    }, nil
}
```

### RPC Server Setup

[`cmd/shinectl/ipc_server.go#L16-48`](cmd/shinectl/ipc_server.go#L16-48):
```go
func startRPCServer(pm *PanelManager, stateMgr *StateManager, cfgPath string) error {
    h := &Handlers{pm: pm, state: stateMgr, cfgPath: cfgPath}

    // JSON-RPC 2.0 method routing
    mux := handler.Map{
        "panel/list":       rpc.HandlerFunc(h.handlePanelList),
        "panel/spawn":      rpc.Handler(h.handlePanelSpawn),
        "panel/kill":       rpc.Handler(h.handlePanelKill),
        "service/status":   rpc.HandlerFunc(h.handleServiceStatus),
        "config/reload":    rpc.HandlerFunc(h.handleConfigReload),
        // Notification handlers (prismctl → shinectl)
        "prism/started":    rpc.Handler(h.handlePrismStarted),
        "prism/crashed":    rpc.Handler(h.handlePrismCrashed),
    }

    rpcServer = rpc.NewServer(paths.ShinectlSocket(), mux, nil)
    return rpcServer.Start()
}
```

### Restart Policy

[`cmd/shinectl/panel_manager.go#L200-280`](cmd/shinectl/panel_manager.go#L200-280):
```go
type PrismRestartState struct {
    RestartCount      int
    RestartTimestamps []time.Time
    ExplicitlyStopped bool
}

func (pm *PanelManager) TriggerRestartPolicy(panelInstance, prismName string, exitCode int) {
    entry := pm.getEntry(panelInstance)

    // Evaluate policy
    switch entry.Restart {
    case "no":
        return  // Never restart
    case "on-failure":
        if exitCode == 0 { return }  // Only restart on failure
    case "unless-stopped":
        if state.ExplicitlyStopped { return }
    case "always":
        // Always restart
    }

    // Check rate limit (max_restarts per hour)
    if entry.MaxRestarts > 0 && state.RestartCount >= entry.MaxRestarts {
        return
    }

    // Async restart after delay
    go pm.restartPrismAsync(panel, prismName, entry.RestartDelay)
}
```

---

## Part 3: `cmd/prismctl` - Panel Supervisor

### Main Initialization

[`cmd/prismctl/main.go#L36-132`](cmd/prismctl/main.go#L36-132):
```go
func main() {
    prismName := os.Args[1]
    instanceName := os.Args[2]  // Optional, defaults to prismName

    // Save terminal state (CRITICAL for restoration)
    termState, _ := newTerminalState()

    // Create mmap state file
    stateMgr, _ := newStateManager(paths.PrismState(instanceName), instanceName)
    defer stateMgr.Remove()  // Clean up on exit

    // Notification manager (bidirectional to shinectl)
    notifyMgr := newNotificationManager(instanceName)
    defer notifyMgr.Close()

    // Create supervisor
    sup := newSupervisor(termState, stateMgr, notifyMgr)

    // Signal handler
    sigHandler := newSignalHandler(sup)
    defer sigHandler.stop()

    // Start RPC server
    rpcServer, _ := startRPCServer(instanceName, sup, stateMgr)

    // Launch initial prism
    sup.startPrism(prismName)

    // Run signal loop (blocks)
    sigHandler.run()
}
```

### Supervisor - MRU Process Management

[`cmd/prismctl/supervisor.go#L16-67`](cmd/prismctl/supervisor.go#L16-67):
```go
type prismState int
const (
    prismForeground prismState = iota  // Visible, receiving I/O
    prismBackground                     // Running, no I/O relay
)

type prismInstance struct {
    name      string
    pid       int
    state     prismState
    ptyMaster *os.File  // Child's PTY master FD
}

type supervisor struct {
    mu           sync.Mutex
    termState    *terminalState
    prismList    []prismInstance    // MRU: [0] = foreground
    surface      *surfaceState      // Active I/O relay
    stateManager *StateManager
    notifyMgr    *NotificationManager
}
```

### Launching a Prism

[`cmd/prismctl/supervisor.go#L106-191`](cmd/prismctl/supervisor.go#L106-191):
```go
func (s *supervisor) launchAndForeground(prismName string) error {
    // Resolve binary
    binaryPath, _ := exec.LookPath(prismName)

    // Move current foreground to background
    if len(s.prismList) > 0 {
        s.prismList[0].state = prismBackground
    }

    // Allocate PTY pair
    ptyMaster, ptySlave, _ := allocatePTY()

    // Sync terminal size (Real PTY → child PTY)
    syncTerminalSize(int(os.Stdin.Fd()), int(ptyMaster.Fd()))

    // CRITICAL: Reset terminal state before fork
    s.termState.resetTerminalState()
    time.Sleep(10 * time.Millisecond)  // Stabilization delay

    // Fork/exec with PTY as controlling terminal
    cmd := exec.Command(binaryPath)
    cmd.Stdin = ptySlave
    cmd.Stdout = ptySlave
    cmd.Stderr = ptySlave
    cmd.SysProcAttr = &syscall.SysProcAttr{
        Setsid:  true,  // New session
        Setctty: true,  // PTY becomes controlling terminal
        Ctty:    0,     // stdin in child
    }
    cmd.Start()
    ptySlave.Close()  // Close in parent

    // Add to MRU list at position [0]
    s.prismList = append([]prismInstance{{
        name: prismName, pid: cmd.Process.Pid,
        state: prismForeground, ptyMaster: ptyMaster,
    }}, s.prismList...)

    // Start surface relay
    s.activateSurfaceToForeground()

    // Notify state & shinectl
    s.notifyMgr.OnPrismStarted(prismName, cmd.Process.Pid)
    return nil
}
```

### Surface - Bidirectional I/O Relay

[`cmd/prismctl/surface.go#L14-79`](cmd/prismctl/surface.go#L14-79):
```go
type surfaceState struct {
    ctx       context.Context
    cancel    context.CancelFunc
    childPTY  *os.File
    active    bool
}

// activateSurface: Real PTY (stdin/stdout) ↔ child PTY master
func activateSurface(ctx context.Context, realPTY, childPTY *os.File) (*surfaceState, error) {
    surfaceCtx, cancel := context.WithCancel(ctx)
    state := &surfaceState{ctx: surfaceCtx, cancel: cancel, childPTY: childPTY, active: true}

    // Real PTY → child PTY (user input)
    go func() {
        io.Copy(childPTY, realPTY)
    }()

    // child PTY → Real PTY (prism output)
    go func() {
        io.Copy(os.Stdout, childPTY)
    }()

    return state, nil
}

func deactivateSurface(state *surfaceState) {
    state.cancel()
    // Force interrupt by setting deadline
    state.childPTY.SetReadDeadline(time.Unix(0, 0))
    state.active = false
}
```

### Hot-Swap Surface

[`cmd/prismctl/supervisor.go#L507-535`](cmd/prismctl/supervisor.go#L507-535):
```go
func (s *supervisor) swapSurface() error {
    startTime := time.Now()

    // Stop current surface
    deactivateSurface(s.surface)

    // Clear screen AFTER stopping old, BEFORE starting new
    os.Stdout.WriteString("\x1b[2J\x1b[H\x1b[0m")

    // Start new surface to foreground
    s.activateSurfaceToForeground()

    // Log latency (target: <50ms)
    if swapLatency := time.Since(startTime); swapLatency > 50*time.Millisecond {
        log.Printf("Warning: swap latency exceeded 50ms: %v", swapLatency)
    }
    return nil
}
```

### Signal Handler

[`cmd/prismctl/signals.go#L37-79`](cmd/prismctl/signals.go#L37-79):
```go
func (sh *signalHandler) run() {
    for sig := range sh.sigCh {
        switch sig {
        case unix.SIGCHLD:
            sh.handleSIGCHLD()  // Reap zombies, handle exits
        case unix.SIGINT:
            // Ctrl+C: kill foreground prism, or shutdown if none
            if sh.handleSIGINT() { return }
        case unix.SIGTERM, unix.SIGHUP:
            sh.supervisor.shutdown()
            return
        case unix.SIGWINCH:
            sh.supervisor.propagateResize()  // Forward to ALL prisms
        }
    }
}

func (sh *signalHandler) handleSIGCHLD() {
    for {
        pid, _ := unix.Wait4(-1, &status, unix.WNOHANG, nil)
        if pid <= 0 { break }

        exitCode := status.ExitStatus()
        sh.supervisor.handleChildExit(pid, exitCode)
    }
}
```

### Terminal State Management

> **⚠️ CRITICAL**: [`cmd/prismctl/terminal.go#L33-66`](cmd/prismctl/terminal.go#L33-66) - Must be called after EVERY child exit:

```go
func (ts *terminalState) resetTerminalState() error {
    // 1. Reset termios to canonical mode
    termios.Lflag |= unix.ICANON | unix.ECHO | unix.ISIG
    unix.IoctlSetTermios(ts.fd, unix.TCSETS, termios)

    // 2. Visual reset sequences
    resetSeq := []byte{
        0x1b, '[', '0', 'm',                     // SGR reset
        0x1b, '[', '?', '1', '0', '4', '9', 'l', // Exit alt screen
        0x1b, '[', '?', '2', '5', 'h',           // Show cursor
        0x1b, '[', '?', '1', '0', '0', '0', 'l', // Disable mouse
    }
    unix.Write(ts.fd, resetSeq)
    return nil
}
```

---

## Part 4: `pkg/rpc` - JSON-RPC 2.0 Infrastructure

### Server

[`pkg/rpc/server.go#L16-60`](pkg/rpc/server.go#L16-60):
```go
type Server struct {
    sockPath string
    listener net.Listener
    mux      handler.Map
    servers  map[net.Conn]*jrpc2.Server  // Multiple concurrent clients
}

func (s *Server) acceptLoop() {
    for {
        conn, _ := s.listener.Accept()
        go s.serveConn(conn)  // Each client gets own jrpc2.Server
    }
}

func (s *Server) serveConn(conn net.Conn) {
    ch := channel.Line(conn, conn)  // Newline-delimited JSON
    srv := jrpc2.NewServer(s.mux, s.opts)
    srv.Start(ch)
    srv.Wait()
}
```

### Typed Clients

[`pkg/rpc/client.go#L75-136`](pkg/rpc/client.go#L75-136):
```go
type PrismClient struct { *Client }

func (c *PrismClient) Up(ctx context.Context, name string) (*UpResult, error) {
    var result UpResult
    err := c.Call(ctx, "prism/up", &UpRequest{Name: name}, &result)
    return &result, err
}

func (c *PrismClient) List(ctx context.Context) (*ListResult, error) {
    var result ListResult
    err := c.Call(ctx, "prism/list", nil, &result)
    return &result, err
}

type ShinectlClient struct { *Client }

// Notifications (prismctl → shinectl)
func (c *ShinectlClient) NotifyPrismCrashed(ctx context.Context, panel, name string, exitCode, signal int) error {
    return c.Notify(ctx, "prism/crashed", &PrismCrashedNotification{
        Panel: panel, Name: name, ExitCode: exitCode, Signal: signal,
    })
}
```

### Error Codes

[`pkg/rpc/errors.go#L8-27`](pkg/rpc/errors.go#L8-27):
```go
const (
    CodePrismNotFound   = -32001
    CodePrismNotRunning = -32002
    CodePanelNotFound   = -32004
    CodeResourceBusy    = -32007
    CodeOperationFailed = -32008
)

func ErrPrismNotFound(name string) error {
    return jrpc2.Errorf(CodePrismNotFound, "prism not found: %s", name)
}
```

---

## Part 5: `pkg/state` - Memory-Mapped State

### Fixed-Size Binary Structures

[`pkg/state/types.go#L41-100`](pkg/state/types.go#L41-100):
```go
// PrismEntry: 80 bytes per prism
type PrismEntry struct {
    NameLen  uint8     // 1 byte
    Name     [63]byte  // 63 bytes (null-padded)
    PID      int32     // 4 bytes
    State    uint8     // 0=bg, 1=fg
    Restarts uint8
    _padding [2]byte
    StartMs  int64     // Unix ms
}

// PrismRuntimeState: 1424 bytes total
type PrismRuntimeState struct {
    Version     uint64           // Sequence counter (odd=writing)
    InstanceLen uint8
    Instance    [63]byte
    FgPrismLen  uint8
    FgPrism     [63]byte
    PrismCount  uint8
    _padding    [3]byte
    Prisms      [16]PrismEntry   // Max 16 prisms
}
```

### Sequence-Based Lock-Free Writes

[`pkg/state/writer.go#L165-176`](pkg/state/writer.go#L165-176):
```go
// beginWrite: version → odd (writing)
func (w *PrismStateWriter) beginWrite() {
    v := atomic.LoadUint64(&w.ptr.Version)
    atomic.StoreUint64(&w.ptr.Version, v+1)
}

// endWrite: version → even (complete), then sync
func (w *PrismStateWriter) endWrite() {
    v := atomic.LoadUint64(&w.ptr.Version)
    atomic.StoreUint64(&w.ptr.Version, v+1)
    w.mmap.Sync()
}
```

### Lock-Free Consistent Reads

[`pkg/state/reader.go#L35-58`](pkg/state/reader.go#L35-58):
```go
func (r *PrismStateReader) Read() (*PrismRuntimeState, error) {
    for i := 0; i < MaxReadRetries; i++ {
        v1 := atomic.LoadUint64(&r.ptr.Version)
        if v1%2 != 0 { continue }  // Writer in progress, retry

        state := *r.ptr  // Copy state

        v2 := atomic.LoadUint64(&r.ptr.Version)
        if v1 == v2 { return &state, nil }  // Consistent!
    }
    return nil, fmt.Errorf("failed consistent read after %d retries", MaxReadRetries)
}
```

---

## Part 6: `pkg/config` - Configuration System

### Three Discovery Types

[`pkg/config/discovery.go#L17-24`](pkg/config/discovery.go#L17-24):
```go
type PrismSource int
const (
    SourceShineToml       // shine.toml [prisms.*]
    SourcePrismDir        // Directory with prism.toml
    SourceStandaloneTOML  // Standalone .toml file
)
```

[`pkg/config/discovery.go#L34-79`](pkg/config/discovery.go#L34-79) - Discovery flow:
```go
func DiscoverPrisms(prismDirs []string) (map[string]*DiscoveredPrism, error) {
    for _, baseDir := range prismDirs {
        entries, _ := os.ReadDir(baseDir)
        for _, entry := range entries {
            if entry.IsDir() {
                // Type 1 & 2: prism.toml in directory
                discoverDirectoryPrism(baseDir, entry.Name())
            } else if strings.HasSuffix(entry.Name(), ".toml") {
                // Type 3: Standalone .toml
                discoverStandalonePrism(baseDir, entry.Name())
            }
        }
    }
}
```

### Config Merging

[`pkg/config/discovery.go#L196-262`](pkg/config/discovery.go#L196-262) - User config overrides prism defaults:
```go
func MergePrismConfigs(prismSource, userConfig *PrismConfig) *PrismConfig {
    merged := &PrismConfig{}

    // Name: user can override
    merged.Name = prismSource.Name
    if userConfig.Name != "" { merged.Name = userConfig.Name }

    // Version: always from source
    merged.Version = prismSource.Version

    // Enabled: OR (either enables it)
    merged.Enabled = userConfig.Enabled || prismSource.Enabled

    // Metadata: ONLY from prism source (user's ignored)
    merged.Metadata = prismSource.Metadata

    return merged
}
```

---

## Part 7: `pkg/panel` - Kitty Integration

### Panel Configuration

[`pkg/panel/config.go#L221-265`](pkg/panel/config.go#L221-265):
```go
type Config struct {
    Type        LayerType    // background, panel, top, overlay
    Origin      Origin       // top-left, center, bottom-right, etc.
    FocusPolicy FocusPolicy  // not-allowed, exclusive, on-demand
    Width       Dimension    // "80" (cols) or "1200px"
    Height      Dimension
    Position    Position     // x,y offset from origin
    OutputName  string       // "DP-2" (CRITICAL: must match monitor)
}
```

### Margin Calculation

[`pkg/panel/config.go#L317-397`](pkg/panel/config.go#L317-397):
```go
func (c *Config) calculateMargins() (top, left, bottom, right int, err error) {
    monWidth, monHeight, _ := getMonitorResolution(c.OutputName)

    // Convert dimensions to pixels
    panelWidth := c.Width.Value
    if !c.Width.IsPixels { panelWidth = c.Width.Value * 10 }

    switch c.Origin {
    case OriginTopRight:
        right = c.Position.X
        top = c.Position.Y
    case OriginCenter:
        // CRITICAL: center edge anchors ALL sides
        left = (monWidth/2) - (panelWidth/2) + c.Position.X
        top = (monHeight/2) - (panelHeight/2) + c.Position.Y
        right = (monWidth/2) - (panelWidth/2) - c.Position.X
        bottom = (monHeight/2) - (panelHeight/2) - c.Position.Y
    }
    return
}
```

### CLI Argument Generation

[`pkg/panel/config.go#L399-465`](pkg/panel/config.go#L399-465):
```go
func (c *Config) ToRemoteControlArgs(componentPath string) []string {
    args := []string{"@", "launch", "--type=os-panel"}

    // --os-panel edge=top
    // --os-panel columns=80
    // --os-panel margin-left=100
    for _, prop := range panelProps {
        args = append(args, "--os-panel", prop)
    }

    args = append(args, componentPath)  // "prismctl clock panel-0"
    return args
}
```

---

## Part 8: `pkg/paths` - Path Management

[`pkg/paths/paths.go#L48-76`](pkg/paths/paths.go#L48-76):
```go
func RuntimeDir() string {
    return filepath.Join("/run/user", fmt.Sprintf("%d", os.Getuid()), "shine")
}

func ShinectlSocket() string {
    return filepath.Join(RuntimeDir(), "shinectl.sock")
}

func PrismSocket(instance string) string {
    return filepath.Join(RuntimeDir(), fmt.Sprintf("prism-%s.sock", instance))
}

func PrismState(instance string) string {
    return filepath.Join(RuntimeDir(), fmt.Sprintf("prism-%s.state", instance))
}
```

---

## Keep in Mind

### ⚠️ CRITICAL: Terminal State

Never modify terminal state outside prismctl. Always call `resetTerminalState()` after every child exit. See [`cmd/prismctl/terminal.go#L33`](cmd/prismctl/terminal.go#L33).

### ⚠️ CRITICAL: Timing Constants

**DO NOT MODIFY** without testing:
- `10ms` stabilization delay after surface switch
- `20ms` shutdown grace period (SIGTERM → SIGKILL)
- `50ms` target for surface swap latency

### ⚠️ CRITICAL: Socket Naming

```
/run/user/{uid}/shine/shinectl.sock      # shinectl
/run/user/{uid}/shine/prism-{name}.sock  # prismctl
/run/user/{uid}/shine/prism-{name}.state # mmap state
```

### ⚠️ CRITICAL: Build Artifacts

Always use `make build` or `go build -o bin/`. Never bare `go build ./cmd/...`.
