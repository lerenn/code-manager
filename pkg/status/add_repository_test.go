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

func TestAddRepository(t *testing.T) {
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
	repoURL := "github.com/octocat/Hello-World"
	params := AddRepositoryParams{
		Path: "/home/user/.cmrepos/github.com/octocat/Hello-World/origin/main",
		Remotes: map[string]Remote{
			"origin": {
				DefaultBranch: "main",
			},
		},
	}

	// Expected repository
	expectedRepo := Repository{
		Path: "/home/user/.cmrepos/github.com/octocat/Hello-World/origin/main",
		Remotes: map[string]Remote{
			"origin": {
				DefaultBranch: "main",
			},
		},
		Worktrees: make(map[string]WorktreeInfo),
	}

	// Existing status
	existingStatus := &Status{
		Repositories: make(map[string]Repository),
		Workspaces:   make(map[string]Workspace),
	}

	existingData, _ := yaml.Marshal(existingStatus)

	// Expected status after addition
	expectedStatus := &Status{
		Repositories: map[string]Repository{
			repoURL: expectedRepo,
		},
		Workspaces: make(map[string]Workspace),
	}

	expectedData, _ := yaml.Marshal(expectedStatus)

	// Mock expectations
	mockFS.EXPECT().Exists("/home/user/.cmstatus.yaml").Return(true, nil)
	mockFS.EXPECT().ReadFile("/home/user/.cmstatus.yaml").Return(existingData, nil)
	mockFS.EXPECT().FileLock("/home/user/.cmstatus.yaml").Return(func() {}, nil)
	mockFS.EXPECT().WriteFileAtomic("/home/user/.cmstatus.yaml", expectedData, gomock.Any()).Return(nil)

	// Execute
	err := manager.AddRepository(repoURL, params)

	// Assert
	assert.NoError(t, err)
}

func TestAddRepository_AlreadyExists(t *testing.T) {
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
	repoURL := "github.com/octocat/Hello-World"
	params := AddRepositoryParams{
		Path: "/home/user/.cmrepos/github.com/octocat/Hello-World/origin/main",
		Remotes: map[string]Remote{
			"origin": {
				DefaultBranch: "main",
			},
		},
	}

	// Existing repository
	existingRepo := Repository{
		Path: "/home/user/.cmrepos/github.com/octocat/Hello-World/origin/main",
		Remotes: map[string]Remote{
			"origin": {
				DefaultBranch: "main",
			},
		},
		Worktrees: make(map[string]WorktreeInfo),
	}

	// Existing status with repository
	existingStatus := &Status{
		Repositories: map[string]Repository{
			repoURL: existingRepo,
		},
		Workspaces: make(map[string]Workspace),
	}

	existingData, _ := yaml.Marshal(existingStatus)

	// Mock expectations
	mockFS.EXPECT().Exists("/home/user/.cmstatus.yaml").Return(true, nil)
	mockFS.EXPECT().ReadFile("/home/user/.cmstatus.yaml").Return(existingData, nil)

	// Execute
	err := manager.AddRepository(repoURL, params)

	// Assert
	assert.ErrorIs(t, err, ErrRepositoryAlreadyExists)
}
