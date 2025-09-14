package main

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
	// Resolve config path
	path := getConfigPath()

	// Handle config file for initialization
	cfg, err := loadConfigForInit(path)
	if err != nil {
		return err
	}

	// Create CM manager
	cmManager, err := cm.NewCM(cm.NewCMParams{
		Config:     cfg,
		ConfigPath: path, // Pass the custom config path
	})
	if err != nil {
		return err
	}
	if config.Verbose {
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

// getConfigPath resolves the configuration file path.
func getConfigPath() string {
	if config.ConfigPath != "" {
		return config.ConfigPath
	}
	homeDir, err := os.UserHomeDir()
	if err != nil {
		homeDir = "."
	}
	return filepath.Join(homeDir, ".cm", "config.yaml")
}

// loadConfigForInit loads configuration for initialization, handling invalid configs gracefully.
func loadConfigForInit(path string) (pkgconfig.Config, error) {
	manager := pkgconfig.NewManager()

	// Check if config file exists
	_, err := os.Stat(path)
	if err == nil {
		// Config file exists, try to load it
		cfg, err := manager.LoadConfig(path)
		if err != nil {
			// Config file exists but is invalid - this is expected for init command
			// We'll use default config and let the init process fix it
			// Return the default config with no error since this is expected behavior
			//nolint:nilerr // Intentionally ignore error - invalid config is expected for init
			return manager.DefaultConfig(), nil
		}
		return cfg, nil
	}
	if os.IsNotExist(err) {
		// Config file doesn't exist, use default config
		return manager.DefaultConfig(), nil
	}
	return pkgconfig.Config{}, fmt.Errorf("failed to check config file: %w", err)
}
