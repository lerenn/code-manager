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
	var all bool

	deleteCmd := &cobra.Command{
		Use:   "delete <branch> [branch2] [branch3] ... [--force/-f] [--workspace/-w] [--all/-a]",
		Short: "Delete worktrees for the specified branches or all worktrees",
		Long:  getDeleteCmdLongDescription(),
		Args:  createDeleteCmdArgsValidator(&all),
		RunE:  createDeleteCmdRunE(&all, &force, &workspaceName),
	}

	addDeleteCmdFlags(deleteCmd, &force, &workspaceName, &all)
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
  cm wt delete --all --force`
}

func createDeleteCmdArgsValidator(all *bool) func(*cobra.Command, []string) error {
	return func(_ *cobra.Command, args []string) error {
		if *all && len(args) > 0 {
			return fmt.Errorf("cannot specify both --all flag and branch names")
		}
		if !*all && len(args) == 0 {
			return fmt.Errorf("must specify either branch names or --all flag")
		}
		return nil
	}
}

func createDeleteCmdRunE(all *bool, force *bool, workspaceName *string) func(*cobra.Command, []string) error {
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
		return runDeleteWorktree(args, *force, *workspaceName)
	}
}

func addDeleteCmdFlags(cmd *cobra.Command, force *bool, workspaceName *string, all *bool) {
	cmd.Flags().BoolVarP(force, "force", "f", false, "Skip confirmation prompts")
	cmd.Flags().StringVarP(workspaceName, "workspace", "w", "",
		"Name of the workspace to delete worktree from (optional)")
	cmd.Flags().BoolVarP(all, "all", "a", false, "Delete all worktrees")
}

func runDeleteWorktree(args []string, force bool, workspaceName string) error {
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

	// Create options struct
	var opts []cm.DeleteWorktreeOpts
	if workspaceName != "" {
		opts = append(opts, cm.DeleteWorktreeOpts{
			WorkspaceName: workspaceName,
		})
	}

	// If workspace is specified, use single worktree deletion for each branch
	if workspaceName != "" {
		for _, branch := range args {
			if err := cmManager.DeleteWorkTree(branch, force, opts...); err != nil {
				return err
			}
		}
		return nil
	}

	// Otherwise use bulk deletion
	return cmManager.DeleteWorkTrees(args, force)
}
