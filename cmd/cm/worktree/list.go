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
	var force bool

	listCmd := &cobra.Command{
		Use:   "list",
		Short: "List all worktrees",
		Long: `List all worktrees for the current repository or workspace.

Examples:
  cm worktree list
  cm wt list
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
			cmManager, err := cm.NewCM(cfg)
			if err != nil {
				return err
			}
			if config.Verbose {
				cmManager.SetLogger(logger.NewVerboseLogger())
			}

			worktrees, projectType, err := cmManager.ListWorktrees(force)
			if err != nil {
				return fmt.Errorf("failed to list worktrees: %w", err)
			}

			if len(worktrees) == 0 {
				fmt.Println("No worktrees found.")
				return nil
			}

			// Display worktrees based on project type
			displayWorktrees(worktrees, projectType)
			return nil
		},
	}

	// Add force flag to list command
	listCmd.Flags().BoolVarP(&force, "force", "f", false, "Force listing without prompts")

	return listCmd
}

// displayWorktrees displays worktrees based on project type.
func displayWorktrees(worktrees []status.WorktreeInfo, projectType cm.ProjectType) {
	switch projectType {
	case cm.ProjectTypeSingleRepo:
		displaySingleRepoWorktrees(worktrees)
	case cm.ProjectTypeWorkspace:
		displayWorkspaceWorktrees(worktrees)
	case cm.ProjectTypeNone:
		// No worktrees to display
		return
	}
}

// displaySingleRepoWorktrees displays worktrees for single repository mode.
func displaySingleRepoWorktrees(worktrees []status.WorktreeInfo) {
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
func displayWorkspaceWorktrees(worktrees []status.WorktreeInfo) {
	fmt.Printf("Worktrees for workspace:\n\n")

	// Group worktrees by branch
	branchGroups := make(map[string][]status.WorktreeInfo)
	for _, worktree := range worktrees {
		branchGroups[worktree.Branch] = append(branchGroups[worktree.Branch], worktree)
	}

	// Display unique branches
	displayUniqueBranches(branchGroups)
}

// displayUniqueBranches displays branches with their remotes.
func displayUniqueBranches(branchGroups map[string][]status.WorktreeInfo) {
	for branch, worktrees := range branchGroups {
		fmt.Printf("  %s:\n", branch)
		for _, worktree := range worktrees {
			remote := worktree.Remote
			if remote == "" {
				remote = "origin"
			}
			fmt.Printf("    [%s]\n", remote)
		}
		fmt.Println()
	}
}
