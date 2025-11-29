# Shine IPC Architecture: JSON-RPC + Two-Tier mmap State

## Status: 100% Complete ✅

**Implementation complete.** All IPC infrastructure is operational.

---

## Architecture Overview

Shine is a **desktop shell platform** for running arbitrary TUI applications (prisms) in Kitty layer shell panels.

```
┌─────────────────────────────────────────────────────────────────────┐
│                           User                                       │
│                             │                                        │
│                      shine CLI                                       │
│                             │                                        │
└─────────────────────────────┼───────────────────────────────────────┘
                              │ JSON-RPC
                              ▼
┌─────────────────────────────────────────────────────────────────────┐
│                        shinectl                                      │
│              (Service manager, provisions panels)                    │
│                             │                                        │
│              ┌──────────────┼──────────────┐                        │
│              ▼              ▼              ▼                        │
│         kitten panel   kitten panel   kitten panel                  │
└─────────────────────────────────────────────────────────────────────┘
                              │
                              ▼
┌─────────────────────────────────────────────────────────────────────┐
│                        prismctl                                      │
│              (PTY supervisor per panel)                              │
│                             │                                        │
│              ┌──────────────┼──────────────┐                        │
│              ▼              ▼              ▼                        │
│           prism A       prism B       prism C                       │
│          (weather)     (spotify)      (clock)                       │
└─────────────────────────────────────────────────────────────────────┘
```

---

## Dual-Layer IPC Design

| Layer | Transport | Purpose | Direction |
|-------|-----------|---------|-----------|
| **State** | Binary mmap | Read-only reflection of runtime state | Write: process, Read: anyone |
| **Command** | JSON-RPC 2.0 socket | Actions, commands, queries | Bidirectional RPC |

**Why mmap?** Zero-copy reads, transparent/debuggable (`hexdump`), multiple readers without overhead.

**Why JSON-RPC?** Battle-tested (LSP), request/response correlation, notification support, standard error codes.

---

## What's Implemented ✅

### pkg/rpc/ (Complete)
- `types.go` - Request/response types for all methods
- `server.go` - Multi-connection RPC server via `creachadair/jrpc2`
- `client.go` - Typed clients (PrismClient, ShinectlClient)
- `errors.go` - Domain-specific RPC errors
- `rpc_test.go` - 414 lines of tests

### pkg/state/ (Complete)
- `types.go` - Binary state structs with sequence counting
- `writer.go` - Atomic mmap writes
- `reader.go` - Lock-free consistent reads
- `mmap.go` - Memory-mapped file abstraction
- `state_test.go` - 464 lines of tests

### cmd/prismctl/ (Complete)
- `ipc.go` - RPC server lifecycle
- `handlers.go` - All `prism/*` and `service/*` handlers
- `state.go` - StateManager wrapper
- `notifications.go` - Bidirectional notifications with reconnect

### cmd/shinectl/ (Complete)
- `ipc_server.go` - RPC server lifecycle
- `panel_handlers.go` - Panel RPC handlers (all implemented)
- `notifications.go` - Notification handlers with restart policy execution
- `state.go` - StateManager wrapper
- `panel_manager.go` - PID tracking, restart policy, RPC client integration

---

## RPC Methods

### prismctl (prism-{instance}.sock)

| Method | Params | Result | Status |
|--------|--------|--------|--------|
| `prism/up` | `{name}` | `{pid, state}` | ✅ |
| `prism/down` | `{name}` | `{stopped}` | ✅ |
| `prism/fg` | `{name}` | `{ok, was_fg}` | ✅ |
| `prism/bg` | `{name}` | `{ok, was_bg}` | ✅ |
| `prism/list` | `{}` | `{prisms: [...]}` | ✅ |
| `service/health` | `{}` | `{healthy, count}` | ✅ |
| `service/shutdown` | `{graceful}` | `{shutting_down}` | ✅ |

### shinectl (shinectl.sock)

| Method | Params | Result | Status |
|--------|--------|--------|--------|
| `panel/list` | `{}` | `{panels: [...]}` | ✅ |
| `panel/spawn` | `{config}` | `{instance, socket}` | ✅ |
| `panel/kill` | `{instance}` | `{killed}` | ✅ |
| `service/status` | `{}` | `{panels, uptime}` | ✅ |
| `config/reload` | `{}` | `{reloaded}` | ✅ |

### Notifications (prismctl → shinectl)

| Notification | Params | Status |
|--------------|--------|--------|
| `prism/started` | `{panel, name, pid}` | ✅ |
| `prism/stopped` | `{panel, name, exit_code}` | ✅ |
| `prism/crashed` | `{panel, name, exit_code, signal}` | ✅ |
| `surface/switched` | `{panel, from, to}` | ✅ |

---

## Completed Work ✅

### 1. `panel/spawn` Implementation ✅
**File:** `cmd/shinectl/panel_handlers.go`

