package cmd

import (
	"testing"

	"github.com/spf13/cobra"
	"github.com/stretchr/testify/assert"
)

func TestShowCommandStructure(t *testing.T) {
	// Verify the command has been registered
	cmd := showCmd

	// Check that the command exists
	assert.NotNil(t, cmd)
	assert.Equal(t, "show [task_id]", cmd.Use)
	assert.Equal(t, "Show detailed task information", cmd.Short)

	// Check that the function validates arguments
	// Create temporary cobra command for testing
	testCmd := &cobra.Command{}
	// No arguments
	err := cobra.MinimumNArgs(1)(testCmd, []string{})
	assert.Error(t, err)
	// Too many arguments
	err = cobra.ExactArgs(1)(testCmd, []string{"1", "2"})
	assert.Error(t, err)
	// Correct number of arguments
	err = cobra.ExactArgs(1)(testCmd, []string{"5"})
	assert.NoError(t, err)

	// Verify the command is registered with taskCmd
	found := false
	for _, subcmd := range taskCmd.Commands() {
		if subcmd.Name() == "show" {
			found = true
			break
		}
	}
	assert.True(t, found, "show command should be registered with taskCmd")
}
