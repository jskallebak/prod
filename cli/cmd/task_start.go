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

// startCmd representaslStartCmd =ts the start command
var taskStartCmd = &cobra.Command{
	Use:   "start",
	Short: "A brief description of your command",
	Long: `A longer description that spans multiple lines and likely contains examples
and usage of using your command. For example:

Cobra is a CLI library for Go that empowers applications.
This application is a tool to generate the needed files
to quickly create a Cobra application.`,
	Run: func(cmd *cobra.Command, args []string) {
		ctx := context.Background()
		inputs, err := util.ParseArgs(args)
		if err != nil {
			fmt.Fprintln(os.Stderr, err)
		}

		dbpool, queries, ok := util.InitDBAndQueriesCLI()
		if !ok {
			fmt.Fprintf(os.Stderr, "Error connecting to database: %v", err)
			os.Exit(1)
		}
		defer dbpool.Close()

		taskService := services.NewTaskService(queries)
		authService := services.NewAuthService(queries)

		user, err := authService.GetCurrentUser(context.Background())
		if err != nil {
			fmt.Println("Needs to be logged to start a task")
			return
		}

		for _, input := range inputs {
			taskID, err := services.GetID(services.GetTaskMap, input)
			if err != nil {
				fmt.Fprintln(os.Stderr, err)
			}
			if err != nil {
				fmt.Fprintf(os.Stderr, "Error invalid task ID\n")
				return
			}

			err = ConfirmCmd(ctx, taskID, user.ID, START, taskService)
			if err != nil {
				fmt.Fprintf(os.Stderr, "%s\n", err)
				return
			}

			task, err := taskService.StartTask(ctx, taskID, user.ID)
			if err != nil {
				fmt.Fprintf(os.Stderr, "No tasks with ID %v\n", taskID)
				return
			}

			fmt.Printf("Task %d marked as active\n", input)
			fmt.Printf("Description: %s\n", task.Description)
			fmt.Printf("updated at: %s\n", task.UpdatedAt.Time.Format("2006-01-02 15:04:05"))

		}

	},
}

func init() {
	taskCmd.AddCommand(taskStartCmd)

	// Here you will define your flags and configuration settings.

	// Cobra supports Persistent Flags which will work for this command
	// and all subcommands, e.g.:
	// startCmd.PersistentFlags().String("foo", "", "A help for foo")

	// Cobra supports local flags which will only run when this command
	// is called directly, e.g.:
	// startCmd.Flags().BoolP("toggle", "t", false, "Help message for toggle")
}
