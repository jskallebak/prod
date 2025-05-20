package cmd

import (
	"context"
	"fmt"
	"os"
	"strings"
	"time"

	"github.com/jskallebak/prod/internal/db/sqlc"
	"github.com/jskallebak/prod/internal/services"
	"github.com/jskallebak/prod/internal/util"
	"github.com/spf13/cobra"
)

var (
	// Flag variables for task add command
	taskPriority  string
	taskDueDate   string
	taskProjectID int
	taskTags      []string
	taskNotes     string
	dependent     int
	interactive   bool
)

// addCmd represents the add command
var addCmd = &cobra.Command{
	Use:   "add [description]",
	Short: "Add a new task",
	Long: `Add a new task to your productivity system.
	
For example:
  prod task add "Make breakfast"
  prod task add "Finish report" --priority=H --due=2025-04-01 --project=2 --tags=work,urgent`,
	Run: func(cmd *cobra.Command, args []string) {
		// Combine all arguments into a single task description
		description := strings.Join(args, " ")
		description = strings.ReplaceAll(description, "\n", "")

		// Initialize DB connection
		dbpool, err := util.InitDB()
		if err != nil {
			fmt.Fprintf(os.Stderr, "Error connecting to database: %v\n", err)
			os.Exit(1)
		}
		defer dbpool.Close()

		// Create queries and services
		queries := sqlc.New(dbpool)
		taskService := services.NewTaskService(queries)
		authService := services.NewAuthService(queries)
		userService := services.NewUserService(queries)

		user, err := authService.GetCurrentUser(context.Background())
		if err != nil {
			fmt.Fprintf(os.Stderr, "Error getting the user: %v\n", err)
		}

		taskMap, err := services.GetTaskMap()
		if err != nil {
			fmt.Println(err)
			return
		}

		if interactive {
			addIMode(user, taskService)
			return
		}

		if len(args) == 0 {
			fmt.Println("Error: Task description is required")
			cmd.Help()
			return
		}

		// reverseTaskMap := ReverseMap(taskMap)

		params := services.TaskParams{
			Description: description,
			Tags:        taskTags,
		}

		// Parse due date if provided
		if cmd.Flags().Changed("due") {
			// Try parsing with full date format (YYYY-MM-DD)
			if taskDueDate == "today" {
				today := time.Now()
				params.DueDate = &today
			} else {
				parsedDate, err := time.Parse("2006-01-02", taskDueDate)

				// If that fails, try parsing just month and day (MM-DD)
				if err != nil {
					// Try MM-DD format and add current year
					shortDate, shortErr := time.Parse("01-02", taskDueDate)
					if shortErr != nil {
						// Also try MM/DD format
						shortDate, shortErr = time.Parse("01/02", taskDueDate)
						if shortErr != nil {
							fmt.Fprintf(os.Stderr, "Invalid date format. Please use YYYY-MM-DD, MM-DD, or MM/DD\n")
							return
						}
					}

					// Take the parsed month and day but use current year
					currentYear := time.Now().Year()
					parsedDate = time.Date(currentYear, shortDate.Month(), shortDate.Day(), 0, 0, 0, 0, time.UTC)
				}

				params.DueDate = &parsedDate
			}
		}

		// Add priority if provided
		if cmd.Flags().Changed("priority") {
			uppercasePriority := strings.ToUpper(taskPriority)
			params.Priority = &uppercasePriority
		}

		// add dependent if provided
		if cmd.Flags().Changed("subtask") {
			// converting input to DB id for foreign-key
			dbIndex, exits := taskMap[dependent]
			input := 0
			if exits {
				params.Dependent = int32(dbIndex)
			} else {
				fmt.Println("Subtask does not exists")
				fmt.Println("Enter another subtask. (Enter for none)")
				for {
					_, err := fmt.Scanln(&input)
					if err != nil {
						fmt.Println("Must be a number (Enter for none)")
					}

					if input == 0 {
						params.Dependent = int32(0)
						break
					}

					dbIndex, exits := taskMap[input]
					if exits {
						params.Dependent = int32(dbIndex)
						break
					}
					fmt.Println("Subtask does not exists")
					fmt.Println("Enter another subtask. (Enter for none)")
				}
			}
		}

		// Add project ID if provided
		if cmd.Flags().Changed("project") && taskProjectID > 0 {
			projectID := int32(taskProjectID)
			params.ProjectID = &projectID
		} else {
			proj, err := userService.GetActiveProject(context.Background(), user.ID)
			if err == nil {
				projectID := int32(proj.ID)
				params.ProjectID = &projectID
			}
		}

		// Add notes if provided
		if cmd.Flags().Changed("notes") && taskNotes != "" {
			params.Notes = &taskNotes
		}

		task, err := taskService.CreateTask(context.Background(), user.ID, params)
		if err != nil {
			fmt.Fprintf(os.Stderr, "Error creating task: %v\n", err)
			return
		}

		taskMap, index, err := services.AppendToMap(task.ID)
		if err != nil {
			fmt.Println(err)
			return
		}

		err = services.MakeTaskMapFile(taskMap)
		if err != nil {
			fmt.Println(err)
			return
		}

		fmt.Printf("Created task: %s (ID: %d) (dbID: %d)\n", description, index, task.ID)
		fmt.Printf("Created at: %s\n", task.CreatedAt.Time.Format("2006-01-02 15:04"))
	},
}

