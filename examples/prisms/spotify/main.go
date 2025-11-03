// Spotify Prism - Display currently playing track with real D-Bus integration
//
// This prism demonstrates:
// - Real Spotify integration via D-Bus MPRIS2 interface
// - Interactive controls (play/pause, next/previous, seek)
// - Beautiful progress bar with lipgloss styling
// - Keyboard input handling
// - Auto-update on track changes
// - Graceful handling when Spotify is not running

package main

import (
	"fmt"
	"log"
	"os"
	"time"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/godbus/dbus/v5"
)

func main() {
	// Set window title for tracking (REQUIRED)
	fmt.Print("\033]0;shine-spotify\007")

	// Check for mock mode flag
	mockMode := false
	if len(os.Args) > 1 && os.Args[1] == "--mock" {
		mockMode = true
	}

	// Create Bubble Tea program without alt screen
	p := tea.NewProgram(initialModel(mockMode))

	if _, err := p.Run(); err != nil {
		log.Fatal(err)
	}
}

// tickMsg is sent every second for progress updates
type tickMsg time.Time

// spotifyStatusMsg contains updated Spotify status
type spotifyStatusMsg struct {
	track     track
	isPlaying bool
	err       error
}

// track represents a Spotify track
type track struct {
	title    string
	artist   string
	album    string
	duration time.Duration // Total duration
	position time.Duration // Current position
}

// model holds the application state
type model struct {
	currentTrack  track
	isPlaying     bool
	spotifyRunning bool
	mockMode      bool
	lastError     error
	width         int
	height        int
}

// initialModel creates the initial application state
func initialModel(mockMode bool) model {
	m := model{
		currentTrack:  track{},
		isPlaying:     false,
		spotifyRunning: false,
		mockMode:      mockMode,
		width:         80,
		height:        3,
	}

	// If in mock mode, populate with example data
	if mockMode {
		m.currentTrack = track{
			title:    "The Less I Know The Better",
			artist:   "Tame Impala",
			album:    "Currents",
			duration: 3*time.Minute + 36*time.Second,
			position: 1*time.Minute + 23*time.Second,
		}
		m.isPlaying = true
		m.spotifyRunning = true
	}

	return m
}

// Init returns the initial command
func (m model) Init() tea.Cmd {
	if m.mockMode {
		return tickCmd()
	}
	return tea.Batch(
		tickCmd(),
		fetchSpotifyStatus(),
	)
}

// tickCmd creates a command that ticks every second
func tickCmd() tea.Cmd {
	return tea.Tick(1*time.Second, func(t time.Time) tea.Msg {
		return tickMsg(t)
	})
}

// Update handles messages and updates the model
func (m model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		m.width = msg.Width
		m.height = msg.Height
		return m, nil

	case tea.KeyMsg:
		switch msg.String() {
		case "ctrl+c", "q":
			return m, tea.Quit

		case " ": // Spacebar - play/pause
			if m.spotifyRunning {
				go spotifyPlayPause()
				return m, fetchSpotifyStatus()
			}
			return m, nil

		case "n": // Next track
			if m.spotifyRunning {
				go spotifyNext()
				return m, fetchSpotifyStatus()
			}
			return m, nil

		case "p": // Previous track
			if m.spotifyRunning {
				go spotifyPrevious()
				return m, fetchSpotifyStatus()
			}
			return m, nil

		case "left": // Seek backward 5s
			if m.spotifyRunning {
				go spotifySeek(-5 * time.Second)
				return m, fetchSpotifyStatus()
			}
			return m, nil

		case "right": // Seek forward 5s
			if m.spotifyRunning {
				go spotifySeek(5 * time.Second)
				return m, fetchSpotifyStatus()
			}
			return m, nil
		}

	case spotifyStatusMsg:
		if msg.err == nil {
			m.currentTrack = msg.track
			m.isPlaying = msg.isPlaying
			m.spotifyRunning = true
			m.lastError = nil
		} else {
			m.spotifyRunning = false
			m.lastError = msg.err
		}
		return m, nil

	case tickMsg:
		// In mock mode, just update progress
		if m.mockMode && m.isPlaying {
			m.currentTrack.position += time.Second
			if m.currentTrack.position >= m.currentTrack.duration {
				m.currentTrack.position = 0
			}
			return m, tickCmd()
		}
		// Fetch fresh Spotify status every second
		return m, tea.Batch(
			tickCmd(),
			fetchSpotifyStatus(),
		)
	}

	return m, nil
}

