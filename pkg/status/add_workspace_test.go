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

func TestAddWorkspace(t *testing.T) {
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
	workspacePath := "/home/user/workspace.code-workspace"
	params := AddWorkspaceParams{
		Repositories: []string{"github.com/octocat/Hello-World", "github.com/lerenn/other"},
	}

	// Expected status file content
	expectedStatus := &Status{
		Repositories: make(map[string]Repository),
		Workspaces: map[string]Workspace{
			workspacePath: {
				Worktrees:    []string{}, // Empty initially, populated when worktrees are created
				Repositories: params.Repositories,
			},
		},
	}

	expectedData, _ := yaml.Marshal(expectedStatus)

	// Mock expectations
	mockFS.EXPECT().Exists("/home/user/.cmstatus.yaml").Return(true, nil)
	mockFS.EXPECT().ReadFile("/home/user/.cmstatus.yaml").Return([]byte(`initialized: true
repositories: {}
workspaces: {}`), nil)
	mockFS.EXPECT().FileLock("/home/user/.cmstatus.yaml").Return(func() {}, nil)
	mockFS.EXPECT().WriteFileAtomic("/home/user/.cmstatus.yaml", expectedData, gomock.Any()).Return(nil)

	// Execute
	err := manager.AddWorkspace(workspacePath, params)

	// Assert
	assert.NoError(t, err)
}

func TestAddWorkspace_Duplicate(t *testing.T) {
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
	workspacePath := "/home/user/workspace.code-workspace"
	params := AddWorkspaceParams{
		Repositories: []string{"github.com/octocat/Hello-World"},
	}

	// Existing status with duplicate workspace
	existingStatus := &Status{
		Repositories: make(map[string]Repository),
		Workspaces: map[string]Workspace{
			workspacePath: {
				Worktrees:    []string{"origin:feature-b"},
				Repositories: []string{"github.com/lerenn/other"},
			},
		},
	}

	existingData, _ := yaml.Marshal(existingStatus)

	// Mock expectations
	mockFS.EXPECT().Exists("/home/user/.cmstatus.yaml").Return(true, nil)
	mockFS.EXPECT().ReadFile("/home/user/.cmstatus.yaml").Return(existingData, nil)

	// Execute
	err := manager.AddWorkspace(workspacePath, params)

	// Assert
	assert.ErrorIs(t, err, ErrWorkspaceAlreadyExists)
}
