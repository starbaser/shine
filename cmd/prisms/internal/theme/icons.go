package theme

// Nerd Font icon constants for common system widgets.
// Requires a Nerd Font terminal font (e.g., JetBrainsMono Nerd Font).

const (
	// System status icons
	IconCPU     = ""  // CPU usage
	IconMemory  = "󰍛"  // Memory usage
	IconDisk    = "󰋊"  // Disk usage
	IconNetwork = "󰖩"  // Network status
	IconBattery = "󰁹"  // Battery

	// Battery states
	IconBatteryFull     = "󰁹"
	IconBatteryThreeQ   = "󰂀"
	IconBatteryHalf     = "󰁿"
	IconBatteryOneQ     = "󰁽"
	IconBatteryEmpty    = "󰁺"
	IconBatteryCharging = "󰂄"

	// Audio icons
	IconVolumeHigh   = "󰕾"
	IconVolumeMedium = "󰖀"
	IconVolumeLow    = "󰕿"
	IconVolumeMute   = "󰝟"

	// Connectivity
	IconWifi         = "󰖩"
	IconWifiOff      = "󰖪"
	IconBluetooth    = "󰂯"
	IconBluetoothOff = "󰂲"
	IconEthernet     = "󰈀"

	// Weather
	IconSunny       = "󰖙"
	IconCloudy      = "󰖐"
	IconRain        = "󰖗"
	IconSnow        = "󰼶"
	IconThunderstorm = "󰙾"
	IconFog         = "󰖑"

	// Time
	IconClock    = "󰥔"
	IconCalendar = "󰃭"
	IconTimer    = "󰔛"

	// Status indicators
	IconSuccess = "󰄬"  // Checkmark
	IconError   = "󰅖"  // X mark
	IconWarning = ""  // Warning triangle
	IconInfo    = "󰋽"  // Info circle

	// Direction arrows
	IconArrowUp    = ""
	IconArrowDown  = ""
	IconArrowLeft  = ""
	IconArrowRight = ""

	// UI elements
	IconMenu     = ""
	IconSettings = ""
	IconSearch   = ""
	IconFilter   = "󰈲"
	IconSort     = "󰒺"
	IconRefresh  = "󰑐"
	IconClose    = "󰅖"
	IconMinimize = ""
	IconMaximize = ""

	// Files and folders
	IconFolder     = "󰉋"
	IconFolderOpen = "󰝰"
	IconFile       = "󰈔"
	IconFileCode   = "󰈮"
	IconFileImage  = "󰈟"
	IconFileAudio  = "󰈣"
	IconFileVideo  = "󰈫"

	// Development
	IconGit       = "󰊢"
	IconGitBranch = ""
	IconGitCommit = ""
	IconGitMerge  = ""
	IconBug       = ""
	IconTerminal  = ""
	IconCode      = "󰅩"

	// Social
	IconChat    = "󰭹"
	IconMail    = "󰇰"
	IconUser    = "󰀄"
	IconUsers   = "󰀏"
	IconNotify  = "󰂚"

	// Playback
	IconPlay     = "󰐊"
	IconPause    = "󰏤"
	IconStop     = "󰓛"
	IconPrevious = "󰒮"
	IconNext     = "󰒭"

	// Power
	IconPower   = "󰐥"
	IconReboot  = "󰑓"
	IconSleep   = "󰒲"
	IconLogout  = "󰍃"

	// Widgets
	IconDashboard = "󰕮"
	IconChart     = "󰄧"
	IconGraph     = "󰺟"
	IconList      = "󰻃"
	IconGrid      = "󰕨"

	// Miscellaneous
	IconStar       = "󰓎"
	IconHeart      = "󰣐"
	IconBookmark   = "󰃃"
	IconPin        = "󰐃"
	IconLock       = "󰍁"
	IconUnlock     = "󰍂"
	IconEye        = "󰈈"
	IconEyeOff     = "󰈉"
	IconLink       = "󰌹"
	IconDownload   = "󰇚"
	IconUpload     = "󰕒"
	IconTrash      = "󰩺"
)

// GetBatteryIcon returns the appropriate battery icon based on charge percentage.
func GetBatteryIcon(percent int, charging bool) string {
	if charging {
		return IconBatteryCharging
	}

	switch {
	case percent >= 90:
		return IconBatteryFull
	case percent >= 60:
		return IconBatteryThreeQ
	case percent >= 40:
		return IconBatteryHalf
	case percent >= 20:
		return IconBatteryOneQ
	default:
		return IconBatteryEmpty
	}
}

// GetVolumeIcon returns the appropriate volume icon based on volume level.
func GetVolumeIcon(percent int, muted bool) string {
	if muted {
		return IconVolumeMute
	}

	switch {
	case percent >= 70:
		return IconVolumeHigh
	case percent >= 30:
		return IconVolumeMedium
	default:
		return IconVolumeLow
	}
}