// View renders the UI
func (m model) View() string {
	// If Spotify is not running, show a friendly message
	if !m.spotifyRunning {
		return renderNotRunning()
	}

	// Define color scheme
	var (
		spotifyGreen = lipgloss.Color("#1DB954")
		textColor    = lipgloss.Color("#FFFFFF")
		dimColor     = lipgloss.Color("#B3B3B3")
		progressFill = lipgloss.Color("#1DB954")
		progressBg   = lipgloss.Color("#404040")
	)

	// Define styles
	containerStyle := lipgloss.NewStyle().
		Border(lipgloss.RoundedBorder()).
		BorderForeground(spotifyGreen).
		Padding(0, 2)

	statusStyle := lipgloss.NewStyle().
		Foreground(spotifyGreen).
		Bold(true).
		Padding(0, 1)

	trackStyle := lipgloss.NewStyle().
		Foreground(textColor).
		Bold(true)

	artistStyle := lipgloss.NewStyle().
		Foreground(dimColor)

	timeStyle := lipgloss.NewStyle().
		Foreground(dimColor)

	helpStyle := lipgloss.NewStyle().
		Foreground(lipgloss.Color("#535353")).
		Italic(true)

	// Playback status icon
	statusIcon := "▶"
	if m.isPlaying {
		statusIcon = "▶"
	} else {
		statusIcon = "⏸"
	}

	// Build track info line
	trackLine := lipgloss.JoinHorizontal(
		lipgloss.Center,
		statusStyle.Render(statusIcon),
		trackStyle.Render(m.currentTrack.title),
		lipgloss.NewStyle().Padding(0, 1).Render("•"),
		artistStyle.Render(m.currentTrack.artist),
	)

	// Build progress bar
	progressBar := renderProgressBar(m.currentTrack.position, m.currentTrack.duration, 40, progressFill, progressBg)
	timeInfo := timeStyle.Render(fmt.Sprintf("%s / %s",
		formatDuration(m.currentTrack.position),
		formatDuration(m.currentTrack.duration)))

	progressLine := lipgloss.JoinHorizontal(
		lipgloss.Left,
		progressBar,
		lipgloss.NewStyle().Padding(0, 1).Render(""),
		timeInfo,
	)

	// Help text
	helpText := helpStyle.Render("space: play/pause • n/p: next/prev • ←/→: seek • q: quit")

	// Combine all elements vertically
	content := lipgloss.JoinVertical(
		lipgloss.Left,
		trackLine,
		progressLine,
		helpText,
	)

	return containerStyle.Render(content)
}

// renderNotRunning displays a message when Spotify is not running
func renderNotRunning() string {
	style := lipgloss.NewStyle().
		Border(lipgloss.RoundedBorder()).
		BorderForeground(lipgloss.Color("#535353")).
		Foreground(lipgloss.Color("#B3B3B3")).
		Padding(0, 2)

	iconStyle := lipgloss.NewStyle().
		Foreground(lipgloss.Color("#535353")).
		Bold(true)

	messageStyle := lipgloss.NewStyle().
		Foreground(lipgloss.Color("#B3B3B3"))

	content := lipgloss.JoinHorizontal(
		lipgloss.Center,
		iconStyle.Render("🎵"),
		lipgloss.NewStyle().Padding(0, 1).Render(""),
		messageStyle.Render("Spotify is not running"),
	)

	return style.Render(content)
}

// renderProgressBar creates a styled progress bar
func renderProgressBar(position, duration time.Duration, width int, fillColor, bgColor lipgloss.Color) string {
	if duration == 0 {
		return lipgloss.NewStyle().
			Foreground(bgColor).
			Render(lipgloss.NewStyle().Width(width).Render(""))
	}

	progress := float64(position) / float64(duration)
	filled := int(progress * float64(width))
	if filled > width {
		filled = width
	}

	fillStyle := lipgloss.NewStyle().Foreground(fillColor)
	bgStyle := lipgloss.NewStyle().Foreground(bgColor)

	var bar string
	for i := 0; i < width; i++ {
		if i < filled {
			bar += fillStyle.Render("━")
		} else {
			bar += bgStyle.Render("─")
		}
	}

	return bar
}

