package worktree

import (
	"fmt"
	"log"

	"github.com/lerenn/code-manager/cmd/cm/internal/cli"
	cm "github.com/lerenn/code-manager/pkg/code-manager"
	"github.com/lerenn/code-manager/pkg/hooks/ide"
	"github.com/lerenn/code-manager/pkg/logger"
	"github.com/spf13/cobra"
)

func createOpenCmd() *cobra.Command {
	var ideName string
	var repositoryName string
	var workspaceName string

	openCmd := &cobra.Command{
		Use:   "open [branch] [--ide <ide-name>] [--workspace <workspace-name>] [--repository <repository-name>]",
		Short: "Open a worktree in the specified IDE",
		Long: `Open a worktree for the specified branch in the specified IDE.

Examples:
  cm worktree open feature-branch                    # Interactive selection of workspace/repository
  cm wt open main
  cm w open feature-branch -i cursor
  cm worktree open main --ide ` + ide.DefaultIDE + `
  cm worktree open feature-branch --workspace my-workspace
  cm worktree open feature-branch --repository my-repo
  cm wt open main --repository /path/to/repo --ide cursor`,
		Args: cobra.MaximumNArgs(1),
		RunE: func(_ *cobra.Command, args []string) error {
			branchName := ""
			if len(args) > 0 {
				branchName = args[0]
			}
			return openWorktree(branchName, ideName, workspaceName, repositoryName)
		},
	}

	// Add IDE, workspace, and repository flags to open command
	openCmd.Flags().StringVarP(&ideName, "ide", "i", "", "Open in specified IDE")
	openCmd.Flags().StringVarP(&workspaceName, "workspace", "w", "",
		"Open worktree for the specified workspace (name from status.yaml, interactive selection if not provided)")
	openCmd.Flags().StringVarP(&repositoryName, "repository", "r", "",
		"Open worktree for the specified repository (name from status.yaml or path, interactive selection if not provided)")

	return openCmd
}

// openWorktree handles the logic for opening a worktree.
func openWorktree(branchName, ideName, workspaceName, repositoryName string) error {
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

	// Determine IDE to use (default to DefaultIDE if not specified)
	ideToUse := ide.DefaultIDE
	if ideName != "" {
		ideToUse = ideName
	}

	// Prepare options for OpenWorktree (interactive selection handled in code-manager)
	var opts []cm.OpenWorktreeOpts
	if workspaceName != "" {
		opts = append(opts, cm.OpenWorktreeOpts{
			WorkspaceName: workspaceName,
		})
	}
	if repositoryName != "" {
		opts = append(opts, cm.OpenWorktreeOpts{
			RepositoryName: repositoryName,
		})
	}

	// Open the worktree (interactive selection handled in code-manager)
	if err := cmManager.OpenWorktree(branchName, ideToUse, opts...); err != nil {
		return fmt.Errorf("failed to open worktree: %w", err)
	}

	// Only log success message in verbose mode
	if cli.Verbose {
		log.Printf("Opened worktree for branch %s", branchName)
	}
	return nil
}
