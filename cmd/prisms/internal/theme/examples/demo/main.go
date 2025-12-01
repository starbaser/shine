package main

import (
	"fmt"

	"github.com/starbased-co/shine/cmd/prisms/internal/theme"

	"github.com/charmbracelet/lipgloss"
)

func main() {
	t := theme.Current()
	fmt.Printf("Active theme: %s\n\n", t.Name())

	// Render comprehensive demo
	demo := lipgloss.JoinVertical(
		lipgloss.Left,
		renderHeader(),
		"",
		renderColorPalette(),
		"",
		renderStyleShowcase(),
		"",
		renderIconShowcase(),
		"",
		renderBorderShowcase(),
	)

	fmt.Println(demo)
}

func renderHeader() string {
	title := theme.TitleStyle().
		Width(80).
		Align(lipgloss.Center).
		Render("Shine Theme System Demo")

	subtitle := theme.SubtitleStyle().
		Width(80).
		Align(lipgloss.Center).
		Render("Srcery Colorscheme")

	return lipgloss.JoinVertical(lipgloss.Center, title, subtitle)
}

func renderColorPalette() string {
	t := theme.Current()

	colors := []struct {
		name  string
		color lipgloss.TerminalColor
	}{
		{"Primary", t.Primary()},
		{"Secondary", t.Secondary()},
		{"Accent", t.Accent()},
		{"Success", t.Success()},
		{"Warning", t.Warning()},
		{"Error", t.Error()},
		{"Info", t.Info()},
		{"Muted", t.Muted()},
	}

	var samples []string
	for _, c := range colors {
		sample := lipgloss.NewStyle().
			Background(c.color).
			Foreground(t.Background()).
			Padding(0, 2).
			Render(c.name)
		samples = append(samples, sample)
	}

	palette := lipgloss.JoinHorizontal(lipgloss.Left, samples...)
	return theme.CardStyle().
		Width(80).
		Render(lipgloss.JoinVertical(
			lipgloss.Left,
			theme.SubtitleStyle().Render("Color Palette"),
			"",
			palette,
		))
}

func renderStyleShowcase() string {
	// Status messages
	statusMsgs := lipgloss.JoinVertical(
		lipgloss.Left,
		theme.SuccessStyle().Render(fmt.Sprintf("%s Operation successful", theme.IconSuccess)),
		theme.WarningStyle().Render(fmt.Sprintf("%s Warning message", theme.IconWarning)),
		theme.ErrorStyle().Render(fmt.Sprintf("%s Error occurred", theme.IconError)),
		theme.InfoStyle().Render(fmt.Sprintf("%s Informational message", theme.IconInfo)),
	)

	// List items
	items := []string{"Active Item", "Normal Item", "Another Item"}
	var listItems []string
	for i, item := range items {
		styled := theme.ListItemStyle(i == 0).Render(item)
		listItems = append(listItems, styled)
	}
	list := lipgloss.JoinVertical(lipgloss.Left, listItems...)

	// Buttons
	activeBtn := theme.ButtonStyle(true).Render("Active")
	normalBtn := theme.ButtonStyle(false).Render("Normal")
	buttons := lipgloss.JoinHorizontal(lipgloss.Left, activeBtn, " ", normalBtn)

	content := lipgloss.JoinVertical(
		lipgloss.Left,
		theme.SubtitleStyle().Render("Styles Showcase"),
		"",
		"Status Messages:",
		statusMsgs,
		"",
		"List Items:",
		list,
		"",
		"Buttons:",
		buttons,
	)

	return theme.CardStyle().Width(80).Render(content)
}

func renderIconShowcase() string {
	t := theme.Current()

	// System icons
	systemRow := lipgloss.JoinHorizontal(
		lipgloss.Left,
		iconBox(theme.IconCPU, "CPU", t.Primary()),
		iconBox(theme.IconMemory, "Memory", t.Secondary()),
		iconBox(theme.IconDisk, "Disk", t.Accent()),
		iconBox(theme.IconNetwork, "Network", t.Success()),
	)

	// Battery icons
	batteryRow := lipgloss.JoinHorizontal(
		lipgloss.Left,
		iconBox(theme.IconBatteryFull, "Full", t.Success()),
		iconBox(theme.IconBatteryHalf, "Half", t.Warning()),
		iconBox(theme.IconBatteryEmpty, "Empty", t.Error()),
		iconBox(theme.IconBatteryCharging, "Charging", t.Info()),
	)

	// Volume icons
	volumeRow := lipgloss.JoinHorizontal(
		lipgloss.Left,
		iconBox(theme.IconVolumeHigh, "High", t.Primary()),
		iconBox(theme.IconVolumeMedium, "Med", t.Secondary()),
		iconBox(theme.IconVolumeLow, "Low", t.Warning()),
		iconBox(theme.IconVolumeMute, "Mute", t.Error()),
	)

	content := lipgloss.JoinVertical(
		lipgloss.Left,
		theme.SubtitleStyle().Render("Icons Showcase"),
		"",
		"System:",
		systemRow,
		"",
		"Battery:",
		batteryRow,
		"",
		"Volume:",
		volumeRow,
	)

	return theme.CardStyle().Width(80).Render(content)
}

func iconBox(icon, label string, color lipgloss.TerminalColor) string {
	t := theme.Current()

	iconStyle := lipgloss.NewStyle().
		Foreground(color).
		Width(8).
		Align(lipgloss.Center)

	labelStyle := lipgloss.NewStyle().
		Foreground(t.TextMuted()).
		Width(8).
		Align(lipgloss.Center)

	return lipgloss.JoinVertical(
		lipgloss.Center,
		iconStyle.Render(icon),
		labelStyle.Render(label),
	)
}

func renderBorderShowcase() string {
	t := theme.Current()

	borders := []struct {
		name   string
		border lipgloss.Border
		color  lipgloss.TerminalColor
	}{
		{"Glass", theme.GlassBorder(), t.BorderSubtle()},
		{"Thick", theme.ThickBorder(), t.Primary()},
		{"Double", theme.DoubleBorder(), t.Accent()},
		{"Dashed", theme.DashedBorder(), t.BorderSubtle()},
		{"Minimal", theme.MinimalBorder(), t.BorderSubtle()},
	}

	var samples []string
	for _, b := range borders {
		sample := lipgloss.NewStyle().
			Border(b.border).
			BorderForeground(b.color).
			Padding(1, 2).
			Render(b.name)
		samples = append(samples, sample)
	}

	borderRow := lipgloss.JoinHorizontal(lipgloss.Left, samples...)

	content := lipgloss.JoinVertical(
		lipgloss.Left,
		theme.SubtitleStyle().Render("Borders Showcase"),
		"",
		borderRow,
	)

	return theme.CardStyle().Width(80).Render(content)
}
