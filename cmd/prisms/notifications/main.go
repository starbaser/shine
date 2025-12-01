package main

import (
	"fmt"
	"log"
	"os"
	"strings"
	"time"

	"github.com/charmbracelet/bubbles/viewport"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/starbased-co/shine/cmd/prisms/internal/theme"
)

type model struct {
	viewport      viewport.Model
	notifications []Notification
	ready         bool
	width         int
	height        int
	err           error
}

type tickMsg time.Time
type historyMsg struct {
	notifications []Notification
	err           error
}

func initialModel() model {
	return model{
		viewport: viewport.New(0, 0),
	}
}

func (m model) Init() tea.Cmd {
	return tea.Batch(
		fetchHistory,
		tickEvery(5*time.Second),
	)
}

func (m model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	var cmd tea.Cmd

	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.String() {
		case "q", "esc", "ctrl+c":
			return m, tea.Quit
		case "d":
			if err := DismissAll(); err != nil {
				m.err = err
			}
			return m, fetchHistory
		}

	case tea.WindowSizeMsg:
		m.width = msg.Width
		m.height = msg.Height

		if !m.ready {
			m.viewport = viewport.New(m.width, m.height-4)
			m.viewport.YPosition = 3
			m.ready = true
		} else {
			m.viewport.Width = m.width
			m.viewport.Height = m.height - 4
		}

		m.viewport.SetContent(m.renderNotifications())

	case historyMsg:
		m.err = msg.err
		if msg.err == nil {
			m.notifications = msg.notifications
		}
		m.viewport.SetContent(m.renderNotifications())

	case tickMsg:
		return m, tea.Batch(fetchHistory, tickEvery(5*time.Second))
	}

	m.viewport, cmd = m.viewport.Update(msg)
	return m, cmd
}

func (m model) View() string {
	if !m.ready {
		return "Loading..."
	}

	t := theme.Current()

	// Header
	title := lipgloss.NewStyle().
		Foreground(t.Primary()).
		Bold(true).
		Render(fmt.Sprintf("%s Notifications", theme.IconNotify))

	header := lipgloss.NewStyle().
		Width(m.width).
		BorderBottom(true).
		BorderStyle(lipgloss.RoundedBorder()).
		BorderForeground(t.BorderSubtle()).
		Padding(0, 1).
		Render(title)

	// Error display
	var errorBar string
	if m.err != nil {
		errorBar = lipgloss.NewStyle().
			Foreground(t.Error()).
			Width(m.width).
			Padding(0, 1).
			Render(fmt.Sprintf("%s Error: %v", theme.IconError, m.err))
	}

	// Footer
	helpStyle := lipgloss.NewStyle().
		Foreground(t.TextMuted()).
		Render("j/k: scroll • d: dismiss all • esc: quit")

	footer := lipgloss.NewStyle().
		Width(m.width).
		BorderTop(true).
		BorderStyle(lipgloss.RoundedBorder()).
		BorderForeground(t.BorderSubtle()).
		Padding(0, 1).
		Render(helpStyle)

	// Compose view
	content := []string{header}
	if errorBar != "" {
		content = append(content, errorBar)
	}
	content = append(content, m.viewport.View(), footer)

	return lipgloss.JoinVertical(lipgloss.Left, content...)
}

// renderNotifications renders the notification list content.
func (m model) renderNotifications() string {
	if len(m.notifications) == 0 {
		return m.renderEmpty()
	}

	t := theme.Current()
	cards := make([]string, 0, len(m.notifications))

	for _, notif := range m.notifications {
		// Card header (app name + timestamp)
		appStyle := lipgloss.NewStyle().
			Foreground(t.Primary()).
			Bold(true)

		timeStyle := lipgloss.NewStyle().
			Foreground(t.TextMuted())

		header := lipgloss.JoinHorizontal(
			lipgloss.Left,
			appStyle.Render(notif.AppName),
			timeStyle.Render("  "+RelativeTime(notif.Timestamp)),
		)

		// Card body
		summaryStyle := lipgloss.NewStyle().
			Foreground(t.TextPrimary()).
			Bold(true)

		bodyStyle := lipgloss.NewStyle().
			Foreground(t.TextPrimary())

		// Build card content
		var cardContent []string
		cardContent = append(cardContent, summaryStyle.Render(notif.Summary))
		if notif.Body != "" && notif.Body != notif.Summary {
			// Truncate body if too long
			body := notif.Body
			if len(body) > 200 {
				body = body[:200] + "..."
			}
			cardContent = append(cardContent, bodyStyle.Render(body))
		}

		content := lipgloss.JoinVertical(
			lipgloss.Left,
			header,
			strings.Join(cardContent, "\n"),
		)

		// Wrap in card style
		card := lipgloss.NewStyle().
			Width(m.width - 4).
			Background(t.Surface1()).
			Foreground(t.TextPrimary()).
			Padding(1, 2).
			Border(lipgloss.RoundedBorder()).
			BorderForeground(t.BorderSubtle()).
			Render(content)

		cards = append(cards, card)
	}

	return lipgloss.JoinVertical(lipgloss.Left, cards...)
}

// renderEmpty renders the empty state.
func (m model) renderEmpty() string {
	t := theme.Current()

	emptyStyle := lipgloss.NewStyle().
		Foreground(t.TextMuted()).
		Padding(2, 0).
		Align(lipgloss.Center).
		Width(m.width)

	return emptyStyle.Render("No notifications")
}

// fetchHistory fetches notification history from Dunst.
func fetchHistory() tea.Msg {
	notifications, err := FetchHistory()
	return historyMsg{
		notifications: notifications,
		err:           err,
	}
}

// tickEvery creates a command that sends a tick message after the specified duration.
func tickEvery(d time.Duration) tea.Cmd {
	return tea.Tick(d, func(t time.Time) tea.Msg {
		return tickMsg(t)
	})
}

func main() {
	if _, err := tea.NewProgram(
		initialModel(),
		tea.WithAltScreen(),
		tea.WithMouseCellMotion(),
	).Run(); err != nil {
		log.Printf("Error running notifications prism: %v", err)
		os.Exit(1)
	}
}
