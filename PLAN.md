# Shine - Development Plan

**Project**: Hyprland Wayland Layer Shell TUI Desktop Shell Toolkit
**Status**: Phase 1 - Prototype
**Architecture**: Go + Bubble Tea components running in Kitty panels

---

## Executive Summary

`shine` is a toolkit for building beautiful TUI-based desktop shell components for Hyprland using the Charm ecosystem (Bubble Tea, Bubbles, Lip Gloss). Instead of implementing Wayland layer shell bindings from scratch, we leverage Kitty's battle-tested GPU-accelerated terminal with built-in layer shell support via `kitten panel`.

**Key Insight**: Kitty already solves the hard problems (GPU rendering, Wayland protocol, layer shell). We focus on building beautiful TUI components and providing a clean Go API for managing them.

---

## Architecture Decision

### Why Kitty Panel?

**Original Plan**:

- Implement Go Wayland layer shell bindings (gotk3-layershell)
- Embed VTE terminal widget in GTK window
- Run Bubble Tea inside embedded terminal

**Problems**:

- Complex: GTK + VTE + Wayland bindings
- Heavy dependencies
- Months of integration work
- Maintaining Wayland protocol code

**Chosen Approach**:

- Use `kitten panel` to launch TUI components
- Kitty handles ALL layer shell integration
- Components are pure Bubble Tea Go programs
- Clean separation: Kitty does plumbing, we do UX

