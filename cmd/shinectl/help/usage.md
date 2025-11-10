# shinectl - Shine Service Manager

Background service that spawns and manages prismctl panel supervisors.

## USAGE

```bash
shinectl [options]
```

## OPTIONS

```text
-config PATH    Path to prism.toml (default: ~/.config/shine/prism.toml)
-version        Print version and exit
-help           Show this help message
```

## BEHAVIOR

shinectl is a long-running daemon that:
- Reads configuration from prism.toml
- Spawns Kitty panels via remote control API
- Launches prismctl supervisors for each panel
- Monitors panel health (30-second interval)
- Handles configuration reloads via SIGHUP

## SIGNALS

```text
SIGHUP          Reload configuration and update panels
SIGTERM/SIGINT  Graceful shutdown of all panels
```

## EXAMPLES

```bash
$ shinectl
```

```bash
$ shinectl -config ~/.config/shine/my-config.toml
```

```bash
$ pkill -HUP shinectl
```

## FILES

```text
Config:  ~/.config/shine/prism.toml
Logs:    ~/.local/share/shine/logs/shinectl.log
Sockets: /run/user/{uid}/shine/prism-*.sock
```

## LEARN MORE
  Use `shine start` to launch shinectl as a service.
  Use `shine logs shinectl` to view logs.
  See prism.toml documentation for configuration options.
