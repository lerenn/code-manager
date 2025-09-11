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

func TestGetWorkspace(t *testing.T) {
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
	expectedWorkspace := Workspace{
		Worktree:     []string{"origin:feature-a"},
		Repositories: []string{"github.com/octocat/Hello-World", "github.com/lerenn/other"},
	}

	// Existing status
	existingStatus := &Status{
		Repositories: make(map[string]Repository),
		Workspaces: map[string]Workspace{
			workspacePath: expectedWorkspace,
		},
	}

	existingData, _ := yaml.Marshal(existingStatus)

	// Mock expectations
	mockFS.EXPECT().Exists("/home/user/.cmstatus.yaml").Return(true, nil)
	mockFS.EXPECT().ReadFile("/home/user/.cmstatus.yaml").Return(existingData, nil)

	// Execute
	workspace, err := manager.GetWorkspace(workspacePath)

	// Assert
	assert.NoError(t, err)
	assert.Equal(t, &expectedWorkspace, workspace)
}

func TestGetWorkspace_NotFound(t *testing.T) {
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

	// Existing status without the workspace
	existingStatus := &Status{
		Repositories: make(map[string]Repository),
		Workspaces:   make(map[string]Workspace),
	}

	existingData, _ := yaml.Marshal(existingStatus)

	// Mock expectations
	mockFS.EXPECT().Exists("/home/user/.cmstatus.yaml").Return(true, nil)
	mockFS.EXPECT().ReadFile("/home/user/.cmstatus.yaml").Return(existingData, nil)

	// Execute
	workspace, err := manager.GetWorkspace(workspacePath)

	// Assert
	assert.Nil(t, workspace)
	assert.ErrorIs(t, err, ErrWorkspaceNotFound)
}
