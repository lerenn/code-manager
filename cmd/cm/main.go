// Package main provides the command-line interface for the CM application.
package main

import (
	"log"
	"os"
	"path/filepath"

	"github.com/lerenn/code-manager/pkg/config"
	"github.com/spf13/cobra"
)

var (
	quiet      bool
	verbose    bool
	configPath string
	ideName    string
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

	// Add subcommands
	rootCmd.AddCommand(createCmd, openCmd, deleteCmd, listCmd, loadCmd, initCmd)

	if err := rootCmd.Execute(); err != nil {
		log.Fatal(err)
	}
}
