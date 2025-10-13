package worktree

import (
	"github.com/lerenn/code-manager/cmd/cm/internal/cli"
	cm "github.com/lerenn/code-manager/pkg/code-manager"
	"github.com/lerenn/code-manager/pkg/hooks/ide"
	"github.com/lerenn/code-manager/pkg/logger"
	"github.com/spf13/cobra"
)

func createLoadCmd() *cobra.Command {
	var ideName string
	var repositoryName string

	loadCmd := &cobra.Command{
		Use:   "load [remote:]<branch-name> [--ide <ide-name>] [--repository <repository-name>]",
		Short: "Load a branch from a remote source",
		Long: `Load a branch from a remote source and create a worktree.

The remote part is optional and defaults to "origin" if not specified.

Examples:
  cm worktree load feature-branch          # Interactive repository selection, uses origin:feature-branch
  cm wt load origin:feature-branch         # Explicitly specify remote
  cm w load upstream:main                  # Use different remote
  cm worktree load feature-branch --ide ` + ide.DefaultIDE + `
  cm worktree load feature-branch --repository my-repo
  cm wt load origin:main --repository /path/to/repo --ide cursor`,
		Args: cobra.MaximumNArgs(1),
		RunE: func(_ *cobra.Command, args []string) error {
			if err := cli.CheckInitialization(); err != nil {
				return err
			}

			cmManager, err := cli.NewCodeManager()
			if err != nil {
				return err
			}
			if cli.Verbose {
				cmManager.SetLogger(logger.NewVerboseLogger())
			}

			// Prepare options for LoadWorktree (interactive selection handled in code-manager)
			var opts cm.LoadWorktreeOpts
			if ideName != "" {
				opts.IDEName = ideName
			}
			if repositoryName != "" {
				opts.RepositoryName = repositoryName
			}

			// Load the worktree (interactive selection handled in code-manager, parsing is handled by CM manager)
			branchRef := ""
			if len(args) > 0 {
				branchRef = args[0]
			}
			return cmManager.LoadWorktree(branchRef, opts)
		},
	}

	// Add IDE and repository flags to load command
	loadCmd.Flags().StringVarP(&ideName, "ide", "i", ide.DefaultIDE, "Open in specified IDE after loading")
	loadCmd.Flags().StringVarP(&repositoryName, "repository", "r", "",
		"Load worktree for the specified repository (name from status.yaml or path, interactive selection if not provided)")

	return loadCmd
}
