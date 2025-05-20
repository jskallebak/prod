package services

import (
	"context"
	"errors"
	"fmt"
	"sort"
	"strings"
	"time"

	"github.com/jackc/pgx/v5/pgtype"
	"github.com/jskallebak/prod/internal/db/sqlc"
	"github.com/jskallebak/prod/internal/util"
)

type TaskNode struct {
	Task     sqlc.Task   // The actual task data
	SubTasks []*TaskNode // Child tasks (subtasks)
	Index    int         // Display index for the task
}

func BuildTaskTree(tasks []sqlc.Task) []*TaskNode {
	taskMap := make(map[int32]*TaskNode)
	for i, task := range tasks {
		taskMap[task.ID] = &TaskNode{
			Task:     task,
			SubTasks: []*TaskNode{},
			Index:    i + 1,
		}
	}
	var rootTasks []*TaskNode
	for _, node := range taskMap {
		if node.Task.Dependent.Valid {
			parentID := node.Task.Dependent.Int32
			if parent, exists := taskMap[parentID]; exists {
				parent.SubTasks = append(parent.SubTasks, node)
			} else {
				rootTasks = append(rootTasks, node)
			}
		} else {
			rootTasks = append(rootTasks, node)
		}
	}
	return rootTasks
}

func ProcessList(tasks []sqlc.Task, q *sqlc.Queries, u *sqlc.User) map[int]int32 {
	taskMap := make(map[int32]sqlc.Task)
	for _, task := range tasks {
		taskMap[task.ID] = task
	}
	displayed := make(map[int32]bool)
	displayIndexToTaskID := make(map[int]int32)
	var rootTasks []sqlc.Task
	for _, task := range tasks {
		if !task.Dependent.Valid {
			rootTasks = append(rootTasks, task)
		}
	}
	sort.Slice(rootTasks, func(i, j int) bool {
		prioI := getPriorityValueAscending(rootTasks[i].Priority)
		prioJ := getPriorityValueAscending(rootTasks[j].Priority)
		if prioI != prioJ {
			return prioI < prioJ
		}
		return rootTasks[i].ID < rootTasks[j].ID
	})
	displayIndex := 1
	var displayTask func(task sqlc.Task, level int, prefix string, childPrefix string)
	displayTask = func(task sqlc.Task, level int, prefix string, childPrefix string) {
		if displayed[task.ID] {
			return
		}
		displayed[task.ID] = true
		thisIndex := displayIndex
		displayIndexToTaskID[thisIndex] = task.ID
		status := "[ ]"
		if task.CompletedAt.Valid {
			status = "[✓]"
		}
		coloredDescription := util.ColoredText(util.ColorGreen, task.Description)
		fmt.Printf("%s%s #%d %d %s\n", prefix, status, displayIndex, task.ID, coloredDescription)
		displayIndex++
		detailPrefix := childPrefix + "    "
		if task.Status == "active" {
			fmt.Printf("%s🎯 Status: %s\n", detailPrefix, task.Status)
		} else {
			fmt.Printf("%s🎯 Status: %s\n", detailPrefix, task.Status)
		}
		if task.CompletedAt.Valid {
			fmt.Printf("%s✅ Completed: %s\n", detailPrefix, task.CompletedAt.Time.Format("2006-01-02 15:04"))
		}
		if task.Priority.Valid {
			priorityName := "Unknown"
			switch task.Priority.String {
			case "H":
				priorityName = util.ColoredText(util.ColorRed, "High")
			case "M":
				priorityName = util.ColoredText(util.ColorYellow, "Medium")
			case "L":
				priorityName = util.ColoredText(util.ColorGreen, "Low")
			}
			fmt.Printf("%s🔄 Priority: %s\n", detailPrefix, priorityName)
		} else {
			fmt.Printf("%s🔄 Priority: --\n", detailPrefix)
		}
		if task.ProjectID.Valid {
			projectService := NewProjectService(q)
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
			now := time.Now()
			today := time.Date(now.Year(), now.Month(), now.Day(), 0, 0, 0, 0, now.Location())
			dueDate := time.Date(
				task.DueDate.Time.Year(),
				task.DueDate.Time.Month(),
				task.DueDate.Time.Day(),
				0, 0, 0, 0,
				task.DueDate.Time.Location(),
			)
			if dueDate.Before(today) && task.Status != "completed" {
				fmt.Printf("%s❗ %s\n", detailPrefix, util.ColoredText(util.ColorRed, "Task overdue"))
			}
		}
		if len(task.Tags) > 0 {
			fmt.Printf("%s🏷️ Tags: %s\n", detailPrefix, strings.Join(task.Tags, ", "))
		}
		if task.Dependent.Valid {
			parentTaskID := task.Dependent.Int32
			var parentDisplayIndex int
			for dispIdx, taskID := range displayIndexToTaskID {
				if taskID == parentTaskID {
					parentDisplayIndex = dispIdx
					break
				}
			}
			fmt.Printf("%s🔗 Dependent: %d\n", detailPrefix, parentDisplayIndex)
		}
		var childTasks []sqlc.Task
		for _, t := range tasks {
			if t.Dependent.Valid && t.Dependent.Int32 == task.ID {
				childTasks = append(childTasks, t)
			}
		}
		if len(childTasks) == 0 {
			return
		}
		sort.Slice(childTasks, func(i, j int) bool {
			prioI := getPriorityValueAscending(childTasks[i].Priority)
			prioJ := getPriorityValueAscending(childTasks[j].Priority)
			if prioI != prioJ {
				return prioI < prioJ
			}
			return childTasks[i].ID < childTasks[j].ID
		})
		for i, childTask := range childTasks {
			isLast := (i == len(childTasks)-1)
			var newPrefix, newChildPrefix string
			if isLast {
				newPrefix = childPrefix + "└── "
				newChildPrefix = childPrefix + "    "
			} else {
				newPrefix = childPrefix + "├── "
				newChildPrefix = childPrefix + "│   "
			}
			displayTask(childTask, level+1, newPrefix, newChildPrefix)
		}
	}
	for _, rootTask := range rootTasks {
		displayTask(rootTask, 0, "", "")
	}
	return displayIndexToTaskID
}

