package main

import (
	cm "github.com/lerenn/code-manager/pkg/cm"
	"github.com/spf13/cobra"
)

func createDeleteCmd() *cobra.Command {
	deleteCmd := &cobra.Command{
		Use:   "delete <branch> [--force]",
		Short: "Delete a worktree for the specified branch",
		Long: `Delete a worktree for the specified branch.

Examples:
  cm delete feature-branch
  cm delete feature-branch --force`,
		Args: cobra.ExactArgs(1),
		RunE: func(_ *cobra.Command, args []string) error {
			cfg := loadConfig()
			cmManager := cm.NewCM(cfg)
			cmManager.SetVerbose(verbose)

			return cmManager.DeleteWorkTree(args[0], force)
		},
	}

	// Add force flag
	deleteCmd.Flags().BoolVar(&force, "force", false, "Skip confirmation prompts")

	return deleteCmd
}
