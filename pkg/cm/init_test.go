//go:build unit

package cm

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/lerenn/code-manager/pkg/config"
	fsmocks "github.com/lerenn/code-manager/pkg/fs/mocks"
	gitmocks "github.com/lerenn/code-manager/pkg/git/mocks"
	"github.com/lerenn/code-manager/pkg/mode/repository"
	repositorymocks "github.com/lerenn/code-manager/pkg/mode/repository/mocks"
	"github.com/lerenn/code-manager/pkg/mode/workspace"
	workspacemocks "github.com/lerenn/code-manager/pkg/mode/workspace/mocks"
	promptmocks "github.com/lerenn/code-manager/pkg/prompt/mocks"
	statusmocks "github.com/lerenn/code-manager/pkg/status/mocks"
	"github.com/stretchr/testify/assert"
	"go.uber.org/mock/gomock"
)

// createInitTestConfig creates a test configuration for use in tests.
func createInitTestConfig() config.Config {
	return config.Config{
		BasePath:   "/test/base/path",
		StatusFile: "/test/status.yaml",
	}
}

func TestRealCM_Init_Success(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockFS := fsmocks.NewMockFS(ctrl)

	mockGit := gitmocks.NewMockGit(ctrl)
	mockStatus := statusmocks.NewMockManager(ctrl)
	mockPrompt := promptmocks.NewMockPrompter(ctrl)
	mockRepository := repositorymocks.NewMockRepository(ctrl)
	mockWorkspace := workspacemocks.NewMockWorkspace(ctrl)

	var cm CM
	var err error

	cm, err = NewCM(NewCMParams{
		RepositoryProvider: func(params repository.NewRepositoryParams) repository.Repository { return mockRepository },
		WorkspaceProvider:  func(params workspace.NewWorkspaceParams) workspace.Workspace { return mockWorkspace },
		Config:             createInitTestConfig(),
		FS:                 mockFS,
		Git:                mockGit,
		Status:             mockStatus,

		Prompt: mockPrompt,
	})
	assert.NoError(t, err)

	// Create a temporary directory for testing
	tempDir := t.TempDir()

	// Mock prompt for base path
	mockPrompt.EXPECT().PromptForBasePath("/test/base/path").Return("~/Code", nil)
	mockFS.EXPECT().ExpandPath("~/Code").Return(tempDir, nil)
	mockFS.EXPECT().CreateDirectory(tempDir, os.FileMode(0755)).Return(nil)
	mockFS.EXPECT().GetHomeDir().Return(filepath.Dir(tempDir), nil).AnyTimes()
	mockFS.EXPECT().Exists("/test/status.yaml").Return(false, nil)
	mockStatus.EXPECT().CreateInitialStatus().Return(nil).AnyTimes()

	err = cm.Init(InitOpts{})
	assert.NoError(t, err)
}

func TestRealCM_Init_InvalidBasePath(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockFS := fsmocks.NewMockFS(ctrl)

	mockGit := gitmocks.NewMockGit(ctrl)
	mockStatus := statusmocks.NewMockManager(ctrl)
	mockPrompt := promptmocks.NewMockPrompter(ctrl)
	mockRepository := repositorymocks.NewMockRepository(ctrl)
	mockWorkspace := workspacemocks.NewMockWorkspace(ctrl)

	var cm CM
	var err error

	cm, err = NewCM(NewCMParams{
		RepositoryProvider: func(params repository.NewRepositoryParams) repository.Repository { return mockRepository },
		WorkspaceProvider:  func(params workspace.NewWorkspaceParams) workspace.Workspace { return mockWorkspace },
		Config:             createInitTestConfig(),
		FS:                 mockFS,
		Git:                mockGit,
		Status:             mockStatus,

		Prompt: mockPrompt,
	})
	assert.NoError(t, err)

	// Mock path expansion failure
	mockFS.EXPECT().ExpandPath("/invalid/path").Return("", assert.AnError)

	opts := InitOpts{
		Force:    false,
		Reset:    false,
		BasePath: "/invalid/path",
	}

	err = cm.Init(opts)
	assert.Error(t, err)
	assert.ErrorIs(t, err, ErrFailedToExpandBasePath)
}

func TestRealCM_Init_ResetSuccess(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockFS := fsmocks.NewMockFS(ctrl)

	mockGit := gitmocks.NewMockGit(ctrl)
	mockStatus := statusmocks.NewMockManager(ctrl)
	mockPrompt := promptmocks.NewMockPrompter(ctrl)
	mockRepository := repositorymocks.NewMockRepository(ctrl)
	mockWorkspace := workspacemocks.NewMockWorkspace(ctrl)

	var cm CM
	var err error

	cm, err = NewCM(NewCMParams{
		RepositoryProvider: func(params repository.NewRepositoryParams) repository.Repository { return mockRepository },
		WorkspaceProvider:  func(params workspace.NewWorkspaceParams) workspace.Workspace { return mockWorkspace },
		Config:             createInitTestConfig(),
		FS:                 mockFS,
		Git:                mockGit,
		Status:             mockStatus,

		Prompt: mockPrompt,
	})
	assert.NoError(t, err)

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

	err = cm.Init(InitOpts{Reset: true})
	assert.NoError(t, err)
}
