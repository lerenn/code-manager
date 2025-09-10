// Package main provides the command-line interface for the CM application.
package main

import (
	"log"

	"github.com/lerenn/code-manager/cmd/cm/internal/config"
	"github.com/lerenn/code-manager/cmd/cm/repository"
	"github.com/lerenn/code-manager/cmd/cm/workspace"
	"github.com/lerenn/code-manager/cmd/cm/worktree"
	"github.com/spf13/cobra"
)

func main() {
	var rootCmd = &cobra.Command{
		Use:   "cm",
		Short: "Code Manager - Git WorkTree Manager",
		Long: `A powerful CLI tool for managing Git worktrees and code development workflows ` +
			`specifically designed for modern IDEs.`,
	}

	// Add global flags
	rootCmd.PersistentFlags().BoolVarP(&config.Quiet, "quiet", "q", false, "Suppress all output except errors")
	rootCmd.PersistentFlags().BoolVarP(&config.Verbose, "verbose", "v", false, "Enable verbose output")
	rootCmd.PersistentFlags().StringVarP(&config.ConfigPath, "config", "c", "", "Specify a custom config file path")

	// Create subcommands
	repositoryCmd := repository.CreateRepositoryCmd()
	worktreeCmd := worktree.CreateWorktreeCmd()
	workspaceCmd := workspace.CreateWorkspaceCmd()
	initCmd := createInitCmd()

	// Add initialization check to all commands except init
	// Note: Individual subcommands will handle their own initialization checks

	// Add subcommands
	rootCmd.AddCommand(repositoryCmd, worktreeCmd, workspaceCmd, initCmd)

	if err := rootCmd.Execute(); err != nil {
		log.Fatal(err)
	}
}
