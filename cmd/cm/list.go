package main

import (
	"fmt"
	"path/filepath"
	"strings"

	cm "github.com/lerenn/cm/pkg/cm"
	"github.com/lerenn/cm/pkg/status"
	"github.com/spf13/cobra"
)

func createListCmd() *cobra.Command {
	listCmd := &cobra.Command{
		Use:   "list",
		Short: "List all worktrees",
		Long: `List all worktrees for the current repository or workspace.

Examples:
  cm list`,
		Args: cobra.NoArgs,
		RunE: func(_ *cobra.Command, _ []string) error {
			cfg := loadConfig()
			cmManager := cm.NewCM(cfg)
			cmManager.SetVerbose(verbose)

			worktrees, projectType, err := cmManager.ListWorktrees()
			if err != nil {
				return fmt.Errorf("failed to list worktrees: %w", err)
			}

			if len(worktrees) == 0 {
				fmt.Println("No worktrees found.")
				return nil
			}

			// Display worktrees based on project type
			displayWorktrees(worktrees, projectType)
			return nil
		},
	}

	return listCmd
}

// displayWorktrees displays worktrees based on project type.
func displayWorktrees(worktrees []status.Repository, projectType cm.ProjectType) {
	switch projectType {
	case cm.ProjectTypeSingleRepo:
		displaySingleRepoWorktrees(worktrees)
	case cm.ProjectTypeWorkspace:
		displayWorkspaceWorktrees(worktrees)
	case cm.ProjectTypeNone:
		// No worktrees to display
		return
	}
}

// displaySingleRepoWorktrees displays worktrees for single repository mode.
func displaySingleRepoWorktrees(worktrees []status.Repository) {
	fmt.Printf("Worktrees for repository %s:\n", worktrees[0].URL)

	for _, worktree := range worktrees {
		remote := worktree.Remote
		if remote == "" {
			remote = "origin"
		}
		fmt.Printf("  [%s] %s\n", remote, worktree.Branch)
	}
}

// displayWorkspaceWorktrees displays worktrees for workspace mode.
func displayWorkspaceWorktrees(worktrees []status.Repository) {
	workspaceName := getWorkspaceName(worktrees)
	fmt.Printf("Worktrees for workspace: %s\n\n", workspaceName)

	// Group worktrees by branch
	branchGroups := make(map[string][]status.Repository)
	for _, worktree := range worktrees {
		branchGroups[worktree.Branch] = append(branchGroups[worktree.Branch], worktree)
	}

	// Display unique branches
	displayUniqueBranches(branchGroups)
}

// getWorkspaceName extracts workspace name from worktrees.
func getWorkspaceName(worktrees []status.Repository) string {
	if len(worktrees) == 0 {
		return "Unknown"
	}

	// Try to extract workspace name from the first worktree's workspace path
	workspacePath := worktrees[0].Workspace
	if workspacePath != "" {
		workspaceFile := filepath.Base(workspacePath)
		return strings.TrimSuffix(workspaceFile, ".code-workspace")
	}

	return "Unknown"
}

// displayUniqueBranches displays branches with their repositories.
func displayUniqueBranches(branchGroups map[string][]status.Repository) {
	for branch, repos := range branchGroups {
		fmt.Printf("  %s:\n", branch)
		for _, repo := range repos {
			remote := repo.Remote
			if remote == "" {
				remote = "origin"
			}
			fmt.Printf("    %s [%s]\n", repo.URL, remote)
		}
		fmt.Println()
	}
}