- Parses `req.Config` (map[string]any) into PrismEntry struct via JSON round-trip
- Validates required fields (name) and restart policy configuration
- Calls `h.pm.SpawnPanel(entry, instanceName)` with proper error handling
- Updates state via `h.state.OnPanelSpawned()`
- Returns `PanelSpawnResult` with instance name and socket path

### 2. PID Tracking for Panels ✅
**File:** `cmd/shinectl/panel_manager.go`

- Added `getPIDFromWindowID()` helper using `kitten @ ls` to retrieve PIDs
- Stores PID in Panel struct after spawning
- Returns actual PID in `panel/list` and `service/status` responses
- Graceful degradation: PID=0 if retrieval fails

### 3. Restart Policy Execution ✅
**File:** `cmd/shinectl/panel_manager.go`, `cmd/shinectl/notifications.go`

- Implemented `PrismRestartState` tracking (count, timestamps, stopped flag)
- Implemented `TriggerRestartPolicy(panel, prism, exitCode)` method
- Supports all policies: no, on-failure, unless-stopped, always
- Respects `restart_delay` and `max_restarts` (rate limiting per hour)
- Async restart via goroutine with RPC `prism/up` call

### 4. Integration Tests ✅
**Files:** `cmd/shinectl/ipc_integration_test.go`, `cmd/shinectl/notifications_test.go`,
         `cmd/prismctl/ipc_integration_test.go`, `cmd/prismctl/notifications_test.go`

Added 1,795 lines of integration tests:
- End-to-end IPC flow tests (spawn → list → kill)
- Notification delivery verification
- State persistence during operations
- Concurrent client access tests
- Error handling coverage

### 5. shine CLI Update ✅
**File:** `cmd/shine/commands.go`

- Uses `rpc.NewShinectlClient()` for all shinectl communication
- Uses `rpc.NewPrismClient()` for direct prismctl communication
- Implements fast status with `state.OpenPrismStateReader()` (instant mmap read)
- Falls back to RPC if state file unavailable
- All commands (start, stop, reload, status) fully migrated

### 6. Deprecated Files Removed ✅
- Deleted `pkg/ipc/protocol.go`
- Deleted `pkg/ipc/client.go`
- Migrated `panel_manager.go` from `pkg/ipc` to `pkg/rpc`

---

## Two-Tier State Model

### Tier 1: Supervisor State (Fixed Schema)
**File:** `/run/user/{uid}/shine/prism-{instance}.state`
**Owner:** prismctl

```go
type PrismRuntimeState struct {
    Version      uint64     // Sequence counter (odd=writing, even=stable)
    InstanceLen  uint8
    Instance     [63]byte
    ActivePrism  uint8      // Index of foreground prism (255 = none)
    PrismCount   uint8
    _padding     [6]byte
    Prisms       [16]PrismEntry
}
```

### Tier 2: Prism App State (Schema-Defined)
**File:** `/run/user/{uid}/shine/panel-{panel}-{prism}.state`
**Owner:** Each prism (optional)
**Schema:** Defined by prism author in `state.schema.json`

*(Future work - prism authors can create their own state files)*

---

## Socket Paths

```
/run/user/{uid}/shine/
├── shinectl.sock           # shinectl RPC
├── shinectl.state          # shinectl mmap state
├── prism-{instance}.sock   # prismctl RPC (per panel)
└── prism-{instance}.state  # prismctl mmap state (per panel)
```

---

## Verification Commands

```bash
# Test prismctl RPC
echo '{"jsonrpc":"2.0","id":1,"method":"prism/list","params":{}}' | \
  socat - UNIX-CONNECT:/run/user/$(id -u)/shine/prism-test.sock

# Test shinectl RPC
echo '{"jsonrpc":"2.0","id":1,"method":"panel/list","params":{}}' | \
  socat - UNIX-CONNECT:/run/user/$(id -u)/shine/shinectl.sock

# Inspect mmap state
hexdump -C /run/user/$(id -u)/shine/prism-test.state

# Run tests
go test ./pkg/rpc/... ./pkg/state/...
```

---

## Dependencies

- `github.com/creachadair/jrpc2 v1.3.3` - JSON-RPC 2.0 library (already in go.mod)

---

## Files Summary

**Core Infrastructure:**
- `pkg/rpc/{types,server,client,errors}.go` - JSON-RPC 2.0 implementation
- `pkg/state/{types,writer,reader,mmap}.go` - Memory-mapped state files

**prismctl (PTY supervisor):**
- `cmd/prismctl/{ipc,handlers,state,notifications}.go`
- `cmd/prismctl/{ipc_integration_test,notifications_test}.go` - Integration tests

**shinectl (service manager):**
- `cmd/shinectl/{ipc_server,panel_handlers,notifications,state,panel_manager}.go`
- `cmd/shinectl/{ipc_integration_test,notifications_test}.go` - Integration tests

**shine CLI:**
- `cmd/shine/commands.go` - Uses new RPC client with mmap fast-path

**Removed (deprecated):**
- ~~`pkg/ipc/protocol.go`~~ - Deleted
- ~~`pkg/ipc/client.go`~~ - Deleted
