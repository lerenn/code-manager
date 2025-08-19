package main

import (
	"fmt"
	"log"

	cm "github.com/lerenn/cm/pkg/cm"
	"github.com/lerenn/cm/pkg/status"
	"github.com/spf13/cobra"
)

func createOpenCmd() *cobra.Command {
	openCmd := &cobra.Command{
		Use:   "open <branch>",
		Short: "Open a worktree in the default IDE",
		Long: `Open a worktree for the specified branch in the default IDE.

Examples:
  cm open feature-branch
  cm open main`,
		Args: cobra.ExactArgs(1),
		RunE: func(_ *cobra.Command, args []string) error {
			cfg := loadConfig()
			cmManager := cm.NewCM(cfg)
			cmManager.SetVerbose(verbose)

			// Get worktrees for the branch
			worktrees, _, err := cmManager.ListWorktrees()
			if err != nil {
				return fmt.Errorf("failed to list worktrees: %w", err)
			}

			// Find the worktree for the specified branch
			var targetWorktree *status.Repository
			for _, worktree := range worktrees {
				if worktree.Branch == args[0] {
					targetWorktree = &worktree
					break
				}
			}

			if targetWorktree == nil {
				return fmt.Errorf("no worktree found for branch: %s", args[0])
			}

			// Open the worktree
			if err := cmManager.OpenWorktree(targetWorktree.URL, targetWorktree.Branch); err != nil {
				return fmt.Errorf("failed to open worktree: %w", err)
			}

			log.Printf("Opened worktree for branch %s", args[0])
			return nil
		},
	}

	return openCmd
}
