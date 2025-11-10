# shinectl config

Configuration file format and loading behavior.

## USAGE

```bash
shinectl -config ~/.config/shine/prism.toml
```

## DEFAULT LOCATION

```text
~/.config/shine/prism.toml
```

## CONFIGURATION FORMAT

Example prism.toml:

```toml
[core]
log_level = "info"

[[prisms]]
name = "shine-clock"
enabled = true
origin = "top-right"
width = "200px"
height = "100px"

[[prisms]]
name = "shine-chat"
enabled = true
origin = "bottom-left"
width = "400px"
height = "300px"
```

## VALIDATION

shinectl validates configuration on startup and reload:
- Prism name must not be empty
- Origin must be valid (top-left, top-right, bottom-left, bottom-right)
- Dimensions must be valid (pixels or percentages)

Invalid configurations cause shinectl to exit or abort reload.

## HOT-RELOAD

```bash
$ pkill -HUP shinectl
```

Configuration reload does NOT restart existing panels.
Only adds/removes panels based on config changes.

## EXAMPLES

```bash
$ shinectl -config ~/.config/shine/dev-config.toml
```

```bash
$ vim ~/.config/shine/prism.toml
$ pkill -HUP shinectl
```

## LEARN MORE
  Use `shinectl help signals` for reload behavior.
  See prism.toml documentation for full reference.
  Use `shine status` to verify configuration.
