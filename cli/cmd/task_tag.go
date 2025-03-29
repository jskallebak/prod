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

var (
	taskTag []string
)

// tagCmd represents the tag command
var tagCmd = &cobra.Command{
	Use:   "tag",
	Short: "A brief description of your command",
	Long: `A longer description that spans multiple lines and likely contains examples
and usage of using your command. For example:

Cobra is a CLI library for Go that empowers applications.
This application is a tool to generate the needed files
to quickly create a Cobra application.`,
	Run: func(cmd *cobra.Command, args []string) {
		fmt.Println("tag called.")

		inputID := args[0]
		taskID, err := getID(getTaskMap, inputID)
		if err != nil {
			fmt.Println("AA")
			fmt.Fprintf(os.Stderr, "%v", err)
			return
		}

		dbpool, queries, ok := util.InitDBAndQueriesCLI()
		if !ok {
			fmt.Fprintf(os.Stderr, "Error connecting to database")
			return
		}
		defer dbpool.Close()

		taskService := services.NewTaskService(queries)
		authService := services.NewAuthService(queries)

		user, err := authService.GetCurrentUser(context.Background())
		if err != nil {
			fmt.Fprintf(os.Stderr, "Error getting the user %v", err)
		}

		if cmd.Flags().Changed("add") {
			err := taskService.AddTag(context.Background(), user.ID, taskID, taskTags)
			if err != nil {
				fmt.Fprintln(os.Stderr, err)
				return
			}
		}

		if cmd.Flags().Changed("clear") {
			err := taskService.ClearTags(context.Background(), user.ID, taskID)
			if err != nil {
				fmt.Fprintln(os.Stderr, err)
				return
			}
		}

		if cmd.Flags().Changed("remove") {
			err := taskService.RemoveTags(context.Background(), user.ID, taskID, taskTags)
			if err != nil {
				fmt.Fprintln(os.Stderr, err)
				return
			}
		}
		return

	},
}

func init() {
	taskCmd.AddCommand(tagCmd)

	tagCmd.Flags().BoolP("clear", "c", false, "Removes all tags")
	tagCmd.Flags().StringSliceVarP(&taskTags, "add", "a", []string{}, "Adds tags (comma-separated)")
	tagCmd.Flags().StringSliceVarP(&taskTags, "remove", "r", []string{}, "removes tags (comma-separated)")
	// Here you will define your flags and configuration settings.

	// Cobra supports Persistent Flags which will work for this command
	// and all subcommands, e.g.:
	// tagCmd.PersistentFlags().String("foo", "", "A help for foo")

	// Cobra supports local flags which will only run when this command
	// is called directly, e.g.:
	// tagCmd.Flags().BoolP("toggle", "t", false, "Help message for toggle")
}