**Benefits**:
✅ Zero Wayland code to maintain
✅ GPU-accelerated rendering (Kitty's OpenGL engine)
✅ Pure Bubble Tea components
✅ Battle-tested layer shell implementation
✅ Focus on component quality, not infrastructure
✅ Weeks instead of months to prototype

---

## What We're Building

### Phase 1: Prototype (Current)

**Goal**: Prove the architecture with a single working component

**Deliverables**:

1. ✅ Chat TUI component (Bubble Tea chat example)
2. ✅ Panel configuration system (Go structs ported from Kitty)
3. ✅ Panel manager (launches components via `kitten panel`)
4. ✅ Remote control client (toggle visibility programmatically)
5. ✅ TOML configuration
6. ✅ `shine` launcher command
7. ✅ `shinectl` control utility

**Scope Limitations**:

- Only one component: `shine-chat`
- Basic configuration (edge, size, margins)
- Simple remote control (toggle only)
- No hot reload, no declarative widgets, no IPC

### Phase 2: Expansion (Future)

**Additional Components**:

- `shine-bar` - Status bar with workspace indicator, clock, system stats
- `shine-dock` - Application launcher dock
- `shine-widgets` - Widget system (weather, media player, etc.)

**Advanced Features**:

- Declarative widget configuration (define widgets in TOML, not Go)
- Hot reload configuration
- IPC between components (event bus)
- Component marketplace/registry
- Theming system

---

## Project Structure

```
shine/
├── cmd/
│   ├── shine/              # Main launcher (orchestrates panels)
│   │   └── main.go
│   ├── shine-chat/         # Chat TUI component (Bubble Tea example)
│   │   └── main.go
│   └── shinectl/           # Control utility (toggle, reload, etc.)
│       └── main.go
│
├── pkg/
│   ├── panel/
│   │   ├── config.go       # LayerShellConfig (ported from Kitty)
│   │   ├── manager.go      # Panel launcher/manager
│   │   ├── remote.go       # Kitty remote control client
│   │   └── args.go         # CLI args builder (part of config.go)
│   │
│   └── config/
│       ├── loader.go       # TOML config loading
│       └── types.go        # Config structures
│
├── examples/
│   └── shine.toml          # Example configuration
│
├── docs/
│   ├── llms/               # LLM-optimized documentation
│   │   ├── research/       # Research findings
│   │   └── man/            # Manual/reference docs
│   └── ARCHITECTURE.md     # Architecture deep dive
│
├── go.mod
├── go.sum
├── PLAN.md                 # This file
└── README.md
```

---

## Key Implementation Details

### 1. Panel Configuration (pkg/panel/config.go)

**Ported from Kitty**: `kitty/types.py:72-88` (LayerShellConfig)

```go
type Config struct {
    // Layer shell properties
    Type                    LayerType      // BACKGROUND, PANEL, TOP, OVERLAY
    Edge                    Edge           // TOP, BOTTOM, LEFT, RIGHT, CENTER
    FocusPolicy             FocusPolicy    // NOT_ALLOWED, EXCLUSIVE, ON_DEMAND

    // Size (cells or pixels)
    Lines                   int            // Height in terminal lines
    Columns                 int            // Width in terminal columns
    LinesPixels             int            // Height in pixels (overrides Lines)
    ColumnsPixels           int            // Width in pixels (overrides Columns)

    // Margins
    MarginTop               int
    MarginLeft              int
    MarginBottom            int
    MarginRight             int

    // Exclusive zone
    ExclusiveZone           int
    OverrideExclusiveZone   bool

    // Behavior
    HideOnFocusLoss         bool
    SingleInstance          bool
    ToggleVisibility        bool

    // Output
    OutputName              string         // Monitor name (e.g., "DP-1")

    // Remote control
    ListenSocket            string         // Unix socket path
}

// ToKittenArgs converts Config to kitten panel CLI arguments
func (c *Config) ToKittenArgs(component string) []string
```

**Key Methods**:

- `ToKittenArgs()` - Convert Go struct to CLI flags for `kitten panel`
- `Edge.String()` - Enum to string conversion
- `FocusPolicy.String()` - Enum to string conversion

### 2. Panel Manager (pkg/panel/manager.go)

**Purpose**: Launch and manage panel instances

```go
type Manager struct {
    instances map[string]*Instance
    mu        sync.RWMutex
}

type Instance struct {
    Name       string
    Command    *exec.Cmd          // Running kitten panel process
    Config     *Config
    Remote     *RemoteControl     // For IPC
}
```

**Key Methods**:

- `Launch(name, config, component)` - Start a panel via `kitten panel`
- `Get(name)` - Retrieve instance by name
- `Stop(name)` - Kill panel process
- `List()` - List all running panels

**Implementation Notes**:

- Uses `exec.Command("kitten", args...)` to spawn panels
- Stores process handle for lifecycle management
- Thread-safe (mutex protected)

### 3. Remote Control Client (pkg/panel/remote.go)

**Purpose**: Send commands to running panels via Kitty's remote control protocol

```go
type RemoteControl struct {
    socketPath string  // Unix socket path (e.g., /tmp/shine-chat.sock)
}

func (rc *RemoteControl) ToggleVisibility() error
```

**Protocol**: JSON over Unix socket

- Connect to `unix:/tmp/shine-{component}.sock`
- Send JSON payload: `{"cmd": "resize-os-window", "action": "toggle-visibility"}`
- Close connection

**Future Commands**:

- `Show()` / `Hide()` - Explicit visibility control
- `UpdateConfig()` - Runtime reconfiguration
- `QueryState()` - Get panel status

### 4. Configuration System (pkg/config/)

**User Configuration** (`~/.config/shine/shine.toml`):

```toml
[chat]
enabled = true
edge = "bottom"
lines = 10
margin_left = 10
margin_right = 10
margin_bottom = 10
single_instance = true
hide_on_focus_loss = true
focus_policy = "on-demand"
```

**Go Structures**:

```go
type Config struct {
    Chat *ChatConfig `toml:"chat"`
}

type ChatConfig struct {
    Enabled           bool   `toml:"enabled"`
    Edge              string `toml:"edge"`
    Lines             int    `toml:"lines"`
    MarginLeft        int    `toml:"margin_left"`
    MarginRight       int    `toml:"margin_right"`
    MarginBottom      int    `toml:"margin_bottom"`
    SingleInstance    bool   `toml:"single_instance"`
    HideOnFocusLoss   bool   `toml:"hide_on_focus_loss"`
    FocusPolicy       string `toml:"focus_policy"`
}

func (cc *ChatConfig) ToPanelConfig() *panel.Config
```

**Flow**:

1. Load TOML from `~/.config/shine/shine.toml`
2. Parse into `config.Config` struct
3. Convert `ChatConfig` to `panel.Config`
4. Pass to `Manager.Launch()`

### 5. Main Launcher (cmd/shine/main.go)

**Responsibilities**:

- Load configuration from `~/.config/shine/shine.toml`
- Create panel manager
- Launch enabled components
- Stay running (orchestrator process)

**Pseudocode**:

```go
cfg := config.Load("~/.config/shine/shine.toml")
mgr := panel.NewManager()

if cfg.Chat.Enabled {
    panelCfg := cfg.Chat.ToPanelConfig()
    mgr.Launch("chat", panelCfg, "shine-chat")
}

select {} // Keep running
```

### 6. Control Utility (cmd/shinectl/main.go)

**Purpose**: CLI tool for controlling running panels

**Usage**:

```bash
shinectl toggle chat      # Toggle chat visibility
shinectl show chat        # Show chat (future)
shinectl hide chat        # Hide chat (future)
shinectl reload           # Reload config (future)
shinectl list             # List panels (future)
```

**Implementation**:

- Maps component name to socket path: `/tmp/shine-{component}.sock`
- Creates `RemoteControl` instance
- Sends command via remote control protocol

### 7. Chat Component (cmd/shine-chat/main.go)

**Source**: `docs/examples/bubbletea/chat/main.go`

**Important**: Use the Bubble Tea chat example AS-IS for prototype. No modifications needed - it's a standalone Bubble Tea application that works perfectly in a terminal (Kitty panel).

**Integration**: Zero integration needed. Just build and launch via `kitten panel shine-chat`.

---

## Dependencies

### Go Modules

```go
module github.com/starbased-co/shine

go 1.21

require (
    github.com/charmbracelet/bubbletea v0.25.0
    github.com/charmbracelet/bubbles v0.18.0
    github.com/charmbracelet/lipgloss v0.9.1
    github.com/BurntSushi/toml v1.3.2
)
```

### System Dependencies

- **Kitty** (>= 0.36.0) - Must have `kitten panel` support
- **Hyprland** (any recent version) - Wayland compositor
- **Go** (>= 1.21) - For building

**Verification**:

```bash
# Check Kitty version
kitty --version

# Check if kitten panel exists
kitten panel --help

# Check Hyprland
hyprctl version
```

---

## Build & Installation

### Development Build

```bash
# Initialize module
go mod init github.com/starbased-co/shine
go mod tidy

# Build all binaries
go build -o bin/shine ./cmd/shine
go build -o bin/shinectl ./cmd/shinectl
go build -o bin/shine-chat ./cmd/shine-chat

# Make accessible
export PATH="$PWD/bin:$PATH"
```

### Installation (Future)

```bash
# Install to $HOME/.local/bin
make install

# Or use go install
go install ./cmd/...
```

---

## Configuration

### User Config Location

`~/.config/shine/shine.toml`

### Example Configuration

```toml
[chat]
enabled = true
edge = "bottom"
lines = 10
margin_left = 10
margin_right = 10
margin_bottom = 10
single_instance = true
hide_on_focus_loss = true
focus_policy = "on-demand"
```

### Configuration Reference

| Field                | Type   | Default         | Description                                                |
| -------------------- | ------ | --------------- | ---------------------------------------------------------- |
| `enabled`            | bool   | `false`         | Enable/disable component                                   |
| `edge`               | string | `"bottom"`      | Panel edge: top, bottom, left, right, center, center-sized |
| `lines`              | int    | `1`             | Height in terminal lines                                   |
| `columns`            | int    | `1`             | Width in terminal columns (for left/right edges)           |
| `margin_top`         | int    | `0`             | Top margin in pixels                                       |
| `margin_left`        | int    | `0`             | Left margin in pixels                                      |
| `margin_bottom`      | int    | `0`             | Bottom margin in pixels                                    |
| `margin_right`       | int    | `0`             | Right margin in pixels                                     |
| `single_instance`    | bool   | `false`         | Only allow one instance per component                      |
| `hide_on_focus_loss` | bool   | `false`         | Hide panel when it loses focus                             |
| `focus_policy`       | string | `"not-allowed"` | Focus policy: not-allowed, exclusive, on-demand            |
| `output_name`        | string | `""`            | Target monitor name (empty = primary)                      |

---

## Usage Examples

### Launch All Configured Panels

```bash
shine
```

### Control Panels

```bash
# Toggle chat visibility
shinectl toggle chat

# Future commands
shinectl show chat
shinectl hide chat
shinectl reload
shinectl list
```

### Hyprland Keybindings

```conf
# ~/.config/hypr/hyprland.conf
bind = SUPER, C, exec, shinectl toggle chat
bind = SUPER SHIFT, R, exec, shinectl reload
```

---

## Development Workflow

### Initial Setup

1. Create project structure
2. Initialize Go module
3. Copy chat example to `cmd/shine-chat/main.go`
4. Implement `pkg/panel/config.go` (port from Kitty)
5. Implement `pkg/panel/manager.go`
6. Implement `pkg/panel/remote.go`
7. Implement `pkg/config/` (TOML loading)
8. Implement `cmd/shine/main.go`
9. Implement `cmd/shinectl/main.go`
10. Test end-to-end

### Testing Strategy

**Phase 1 (Prototype)**:

- Manual testing only
- Verify chat launches in panel
- Verify toggle works
- Verify configuration loads

**Phase 2 (Future)**:

- Unit tests for config parsing
- Integration tests for panel manager
- End-to-end tests for component lifecycle

### Debug Commands

```bash
# Test kitten panel directly
kitten panel --edge=bottom --lines=10 \
    --margin-left=10 --margin-right=10 \
    --listen-on=unix:/tmp/test.sock \
    shine-chat

# Test remote control
echo '{"cmd":"resize-os-window","action":"toggle-visibility"}' | \
    nc -U /tmp/test.sock

# Check running processes
ps aux | grep shine
ps aux | grep kitten

# Monitor logs
journalctl -f -t shine
```

---

## Technical Decisions & Rationale

### Why Not Direct Wayland Bindings?

**Considered**: `gotk3-layershell`, `neurlang/wayland`

**Rejected Because**:

- GTK dependency (heavyweight)
- Complex integration with TUI
- Maintenance burden (Wayland protocol changes)
- Not leveraging Kitty's strengths

**Kitty Wins**: Already installed, already works, GPU-accelerated, battle-tested.

### Why Not Build Custom Terminal Renderer?

**Considered**: ANSI → pixel renderer + `neurlang/wayland`

**Rejected Because**:

- Months of development
- Complex: font rendering, OpenGL, input handling
- Unlikely to match Kitty's performance

**Kitty Wins**: Professional-grade terminal rendering for free.

### Why Go Instead of Rust/Python?

**Go Strengths**:

- Fast compile times (rapid iteration)
- Simple concurrency (goroutines for component management)
- Great stdlib (exec, JSON, networking)
- Bubble Tea ecosystem (Charm libraries)
- Single binary distribution

**Rust**: Overkill for this use case, slow compile times
**Python**: Kitty's already Python, Go gives us type safety + speed

### Component as Separate Binaries vs. Single Binary

**Chosen**: Separate binaries (`shine-chat`, `shine-bar`, etc.)

**Rationale**:

- Clean process isolation
- Independent crashes don't kill all components
- Easier to develop/debug individual components
- Users can run components standalone
- Simpler build process (no plugin loading)

**Alternative Rejected**: Monolithic binary with plugin system

- More complex
- Runtime plugin loading issues
- Harder to distribute

---

## Known Limitations

### Prototype (Phase 1)

1. **No hot reload** - Must restart `shine` to reload config
2. **No declarative widgets** - Must write Go code for components
3. **No IPC** - Components can't communicate
4. **Basic error handling** - Errors may not be user-friendly
5. **No logging** - Debug output to stdout only
6. **Single component** - Only chat implemented

### Architecture

1. **Kitty dependency** - Users must have Kitty installed
2. **Terminal constraints** - Limited to terminal rendering (no custom graphics)
3. **Kitty panel API** - Limited by what `kitten panel` supports

**Mitigation**: These are acceptable tradeoffs for the massive simplification gained.

---

## Success Criteria

### Phase 1 (Prototype)

✅ Chat TUI launches in Kitty panel at bottom of screen
✅ Panel has correct margins (10px left/right/bottom)
✅ `shinectl toggle chat` shows/hides panel
✅ Configuration loaded from `~/.config/shine/shine.toml`
✅ Single instance mode works (second launch reuses first)
✅ Hide on focus loss works (panel hides when clicking elsewhere)

### Phase 2 (Expansion)

- Multiple components running simultaneously
- Declarative widget configuration
- Hot reload without restart
- IPC event bus for component communication
- Documentation and examples

---

## Next Steps

### Immediate (Prototype Implementation)

1. ✅ Create project structure
2. ✅ Initialize Go module with dependencies
3. ✅ Copy Bubble Tea chat example
4. ✅ Implement `pkg/panel/config.go`
5. ✅ Implement `pkg/panel/manager.go`
6. ✅ Implement `pkg/panel/remote.go`
7. ✅ Implement `pkg/config/` package
8. ✅ Implement `cmd/shine/main.go`
9. ✅ Implement `cmd/shinectl/main.go`
10. ✅ Test end-to-end
11. ✅ Document build/installation

### Short Term (Post-Prototype)

- Add `shine-bar` component (workspace + clock)
- Implement hot reload
- Add logging with levels
- Error handling improvements
- Write ARCHITECTURE.md

### Long Term (Phase 2)

- Declarative widget system
- Component marketplace
- Theming system
- Advanced IPC
- Performance profiling
- Community components

---

## Reference Documentation

### Research Documents

- `docs/llms/research/git-miner/kitty-wayland-panel.md` - Kitty layer shell deep dive
- `docs/llms/research/github/go-wayland-layershell-libraries.md` - Go Wayland library analysis

### Charm Ecosystem

- `docs/llms/man/charm/bubbletea.md` - Bubble Tea documentation
- `docs/llms/man/charm/bubbles.md` - Bubbles components
- `docs/llms/man/charm/lipgloss.md` - Styling library
- `docs/llms/man/charm/glamour.md` - Markdown rendering

### Example Code

- `docs/examples/bubbletea/chat/main.go` - Chat example (prototype component)
- `docs/examples/bubbletea/chat/README.md` - Chat documentation

### Kitty Source References

- `glfw/wl_window.c:1163-1185` - Layer shell window creation
- `kitty/types.py:72-88` - LayerShellConfig definition
- `kittens/panel/main.py:66` - Config conversion logic
- `kittens/panel/main.py:195-232` - Panel launch flow

---

## Questions & Decisions Log

### Q: Should components be Go plugins or separate binaries?

**A**: Separate binaries for simplicity and isolation.

### Q: How do we handle component discovery?

**A**: Explicit configuration in TOML. No auto-discovery in Phase 1.

### Q: What about users without Kitty?

**A**: Kitty is a hard requirement. Document clearly in README.

### Q: Can components use non-TUI rendering?

**A**: Phase 1: TUI only. Phase 2: Investigate Sixel/image protocol.

### Q: How to handle component crashes?

**A**: Phase 1: Manual restart. Phase 2: Supervisor/auto-restart.

### Q: Declarative widgets in Phase 2 - how?

**A**: TOML config generates Bubble Tea models. Example:

```toml
[bar.widgets.clock]
type = "clock"
format = "15:04:05"
position = "right"
```

Code generation or reflection to instantiate.

---

## Glossary

- **Layer Shell**: Wayland protocol extension for desktop shell components
- **Exclusive Zone**: Reserved screen space that compositor respects
- **Panel**: Desktop shell element (bar, dock, widget)
- **Kitten**: Kitty plugin/extension system
- **TUI**: Text User Interface (terminal-based UI)
- **Bubble Tea**: Elm-inspired TUI framework for Go
- **Charm**: Organization behind Bubble Tea ecosystem

---

## Appendix: Kitty Panel CLI Reference

```bash
kitten panel [OPTIONS] COMMAND

Options:
  --edge=EDGE                Edge placement (top|bottom|left|right|center|center-sized|background|none)
  --layer=LAYER              Wayland layer (background|bottom|top|overlay)
  --lines=NUM[px]            Height in lines or pixels
  --columns=NUM[px]          Width in columns or pixels
  --margin-top=PX            Top margin in pixels
  --margin-left=PX           Left margin in pixels
  --margin-bottom=PX         Bottom margin in pixels
  --margin-right=PX          Right margin in pixels
  --focus-policy=POLICY      Focus policy (not-allowed|exclusive|on-demand)
  --exclusive-zone=NUM       Exclusive zone size (-1 = auto)
  --override-exclusive-zone  Override default exclusive zone
  --hide-on-focus-loss       Auto-hide when focus is lost
  --single-instance          Only one instance per invocation
  --toggle-visibility        Toggle if already running
  --output-name=NAME         Target monitor name
  -o allow_remote_control=socket-only  Enable remote control
  --listen-on=unix:PATH      Remote control socket path
```

---

**Document Version**: 1.0
**Last Updated**: 2025-11-01
**Status**: Ready for Implementation
