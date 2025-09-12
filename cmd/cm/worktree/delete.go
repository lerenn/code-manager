package worktree

import (
	"github.com/lerenn/code-manager/cmd/cm/internal/config"
	cm "github.com/lerenn/code-manager/pkg/cm"
	"github.com/lerenn/code-manager/pkg/logger"
	"github.com/spf13/cobra"
)

func createDeleteCmd() *cobra.Command {
	var force bool
	deleteCmd := &cobra.Command{
		Use:   "delete <branch> [branch2] [branch3] ... [--force/-f]",
		Short: "Delete worktrees for the specified branches",
		Long: `Delete worktrees for the specified branches.

You can delete multiple worktrees at once by providing multiple branch names.

Examples:
  cm worktree delete feature-branch
  cm wt delete feature-branch --force
  cm w delete feature-branch --force
  cm worktree delete branch1 branch2 branch3
  cm wt delete branch1 branch2 --force`,
		Args: cobra.MinimumNArgs(1),
		RunE: func(_ *cobra.Command, args []string) error {
			if err := config.CheckInitialization(); err != nil {
				return err
			}

			cfg, err := config.LoadConfig()
			if err != nil {
				return err
			}
			cmManager, err := cm.NewCM(cm.NewCMParams{
				Config: cfg,
			})
			if err != nil {
				return err
			}
			if config.Verbose {
				cmManager.SetLogger(logger.NewVerboseLogger())
			}

			return cmManager.DeleteWorkTrees(args, force)
		},
	}

	// Add force flag
	deleteCmd.Flags().BoolVarP(&force, "force", "f", false, "Skip confirmation prompts")

	return deleteCmd
}
