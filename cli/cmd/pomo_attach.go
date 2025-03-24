package cmd

import (
	"context"
	"fmt"
	"os"
	"strconv"

	"github.com/jskallebak/prod/internal/services"
	"github.com/jskallebak/prod/internal/util"
	"github.com/spf13/cobra"
)

var attachCmd = &cobra.Command{
	Use:   "attach [task-id]",
	Short: "Attach a task to the current Pomodoro session",
	Long: `Attach an existing task to the current active Pomodoro session.

Examples:
  prod pomo attach 5  # Attach task with ID 5 to the current Pomodoro session`,

	Args: cobra.ExactArgs(1),
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
			fmt.Println("You need to be logged in to attach a task to a Pomodoro session")
			fmt.Println("Use 'prod login' to authenticate")
			return
		}

		// Parse task ID
		taskID, err := strconv.Atoi(args[0])
		if err != nil {
			fmt.Fprintf(os.Stderr, "Invalid task ID: %v\n", err)
			return
		}

		// Verify the task exists and belongs to the user
		taskService := services.NewTaskService(queries)
		task, err := taskService.GetTask(context.Background(), int32(taskID), user.ID)
		if err != nil {
			fmt.Fprintf(os.Stderr, "Error: %v\n", err)
			return
		}

		// Check if there's an active session
		pomoService := services.NewPomodoroService(queries)
		activeSession, err := pomoService.GetActiveSession(context.Background(), user.ID)
		if err != nil {
			fmt.Println("You don't have an active Pomodoro session")
			fmt.Println("Use 'prod pomo start' to start a new session first")
			return
		}

		// Check if the session already has the same task attached
		if activeSession.TaskID != nil && *activeSession.TaskID == int32(taskID) {
			fmt.Printf("Task '%s' (ID: %d) is already attached to the current Pomodoro session\n",
				task.Description, task.ID)
			return
		}

		// Attach the task to the session
		updatedSession, err := pomoService.AttachTask(context.Background(), user.ID, int32(taskID))
		if err != nil {
			fmt.Printf("Error attaching task to Pomodoro session: %v\n", err)
			return
		}

		fmt.Printf("📎 Task '%s' (ID: %d) attached to the current Pomodoro session\n",
			task.Description, task.ID)

		// Print session status
		if updatedSession.Status == services.StatusActive {
			fmt.Println("Pomodoro session is active")
			startTime := updatedSession.StartTime.Time
			workDuration := updatedSession.WorkDuration
			endTime := startTime.Add(workDuration)
			fmt.Printf("Work until: %s\n", endTime.Format("15:04:05"))
		} else if updatedSession.Status == services.StatusPaused {
			fmt.Println("Pomodoro session is paused")
			fmt.Println("Use 'prod pomo resume' to continue")
		}
	},
}

func init() {
	pomoCmd.AddCommand(attachCmd)
}
