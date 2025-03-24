package cmd

import (
	"context"
	"fmt"

	"github.com/jskallebak/prod/internal/services"
	"github.com/jskallebak/prod/internal/util"
	"github.com/spf13/cobra"
)

var detachCmd = &cobra.Command{
	Use:   "detach",
	Short: "Detach task from the current Pomodoro session",
	Long: `Remove task attachment from the current Pomodoro session.

Examples:
  prod pomo detach  # Remove task attachment from the current Pomodoro session`,

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
			fmt.Println("You need to be logged in to detach a task from a Pomodoro session")
			fmt.Println("Use 'prod login' to authenticate")
			return
		}

		// Check if there's an active session
		pomoService := services.NewPomodoroService(queries)
		activeSession, err := pomoService.GetActiveSession(context.Background(), user.ID)
		if err != nil {
			fmt.Println("You don't have an active Pomodoro session")
			return
		}

		// Check if there's a task attached
		if activeSession.TaskID == nil {
			fmt.Println("There is no task attached to the current Pomodoro session")
			return
		}

		// Get task details for display
		taskService := services.NewTaskService(queries)
		task, err := taskService.GetTask(context.Background(), *activeSession.TaskID, user.ID)
		if err != nil {
			fmt.Printf("Error retrieving task: %v\n", err)
			return
		}

		// Detach the task
		_, err = pomoService.DetachTask(context.Background(), user.ID)
		if err != nil {
			fmt.Printf("Error detaching task from Pomodoro session: %v\n", err)
			return
		}

		fmt.Printf("Task '%s' (ID: %d) detached from the current Pomodoro session\n",
			task.Description, task.ID)

		// Print session status
		if activeSession.Status == services.StatusActive {
			fmt.Println("Pomodoro session is still active")
			startTime := activeSession.StartTime.Time
			workDuration := activeSession.WorkDuration
			endTime := startTime.Add(workDuration)
			fmt.Printf("Work until: %s\n", endTime.Format("15:04:05"))
		} else if activeSession.Status == services.StatusPaused {
			fmt.Println("Pomodoro session is paused")
			fmt.Println("Use 'prod pomo resume' to continue")
		}
	},
}

func init() {
	pomoCmd.AddCommand(detachCmd)
}