func init() {
	taskCmd.AddCommand(addCmd)

	// Define flags for the add command
	addCmd.Flags().StringVarP(&taskPriority, "priority", "p", "", "Task priority (H, M, L)")
	addCmd.Flags().StringVarP(&taskDueDate, "due", "d", "", "Due date (YYYY-MM-DD)")
	addCmd.Flags().IntVarP(&taskProjectID, "project", "P", 0, "Project ID")
	addCmd.Flags().StringSliceVarP(&taskTags, "tags", "t", []string{}, "Task tags (comma-separated)")
	addCmd.Flags().StringVarP(&taskNotes, "notes", "n", "", "Additional notes for the task")
	addCmd.Flags().IntVarP(&dependent, "subtask", "s", 0, "Makes a sub task of a task")
	addCmd.Flags().BoolVarP(&interactive, "interactive", "i", false, "Interactive add")
}

func addIMode(user *sqlc.User, ts *services.TaskService) {
	params := services.TaskParams{}
	var description string

	for description == "" {
		fmt.Printf("Enter description: ")
		fmt.Scanln(&description)
		if description == "" {
			fmt.Println("Must enter a description. (ctrl-c to exit)")
		}
	}
	params.Description = description

	var priority string
	fmt.Printf("Enter priority (H, M, L), leave blank for none: ")
	fmt.Scanln(&priority)
	if priority != "" {
		priority := strings.ToUpper(priority)
		params.Priority = &priority
	}

	var project int32
	fmt.Printf("Enter project ID, leave black for none: ")
	fmt.Scanln(&project)
	if project != 0 {
		params.ProjectID = &project
	}

	var subtask int
	fmt.Printf("If this task is gonna be subtask, enter ID of main task: ")
	fmt.Scanln(&subtask)
	if subtask != 0 {
		taskMap, err := services.GetTaskMap()
		if err != nil {
			fmt.Println("error getting task map.")
			return
		}

		taskID, exits := taskMap[subtask]
		for !exits {
			fmt.Println("No task with ID, enter again. (leave empty to skip)")
			fmt.Scanln(&subtask)
			taskID, exits = taskMap[subtask]
		}

		params.Dependent = taskID
	}

	var tags []string
	var tagsString string
	fmt.Printf("Enter tag(s): ")
	fmt.Scanln(&tagsString)
	tags = strings.Split(tagsString, " ")

	params.Tags = tags

	taskMap, err := services.GetTaskMap()
	if err != nil {
		fmt.Fprintf(os.Stderr, "%s", err)
		return
	}

	task, err := ts.CreateTask(context.Background(), user.ID, params)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error creating task: %v\n", err)
		return
	}

	taskMap, index, err := services.AppendToMap(task.ID)
	if err != nil {
		fmt.Println(err)
		return
	}

	err = services.MakeTaskMapFile(taskMap)
	if err != nil {
		fmt.Println(err)
		return
	}

	fmt.Printf("Created task: %s (ID: %d) (dbID: %s)\n", description, index, task.ID)
	fmt.Printf("Created at: %s\n", task.CreatedAt.Time.Format("2006-01-02 15:04"))
}
