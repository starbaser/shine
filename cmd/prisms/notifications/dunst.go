package main

import (
	"encoding/json"
	"fmt"
	"os/exec"
	"time"
)

// Notification represents a parsed Dunst notification from history.
type Notification struct {
	AppName   string
	Summary   string
	Body      string
	Timestamp time.Time
}

// DunstHistory represents the JSON structure from dunstctl history.
type DunstHistory struct {
	Type string                       `json:"type"`
	Data []map[string]DunstFieldValue `json:"data"`
}

// DunstFieldValue represents a typed field in the Dunst JSON output.
type DunstFieldValue struct {
	Type string      `json:"type"`
	Data interface{} `json:"data"`
}

// FetchHistory retrieves notification history from Dunst.
// Returns empty slice if Dunst is not running or no notifications exist.
func FetchHistory() ([]Notification, error) {
	cmd := exec.Command("dunstctl", "history")
	output, err := cmd.Output()
	if err != nil {
		return nil, fmt.Errorf("dunstctl history failed: %w", err)
	}

	var history DunstHistory
	if err := json.Unmarshal(output, &history); err != nil {
		return nil, fmt.Errorf("failed to parse history: %w", err)
	}

	notifications := make([]Notification, 0, len(history.Data))
	for _, item := range history.Data {
		notif := Notification{
			AppName:   getStringField(item, "appname"),
			Summary:   getStringField(item, "summary"),
			Body:      getStringField(item, "body"),
			Timestamp: getTimestampField(item, "timestamp"),
		}
		notifications = append(notifications, notif)
	}

	return notifications, nil
}

// DismissAll dismisses all visible notifications.
func DismissAll() error {
	cmd := exec.Command("dunstctl", "close-all")
	if err := cmd.Run(); err != nil {
		return fmt.Errorf("dunstctl close-all failed: %w", err)
	}
	return nil
}

// getStringField safely extracts a string value from a Dunst field.
func getStringField(fields map[string]DunstFieldValue, key string) string {
	if field, ok := fields[key]; ok {
		if str, ok := field.Data.(string); ok {
			return str
		}
	}
	return ""
}

// getTimestampField safely extracts a timestamp from a Dunst field.
// Dunst timestamps are in microseconds since boot (monotonic), not Unix time.
// We convert to an approximate time for relative display.
func getTimestampField(fields map[string]DunstFieldValue, key string) time.Time {
	if field, ok := fields[key]; ok {
		switch v := field.Data.(type) {
		case float64:
			// Convert microseconds to seconds and use as duration ago
			microseconds := int64(v)
			seconds := microseconds / 1_000_000
			return time.Now().Add(-time.Duration(seconds) * time.Second)
		}
	}
	return time.Time{}
}

// RelativeTime formats a timestamp as a relative time string (e.g., "2m ago").
func RelativeTime(t time.Time) string {
	if t.IsZero() {
		return "unknown"
	}

	duration := time.Since(t)
	switch {
	case duration < time.Minute:
		return "just now"
	case duration < time.Hour:
		mins := int(duration.Minutes())
		return fmt.Sprintf("%dm ago", mins)
	case duration < 24*time.Hour:
		hours := int(duration.Hours())
		return fmt.Sprintf("%dh ago", hours)
	default:
		days := int(duration.Hours() / 24)
		return fmt.Sprintf("%dd ago", days)
	}
}
