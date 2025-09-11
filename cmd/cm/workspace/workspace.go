package workspace

import (
	"github.com/spf13/cobra"
)

// CreateWorkspaceCmd creates the workspace command with all its subcommands.
func CreateWorkspaceCmd() *cobra.Command {
	workspaceCmd := &cobra.Command{
		Use:     "workspace",
		Aliases: []string{"ws"},
		Short:   "Workspace management commands",
		Long:    `Commands for managing workspaces in CM.`,
	}

	// Add workspace subcommands
	createCmd := createCreateCmd()
	workspaceCmd.AddCommand(createCmd)

	listCmd := createListCmd()
	workspaceCmd.AddCommand(listCmd)

	return workspaceCmd
}
