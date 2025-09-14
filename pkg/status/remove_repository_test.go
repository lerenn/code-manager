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

func TestRemoveRepository_Success(t *testing.T) {
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
	repoURL := "github.com/test/repo"
	initialStatus := &Status{
		Repositories: map[string]Repository{
			repoURL: {
				Path: "/path/to/repo",
				Remotes: map[string]Remote{
					"origin": {
						DefaultBranch: "main",
					},
				},
				Worktrees: map[string]WorktreeInfo{
					"main": {
						Remote: "origin",
						Branch: "main",
					},
				},
			},
		},
		Workspaces: make(map[string]Workspace),
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
	err := manager.RemoveRepository(repoURL)

	// Verify
	assert.NoError(t, err)
}

func TestRemoveRepository_NotFound(t *testing.T) {
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

	// Test data - repository doesn't exist
	repoURL := "github.com/non-existent/repo"
	initialStatus := &Status{
		Repositories: make(map[string]Repository),
		Workspaces:   make(map[string]Workspace),
	}

	initialData, _ := yaml.Marshal(initialStatus)

	// Mock expectations
	mockFS.EXPECT().Exists(cfg.StatusFile).Return(true, nil)
	mockFS.EXPECT().ReadFile(cfg.StatusFile).Return(initialData, nil)

	// Execute
	err := manager.RemoveRepository(repoURL)

	// Verify
	assert.Error(t, err)
	assert.ErrorIs(t, err, ErrRepositoryNotFound)
}

func TestRemoveRepository_LoadStatusFailure(t *testing.T) {
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

	repoURL := "github.com/test/repo"

	// Mock expectations - file doesn't exist, so loadStatus will create it
	mockFS.EXPECT().Exists(cfg.StatusFile).Return(false, nil)
	mockFS.EXPECT().FileLock(cfg.StatusFile).Return(func() {}, nil)
	mockFS.EXPECT().WriteFileAtomic(cfg.StatusFile, gomock.Any(), gomock.Any()).Return(nil)

	// Execute
	err := manager.RemoveRepository(repoURL)

	// Verify
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "repository not found in status")
}

func TestRemoveRepository_SaveStatusFailure(t *testing.T) {
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
	repoURL := "github.com/test/repo"
	initialStatus := &Status{
		Repositories: map[string]Repository{
			repoURL: {
				Path: "/path/to/repo",
				Remotes: map[string]Remote{
					"origin": {
						DefaultBranch: "main",
					},
				},
				Worktrees: make(map[string]WorktreeInfo),
			},
		},
		Workspaces: make(map[string]Workspace),
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
	err := manager.RemoveRepository(repoURL)

	// Verify
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "failed to save status")
}

func TestRemoveRepository_FileReadFailure(t *testing.T) {
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

	repoURL := "github.com/test/repo"

	// Mock expectations
	mockFS.EXPECT().Exists(cfg.StatusFile).Return(true, nil)
	mockFS.EXPECT().ReadFile(cfg.StatusFile).Return(nil, assert.AnError)

	// Execute
	err := manager.RemoveRepository(repoURL)

	// Verify
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "failed to load status")
}

func TestRemoveRepository_MultipleRepositories(t *testing.T) {
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

	// Test data - multiple repositories
	repoURL := "github.com/repo-to-delete"
	initialStatus := &Status{
		Repositories: map[string]Repository{
			repoURL: {
				Path: "/path/to/repo-to-delete",
				Remotes: map[string]Remote{
					"origin": {
						DefaultBranch: "main",
					},
				},
				Worktrees: make(map[string]WorktreeInfo),
			},
			"github.com/other/repo": {
				Path: "/path/to/other-repo",
				Remotes: map[string]Remote{
					"origin": {
						DefaultBranch: "main",
					},
				},
				Worktrees: make(map[string]WorktreeInfo),
			},
		},
		Workspaces: make(map[string]Workspace),
	}

	// Expected status after removal - other repository should remain
	expectedStatus := &Status{
		Repositories: map[string]Repository{
			"github.com/other/repo": {
				Path: "/path/to/other-repo",
				Remotes: map[string]Remote{
					"origin": {
						DefaultBranch: "main",
					},
				},
				Worktrees: make(map[string]WorktreeInfo),
			},
		},
		Workspaces: make(map[string]Workspace),
	}

	initialData, _ := yaml.Marshal(initialStatus)
	expectedData, _ := yaml.Marshal(expectedStatus)

	// Mock expectations
	mockFS.EXPECT().Exists(cfg.StatusFile).Return(true, nil)
	mockFS.EXPECT().ReadFile(cfg.StatusFile).Return(initialData, nil)
	mockFS.EXPECT().FileLock(cfg.StatusFile).Return(func() {}, nil)
	mockFS.EXPECT().WriteFileAtomic(cfg.StatusFile, expectedData, gomock.Any()).Return(nil)

	// Execute
	err := manager.RemoveRepository(repoURL)

	// Verify
	assert.NoError(t, err)
}

func TestRemoveRepository_WithWorktrees(t *testing.T) {
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

	// Test data - repository with worktrees
	repoURL := "github.com/test/repo"
	initialStatus := &Status{
		Repositories: map[string]Repository{
			repoURL: {
				Path: "/path/to/repo",
				Remotes: map[string]Remote{
					"origin": {
						DefaultBranch: "main",
					},
				},
				Worktrees: map[string]WorktreeInfo{
					"main": {
						Remote: "origin",
						Branch: "main",
					},
					"feature": {
						Remote: "origin",
						Branch: "feature",
					},
				},
			},
		},
		Workspaces: make(map[string]Workspace),
	}

	// Expected status after removal - repository and all worktrees should be removed
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
	err := manager.RemoveRepository(repoURL)

	// Verify
	assert.NoError(t, err)
}
