package worktree

import (
	"fmt"

	"github.com/lerenn/code-manager/cmd/cm/internal/config"
	cm "github.com/lerenn/code-manager/pkg/cm"
	"github.com/lerenn/code-manager/pkg/logger"
	"github.com/lerenn/code-manager/pkg/status"
	"github.com/spf13/cobra"
)

const defaultRemote = "origin"

func createListCmd() *cobra.Command {
	var workspaceName string
	var repositoryName string

	listCmd := &cobra.Command{
		Use:   "list",
		Short: "List all worktrees for a workspace or repository",
		Long:  getListCmdLongDescription(),
		Args:  cobra.NoArgs,
		RunE:  createListCmdRunE(&workspaceName, &repositoryName),
	}

	// Add workspace and repository flags to list command (optional)
	listCmd.Flags().StringVarP(&workspaceName, "workspace", "w", "",
		"Name of the workspace to list worktrees for (optional)")
	listCmd.Flags().StringVarP(&repositoryName, "repository", "r", "",
		"Name of the repository to list worktrees for (optional)")

	return listCmd
}

func getListCmdLongDescription() string {
	return `List all worktrees for a specific workspace, repository, or current repository.

Examples:
  cm worktree list                    # List worktrees for current repository
  cm worktree list --workspace my-workspace  # List worktrees for specific workspace
  cm worktree list --repository my-repo      # List worktrees for specific repository
  cm wt list -w my-workspace
  cm w list
  cm wt list -r /path/to/repo`
}

func createListCmdRunE(workspaceName, repositoryName *string) func(*cobra.Command, []string) error {
	return func(_ *cobra.Command, _ []string) error {
		cmManager, err := initializeCMForList()
		if err != nil {
			return err
		}

		// Validate that workspace and repository are not both specified
		if *workspaceName != "" && *repositoryName != "" {
			return fmt.Errorf("cannot specify both --workspace and --repository flags")
		}

		opts := buildListWorktreesOptions(*workspaceName, *repositoryName)

		worktrees, err := cmManager.ListWorktrees(opts...)
		if err != nil {
			return fmt.Errorf("failed to list worktrees: %w", err)
		}

		displayWorktrees(worktrees, *workspaceName, *repositoryName)
		return nil
	}
}

func initializeCMForList() (cm.CM, error) {
	if err := config.CheckInitialization(); err != nil {
		return nil, err
	}

	cfg, err := config.LoadConfig()
	if err != nil {
		return nil, err
	}
	cmManager, err := cm.NewCM(cm.NewCMParams{
		Config:     cfg,
		ConfigPath: config.GetConfigPath(),
	})
	if err != nil {
		return nil, err
	}
	if config.Verbose {
		cmManager.SetLogger(logger.NewVerboseLogger())
	}
	return cmManager, nil
}

func buildListWorktreesOptions(workspaceName, repositoryName string) []cm.ListWorktreesOpts {
	var opts []cm.ListWorktreesOpts
	if workspaceName != "" {
		opts = append(opts, cm.ListWorktreesOpts{
			WorkspaceName: workspaceName,
		})
	}
	if repositoryName != "" {
		opts = append(opts, cm.ListWorktreesOpts{
			RepositoryName: repositoryName,
		})
	}
	return opts
}

func displayWorktrees(worktrees []status.WorktreeInfo, workspaceName, repositoryName string) {
	switch {
	case workspaceName != "":
		displayWorkspaceWorktrees(worktrees, workspaceName)
	case repositoryName != "":
		displayRepositoryWorktreesWithName(worktrees, repositoryName)
	default:
		displayRepositoryWorktrees(worktrees)
	}
}

// displayRepositoryWorktrees displays worktrees for repository mode.
func displayRepositoryWorktrees(worktrees []status.WorktreeInfo) {
	if len(worktrees) == 0 {
		fmt.Println("No worktrees found.")
		return
	}

	fmt.Printf("Worktrees:\n")

	for _, worktree := range worktrees {
		remote := worktree.Remote
		if remote == "" {
			remote = defaultRemote
		}
		fmt.Printf("  [%s] %s\n", remote, worktree.Branch)
	}
}

// displayRepositoryWorktreesWithName displays worktrees for repository mode with repository name.
func displayRepositoryWorktreesWithName(worktrees []status.WorktreeInfo, repositoryName string) {
	fmt.Printf("Worktrees for repository '%s':\n", repositoryName)

	if len(worktrees) == 0 {
		fmt.Println("No worktrees found.")
		return
	}

	// Display worktrees in the format [remote] branch-name
	for _, worktree := range worktrees {
		remote := worktree.Remote
		if remote == "" {
			remote = defaultRemote
		}
		fmt.Printf("  [%s] %s\n", remote, worktree.Branch)
	}
}

// displayWorkspaceWorktrees displays worktrees for workspace mode.
func displayWorkspaceWorktrees(worktrees []status.WorktreeInfo, workspaceName string) {
	fmt.Printf("Worktrees for workspace '%s':\n", workspaceName)

	if len(worktrees) == 0 {
		fmt.Println("No worktrees found.")
		return
	}

	// Display worktrees in the format [remote] branch-name
	for _, worktree := range worktrees {
		remote := worktree.Remote
		if remote == "" {
			remote = defaultRemote
		}
		fmt.Printf("  [%s] %s\n", remote, worktree.Branch)
	}
}
