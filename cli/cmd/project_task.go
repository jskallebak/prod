package cmd

import (
	"github.com/spf13/cobra"
)

// projectTaskCmd represents the task subcommand under project
var projectTaskCmd = &cobra.Command{
	Use:   "task",
	Short: "Manage tasks within a project",
	Long: `Manage tasks associated with a project.

Available Commands:
  add       Add a task to a project
  list      List all tasks in a project
  remove    Remove a task from a project`,
}

func init() {
	projectCmd.AddCommand(projectTaskCmd)
}
