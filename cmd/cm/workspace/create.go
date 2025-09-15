// Package workspace provides workspace management commands for the CM CLI.
package workspace

import (
	"fmt"

	"github.com/lerenn/code-manager/cmd/cm/internal/cli"
	cm "github.com/lerenn/code-manager/pkg/code-manager"
	"github.com/lerenn/code-manager/pkg/logger"
	"github.com/spf13/cobra"
)

func createCreateCmd() *cobra.Command {
	createCmd := &cobra.Command{
		Use:   "create <workspace-name> [repositories...]",
		Short: "Create a new workspace with repository selection",
		Long:  getCreateCommandLongDescription(),
		Args:  cobra.MinimumNArgs(1),
		RunE:  createCreateCmdRunE,
	}

	return createCmd
}

// getCreateCommandLongDescription returns the long description for the create command.
func getCreateCommandLongDescription() string {
	return `Create a new workspace with repository selection.

This command allows you to create new workspace definitions in the status.yaml file.
You can specify repositories using:
- Repository names from status.yaml (e.g., repo1, repo2)
- Absolute paths (e.g., /path/to/repo1, /path/to/repo2)
- Relative paths (e.g., ./repo1, ../repo2)

Examples:
  # Create workspace with repository names from status.yaml
  cm workspace create my-workspace repo1 repo2

  # Create workspace with absolute paths
  cm workspace create my-workspace /path/to/repo1 /path/to/repo2

  # Create workspace with relative paths
  cm workspace create my-workspace ./repo1 ../repo2

  # Create workspace with mixed repository sources
  cm workspace create my-workspace repo1 /path/to/repo2 ./repo3`
}

// createCreateCmdRunE creates the RunE function for the create command.
func createCreateCmdRunE(_ *cobra.Command, args []string) error {
	workspaceName := args[0]
	repositories := args[1:]

	// Create CM instance
	cmManager, err := cli.NewCodeManager()
	if err != nil {
		return fmt.Errorf("failed to create CM instance: %w", err)
	}

	// Set logger based on verbosity
	if cli.Verbose {
		cmManager.SetLogger(logger.NewVerboseLogger())
	}

	// Create workspace
	params := cm.CreateWorkspaceParams{
		WorkspaceName: workspaceName,
		Repositories:  repositories,
	}

	if err := cmManager.CreateWorkspace(params); err != nil {
		return err
	}

	// Print success message
	if !cli.Quiet {
		fmt.Printf("✓ Workspace '%s' created successfully\n", workspaceName)
	}

	return nil
}
