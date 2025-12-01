package theme_test

import (
	"fmt"

	"github.com/starbased-co/shine/cmd/prisms/internal/theme"

	"github.com/charmbracelet/lipgloss"
)

// Example demonstrates basic theme usage.
func Example() {
	// Get current theme (Srcery auto-registered)
	t := theme.Current()

	// Use theme colors directly
	style := lipgloss.NewStyle().
		Foreground(t.Primary()).
		Background(t.Surface0()).
		Padding(1, 2)

	fmt.Println(style.Render("Themed text"))
}

// Example_styleBuilders demonstrates pre-built style usage.
func Example_styleBuilders() {
	// Panel styles
	panel := theme.PanelStyle().
		Width(40).
		Height(10).
		Render("Panel content")

	activePanel := theme.ActivePanelStyle().
		Width(40).
		Height(10).
		Render("Active panel")

	// Card style
	card := theme.CardStyle().Render("Card content")

	// Title styles
	title := theme.TitleStyle().Render("Main Title")
	subtitle := theme.SubtitleStyle().Render("Subtitle")

	// Status messages
	success := theme.SuccessStyle().Render(fmt.Sprintf("%s Success", theme.IconSuccess))
	error := theme.ErrorStyle().Render(fmt.Sprintf("%s Error", theme.IconError))
	warning := theme.WarningStyle().Render(fmt.Sprintf("%s Warning", theme.IconWarning))

	fmt.Println(panel)
	fmt.Println(activePanel)
	fmt.Println(card)
	fmt.Println(title)
	fmt.Println(subtitle)
	fmt.Println(success)
	fmt.Println(error)
	fmt.Println(warning)
}

// Example_icons demonstrates icon usage.
func Example_icons() {
	// System status icons
	cpu := fmt.Sprintf("%s CPU: 45%%", theme.IconCPU)
	memory := fmt.Sprintf("%s Memory: 8.2 GB", theme.IconMemory)
	disk := fmt.Sprintf("%s Disk: 120 GB free", theme.IconDisk)

	// Battery status
	batteryIcon := theme.GetBatteryIcon(75, false)
	battery := fmt.Sprintf("%s 75%%", batteryIcon)

	chargingIcon := theme.GetBatteryIcon(50, true)
	charging := fmt.Sprintf("%s 50%% (charging)", chargingIcon)

	// Volume status
	volumeIcon := theme.GetVolumeIcon(80, false)
	volume := fmt.Sprintf("%s 80%%", volumeIcon)

	mutedIcon := theme.GetVolumeIcon(0, true)
	muted := fmt.Sprintf("%s Muted", mutedIcon)

	fmt.Println(cpu)
	fmt.Println(memory)
	fmt.Println(disk)
	fmt.Println(battery)
	fmt.Println(charging)
	fmt.Println(volume)
	fmt.Println(muted)
}

// Example_customBorders demonstrates border customization.
func Example_customBorders() {
	t := theme.Current()

	// Glass border
	glass := lipgloss.NewStyle().
		Border(theme.GlassBorder()).
		BorderForeground(t.BorderSubtle()).
		Padding(1, 2).
		Render("Glass panel")

	// Thick border
	thick := lipgloss.NewStyle().
		Border(theme.ThickBorder()).
		BorderForeground(t.Primary()).
		Padding(1, 2).
		Render("Thick border")

	// Dashed border
	dashed := lipgloss.NewStyle().
		Border(theme.DashedBorder()).
		BorderForeground(t.BorderSubtle()).
		Padding(1, 2).
		Render("Dashed border")

	// Double border
	double := lipgloss.NewStyle().
		Border(theme.DoubleBorder()).
		BorderForeground(t.Accent()).
		Padding(1, 2).
		Render("Double border")

	fmt.Println(glass)
	fmt.Println(thick)
	fmt.Println(dashed)
	fmt.Println(double)
}

// Example_listItems demonstrates list item styling.
func Example_listItems() {
	items := []string{"Item 1", "Item 2", "Item 3"}
	selectedIdx := 1

	for i, item := range items {
		styled := theme.ListItemStyle(i == selectedIdx).Render(item)
		fmt.Println(styled)
	}
}

// Example_progressBar demonstrates progress bar styling.
func Example_progressBar() {
	filled, empty := theme.ProgressBarStyle()

	progress := 60 // 60%
	barWidth := 20

	filledWidth := (progress * barWidth) / 100
	emptyWidth := barWidth - filledWidth

	bar := lipgloss.JoinHorizontal(
		lipgloss.Left,
		filled.Render(lipgloss.NewStyle().Width(filledWidth).Render("")),
		empty.Render(lipgloss.NewStyle().Width(emptyWidth).Render("")),
	)

	fmt.Printf("Progress: %s %d%%\n", bar, progress)
}

// Example_statusBar demonstrates status bar creation.
func Example_statusBar() {
	t := theme.Current()

	left := lipgloss.NewStyle().
		Foreground(t.Primary()).
		Render(fmt.Sprintf("%s My App", theme.IconDashboard))

	center := lipgloss.NewStyle().
		Foreground(t.TextSecondary()).
		Render(fmt.Sprintf("%s 15:30", theme.IconClock))

	right := lipgloss.NewStyle().
		Foreground(t.Success()).
		Render(fmt.Sprintf("%s Connected", theme.IconWifi))

	statusBar := theme.StatusBarStyle().
		Width(80).
		Render(lipgloss.JoinHorizontal(
			lipgloss.Left,
			left,
			lipgloss.NewStyle().Width(30).Align(lipgloss.Center).Render(center),
			lipgloss.NewStyle().Width(30).Align(lipgloss.Right).Render(right),
		))

	fmt.Println(statusBar)
}

// Example_themeSwitching demonstrates runtime theme switching.
func Example_themeSwitching() {
	// List available themes
	themes := theme.List()
	fmt.Printf("Available themes: %v\n", themes)

	// Get current theme name
	current := theme.Current().Name()
	fmt.Printf("Current theme: %s\n", current)

	// Switch theme (when more themes are added)
	err := theme.SetCurrent("srcery")
	if err != nil {
		fmt.Printf("Error: %v\n", err)
	}

	// Get specific theme without activating
	srcery := theme.Get("srcery")
	if srcery != nil {
		fmt.Printf("Srcery theme available: %s\n", srcery.Name())
	}
}
