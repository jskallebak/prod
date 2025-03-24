package cmd

import (
	"context"
	"fmt"
	"os"
	"strconv"
	"time"

	"github.com/jskallebak/prod/internal/services"
	"github.com/jskallebak/prod/internal/util"
	"github.com/spf13/cobra"
)

var (
	pomoWorkDuration  int
	pomoBreakDuration int
	pomodoroNote      string
)

var startCmd = &cobra.Command{
	Use:   "start [task-id]",
	Short: "Start a new Pomodoro session",
	Long: `Start a new Pomodoro session, optionally linked to a task.

Examples:
  prod pomo start            # Start a Pomodoro without a task
  prod pomo start 5          # Start a Pomodoro linked to task ID 5
  prod pomo start --work 25  # Start a Pomodoro with custom work duration (25 minutes)
  prod pomo start 5 --note "Working on feature X"`,
	Args: cobra.MaximumNArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		dbpool, queries, ok := util.InitDBAndQueriesCLI()
		if !ok {
			return
		}
		defer dbpool.Close()

		// Get authenticated user
		authService := services.NewAuthService(queries)
		user, err := authService.GetCurrentUser(context.Background())
		if err != nil {
			fmt.Println("You need to be logged in to start a Pomodoro session")
			fmt.Println("Use 'prod login' to authenticate")
			return
		}

		// Check if there's already an active session
		pomoService := services.NewPomodoroService(queries)
		activeSession, err := pomoService.GetActiveSession(context.Background(), user.ID)
		if err == nil && activeSession != nil {
			fmt.Println("You already have an active Pomodoro session")
			fmt.Println("Use 'prod pomo status' to check its status")
			fmt.Println("Use 'prod pomo stop' to stop it before starting a new one")
			return
		}

		// Get task ID if provided
		var taskID *int32
		if len(args) == 1 {
			id, err := strconv.Atoi(args[0])
			if err != nil {
				fmt.Fprintf(os.Stderr, "Invalid task ID: %v\n", err)
				return
			}

			// Verify the task exists and belongs to the user
			taskService := services.NewTaskService(queries)
			_, err = taskService.GetTask(context.Background(), int32(id), user.ID)
			if err != nil {
				fmt.Fprintf(os.Stderr, "Error: %v\n", err)
				return
			}

			intID := int32(id)
			taskID = &intID
		}

		// Use default durations if not specified
		workDuration := pomoWorkDuration
		if workDuration <= 0 {
			// Get from user config or use default
			config, err := pomoService.GetUserConfig(context.Background(), user.ID)
			if err == nil {
				workDuration = int(config.WorkDuration)
			} else {
				workDuration = 25 // Default: 25 minutes
			}
		}

		breakDuration := pomoBreakDuration
		if breakDuration <= 0 {
			// Get from user config or use default
			config, err := pomoService.GetUserConfig(context.Background(), user.ID)
			if err == nil {
				breakDuration = int(config.BreakDuration)
			} else {
				breakDuration = 5 // Default: 5 minutes
			}
		}

		// Start the Pomodoro session
		session, err := pomoService.StartSession(
			context.Background(),
			user.ID,
			taskID,
			time.Duration(workDuration)*time.Minute,
			time.Duration(breakDuration)*time.Minute,
			pomodoroNote,
		)
		if err != nil {
			fmt.Fprintf(os.Stderr, "Error starting Pomodoro session: %v\n", err)
			return
		}

		fmt.Println("🍅 Pomodoro session started!")
		fmt.Printf("Work duration: %d minutes\n", workDuration)
		fmt.Printf("Break duration: %d minutes\n", breakDuration)

		if taskID != nil {
			taskService := services.NewTaskService(queries)
			task, _ := taskService.GetTask(context.Background(), *taskID, user.ID)
			fmt.Printf("Task: %s (ID: %d)\n", task.Description, task.ID)
		}

		if pomodoroNote != "" {
			fmt.Printf("Note: %s\n", pomodoroNote)
		}

		// Display end time
		startTime := session.StartTime.Time
		endTime := startTime.Add(time.Duration(workDuration) * time.Minute)
		fmt.Printf("Started at: %s\n", startTime.Format("15:04:05"))
		fmt.Printf("Work until: %s\n", endTime.Format("15:04:05"))

		fmt.Println("\nUse 'prod pomo status' to check your progress")
		fmt.Println("Use 'prod pomo stop' when you're done")
	},
}

func init() {
	pomoCmd.AddCommand(startCmd)

	// Add flags
	startCmd.Flags().IntVar(&pomoWorkDuration, "work", 0, "Work duration in minutes (default: from config or 25)")
	startCmd.Flags().IntVar(&pomoBreakDuration, "break", 0, "Break duration in minutes (default: from config or 5)")
	startCmd.Flags().StringVar(&pomodoroNote, "note", "", "Add a note to this Pomodoro session")
}
