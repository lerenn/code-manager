// Package workspace provides workspace management commands for the CM CLI.
package workspace

import (
	"fmt"

	"github.com/lerenn/code-manager/cmd/cm/internal/cli"
	cm "github.com/lerenn/code-manager/pkg/code-manager"
	"github.com/lerenn/code-manager/pkg/logger"
	"github.com/spf13/cobra"
)

func createAddCmd() *cobra.Command {
	var workspaceName string

	addCmd := &cobra.Command{
		Use:   "add [repository_name] [-w workspace_name]",
		Short: "Add a repository to an existing workspace",
		Long:  getAddCommandLongDescription(),
		Args:  cobra.MaximumNArgs(1),
		RunE:  createAddCmdRunE,
	}

	// Add workspace flag
	addCmd.Flags().StringVarP(&workspaceName, "workspace", "w", "",
		"Add repository to the specified workspace (interactive selection if not provided)")

	return addCmd
}

// getAddCommandLongDescription returns the long description for the add command.
func getAddCommandLongDescription() string {
	return `Add a repository to an existing workspace.

This command adds a repository to an existing workspace definition in the status.yaml file.
The command will:
- Add the repository to the workspace's repository list
- Create worktrees in the new repository for all branches that already have worktrees in ALL existing repositories
- Update all existing .code-workspace files to include the new repository

You can specify the repository using:
- Repository name from status.yaml (e.g., repo1)
- Absolute path (e.g., /path/to/repo1)
- Relative path (e.g., ./repo1, ../repo2)

If no workspace name is provided, you will be prompted to select one interactively.
If no repository name is provided, you will be prompted to select one interactively.

Examples:
  # Add repository to workspace
  cm workspace add repo1 -w my-workspace

  # Add repository with absolute path
  cm ws add /path/to/repo -w my-workspace

  # Interactive selection for both workspace and repository
  cm workspace add`
}

// createAddCmdRunE creates the RunE function for the add command.
func createAddCmdRunE(cmd *cobra.Command, args []string) error {
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

	// Create add parameters (interactive selection handled in code-manager)
	params := cm.AddRepositoryToWorkspaceParams{
		WorkspaceName: workspaceName,
		Repository:    repoName,
	}

	// Add repository to workspace (interactive selection handled in code-manager)
	if err := cmManager.AddRepositoryToWorkspace(params); err != nil {
		return err
	}

	// Print success message (params may have been updated by interactive selection)
	if !cli.Quiet {
		fmt.Printf("✓ Repository '%s' added to workspace '%s' successfully\n", params.Repository, params.WorkspaceName)
	}

	return nil
}
