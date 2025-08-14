// Package main provides the command-line interface for the WTM application.
package main

import (
	"fmt"
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

		defaultConfigPath := filepath.Join(homeDir, ".wtm", "config.yaml")
		cfg, err = config.LoadConfigWithFallback(defaultConfigPath)
		if err != nil {
			// If there's an error, use default config
			cfg = config.NewManager().DefaultConfig()
		}
	}

	return cfg
}

func createCreateCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "create [branch]",
		Short: "Create worktree(s) for the specified branch",
		Long:  `Create worktree(s) for the specified branch. Currently supports single repository mode.`,
		Args:  cobra.ExactArgs(1),
		RunE: func(_ *cobra.Command, args []string) error {
			branch := args[0]
			cfg := loadConfig()
			cgwtManager := wtm.NewWTM(cfg)
			cgwtManager.SetVerbose(verbose)

			// Create worktree with IDE if specified
			if ideName != "" {
				return cgwtManager.CreateWorkTree(branch, &ideName)
			}

			// Just create worktree without IDE
			return cgwtManager.CreateWorkTree(branch, nil)
		},
	}
}

func createOpenCmd() *cobra.Command {
	openCmd := &cobra.Command{
		Use:   "open [worktree-name]",
		Short: "Open existing worktree in IDE",
		Long:  "Open an existing worktree in the specified IDE",
		Args:  cobra.ExactArgs(1),
		RunE: func(_ *cobra.Command, args []string) error {
			worktreeName := args[0]

			if ideName == "" {
				return fmt.Errorf("IDE name is required. Use -i or --ide flag")
			}

			// Load configuration and create WTM instance
			cfg := loadConfig()
			wtmManager := wtm.NewWTM(cfg)
			wtmManager.SetVerbose(verbose)

			// Open worktree in IDE
			return wtmManager.OpenWorktree(worktreeName, ideName)
		},
	}

	// Add IDE flag to open command
	openCmd.Flags().StringVarP(&ideName, "ide", "i", "", "IDE to open worktree in")
	err := openCmd.MarkFlagRequired("ide")
	if err != nil {
		log.Fatal(err)
	}

	return openCmd
}

func createDeleteCmd() *cobra.Command {
	var force bool

	deleteCmd := &cobra.Command{
		Use:   "delete [branch-name]",
		Short: "Delete a worktree",
		Long:  "Delete a worktree and clean up Git state",
		Args:  cobra.ExactArgs(1),
		RunE: func(_ *cobra.Command, args []string) error {
			branch := args[0]

			// Load configuration and create WTM instance
			cfg := loadConfig()
			wtmManager := wtm.NewWTM(cfg)
			wtmManager.SetVerbose(verbose)

			// Delete worktree
			return wtmManager.DeleteWorkTree(branch, force)
		},
	}

	// Add force flag to delete command
	deleteCmd.Flags().BoolVarP(&force, "force", "f", false, "Force deletion without confirmation")

	return deleteCmd
}

func main() {
	var rootCmd = &cobra.Command{
		Use:   "cgwt",
		Short: "Cursor Git WorkTree Manager",
		Long:  `A powerful CLI tool for managing Git worktrees specifically designed for Cursor IDE.`,
	}

	// Add global flags
	rootCmd.PersistentFlags().BoolVarP(&quiet, "quiet", "q", false, "Suppress all output except errors")
	rootCmd.PersistentFlags().BoolVarP(&verbose, "verbose", "v", false, "Enable verbose output")
	rootCmd.PersistentFlags().StringVarP(&configPath, "config", "c", "", "Specify a custom config file path")

	// Create commands
	createCmd := createCreateCmd()
	openCmd := createOpenCmd()
	deleteCmd := createDeleteCmd()

	// Add IDE flag to create command
	createCmd.Flags().StringVarP(&ideName, "ide", "i", "", "Open in specified IDE after creation")

	// Add subcommands
	rootCmd.AddCommand(createCmd, openCmd, deleteCmd)

	if err := rootCmd.Execute(); err != nil {
		log.Fatal(err)
	}
}
