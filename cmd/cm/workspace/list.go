// Package workspace provides workspace management commands for the CM CLI.
package workspace

import (
	"fmt"

	"github.com/lerenn/code-manager/cmd/cm/internal/cli"
	cm "github.com/lerenn/code-manager/pkg/code-manager"
	"github.com/lerenn/code-manager/pkg/logger"
	"github.com/spf13/cobra"
)

func createListCmd() *cobra.Command {
	listCmd := &cobra.Command{
		Use:   "list",
		Short: "List all workspaces",
		Long:  getListCommandLongDescription(),
		RunE:  createListCmdRunE,
	}

	return listCmd
}

// getListCommandLongDescription returns the long description for the list command.
func getListCommandLongDescription() string {
	return `List all workspaces defined in the status.yaml file.

This command displays all workspaces with their associated repositories and worktrees.
Each workspace shows:
- Workspace name
- Associated repositories
- Associated worktrees

Examples:
  # List all workspaces
  cm workspace list

  # List all workspaces (using alias)
  cm ws list`
}

// createListCmdRunE creates the RunE function for the list command.
func createListCmdRunE(_ *cobra.Command, _ []string) error {
	// Create CM instance
	cmManager, err := cli.NewCodeManager()
	if err != nil {
		return fmt.Errorf("failed to create CM instance: %w", err)
	}

	// Set logger based on verbosity
	if cli.Verbose {
		cmManager.SetLogger(logger.NewVerboseLogger())
	}

	// List workspaces
	workspaces, err := cmManager.ListWorkspaces()
	if err != nil {
		return err
	}

	// Print results
	if !cli.Quiet {
		return printWorkspaces(workspaces)
	}

	return nil
}

// printWorkspaces prints the list of workspaces to stdout.
func printWorkspaces(workspaces []cm.WorkspaceInfo) error {
	if len(workspaces) == 0 {
		fmt.Println("No workspaces found.")
		return nil
	}

	fmt.Printf("Found %d workspace(s):\n\n", len(workspaces))
	for _, workspace := range workspaces {
		fmt.Printf("Workspace: %s\n", workspace.Name)

		if len(workspace.Repositories) > 0 {
			fmt.Printf("  Repositories (%d):\n", len(workspace.Repositories))
			for _, repo := range workspace.Repositories {
				fmt.Printf("    - %s\n", repo)
			}
		} else {
			fmt.Printf("  Repositories: none\n")
		}

		if len(workspace.Worktrees) > 0 {
			fmt.Printf("  Worktrees (%d):\n", len(workspace.Worktrees))
			for _, worktree := range workspace.Worktrees {
				fmt.Printf("    - %s\n", worktree)
			}
		} else {
			fmt.Printf("  Worktrees: none\n")
		}

		fmt.Println()
	}

	return nil
}
