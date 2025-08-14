// Package main provides the command-line interface for the WTM application.
package main

import (
	"log"
	"os"
	"path/filepath"

	"github.com/lerenn/wtm/pkg/config"
	"github.com/lerenn/wtm/pkg/wtm"
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

		defaultConfigPath := filepath.Join(homeDir, ".cgwt", "config.yaml")
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
		Use:   "cgwt",
		Short: "Cursor Git WorkTree Manager",
		Long:  `A powerful CLI tool for managing Git worktrees specifically designed for Cursor IDE.`,
	}

	var createCmd = &cobra.Command{
		Use:   "create [branch]",
		Short: "Create worktree(s) for the specified branch",
		Long:  `Create worktree(s) for the specified branch. Currently supports single repository mode.`,
		Args:  cobra.ExactArgs(1),
		RunE: func(_ *cobra.Command, args []string) error {
			branch := args[0]
			cfg := loadConfig()
			cgwtManager := wtm.NewWTM(cfg)
			if verbose {
				cgwtManager.SetVerbose(true)
			}
			return cgwtManager.CreateWorkTree(branch)
		},
	}

	// Add global flags
	rootCmd.PersistentFlags().BoolVarP(&quiet, "quiet", "q", false, "Suppress all output except errors")
	rootCmd.PersistentFlags().BoolVarP(&verbose, "verbose", "v", false, "Enable verbose output")
	rootCmd.PersistentFlags().StringVarP(&configPath, "config", "c", "", "Specify a custom config file path")

	// Add subcommands
	rootCmd.AddCommand(createCmd)

	if err := rootCmd.Execute(); err != nil {
		log.Fatal(err)
	}
}
