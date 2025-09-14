package worktree

import (
	"fmt"
	"log"

	"github.com/lerenn/code-manager/cmd/cm/internal/config"
	cm "github.com/lerenn/code-manager/pkg/cm"
	"github.com/lerenn/code-manager/pkg/hooks/ide"
	"github.com/lerenn/code-manager/pkg/logger"
	"github.com/spf13/cobra"
)

func createOpenCmd() *cobra.Command {
	var ideName string
	var repositoryName string

	openCmd := &cobra.Command{
		Use:   "open <branch> [--ide <ide-name>] [--repository <repository-name>]",
		Short: "Open a worktree in the specified IDE",
		Long: `Open a worktree for the specified branch in the specified IDE.

Examples:
  cm worktree open feature-branch
  cm wt open main
  cm w open feature-branch -i cursor
  cm worktree open main --ide ` + ide.DefaultIDE + `
  cm worktree open feature-branch --repository my-repo
  cm wt open main --repository /path/to/repo --ide cursor`,
		Args: cobra.ExactArgs(1),
		RunE: func(_ *cobra.Command, args []string) error {
			return openWorktree(args[0], ideName, repositoryName)
		},
	}

	// Add IDE and repository flags to open command
	openCmd.Flags().StringVarP(&ideName, "ide", "i", "", "Open in specified IDE")
	openCmd.Flags().StringVarP(&repositoryName, "repository", "r", "",
		"Open worktree for the specified repository (name from status.yaml or path)")

	return openCmd
}

// openWorktree handles the logic for opening a worktree.
func openWorktree(branchName, ideName, repositoryName string) error {
	if err := config.CheckInitialization(); err != nil {
		return err
	}

	cfg, err := config.LoadConfig()
	if err != nil {
		return err
	}
	cmManager, err := cm.NewCM(cm.NewCMParams{
		Config:     cfg,
		ConfigPath: config.GetConfigPath(),
	})
	if err != nil {
		return err
	}
	if config.Verbose {
		cmManager.SetLogger(logger.NewVerboseLogger())
	}

	// Determine IDE to use (default to DefaultIDE if not specified)
	ideToUse := ide.DefaultIDE
	if ideName != "" {
		ideToUse = ideName
	}

	// Prepare options for OpenWorktree
	var opts []cm.OpenWorktreeOpts
	if repositoryName != "" {
		opts = append(opts, cm.OpenWorktreeOpts{
			RepositoryName: repositoryName,
		})
	}

	// Open the worktree
	if err := cmManager.OpenWorktree(branchName, ideToUse, opts...); err != nil {
		return fmt.Errorf("failed to open worktree: %w", err)
	}

	// Only log success message in verbose mode
	if config.Verbose {
		log.Printf("Opened worktree for branch %s", branchName)
	}
	return nil
}