func getPriorityValueAscending(priority pgtype.Text) int {
	if !priority.Valid {
		return 0
	}
	switch priority.String {
	case "L":
		return 1
	case "M":
		return 2
	case "H":
		return 3
	default:
		return 0
	}
}

func SortTaskList(taskList []sqlc.Task) []sqlc.Task {
	taskMap := make(map[int32]sqlc.Task)
	for _, task := range taskList {
		taskMap[task.ID] = task
	}
	depthMap := make(map[int32]int)
	for _, task := range taskList {
		calculateDepth(task.ID, taskMap, depthMap)
	}
	sortedTasks := make([]sqlc.Task, len(taskList))
	copy(sortedTasks, taskList)
	sort.Slice(sortedTasks, func(i, j int) bool {
		taskI := sortedTasks[i]
		taskJ := sortedTasks[j]
		pathI := getTaskPath(taskI.ID, taskMap)
		pathJ := getTaskPath(taskJ.ID, taskMap)
		minLen := len(pathI)
		if len(pathJ) < minLen {
			minLen = len(pathJ)
		}
		for k := 0; k < minLen; k++ {
			if pathI[k] != pathJ[k] {
				return pathI[k] < pathJ[k]
			}
		}
		if len(pathI) != len(pathJ) {
			return len(pathI) < len(pathJ)
		}
		priorityI := getPriorityValue(taskI.Priority)
		priorityJ := getPriorityValue(taskJ.Priority)
		if priorityI != priorityJ {
			return priorityI > priorityJ
		}
		return taskI.ID < taskJ.ID
	})
	return sortedTasks
}

func calculateDepth(taskID int32, taskMap map[int32]sqlc.Task, depthMap map[int32]int) int {
	if depth, found := depthMap[taskID]; found {
		return depth
	}
	task, exists := taskMap[taskID]
	if !exists {
		return 0
	}
	if !task.Dependent.Valid {
		depthMap[taskID] = 0
		return 0
	}
	parentID := task.Dependent.Int32
	parentDepth := calculateDepth(parentID, taskMap, depthMap)
	depth := parentDepth + 1
	depthMap[taskID] = depth
	return depth
}

func getTaskPath(taskID int32, taskMap map[int32]sqlc.Task) []int32 {
	var path []int32
	currentID := taskID
	for {
		task, exists := taskMap[currentID]
		if !exists {
			break
		}
		path = append([]int32{currentID}, path...)
		if !task.Dependent.Valid {
			break
		}
		currentID = task.Dependent.Int32
	}
	return path
}

func getPriorityValue(priority pgtype.Text) int {
	if !priority.Valid {
		return 0
	}
	switch priority.String {
	case "H":
		return 3
	case "M":
		return 2
	case "L":
		return 1
	default:
		return 0
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