// formatDuration formats a duration as MM:SS
func formatDuration(d time.Duration) string {
	minutes := int(d.Minutes())
	seconds := int(d.Seconds()) % 60
	return fmt.Sprintf("%d:%02d", minutes, seconds)
}

// D-Bus integration functions

// fetchSpotifyStatus creates a command to fetch current Spotify status
func fetchSpotifyStatus() tea.Cmd {
	return func() tea.Msg {
		track, isPlaying, err := getSpotifyStatus()
		return spotifyStatusMsg{
			track:     track,
			isPlaying: isPlaying,
			err:       err,
		}
	}
}

// getSpotifyStatus queries Spotify via D-Bus MPRIS2 interface
func getSpotifyStatus() (track, bool, error) {
	conn, err := dbus.SessionBus()
	if err != nil {
		return track{}, false, err
	}

	obj := conn.Object("org.mpris.MediaPlayer2.spotify", "/org/mpris/MediaPlayer2")

	// Get metadata
	var metadata map[string]dbus.Variant
	err = obj.Call("org.freedesktop.DBus.Properties.Get", 0,
		"org.mpris.MediaPlayer2.Player", "Metadata").Store(&metadata)
	if err != nil {
		return track{}, false, err
	}

	// Get playback status
	var status string
	err = obj.Call("org.freedesktop.DBus.Properties.Get", 0,
		"org.mpris.MediaPlayer2.Player", "PlaybackStatus").Store(&status)
	if err != nil {
		return track{}, false, err
	}

	// Get position
	var position int64
	err = obj.Call("org.freedesktop.DBus.Properties.Get", 0,
		"org.mpris.MediaPlayer2.Player", "Position").Store(&position)
	if err != nil {
		position = 0
	}

	// Parse metadata
	title := ""
	if v, ok := metadata["xesam:title"]; ok {
		title = v.Value().(string)
	}

	artist := "Unknown Artist"
	if v, ok := metadata["xesam:artist"]; ok {
		artists := v.Value().([]string)
		if len(artists) > 0 {
			artist = artists[0]
		}
	}

	album := ""
	if v, ok := metadata["xesam:album"]; ok {
		album = v.Value().(string)
	}

	duration := time.Duration(0)
	if v, ok := metadata["mpris:length"]; ok {
		duration = time.Duration(v.Value().(int64)) * time.Microsecond
	}

	return track{
		title:    title,
		artist:   artist,
		album:    album,
		duration: duration,
		position: time.Duration(position) * time.Microsecond,
	}, status == "Playing", nil
}

// spotifyPlayPause toggles play/pause
func spotifyPlayPause() error {
	conn, err := dbus.SessionBus()
	if err != nil {
		return err
	}

	obj := conn.Object("org.mpris.MediaPlayer2.spotify", "/org/mpris/MediaPlayer2")
	return obj.Call("org.mpris.MediaPlayer2.Player.PlayPause", 0).Err
}

// spotifyNext skips to next track
func spotifyNext() error {
	conn, err := dbus.SessionBus()
	if err != nil {
		return err
	}

	obj := conn.Object("org.mpris.MediaPlayer2.spotify", "/org/mpris/MediaPlayer2")
	return obj.Call("org.mpris.MediaPlayer2.Player.Next", 0).Err
}

// spotifyPrevious goes to previous track
func spotifyPrevious() error {
	conn, err := dbus.SessionBus()
	if err != nil {
		return err
	}

	obj := conn.Object("org.mpris.MediaPlayer2.spotify", "/org/mpris/MediaPlayer2")
	return obj.Call("org.mpris.MediaPlayer2.Player.Previous", 0).Err
}

// spotifySeek seeks by the specified offset
func spotifySeek(offset time.Duration) error {
	conn, err := dbus.SessionBus()
	if err != nil {
		return err
	}

	obj := conn.Object("org.mpris.MediaPlayer2.spotify", "/org/mpris/MediaPlayer2")
	offsetMicroseconds := int64(offset / time.Microsecond)
	return obj.Call("org.mpris.MediaPlayer2.Player.Seek", 0, offsetMicroseconds).Err
}
