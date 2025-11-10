# prismctl signals

Signal handling for process management and terminal control.

## SIGNAL HANDLING

prismctl handles the following signals:

### SIGCHLD - Child Process Exit

When a prism process exits:
1. Supervisor detects exit via SIGCHLD
2. Logs exit status
3. Applies restart policy if configured
4. Updates MRU list if crash recovery enabled

### SIGTERM/SIGINT - Graceful Shutdown

When prismctl receives SIGTERM or SIGINT:
1. Terminates foreground prism (SIGTERM → SIGKILL)
2. Terminates all background prisms
3. Restores terminal state
4. Closes IPC socket
5. Exits cleanly

### SIGWINCH - Terminal Resize

When terminal is resized:
1. prismctl forwards SIGWINCH to foreground prism
2. Prism's Bubble Tea program handles resize internally
3. Background prisms remain unaffected

## PROCESS MANAGEMENT

prismctl uses signals to manage prism processes:

### Suspend (Background)

```text
SIGSTOP → Suspend process execution
```

When switching prisms:
1. Save terminal state
2. Send SIGSTOP to suspend process
3. Process enters background MRU list

### Resume (Foreground)

```text
SIGCONT → Resume process execution
```

When bringing prism to foreground:
1. Restore terminal state
2. Send SIGCONT to resume process
3. Wait 10ms for stabilization

### Terminate

```text
SIGTERM → Graceful shutdown request
SIGKILL → Forced termination (after 20ms)
```

Shutdown sequence:
1. Send SIGTERM to prism
2. Wait 20ms for graceful exit
3. Send SIGKILL if still running

## TIMING CONSTANTS

**CRITICAL**: Do not modify these without testing:
- 10ms stabilization delay after SIGCONT
- 20ms shutdown grace period (SIGTERM → SIGKILL)

Terminal state restoration requires exact sequencing.

## EXAMPLES

```bash
$ prismctl shine-clock
```

```bash
$ pkill -TERM prismctl
```

```bash
$ kill -WINCH $(pgrep prismctl)
```

## LEARN MORE
  Use `prismctl help usage` for main usage.
  Use `prismctl help ipc` for IPC commands.
  See terminal state management documentation.
