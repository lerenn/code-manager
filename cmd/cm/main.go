// Package main provides the command-line interface for the CM application.
package main

import (
	"log"
	"os"
	"path/filepath"

	"github.com/lerenn/cm/pkg/config"
	"github.com/spf13/cobra"
)

var (
	quiet      bool
	verbose    bool
	configPath string
	ideName    string
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

	// Add IDE flag to create command
	createCmd.Flags().StringVarP(&ideName, "ide", "i", "", "Open in specified IDE after creation")

	// Add IDE flag to load command
	loadCmd.Flags().StringVarP(&ideName, "ide", "i", "", "Open in specified IDE after loading")

	// Add IDE flag to open command
	openCmd.Flags().StringVarP(&ideName, "ide", "i", "", "Open in specified IDE")

	// Add subcommands
	rootCmd.AddCommand(createCmd, openCmd, deleteCmd, listCmd, loadCmd, initCmd)

	if err := rootCmd.Execute(); err != nil {
		log.Fatal(err)
	}
}
