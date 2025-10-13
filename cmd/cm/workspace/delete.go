// Package workspace provides workspace management commands for the CM CLI.
package workspace

import (
	"fmt"

	"github.com/lerenn/code-manager/cmd/cm/internal/cli"
	cm "github.com/lerenn/code-manager/pkg/code-manager"
	"github.com/lerenn/code-manager/pkg/logger"
	"github.com/spf13/cobra"
)

func createDeleteCmd() *cobra.Command {
	deleteCmd := &cobra.Command{
		Use:   "delete [workspace-name]",
		Short: "Delete a workspace and all associated resources",
		Long:  getDeleteCommandLongDescription(),
		Args:  cobra.MaximumNArgs(1),
		RunE:  createDeleteCmdRunE,
	}

	// Add force flag
	deleteCmd.Flags().BoolP("force", "f", false, "Skip confirmation prompts")

	return deleteCmd
}

// getDeleteCommandLongDescription returns the long description for the delete command.
func getDeleteCommandLongDescription() string {
	return `Delete a workspace and all associated resources.

This command completely removes a workspace definition from the status.yaml file
and deletes all associated worktrees, workspace files, and worktree-specific
workspace files.

The deletion process includes:
- Deleting all worktrees associated with the workspace from all repositories
- Deleting the main workspace file (.code-workspace)
- Deleting all worktree-specific workspace files
- Removing the workspace entry from status.yaml
- Preserving individual repository entries (they may be used by other workspaces)

If no workspace name is provided, you will be prompted to select one interactively.

By default, the command will show a confirmation prompt with a detailed summary
of what will be deleted. Use the --force flag to skip confirmation prompts.

Examples:
  # Delete workspace with confirmation
  cm workspace delete my-workspace

  # Delete workspace without confirmation
  cm workspace delete my-workspace --force

  # Interactive selection
  cm ws delete`
}

// createDeleteCmdRunE creates the RunE function for the delete command.
func createDeleteCmdRunE(cmd *cobra.Command, args []string) error {
	// Get force flag
	force, err := cmd.Flags().GetBool("force")
	if err != nil {
		return fmt.Errorf("failed to get force flag: %w", err)
	}

	// Create CM instance
	cmManager, err := cli.NewCodeManager()
	if err != nil {
		return fmt.Errorf("failed to create CM instance: %w", err)
	}

	// Set logger based on verbosity
	if cli.Verbose {
		cmManager.SetLogger(logger.NewVerboseLogger())
	}

	// Create delete parameters (interactive selection handled in code-manager)
	params := cm.DeleteWorkspaceParams{
		WorkspaceName: "",
		Force:         force,
	}
	if len(args) > 0 {
		params.WorkspaceName = args[0]
	}

	// Delete workspace (interactive selection handled in code-manager)
	if err := cmManager.DeleteWorkspace(params); err != nil {
		return err
	}

	// Print success message
	if !cli.Quiet {
		fmt.Printf("✓ Workspace '%s' deleted successfully\n", params.WorkspaceName)
	}

	return nil
}
