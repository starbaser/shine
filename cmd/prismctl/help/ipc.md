# prismctl ipc

IPC protocol and command reference for hot-swapping prisms.

## USAGE

```bash
echo '{"action":"<action>"}' | socat - UNIX-CONNECT:<socket-path>
```

## IPC COMMANDS

### start

Start or resume a prism (idempotent).

```json
{"action":"start","prism":"shine-clock"}
```

Behavior:
- If prism is already running, brings it to foreground
- If different prism is running, suspends it and starts new one
- If no prism is running, starts the specified prism

### kill

Kill the current prism and resume next in MRU list.

```json
{"action":"kill","prism":"shine-clock"}
```

Behavior:
- Terminates the specified prism process
- Automatically resumes most recently used prism
- Removes killed prism from MRU list

### status

Query current supervisor status.

```json
{"action":"status"}
```

Response:
```json
{
  "success": true,
  "foreground": "shine-clock",
  "background": ["shine-chat", "shine-spotify"]
}
```

### stop

Graceful shutdown of prismctl supervisor.

```json
{"action":"stop"}
```

Behavior:
- Terminates foreground prism
- Terminates all background prisms
- Closes IPC socket
- Exits prismctl process

## EXAMPLES

```bash
$ SOCK=$(ls -t /run/user/$(id -u)/shine/prism-*.sock | head -1)
$ echo '{"action":"status"}' | socat - UNIX-CONNECT:$SOCK
```

```bash
$ SOCK=$(ls -t /run/user/$(id -u)/shine/prism-panel-0.*.sock | head -1)
$ echo '{"action":"start","prism":"shine-clock"}' | socat - UNIX-CONNECT:$SOCK
```

```bash
$ SOCK=$(ls -t /run/user/$(id -u)/shine/prism-*.sock | head -1)
$ echo '{"action":"kill","prism":"shine-chat"}' | socat - UNIX-CONNECT:$SOCK
```

## SOCKET DISCOVERY

Always find the current socket before sending commands:

```bash
# Find most recent socket
SOCK=$(ls -t /run/user/$(id -u)/shine/prism-*.sock | head -1)

# Or find by component name
SOCK=$(ls -t /run/user/$(id -u)/shine/prism-test-prism.*.sock | head -1)
```

Each prismctl instance creates a socket with its PID in the name.
When prismctl restarts, the old socket is removed and a new one is created.

## LEARN MORE
  Use `prismctl help usage` for main usage.
  Use `prismctl help signals` for signal handling.
  See IPC protocol documentation for advanced usage.
