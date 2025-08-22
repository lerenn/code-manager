package main

import (
	"fmt"
	"log"

	cm "github.com/lerenn/code-manager/pkg/cm"
	"github.com/spf13/cobra"
)

func createOpenCmd() *cobra.Command {
	var ideName string

	openCmd := &cobra.Command{
		Use:   "open <branch>",
		Short: "Open a worktree in the specified IDE",
		Long: `Open a worktree for the specified branch in the specified IDE.

Examples:
  cm open feature-branch
  cm open main
  cm open feature-branch -i vscode
  cm open main --ide cursor`,
		Args: cobra.ExactArgs(1),
		RunE: func(_ *cobra.Command, args []string) error {
			cfg, err := loadConfig()
			if err != nil {
				return err
			}
			cmManager := cm.NewCM(cfg)
			cmManager.SetVerbose(verbose)

			// Determine IDE to use (default to "cursor" if not specified)
			ideToUse := "cursor"
			if ideName != "" {
				ideToUse = ideName
			}

			// Open the worktree
			if err := cmManager.OpenWorktree(args[0], ideToUse); err != nil {
				return fmt.Errorf("failed to open worktree: %w", err)
			}

			// Only log success message in verbose mode
			if verbose {
				log.Printf("Opened worktree for branch %s", args[0])
			}
			return nil
		},
	}

	// Add IDE flag to open command
	openCmd.Flags().StringVarP(&ideName, "ide", "i", "", "Open in specified IDE")

	return openCmd
}
