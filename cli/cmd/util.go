package cmd

import (
	"encoding/json"
	"errors"
	"fmt"
	"math"
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

func makeTaskMapFile(m map[int]int32) error {
	homeDir, err := os.UserHomeDir()
	if err != nil {
		return fmt.Errorf("Error getting home directory: %w\n", err)
	}

	// Create a .prod directory in the home directory
	prodDir := filepath.Join(homeDir, ".prod")
	err = os.MkdirAll(prodDir, 0755) // Creates directory if it doesn't exist
	if err != nil {
		return fmt.Errorf("Error creating directory: %w\n", err)
	}

	// Save the file in the .prod directory
	filePath := filepath.Join(prodDir, "taskMap.json")

	// Create a file (or truncate it if it already exists)
	file, err := os.Create(filePath)
	if err != nil {
		return fmt.Errorf("Error creating file: %w\n", err)
	}
	defer file.Close()

	// Marshal map to JSON with indentation
	jsonData, err := json.MarshalIndent(m, "", "  ")
	if err != nil {
		return fmt.Errorf("Error marshaling to JSON: %w", err)
	}
	// Write JSON to file
	_, err = file.Write(jsonData)
	if err != nil {
		return fmt.Errorf("Error writing to file: %w", err)
	}

	return nil
}

func getTaskMap() (map[int]int32, error) {
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

func getTaskID(m map[int]int32, input string) (int32, error) {
	inputInt, err := strconv.Atoi(input)
	if err != nil {
		return 0, fmt.Errorf("Error converting input to int %w", err)
	}

	taskID, exits := m[inputInt]
	if !exits {
		return 0, fmt.Errorf("No tasks with %s iD", input)
	}
	return taskID, nil
}

func appendToMap(m map[int]int32, input string) (map[int]int32, error) {
	taskMap, err := getTaskMap()
	if err != nil {
		return nil, err
	}

	i, err := strconv.Atoi(input)
	if err != nil {
		return nil, errors.New("Invalid input type")
	}

	highestKey, err := findHighestKey(taskMap)
	if err != nil {
		taskMap[1] = int32(i)
		return taskMap, nil
	}
	taskMap[highestKey+1] = int32(i)

	return taskMap, nil
}

func removeFromMap(m map[int]int32, input string) error {
	inputInt, err := strconv.Atoi(input)
	if err != nil {
		return fmt.Errorf("Error converting input to int %w", err)
	}

	fmt.Println("inputInt:", inputInt)

	_, exits := m[inputInt]
	if !exits {
		return fmt.Errorf("No tasks with %s ID", input)
	}

	delete(m, inputInt)

	return nil
}

func findHighestKey(m map[int]int32) (int, error) {
	// Check if map is empty
	if len(m) == 0 {
		return 0, errors.New("map is empty")
	}

	// Start with the lowest possible int value
	highestKey := math.MinInt

	// Iterate through all keys
	for k := range m {
		if k > highestKey {
			highestKey = k
		}
	}

	return highestKey, nil
}
