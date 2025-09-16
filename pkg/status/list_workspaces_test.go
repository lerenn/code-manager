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

func TestListWorkspaces_Success(t *testing.T) {
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
	expectedWorkspaces := map[string]Workspace{
		"workspace1": {
			Worktrees:    []string{"worktree1"},
			Repositories: []string{"github.com/test/repo1"},
		},
		"workspace2": {
			Worktrees:    []string{"worktree2", "worktree3"},
			Repositories: []string{"github.com/test/repo2"},
		},
	}

	status := &Status{
		Repositories: make(map[string]Repository),
		Workspaces:   expectedWorkspaces,
	}

	statusData, _ := yaml.Marshal(status)

	// Mock expectations
	mockFS.EXPECT().Exists(cfg.StatusFile).Return(true, nil)
	mockFS.EXPECT().ReadFile(cfg.StatusFile).Return(statusData, nil)

	// Execute
	workspaces, err := manager.ListWorkspaces()

	// Verify
	assert.NoError(t, err)
	assert.Equal(t, expectedWorkspaces, workspaces)
}

func TestListWorkspaces_EmptyWorkspaces(t *testing.T) {
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

	// Test data - empty workspaces
	status := &Status{
		Repositories: make(map[string]Repository),
		Workspaces:   make(map[string]Workspace),
	}

	statusData, _ := yaml.Marshal(status)

	// Mock expectations
	mockFS.EXPECT().Exists(cfg.StatusFile).Return(true, nil)
	mockFS.EXPECT().ReadFile(cfg.StatusFile).Return(statusData, nil)

	// Execute
	workspaces, err := manager.ListWorkspaces()

	// Verify
	assert.NoError(t, err)
	assert.Empty(t, workspaces)
}

func TestListWorkspaces_LoadStatusFailure(t *testing.T) {
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

	// Mock expectations - file doesn't exist, so loadStatus will create it
	mockFS.EXPECT().Exists(cfg.StatusFile).Return(false, nil)
	mockFS.EXPECT().FileLock(cfg.StatusFile).Return(func() {}, nil)
	mockFS.EXPECT().WriteFileAtomic(cfg.StatusFile, gomock.Any(), gomock.Any()).Return(nil)

	// Execute
	workspaces, err := manager.ListWorkspaces()

	// Verify
	assert.NoError(t, err)
	assert.Empty(t, workspaces)
}

func TestListWorkspaces_FileReadFailure(t *testing.T) {
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

	// Mock expectations
	mockFS.EXPECT().Exists(cfg.StatusFile).Return(true, nil)
	mockFS.EXPECT().ReadFile(cfg.StatusFile).Return(nil, assert.AnError)

	// Execute
	workspaces, err := manager.ListWorkspaces()

	// Verify
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "failed to load status")
	assert.Nil(t, workspaces)
}

func TestListWorkspaces_FileExistsFailure(t *testing.T) {
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

	// Mock expectations - file existence check fails
	mockFS.EXPECT().Exists(cfg.StatusFile).Return(false, assert.AnError)

	// Execute
	workspaces, err := manager.ListWorkspaces()

	// Verify
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "failed to check status file existence")
	assert.Nil(t, workspaces)
}

func TestListWorkspaces_WithRepositories(t *testing.T) {
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

	// Test data - workspaces with repositories
	expectedWorkspaces := map[string]Workspace{
		"workspace1": {
			Worktrees:    []string{},
			Repositories: []string{"github.com/test/repo1", "github.com/test/repo2"},
		},
	}

	status := &Status{
		Repositories: map[string]Repository{
			"github.com/test/repo1": {
				Path:      "/path/to/repo1",
				Remotes:   make(map[string]Remote),
				Worktrees: make(map[string]WorktreeInfo),
			},
			"github.com/test/repo2": {
				Path:      "/path/to/repo2",
				Remotes:   make(map[string]Remote),
				Worktrees: make(map[string]WorktreeInfo),
			},
		},
		Workspaces: expectedWorkspaces,
	}

	statusData, _ := yaml.Marshal(status)

	// Mock expectations
	mockFS.EXPECT().Exists(cfg.StatusFile).Return(true, nil)
	mockFS.EXPECT().ReadFile(cfg.StatusFile).Return(statusData, nil)

	// Execute
	workspaces, err := manager.ListWorkspaces()

	// Verify
	assert.NoError(t, err)
	assert.Equal(t, expectedWorkspaces, workspaces)
}
