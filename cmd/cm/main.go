// Package main provides the command-line interface for the CM application.
package main

import (
	"fmt"
	"log"
	"os"
	"path/filepath"

	"github.com/lerenn/code-manager/pkg/config"
	"github.com/lerenn/code-manager/pkg/fs"
	"github.com/lerenn/code-manager/pkg/status"
	"github.com/spf13/cobra"
)

var (
	quiet      bool
	verbose    bool
	configPath string
)

// loadConfig loads the configuration strictly, failing if not found.
func loadConfig() *config.Config {
	var cfg *config.Config
	var err error
	manager := config.NewManager()

	var path string
	if configPath != "" {
		path = configPath
	} else {
		homeDir, derr := os.UserHomeDir()
		if derr != nil {
			homeDir = "."
		}
		path = filepath.Join(homeDir, ".cm", "config.yaml")
	}

	cfg, err = manager.LoadConfigStrict(path)
	if err != nil {
		if configPath != "" {
			log.Fatalf("Configuration not found at %s. Run: cm init -c %s", path, path)
		}
		log.Fatalf("Configuration not found at %s. Run: cm init", path)
	}

	return cfg
}

// checkInitialization checks if CM is initialized and returns an error if not.
func checkInitialization() error {
	cfg := loadConfig()
	fsInstance := fs.NewFS()

	// Check if status file exists
	exists, err := fsInstance.Exists(cfg.StatusFile)
	if err != nil {
		return fmt.Errorf("failed to check status file existence: %w", err)
	}

	if !exists {
		return status.ErrNotInitialized
	}

	return nil
}

// addInitializationCheck adds a pre-run check to ensure CM is initialized.
func addInitializationCheck(cmd *cobra.Command) {
	originalRunE := cmd.RunE
	cmd.RunE = func(cmd *cobra.Command, args []string) error {
		if err := checkInitialization(); err != nil {
			return err
		}
		if originalRunE != nil {
			return originalRunE(cmd, args)
		}
		return nil
	}
}

func main() {
	var rootCmd = &cobra.Command{
		Use:   "cm",
		Short: "Code Manager - Git WorkTree Manager",
		Long: `A powerful CLI tool for managing Git worktrees and code development workflows ` +
			`specifically designed for modern IDEs.`,
	}

	// Add global flags
	rootCmd.PersistentFlags().BoolVarP(&quiet, "quiet", "q", false, "Suppress all output except errors")
	rootCmd.PersistentFlags().BoolVarP(&verbose, "verbose", "v", false, "Enable verbose output")
	rootCmd.PersistentFlags().StringVarP(&configPath, "config", "c", "", "Specify a custom config file path")

	// Create commands
	createCmd := createCreateCmd()
	openCmd := createOpenCmd()
	deleteCmd := createDeleteCmd()
	listCmd := createListCmd()
	loadCmd := createLoadCmd()
	initCmd := createInitCmd()
	cloneCmd := createCloneCmd()

	// Add initialization check to all commands except init
	addInitializationCheck(createCmd)
	addInitializationCheck(openCmd)
	addInitializationCheck(deleteCmd)
	addInitializationCheck(listCmd)
	addInitializationCheck(loadCmd)

	// Add subcommands
	rootCmd.AddCommand(createCmd, openCmd, deleteCmd, listCmd, loadCmd, initCmd, cloneCmd)

	if err := rootCmd.Execute(); err != nil {
		log.Fatal(err)
	}
}
