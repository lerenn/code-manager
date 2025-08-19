package main

import (
	cm "github.com/lerenn/cm/pkg/cm"
	"github.com/spf13/cobra"
)

func createCreateCmd() *cobra.Command {
	createCmd := &cobra.Command{
		Use:   "create <branch> [--ide <ide-name>]",
		Short: "Create a worktree for the specified branch",
		Long: `Create a worktree for the specified branch in the current repository or workspace.

Examples:
  cm create feature-branch
  cm create feature-branch --ide vscode
  cm create feature-branch --ide cursor`,
		Args: cobra.ExactArgs(1),
		RunE: func(_ *cobra.Command, args []string) error {
			cfg := loadConfig()
			cmManager := cm.NewCM(cfg)
			cmManager.SetVerbose(verbose)

			var opts []cm.CreateWorkTreeOpts
			if ideName != "" {
				opts = append(opts, cm.CreateWorkTreeOpts{})
			}

			return cmManager.CreateWorkTree(args[0], opts...)
		},
	}

	return createCmd
}
