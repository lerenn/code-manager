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

func TestListAllWorktrees(t *testing.T) {
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

	// Expected worktrees
	expectedWorktrees := []WorktreeInfo{
		{
			Remote: "origin",
			Branch: "feature-a",
		},
		{
			Remote: "origin",
			Branch: "feature-b",
		},
	}

	// Existing status
	existingStatus := &Status{
		Repositories: map[string]Repository{
			"github.com/lerenn/example": {
				Path: "/home/user/.cmrepos/github.com/lerenn/example/origin/main",
				Remotes: map[string]Remote{
					"origin": {
						DefaultBranch: "main",
					},
				},
				Worktrees: map[string]WorktreeInfo{
					"origin:feature-a": expectedWorktrees[0],
					"origin:feature-b": expectedWorktrees[1],
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
	worktrees, err := manager.ListAllWorktrees()

	// Assert
	assert.NoError(t, err)
	assert.ElementsMatch(t, expectedWorktrees, worktrees)
}

func TestListAllWorktrees_Empty(t *testing.T) {
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

	// Empty status
	existingStatus := &Status{
		Repositories: make(map[string]Repository),
		Workspaces:   make(map[string]Workspace),
	}

	existingData, _ := yaml.Marshal(existingStatus)

	// Mock expectations
	mockFS.EXPECT().Exists("/home/user/.cmstatus.yaml").Return(true, nil)
	mockFS.EXPECT().ReadFile("/home/user/.cmstatus.yaml").Return(existingData, nil)

	// Execute
	worktrees, err := manager.ListAllWorktrees()

	// Assert
	assert.NoError(t, err)
	assert.Empty(t, worktrees)
}
