package worktree

import (
	"github.com/lerenn/code-manager/cmd/cm/internal/config"
	cm "github.com/lerenn/code-manager/pkg/cm"
	"github.com/spf13/cobra"
)

func createDeleteCmd() *cobra.Command {
	var force bool
	deleteCmd := &cobra.Command{
		Use:   "delete <branch> [--force]",
		Short: "Delete a worktree for the specified branch",
		Long: `Delete a worktree for the specified branch.

Examples:
  cm worktree delete feature-branch
  cm wt delete feature-branch --force
  cm w delete feature-branch --force`,
		Args: cobra.ExactArgs(1),
		RunE: func(_ *cobra.Command, args []string) error {
			if err := config.CheckInitialization(); err != nil {
				return err
			}

			cfg, err := config.LoadConfig()
			if err != nil {
				return err
			}
			cmManager := cm.NewCM(cfg)
			cmManager.SetVerbose(config.Verbose)

			return cmManager.DeleteWorkTree(args[0], force)
		},
	}

	// Add force flag
	deleteCmd.Flags().BoolVar(&force, "force", false, "Skip confirmation prompts")

	return deleteCmd
}
