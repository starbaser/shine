// Package theme provides a generic, reusable theming system for Bubble Tea TUI applications.
//
// The theme package offers:
//   - Generic Theme interface not tied to any specific colorscheme
//   - Theme registry for runtime theme switching
//   - Pre-built Lipgloss style builders for common UI patterns
//   - Nerd Font icon constants for system widgets
//   - Custom border definitions for panels and containers
//
// # Basic Usage
//
// The package auto-registers the Srcery theme on import. Use theme.Current() to access
// the active theme:
//
//	import "github.com/starbased-co/shine/cmd/prisms/internal/theme"
//
//	func main() {
//	    t := theme.Current()
//	    style := lipgloss.NewStyle().
//	        Foreground(t.Primary()).
//	        Background(t.Surface0())
//	}
//
// # Pre-Built Styles
//
// Use style builders for common UI patterns:
//
//	panel := theme.PanelStyle().Width(40).Render("Content")
//	card := theme.CardStyle().Render("Card content")
//	title := theme.TitleStyle().Render("Title")
//	success := theme.SuccessStyle().Render("✓ Success")
//
// # Icons
//
// Access Nerd Font icons for system widgets:
//
//	cpu := theme.IconCPU
//	battery := theme.GetBatteryIcon(75, false)  // 75%, not charging
//	volume := theme.GetVolumeIcon(80, false)    // 80%, not muted
//
// # Custom Borders
//
//	glass := lipgloss.NewStyle().Border(theme.GlassBorder())
//	thick := lipgloss.NewStyle().Border(theme.ThickBorder())
//
// # Theme Interface
//
// Implement the Theme interface to add new themes:
//
//	type MyTheme struct{}
//
//	func (t MyTheme) Name() string { return "mytheme" }
//	func (t MyTheme) Background() lipgloss.TerminalColor { ... }
//	// ... implement all required methods
//
//	func init() {
//	    theme.Register(MyTheme{})
//	}
//
// # Color Semantics
//
// The Theme interface provides colors organized by purpose:
//
//   - Base: Background, Foreground
//   - Semantic: Primary, Secondary, Accent, Muted
//   - Status: Success, Warning, Error, Info
//   - Surface: Surface0, Surface1, Surface2 (layered backgrounds)
//   - Border: BorderSubtle, BorderActive
//   - Text: TextPrimary, TextSecondary, TextMuted
//
// This organization ensures themes remain consistent and interchangeable.
//
// # Requirements
//
//   - Nerd Font (recommended: JetBrainsMono Nerd Font)
//   - Terminal with true color support (24-bit)
//   - Charm Bracelet Lipgloss library
//
// # Example
//
// See examples/demo/main.go for a comprehensive demonstration of all features.
package theme
