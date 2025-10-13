package codemanager

import (
	"fmt"
	"sort"

	"github.com/lerenn/code-manager/pkg/code-manager/consts"
	"github.com/lerenn/code-manager/pkg/prompt"
)

// TargetSelectionResult represents the result of an interactive target selection.
type TargetSelectionResult struct {
	Name     string // The name of the selected repository or workspace
	Type     string // The type of the selected target (repository or workspace)
	Worktree string // The selected worktree name (empty for single-step selection)
}

// promptSelectTargetAndWorktree prompts the user to select a repository/workspace first, then a worktree.
func (c *realCodeManager) promptSelectTargetAndWorktree() (TargetSelectionResult, error) {
	// Step 1: Select target
	targetResult, err := c.promptSelectTarget("", "two-step selection")
	if err != nil {
		return TargetSelectionResult{}, err
	}

	// Step 2: Select worktree from the chosen target
	worktreeChoices, err := c.buildWorktreeChoices(targetResult.Type, targetResult.Name)
	if err != nil {
		return TargetSelectionResult{}, fmt.Errorf("failed to build worktree choices: %w", err)
	}

	if len(worktreeChoices) == 0 {
		return TargetSelectionResult{}, fmt.Errorf("no worktrees available for selected %s: %s",
			targetResult.Type, targetResult.Name)
	}

	if c.deps.Logger != nil {
		c.deps.Logger.Logf("Step 2: Prompting user to select worktree from %d choices", len(worktreeChoices))
	}

	// Use the prompt package to get worktree selection
	selectedWorktreeChoice, err := c.deps.Prompt.PromptSelectTarget(worktreeChoices, false)
	if err != nil {
		return TargetSelectionResult{}, fmt.Errorf("failed to get worktree selection: %w", err)
	}

	if c.deps.Logger != nil {
		c.deps.Logger.Logf("User selected worktree: %s", selectedWorktreeChoice.Name)
	}

	return TargetSelectionResult{
		Name:     targetResult.Name,
		Type:     targetResult.Type,
		Worktree: selectedWorktreeChoice.Name,
	}, nil
}

// buildWorktreeChoices builds a list of worktree choices for a specific target.
func (c *realCodeManager) buildWorktreeChoices(targetType, targetName string) ([]prompt.TargetChoice, error) {
	var choices []prompt.TargetChoice

	switch targetType {
	case prompt.TargetRepository:
		// Get worktrees for the repository
		worktrees, err := c.ListWorktrees(ListWorktreesOpts{
			RepositoryName: targetName,
		})
		if err != nil {
			return nil, fmt.Errorf("failed to list worktrees for repository %s: %w", targetName, err)
		}

		for _, worktree := range worktrees {
			choices = append(choices, prompt.TargetChoice{
				Type: prompt.TargetRepository, // Keep same type for consistency
				Name: worktree.Branch,
			})
		}
	case prompt.TargetWorkspace:
		// Get worktrees for the workspace
		worktrees, err := c.ListWorktrees(ListWorktreesOpts{
			WorkspaceName: targetName,
		})
		if err != nil {
			return nil, fmt.Errorf("failed to list worktrees for workspace %s: %w", targetName, err)
		}

		// Deduplicate worktrees by branch name to avoid duplicates
		seenBranches := make(map[string]bool)
		for _, worktree := range worktrees {
			if !seenBranches[worktree.Branch] {
				seenBranches[worktree.Branch] = true
				choices = append(choices, prompt.TargetChoice{
					Type: prompt.TargetWorkspace, // Keep same type for consistency
					Name: worktree.Branch,
				})
			}
		}
	default:
		return nil, fmt.Errorf("unknown target type: %s", targetType)
	}

	// Sort choices by name
	sort.Slice(choices, func(i, j int) bool {
		return choices[i].Name < choices[j].Name
	})

	return choices, nil
}

// promptSelectTargetOnly prompts the user to select a repository or workspace only (no worktree selection).
func (c *realCodeManager) promptSelectTargetOnly() (TargetSelectionResult, error) {
	return c.promptSelectTarget("", "target-only selection")
}

