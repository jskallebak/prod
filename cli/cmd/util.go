package cmd

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"math"
	"os"
	"path/filepath"
	"sort"
	"strconv"
	"strings"
	"time"

	"github.com/jackc/pgx/v5/pgtype"
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
		return 0, fmt.Errorf("No tasks with ID %d\n", input)
	}
	return taskID, nil
}

func appendToMap(input int32) (map[int]int32, int, error) {
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

func removeFromMap(input int) error {
	m, err := getTaskMap()
	if err != nil {
		return fmt.Errorf("removeFromMap: getTaskMap: %v", err)
	}

	_, exits := m[int(input)]
	if !exits {
		return fmt.Errorf("No tasks with %d ID", input)
	}

	delete(m, int(input))
	fmt.Println(m)
	makeTaskMapFile(m)

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

// TaskNode represents a task and its subtasks in a tree structure
type TaskNode struct {
	Task     sqlc.Task   // The actual task data
	SubTasks []*TaskNode // Child tasks (subtasks)
	Index    int         // Display index for the task
}

// BuildTaskTree converts a flat list of tasks into a hierarchical tree structure
func BuildTaskTree(tasks []sqlc.Task) []*TaskNode {
	// Map to store tasks by their ID for quick lookup
	taskMap := make(map[int32]*TaskNode)

	// First pass: create nodes for all tasks
	for i, task := range tasks {
		taskMap[task.ID] = &TaskNode{
			Task:     task,
			SubTasks: []*TaskNode{},
			Index:    i + 1, // 1-based index for display
		}
	}

	// Root tasks (those without parents or with invalid parents)
	var rootTasks []*TaskNode

	// Second pass: build the tree structure by connecting parents and children
	for _, node := range taskMap {
		// If task has a valid dependent (parent task)
		if node.Task.Dependent.Valid {
			parentID := node.Task.Dependent.Int32
			if parent, exists := taskMap[parentID]; exists {
				// Add this task as a subtask of its parent
				parent.SubTasks = append(parent.SubTasks, node)
			} else {
				// If parent doesn't exist, treat as root task
				rootTasks = append(rootTasks, node)
			}
		} else {
			// Tasks without parents are root tasks
			rootTasks = append(rootTasks, node)
		}
	}

	return rootTasks
}

// ProcessList processes and displays tasks in a tree structure and returns a map of display index to task ID
func ProcessList(tasks []sqlc.Task, q *sqlc.Queries, u *sqlc.User) map[int]int32 {
	// Step 1: Build a map of task ID to task for quick lookups
	taskMap := make(map[int32]sqlc.Task)
	for _, task := range tasks {
		taskMap[task.ID] = task
	}

	// Step 2: Create a map to track which tasks have been displayed
	displayed := make(map[int32]bool)

	// Step 3: Create a map to track display index to actual task ID
	displayIndexToTaskID := make(map[int]int32)

	// Step 4: Get all root tasks (no dependency)
	var rootTasks []sqlc.Task
	for _, task := range tasks {
		if !task.Dependent.Valid {
			rootTasks = append(rootTasks, task)
		}
	}

	// Step 5: Sort root tasks by priority (low -> medium -> high)
	sort.Slice(rootTasks, func(i, j int) bool {
		// Convert priorities to numeric values
		prioI := getPriorityValueAscending(rootTasks[i].Priority)
		prioJ := getPriorityValueAscending(rootTasks[j].Priority)

		// Sort by priority first
		if prioI != prioJ {
			return prioI < prioJ // Low to High
		}

		// If priorities are equal, sort by ID for consistency
		return rootTasks[i].ID < rootTasks[j].ID
	})

	// Step 6: Track the index for display purposes
	displayIndex := 1

	// Step 7: Recursive function to display a task and all its descendants
	var displayTask func(task sqlc.Task, level int, prefix string, childPrefix string)
	displayTask = func(task sqlc.Task, level int, prefix string, childPrefix string) {
		// Skip if already displayed
		if displayed[task.ID] {
			return
		}

		// Mark as displayed
		displayed[task.ID] = true

		// Store the mapping of display index to actual task ID
		thisIndex := displayIndex
		displayIndexToTaskID[thisIndex] = task.ID

		// Display the task with its prefix
		status := "[ ]"
		if task.CompletedAt.Valid {
			status = "[✓]"
		}

		coloredDescription := coloredText(ColorGreen, task.Description)
		fmt.Printf("%s%s #%d %d %s\n", prefix, status, displayIndex, task.ID, coloredDescription)
		displayIndex++

		// Display task details with consistent formatting
		detailPrefix := childPrefix + "    "

		// Show task status
		if task.Status == "active" {
			fmt.Printf("%s🎯 Status: %s\n", detailPrefix, coloredText(ColorBrightCyan, task.Status))
		} else {
			fmt.Printf("%s🎯 Status: %s\n", detailPrefix, task.Status)
		}

		if task.CompletedAt.Valid {
			fmt.Printf("%s✅ Completed: %s\n", detailPrefix, task.CompletedAt.Time.Format("2006-01-02 15:04"))
		}

		if task.Priority.Valid {
			// Convert priority letter to full name
			priorityName := "Unknown"
			switch task.Priority.String {
			case "H":
				priorityName = coloredText(ColorRed, "High")
			case "M":
				priorityName = coloredText(ColorYellow, "Medium")
			case "L":
				priorityName = coloredText(ColorGreen, "Low")
			}
			fmt.Printf("%s🔄 Priority: %s\n", detailPrefix, priorityName)
		} else {
			fmt.Printf("%s🔄 Priority: --\n", detailPrefix)
		}

		if task.ProjectID.Valid {
			projectService := services.NewProjectService(q)
			project, err := projectService.GetProject(context.Background(), task.ProjectID.Int32, u.ID)
			if err == nil && project != nil {
				fmt.Printf("%s📁 Project: %s\n", detailPrefix, project.Name)
			} else {
				fmt.Printf("%s📁 Project: ID %d\n", detailPrefix, task.ProjectID.Int32)
			}
		}

		if task.StartDate.Valid {
			fmt.Printf("%s📅 Started at: %s\n", detailPrefix, task.StartDate.Time.Format("Mon, Jan 2, 2006"))
		}

		if task.DueDate.Valid {
			fmt.Printf("%s📅 Due: %s\n", detailPrefix, task.DueDate.Time.Format("Mon, Jan 2, 2006"))

			// Get today's date (without time component)
			now := time.Now()
			today := time.Date(now.Year(), now.Month(), now.Day(), 0, 0, 0, 0, now.Location())
			// tomorrow := today.AddDate(0, 0, 1)

			// Get due date without time component
			dueDate := time.Date(
				task.DueDate.Time.Year(),
				task.DueDate.Time.Month(),
				task.DueDate.Time.Day(),
				0, 0, 0, 0,
				task.DueDate.Time.Location(),
			)

			// Task is overdue if the due date is strictly before today
			if dueDate.Before(today) && task.Status != "completed" {
				text := coloredText(ColorRed, "Task overdue")
				fmt.Printf("%s❗ %s\n", detailPrefix, text)
			}
		}

		if len(task.Tags) > 0 {
			fmt.Printf("%s🏷️ Tags: %s\n", detailPrefix, strings.Join(task.Tags, ", "))
		}

		// Show dependency information
		if task.Dependent.Valid {
			// Find the display index for this task's parent
			parentTaskID := task.Dependent.Int32

			// Find display index for the parent task
			var parentDisplayIndex int
			for dispIdx, taskID := range displayIndexToTaskID {
				if taskID == parentTaskID {
					parentDisplayIndex = dispIdx
					break
				}
			}

			fmt.Printf("%s🔗 Dependent: %d\n", detailPrefix, parentDisplayIndex)
		}

		// Find all children of this task
		var childTasks []sqlc.Task
		for _, t := range tasks {
			if t.Dependent.Valid && t.Dependent.Int32 == task.ID {
				childTasks = append(childTasks, t)
			}
		}

		// Skip further processing if no children
		if len(childTasks) == 0 {
			return
		}

		// Sort children by priority (low -> medium -> high)
		sort.Slice(childTasks, func(i, j int) bool {
			// Convert priorities to numeric values
			prioI := getPriorityValueAscending(childTasks[i].Priority)
			prioJ := getPriorityValueAscending(childTasks[j].Priority)

			// Sort by priority first
			if prioI != prioJ {
				return prioI < prioJ // Low to High
			}

			// If priorities are equal, sort by ID for consistency
			return childTasks[i].ID < childTasks[j].ID
		})

		// Display children recursively
		for i, childTask := range childTasks {
			isLast := (i == len(childTasks)-1)

			// Determine the prefix for this child and its children
			var newPrefix, newChildPrefix string

			if isLast {
				// Last child gets the └── prefix
				newPrefix = childPrefix + "└── "
				// Its children get more indentation with spaces
				newChildPrefix = childPrefix + "    "
			} else {
				// Non-last children get the ├── prefix
				newPrefix = childPrefix + "├── "
				// Their children get more indentation with a vertical bar
				newChildPrefix = childPrefix + "│   "
			}

			displayTask(childTask, level+1, newPrefix, newChildPrefix)
		}
	}

	// Step 8: Display all root tasks and their hierarchies
	for _, rootTask := range rootTasks {
		// Root tasks have empty prefixes
		displayTask(rootTask, 0, "", "")
	}

	// Return the map of display indices to task IDs
	return displayIndexToTaskID
}

// Helper function to convert priority to a numeric value for ascending sorting (L->M->H)
func getPriorityValueAscending(priority pgtype.Text) int {
	if !priority.Valid {
		return 0 // No priority (lowest)
	}

	switch priority.String {
	case "L":
		return 1 // Low priority
	case "M":
		return 2 // Medium priority
	case "H":
		return 3 // High priority
	default:
		return 0 // Unknown priority (treat as lowest)
	}
}

// SortTaskList sorts tasks hierarchically with consistent ordering
func SortTaskList(taskList []sqlc.Task) []sqlc.Task {
	// First, build a map of tasks by ID for quick lookup
	taskMap := make(map[int32]sqlc.Task)
	for _, task := range taskList {
		taskMap[task.ID] = task
	}

	// Create a map to track the tree depth of each task
	depthMap := make(map[int32]int)

	// Calculate the depth of each task in the tree
	for _, task := range taskList {
		calculateDepth(task.ID, taskMap, depthMap)
	}

	// Create a sorted copy of the task list
	sortedTasks := make([]sqlc.Task, len(taskList))
	copy(sortedTasks, taskList)

	// Sort the tasks by:
	// 1. Tree path (parent-child hierarchy)
	// 2. Priority within siblings
	// 3. ID for consistency
	sort.Slice(sortedTasks, func(i, j int) bool {
		taskI := sortedTasks[i]
		taskJ := sortedTasks[j]

		// Get the complete path to each task
		pathI := getTaskPath(taskI.ID, taskMap)
		pathJ := getTaskPath(taskJ.ID, taskMap)

		// Compare paths element by element
		minLen := len(pathI)
		if len(pathJ) < minLen {
			minLen = len(pathJ)
		}

		for k := 0; k < minLen; k++ {
			if pathI[k] != pathJ[k] {
				// If different parent, sort by parent ID
				return pathI[k] < pathJ[k]
			}
		}

		// If one path is a prefix of the other, the shorter one comes first
		if len(pathI) != len(pathJ) {
			return len(pathI) < len(pathJ)
		}

		// If tasks are siblings (same path), sort by priority
		priorityI := getPriorityValue(taskI.Priority)
		priorityJ := getPriorityValue(taskJ.Priority)

		if priorityI != priorityJ {
			return priorityI > priorityJ // Higher priority first
		}

		// If priority is the same, sort by ID for consistency
		return taskI.ID < taskJ.ID
	})

	return sortedTasks
}

// calculateDepth calculates the depth of a task in the tree recursively
func calculateDepth(taskID int32, taskMap map[int32]sqlc.Task, depthMap map[int32]int) int {
	// If we've already calculated the depth, return it
	if depth, found := depthMap[taskID]; found {
		return depth
	}

	task, exists := taskMap[taskID]
	if !exists {
		return 0
	}

	// Root tasks have depth 0
	if !task.Dependent.Valid {
		depthMap[taskID] = 0
		return 0
	}

	// The depth is one more than the parent's depth
	parentID := task.Dependent.Int32
	parentDepth := calculateDepth(parentID, taskMap, depthMap)
	depth := parentDepth + 1
	depthMap[taskID] = depth

	return depth
}

// getTaskPath returns the path from root to the given task as a slice of task IDs
func getTaskPath(taskID int32, taskMap map[int32]sqlc.Task) []int32 {
	var path []int32

	// Start with the current task
	currentID := taskID

	// Build path from task to root
	for {
		task, exists := taskMap[currentID]
		if !exists {
			break
		}

		// Add the current task to the beginning of the path
		path = append([]int32{currentID}, path...)

		// If this is a root task, we're done
		if !task.Dependent.Valid {
			break
		}

		// Move to the parent
		currentID = task.Dependent.Int32
	}

	return path
}

// Helper function to convert priority to a numeric value for sorting
func getPriorityValue(priority pgtype.Text) int {
	if !priority.Valid {
		return 0 // No priority
	}

	switch priority.String {
	case "H":
		return 3 // High priority
	case "M":
		return 2 // Medium priority
	case "L":
		return 1 // Low priority
	default:
		return 0 // Unknown priority
	}
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

func ConfirmCmd(ctx context.Context, taskID, userID int32, action ActionType, ts *services.TaskService) error {
	// Get task info for confirmation
	task, err := ts.GetTask(ctx, taskID, userID)
	if err != nil {
		return fmt.Errorf("Error: Failed to find task with ID %d: %v", taskID, err)
	}

	taskMap, err := getTaskMap()
	if err != nil {
		return fmt.Errorf("ConfirmCmd: getTaskMap: %v", err)
	}

	reverseMap := ReverseMap(taskMap)
	taskID = int32(reverseMap[taskID])

	// Confirm deletion unless --yes flag is used
	fmt.Printf("You are about to %s task %d: \"%s\"\n", action, taskID, task.Description)
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

func RecursiveSubtasks(
	ctx context.Context,
	userID int32,
	taskID int32,
	ts *services.TaskService,
	action ActionType,
	input int,
	executeFunc func(context.Context, int32, int32) (*sqlc.Task, error),
) error {
	subtasks, err := ts.GetDependent(ctx, userID, taskID)
	if err != nil {
		return fmt.Errorf("RecursiveSubtasks: Error with GetDependent: %v", err)
	}

	if len(subtasks) != 0 {
		fmt.Printf("The task have subtask(s), confirm to %s\n", action)
		for _, st := range subtasks {
			err = ConfirmCmd(ctx, st.ID, userID, action, ts)
			if err != nil {
				return fmt.Errorf("RecursiveSubtasks: Error with ConfirmCmd: %v", err)
			}

			_, err = executeFunc(ctx, st.ID, userID)
			if err != nil {
				return fmt.Errorf("RecursiveSubtasks: Error with DeleteTask: %v", err)
			}
			fmt.Printf("Task %d deleted successfully\n", input)
		}
	}

	return nil
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
