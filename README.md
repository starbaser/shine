# `shine ⬘`

## Overview

```
                      ┌─────────────────┐
                      │      shine      │ <- $ shine start/stop/status
                      └────────┬────────┘
                              ↓│↑ JSON-RPC 2.0 ⇆
┌──────────────────────────────┴──────────────────────┐
│  shinectl                                           │
│  ├─ Loads ~/.config/shine/shine.toml                │
│  │  └─ Discovers prism.toml defs in prisms/         │
│  ├─ Manages prism lifecycle                         │
│  └─ Launches prisms via Kitty remote control API    │
└──────────────────────┬──────────────────────────────┘
    unix:@shine.sock ⇆ │ ⇆ unix:@mykitty
                       │ kitten @ launch --type=os-panel prismctl {instance}
              ┌────────┴────────┐
              │      kitty      │ <- wl_surface
              └────────┬────────┘
                      ↓│↑ stdio
              ┌────────┴────────┐
 /dev/ptmx -> │      pty_M      │
              └────────┬────────┘
                      ↓│↑
              ┌────────┴────────┐     ┌──────────────┐
/dev/pts/n -> │      pty_S      ├─────┤   prismctl   │ <- prism-{instance}.sock
              └─────────────────┘     └─┬────┬─────┬─┘
                                 ┌──────┘    │     └─────┐
                            ┌────┴─────┐┌────┴─────┐┌────┴─────┐
                            │   PTY1   ││   PTY2   ││  *PTY3   │ <- (* = foreground)
                            └────┬─────┘└────┬─────┘└────┬─────┘
                            ┌────┴─────┐┌────┴─────┐┌────┴─────┐
                            │  clock   ││  wabar   ││   app3   │ <- prisms/bin/
                            └──────────┘└──────────┘└──────────┘
```
