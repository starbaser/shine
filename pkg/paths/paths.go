package paths

import (
	"fmt"
	"os"
	"path/filepath"
)

func ExpandHome(path string) string {
	if len(path) == 0 || path[0] != '~' {
		return path
	}
	home, err := os.UserHomeDir()
	if err != nil {
		return path
	}
	if len(path) == 1 {
		return home
	}
	return filepath.Join(home, path[1:])
}

func ConfigDir() string {
	if xdg := os.Getenv("XDG_CONFIG_HOME"); xdg != "" {
		return filepath.Join(xdg, "shine")
	}
	home, _ := os.UserHomeDir()
	return filepath.Join(home, ".config", "shine")
}

func DataDir() string {
	if xdg := os.Getenv("XDG_DATA_HOME"); xdg != "" {
		return filepath.Join(xdg, "shine")
	}
	home, _ := os.UserHomeDir()
	return filepath.Join(home, ".local", "share", "shine")
}

func LogDir() string {
	return filepath.Join(DataDir(), "logs")
}

func RuntimeDir() string {
	uid := os.Getuid()
	return filepath.Join("/run/user", fmt.Sprintf("%d", uid), "shine")
}

func ShinedSocket() string {
	return filepath.Join(RuntimeDir(), "shine.sock")
}

func PrismSocket(instance string) string {
	return filepath.Join(RuntimeDir(), fmt.Sprintf("prism-%s.sock", instance))
}

func PrismState(instance string) string {
	return filepath.Join(RuntimeDir(), fmt.Sprintf("prism-%s.state", instance))
}

func ShinedState() string {
	return filepath.Join(RuntimeDir(), "shined.state")
}

func DefaultConfigPath() string {
	return filepath.Join(ConfigDir(), "shine.toml")
}
