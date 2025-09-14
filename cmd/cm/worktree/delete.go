package worktree

import (
	"fmt"

	"github.com/lerenn/code-manager/cmd/cm/internal/config"
	cm "github.com/lerenn/code-manager/pkg/cm"
	"github.com/lerenn/code-manager/pkg/logger"
	"github.com/spf13/cobra"
)

func createDeleteCmd() *cobra.Command {
	var force bool
	var workspaceName string
	var repositoryName string
	var all bool

	deleteCmd := &cobra.Command{
		Use:   "delete <branch> [branch2] [branch3] ... [--force/-f] [--workspace/-w] [--repository/-r] [--all/-a]",
		Short: "Delete worktrees for the specified branches or all worktrees",
		Long:  getDeleteCmdLongDescription(),
		Args:  createDeleteCmdArgsValidator(&all, &workspaceName, &repositoryName),
		RunE:  createDeleteCmdRunE(&all, &force, &workspaceName, &repositoryName),
	}

	addDeleteCmdFlags(deleteCmd, &force, &workspaceName, &repositoryName, &all)
	return deleteCmd
}

func getDeleteCmdLongDescription() string {
	return `Delete worktrees for the specified branches or all worktrees.

You can delete multiple worktrees at once by providing multiple branch names,
or delete all worktrees using the --all flag.

Examples:
  cm worktree delete feature-branch
  cm wt delete feature-branch --force
  cm w delete feature-branch --force
  cm wt delete feature-branch --workspace my-workspace
  cm wt delete feature-branch -w my-workspace --force
  cm worktree delete branch1 branch2 branch3
  cm wt delete branch1 branch2 --force
  cm worktree delete --all
  cm wt delete --all --force
  cm worktree delete feature-branch --repository my-repo
  cm wt delete feature-branch --repository /path/to/repo --force`
}

func createDeleteCmdArgsValidator(
	all *bool,
	workspaceName *string,
	repositoryName *string,
) func(*cobra.Command, []string) error {
	return func(_ *cobra.Command, args []string) error {
		// Validate that workspace and repository are not both specified
		if *workspaceName != "" && *repositoryName != "" {
			return fmt.Errorf("cannot specify both --workspace and --repository flags")
		}

		if *all && len(args) > 0 {
			return fmt.Errorf("cannot specify both --all flag and branch names")
		}
		if !*all && len(args) == 0 {
			return fmt.Errorf("must specify either branch names or --all flag")
		}
		return nil
	}
}

func createDeleteCmdRunE(
	all *bool,
	force *bool,
	workspaceName *string,
	repositoryName *string,
) func(*cobra.Command, []string) error {
	return func(_ *cobra.Command, args []string) error {
		if err := config.CheckInitialization(); err != nil {
			return err
		}

		cfg, err := config.LoadConfig()
		if err != nil {
			return err
		}
		cmManager, err := cm.NewCM(cm.NewCMParams{
			Config: cfg,
		})
		if err != nil {
			return err
		}
		if config.Verbose {
			cmManager.SetLogger(logger.NewVerboseLogger())
		}

		if *all {
			return cmManager.DeleteAllWorktrees(*force)
		}
		return runDeleteWorktree(args, *force, *workspaceName, *repositoryName)
	}
}

func addDeleteCmdFlags(cmd *cobra.Command, force *bool, workspaceName *string, repositoryName *string, all *bool) {
	cmd.Flags().BoolVarP(force, "force", "f", false, "Skip confirmation prompts")
	cmd.Flags().StringVarP(workspaceName, "workspace", "w", "",
		"Name of the workspace to delete worktree from (optional)")
	cmd.Flags().StringVarP(repositoryName, "repository", "r", "",
		"Name of the repository to delete worktree from (optional)")
	cmd.Flags().BoolVarP(all, "all", "a", false, "Delete all worktrees")
}

func runDeleteWorktree(args []string, force bool, workspaceName string, repositoryName string) error {
	cmManager, err := initializeCM()
	if err != nil {
		return err
	}

	opts := buildDeleteWorktreeOptions(workspaceName, repositoryName)

	// If workspace or repository is specified, use single worktree deletion for each branch
	if workspaceName != "" || repositoryName != "" {
		return deleteWorktreesIndividually(cmManager, args, force, opts)
	}

	// Otherwise use bulk deletion
	return cmManager.DeleteWorkTrees(args, force)
}

func initializeCM() (cm.CM, error) {
	if err := config.CheckInitialization(); err != nil {
		return nil, err
	}

	cfg, err := config.LoadConfig()
	if err != nil {
		return nil, err
	}
	cmManager, err := cm.NewCM(cm.NewCMParams{
		Config: cfg,
	})
	if err != nil {
		return nil, err
	}
	if config.Verbose {
		cmManager.SetLogger(logger.NewVerboseLogger())
	}
	return cmManager, nil
}

func buildDeleteWorktreeOptions(workspaceName, repositoryName string) []cm.DeleteWorktreeOpts {
	var opts []cm.DeleteWorktreeOpts
	if workspaceName != "" {
		opts = append(opts, cm.DeleteWorktreeOpts{
			WorkspaceName: workspaceName,
		})
	}
	if repositoryName != "" {
		opts = append(opts, cm.DeleteWorktreeOpts{
			RepositoryName: repositoryName,
		})
	}
	return opts
}

func deleteWorktreesIndividually(cmManager cm.CM, args []string, force bool, opts []cm.DeleteWorktreeOpts) error {
	for _, branch := range args {
		if err := cmManager.DeleteWorkTree(branch, force, opts...); err != nil {
			return err
		}
	}
	return nil
}
