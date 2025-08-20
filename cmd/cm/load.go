package main

import (
	"strings"

	cm "github.com/lerenn/cm/pkg/cm"
	"github.com/spf13/cobra"
)

func createLoadCmd() *cobra.Command {
	var ideName string

	loadCmd := &cobra.Command{
		Use:   "load <remote-source:branch-name> [--ide <ide-name>]",
		Short: "Load a branch from a remote source",
		Long: `Load a branch from a remote source and create a worktree.

Examples:
  cm load origin:feature-branch
  cm load upstream:main
  cm load origin:feature-branch --ide vscode`,
		Args: cobra.ExactArgs(1),
		RunE: func(_ *cobra.Command, args []string) error {
			cfg := loadConfig()
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

			// Load the worktree
			var opts []cm.LoadWorktreeOpts
			if ideName != "" {
				opts = append(opts, cm.LoadWorktreeOpts{IDEName: ideName})
			}

			return cmManager.LoadWorktree(args[0], opts...)
		},
	}

	// Add IDE flag to load command
	loadCmd.Flags().StringVarP(&ideName, "ide", "i", "", "Open in specified IDE after loading")

	return loadCmd
}
