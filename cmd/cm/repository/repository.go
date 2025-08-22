package repository

import (
	"github.com/spf13/cobra"
)

// CreateRepositoryCmd creates the repository command with all its subcommands.
func CreateRepositoryCmd() *cobra.Command {
	repositoryCmd := &cobra.Command{
		Use:     "repository",
		Aliases: []string{"r", "repo"},
		Short:   "Repository management commands",
		Long:    `Commands for managing Git repositories in CM.`,
	}

	// Add repository subcommands
	cloneCmd := createCloneCmd()
	repositoryCmd.AddCommand(cloneCmd)

	return repositoryCmd
}
