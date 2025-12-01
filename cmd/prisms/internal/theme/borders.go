package theme

import "github.com/charmbracelet/lipgloss"

// GlassBorder returns a custom border with rounded corners for glass panel effects.
func GlassBorder() lipgloss.Border {
	return lipgloss.Border{
		Top:         "─",
		Bottom:      "─",
		Left:        "│",
		Right:       "│",
		TopLeft:     "╭",
		TopRight:    "╮",
		BottomLeft:  "╰",
		BottomRight: "╯",
	}
}

// ThickBorder returns a border with heavy/thick lines.
func ThickBorder() lipgloss.Border {
	return lipgloss.Border{
		Top:         "━",
		Bottom:      "━",
		Left:        "┃",
		Right:       "┃",
		TopLeft:     "┏",
		TopRight:    "┓",
		BottomLeft:  "┗",
		BottomRight: "┛",
	}
}

// DoubleBorder returns a double-line border.
func DoubleBorder() lipgloss.Border {
	return lipgloss.Border{
		Top:         "═",
		Bottom:      "═",
		Left:        "║",
		Right:       "║",
		TopLeft:     "╔",
		TopRight:    "╗",
		BottomLeft:  "╚",
		BottomRight: "╝",
	}
}

// DashedBorder returns a border with dashed lines.
func DashedBorder() lipgloss.Border {
	return lipgloss.Border{
		Top:         "┄",
		Bottom:      "┄",
		Left:        "┆",
		Right:       "┆",
		TopLeft:     "╭",
		TopRight:    "╮",
		BottomLeft:  "╰",
		BottomRight: "╯",
	}
}

// MinimalBorder returns a subtle border with minimal visual weight.
func MinimalBorder() lipgloss.Border {
	return lipgloss.Border{
		Top:         "─",
		Bottom:      "─",
		Left:        "│",
		Right:       "│",
		TopLeft:     "┌",
		TopRight:    "┐",
		BottomLeft:  "└",
		BottomRight: "┘",
	}
}

// BracketBorder returns a border using bracket-style corners.
func BracketBorder() lipgloss.Border {
	return lipgloss.Border{
		Top:         "─",
		Bottom:      "─",
		Left:        "│",
		Right:       "│",
		TopLeft:     "⎾",
		TopRight:    "⏋",
		BottomLeft:  "⎿",
		BottomRight: "⏌",
	}
}
