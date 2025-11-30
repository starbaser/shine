package help

import (
	"sort"
)

type Topic struct {
	Name     string
	Category string
	Synopsis string
	Usage    string
	Related  []string
	Content  string // Markdown content
}

type Registry struct {
	topics map[string]*Topic
}

func NewRegistry() *Registry {
	return &Registry{
		topics: make(map[string]*Topic),
	}
}

func (r *Registry) Register(topic *Topic) {
	r.topics[topic.Name] = topic
}

func (r *Registry) Get(name string) (*Topic, bool) {
	t, ok := r.topics[name]
	return t, ok
}

func (r *Registry) List() []*Topic {
	topics := make([]*Topic, 0, len(r.topics))
	for _, t := range r.topics {
		topics = append(topics, t)
	}
	sort.Slice(topics, func(i, j int) bool {
		return topics[i].Name < topics[j].Name
	})
	return topics
}

func (r *Registry) ListByCategory() map[string][]*Topic {
	byCategory := make(map[string][]*Topic)
	for _, t := range r.topics {
		byCategory[t.Category] = append(byCategory[t.Category], t)
	}
	for cat := range byCategory {
		topics := byCategory[cat]
		sort.Slice(topics, func(i, j int) bool {
			return topics[i].Name < topics[j].Name
		})
	}
	return byCategory
}

func (r *Registry) Categories() []string {
	seen := make(map[string]bool)
	for _, t := range r.topics {
		seen[t.Category] = true
	}
	cats := make([]string, 0, len(seen))
	for c := range seen {
		cats = append(cats, c)
	}
	sort.Strings(cats)
	return cats
}

func (r *Registry) Names() []string {
	names := make([]string, 0, len(r.topics))
	for name := range r.topics {
		names = append(names, name)
	}
	sort.Strings(names)
	return names
}
