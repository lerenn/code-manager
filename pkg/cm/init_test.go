//go:build unit

package cm

import (
	"os"
	"testing"

	basepkg "github.com/lerenn/cm/internal/base"
	"github.com/lerenn/cm/pkg/fs"
	"github.com/lerenn/cm/pkg/git"
	"github.com/lerenn/cm/pkg/ide"
	"github.com/lerenn/cm/pkg/logger"
	"github.com/lerenn/cm/pkg/prompt"
	"github.com/lerenn/cm/pkg/status"
	"github.com/stretchr/testify/assert"
	"go.uber.org/mock/gomock"
)

func TestRealCM_IsInitialized(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockFS := fs.NewMockFS(ctrl)
	mockGit := git.NewMockGit(ctrl)
	mockStatus := status.NewMockManager(ctrl)
	mockLogger := logger.NewNoopLogger()
	mockPrompt := prompt.NewMockPrompt(ctrl)
	cm := &realCM{
		Base: basepkg.NewBase(basepkg.NewBaseParams{
			FS:            mockFS,
			Git:           mockGit,
			Config:        createTestConfig(),
			StatusManager: mockStatus,
			Logger:        mockLogger,
			Prompt:        mockPrompt,
			Verbose:       false,
		}),
		ideManager: nil,
	}

	// Test when initialized
	mockStatus.EXPECT().IsInitialized().Return(true, nil)
	initialized, err := cm.IsInitialized()
	assert.NoError(t, err)
	assert.True(t, initialized)

	// Test when not initialized
	mockStatus.EXPECT().IsInitialized().Return(false, nil)
	initialized, err = cm.IsInitialized()
	assert.NoError(t, err)
	assert.False(t, initialized)
}

func TestRealCM_Init_Success(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockFS := fs.NewMockFS(ctrl)
	mockGit := git.NewMockGit(ctrl)
	mockStatus := status.NewMockManager(ctrl)
	mockLogger := logger.NewNoopLogger()
	mockPrompt := prompt.NewMockPrompt(ctrl)

	cm := &realCM{
		Base: basepkg.NewBase(basepkg.NewBaseParams{
			FS:            mockFS,
			Git:           mockGit,
			Config:        createTestConfig(),
			StatusManager: mockStatus,
			Logger:        mockLogger,
			Prompt:        mockPrompt,
			Verbose:       false,
		}),
		ideManager: ide.NewManager(mockFS, mockLogger),
	}

	// Mock initialization checks
	mockStatus.EXPECT().IsInitialized().Return(false, nil)
	mockPrompt.EXPECT().PromptForBasePath().Return("~/Code", nil)
	mockFS.EXPECT().ExpandPath("~/Code").Return("/home/user/Code", nil)
	mockFS.EXPECT().CreateDirectory("/home/user/Code", os.FileMode(0755)).Return(nil)
	mockFS.EXPECT().GetHomeDir().Return("/home/user", nil).AnyTimes()
	mockStatus.EXPECT().CreateInitialStatus().Return(nil).AnyTimes()
	mockStatus.EXPECT().SetInitialized(true).Return(nil)

	err := cm.Init(InitOpts{})
	assert.NoError(t, err)
}

func TestRealCM_Init_AlreadyInitialized(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockFS := fs.NewMockFS(ctrl)
	mockGit := git.NewMockGit(ctrl)
	mockStatus := status.NewMockManager(ctrl)
	mockLogger := logger.NewNoopLogger()
	mockPrompt := prompt.NewMockPrompt(ctrl)

	cm := &realCM{
		Base: basepkg.NewBase(basepkg.NewBaseParams{
			FS:            mockFS,
			Git:           mockGit,
			Config:        createTestConfig(),
			StatusManager: mockStatus,
			Logger:        mockLogger,
			Prompt:        mockPrompt,
			Verbose:       false,
		}),
		ideManager: ide.NewManager(mockFS, mockLogger),
	}

	// Mock already initialized
	mockStatus.EXPECT().IsInitialized().Return(true, nil)

	err := cm.Init(InitOpts{})
	assert.ErrorIs(t, err, ErrAlreadyInitialized)
}

func TestRealCM_Init_InvalidBasePath(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockFS := fs.NewMockFS(ctrl)
	mockGit := git.NewMockGit(ctrl)
	mockStatus := status.NewMockManager(ctrl)
	mockLogger := logger.NewNoopLogger()
	mockPrompt := prompt.NewMockPrompt(ctrl)
	cm := &realCM{
		Base: basepkg.NewBase(basepkg.NewBaseParams{
			FS:            mockFS,
			Git:           mockGit,
			Config:        createTestConfig(),
			StatusManager: mockStatus,
			Logger:        mockLogger,
			Prompt:        mockPrompt,
			Verbose:       false,
		}),
		ideManager: nil,
	}

	// Mock that CM is not initialized
	mockStatus.EXPECT().IsInitialized().Return(false, nil)

	// Mock path expansion failure
	mockFS.EXPECT().ExpandPath("/invalid/path").Return("", assert.AnError)

	opts := InitOpts{
		Force:    false,
		Reset:    false,
		BasePath: "/invalid/path",
	}

	err := cm.Init(opts)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "failed to expand base path")
}

func TestRealCM_Init_ResetSuccess(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockFS := fs.NewMockFS(ctrl)
	mockGit := git.NewMockGit(ctrl)
	mockStatus := status.NewMockManager(ctrl)
	mockLogger := logger.NewNoopLogger()
	mockPrompt := prompt.NewMockPrompt(ctrl)

	cm := &realCM{
		Base: basepkg.NewBase(basepkg.NewBaseParams{
			FS:            mockFS,
			Git:           mockGit,
			Config:        createTestConfig(),
			StatusManager: mockStatus,
			Logger:        mockLogger,
			Prompt:        mockPrompt,
			Verbose:       false,
		}),
		ideManager: ide.NewManager(mockFS, mockLogger),
	}

	// Mock reset initialization
	mockPrompt.EXPECT().PromptForConfirmation(
		"This will reset your CM configuration and remove all existing worktrees. Are you sure?", false).Return(true, nil)
	mockPrompt.EXPECT().PromptForBasePath().Return("~/Code", nil)
	mockFS.EXPECT().ExpandPath("~/Code").Return("/home/user/Code", nil)
	mockFS.EXPECT().CreateDirectory("/home/user/Code", os.FileMode(0755)).Return(nil)
	mockFS.EXPECT().GetHomeDir().Return("/home/user", nil).AnyTimes()
	mockStatus.EXPECT().CreateInitialStatus().Return(nil).AnyTimes()
	mockStatus.EXPECT().SetInitialized(true).Return(nil)

	err := cm.Init(InitOpts{Reset: true})
	assert.NoError(t, err)
}
