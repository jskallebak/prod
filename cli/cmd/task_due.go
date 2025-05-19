/*
Copyright © 2025 NAME HERE <EMAIL ADDRESS>
*/
package cmd

import (
	"context"
	"fmt"
	"os"
	"time"

	"github.com/jskallebak/prod/internal/services"
	"github.com/jskallebak/prod/internal/util"
	"github.com/spf13/cobra"
)

var (
	confirm bool
	date    string
)

// dueCmd represents the due command
var dueCmd = &cobra.Command{
	Use:   "due",
	Short: "A brief description of your command",
	Long: `A longer description that spans multiple lines and likely contains examples
and usage of using your command. For example:

Cobra is a CLI library for Go that empowers applications.
This application is a tool to generate the needed files
to quickly create a Cobra application.`,
	Run: func(cmd *cobra.Command, args []string) {
		fmt.Println("due called")
		ctx := context.Background()
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
			fmt.Fprintf(os.Stderr, "Needs to be logged in to show tasks")
			return
		}

		inputs, err := ParseArgs(args)
		if err != nil {
			fmt.Fprintf(os.Stderr, "due: error in ParseArgs: %v\n", err)
		}

		for _, input := range inputs {
			taskID, err := getID(getTaskMap, input)
			if err != nil {
				fmt.Fprintf(os.Stderr, "due_today: error invalid task ID\n")
				return
			}

			if !confirm {
				err = ConfirmCmd(ctx, taskID, user.ID, DUE, taskService)
				if err != nil {
					fmt.Fprintf(os.Stderr, "due_today: %v\n", err)
					return
				}
			}

			var parsedDate *time.Time
			if date != "" {
				parsed, err := time.Parse("2006-01-02", date)
				if err != nil {
					fmt.Fprintf(os.Stderr, "Due: Invalid date format %v\n", err)
					return
				}
				parsedDate = &parsed
			}

			task, err := taskService.SetDue(ctx, user.ID, taskID, parsedDate)
			if err != nil {
				fmt.Fprintf(os.Stderr, "Due: %v", err)
				return
			}

			fmt.Printf("Task %d due_date set to today\n", input)
			fmt.Printf("Description: %s\n", task.Description)
			fmt.Printf("updated at: %s\n", task.UpdatedAt.Time.Format("2006-01-02 15:04:05"))

		}

	},
}

func init() {
	taskCmd.AddCommand(dueCmd)

	// Define flags for the delete command
	dueCmd.Flags().BoolVar(&confirm, "yes", false, "Delete without confirmation")
	dueCmd.Flags().StringVarP(&date, "date", "d", "", "Specify a date (YYYY-MM-DD)")

	// Here you will define your flags and configuration settings.

	// Cobra supports Persistent Flags which will work for this command
	// and all subcommands, e.g.:
	// dueCmd.PersistentFlags().String("foo", "", "A help for foo")

	// Cobra supports local flags which will only run when this command
	// is called directly, e.g.:
	// dueCmd.Flags().BoolP("toggle", "t", false, "Help message for toggle")
}
