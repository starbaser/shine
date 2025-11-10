package main

// CommandHelp defines structured metadata for command help
type CommandHelp struct {
	Name        string   // Command name
	Category    string   // Command category for grouping
	Synopsis    string   // Brief one-line description
	Description string   // Longer description (optional)
	Usage       string   // Usage syntax
	Content     string   // Full help content (markdown)
	Related     []string // Related commands
	SeeAlso     []string // Additional topics/resources
}

// helpRegistry contains all command help metadata
var helpRegistry = map[string]*CommandHelp{
	"start": {
		Name:     "start",
		Category: "Service Management",
		Synopsis: "Start the shine service and enabled panels",
		Usage:    "shine start",
		Content:  startHelp,
		Related:  []string{"stop", "status", "reload"},
		SeeAlso:  []string{"Configuration: ~/.config/shine/prism.toml"},
	},
	"stop": {
		Name:     "stop",
		Category: "Service Management",
		Synopsis: "Stop all running panels gracefully",
		Usage:    "shine stop",
		Content:  stopHelp,
		Related:  []string{"start", "status"},
		SeeAlso:  []string{"IPC protocol", "Socket management"},
	},
	"status": {
		Name:     "status",
		Category: "Monitoring",
		Synopsis: "Display current state of panels and prisms",
		Usage:    "shine status",
		Content:  statusHelp,
		Related:  []string{"start", "stop", "logs"},
		SeeAlso:  []string{"MRU ordering", "Process states"},
	},
	"reload": {
		Name:     "reload",
		Category: "Configuration",
		Synopsis: "Reload configuration without restarting",
		Usage:    "shine reload",
		Content:  reloadHelp,
		Related:  []string{"start", "status"},
		SeeAlso:  []string{"SIGHUP behavior", "Hot-reload limitations"},
	},
	"logs": {
		Name:     "logs",
		Category: "Monitoring",
		Synopsis: "View service and panel log files",
		Usage:    "shine logs [panel-id]",
		Content:  logsHelp,
		Related:  []string{"status"},
		SeeAlso:  []string{"Log directory: ~/.local/share/shine/logs/"},
	},
}

// commandCategories defines the order and grouping of categories
var commandCategories = []struct {
	Name        string
	Description string
}{
	{"Service Management", "Starting, stopping, and managing the shine service"},
	{"Monitoring", "Viewing status, logs, and debugging"},
	{"Configuration", "Managing configuration and settings"},
}

// getCommandsByCategory returns commands grouped by category
func getCommandsByCategory() map[string][]*CommandHelp {
	result := make(map[string][]*CommandHelp)

	for _, cmd := range helpRegistry {
		result[cmd.Category] = append(result[cmd.Category], cmd)
	}

	return result
}

// getAllCommands returns all commands sorted by name
func getAllCommands() []*CommandHelp {
	commands := make([]*CommandHelp, 0, len(helpRegistry))
	for _, cmd := range helpRegistry {
		commands = append(commands, cmd)
	}
	return commands
}

// getCommandHelp returns help metadata for a specific command
func getCommandHelp(name string) (*CommandHelp, bool) {
	cmd, ok := helpRegistry[name]
	return cmd, ok
}

// getCommandNames returns all command names (useful for completion)
func getCommandNames() []string {
	names := make([]string, 0, len(helpRegistry))
	for name := range helpRegistry {
		names = append(names, name)
	}
	return names
}
