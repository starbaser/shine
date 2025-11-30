package help

import (
	"fmt"
	"strings"

	"github.com/charmbracelet/glamour"
	"github.com/charmbracelet/lipgloss"
)

type RenderOptions struct {
	Width int
	Style string
}

func (r *Registry) Render(topicName string, opts RenderOptions) (string, error) {
	if topicName == "" {
		return r.renderTopicList(opts), nil
	}

	switch topicName {
	case "list":
		return r.renderTopicList(opts), nil
	case "categories":
		return r.renderCategories(opts), nil
	}

	topic, ok := r.Get(topicName)
	if !ok {
		return "", fmt.Errorf("topic %q not found", topicName)
	}

	renderer, err := glamour.NewTermRenderer(
		glamour.WithAutoStyle(),
		glamour.WithWordWrap(opts.Width),
	)
	if err != nil {
		return "", fmt.Errorf("failed to create renderer: %w", err)
	}

	rendered, err := renderer.Render(topic.Content)
	if err != nil {
		return "", fmt.Errorf("failed to render content: %w", err)
	}

	return rendered, nil
}

func (r *Registry) renderTopicList(opts RenderOptions) string {
	var b strings.Builder

	titleStyle := lipgloss.NewStyle().Bold(true).Foreground(lipgloss.Color("12"))
	categoryStyle := lipgloss.NewStyle().Bold(true).Foreground(lipgloss.Color("14"))

	b.WriteString(titleStyle.Render("Available Help Topics"))
	b.WriteString("\n\n")

	byCategory := r.ListByCategory()
	categories := r.Categories()

	for _, cat := range categories {
		b.WriteString(categoryStyle.Render(cat))
		b.WriteString("\n")

		for _, topic := range byCategory[cat] {
			b.WriteString(fmt.Sprintf("  %-15s  %s\n", topic.Name, topic.Synopsis))
		}
		b.WriteString("\n")
	}

	return b.String()
}

func (r *Registry) renderCategories(opts RenderOptions) string {
	var b strings.Builder

	titleStyle := lipgloss.NewStyle().Bold(true).Foreground(lipgloss.Color("12"))

	b.WriteString(titleStyle.Render("Help Categories"))
	b.WriteString("\n\n")

	byCategory := r.ListByCategory()
	categories := r.Categories()

	for _, cat := range categories {
		topics := byCategory[cat]
		b.WriteString(fmt.Sprintf("  %-20s  (%d topics)\n", cat, len(topics)))
	}

	return b.String()
}
