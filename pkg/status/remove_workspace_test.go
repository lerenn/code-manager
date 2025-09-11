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
