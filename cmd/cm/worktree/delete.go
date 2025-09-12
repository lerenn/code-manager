package worktree

import (
	"fmt"

	"github.com/lerenn/code-manager/cmd/cm/internal/config"
	cm "github.com/lerenn/code-manager/pkg/cm"
	"github.com/lerenn/code-manager/pkg/logger"
	"github.com/spf13/cobra"
)

func createDeleteCmd() *cobra.Command {
	var force bool
	var all bool
	deleteCmd := &cobra.Command{
		Use:   "delete <branch> [branch2] [branch3] ... [--force/-f] [--all/-a]",
		Short: "Delete worktrees for the specified branches or all worktrees",
		Long: `Delete worktrees for the specified branches or all worktrees.

You can delete multiple worktrees at once by providing multiple branch names,
or delete all worktrees using the --all flag.

Examples:
  cm worktree delete feature-branch
  cm wt delete feature-branch --force
  cm w delete feature-branch --force
  cm worktree delete branch1 branch2 branch3
  cm wt delete branch1 branch2 --force
  cm worktree delete --all
  cm wt delete --all --force`,
		Args: func(_ *cobra.Command, args []string) error {
			if all && len(args) > 0 {
				return fmt.Errorf("cannot specify both --all flag and branch names")
			}
			if !all && len(args) == 0 {
				return fmt.Errorf("must specify either branch names or --all flag")
			}
			return nil
		},
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

			if all {
				return cmManager.DeleteAllWorktrees(force)
			}
			return cmManager.DeleteWorkTrees(args, force)
		},
	}

	// Add force flag
	deleteCmd.Flags().BoolVarP(&force, "force", "f", false, "Skip confirmation prompts")

	// Add all flag
	deleteCmd.Flags().BoolVarP(&all, "all", "a", false, "Delete all worktrees")

	return deleteCmd
}
