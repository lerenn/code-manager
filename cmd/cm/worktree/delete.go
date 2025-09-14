package worktree

import (
	"github.com/lerenn/code-manager/cmd/cm/internal/config"
	cm "github.com/lerenn/code-manager/pkg/cm"
	"github.com/lerenn/code-manager/pkg/logger"
	"github.com/spf13/cobra"
)

func createDeleteCmd() *cobra.Command {
	var force bool
	var workspaceName string
	deleteCmd := &cobra.Command{
		Use:   "delete <branch> [branch2] [branch3] ... [--force/-f] [--workspace/-w]",
		Short: "Delete worktrees for the specified branches",
		Long: `Delete worktrees for the specified branches.

You can delete multiple worktrees at once by providing multiple branch names.

Examples:
  cm worktree delete feature-branch
  cm wt delete feature-branch --force
  cm w delete feature-branch --force
  cm wt delete feature-branch --workspace my-workspace
  cm wt delete feature-branch -w my-workspace --force
  cm worktree delete branch1 branch2 branch3
  cm wt delete branch1 branch2 --force`,
		Args: cobra.MinimumNArgs(1),
		RunE: func(_ *cobra.Command, args []string) error {
			return runDeleteWorktree(args, force, workspaceName)
		},
	}

	// Add flags
	deleteCmd.Flags().BoolVarP(&force, "force", "f", false, "Skip confirmation prompts")
	deleteCmd.Flags().StringVarP(&workspaceName, "workspace", "w", "",
		"Name of the workspace to delete worktree from (optional)")

	return deleteCmd
}

func runDeleteWorktree(args []string, force bool, workspaceName string) error {
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

	// Create options struct
	var opts []cm.DeleteWorktreeOpts
	if workspaceName != "" {
		opts = append(opts, cm.DeleteWorktreeOpts{
			WorkspaceName: workspaceName,
		})
	}

	// If workspace is specified, use single worktree deletion for each branch
	if workspaceName != "" {
		for _, branch := range args {
			if err := cmManager.DeleteWorkTree(branch, force, opts...); err != nil {
				return err
			}
		}
		return nil
	}

	// Otherwise use bulk deletion
	return cmManager.DeleteWorkTrees(args, force)
}
