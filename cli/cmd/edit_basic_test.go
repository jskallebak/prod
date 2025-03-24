package cmd

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestEditCommandFlags(t *testing.T) {
	// Verify the command has been registered
	cmd := editCmd

	// Check that the command exists
	assert.NotNil(t, cmd)
	assert.Equal(t, "edit [task_id]", cmd.Use)
	assert.Equal(t, "Edit an existing task", cmd.Short)

	// Check the flags
	descFlag := cmd.Flag("desc")
	assert.NotNil(t, descFlag, "desc flag should exist")

	priorityFlag := cmd.Flag("priority")
	assert.NotNil(t, priorityFlag, "priority flag should exist")

	dueFlag := cmd.Flag("due")
	assert.NotNil(t, dueFlag, "due flag should exist")

	projectFlag := cmd.Flag("project")
	assert.NotNil(t, projectFlag, "project flag should exist")

	tagsFlag := cmd.Flag("tags")
	assert.NotNil(t, tagsFlag, "tags flag should exist")

	notesFlag := cmd.Flag("notes")
	assert.NotNil(t, notesFlag, "notes flag should exist")
}
