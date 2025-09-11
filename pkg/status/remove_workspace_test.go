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

func TestRemoveWorkspace(t *testing.T) {
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
	workspaceName := "my-workspace"
	workspacePath := "/home/user/my-workspace.code-workspace"

	// Existing status with workspace to remove
	existingStatus := &Status{
		Repositories: map[string]Repository{
			"github.com/octocat/Hello-World": {
				Path: "/home/user/.cm/github.com/octocat/Hello-World",
				Remotes: map[string]Remote{
					"origin": {DefaultBranch: "main"},
				},
				Worktrees: make(map[string]WorktreeInfo),
			},
		},
		Workspaces: map[string]Workspace{
			workspacePath: {
				Worktree:     []string{"origin:feature-branch"},
				Repositories: []string{"github.com/octocat/Hello-World"},
			},
		},
	}

	// Expected status after removal
	expectedStatus := &Status{
		Repositories: map[string]Repository{
			"github.com/octocat/Hello-World": {
				Path: "/home/user/.cm/github.com/octocat/Hello-World",
				Remotes: map[string]Remote{
					"origin": {DefaultBranch: "main"},
				},
				Worktrees: make(map[string]WorktreeInfo),
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
	err := manager.RemoveWorkspace(workspaceName)

	// Assert
	assert.NoError(t, err)
}

func TestRemoveWorkspace_NotFound(t *testing.T) {
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
	workspaceName := "non-existent-workspace"

	// Existing status without the workspace
	existingStatus := &Status{
		Repositories: make(map[string]Repository),
		Workspaces: map[string]Workspace{
			"/home/user/other-workspace.code-workspace": {
				Worktree:     []string{"origin:feature-branch"},
				Repositories: []string{"github.com/octocat/Hello-World"},
			},
		},
	}

	existingData, _ := yaml.Marshal(existingStatus)

	// Mock expectations
	mockFS.EXPECT().Exists("/home/user/.cmstatus.yaml").Return(true, nil)
	mockFS.EXPECT().ReadFile("/home/user/.cmstatus.yaml").Return(existingData, nil)

	// Execute
	err := manager.RemoveWorkspace(workspaceName)

	// Assert
	assert.ErrorIs(t, err, ErrWorkspaceNotFound)
}

func TestGetWorkspaceByName(t *testing.T) {
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
	workspaceName := "my-workspace"
	workspacePath := "/home/user/my-workspace.code-workspace"
	expectedWorkspace := Workspace{
		Worktree:     []string{"origin:feature-branch"},
		Repositories: []string{"github.com/octocat/Hello-World"},
	}

	// Existing status with workspace
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
	workspace, err := manager.GetWorkspaceByName(workspaceName)

	// Assert
	assert.NoError(t, err)
	assert.Equal(t, &expectedWorkspace, workspace)
}

func TestGetWorkspaceByName_NotFound(t *testing.T) {
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
	workspaceName := "non-existent-workspace"

	// Existing status without the workspace
	existingStatus := &Status{
		Repositories: make(map[string]Repository),
		Workspaces: map[string]Workspace{
			"/home/user/other-workspace.code-workspace": {
				Worktree:     []string{"origin:feature-branch"},
				Repositories: []string{"github.com/octocat/Hello-World"},
			},
		},
	}

	existingData, _ := yaml.Marshal(existingStatus)

	// Mock expectations
	mockFS.EXPECT().Exists("/home/user/.cmstatus.yaml").Return(true, nil)
	mockFS.EXPECT().ReadFile("/home/user/.cmstatus.yaml").Return(existingData, nil)

	// Execute
	workspace, err := manager.GetWorkspaceByName(workspaceName)

	// Assert
	assert.ErrorIs(t, err, ErrWorkspaceNotFound)
	assert.Nil(t, workspace)
}

func TestRemoveWorkspace_LoadStatusError(t *testing.T) {
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
	workspaceName := "my-workspace"

	// Mock expectations - simulate file read error
	mockFS.EXPECT().Exists("/home/user/.cmstatus.yaml").Return(true, nil)
	mockFS.EXPECT().ReadFile("/home/user/.cmstatus.yaml").Return(nil, assert.AnError)

	// Execute
	err := manager.RemoveWorkspace(workspaceName)

	// Assert
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "failed to load status")
}

func TestGetWorkspaceByName_LoadStatusError(t *testing.T) {
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
	workspaceName := "my-workspace"

	// Mock expectations - simulate file read error
	mockFS.EXPECT().Exists("/home/user/.cmstatus.yaml").Return(true, nil)
	mockFS.EXPECT().ReadFile("/home/user/.cmstatus.yaml").Return(nil, assert.AnError)

	// Execute
	workspace, err := manager.GetWorkspaceByName(workspaceName)

	// Assert
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "failed to load status")
	assert.Nil(t, workspace)
}

func TestRemoveWorkspace_SaveStatusError(t *testing.T) {
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
	workspaceName := "my-workspace"
	workspacePath := "/home/user/my-workspace.code-workspace"

	// Existing status with workspace to remove
	existingStatus := &Status{
		Repositories: make(map[string]Repository),
		Workspaces: map[string]Workspace{
			workspacePath: {
				Worktree:     []string{"origin:feature-branch"},
				Repositories: []string{"github.com/octocat/Hello-World"},
			},
		},
	}

	existingData, _ := yaml.Marshal(existingStatus)

	// Mock expectations - simulate save error
	mockFS.EXPECT().Exists("/home/user/.cmstatus.yaml").Return(true, nil)
	mockFS.EXPECT().ReadFile("/home/user/.cmstatus.yaml").Return(existingData, nil)
	mockFS.EXPECT().FileLock("/home/user/.cmstatus.yaml").Return(func() {}, nil)
	mockFS.EXPECT().WriteFileAtomic("/home/user/.cmstatus.yaml", gomock.Any(), gomock.Any()).Return(assert.AnError)

	// Execute
	err := manager.RemoveWorkspace(workspaceName)

	// Assert
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "failed to save status")
}
