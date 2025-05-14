package cmd

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"math"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"time"

	"github.com/jskallebak/prod/internal/db/sqlc"
	"github.com/jskallebak/prod/internal/services"
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

func getID(mapFunc func() (map[int]int32, error), input int) (int32, error) {
	m, err := mapFunc()
	if err != nil {
		return 0, err
	}

	taskID, exits := m[input]
	if !exits {
		return 0, fmt.Errorf("No tasks with %s iD", input)
	}
	return taskID, nil
}

func appendToMap(m map[int]int32, input int32) (map[int]int32, int, error) {
	taskMap, err := getTaskMap()
	if err != nil {
		return nil, 0, err
	}

	highestKey, err := findHighestKey(taskMap)
	index := 1
	if err != nil {
		taskMap[index] = input
		return taskMap, 0, nil
	}
	index = highestKey + 1
	taskMap[index] = input

	return taskMap, index, nil
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

func ProcessList(list []sqlc.Task, q *sqlc.Queries, u *sqlc.User) {
	indent := ""

	for i, item := range list {
		if item.Dependent.Valid {
			fmt.Println("    ↳")
			indent = "        "
		} else {
			indent = ""
		}

		// Show task status with checkbox
		status := "[ ]"
		if item.CompletedAt.Valid {
			status = "[✓]"
		}
		coloredDescription := coloredText(ColorGreen, item.Description)
		fmt.Printf("%s%s #%d %d %s\n", indent, status, i+1, item.ID, coloredDescription)

		// Show task details with emojis and consistent formatting
		if item.Status == "active" {
			fmt.Printf("    %s🎯 Status: %s\n", indent, coloredText(ColorBrightCyan, item.Status)) // Target/goal
		} else {
			fmt.Printf("    %s🎯 Status: %s\n", indent, item.Status) // Target/goal
		}
		if item.CompletedAt.Valid {
			fmt.Printf("    %s✅ Completed: %s\n", indent, item.CompletedAt.Time.Format("2006-01-02 15:04"))
		}
		if item.Priority.Valid {
			// Convert priority letter to full name
			priorityName := "Unknown"
			switch item.Priority.String {
			case "H":
				priorityName = coloredText(ColorRed, "High")
			case "M":
				priorityName = coloredText(ColorYellow, "Medium")
			case "L":
				priorityName = coloredText(ColorGreen, "Low")
			}
			fmt.Printf("    %s🔄 Priority: %s\n", indent, priorityName)
		} else {
			fmt.Printf("    %s🔄 Priority: --\n", indent)
		}
		if item.ProjectID.Valid {
			// Get project name instead of just showing the ID
			projectService := services.NewProjectService(q)
			project, err := projectService.GetProject(context.Background(), item.ProjectID.Int32, u.ID)
			if err == nil && project != nil {
				fmt.Printf("    %s📁 Project: %s\n", indent, project.Name)
			} else {
				fmt.Printf("    %s📁 Project: ID %d\n", indent, item.ProjectID.Int32)
			}
		}
		// else {
		// 	fmt.Printf("    📁\tProject: --\n")
		// }
		if item.StartDate.Valid {
			fmt.Printf("    %s📅 Started at: %s\n", indent, item.StartDate.Time.Format("Mon, Jan 2, 2006"))
		}
		// else {
		// 	fmt.Printf("\t📅 Started at: --\n")
		// }
		if item.DueDate.Valid {
			fmt.Printf("    %s📅 Due: %s\n", indent, item.DueDate.Time.Format("Mon, Jan 2, 2006"))
		}
		// else {
		// 	fmt.Printf("\t📅 Due: --\n")
		// }
		if len(item.Tags) > 0 {
			fmt.Printf("    %s🏷️ Tags: %s\n", indent, strings.Join(item.Tags, ", "))
		}
		// else {
		// 	fmt.Printf("\t🏷️ Tags: --\n")
		// }

		taskMap := MakeTaskMap(list)
		reverseMap := ReverseMap(taskMap)

		if true {
			index, exits := reverseMap[item.Dependent.Int32]
			if exits {
				fmt.Printf("    %s🔗 Dependent: %d\n", indent, index)
			}
		}
	}
}

func SortTaskList(taskList []sqlc.Task) []sqlc.Task {
	subtaskMap := MakeSubtaskMap(taskList)
	sortedList := []sqlc.Task{}

	counter := 0

	for _, task := range taskList {
		if task.Dependent.Valid {
			// fmt.Println(i)
			continue
		}

		sortedList = append(sortedList, task)
		// fmt.Println(i)

		// Make a dict with subtasks, if the current task are in the dict, append them
		subtasks, exits := subtaskMap[task.ID]
		if exits {
			for _, subtask := range subtasks {
				task, err := findTask(taskList, int32(subtask))
				if err != nil {
					continue
				}

				sortedList = append(sortedList, task)
			}
		}
		counter += 1

	}
	return sortedList
}

func MakeTaskMap(taskList []sqlc.Task) map[int]int32 {
	taskMap := make(map[int]int32)
	for i, task := range taskList {
		taskMap[i+1] = task.ID
	}
	return taskMap
}

func ReverseMap[K comparable, V comparable](m map[K]V) map[V]K {
	reversed := make(map[V]K, len(m))
	for key, value := range m {
		reversed[value] = key
	}
	return reversed
}

func MakeSubtaskMap(taskList []sqlc.Task) map[int32][]int32 {
	subtaskMap := make(map[int32][]int32)
	for _, task := range taskList {
		if task.Dependent.Valid {
			subtaskMap[task.Dependent.Int32] = append(subtaskMap[task.Dependent.Int32], task.ID)
		}
	}
	return subtaskMap
}

func findTask(taskList []sqlc.Task, taskID int32) (sqlc.Task, error) {
	for _, task := range taskList {
		if task.ID == taskID {
			return task, nil
		}
	}
	return sqlc.Task{}, errors.New("could not find the task in list")
}

func ConfirmCmd(ctx context.Context, input int, taskID, userID int32, name string, ts *services.TaskService) error {
	// Get task info for confirmation
	task, err := ts.GetTask(ctx, int32(taskID), userID)
	if err != nil {
		return fmt.Errorf("Error: Failed to find task with ID %d: %v", taskID, err)
	}

	// Confirm deletion unless --yes flag is used
	fmt.Printf("You are about to %s task %d: \"%s\"\n", name, input, task.Description)
	fmt.Print("Are you sure? (y/N): ")
	var confirmation string
	fmt.Scanln(&confirmation)
	if confirmation != "y" && confirmation != "Y" {
		return errors.New("complete cancelled")
	}
	return nil

}

// ParseArgs parses arguments in formats like "1", "1-3", "1,2,4", "1 2 4" or combinations
// Returns a slice of integers containing all the specified values
func ParseArgs(args []string) ([]int, error) {
	argStr := strings.Join(args, " ")
	if argStr == "" {
		return nil, fmt.Errorf("empty argument string")
	}

	var result []int

	// First split by commas
	commaSeparated := strings.Split(argStr, ",")

	for _, commaPart := range commaSeparated {
		// Then split each comma-separated part by spaces
		spaceParts := strings.Fields(commaPart)

		for _, part := range spaceParts {
			part = strings.TrimSpace(part)
			if part == "" {
				continue
			}

			// Check if it's a range (e.g., "1-3")
			if strings.Contains(part, "-") {
				rangeParts := strings.Split(part, "-")
				if len(rangeParts) != 2 {
					return nil, fmt.Errorf("invalid range format: %s", part)
				}

				start, err := strconv.Atoi(strings.TrimSpace(rangeParts[0]))
				if err != nil {
					return nil, fmt.Errorf("invalid range start: %s", rangeParts[0])
				}

				end, err := strconv.Atoi(strings.TrimSpace(rangeParts[1]))
				if err != nil {
					return nil, fmt.Errorf("invalid range end: %s", rangeParts[1])
				}

				if end < start {
					return nil, fmt.Errorf("range end cannot be less than start: %s", part)
				}

				for i := start; i <= end; i++ {
					result = append(result, i)
				}
			} else {
				// Single number
				num, err := strconv.Atoi(part)
				if err != nil {
					return nil, fmt.Errorf("invalid number: %s", part)
				}
				result = append(result, num)
			}
		}
	}

	return result, nil
}

func Input2Int(input string) (int, error) {
	i, err := strconv.Atoi(input)
	if err != nil {
		return 0, err
	}
	return i, nil
}

// func parseArgs(args []string) []string {
// 	var parsedArgs []string
// 	pattern := regexp.MustCompile(`\d+-\d+`)
//
// 	for _, arg := range args {
// 		matches := pattern.FindAllString(arg, -1)
// 		for _, match := range matches {
// 			splitMatch := strings.Split(match, "-")
//
// 			for _, part := range sort.Slice(splitMatch, func(i, j int) bool {
// 				return i > j
// 			}) {
// 				fmt.Println(part)
// 			}
//
// 		}
// 	}
//
// 	return parsedArgs
