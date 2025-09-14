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

func TestCreateInitialStatus_Success(t *testing.T) {
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

	// Expected initial status
	expectedStatus := &Status{
		Repositories: make(map[string]Repository),
		Workspaces:   make(map[string]Workspace),
	}

	expectedData, _ := yaml.Marshal(expectedStatus)

	// Mock expectations
	mockFS.EXPECT().FileLock(cfg.StatusFile).Return(func() {}, nil)
	mockFS.EXPECT().WriteFileAtomic(cfg.StatusFile, expectedData, gomock.Any()).Return(nil)

	// Execute
	err := manager.CreateInitialStatus()

	// Verify
	assert.NoError(t, err)
}

func TestCreateInitialStatus_SaveStatusFailure(t *testing.T) {
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

	// Expected initial status
	expectedStatus := &Status{
		Repositories: make(map[string]Repository),
		Workspaces:   make(map[string]Workspace),
	}

	expectedData, _ := yaml.Marshal(expectedStatus)

	// Mock expectations - save status fails
	mockFS.EXPECT().FileLock(cfg.StatusFile).Return(func() {}, nil)
	mockFS.EXPECT().WriteFileAtomic(cfg.StatusFile, expectedData, gomock.Any()).Return(assert.AnError)

	// Execute
	err := manager.CreateInitialStatus()

	// Verify
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "failed to write status file")
}

func TestCreateInitialStatus_FileLockFailure(t *testing.T) {
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

	// Mock expectations - file lock fails
	mockFS.EXPECT().FileLock(cfg.StatusFile).Return(nil, assert.AnError)

	// Execute
	err := manager.CreateInitialStatus()

	// Verify
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "failed to acquire file lock")
}
