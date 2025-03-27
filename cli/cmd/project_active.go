/*
Copyright © 2025 NAME HERE <EMAIL ADDRESS>
*/
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

// activeCmd represents the active command
var activeCmd = &cobra.Command{
	Use:   "active",
	Short: "A brief description of your command",
	Long: `A longer description that spans multiple lines and likely contains examples
and usage of using your command. For example:

Cobra is a CLI library for Go that empowers applications.
This application is a tool to generate the needed files
to quickly create a Cobra application.`,
	Run: func(cmd *cobra.Command, args []string) {
		fmt.Println("active called")

		dbpool, queries, ok := util.InitDBAndQueriesCLI()
		if !ok {
			fmt.Fprintf(os.Stderr, "Error connection to database: %w")
			return
		}
		defer dbpool.Close()

		authService := services.NewAuthService(queries)
		userService := services.NewUserService(queries)

		user, err := authService.GetCurrentUser(context.Background())
		if err != nil {
			fmt.Fprintf(os.Stderr, "Error getting user %w", err)
			return
		}

		projectID, err := strconv.Atoi(args[0])
		if err != nil {
			fmt.Fprintf(os.Stderr, "Error: Invalid project ID\n")
			return
		}

		err = userService.SetActiveProject(context.Background(), user.ID, int32(projectID))
		if err != nil {
			fmt.Fprintf(os.Stderr, "Error setting active project %w", err)
			return
		}

		fmt.Printf("Active project set to %d\n", projectID)

	},
}

func init() {
	projectCmd.AddCommand(activeCmd)

	// Here you will define your flags and configuration settings.

	// Cobra supports Persistent Flags which will work for this command
	// and all subcommands, e.g.:
	// activeCmd.PersistentFlags().String("foo", "", "A help for foo")

	// Cobra supports local flags which will only run when this command
	// is called directly, e.g.:
	// activeCmd.Flags().BoolP("toggle", "t", false, "Help message for toggle")
}
