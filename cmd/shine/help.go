package main

import (
	_ "embed"
	"fmt"
	"os"

	"github.com/charmbracelet/glamour"
)

//go:embed help/usage.md
var usageHelp string

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