// promptSelectTarget is the unified method for target selection with filtering support.
// filterType can be "repository", "workspace", or "" (empty for both).
// context is used for logging purposes.
func (c *realCodeManager) promptSelectTarget(filterType, context string) (TargetSelectionResult, error) {
	// Prepare parameters for hooks
	params := map[string]interface{}{
		"showWorktreeLabel": false,
	}

	var selectedName, selectedType string
	err := c.executeWithHooks(consts.PromptSelectTarget, params, func() error {
		if c.deps.Logger != nil {
			c.deps.Logger.Logf("Building target choices for %s", context)
		}

		// Build choices based on filter type
		choices, err := c.buildTargetChoices(false, filterType)
		if err != nil {
			return fmt.Errorf("failed to build target choices: %w", err)
		}

		if len(choices) == 0 {
			return c.getNoChoicesError(filterType)
		}

		if c.deps.Logger != nil {
			c.deps.Logger.Logf("Prompting user to select target from %d choices", len(choices))
		}

		// Use the prompt package to get user selection
		selected, err := c.deps.Prompt.PromptSelectTarget(choices, false)
		if err != nil {
			return fmt.Errorf("failed to get target selection: %w", err)
		}

		selectedName = selected.Name
		selectedType = selected.Type

		if c.deps.Logger != nil {
			c.deps.Logger.Logf("User selected %s: %s", selectedType, selectedName)
		}

		return nil
	})

	if err != nil {
		return TargetSelectionResult{}, err
	}

	return TargetSelectionResult{
		Name:     selectedName,
		Type:     selectedType,
		Worktree: "", // No worktree for target-only selection
	}, nil
}

// getNoChoicesError returns an appropriate error message based on the filter type.
func (c *realCodeManager) getNoChoicesError(filterType string) error {
	switch filterType {
	case prompt.TargetRepository:
		return fmt.Errorf("no repositories available for selection")
	case prompt.TargetWorkspace:
		return fmt.Errorf("no workspaces available for selection")
	default:
		return fmt.Errorf("no repositories or workspaces available for selection")
	}
}

// buildTargetChoices builds a list of target choices from repositories and workspaces.
// filterType can be "repository", "workspace", or "" (empty for both).
func (c *realCodeManager) buildTargetChoices(showWorktreeLabel bool, filterType string) ([]prompt.TargetChoice, error) {
	var choices []prompt.TargetChoice

	// Add repositories (if not filtering to workspaces only)
	if filterType == "" || filterType == prompt.TargetRepository {
		repoChoices, err := c.buildRepositoryChoices(showWorktreeLabel)
		if err != nil {
			return nil, fmt.Errorf("failed to build repository choices: %w", err)
		}
		choices = append(choices, repoChoices...)
	}

	// Add workspaces (if not filtering to repositories only)
	if filterType == "" || filterType == prompt.TargetWorkspace {
		workspaceChoices, err := c.buildWorkspaceChoices(showWorktreeLabel)
		if err != nil {
			return nil, fmt.Errorf("failed to build workspace choices: %w", err)
		}
		choices = append(choices, workspaceChoices...)
	}

	// Sort choices
	c.sortChoices(choices, filterType)

	return choices, nil
}

// buildRepositoryChoices builds choices for repositories.
func (c *realCodeManager) buildRepositoryChoices(showWorktreeLabel bool) ([]prompt.TargetChoice, error) {
	repositories, err := c.ListRepositories()
	if err != nil {
		return nil, err
	}

	var choices []prompt.TargetChoice
	for _, repo := range repositories {
		choice := prompt.TargetChoice{
			Type: prompt.TargetRepository,
			Name: repo.Name,
		}

		if showWorktreeLabel {
			choice.Worktree = c.getFirstWorktreeForRepositorySafe(repo.Name)
		}

		choices = append(choices, choice)
	}

	return choices, nil
}

