// Package main provides the command-line interface for the CM application.
package main

import (
	"log"
	"os"
	"path/filepath"

	"github.com/lerenn/cm/pkg/config"
	"github.com/lerenn/cm/pkg/fs"
	"github.com/lerenn/cm/pkg/status"
	"github.com/spf13/cobra"
)

var (
	quiet      bool
	verbose    bool
	configPath string
)

// loadConfig loads the configuration with fallback to default.
func loadConfig() *config.Config {
	var cfg *config.Config
	var err error

	if configPath != "" {
		// Use custom config path if provided
		manager := config.NewManager()
		cfg, err = manager.LoadConfig(configPath)
		if err != nil {
			log.Printf("Failed to load custom config from %s: %v", configPath, err)
			// Fall back to default config
			cfg = manager.DefaultConfig()
		}
	} else {
		// Use default config loading logic
		homeDir, err := os.UserHomeDir()
		if err != nil {
			// Fallback to current directory if home directory cannot be determined
			homeDir = "."
		}

		defaultConfigPath := filepath.Join(homeDir, ".cm", "config.yaml")
		cfg, err = config.LoadConfigWithFallback(defaultConfigPath)
		if err != nil {
			// If there's an error, use default config
			cfg = config.NewManager().DefaultConfig()
		}
	}

	return cfg
}

// checkInitialization checks if CM is initialized and returns an error if not.
func checkInitialization() error {
	cfg := loadConfig()
	fsInstance := fs.NewFS()
	statusManager := status.NewManager(fsInstance, cfg)

	initialized, err := statusManager.IsInitialized()
	if err != nil {
		return err
	}

	if !initialized {
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

	// Add initialization check to all commands except init
	addInitializationCheck(createCmd)
	addInitializationCheck(openCmd)
	addInitializationCheck(deleteCmd)
	addInitializationCheck(listCmd)
	addInitializationCheck(loadCmd)

	// Add subcommands
	rootCmd.AddCommand(createCmd, openCmd, deleteCmd, listCmd, loadCmd, initCmd)

	if err := rootCmd.Execute(); err != nil {
		log.Fatal(err)
	}
}
