//go:build unit

package cm

import (
	"os"
	"path/filepath"
	"testing"

	basepkg "github.com/lerenn/code-manager/internal/base"
	"github.com/lerenn/code-manager/pkg/fs"
	"github.com/lerenn/code-manager/pkg/git"
	"github.com/lerenn/code-manager/pkg/hooks/ide_opening"
	"github.com/lerenn/code-manager/pkg/logger"
	"github.com/lerenn/code-manager/pkg/prompt"
	"github.com/lerenn/code-manager/pkg/status"
	"github.com/stretchr/testify/assert"
	"go.uber.org/mock/gomock"
)

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
		ideManager: ide_opening.NewManager(mockFS, mockLogger),
	}

	// Create a temporary directory for testing
	tempDir := t.TempDir()

	// Mock prompt for base path
	mockPrompt.EXPECT().PromptForBasePath("/test/base/path").Return("~/Code", nil)
	mockFS.EXPECT().ExpandPath("~/Code").Return(tempDir, nil)
	mockFS.EXPECT().CreateDirectory(tempDir, os.FileMode(0755)).Return(nil)
	mockFS.EXPECT().GetHomeDir().Return(filepath.Dir(tempDir), nil).AnyTimes()
	mockFS.EXPECT().Exists("/test/status.yaml").Return(false, nil)
	mockStatus.EXPECT().CreateInitialStatus().Return(nil).AnyTimes()

	err := cm.Init(InitOpts{})
	assert.NoError(t, err)
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

	// Mock path expansion failure
	mockFS.EXPECT().ExpandPath("/invalid/path").Return("", assert.AnError)

	opts := InitOpts{
		Force:    false,
		Reset:    false,
		BasePath: "/invalid/path",
	}

	err := cm.Init(opts)
	assert.Error(t, err)
	assert.ErrorIs(t, err, ErrFailedToExpandBasePath)
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
		ideManager: ide_opening.NewManager(mockFS, mockLogger),
	}

	// Create a temporary directory for testing
	tempDir := t.TempDir()

	// Mock reset initialization
	mockPrompt.EXPECT().PromptForConfirmation(
		"This will reset your CM configuration and remove all existing worktrees. Are you sure?", false).Return(true, nil)
	mockPrompt.EXPECT().PromptForBasePath("/test/base/path").Return("~/Code", nil)
	mockFS.EXPECT().ExpandPath("~/Code").Return(tempDir, nil)
	mockFS.EXPECT().CreateDirectory(tempDir, os.FileMode(0755)).Return(nil)
	mockFS.EXPECT().GetHomeDir().Return(filepath.Dir(tempDir), nil).AnyTimes()
	mockStatus.EXPECT().CreateInitialStatus().Return(nil).AnyTimes()
	// Add expectation for status file existence check
	mockFS.EXPECT().Exists("/test/status.yaml").Return(false, nil)

	err := cm.Init(InitOpts{Reset: true})
	assert.NoError(t, err)
}
