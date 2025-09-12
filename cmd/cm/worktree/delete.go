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
		Use:   "delete <branch> [--force/-f] [--workspace/-w]",
		Short: "Delete a worktree for the specified branch",
		Long: `Delete a worktree for the specified branch.

Examples:
  cm worktree delete feature-branch
  cm wt delete feature-branch --force
  cm w delete feature-branch --force
  cm wt delete feature-branch --workspace my-workspace
  cm wt delete feature-branch -w my-workspace --force`,
		Args: cobra.ExactArgs(1),
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

			// Create options struct
			var opts []cm.DeleteWorktreeOpts
			if workspaceName != "" {
				opts = append(opts, cm.DeleteWorktreeOpts{
					WorkspaceName: workspaceName,
				})
			}

			return cmManager.DeleteWorkTree(args[0], force, opts...)
		},
	}

	// Add flags
	deleteCmd.Flags().BoolVarP(&force, "force", "f", false, "Skip confirmation prompts")
	deleteCmd.Flags().StringVarP(&workspaceName, "workspace", "w", "",
		"Name of the workspace to delete worktree from (optional)")

	return deleteCmd
}
