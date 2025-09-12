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

func TestRemoveWorkspace_Success(t *testing.T) {
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
	workspaceName := "test-workspace"
	initialStatus := &Status{
		Repositories: make(map[string]Repository),
		Workspaces: map[string]Workspace{
			workspaceName: {
				Worktrees:    []string{"test-worktree"},
				Repositories: []string{"github.com/test/repo"},
			},
		},
	}

	// Expected status after removal
	expectedStatus := &Status{
		Repositories: make(map[string]Repository),
		Workspaces:   make(map[string]Workspace),
	}

	initialData, _ := yaml.Marshal(initialStatus)
	expectedData, _ := yaml.Marshal(expectedStatus)

	// Mock expectations
	mockFS.EXPECT().Exists(cfg.StatusFile).Return(true, nil)
	mockFS.EXPECT().ReadFile(cfg.StatusFile).Return(initialData, nil)
	mockFS.EXPECT().FileLock(cfg.StatusFile).Return(func() {}, nil)
	mockFS.EXPECT().WriteFileAtomic(cfg.StatusFile, expectedData, gomock.Any()).Return(nil)

	// Execute
	err := manager.RemoveWorkspace(workspaceName)

	// Verify
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

	// Test data - workspace doesn't exist
	workspaceName := "non-existent-workspace"
	initialStatus := &Status{
		Repositories: make(map[string]Repository),
		Workspaces:   make(map[string]Workspace),
	}

	initialData, _ := yaml.Marshal(initialStatus)

	// Mock expectations
	mockFS.EXPECT().Exists(cfg.StatusFile).Return(true, nil)
	mockFS.EXPECT().ReadFile(cfg.StatusFile).Return(initialData, nil)

	// Execute
	err := manager.RemoveWorkspace(workspaceName)

	// Verify
	assert.Error(t, err)
	assert.ErrorIs(t, err, ErrWorkspaceNotFound)
}

func TestRemoveWorkspace_LoadStatusFailure(t *testing.T) {
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

	workspaceName := "test-workspace"

	// Mock expectations - file doesn't exist, so loadStatus will create it
	mockFS.EXPECT().Exists(cfg.StatusFile).Return(false, nil)
	mockFS.EXPECT().FileLock(cfg.StatusFile).Return(func() {}, nil)
	mockFS.EXPECT().WriteFileAtomic(cfg.StatusFile, gomock.Any(), gomock.Any()).Return(nil)

	// Execute
	err := manager.RemoveWorkspace(workspaceName)

	// Verify
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "workspace not found in status")
}

func TestRemoveWorkspace_SaveStatusFailure(t *testing.T) {
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
	workspaceName := "test-workspace"
	initialStatus := &Status{
		Repositories: make(map[string]Repository),
		Workspaces: map[string]Workspace{
			workspaceName: {
				Worktrees:    []string{"test-worktree"},
				Repositories: []string{"github.com/test/repo"},
			},
		},
	}

	// Expected status after removal
	expectedStatus := &Status{
		Repositories: make(map[string]Repository),
		Workspaces:   make(map[string]Workspace),
	}

	initialData, _ := yaml.Marshal(initialStatus)
	expectedData, _ := yaml.Marshal(expectedStatus)

	// Mock expectations
	mockFS.EXPECT().Exists(cfg.StatusFile).Return(true, nil)
	mockFS.EXPECT().ReadFile(cfg.StatusFile).Return(initialData, nil)
	mockFS.EXPECT().FileLock(cfg.StatusFile).Return(func() {}, nil)
	mockFS.EXPECT().WriteFileAtomic(cfg.StatusFile, expectedData, gomock.Any()).Return(assert.AnError)

	// Execute
	err := manager.RemoveWorkspace(workspaceName)

	// Verify
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "failed to save status")
}

func TestRemoveWorkspace_FileReadFailure(t *testing.T) {
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

	workspaceName := "test-workspace"

	// Mock expectations
	mockFS.EXPECT().Exists(cfg.StatusFile).Return(true, nil)
	mockFS.EXPECT().ReadFile(cfg.StatusFile).Return(nil, assert.AnError)

	// Execute
	err := manager.RemoveWorkspace(workspaceName)

	// Verify
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "failed to load status")
}

func TestRemoveWorkspace_MultipleWorkspaces(t *testing.T) {
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

	// Test data - multiple workspaces
	workspaceName := "workspace-to-delete"
	initialStatus := &Status{
		Repositories: make(map[string]Repository),
		Workspaces: map[string]Workspace{
			workspaceName: {
				Worktrees:    []string{"test-worktree"},
				Repositories: []string{"github.com/test/repo"},
			},
			"other-workspace": {
				Worktrees:    []string{"other-worktree"},
				Repositories: []string{"github.com/other/repo"},
			},
		},
	}

	// Expected status after removal - other workspace should remain
	expectedStatus := &Status{
		Repositories: make(map[string]Repository),
		Workspaces: map[string]Workspace{
			"other-workspace": {
				Worktrees:    []string{"other-worktree"},
				Repositories: []string{"github.com/other/repo"},
			},
		},
	}

	initialData, _ := yaml.Marshal(initialStatus)
	expectedData, _ := yaml.Marshal(expectedStatus)

	// Mock expectations
	mockFS.EXPECT().Exists(cfg.StatusFile).Return(true, nil)
	mockFS.EXPECT().ReadFile(cfg.StatusFile).Return(initialData, nil)
	mockFS.EXPECT().FileLock(cfg.StatusFile).Return(func() {}, nil)
	mockFS.EXPECT().WriteFileAtomic(cfg.StatusFile, expectedData, gomock.Any()).Return(nil)

	// Execute
	err := manager.RemoveWorkspace(workspaceName)

	// Verify
	assert.NoError(t, err)
}
