package main

import (
	"fmt"
	"log"
	"time"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/starbased-co/shine/cmd/prisms/internal/theme"
)

func main() {
	fmt.Print("\033]0;shine-clock\007")

	p := tea.NewProgram(
		initialModel(),
		tea.WithAltScreen(),
	)

	if _, err := p.Run(); err != nil {
		log.Fatal(err)
	}
}

type tickMsg time.Time

type model struct {
	currentTime time.Time
	width       int
	height      int
}

func initialModel() model {
	return model{
		currentTime: time.Now(),
		width:       20,
		height:      5,
	}
}

func (m model) Init() tea.Cmd {
	return tickCmd()
}

func tickCmd() tea.Cmd {
	return tea.Tick(time.Second, func(t time.Time) tea.Msg {
		return tickMsg(t)
	})
}

func (m model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		m.width = msg.Width
		m.height = msg.Height
		return m, nil

	case tea.KeyMsg:
		switch msg.Type {
		case tea.KeyEsc:
			return m, tea.Quit
		}

	case tickMsg:
		m.currentTime = time.Time(msg)
		return m, tickCmd()
	}

	return m, nil
}

func (m model) View() string {
	t := theme.Current()

	timeStyle := lipgloss.NewStyle().
		Foreground(t.Accent()).
		Background(t.Background()).
		Bold(true).
		Align(lipgloss.Center, lipgloss.Center).
		Width(m.width).
		Height(m.height)

	dateStyle := lipgloss.NewStyle().
		Foreground(t.TextSecondary()).
		Align(lipgloss.Center).
		Width(m.width)

	timeStr := m.currentTime.Format("15:04")
	dateStr := m.currentTime.Format("Mon, Jan 2")

	content := lipgloss.JoinVertical(lipgloss.Center,
		timeStyle.Render(timeStr),
		dateStyle.Render(dateStr),
	)

	panel := theme.GlassPanel().
		Width(m.width).
		Height(m.height).
		Align(lipgloss.Center, lipgloss.Center)

	return panel.Render(content)
}
