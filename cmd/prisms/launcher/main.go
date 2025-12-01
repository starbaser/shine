package main

import (
	"fmt"
	"log"
	"os"
	"os/exec"
	"strings"

	"github.com/charmbracelet/bubbles/textinput"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/starbased-co/shine/cmd/prisms/internal/theme"
)

func main() {
	fmt.Print("\033]0;shine-launcher\007")

	entries, err := LoadDesktopEntries()
	if err != nil {
		log.Fatal(err)
	}

	p := tea.NewProgram(
		initialModel(entries),
		tea.WithAltScreen(),
	)

	if _, err := p.Run(); err != nil {
		log.Fatal(err)
	}
}

type model struct {
	allEntries      []DesktopEntry
	filteredEntries []DesktopEntry
	searchInput     textinput.Model
	cursor          int
	width           int
	height          int
}

func initialModel(entries []DesktopEntry) model {
	ti := textinput.New()
	ti.Placeholder = "Search applications..."
	ti.Focus()
	ti.CharLimit = 100
	ti.Width = 40

	return model{
		allEntries:      entries,
		filteredEntries: entries,
		searchInput:     ti,
		cursor:          0,
		width:           50,
		height:          20,
	}
}

func (m model) Init() tea.Cmd {
	return textinput.Blink
}

func (m model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	var cmd tea.Cmd

	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		m.width = msg.Width
		m.height = msg.Height
		return m, nil

	case tea.KeyMsg:
		switch msg.Type {
		case tea.KeyEsc, tea.KeyCtrlC:
			return m, tea.Quit

		case tea.KeyEnter:
			if len(m.filteredEntries) > 0 && m.cursor < len(m.filteredEntries) {
				selected := m.filteredEntries[m.cursor]
				if err := launchApp(selected); err != nil {
					return m, tea.Quit
				}
				return m, tea.Quit
			}

		case tea.KeyUp:
			if m.cursor > 0 {
				m.cursor--
			}

		case tea.KeyDown:
			if m.cursor < len(m.filteredEntries)-1 {
				m.cursor++
			}

		default:
			m.searchInput, cmd = m.searchInput.Update(msg)
			query := m.searchInput.Value()
			m.filteredEntries = FilterEntries(m.allEntries, query)
			if m.cursor >= len(m.filteredEntries) {
				m.cursor = max(0, len(m.filteredEntries)-1)
			}
			return m, cmd
		}
	}

	return m, nil
}

func (m model) View() string {
	t := theme.Current()

	searchStyle := lipgloss.NewStyle().
		Foreground(t.Info()).
		Bold(true)

	headerStyle := lipgloss.NewStyle().
		Foreground(t.TextSecondary()).
		Padding(0, 1)

	selectedStyle := lipgloss.NewStyle().
		Background(t.Surface1()).
		Foreground(t.Accent()).
		Padding(0, 1).
		Width(m.width - 4)

	itemStyle := lipgloss.NewStyle().
		Foreground(t.TextPrimary()).
		Padding(0, 1).
		Width(m.width - 4)

	commentStyle := lipgloss.NewStyle().
		Foreground(t.TextMuted()).
		Padding(0, 1).
		Width(m.width - 4)

	separatorStyle := lipgloss.NewStyle().
		Foreground(t.BorderSubtle())

	var header string
	header = lipgloss.JoinHorizontal(
		lipgloss.Left,
		searchStyle.Render(theme.IconSearch+" "),
		m.searchInput.View(),
	)

	header = headerStyle.Render(header)

	separator := separatorStyle.Render(strings.Repeat("─", m.width-4))

	maxVisible := m.height - 6
	if maxVisible < 5 {
		maxVisible = 5
	}

	var items []string
	visibleStart := 0
	visibleEnd := len(m.filteredEntries)

	if len(m.filteredEntries) > maxVisible {
		if m.cursor > maxVisible/2 {
			visibleStart = m.cursor - maxVisible/2
		}
		visibleEnd = visibleStart + maxVisible
		if visibleEnd > len(m.filteredEntries) {
			visibleEnd = len(m.filteredEntries)
			visibleStart = max(0, visibleEnd-maxVisible)
		}
	}

	for i := visibleStart; i < visibleEnd; i++ {
		entry := m.filteredEntries[i]
		appName := fmt.Sprintf("  %s", entry.Name)
		comment := ""
		if entry.Comment != "" {
			comment = fmt.Sprintf("  %s", entry.Comment)
		}

		if i == m.cursor {
			items = append(items, selectedStyle.Render(appName))
			if comment != "" {
				items = append(items, commentStyle.Render(comment))
			}
		} else {
			items = append(items, itemStyle.Render(appName))
			if comment != "" {
				items = append(items, commentStyle.Render(comment))
			}
		}
	}

	if len(items) == 0 {
		noResults := lipgloss.NewStyle().
			Foreground(t.TextMuted()).
			Padding(1).
			Align(lipgloss.Center).
			Width(m.width - 4).
			Render("No applications found")
		items = append(items, noResults)
	}

	content := lipgloss.JoinVertical(
		lipgloss.Left,
		header,
		separator,
		lipgloss.JoinVertical(lipgloss.Left, items...),
	)

	panel := theme.GlassPanel().
		Width(m.width).
		Height(m.height)

	return panel.Render(content)
}

func launchApp(entry DesktopEntry) error {
	parts := strings.Fields(entry.Exec)
	if len(parts) == 0 {
		return fmt.Errorf("empty exec field")
	}

	cmd := exec.Command(parts[0], parts[1:]...)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr

	return cmd.Start()
}

func max(a, b int) int {
	if a > b {
		return a
	}
	return b
}
