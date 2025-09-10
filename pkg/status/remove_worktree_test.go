//go:build unit

package status

import (
	"testing"

	"github.com/lerenn/code-manager/pkg/config"
	fsmocks "github.com/lerenn/code-manager/pkg/fs/mocks"
	"github.com/stretchr/testify/assert"
	"go.uber.org/mock/gomock"
	"gopkg.in/yaml.v3"
)

func TestRemoveWorktree(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockFS := fsmocks.NewMockFS(ctrl)

	cfg := config.Config{
		BasePath:   "/home/user/.cm",
		StatusFile: "/home/user/.cmstatus.yaml",
	}

	manager := &realManager{
		fs:     mockFS,
		config: cfg,
	}

	// Test data
	repoURL := "github.com/octocat/Hello-World"
	branch := "feature-a"

	// Existing status with the worktree to remove
	existingStatus := &Status{
		Repositories: map[string]Repository{
			repoURL: {
				Path: "/home/user/.cmrepos/github.com/octocat/Hello-World/origin/main",
				Remotes: map[string]Remote{
					"origin": {
						DefaultBranch: "main",
					},
				},
				Worktrees: map[string]WorktreeInfo{
					"origin:feature-a": {
						Remote: "origin",
						Branch: branch,
					},
					"origin:feature-b": {
						Remote: "origin",
						Branch: "feature-b",
					},
				},
			},
		},
		Workspaces: make(map[string]Workspace),
	}

	// Expected status after removal
	expectedStatus := &Status{
		Repositories: map[string]Repository{
			repoURL: {
				Path: "/home/user/.cmrepos/github.com/octocat/Hello-World/origin/main",
				Remotes: map[string]Remote{
					"origin": {
						DefaultBranch: "main",
					},
				},
				Worktrees: map[string]WorktreeInfo{
					"origin:feature-b": {
						Remote: "origin",
						Branch: "feature-b",
					},
				},
			},
		},
		Workspaces: make(map[string]Workspace),
	}

	existingData, _ := yaml.Marshal(existingStatus)
	expectedData, _ := yaml.Marshal(expectedStatus)

	// Mock expectations
	mockFS.EXPECT().Exists("/home/user/.cmstatus.yaml").Return(true, nil)
	mockFS.EXPECT().ReadFile("/home/user/.cmstatus.yaml").Return(existingData, nil)
	mockFS.EXPECT().FileLock("/home/user/.cmstatus.yaml").Return(func() {}, nil)
	mockFS.EXPECT().WriteFileAtomic("/home/user/.cmstatus.yaml", expectedData, gomock.Any()).Return(nil)

	// Execute
	err := manager.RemoveWorktree(repoURL, branch)

	// Assert
	assert.NoError(t, err)
}

func TestRemoveWorktree_NotFound(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockFS := fsmocks.NewMockFS(ctrl)

	cfg := config.Config{
		BasePath:   "/home/user/.cm",
		StatusFile: "/home/user/.cmstatus.yaml",
	}

	manager := &realManager{
		fs:     mockFS,
		config: cfg,
	}

	// Test data
	repoURL := "github.com/octocat/Hello-World"
	branch := "feature-a"

	// Existing status without the worktree to remove
	existingStatus := &Status{
		Repositories: map[string]Repository{
			repoURL: {
				Path: "/home/user/.cmrepos/github.com/octocat/Hello-World/origin/main",
				Remotes: map[string]Remote{
					"origin": {
						DefaultBranch: "main",
					},
				},
				Worktrees: map[string]WorktreeInfo{
					"origin:feature-b": {
						Remote: "origin",
						Branch: "feature-b",
					},
				},
			},
		},
		Workspaces: make(map[string]Workspace),
	}

	existingData, _ := yaml.Marshal(existingStatus)

	// Mock expectations
	mockFS.EXPECT().Exists("/home/user/.cmstatus.yaml").Return(true, nil)
	mockFS.EXPECT().ReadFile("/home/user/.cmstatus.yaml").Return(existingData, nil)

	// Execute
	err := manager.RemoveWorktree(repoURL, branch)

	// Assert
	assert.ErrorIs(t, err, ErrWorktreeNotFound)
}

func TestRemoveWorktree_RepositoryNotFound(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockFS := fsmocks.NewMockFS(ctrl)

	cfg := config.Config{
		BasePath:   "/home/user/.cm",
		StatusFile: "/home/user/.cmstatus.yaml",
	}

	manager := &realManager{
		fs:     mockFS,
		config: cfg,
	}

	// Test data
	repoURL := "github.com/octocat/Hello-World"
	branch := "feature-a"

	// Existing status without the repository
	existingStatus := &Status{
		Repositories: make(map[string]Repository),
		Workspaces:   make(map[string]Workspace),
	}

	existingData, _ := yaml.Marshal(existingStatus)

	// Mock expectations
	mockFS.EXPECT().Exists("/home/user/.cmstatus.yaml").Return(true, nil)
	mockFS.EXPECT().ReadFile("/home/user/.cmstatus.yaml").Return(existingData, nil)

	// Execute
	err := manager.RemoveWorktree(repoURL, branch)

	// Assert
	assert.ErrorIs(t, err, ErrRepositoryNotFound)
}
