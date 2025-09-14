//go:build unit

package status

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestWorkspace_HasRepository(t *testing.T) {
	tests := []struct {
		name           string
		workspace      Workspace
		repositoryName string
		expected       bool
	}{
		{
			name: "repository exists in workspace",
			workspace: Workspace{
				Repositories: []string{"github.com/user/repo1", "github.com/user/repo2"},
			},
			repositoryName: "github.com/user/repo1",
			expected:       true,
		},
		{
			name: "repository does not exist in workspace",
			workspace: Workspace{
				Repositories: []string{"github.com/user/repo1", "github.com/user/repo2"},
			},
			repositoryName: "github.com/user/repo3",
			expected:       false,
		},
		{
			name: "empty workspace",
			workspace: Workspace{
				Repositories: []string{},
			},
			repositoryName: "github.com/user/repo1",
			expected:       false,
		},
		{
			name: "nil repositories slice",
			workspace: Workspace{
				Repositories: nil,
			},
			repositoryName: "github.com/user/repo1",
			expected:       false,
		},
		{
			name: "exact match required",
			workspace: Workspace{
				Repositories: []string{"github.com/user/repo1", "github.com/user/repo2"},
			},
			repositoryName: "github.com/user/repo",
			expected:       false,
		},
		{
			name: "case sensitive match",
			workspace: Workspace{
				Repositories: []string{"github.com/user/repo1", "github.com/user/repo2"},
			},
			repositoryName: "GITHUB.COM/USER/REPO1",
			expected:       false,
		},
		{
			name: "workspace with worktrees and repositories",
			workspace: Workspace{
				Worktrees:    []string{"worktree1", "worktree2"},
				Repositories: []string{"github.com/user/repo1", "github.com/user/repo2"},
			},
			repositoryName: "github.com/user/repo2",
			expected:       true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := tt.workspace.HasRepository(tt.repositoryName)
			assert.Equal(t, tt.expected, result)
		})
	}
}
