/*
Copyright © 2025 NAME HERE <EMAIL ADDRESS>
*/
package cmd

import (
	"context"
	"fmt"
	"os"

	"github.com/jskallebak/prod/internal/db/sqlc"
	"github.com/jskallebak/prod/internal/services"
	"github.com/jskallebak/prod/internal/util"
	"github.com/spf13/cobra"
)

// listCmd represents the list command
var (
	listPriority string
)

var listCmd = &cobra.Command{
	Use:   "list",
	Short: "A brief description of your command",
	Long: `A longer description that spans multiple lines and likely contains examples
and usage of using your command. For example:

Cobra is a CLI library for Go that empowers applications.
This application is a tool to generate the needed files
to quickly create a Cobra application.`,
	Run: func(cmd *cobra.Command, args []string) {
		fmt.Println("list called")

		dbpool, err := util.InitDB()
		if err != nil {
			fmt.Fprintf(os.Stderr, "Error connecting to database: %v\n", err)
		}
		defer dbpool.Close()

		queries := sqlc.New(dbpool)
		taskService := services.NewTaskService(queries)

		// TODO: This is a place holder for userID, when login system are made, fix this
		userID := 1

		// Handle priority flag
		var priorityPtr *string
		if cmd.Flags().Changed("priority") {
			priorityPtr = &listPriority
		}

		tasks, err := taskService.ListTasks(context.Background(), userID, priorityPtr)
		if err != nil {
			fmt.Fprintf(os.Stderr, "Error getting list of tasks: %v\n", err)
			return
		}

		if len(tasks) == 0 {
			if cmd.Flags().Changed("priority") {
				fmt.Printf("No tasks found with priority '%s'\n", listPriority)
			} else {
				fmt.Println("No tasks found")
			}
			return
		}

		for _, task := range tasks {
			fmt.Printf("ID: %d\n", task.ID)
			fmt.Printf("Description: %s\n", task.Description)
			if task.CompletedAt.Valid {
				fmt.Printf("Completed: %s\n", task.CompletedAt.Time.Format("2006-01-02 15:04"))
			} else {
				fmt.Println("Status: Not completed")
			}
			if task.Priority.Valid {
				fmt.Printf("Priority: %s\n", task.Priority.String)
			} else {
				fmt.Println("Priority: None")
			}
			if task.DueDate.Valid {
				fmt.Printf("Due: %s\n", task.DueDate.Time.Format("2006-01-02"))
			} else {
				fmt.Println("Due: No deadline")
			}
			if len(task.Tags) > 0 {
				fmt.Printf("Tags: %v\n", task.Tags)
			} else {
				fmt.Println("Tags: None")
			}
			fmt.Println()
		}

	},
}

func init() {
	taskCmd.AddCommand(listCmd)

	// Add priority flag for filtering tasks
	listCmd.Flags().StringVarP(&listPriority, "priority", "p", "", "Filter tasks by priority (H, M, L)")

	// Here you will define your flags and configuration settings.

	// Cobra supports Persistent Flags which will work for this command
	// and all subcommands, e.g.:
	// listCmd.PersistentFlags().String("foo", "", "A help for foo")

	// Cobra supports local flags which will only run when this command
	// is called directly, e.g.:
	// listCmd.Flags().BoolP("toggle", "t", false, "Help message for toggle")
}
