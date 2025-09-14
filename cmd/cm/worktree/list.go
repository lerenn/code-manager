package worktree

import (
	"fmt"

	"github.com/lerenn/code-manager/cmd/cm/internal/config"
	cm "github.com/lerenn/code-manager/pkg/cm"
	"github.com/lerenn/code-manager/pkg/logger"
	"github.com/lerenn/code-manager/pkg/status"
	"github.com/spf13/cobra"
)

func createListCmd() *cobra.Command {
	var workspaceName string

	listCmd := &cobra.Command{
		Use:   "list",
		Short: "List all worktrees for a workspace or repository",
		Long: `List all worktrees for a specific workspace or current repository.

Examples:
  cm worktree list                    # List worktrees for current repository
  cm worktree list --workspace my-workspace  # List worktrees for specific workspace
  cm wt list -w my-workspace
  cm w list`,
		Args: cobra.NoArgs,
		RunE: func(_ *cobra.Command, _ []string) error {
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
			var opts []cm.ListWorktreesOpts
			if workspaceName != "" {
				opts = append(opts, cm.ListWorktreesOpts{
					WorkspaceName: workspaceName,
				})
			}

			worktrees, err := cmManager.ListWorktrees(opts...)
			if err != nil {
				return fmt.Errorf("failed to list worktrees: %w", err)
			}

			// Display worktrees based on whether workspace was specified
			if workspaceName != "" {
				displayWorkspaceWorktrees(worktrees, workspaceName)
			} else {
				displayRepositoryWorktrees(worktrees)
			}
			return nil
		},
	}

	// Add workspace flag to list command (optional)
	listCmd.Flags().StringVarP(&workspaceName, "workspace", "w", "",
		"Name of the workspace to list worktrees for (optional)")

	return listCmd
}

// displayRepositoryWorktrees displays worktrees for repository mode.
func displayRepositoryWorktrees(worktrees []status.WorktreeInfo) {
	if len(worktrees) == 0 {
		fmt.Println("No worktrees found.")
		return
	}

	fmt.Printf("Worktrees:\n")

	for _, worktree := range worktrees {
		remote := worktree.Remote
		if remote == "" {
			remote = "origin"
		}
		fmt.Printf("  [%s] %s\n", remote, worktree.Branch)
	}
}

// displayWorkspaceWorktrees displays worktrees for workspace mode.
func displayWorkspaceWorktrees(worktrees []status.WorktreeInfo, workspaceName string) {
	fmt.Printf("Worktrees for workspace '%s':\n", workspaceName)

	if len(worktrees) == 0 {
		fmt.Println("No worktrees found.")
		return
	}

	// Display worktrees in the format [remote] branch-name
	for _, worktree := range worktrees {
		remote := worktree.Remote
		if remote == "" {
			remote = "origin"
		}
		fmt.Printf("  [%s] %s\n", remote, worktree.Branch)
	}
}
