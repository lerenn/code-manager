package worktree

import (
	"github.com/lerenn/code-manager/cmd/cm/internal/config"
	cm "github.com/lerenn/code-manager/pkg/cm"
	"github.com/lerenn/code-manager/pkg/hooks/ide"
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
  cm worktree load feature-branch          # Uses origin:feature-branch
  cm wt load origin:feature-branch         # Explicitly specify remote
  cm w load upstream:main                  # Use different remote
  cm worktree load feature-branch --ide ` + ide.DefaultIDE + ``,
		Args: cobra.ExactArgs(1),
		RunE: func(_ *cobra.Command, args []string) error {
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

			// Prepare options for LoadWorktree
			var opts cm.LoadWorktreeOpts
			if ideName != "" {
				opts.IDEName = ideName
			}

			// Load the worktree (parsing is handled by CM manager)
			return cmManager.LoadWorktree(args[0], opts)
		},
	}

	// Add IDE flag to load command
	loadCmd.Flags().StringVarP(&ideName, "ide", "i", ide.DefaultIDE, "Open in specified IDE after loading")

	return loadCmd
}
