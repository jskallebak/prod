package cmd

import (
	"fmt"
	"time"
)

// formatDuration formats a duration in a human-readable format (e.g., "25m 30s")
func formatDuration(d time.Duration) string {
	d = d.Round(time.Second)
	h := d / time.Hour
	d -= h * time.Hour
	m := d / time.Minute
	d -= m * time.Minute
	s := d / time.Second

	if h > 0 {
		return fmt.Sprintf("%dh %dm %ds", h, m, s)
	}
	return fmt.Sprintf("%dm %ds", m, s)
}

// formatDurationSeconds formats seconds into a human-readable duration string
func formatDurationSeconds(seconds int64) string {
	hours := seconds / 3600
	minutes := (seconds % 3600) / 60
	remainingSeconds := seconds % 60

	if hours > 0 {
		return fmt.Sprintf("%dh %dm %ds", hours, minutes, remainingSeconds)
	} else if minutes > 0 {
		return fmt.Sprintf("%dm %ds", minutes, remainingSeconds)
	} else {
		return fmt.Sprintf("%ds", remainingSeconds)
	}
}

func coloredText(color Color, text string) string {
	return fmt.Sprintf("%s%s%s", string(color), text, string(ColorReset))
}
