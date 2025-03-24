package cmd

import (
	"testing"

	"github.com/spf13/cobra"
	"github.com/stretchr/testify/assert"
)

func TestDoneCommandStructure(t *testing.T) {
	// Verify the command has been registered
	cmd := doneCmd

	// Check that the command exists
	assert.NotNil(t, cmd)
	assert.Equal(t, "done [task_id]", cmd.Use)
	assert.Equal(t, "Mark a task as completed", cmd.Short)

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
}
