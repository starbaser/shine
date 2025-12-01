package theme

import (
	"fmt"
	"sync"

	"github.com/charmbracelet/lipgloss"
)

// Theme defines a complete color scheme for TUI applications.
// Implementations should provide cohesive color palettes that work well together.
type Theme interface {
	Name() string

	// Base colors
	Background() lipgloss.TerminalColor
	Foreground() lipgloss.TerminalColor

	// Semantic colors
	Primary() lipgloss.TerminalColor
	Secondary() lipgloss.TerminalColor
	Accent() lipgloss.TerminalColor
	Muted() lipgloss.TerminalColor

	// Status colors
	Success() lipgloss.TerminalColor
	Warning() lipgloss.TerminalColor
	Error() lipgloss.TerminalColor
	Info() lipgloss.TerminalColor

	// Surface colors (layered backgrounds)
	Surface0() lipgloss.TerminalColor // Base surface
	Surface1() lipgloss.TerminalColor // Raised surface
	Surface2() lipgloss.TerminalColor // Highest surface

	// Border colors
	BorderSubtle() lipgloss.TerminalColor
	BorderActive() lipgloss.TerminalColor

	// Text hierarchy
	TextPrimary() lipgloss.TerminalColor
	TextSecondary() lipgloss.TerminalColor
	TextMuted() lipgloss.TerminalColor
}

// Registry manages available themes and the current active theme.
type Registry struct {
	mu      sync.RWMutex
	themes  map[string]Theme
	current Theme
}

var (
	globalRegistry = &Registry{
		themes: make(map[string]Theme),
	}
)

// Register adds a theme to the global registry.
// If this is the first theme registered, it becomes the current theme.
func Register(t Theme) {
	globalRegistry.mu.Lock()
	defer globalRegistry.mu.Unlock()

	name := t.Name()
	globalRegistry.themes[name] = t

	if globalRegistry.current == nil {
		globalRegistry.current = t
	}
}

// SetCurrent sets the active theme by name.
// Returns an error if the theme is not registered.
func SetCurrent(name string) error {
	globalRegistry.mu.Lock()
	defer globalRegistry.mu.Unlock()

	t, ok := globalRegistry.themes[name]
	if !ok {
		return fmt.Errorf("theme not found: %s", name)
	}

	globalRegistry.current = t
	return nil
}

// Current returns the currently active theme.
// If no theme is set, it panics. Always register at least one theme.
func Current() Theme {
	globalRegistry.mu.RLock()
	defer globalRegistry.mu.RUnlock()

	if globalRegistry.current == nil {
		panic("no theme registered")
	}

	return globalRegistry.current
}

// List returns all registered theme names.
func List() []string {
	globalRegistry.mu.RLock()
	defer globalRegistry.mu.RUnlock()

	names := make([]string, 0, len(globalRegistry.themes))
	for name := range globalRegistry.themes {
		names = append(names, name)
	}

	return names
}

// Get retrieves a theme by name without setting it as current.
// Returns nil if the theme is not registered.
func Get(name string) Theme {
	globalRegistry.mu.RLock()
	defer globalRegistry.mu.RUnlock()

	return globalRegistry.themes[name]
}
