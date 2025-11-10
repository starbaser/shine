package main

import (
	_ "embed"
	"encoding/json"
	"fmt"
	"os"

	"github.com/charmbracelet/glamour"
)

//go:embed help/usage.md
var usageHelp string

//go:embed help/ipc.md
var ipcHelp string

//go:embed help/signals.md
var signalsHelp string

var renderer *glamour.TermRenderer

func init() {
	var err error
	renderer, err = glamour.NewTermRenderer(
		glamour.WithAutoStyle(),
		glamour.WithWordWrap(100),
	)
	if err != nil {
		// Fallback: renderer stays nil, will use plain text
		fmt.Fprintf(os.Stderr, "Warning: Failed to initialize help renderer: %v\n", err)
	}
}

func renderHelp(markdown string) string {
	if renderer == nil {
		return markdown
	}

	out, err := renderer.Render(markdown)
	if err != nil {
		return markdown
	}
	return out
}

// showHelp displays help for a specific topic or general usage
func showHelp(topic string) {
	var content string

	// Check for special topics
	switch topic {
	case "topics":
		content = helpTopics()
	case "list":
		content = helpList()
	case "categories":
		content = helpCategories()
	case "", "help":
		content = usageHelp
	default:
		// Look up topic in registry
		if cmd, ok := helpRegistry[topic]; ok {
			content = cmd.Content
		} else {
			// Unknown topic, show main help
			content = usageHelp
		}
	}

	fmt.Print(renderHelp(content))
}

// helpTopics returns a list of available help topics
func helpTopics() string {
	return `# prismctl - Available Help Topics

Use **prismctl help <topic>** to view detailed help for a specific topic.

## Available Topics

- **usage** - General usage and command-line interface
- **ipc** - IPC protocol and command reference
- **signals** - Signal handling and process management

## Examples

View help for IPC commands:
` + "```bash" + `
prismctl help ipc
` + "```" + `

View signal handling help:
` + "```bash" + `
prismctl help signals
` + "```" + `

## Getting Started

For general usage information:
` + "```bash" + `
prismctl --help
` + "```" + `

## Files

Logs: ` + "`~/.local/share/shine/logs/prismctl.log`" + `
Sockets: ` + "`/run/user/{uid}/shine/prism-*.sock`" + `

## Documentation

For complete documentation, see: https://github.com/starbased-co/shine
`
}

// helpList generates a compact list of all topics with brief descriptions
func helpList() string {
	content := "# prismctl - Help Topics\n\n"
	content += "All available help topics with brief descriptions.\n\n"
	content += "## Topics\n\n"

	// Get topics by category
	byCategory := getTopicsByCategory()

	// Display by category in defined order
	for _, cat := range topicCategories {
		topics := byCategory[cat.Name]
		if len(topics) == 0 {
			continue
		}

		content += fmt.Sprintf("### %s\n\n", cat.Name)
		content += fmt.Sprintf("*%s*\n\n", cat.Description)

		for _, topic := range topics {
			content += fmt.Sprintf("**`%s`** - %s\n", topic.Name, topic.Synopsis)
			if len(topic.Related) > 0 {
				content += fmt.Sprintf("  Related: %s\n", joinTopics(topic.Related))
			}
			content += "\n"
		}
	}

	content += "## Getting More Help\n\n"
	content += "Use `prismctl help <topic>` to view detailed help for a specific topic.\n"

	return content
}

// helpCategories shows topics organized by category with descriptions
func helpCategories() string {
	content := "# prismctl - Help Categories\n\n"
	content += "Help topics organized by functional category.\n\n"

	// Get topics by category
	byCategory := getTopicsByCategory()

	// Display by category in defined order
	for _, cat := range topicCategories {
		topics := byCategory[cat.Name]
		if len(topics) == 0 {
			continue
		}

		content += fmt.Sprintf("## %s\n\n", cat.Name)
		content += fmt.Sprintf("%s\n\n", cat.Description)

		for _, topic := range topics {
			content += fmt.Sprintf("- **%s** - %s\n", topic.Name, topic.Synopsis)
		}
		content += "\n"
	}

	content += "## Usage\n\n"
	content += "```bash\n"
	content += "# View all topics\n"
	content += "prismctl help list\n\n"
	content += "# View detailed help for a topic\n"
	content += "prismctl help <topic>\n"
	content += "```\n"

	return content
}

// joinTopics joins topic names with commas and backticks
func joinTopics(topics []string) string {
	if len(topics) == 0 {
		return ""
	}

	result := ""
	for i, topic := range topics {
		if i > 0 {
			result += ", "
		}
		result += fmt.Sprintf("`%s`", topic)
	}
	return result
}

// helpJSON outputs help metadata in JSON format (for machine consumption)
func helpJSON(topic string) error {
	var output interface{}

	if topic == "" {
		// Output all topics metadata
		output = getAllTopics()
	} else if topic == "categories" {
		// Output categories structure
		output = topicCategories
	} else if topic == "names" {
		// Output just topic names (for completion)
		output = getTopicNames()
	} else {
		// Output specific topic metadata
		if cmd, ok := getTopicHelp(topic); ok {
			// Omit the full markdown content for JSON output
			type TopicMeta struct {
				Name     string   `json:"name"`
				Category string   `json:"category"`
				Synopsis string   `json:"synopsis"`
				Usage    string   `json:"usage"`
				Related  []string `json:"related"`
				SeeAlso  []string `json:"see_also"`
			}
			output = TopicMeta{
				Name:     cmd.Name,
				Category: cmd.Category,
				Synopsis: cmd.Synopsis,
				Usage:    cmd.Usage,
				Related:  cmd.Related,
				SeeAlso:  cmd.SeeAlso,
			}
		} else {
			return fmt.Errorf("unknown topic: %s", topic)
		}
	}

	encoder := json.NewEncoder(os.Stdout)
	encoder.SetIndent("", "  ")
	return encoder.Encode(output)
}
