package main

import (
	"log"
	"os"
	"path/filepath"

	"github.com/lerenn/cgwt/pkg/cgwt"
	"github.com/lerenn/cgwt/pkg/config"
	"github.com/spf13/cobra"
)

var (
	quiet   bool
	verbose bool
)

// loadConfig loads the configuration with fallback to default.
func loadConfig() *config.Config {
	homeDir, err := os.UserHomeDir()
	if err != nil {
		// Fallback to current directory if home directory cannot be determined
		homeDir = "."
	}

	configPath := filepath.Join(homeDir, ".cursor", "cgwt", "config.yaml")
	cfg, err := config.LoadConfigWithFallback(configPath)
	if err != nil {
		// If there's an error, use default config
		cfg = config.NewManager().DefaultConfig()
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
		Long:  `Create worktree(s) for the specified branch. Currently only detects Git repository mode.`,
		Args:  cobra.ExactArgs(1),
		RunE: func(_ *cobra.Command, _ []string) error {
			// For now, just call the detection logic
			// The branch argument is not used yet
			cfg := loadConfig()
			cgwtManager := cgwt.NewCGWT(cfg)
			if verbose {
				cgwtManager.SetVerbose(true)
			}
			return cgwtManager.CreateWorkTree()
		},
	}

	// Add global flags
	rootCmd.PersistentFlags().BoolVarP(&quiet, "quiet", "q", false, "Suppress all output except errors")
	rootCmd.PersistentFlags().BoolVarP(&verbose, "verbose", "v", false, "Enable verbose output")

	// Add subcommands
	rootCmd.AddCommand(createCmd)

	if err := rootCmd.Execute(); err != nil {
		log.Fatal(err)
	}
}
