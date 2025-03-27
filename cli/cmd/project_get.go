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

// getCmd represents the get command
var getCmd = &cobra.Command{
	Use:   "get",
	Short: "A brief description of your command",
	Long: `A longer description that spans multiple lines and likely contains examples
and usage of using your command. For example:

Cobra is a CLI library for Go that empowers applications.
This application is a tool to generate the needed files
to quickly create a Cobra application.`,
	Run: func(cmd *cobra.Command, args []string) {
		fmt.Println("get called")

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

		proj, err := userService.GetActiveProject(context.Background(), user.ID)
		if err != nil {
			fmt.Fprintf(os.Stderr, "Error getting the active project %w", err)
			return
		}

		fmt.Printf("Active project is: %s - ID: %d\n", proj.Name, proj.ID)
	},
}

func init() {
	projectCmd.AddCommand(getCmd)

	// Here you will define your flags and configuration settings.

	// Cobra supports Persistent Flags which will work for this command
	// and all subcommands, e.g.:
	// getCmd.PersistentFlags().String("foo", "", "A help for foo")

	// Cobra supports local flags which will only run when this command
	// is called directly, e.g.:
	// getCmd.Flags().BoolP("toggle", "t", false, "Help message for toggle")
}
