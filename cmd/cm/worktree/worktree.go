package worktree

import (
	"github.com/spf13/cobra"
)

// CreateWorktreeCmd creates the worktree command with all its subcommands.
func CreateWorktreeCmd() *cobra.Command {
	worktreeCmd := &cobra.Command{
		Use:     "worktree",
		Aliases: []string{"w", "wt"},
		Short:   "Worktree management commands",
		Long:    `Commands for managing Git worktrees in CM.`,
	}

	// Add worktree subcommands
	createCmd := createCreateCmd()
	openCmd := createOpenCmd()
	deleteCmd := createDeleteCmd()
	listCmd := createListCmd()
	loadCmd := createLoadCmd()

	worktreeCmd.AddCommand(createCmd, openCmd, deleteCmd, listCmd, loadCmd)

	return worktreeCmd
}
