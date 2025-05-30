package cmd

import (
	"context"
	"fmt"
	"os"
	"strings"
	"time"

	"github.com/jackc/pgx/v5/pgtype"
	"github.com/jskallebak/prod/internal/db/sqlc"
	"github.com/jskallebak/prod/internal/services"
	"github.com/jskallebak/prod/internal/util"
	"github.com/spf13/cobra"
)

var (
	// Flag variables for task edit command
	editPriority  string
	editDueDate   string
	editProjectID int
	editTags      []string
	editNotes     string
	editDesc      string
	editStatus    string
)

// editCmd represents the edit command
var editCmd = &cobra.Command{
	Use:   "edit [task_id]",
	Short: "Edit an existing task",
	Long: `Edit an existing task in your productivity system.
	
For example:
  prod task edit 5 --desc="Updated task description"
  prod task edit 5 --priority=H --due=2025-04-01 --project=2 --tags=work,urgent --notes="Important update"
  prod task edit 5 --status=completed

Available flags:
  --desc        Update task description
  --priority    Set priority (H/M/L)
  --due         Set due date (YYYY-MM-DD)
  --project     Set project ID
  --tags        Set tags (comma separated)
  --notes       Set additional notes
  --status      Set status (pending/completed)`,
	// Args: cobra.ExactArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		// Parse task ID from arguments
		inputs, err := util.ParseArgs(args)

		// Initialize DB connection
		dbpool, err := util.InitDB()
		if err != nil {
			fmt.Fprintf(os.Stderr, "Error connecting to database: %v\n", err)
			os.Exit(1)
		}
		defer dbpool.Close()

		// Create queries and service
		queries := sqlc.New(dbpool)
		taskService := services.NewTaskService(queries)
		authService := services.NewAuthService(queries)

		user, err := authService.GetCurrentUser(context.Background())
		if err != nil {
			fmt.Fprintf(os.Stderr, "Failed to get user: %s", err)
		}

		for _, input := range inputs {
			taskID, err := services.GetID(services.GetTaskMap, input)
			ctx := context.Background()
			if err != nil {
				fmt.Fprintf(os.Stderr, "Error: Invalid task ID\n")
				return
			}

			// Get existing task to edit
			existingTask, err := taskService.GetTask(ctx, taskID, user.ID)
			if err != nil {
				fmt.Fprintf(os.Stderr, "Error: Failed to find task with ID %d: %v\n", taskID, err)
				return
			}

			// Prepare update params with existing values
			updateParams := sqlc.UpdateTaskParams{
				ID:          existingTask.ID,
				UserID:      pgtype.Int4{Int32: user.ID, Valid: true},
				Description: existingTask.Description,
				Status:      existingTask.Status,
				Priority:    existingTask.Priority,
				DueDate:     existingTask.DueDate,
				StartDate:   existingTask.StartDate,
				ProjectID:   existingTask.ProjectID,
				Recurrence:  existingTask.Recurrence,
				Tags:        existingTask.Tags,
				Notes:       existingTask.Notes,
			}

			// Only update fields that were specifically provided
			if cmd.Flags().Changed("desc") {
				updateParams.Description = editDesc
			}

			if cmd.Flags().Changed("priority") {
				// Convert priority to uppercase for case-insensitive matching
				upperPriority := strings.ToUpper(editPriority)
				updateParams.Priority = pgtype.Text{
					String: upperPriority,
					Valid:  true,
				}
			}

			if cmd.Flags().Changed("due") {
				parsedDate, err := time.Parse("2006-01-02", editDueDate)
				if err != nil {
					fmt.Fprintf(os.Stderr, "Invalid date format. Please use YYYY-MM-DD\n")
					return
				}
				updateParams.DueDate = pgtype.Timestamptz{
					Time:  parsedDate,
					Valid: true,
				}
			}

			if cmd.Flags().Changed("project") && editProjectID > 0 {
				updateParams.ProjectID = pgtype.Int4{
					Int32: int32(editProjectID),
					Valid: true,
				}
			}

			if cmd.Flags().Changed("tags") {
				updateParams.Tags = editTags
			}

			if cmd.Flags().Changed("notes") {
				updateParams.Notes = pgtype.Text{
					String: editNotes,
					Valid:  true,
				}
			}

			if cmd.Flags().Changed("status") {
				updateParams.Status = editStatus

			}

			err = ConfirmCmd(ctx, taskID, user.ID, EDIT, taskService)
			if err != nil {
				fmt.Fprintf(os.Stderr, "%s\n", err)
				return
			}

			// Call the service to update the task
			updatedTask, err := queries.UpdateTask(ctx, updateParams)
			if err != nil {
				fmt.Fprintf(os.Stderr, "Error updating task: %v\n", err)
				return
			}

			fmt.Printf("Task %d updated successfully\n", input)
			fmt.Printf("Description: %s\n", updatedTask.Description)
			if updatedTask.Priority.Valid {
				fmt.Printf("Priority: %s\n", updatedTask.Priority.String)
			}
			if updatedTask.DueDate.Valid {
				fmt.Printf("Due date: %s\n", updatedTask.DueDate.Time.Format("2006-01-02"))
			}
			if len(updatedTask.Tags) > 0 {
				fmt.Printf("Tags: %v\n", updatedTask.Tags)
			}
			if updatedTask.Notes.Valid {
				fmt.Printf("Notes: %s\n", updatedTask.Notes.String)
			}

		}
	},
}

func init() {
	taskCmd.AddCommand(editCmd)

	// Define flags for the edit command
	editCmd.Flags().StringVar(&editDesc, "desc", "", "Updated task description")
	editCmd.Flags().StringVarP(&editPriority, "priority", "p", "", "Task priority (H, M, L)")
	editCmd.Flags().StringVarP(&editDueDate, "due", "d", "", "Due date (YYYY-MM-DD)")
	editCmd.Flags().IntVarP(&editProjectID, "project", "P", 0, "Project ID")
	editCmd.Flags().StringSliceVarP(&editTags, "tags", "t", []string{}, "Task tags (comma-separated)")
	editCmd.Flags().StringVar(&editNotes, "notes", "", "Additional notes for the task")
	editCmd.Flags().StringVarP(&editStatus, "status", "s", "", "Task status (pending, active, completed, archived)")
}
