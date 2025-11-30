package main

import (
	"embed"
	"fmt"
	"os"

	"github.com/starbased-co/shine/pkg/help"
)

//go:embed help/*.md
var helpFiles embed.FS

var registry = help.NewRegistry()

func init() {
	// Read main usage help
	usageContent, err := helpFiles.ReadFile("help/usage.md")
	if err == nil {
		registry.Register(&help.Topic{
			Name:     "",
			Category: "General",
			Synopsis: "General usage information",
			Content:  string(usageContent),
		})
	}

	// Register all topics from registry
	for _, meta := range helpRegistry {
		registry.Register(&help.Topic{
			Name:     meta.Name,
			Category: meta.Category,
			Synopsis: meta.Synopsis,
			Usage:    meta.Usage,
			Related:  meta.Related,
			Content:  meta.Content,
		})
	}
}

// showHelp displays help for a specific topic or general usage
func showHelp(topic string) {
	output, err := registry.Render(topic, help.RenderOptions{Width: 100})
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error: %s\n", err)
		os.Exit(1)
	}
	fmt.Print(output)
}
