// Package repository provides repository management commands for the CM CLI.
package repository

import (
	"github.com/lerenn/code-manager/cmd/cm/internal/cli"
	cm "github.com/lerenn/code-manager/pkg/code-manager"
	"github.com/lerenn/code-manager/pkg/logger"
	"github.com/spf13/cobra"
)

func createDeleteCmd() *cobra.Command {
	var force bool

	deleteCmd := &cobra.Command{
		Use:   "delete [repository-name]",
		Short: "Delete a repository and all associated resources",
		Long: `Delete a repository from CM and remove all associated worktrees and files.

This command will:
  • Delete all worktrees associated with the repository
  • Remove the repository from the status file
  • Delete the repository directory (if within base path)

If no repository name is provided, you will be prompted to select one interactively.

Use the --force flag to skip confirmation prompts.

Examples:
  cm repository delete my-repo
  cm repo delete https://github.com/user/repo.git
  cm r delete my-repo --force
  cm r delete  # Interactive selection`,
		Args: cobra.MaximumNArgs(1),
		RunE: func(_ *cobra.Command, args []string) error {
			if err := cli.CheckInitialization(); err != nil {
				return err
			}

			cmManager, err := cli.NewCodeManager()
			if err != nil {
				return err
			}
			if cli.Verbose {
				cmManager.SetLogger(logger.NewVerboseLogger())
			}

			// Create delete parameters (interactive selection handled in code-manager)
			params := cm.DeleteRepositoryParams{
				Force: force,
			}
			if len(args) > 0 {
				params.RepositoryName = args[0]
			}

			return cmManager.DeleteRepository(params)
		},
	}

	// Add flags
	deleteCmd.Flags().BoolVarP(&force, "force", "f", false, "Skip confirmation prompts")

	return deleteCmd
}
