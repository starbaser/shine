# SHINE Architecture

## Diagram Conventions

| Arrow | Meaning |
|-------|---------|
| `-->` | Compile-time import |
| `-.->` | Runtime IPC |

## System Overview

Commands flow **down**, notifications flow **up**.

```mermaid
flowchart TB
    subgraph CONTROL["Control Plane"]
        shine["shine<br/>CLI"]
        shined["shined<br/>daemon"]
    end

    subgraph PANEL["Panel (per instance)"]
        prismctl["prismctl<br/>PTY supervisor"]
        prisms["TUIs: bar · clock · sysinfo · chat"]
    end

    subgraph PKG["pkg/"]
        direction LR
        rpc[rpc] --- state[state] --- paths[paths]
        config[config] --- panel[panel] --- help[help]
    end

    shine -.->|"shine.sock"| shined
    shined -.->|"kitten @ launch"| prismctl
    prismctl -.->|"PTY"| prisms
    prismctl -.->|"notify"| shined

    CONTROL --> PKG
    PANEL --> PKG
```

## Runtime Communication

Detailed IPC flow showing socket-level interactions.

```mermaid
flowchart LR
    subgraph CLI
        shine_cmd[commands.go]
    end

    subgraph DAEMON
        shined_ipc[ipc_server.go]
        shined_mgr[panel_manager.go]
        shined_notif[notifications.go]
    end

    subgraph SUPERVISOR
        prism_ipc[ipc.go]
        prism_super[supervisor.go]
        prism_notif[notifications.go]
    end

    subgraph APPS
        bar[bar]
        clock[clock]
    end

    shine_cmd -.->|"shine.sock<br/>panel/spawn"| shined_ipc
    shined_ipc --> shined_mgr
    shined_mgr -.->|"kitten @ launch<br/>--type=os-panel"| prism_ipc
    shined_ipc -.->|"prism-*.sock<br/>prism/up"| prism_ipc
    prism_ipc --> prism_super
    prism_super -.->|"execve"| bar & clock
    prism_notif -.->|"prism/started"| shined_notif
```

## Binary Structure

Each binary's internal file organization.

### cmd/shine/ (CLI)

```mermaid
flowchart LR
    main[main.go] --> commands[commands.go]
    commands --> output[output.go]
    commands --> help[help.go]
```

### cmd/shined/ (Daemon)

```mermaid
flowchart TB
    main[main.go] --> config[config.go]
    main --> ipc[ipc_server.go]
    ipc --> handlers[panel_handlers.go]
    handlers --> mgr[panel_manager.go]
    mgr --> state[state.go]
    mgr --> notif[notifications.go]
    mgr --> newprism[newprism.go]
```

### cmd/prismctl/ (PTY Supervisor)

```mermaid
flowchart TB
    main[main.go] --> super[supervisor.go]
    main --> ipc[ipc.go]
    ipc --> handlers[handlers.go]
    super --> mirror[mirror.go]
    super --> pty[pty_manager.go]
    super --> term[terminal.go]
    super --> signals[signals.go]
    super --> state[state.go]
    super --> notif[notifications.go]
```

## Package Dependencies

How `pkg/` modules relate to each other and to binaries.

```mermaid
flowchart TB
    subgraph BINS["Binaries"]
        shine[shine]
        shined[shined]
        prismctl[prismctl]
    end

    subgraph CORE["Core Infrastructure"]
        rpc[pkg/rpc]
        state[pkg/state]
        paths[pkg/paths]
    end

    subgraph DOMAIN["Domain Logic"]
        config[pkg/config]
        panel[pkg/panel]
    end

    subgraph UI["UI Support"]
        help[pkg/help]
    end

    %% Binary deps
    shine --> rpc & state & paths & help
    shined --> rpc & state & paths & config & panel
    prismctl --> rpc & state & paths

    %% Internal pkg deps
    config --> panel
    config --> paths

    %% Styling
    classDef core fill:#3182ce,color:#fff
    classDef domain fill:#38a169,color:#fff
    classDef ui fill:#d69e2e,color:#fff
    class rpc,state,paths core
    class config,panel domain
    class help ui
```

## File Inventory

| Directory | Files | Role |
|-----------|------:|------|
| `cmd/shine/` | 4 | User CLI |
| `cmd/shined/` | 9 | Service daemon |
| `cmd/prismctl/` | 12 | PTY supervisor |
| `cmd/prisms/` | 4 | Example TUIs |
| `pkg/config/` | 5 | Config & discovery |
| `pkg/panel/` | 2 | Kitty integration |
| `pkg/rpc/` | 4 | JSON-RPC 2.0 |
| `pkg/state/` | 4 | Mmap state files |
| `pkg/paths/` | 1 | Path utilities |
| `pkg/help/` | 2 | Help rendering |
| **Total** | **37** | |
