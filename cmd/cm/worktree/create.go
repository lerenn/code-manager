// Package worktree provides worktree management commands for the CM CLI.
package worktree

import (
	"github.com/lerenn/code-manager/cmd/cm/internal/config"
	cm "github.com/lerenn/code-manager/pkg/cm"
	"github.com/spf13/cobra"
)

func createCreateCmd() *cobra.Command {
	var ideName string
	var force bool

	createCmd := &cobra.Command{
		Use:   "create <branch> [--ide <ide-name>]",
		Short: "Create a worktree for the specified branch",
		Long: `Create a worktree for the specified branch in the current repository or workspace.

Examples:
  cm worktree create feature-branch
  cm wt create feature-branch --ide cursor
  cm w create feature-branch --ide vscode`,
		Args: cobra.ExactArgs(1),
		RunE: func(_ *cobra.Command, args []string) error {
			if err := config.CheckInitialization(); err != nil {
				return err
			}

			cfg, err := config.LoadConfig()
			if err != nil {
				return err
			}
			cmManager := cm.NewCM(cfg)
			cmManager.SetVerbose(config.Verbose)

			var opts cm.CreateWorkTreeOpts
			if ideName != "" {
				opts.IDEName = ideName
			}
			opts.Force = force

			return cmManager.CreateWorkTree(args[0], opts)
		},
	}

	// Add IDE flag to create command
	createCmd.Flags().StringVarP(&ideName, "ide", "i", "", "Open in specified IDE after creation")
	// Add force flag to create command
	createCmd.Flags().BoolVarP(&force, "force", "f", false, "Force creation without prompts")

	return createCmd
}
