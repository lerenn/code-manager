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
		Config:     cfg,
		ConfigPath: config.GetConfigPath(),
	})
	if err != nil {
		return fmt.Errorf("failed to create CM instance: %w", err)
	}

	// Set logger based on verbosity
	if config.Verbose {
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
	if !config.Quiet {
		fmt.Printf("✓ Workspace '%s' created successfully\n", workspaceName)
	}

	return nil
}
