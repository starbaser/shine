# shine Help System

Comprehensive documentation for the shine CLI help system.

## Overview

The shine help system provides three key features:

1. **Human-readable help** - Beautifully rendered markdown via Glamour
2. **Structured metadata** - Organized command information with categories
3. **Machine-readable output** - JSON format for tooling integration

## Architecture

### Hybrid Approach

The help system uses a hybrid architecture:

- **Markdown files** (`cmd/shine/help/*.md`) - Long-form content, examples, troubleshooting
- **Go structs** (`help_metadata.go`) - Structured metadata for organization and tooling
- **Dynamic generation** - Help listings generated from metadata at runtime

### Components

```
cmd/shine/
├── help/
│   ├── usage.md      # Main help page
│   ├── start.md      # Per-command help pages
│   ├── stop.md
│   ├── status.md
│   ├── reload.md
│   └── logs.md
├── help.go           # Rendering and display logic
├── help_metadata.go  # Structured metadata and registry
└── main.go           # CLI integration
```

## Usage

### Human-Readable Help

```bash
# Main help
shine --help
shine -h

# Per-command help (full markdown)
shine help start
shine help stop
shine help status
shine help reload
shine help logs

# Command listings
shine help list       # Detailed list with descriptions
shine help topics     # Quick topic overview
shine help categories # Commands organized by category
```

### Machine-Readable Help (JSON)

```bash
# Get metadata for a specific command
shine help start --json

# Get all command names (for completion)
shine help --json names

# Get all commands with metadata
shine help --json

# Get category structure
shine help --json categories
```

## Integration Examples

### Shell Completion (zsh)

Create `~/.zfunc/_shine`:

```zsh
#compdef shine

_shine() {
  local -a commands
  commands=(${(f)"$(shine help --json names 2>/dev/null)"})

  _arguments \
    '1: :->command' \
    '*::arg:->args'

  case $state in
    command)
      _describe 'command' commands
      ;;
  esac
}

_shine
```

### Shell Completion (bash)

Create `/etc/bash_completion.d/shine`:

```bash
_shine_completions() {
  local cur prev
  COMPREPLY=()
  cur="${COMP_WORDS[COMP_CWORD]}"
  prev="${COMP_WORDS[COMP_CWORD-1]}"

  if [ $COMP_CWORD -eq 1 ]; then
    COMPREPLY=($(compgen -W "$(shine help --json names 2>/dev/null)" -- "$cur"))
  fi
}

complete -F _shine_completions shine
```

### IDE Integration

Get command synopsis for hover text:

```javascript
const { execSync } = require('child_process');

function getCommandHelp(command) {
  try {
    const json = execSync(`shine help ${command} --json`, { encoding: 'utf8' });
    return JSON.parse(json);
  } catch (err) {
    return null;
  }
}

// Usage
const help = getCommandHelp('start');
console.log(help.synopsis);  // "Start the shine service and enabled panels"
console.log(help.related);   // ["stop", "status", "reload"]
```

### Man Page Generation

The structured metadata enables future man page generation:

```bash
# Planned future feature
shine help start --man > /usr/local/share/man/man1/shine-start.1
```

## CommandHelp Structure

Each command in the registry has the following metadata:

```go
type CommandHelp struct {
    Name        string   // Command name
    Category    string   // Command category for grouping
    Synopsis    string   // Brief one-line description
    Description string   // Longer description (optional)
    Usage       string   // Usage syntax
    Content     string   // Full help content (markdown)
    Related     []string // Related commands
    SeeAlso     []string // Additional topics/resources
}
```

### Categories

Commands are organized into three categories:

1. **Service Management** - Starting, stopping, and managing the shine service
2. **Monitoring** - Viewing status, logs, and debugging
3. **Configuration** - Managing configuration and settings

## Adding New Commands

To add help for a new command:

1. Create markdown file: `cmd/shine/help/mycommand.md`
2. Add embed directive in `help.go`:
   ```go
   //go:embed help/mycommand.md
   var mycommandHelp string
   ```
3. Add entry to `helpRegistry` in `help_metadata.go`:
   ```go
   "mycommand": {
       Name:     "mycommand",
       Category: "Service Management",
       Synopsis: "Brief description",
       Usage:    "shine mycommand [args]",
       Content:  mycommandHelp,
       Related:  []string{"relatedcmd"},
       SeeAlso:  []string{"Additional info"},
   }
   ```
4. Rebuild: `go build -o bin/shine ./cmd/shine`

## Benefits

### Maintainability

- **Separation of concerns** - Content in markdown, metadata in Go
- **Single source of truth** - Each command documented once
- **Easy updates** - Edit markdown without touching code structure

### Extensibility

- **Multiple output formats** - Markdown, JSON, (future: man pages, HTML)
- **Tooling integration** - Shell completion, IDE support, documentation generators
- **Programmatic access** - Query help system from other tools

### User Experience

- **Beautiful rendering** - Glamour provides terminal-optimized display
- **Organized navigation** - Categories and listings help discovery
- **Rich content** - Examples, troubleshooting, cross-references

## Future Enhancements

Potential improvements to the help system:

1. **Man page generation** - Convert metadata + markdown to groff format
2. **HTML documentation** - Static site generation from help content
3. **Search functionality** - `shine help search <term>`
4. **Interactive help** - TUI browser for help content
5. **Localization** - Multi-language support via metadata
6. **Help versioning** - Version-specific help content
7. **Extended metadata** - Tags, aliases, deprecation notices

## Implementation Notes

### Why Hybrid?

Pure structured approach (all help in Go structs) would require:
- Embedding long markdown in string literals
- Losing syntax highlighting for embedded examples
- Harder to edit and maintain content

Pure markdown approach lacks:
- Programmatic access to metadata
- Ability to generate listings dynamically
- Machine-readable output for tooling

The hybrid approach combines the best of both:
- Maintainable markdown files for content
- Structured metadata for organization
- Dynamic generation for listings and JSON output

### Performance

Help rendering is fast because:
- Markdown files embedded at compile time (no file I/O)
- Glamour renderer initialized once via `init()`
- JSON encoding uses standard library

### Compatibility

The help system maintains backward compatibility:
- `--help` flag works as before
- Per-command help unchanged for users
- New features are purely additive
