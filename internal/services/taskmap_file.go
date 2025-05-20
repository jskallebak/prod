package services

import (
	"encoding/json"
	"errors"
	"fmt"
	"math"
	"os"
	"path/filepath"
)

func MakeTaskMapFile(m map[int]int32) error {
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

func GetTaskMap() (map[int]int32, error) {
	homeDir, err := os.UserHomeDir()
	if err != nil {
		return nil, fmt.Errorf("Error getting home directory: %w\n", err)
	}

	filePath := filepath.Join(homeDir, ".prod", "taskMap.json")
	_, err = os.Stat(filePath)
	if err != nil {
		return nil, fmt.Errorf("taskmap.json doesnt exist")
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

func GetID(mapFunc func() (map[int]int32, error), input int) (int32, error) {
	m, err := mapFunc()
	if err != nil {
		return 0, err
	}
	taskID, exists := m[input]
	if !exists {
		return 0, fmt.Errorf("No tasks with ID %d\n", input)
	}
	return taskID, nil
}

func AppendToMap(input int32) (map[int]int32, int, error) {
	taskMap, err := GetTaskMap()
	if err != nil {
		return nil, 0, err
	}

	highestKey, err := FindHighestKey(taskMap)
	index := 1
	if err != nil {
		taskMap[index] = input
		return taskMap, 0, nil
	}
	index = highestKey + 1
	taskMap[index] = input

	return taskMap, index, nil
}

func RemoveFromMap(input int) error {
	m, err := GetTaskMap()
	if err != nil {
		return fmt.Errorf("removeFromMap: GetTaskMap: %v", err)
	}

	_, exists := m[int(input)]
	if !exists {
		return fmt.Errorf("No tasks with %d ID", input)
	}

	delete(m, int(input))
	MakeTaskMapFile(m)

	return nil
}

func FindHighestKey(m map[int]int32) (int, error) {
	if len(m) == 0 {
		return 0, errors.New("map is empty")
	}

	highestKey := math.MinInt
	for k := range m {
		if k > highestKey {
			highestKey = k
		}
	}

	return highestKey, nil
}