// buildWorkspaceChoices builds choices for workspaces.
func (c *realCodeManager) buildWorkspaceChoices(showWorktreeLabel bool) ([]prompt.TargetChoice, error) {
	workspaces, err := c.ListWorkspaces()
	if err != nil {
		return nil, err
	}

	var choices []prompt.TargetChoice
	for _, workspace := range workspaces {
		choice := prompt.TargetChoice{
			Type: prompt.TargetWorkspace,
			Name: workspace.Name,
		}

		if showWorktreeLabel {
			choice.Worktree = c.getFirstWorktreeForWorkspaceSafe(workspace.Name)
		}

		choices = append(choices, choice)
	}

	return choices, nil
}

// sortChoices sorts the choices based on filter type.
func (c *realCodeManager) sortChoices(choices []prompt.TargetChoice, filterType string) {
	if filterType == "" {
		// Mixed types: sort by type first, then by name
		sort.Slice(choices, func(i, j int) bool {
			if choices[i].Type != choices[j].Type {
				return choices[i].Type < choices[j].Type
			}
			return choices[i].Name < choices[j].Name
		})
	} else {
		// Single type: sort by name only
		sort.Slice(choices, func(i, j int) bool {
			return choices[i].Name < choices[j].Name
		})
	}
}

// promptSelectRepositoryOnly prompts the user to select a repository only.
func (c *realCodeManager) promptSelectRepositoryOnly() (TargetSelectionResult, error) {
	return c.promptSelectTarget(prompt.TargetRepository, "repository-only selection")
}

// promptSelectWorkspaceOnly prompts the user to select a workspace only.
func (c *realCodeManager) promptSelectWorkspaceOnly() (TargetSelectionResult, error) {
	return c.promptSelectTarget(prompt.TargetWorkspace, "workspace-only selection")
}

// getFirstWorktreeForRepository gets the first worktree alphabetically for a repository.
func (c *realCodeManager) getFirstWorktreeForRepository(repoName string) (string, error) {
	// Get worktrees for the repository
	worktrees, err := c.ListWorktrees(ListWorktreesOpts{
		RepositoryName: repoName,
	})
	if err != nil {
		return "", err
	}

	if len(worktrees) == 0 {
		return "", nil
	}

	// Sort worktrees by branch name and return the first one
	sort.Slice(worktrees, func(i, j int) bool {
		return worktrees[i].Branch < worktrees[j].Branch
	})

	return worktrees[0].Branch, nil
}

// getFirstWorktreeForWorkspace gets the first worktree alphabetically for a workspace.
func (c *realCodeManager) getFirstWorktreeForWorkspace(workspaceName string) (string, error) {
	// Get worktrees for the workspace
	worktrees, err := c.ListWorktrees(ListWorktreesOpts{
		WorkspaceName: workspaceName,
	})
	if err != nil {
		return "", err
	}

	if len(worktrees) == 0 {
		return "", nil
	}

	// Sort worktrees by branch name and return the first one
	sort.Slice(worktrees, func(i, j int) bool {
		return worktrees[i].Branch < worktrees[j].Branch
	})

	return worktrees[0].Branch, nil
}

// getFirstWorktreeForRepositorySafe gets the first worktree for a repository, returning empty string on error.
func (c *realCodeManager) getFirstWorktreeForRepositorySafe(repoName string) string {
	worktree, err := c.getFirstWorktreeForRepository(repoName)
	if err != nil {
		if c.deps.Logger != nil {
			c.deps.Logger.Logf("Failed to get worktree for repository %s: %v", repoName, err)
		}
		return ""
	}
	return worktree
}

// getFirstWorktreeForWorkspaceSafe gets the first worktree for a workspace, returning empty string on error.
func (c *realCodeManager) getFirstWorktreeForWorkspaceSafe(workspaceName string) string {
	worktree, err := c.getFirstWorktreeForWorkspace(workspaceName)
	if err != nil {
		if c.deps.Logger != nil {
			c.deps.Logger.Logf("Failed to get worktree for workspace %s: %v", workspaceName, err)
		}
		return ""
	}
	return worktree
}
