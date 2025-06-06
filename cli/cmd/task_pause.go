/*
Copyright © 2025 NAME HERE <EMAIL ADDRESS>
*/
package cmd

import (
	"context"
	"fmt"
	"os"

	"github.com/jskallebak/prod/internal/services"
	"github.com/jskallebak/prod/internal/util"
	"github.com/spf13/cobra"
)

// pauseCmd represents the pause command
var taskPauseCmd = &cobra.Command{
	Use:   "pause",
	Short: "A brief description of your command",
	Long: `A longer description that spans multiple lines and likely contains examples
and usage of using your command. For example:

Cobra is a CLI library for Go that empowers applications.
This application is a tool to generate the needed files
to quickly create a Cobra application.`,
	Run: func(cmd *cobra.Command, args []string) {
		inputs, err := util.ParseArgs(args)
		ctx := context.Background()
		if err != nil {
			fmt.Fprintf(os.Stderr, "pause: error in ParseArgs: %v", err)
			return
		}

		dbpool, queries, ok := util.InitDBAndQueriesCLI()
		if !ok {
			fmt.Fprintf(os.Stderr, "Error connection to database")
			os.Exit(1)
		}
		defer dbpool.Close()

		taskService := services.NewTaskService(queries)
		authService := services.NewAuthService(queries)

		user, err := authService.GetCurrentUser(context.Background())
		if err != nil {
			fmt.Fprintf(os.Stderr, "Needs to be logged in to show tasks\n")
			return
		}

		for _, input := range inputs {
			taskID, err := services.GetID(services.GetTaskMap, input)
			if err != nil {
				fmt.Fprintf(os.Stderr, "Error invalid task ID\n")
				return
			}

			adaptedConfirm := func(ctx context.Context, taskID int32, userID int32, action string, ts *services.TaskService) error {
				return ConfirmCmd(ctx, taskID, userID, ActionType(action), ts)
			}

			err = services.RecursiveSubtasks(ctx, user.ID, taskID, taskService, "pause", input, adaptedConfirm, taskService.PauseTask)
			if err != nil {
				fmt.Fprintf(os.Stderr, "%s\n", err)
				return
			}

			err = ConfirmCmd(ctx, taskID, user.ID, PAUSE, taskService)
			if err != nil {
				fmt.Fprintf(os.Stderr, "%s\n", err)
				return
			}

			task, err := taskService.PauseTask(context.Background(), int32(taskID), user.ID)
			if err != nil {
				fmt.Fprintf(os.Stderr, "Error: Failed to find task with ID %d: %v\n", taskID, err)
				return
			}

			fmt.Printf("Task %d marked as pending\n", input)
			fmt.Printf("Description: %s\n", task.Description)
			fmt.Printf("updated at: %s\n", task.UpdatedAt.Time.Format("2006-01-02 15:04:05"))

		}

	},
}

func init() {
	taskCmd.AddCommand(taskPauseCmd)

	// Here you will define your flags and configuration settings.

	// Cobra supports Persistent Flags which will work for this command
	// and all subcommands, e.g.:
	// pauseCmd.PersistentFlags().String("foo", "", "A help for foo")

	// Cobra supports local flags which will only run when this command
	// is called directly, e.g.:
	// pauseCmd.Flags().BoolP("toggle", "t", false, "Help message for toggle")
}
