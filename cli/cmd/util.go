package cmd

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strconv"
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

func loadTaskMap() (map[int]int32, error) {
	homeDir, err := os.UserHomeDir()
	if err != nil {
		return nil, fmt.Errorf("Error getting hom edirectory: %w\n", err)
	}

	filePath := filepath.Join(homeDir, ".prod", "taskMap.json")
	_, err = os.Stat(filePath)
	if err != nil {
		return nil, fmt.Errorf("taskmap.json doesnt exits")
	}
	bytes, err := os.ReadFile(filePath)
	if err != nil {
		return nil, fmt.Errorf("Error reading the json file")
	}

	taskMap := map[int]int32{}

	err = json.Unmarshal(bytes, &taskMap)
	if err != nil {
		return nil, fmt.Errorf("Error unmarshal the json file")
	}

	return taskMap, nil
}

func getTaskID(input string) (int32, error) {
	taskMap, err := loadTaskMap()
	if err != nil {
		return 0, err
	}

	inputInt, err := strconv.Atoi(input)
	if err != nil {
		return 0, fmt.Errorf("Error converting input to int %w", err)
	}

	taskID, exits := taskMap[inputInt]
	if !exits {
		return 0, fmt.Errorf("No tasks with %s iD", input)
	}
	return taskID, nil
}
