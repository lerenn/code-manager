package worktree

import (
	"fmt"
	"log"

	"github.com/lerenn/code-manager/cmd/cm/internal/config"
	cm "github.com/lerenn/code-manager/pkg/cm"
	"github.com/lerenn/code-manager/pkg/hooks/ide"
	"github.com/spf13/cobra"
)

func createOpenCmd() *cobra.Command {
	var ideName string

	openCmd := &cobra.Command{
		Use:   "open <branch>",
		Short: "Open a worktree in the specified IDE",
		Long: `Open a worktree for the specified branch in the specified IDE.

Examples:
  cm worktree open feature-branch
  cm wt open main
  cm w open feature-branch -i cursor
  cm worktree open main --ide ` + ide.DefaultIDE + ``,
		Args: cobra.ExactArgs(1),
		RunE: func(_ *cobra.Command, args []string) error {
			return openWorktree(args[0], ideName)
		},
	}

	// Add IDE flag to open command
	openCmd.Flags().StringVarP(&ideName, "ide", "i", "", "Open in specified IDE")

	return openCmd
}

// openWorktree handles the logic for opening a worktree.
func openWorktree(branchName, ideName string) error {
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
	cmManager.SetVerbose(config.Verbose)

	// Determine IDE to use (default to DefaultIDE if not specified)
	ideToUse := ide.DefaultIDE
	if ideName != "" {
		ideToUse = ideName
	}

	// Open the worktree
	if err := cmManager.OpenWorktree(branchName, ideToUse); err != nil {
		return fmt.Errorf("failed to open worktree: %w", err)
	}

	// Only log success message in verbose mode
	if config.Verbose {
		log.Printf("Opened worktree for branch %s", branchName)
	}
	return nil
}
