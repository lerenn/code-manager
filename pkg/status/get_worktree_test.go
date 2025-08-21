//go:build unit

package status

import (
	"testing"

	"github.com/lerenn/code-manager/pkg/config"
	"github.com/lerenn/code-manager/pkg/fs"
	"github.com/stretchr/testify/assert"
	"go.uber.org/mock/gomock"
	"gopkg.in/yaml.v3"
)

func TestGetWorktree(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockFS := fs.NewMockFS(ctrl)

	cfg := &config.Config{
		BasePath:   "/home/user/.cm",
		StatusFile: "/home/user/.cmstatus.yaml",
	}

	manager := &realManager{
		fs:     mockFS,
		config: cfg,
	}

	// Test data
	repoURL := "github.com/lerenn/example"
	branch := "feature-a"
	expectedWorktree := WorktreeInfo{
		Remote: "origin",
		Branch: branch,
	}

	// Existing status
	existingStatus := &Status{
		Repositories: map[string]Repository{
			repoURL: {
				Path: "/home/user/.cmrepos/github.com/lerenn/example/origin/main",
				Remotes: map[string]Remote{
					"origin": {
						DefaultBranch: "main",
					},
				},
				Worktrees: map[string]WorktreeInfo{
					"origin:feature-a": expectedWorktree,
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
	worktree, err := manager.GetWorktree(repoURL, branch)

	// Assert
	assert.NoError(t, err)
	assert.Equal(t, &expectedWorktree, worktree)
}

func TestGetWorktree_NotFound(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockFS := fs.NewMockFS(ctrl)

	cfg := &config.Config{
		BasePath:   "/home/user/.cm",
		StatusFile: "/home/user/.cmstatus.yaml",
	}

	manager := &realManager{
		fs:     mockFS,
		config: cfg,
	}

	// Test data
	repoURL := "github.com/lerenn/example"
	branch := "feature-a"

	// Existing status without the requested worktree
	existingStatus := &Status{
		Repositories: map[string]Repository{
			repoURL: {
				Path: "/home/user/.cmrepos/github.com/lerenn/example/origin/main",
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
	worktree, err := manager.GetWorktree(repoURL, branch)

	// Assert
	assert.Nil(t, worktree)
	assert.ErrorIs(t, err, ErrWorktreeNotFound)
}

func TestGetWorktree_RepositoryNotFound(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockFS := fs.NewMockFS(ctrl)

	cfg := &config.Config{
		BasePath:   "/home/user/.cm",
		StatusFile: "/home/user/.cmstatus.yaml",
	}

	manager := &realManager{
		fs:     mockFS,
		config: cfg,
	}

	// Test data
	repoURL := "github.com/lerenn/example"
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
	worktree, err := manager.GetWorktree(repoURL, branch)

	// Assert
	assert.Nil(t, worktree)
	assert.ErrorIs(t, err, ErrRepositoryNotFound)
}
