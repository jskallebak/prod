/*
Copyright © 2025 NAME HERE <EMAIL ADDRESS>
*/
package cmd

import (
	"context"
	"fmt"

	"github.com/jskallebak/prod/internal/services"
	"github.com/jskallebak/prod/internal/util"
	"github.com/spf13/cobra"
)

// testCmd represents the test command
var testCmd = &cobra.Command{
	Use:   "test",
	Short: "Run tests and verify system functionality",
	Long: `Execute various tests to ensure the productivity system is functioning correctly.
This is primarily used for development and debugging purposes.`,
	Run: func(cmd *cobra.Command, args []string) {
		fmt.Println("Testing current user authentication...")

		dbpool, queries, ok := util.InitDBAndQueriesCLI()
		if !ok {
			return
		}
		defer dbpool.Close()

		authS := services.NewAuthService(queries)
		taskS := services.NewTaskService(queries)

		// Try to get the current user
		user, err := authS.GetCurrentUser(context.Background())
		if err != nil {
			fmt.Printf("No authenticated user found: %v\n", err)
			return
		}

		fmt.Println("Successfully retrieved current user:")
		fmt.Printf("ID: %d\n", user.ID)
		fmt.Printf("Email: %s\n", user.Email)
		if user.Name.Valid {
			fmt.Printf("Name: %s\n", user.Name.String)
		}

		// taskMap, err := getTaskMap()
		// if err != nil {
		// 	fmt.Println(err)
		// 	return
		// }

		tasks, err := taskS.ListTasks(context.Background(), user.ID, nil, nil, nil, nil, false)
		if err != nil {

		}

		_ = MakeTaskMap(tasks)

		//list := SortTaskList(tasks)
		fmt.Println(MakeSubtaskMap(tasks))

	},
}

func init() {
	rootCmd.AddCommand(testCmd)

	// Here you will define your flags and configuration settings.

	// Cobra supports Persistent Flags which will work for this command
	// and all subcommands, e.g.:
	// testCmd.PersistentFlags().String("foo", "", "A help for foo")

	// Cobra supports local flags which will only run when this command
	// is called directly, e.g.:
	// testCmd.Flags().BoolP("toggle", "t", false, "Help message for toggle")
}
