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

func TestAddWorktree(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockFS := fsmocks.NewMockFS(ctrl)

	cfg := config.Config{
		RepositoriesDir: "/home/user/.cm",
		StatusFile:      "/home/user/.cmstatus.yaml",
	}

	manager := &realManager{
		fs:     mockFS,
		config: cfg,
	}

	// Test data
	repoURL := "github.com/octocat/Hello-World"
	branch := "feature-a"
	remote := "origin"
	worktreePath := "/home/user/.cmrepos/github.com/octocat/Hello-World/feature-a"
	workspacePath := ""

	// Expected status file content
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
					"origin:feature-a": {
						Remote: remote,
						Branch: branch,
					},
				},
			},
		},
		Workspaces: make(map[string]Workspace),
	}

	expectedData, _ := yaml.Marshal(expectedStatus)

	// Mock expectations
	mockFS.EXPECT().Exists("/home/user/.cmstatus.yaml").Return(true, nil)
	mockFS.EXPECT().ReadFile("/home/user/.cmstatus.yaml").Return([]byte(`initialized: true
repositories:
  github.com/octocat/Hello-World:
    path: /home/user/.cmrepos/github.com/octocat/Hello-World/origin/main
    remotes:
      origin:
        default_branch: main
    worktrees: {}
workspaces: {}`), nil)
	mockFS.EXPECT().FileLock("/home/user/.cmstatus.yaml").Return(func() {}, nil)
	mockFS.EXPECT().WriteFileAtomic("/home/user/.cmstatus.yaml", expectedData, gomock.Any()).Return(nil)

	// Execute
	err := manager.AddWorktree(AddWorktreeParams{
		RepoURL:       repoURL,
		Branch:        branch,
		WorktreePath:  worktreePath,
		WorkspacePath: workspacePath,
		Remote:        remote,
	})

	// Assert
	assert.NoError(t, err)
}

func TestAddWorktree_Duplicate(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockFS := fsmocks.NewMockFS(ctrl)

	cfg := config.Config{
		RepositoriesDir: "/home/user/.cm",
		StatusFile:      "/home/user/.cmstatus.yaml",
	}

	manager := &realManager{
		fs:     mockFS,
		config: cfg,
	}

	// Test data
	repoURL := "github.com/octocat/Hello-World"
	branch := "feature-a"
	remote := "origin"

	// Existing status with duplicate
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
						Remote: remote,
						Branch: branch,
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
	err := manager.AddWorktree(AddWorktreeParams{
		RepoURL: repoURL,
		Branch:  branch,
		Remote:  remote,
	})

	// Assert
	assert.ErrorIs(t, err, ErrWorktreeAlreadyExists)
}

func TestAddWorktree_RepositoryNotFound(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockFS := fsmocks.NewMockFS(ctrl)

	cfg := config.Config{
		RepositoriesDir: "/home/user/.cm",
		StatusFile:      "/home/user/.cmstatus.yaml",
	}

	manager := &realManager{
		fs:     mockFS,
		config: cfg,
	}

	// Test data
	repoURL := "github.com/octocat/Hello-World"
	branch := "feature-a"
	remote := "origin"
	worktreePath := "/home/user/.cmrepos/github.com/octocat/Hello-World/origin/feature-a"

	// Existing status without the repository
	existingStatus := &Status{
		Repositories: make(map[string]Repository),
		Workspaces:   make(map[string]Workspace),
	}

	existingData, _ := yaml.Marshal(existingStatus)

	// Mock expectations for status file operations
	mockFS.EXPECT().Exists("/home/user/.cmstatus.yaml").Return(true, nil)
	mockFS.EXPECT().ReadFile("/home/user/.cmstatus.yaml").Return(existingData, nil)

	// Execute
	err := manager.AddWorktree(AddWorktreeParams{
		RepoURL:      repoURL,
		Branch:       branch,
		Remote:       remote,
		WorktreePath: worktreePath,
	})

	// Assert
	assert.ErrorIs(t, err, ErrRepositoryNotFound)
}
