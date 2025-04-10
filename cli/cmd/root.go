// In cli/cmd/root.go

package cmd

import (
	"context"
	"fmt"
	"os"
	"sort"

	"github.com/jskallebak/prod/internal/auth"
	"github.com/jskallebak/prod/internal/services"
	"github.com/jskallebak/prod/internal/util"
	"github.com/spf13/cobra"
)

// rootCmd represents the base command when called without any subcommands
var rootCmd = &cobra.Command{
	Use:   "prod",
	Short: "Productivity CLI tool for tasks, notes, pomodoros, calendar, habits, and projects",
	Long: `A comprehensive productivity application combining task management (TaskWarrior-like),
notes, Pomodoro timer, calendar, habit tracking, and project management with Kanban boards.

Type 'prod help' for a list of available commands.`,
	Run: func(cmd *cobra.Command, args []string) {
		// Display app version
		fmt.Println("Productivity CLI v0.1.0")

		// Check authentication status
		token, err := auth.ReadToken()
		if err != nil {
			fmt.Println("Status: Not logged in. Run 'prod login' to authenticate.")
			return
		} else {
			// Verify token and get user info
			claim, err := auth.VerifyJWT(token)
			if err != nil {
				fmt.Println("Status: Authentication token expired or invalid. Run 'prod login' to re-authenticate.")
				return
			} else {
				// Get user email from claims
				fmt.Printf("Status: Logged in as %s\n", claim.Email)
			}

			dbpool, queries, ok := util.InitDBAndQueriesCLI()
			if !ok {
				fmt.Fprintf(os.Stderr, "Error connection to database")
				return
			}
			defer dbpool.Close()

			authService := services.NewAuthService(queries)
			userService := services.NewUserService(queries)

			user, err := authService.GetCurrentUser(context.Background())
			if err != nil {
				fmt.Fprintf(os.Stderr, "Error getting user\n")
				return
			}

			proj, err := userService.GetActiveProject(context.Background(), user.ID)
			if err != nil {
				fmt.Println("No active project")
			} else {
				fmt.Println("Active project: ", proj.Name)
			}
		}

		// Show available commands
		fmt.Println("\nUsage:")
		fmt.Printf("  %s [command]\n", cmd.Name())

		fmt.Println("\nAvailable Commands:")

		// Get all commands and sort them alphabetically
		cmds := cmd.Commands()
		sort.Slice(cmds, func(i, j int) bool {
			return cmds[i].Name() < cmds[j].Name()
		})

		// Display each command with its short description
		for _, c := range cmds {
			if !c.Hidden {
				fmt.Printf("  %-12s %s\n", c.Name(), c.Short)
			}
		}

		// Show flags
		fmt.Println("\nFlags:")
		fmt.Println("  -h, --help     help for " + cmd.Name())
		if cmd.Flags().Lookup("toggle") != nil {
			fmt.Println("  -t, --toggle   Help message for toggle")
		}

		// Show help hint
		fmt.Printf("\nUse \"%s [command] --help\" for more information about a command.\n", cmd.Name())
	},
}

// Execute adds all child commands to the root command and sets flags appropriately.
// This is called by main.main(). It only needs to happen once to the rootCmd.
func Execute() {
	err := rootCmd.Execute()
	if err != nil {
		os.Exit(1)
	}
}

func init() {
	// Here you will define your flags and configuration settings.
	// Cobra supports persistent flags, which, if defined here,
	// will be global for your application.

	// rootCmd.PersistentFlags().StringVar(&cfgFile, "config", "", "config file (default is $HOME/.cli.yaml)")

	// Cobra also supports local flags, which will only run
	// when this action is called directly.
	rootCmd.Flags().BoolP("toggle", "t", false, "Help message for toggle")
}
