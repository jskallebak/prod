package cmd

import (
	"context"
	"fmt"
	"time"

	"github.com/jskallebak/prod/internal/services"
	"github.com/jskallebak/prod/internal/util"
	"github.com/spf13/cobra"
)

var statusCmd = &cobra.Command{
	Use:   "status",
	Short: "Show current Pomodoro status",
	Long: `Display information about the current Pomodoro session, if one exists.

Examples:
  prod pomo status  # Show status of the current Pomodoro session`,

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
			fmt.Println("You need to be logged in to view Pomodoro status")
			fmt.Println("Use 'prod login' to authenticate")
			return
		}

		// Check if there's an active session
		pomoService := services.NewPomodoroService(queries)
		activeSession, err := pomoService.GetActiveSession(context.Background(), user.ID)
		if err != nil {
			fmt.Println("You don't have an active Pomodoro session")
			fmt.Println("Use 'prod pomo start' to start a new session")
			return
		}

		// Display session status
		now := time.Now()
		startTime := activeSession.StartTime.Time
		workDuration := activeSession.WorkDuration
		var elapsedTime time.Duration
		var remainingTime time.Duration

		if activeSession.Status == services.StatusActive {
			elapsedTime = now.Sub(startTime)

			// Account for pause duration if any
			if activeSession.TotalPauseDuration > 0 {
				elapsedTime -= activeSession.TotalPauseDuration
			}

			remainingTime = workDuration - elapsedTime
			if remainingTime < 0 {
				remainingTime = 0
			}

			// Print active session status
			fmt.Println("🍅 Pomodoro session is ACTIVE")
			fmt.Printf("Started at: %s\n", startTime.Format("15:04:05"))
			fmt.Printf("Current time: %s\n", now.Format("15:04:05"))
			fmt.Printf("Time elapsed: %s\n", util.FormatDuration(elapsedTime))
			fmt.Printf("Time remaining: %s\n", util.FormatDuration(remainingTime))

			// Progress bar (25 chars wide)
			progress := 1.0
			if workDuration > 0 {
				progress = float64(elapsedTime) / float64(workDuration)
				if progress > 1.0 {
					progress = 1.0
				}
			}
			renderProgressBar(progress, 25)

		} else if activeSession.Status == services.StatusPaused {
			// For paused sessions, the elapsed time is the time until the pause
			pauseTime := activeSession.PauseTime.Time
			elapsedTime = pauseTime.Sub(startTime)

			// Account for previous pause durations
			if activeSession.TotalPauseDuration > 0 {
				elapsedTime -= activeSession.TotalPauseDuration
			}

			remainingTime = workDuration - elapsedTime
			if remainingTime < 0 {
				remainingTime = 0
			}

			pauseDuration := now.Sub(pauseTime)

			// Print paused session status
			fmt.Println("⏸️  Pomodoro session is PAUSED")
			fmt.Printf("Started at: %s\n", startTime.Format("15:04:05"))
			fmt.Printf("Paused at: %s\n", pauseTime.Format("15:04:05"))
			fmt.Printf("Pause duration: %s\n", util.FormatDuration(pauseDuration))
			fmt.Printf("Active time: %s\n", util.FormatDuration(elapsedTime))
			fmt.Printf("Time remaining: %s\n", util.FormatDuration(remainingTime))

			// Progress bar (25 chars wide)
			progress := 1.0
			if workDuration > 0 {
				progress = float64(elapsedTime) / float64(workDuration)
				if progress > 1.0 {
					progress = 1.0
				}
			}
			renderProgressBar(progress, 25)

			fmt.Println("\nUse 'prod pomo resume' to continue")
		}

		// Show attached task if any
		if activeSession.TaskID != nil {
			taskService := services.NewTaskService(queries)
			task, err := taskService.GetTask(context.Background(), *activeSession.TaskID, user.ID)
			if err == nil {
				fmt.Println("\nAttached Task:")
				fmt.Printf("ID: %d\n", task.ID)
				fmt.Printf("Description: %s\n", task.Description)
				if task.Priority.Valid {
					fmt.Printf("Priority: %s\n", task.Priority.String)
				}

				// Add project information if task has an associated project
				if task.ProjectID.Valid {
					projectService := services.NewProjectService(queries)
					project, err := projectService.GetProject(context.Background(), task.ProjectID.Int32, user.ID)
					if err == nil && project != nil {
						fmt.Printf("Project: %s\n", project.Name)
					}
				}
			}
		}

		// Show session note if any
		if activeSession.Note != "" {
			fmt.Printf("\nNote: %s\n", activeSession.Note)
		}

		// Show available commands
		fmt.Println("\nAvailable commands:")
		if activeSession.Status == services.StatusActive {
			fmt.Println("  prod pomo pause  - Pause this session")
		} else if activeSession.Status == services.StatusPaused {
			fmt.Println("  prod pomo resume - Resume this session")
		}
		fmt.Println("  prod pomo stop   - End this session")
		if activeSession.TaskID == nil {
			fmt.Println("  prod pomo attach - Attach a task to this session")
		} else {
			fmt.Println("  prod pomo detach - Remove task from this session")
		}
	},
}

// renderProgressBar displays a text-based progress bar
func renderProgressBar(progress float64, width int) {
	fmt.Println()
	filled := int(progress * float64(width))
	if filled > width {
		filled = width
	}

	fmt.Print("[")
	for i := 0; i < width; i++ {
		if i < filled {
			fmt.Print("=")
		} else if i == filled && filled < width {
			fmt.Print(">")
		} else {
			fmt.Print(" ")
		}
	}
	fmt.Printf("] %.0f%%\n", progress*100)
}

func init() {
	pomoCmd.AddCommand(statusCmd)
}
