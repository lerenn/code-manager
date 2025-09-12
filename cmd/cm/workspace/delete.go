// Package workspace provides workspace management commands for the CM CLI.
package workspace

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/lerenn/code-manager/cmd/cm/internal/config"
	cm "github.com/lerenn/code-manager/pkg/cm"
	pkgconfig "github.com/lerenn/code-manager/pkg/config"
	"github.com/lerenn/code-manager/pkg/logger"
	"github.com/spf13/cobra"
)

func createDeleteCmd() *cobra.Command {
	deleteCmd := &cobra.Command{
		Use:   "delete <workspace-name>",
		Short: "Delete a workspace and all associated resources",
		Long:  getDeleteCommandLongDescription(),
		Args:  cobra.ExactArgs(1),
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

By default, the command will show a confirmation prompt with a detailed summary
of what will be deleted. Use the --force flag to skip confirmation prompts.

Examples:
  # Delete workspace with confirmation
  cm workspace delete my-workspace

  # Delete workspace without confirmation
  cm workspace delete my-workspace --force`
}

// createDeleteCmdRunE creates the RunE function for the delete command.
func createDeleteCmdRunE(cmd *cobra.Command, args []string) error {
	workspaceName := args[0]

	// Get force flag
	force, err := cmd.Flags().GetBool("force")
	if err != nil {
		return fmt.Errorf("failed to get force flag: %w", err)
	}

	// Resolve config path
	var path string
	if config.ConfigPath != "" {
		path = config.ConfigPath
	} else {
		homeDir, err := os.UserHomeDir()
		if err != nil {
			homeDir = "."
		}
		path = filepath.Join(homeDir, ".cm", "config.yaml")
	}

	// Load configuration
	manager := pkgconfig.NewManager()
	cfg, err := manager.LoadConfig(path)
	if err != nil {
		return fmt.Errorf("failed to load config: %w", err)
	}

	// Create CM instance
	cmManager, err := cm.NewCM(cm.NewCMParams{
		Config: cfg,
	})
	if err != nil {
		return fmt.Errorf("failed to create CM instance: %w", err)
	}

	// Set logger based on verbosity
	if config.Verbose {
		cmManager.SetLogger(logger.NewVerboseLogger())
	}

	// Delete workspace
	params := cm.DeleteWorkspaceParams{
		WorkspaceName: workspaceName,
		Force:         force,
	}

	if err := cmManager.DeleteWorkspace(params); err != nil {
		return err
	}

	// Print success message
	if !config.Quiet {
		fmt.Printf("✓ Workspace '%s' deleted successfully\n", workspaceName)
	}

	return nil
}
