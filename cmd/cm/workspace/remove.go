// Package workspace provides workspace management commands for the CM CLI.
package workspace

import (
	"fmt"

	"github.com/lerenn/code-manager/cmd/cm/internal/cli"
	cm "github.com/lerenn/code-manager/pkg/code-manager"
	"github.com/lerenn/code-manager/pkg/logger"
	"github.com/spf13/cobra"
)

func createRemoveCmd() *cobra.Command {
	var workspaceName string

	removeCmd := &cobra.Command{
		Use:   "remove [repository_name] [-w workspace_name]",
		Short: "Remove a repository from an existing workspace",
		Long:  getRemoveCommandLongDescription(),
		Args:  cobra.MaximumNArgs(1),
		RunE:  createRemoveCmdRunE,
	}

	// Add workspace flag
	removeCmd.Flags().StringVarP(&workspaceName, "workspace", "w", "",
		"Remove repository from the specified workspace (interactive selection if not provided)")

	return removeCmd
}

// getRemoveCommandLongDescription returns the long description for the remove command.
func getRemoveCommandLongDescription() string {
	return `Remove a repository from an existing workspace.

This command removes a repository from an existing workspace definition in the status.yaml file.
The command will:
- Remove the repository from the workspace's repository list
- Update all existing .code-workspace files to remove the repository folder entries
- Preserve all worktrees (they are not deleted, only removed from the workspace)

You can specify the repository using:
- Repository name from status.yaml (e.g., repo1)
- Absolute path (e.g., /path/to/repo1)
- Relative path (e.g., ./repo1, ../repo2)

If no workspace name is provided, you will be prompted to select one interactively.
If no repository name is provided, you will be prompted to select one interactively.

Examples:
  # Remove repository from workspace
  cm workspace remove repo1 -w my-workspace

  # Remove repository with absolute path
  cm ws remove /path/to/repo -w my-workspace

  # Interactive selection for both workspace and repository
  cm workspace remove`
}

// createRemoveCmdRunE creates the RunE function for the remove command.
func createRemoveCmdRunE(cmd *cobra.Command, args []string) error {
	// Get workspace flag
	workspaceName, err := cmd.Flags().GetString("workspace")
	if err != nil {
		return fmt.Errorf("failed to get workspace flag: %w", err)
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

	// Get repository name from args
	repoName := ""
	if len(args) > 0 {
		repoName = args[0]
	}

	// Create remove parameters (interactive selection handled in code-manager)
	params := cm.RemoveRepositoryFromWorkspaceParams{
		WorkspaceName: workspaceName,
		Repository:    repoName,
	}

	// Remove repository from workspace (interactive selection handled in code-manager)
	if err := cmManager.RemoveRepositoryFromWorkspace(params); err != nil {
		return err
	}

	// Print success message (params may have been updated by interactive selection)
	if !cli.Quiet {
		fmt.Printf("✓ Repository '%s' removed from workspace '%s' successfully\n", params.Repository, params.WorkspaceName)
	}

	return nil
}
