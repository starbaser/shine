# shine - Prism TUI Manager

A TUI-based desktop shell toolkit for Hyprland using Kitty panels.

## Usage

```bash
shine <command> [arguments]
```

## Commands

- `start` - Start/resume the shine service
- `stop` - Gracefully stop all panels
- `reload` - Reload configuration and update panels
- `status` - Show status of all panels
- `logs [panel-id]` - View logs (all or specific panel)
- `help [topic]` - Show detailed help for a command or topic

## Options

- `-h, --help` - Show this help message
- `-v, --version` - Show version information

## Examples

Start the service:
```bash
shine start
```

View panel status:
```bash
shine status
```

List all log files:
```bash
shine logs
```

View shinectl logs:
```bash
shine logs shinectl
```

Stop all panels:
```bash
shine stop
```

Get detailed help for a command:
```bash
shine help start
shine help status
```

List all available commands:
```bash
shine help list       # Detailed list with descriptions
shine help topics     # Quick topic overview
shine help categories # Commands organized by category
```

## Configuration

- **Config file**: `~/.config/shine/prism.toml`
- **Log files**: `~/.local/share/shine/logs/`

## Getting Started

1. Configure panels in `~/.config/shine/prism.toml`
2. Start the service: `shine start`
3. Check status: `shine status`

## Documentation

For more information, see the documentation at https://github.com/starbased-co/shine
