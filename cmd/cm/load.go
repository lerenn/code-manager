package main

import (
	"strings"

	cm "github.com/lerenn/code-manager/pkg/cm"
	"github.com/spf13/cobra"
)

func createLoadCmd() *cobra.Command {
	var ideName string

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
			cfg, err := loadConfig()
			if err != nil {
				return err
			}
			cmManager := cm.NewCM(cfg)
			cmManager.SetVerbose(verbose)

			// Parse remote source and branch name
			parts := strings.SplitN(args[0], ":", 2)
			if len(parts) != 2 {
				return cm.ErrInvalidArgumentFormat
			}

			remoteSource := strings.TrimSpace(parts[0])
			branchName := strings.TrimSpace(parts[1])

			if remoteSource == "" {
				return cm.ErrEmptyRemoteSource
			}

			if branchName == "" {
				return cm.ErrEmptyBranchName
			}

			// Check if branch name contains colon (invalid)
			if strings.Contains(branchName, ":") {
				return cm.ErrBranchNameContainsColon
			}

			// Prepare options for LoadWorktree
			var opts cm.LoadWorktreeOpts
			if ideName != "" {
				opts.IDEName = ideName
			}

			// Load the worktree
			return cmManager.LoadWorktree(args[0], opts)
		},
	}

	// Add IDE flag to load command
	loadCmd.Flags().StringVarP(&ideName, "ide", "i", "", "Open in specified IDE after loading")

	return loadCmd
}
