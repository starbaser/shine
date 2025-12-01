package theme

import "github.com/charmbracelet/lipgloss"

// StyleBuilder provides reusable Lipgloss style constructors based on the current theme.

// PanelStyle creates a base panel style with theme background and border.
func PanelStyle() lipgloss.Style {
	t := Current()
	return lipgloss.NewStyle().
		Background(t.Background()).
		Foreground(t.Foreground()).
		Border(lipgloss.RoundedBorder()).
		BorderForeground(t.BorderSubtle())
}

// ActivePanelStyle creates a panel style with active border highlighting.
func ActivePanelStyle() lipgloss.Style {
	t := Current()
	return PanelStyle().
		BorderForeground(t.BorderActive())
}

// CardStyle creates a raised surface card with padding.
func CardStyle() lipgloss.Style {
	t := Current()
	return lipgloss.NewStyle().
		Background(t.Surface1()).
		Foreground(t.TextPrimary()).
		Padding(1, 2).
		Border(lipgloss.RoundedBorder()).
		BorderForeground(t.BorderSubtle())
}

// TitleStyle creates a prominent title style.
func TitleStyle() lipgloss.Style {
	t := Current()
	return lipgloss.NewStyle().
		Foreground(t.Primary()).
		Bold(true)
}

// SubtitleStyle creates a secondary title style.
func SubtitleStyle() lipgloss.Style {
	t := Current()
	return lipgloss.NewStyle().
		Foreground(t.Secondary()).
		Bold(true)
}

// TextStyle creates a base text style.
func TextStyle() lipgloss.Style {
	t := Current()
	return lipgloss.NewStyle().
		Foreground(t.TextPrimary())
}

// MutedTextStyle creates a muted/subtle text style.
func MutedTextStyle() lipgloss.Style {
	t := Current()
	return lipgloss.NewStyle().
		Foreground(t.TextMuted())
}

// ErrorStyle creates an error message style.
func ErrorStyle() lipgloss.Style {
	t := Current()
	return lipgloss.NewStyle().
		Foreground(t.Error()).
		Bold(true)
}

// WarningStyle creates a warning message style.
func WarningStyle() lipgloss.Style {
	t := Current()
	return lipgloss.NewStyle().
		Foreground(t.Warning()).
		Bold(true)
}

// SuccessStyle creates a success message style.
func SuccessStyle() lipgloss.Style {
	t := Current()
	return lipgloss.NewStyle().
		Foreground(t.Success()).
		Bold(true)
}

// InfoStyle creates an informational message style.
func InfoStyle() lipgloss.Style {
	t := Current()
	return lipgloss.NewStyle().
		Foreground(t.Info())
}

// StatusBarStyle creates a status bar with distinct surface.
func StatusBarStyle() lipgloss.Style {
	t := Current()
	return lipgloss.NewStyle().
		Background(t.Surface2()).
		Foreground(t.TextPrimary()).
		Padding(0, 1)
}

// ButtonStyle creates a clickable button style.
func ButtonStyle(active bool) lipgloss.Style {
	t := Current()
	if active {
		return lipgloss.NewStyle().
			Background(t.Primary()).
			Foreground(t.Background()).
			Padding(0, 2).
			Bold(true)
	}
	return lipgloss.NewStyle().
		Background(t.Surface1()).
		Foreground(t.TextSecondary()).
		Padding(0, 2)
}

// BadgeStyle creates a small badge/tag style.
func BadgeStyle(color lipgloss.TerminalColor) lipgloss.Style {
	t := Current()
	return lipgloss.NewStyle().
		Background(color).
		Foreground(t.Background()).
		Padding(0, 1).
		Bold(true)
}

// ListItemStyle creates a list item style with optional selection.
func ListItemStyle(selected bool) lipgloss.Style {
	t := Current()
	if selected {
		return lipgloss.NewStyle().
			Background(t.Surface1()).
			Foreground(t.Primary()).
			Padding(0, 1)
	}
	return lipgloss.NewStyle().
		Foreground(t.TextPrimary()).
		Padding(0, 1)
}

// SeparatorStyle creates a horizontal separator line.
func SeparatorStyle() lipgloss.Style {
	t := Current()
	return lipgloss.NewStyle().
		Foreground(t.BorderSubtle())
}

// CodeBlockStyle creates a styled code block.
func CodeBlockStyle() lipgloss.Style {
	t := Current()
	return lipgloss.NewStyle().
		Background(t.Surface2()).
		Foreground(t.TextPrimary()).
		Padding(1, 2).
		Border(lipgloss.RoundedBorder()).
		BorderForeground(t.BorderSubtle())
}

// ProgressBarStyle creates a progress bar style.
func ProgressBarStyle() (filled, empty lipgloss.Style) {
	t := Current()
	filledStyle := lipgloss.NewStyle().
		Background(t.Primary()).
		Foreground(t.Background())

	emptyStyle := lipgloss.NewStyle().
		Background(t.Surface1()).
		Foreground(t.TextMuted())

	return filledStyle, emptyStyle
}

// GlassPanel creates a translucent panel effect using borders and subtle colors.
func GlassPanel() lipgloss.Style {
	t := Current()
	return lipgloss.NewStyle().
		Background(t.Surface0()).
		Foreground(t.TextPrimary()).
		Border(GlassBorder()).
		BorderForeground(t.BorderSubtle()).
		Padding(1, 2)
}

// HighlightStyle creates a highlighted text style.
func HighlightStyle() lipgloss.Style {
	t := Current()
	return lipgloss.NewStyle().
		Background(t.Accent()).
		Foreground(t.Background()).
		Bold(true)
}
