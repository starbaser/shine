package main

import (
	"bufio"
	"os"
	"path/filepath"
	"strings"
)

// DesktopEntry represents a parsed XDG desktop entry.
type DesktopEntry struct {
	Name    string
	Exec    string
	Icon    string
	Comment string
	Path    string
}

// LoadDesktopEntries scans XDG application directories and parses .desktop files.
func LoadDesktopEntries() ([]DesktopEntry, error) {
	var entries []DesktopEntry

	homeDir, err := os.UserHomeDir()
	if err != nil {
		return nil, err
	}

	searchPaths := []string{
		"/usr/share/applications/",
		filepath.Join(homeDir, ".local/share/applications/"),
	}

	for _, dir := range searchPaths {
		dirEntries, err := os.ReadDir(dir)
		if err != nil {
			continue
		}

		for _, entry := range dirEntries {
			if entry.IsDir() || !strings.HasSuffix(entry.Name(), ".desktop") {
				continue
			}

			fullPath := filepath.Join(dir, entry.Name())
			desktopEntry, err := parseDesktopEntry(fullPath)
			if err != nil {
				continue
			}

			if desktopEntry != nil {
				entries = append(entries, *desktopEntry)
			}
		}
	}

	return entries, nil
}

// parseDesktopEntry parses a single .desktop file.
func parseDesktopEntry(path string) (*DesktopEntry, error) {
	file, err := os.Open(path)
	if err != nil {
		return nil, err
	}
	defer file.Close()

	var (
		entry       = &DesktopEntry{Path: path}
		inSection   = false
		noDisplay   = false
		hasName     = false
		hasExec     = false
	)

	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())

		if line == "" || strings.HasPrefix(line, "#") {
			continue
		}

		if line == "[Desktop Entry]" {
			inSection = true
			continue
		}

		if strings.HasPrefix(line, "[") && line != "[Desktop Entry]" {
			break
		}

		if !inSection {
			continue
		}

		key, value, found := strings.Cut(line, "=")
		if !found {
			continue
		}

		key = strings.TrimSpace(key)
		value = strings.TrimSpace(value)

		switch key {
		case "Name":
			entry.Name = value
			hasName = true
		case "Exec":
			entry.Exec = cleanExecField(value)
			hasExec = true
		case "Icon":
			entry.Icon = value
		case "Comment":
			entry.Comment = value
		case "NoDisplay":
			noDisplay = strings.ToLower(value) == "true"
		case "Type":
			if value != "Application" {
				return nil, nil
			}
		}
	}

	if err := scanner.Err(); err != nil {
		return nil, err
	}

	if noDisplay || !hasName || !hasExec {
		return nil, nil
	}

	return entry, nil
}

// cleanExecField removes field codes (%u, %U, %f, %F, etc.) from Exec field.
func cleanExecField(exec string) string {
	parts := strings.Fields(exec)
	var cleaned []string

	for _, part := range parts {
		if strings.HasPrefix(part, "%") {
			continue
		}
		cleaned = append(cleaned, part)
	}

	return strings.Join(cleaned, " ")
}

// FilterEntries performs case-insensitive substring matching on app names.
func FilterEntries(entries []DesktopEntry, query string) []DesktopEntry {
	if query == "" {
		return entries
	}

	query = strings.ToLower(query)
	var filtered []DesktopEntry

	for _, entry := range entries {
		if strings.Contains(strings.ToLower(entry.Name), query) {
			filtered = append(filtered, entry)
		}
	}

	return filtered
}
