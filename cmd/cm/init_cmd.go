package main

import (
	"github.com/lerenn/code-manager/cmd/cm/internal/cli"
	cm "github.com/lerenn/code-manager/pkg/code-manager"
	"github.com/lerenn/code-manager/pkg/dependencies"
	"github.com/lerenn/code-manager/pkg/logger"
	"github.com/spf13/cobra"
)

var (
	force           bool
	reset           bool
	repositoriesDir string
	workspacesDir   string
	statusFile      string
)

func createInitCmd() *cobra.Command {
	initCmd := &cobra.Command{
		Use:   "init [--force] [--repositories-dir <path>] [--workspaces-dir <path>] [--reset]",
		Short: "Initialize CM configuration",
		Long: `Initialize CM configuration with interactive prompts or direct path specification.

Flags:
  --force, -f              Skip interactive confirmation when using --reset flag
  --repositories-dir, -r   Set the repositories directory directly (skips interactive prompt)
  --workspaces-dir, -w     Set the workspaces directory directly
  --reset, -R              Reset existing CM configuration and start fresh`,
		Args: cobra.NoArgs,
		RunE: func(_ *cobra.Command, _ []string) error {
			return runInitCommand()
		},
	}

	// Add flags
	initCmd.Flags().BoolVarP(&force, "force", "f", false, "Skip interactive confirmation when using --reset flag")
	initCmd.Flags().StringVarP(&repositoriesDir, "repositories-dir", "r", "",
		"Set the repositories directory directly (skips interactive prompt)")
	initCmd.Flags().StringVarP(&workspacesDir, "workspaces-dir", "w", "",
		"Set the workspaces directory directly (skips interactive prompt)")
	initCmd.Flags().StringVarP(&statusFile, "status-file", "s", "",
		"Set the status file location directly (skips interactive prompt)")
	initCmd.Flags().BoolVarP(&reset, "reset", "R", false, "Reset existing CM configuration and start fresh")

	return initCmd
}

// runInitCommand executes the init command logic.
func runInitCommand() error {
	// Create config manager
	configManager := cli.NewConfigManager()

	// Create CM manager
	cmManager, err := cm.NewCodeManager(cm.NewCodeManagerParams{
		Dependencies: dependencies.New().
			WithConfig(configManager),
	})
	if err != nil {
		return err
	}
	if cli.Verbose {
		cmManager.SetLogger(logger.NewVerboseLogger())
	}

	// Prepare init options
	opts := cm.InitOpts{
		Force:           force,
		Reset:           reset,
		RepositoriesDir: repositoriesDir,
		WorkspacesDir:   workspacesDir,
		StatusFile:      statusFile,
	}

	return cmManager.Init(opts)
}
