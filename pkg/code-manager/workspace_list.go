package codemanager

import (
	"fmt"
	"sort"

	"github.com/lerenn/code-manager/pkg/code-manager/consts"
)

// WorkspaceInfo contains information about a workspace for display purposes.
type WorkspaceInfo struct {
	Name         string
	Repositories []string
	Worktrees    []string
}

// ListWorkspaces lists all workspaces from the status file.
func (c *realCodeManager) ListWorkspaces() ([]WorkspaceInfo, error) {
	// Prepare parameters for hooks
	params := map[string]interface{}{}

	// Execute with hooks
	return c.executeWithHooksAndReturnWorkspaces(consts.ListWorkspaces, params, func() ([]WorkspaceInfo, error) {
		if c.deps.Logger != nil {
			c.deps.Logger.Logf("Loading workspaces from status file")
		}

		// Get all workspaces from status manager
		workspaces, err := c.deps.StatusManager.ListWorkspaces()
		if err != nil {
			return nil, fmt.Errorf("failed to load workspaces: %w", err)
		}

		if c.deps.Logger != nil {
			c.deps.Logger.Logf("Formatting workspace list")
		}

		// Convert to WorkspaceInfo slice
		var workspaceInfos []WorkspaceInfo
		for workspaceName, workspace := range workspaces {
			workspaceInfo := WorkspaceInfo{
				Name:         workspaceName,
				Repositories: workspace.Repositories,
				Worktrees:    workspace.Worktrees,
			}
			workspaceInfos = append(workspaceInfos, workspaceInfo)
		}

		// Sort workspaces by name for consistent ordering
		sort.Slice(workspaceInfos, func(i, j int) bool {
			return workspaceInfos[i].Name < workspaceInfos[j].Name
		})

		if c.deps.Logger != nil {
			c.deps.Logger.Logf("Found %d workspaces", len(workspaceInfos))
		}

		return workspaceInfos, nil
	})
}
