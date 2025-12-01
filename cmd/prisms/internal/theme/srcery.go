package theme

import "github.com/charmbracelet/lipgloss"

// Srcery implements the Srcery colorscheme.
// Based on srcery.sh palette with vim highlight group semantics.
type Srcery struct{}

// Palette constants based on srcery.sh
const (
	// Base colors
	srceryBlack        = lipgloss.Color("#1C1B19")
	srceryRed          = lipgloss.Color("#EF2F27")
	srceryGreen        = lipgloss.Color("#519F50")
	srceryYellow       = lipgloss.Color("#FBB829")
	srceryBlue         = lipgloss.Color("#2C78BF")
	srceryMagenta      = lipgloss.Color("#E02C6D")
	srceryCyan         = lipgloss.Color("#0AAEB3")
	srceryWhite        = lipgloss.Color("#918175")
	srceryBrightBlack  = lipgloss.Color("#2D2C29")
	srceryBrightRed    = lipgloss.Color("#F75341")
	srceryBrightGreen  = lipgloss.Color("#98BC37")
	srceryBrightYellow = lipgloss.Color("#FED06E")
	srceryBrightBlue   = lipgloss.Color("#68A8E4")
	srceryBrightMagenta = lipgloss.Color("#FF5C8F")
	srceryBrightCyan   = lipgloss.Color("#53FDE9")
	srceryBrightWhite  = lipgloss.Color("#FCE8C3")

	// Extended grays
	srceryGray1 = lipgloss.Color("#262626")
	srceryGray2 = lipgloss.Color("#303030")
	srceryGray3 = lipgloss.Color("#3A3A3A")
	srceryGray4 = lipgloss.Color("#4E4E4E")
	srceryGray5 = lipgloss.Color("#626262")
	srceryGray6 = lipgloss.Color("#767676")

	// Special colors
	srceryOrange       = lipgloss.Color("#FF5F00")
	srceryBrightOrange = lipgloss.Color("#FF8700")
	srceryHardBlack    = lipgloss.Color("#121212")
	srceryXgray1       = lipgloss.Color("#262626")
	srceryXgray2       = lipgloss.Color("#303030")
	srceryXgray3       = lipgloss.Color("#3A3A3A")
	srceryXgray4       = lipgloss.Color("#4E4E4E")
	srceryXgray5       = lipgloss.Color("#626262")
)

func (s Srcery) Name() string {
	return "srcery"
}

// Base colors
func (s Srcery) Background() lipgloss.TerminalColor {
	return srceryBlack
}

func (s Srcery) Foreground() lipgloss.TerminalColor {
	return srceryBrightWhite
}

// Semantic colors
func (s Srcery) Primary() lipgloss.TerminalColor {
	return srceryBrightBlue
}

func (s Srcery) Secondary() lipgloss.TerminalColor {
	return srceryBrightCyan
}

func (s Srcery) Accent() lipgloss.TerminalColor {
	return srceryBrightMagenta
}

func (s Srcery) Muted() lipgloss.TerminalColor {
	return srceryGray5
}

// Status colors
func (s Srcery) Success() lipgloss.TerminalColor {
	return srceryBrightGreen
}

func (s Srcery) Warning() lipgloss.TerminalColor {
	return srceryBrightYellow
}

func (s Srcery) Error() lipgloss.TerminalColor {
	return srceryBrightRed
}

func (s Srcery) Info() lipgloss.TerminalColor {
	return srceryBrightCyan
}

// Surface colors
func (s Srcery) Surface0() lipgloss.TerminalColor {
	return srceryBlack
}

func (s Srcery) Surface1() lipgloss.TerminalColor {
	return srceryGray2
}

func (s Srcery) Surface2() lipgloss.TerminalColor {
	return srceryGray3
}

// Border colors
func (s Srcery) BorderSubtle() lipgloss.TerminalColor {
	return srceryGray4
}

func (s Srcery) BorderActive() lipgloss.TerminalColor {
	return srceryBrightBlue
}

// Text hierarchy
func (s Srcery) TextPrimary() lipgloss.TerminalColor {
	return srceryBrightWhite
}

func (s Srcery) TextSecondary() lipgloss.TerminalColor {
	return srceryWhite
}

func (s Srcery) TextMuted() lipgloss.TerminalColor {
	return srceryGray6
}

// Vim Highlight Group Mappings
// Extended methods for vim-style highlight groups

func (s Srcery) Normal() (fg, bg lipgloss.TerminalColor) {
	return srceryBrightWhite, srceryBlack
}

func (s Srcery) CursorLine() lipgloss.TerminalColor {
	return srceryGray2
}

func (s Srcery) Visual() lipgloss.TerminalColor {
	return srceryGray3
}

func (s Srcery) Comment() lipgloss.TerminalColor {
	return srceryGray5
}

func (s Srcery) String() lipgloss.TerminalColor {
	return srceryBrightGreen
}

func (s Srcery) Number() lipgloss.TerminalColor {
	return srceryBrightOrange
}

func (s Srcery) Constant() lipgloss.TerminalColor {
	return srceryBrightMagenta
}

func (s Srcery) Identifier() lipgloss.TerminalColor {
	return srceryBrightBlue
}

func (s Srcery) Statement() lipgloss.TerminalColor {
	return srceryRed
}

func (s Srcery) PreProc() lipgloss.TerminalColor {
	return srceryBrightOrange
}

func (s Srcery) Type() lipgloss.TerminalColor {
	return srceryBrightYellow
}

func (s Srcery) Special() lipgloss.TerminalColor {
	return srceryOrange
}

func (s Srcery) Underlined() lipgloss.TerminalColor {
	return srceryBrightBlue
}

func (s Srcery) Todo() lipgloss.TerminalColor {
	return srceryBrightYellow
}

func (s Srcery) Search() (fg, bg lipgloss.TerminalColor) {
	return srceryBlack, srceryBrightYellow
}

func (s Srcery) IncSearch() (fg, bg lipgloss.TerminalColor) {
	return srceryBlack, srceryBrightOrange
}

func (s Srcery) DiffAdd() lipgloss.TerminalColor {
	return srceryGreen
}

func (s Srcery) DiffChange() lipgloss.TerminalColor {
	return srceryYellow
}

func (s Srcery) DiffDelete() lipgloss.TerminalColor {
	return srceryRed
}

func (s Srcery) DiffText() lipgloss.TerminalColor {
	return srceryBlue
}

func init() {
	Register(Srcery{})
}
