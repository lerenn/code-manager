package main

import (
	cm "github.com/lerenn/cm/pkg/cm"
	"github.com/spf13/cobra"
)

func createLoadCmd() *cobra.Command {
	loadCmd := &cobra.Command{
		Use:   "load [remote:]<branch-name> [--ide <ide-name>]",
		Short: "Load a branch from a remote source",
		Long: `Load a branch from a remote source and create a worktree.

The remote part is optional and defaults to "origin" if not specified.

Examples:
  cm load feature-branch          # Uses origin:feature-branch
  cm load origin:feature-branch   # Explicitly specify remote
  cm load upstream:main           # Use different remote
  cm load feature-branch --ide vscode`,
		Args: cobra.ExactArgs(1),
		RunE: func(_ *cobra.Command, args []string) error {
			cfg := loadConfig()
			cmManager := cm.NewCM(cfg)
			cmManager.SetVerbose(verbose)

			// Let CM package handle all parsing and validation
			return cmManager.LoadWorktree(args[0])
		},
	}

	return loadCmd
}
