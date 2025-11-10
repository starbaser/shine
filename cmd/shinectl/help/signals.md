# shinectl signals

Signal handling for configuration reload and shutdown.

## USAGE

```bash
pkill -HUP shinectl    # Reload configuration
pkill -TERM shinectl   # Graceful shutdown
```

## SIGHUP - Configuration Reload

When shinectl receives SIGHUP, it:
1. Reloads prism.toml configuration
2. Validates the new configuration
3. Compares current panels with new config
4. Removes panels no longer in config
5. Adds panels for new prisms

Existing panels are NOT restarted during reload.

## SIGTERM/SIGINT - Graceful Shutdown

When shinectl receives SIGTERM or SIGINT:
1. Logs shutdown message
2. Calls PanelManager.Shutdown()
3. Terminates all prismctl supervisors
4. Closes Kitty panels
5. Exits cleanly

## EXAMPLES

```bash
$ pkill -HUP shinectl
```

```bash
$ pkill -TERM shinectl
```

```bash
$ kill -HUP $(pgrep shinectl)
```

## LEARN MORE
  Use `shine reload` to send SIGHUP automatically.
  Use `shine stop` to send SIGTERM automatically.
  See `shinectl help` for main usage.
