//go:build unit

package cm

import (
	"testing"

	"github.com/lerenn/code-manager/pkg/config"
	"github.com/lerenn/code-manager/pkg/fs"
	"github.com/lerenn/code-manager/pkg/git"
	"github.com/lerenn/code-manager/pkg/ide"
	"github.com/lerenn/code-manager/pkg/repository"
	"github.com/lerenn/code-manager/pkg/workspace"
	"github.com/stretchr/testify/assert"
	"go.uber.org/mock/gomock"
)

// createTestConfig creates a test configuration for use in tests.
func createTestConfig() *config.Config {
	return &config.Config{
		BasePath:   "/test/base/path",
		StatusFile: "/test/status.yaml",
	}
}

func TestCM_Run_SingleRepository(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockRepository := repository.NewMockRepository(ctrl)
	mockWorkspace := workspace.NewMockWorkspace(ctrl)
	mockIDE := ide.NewMockManagerInterface(ctrl)

	// Create CM with mocked dependencies
	cm := NewCMWithDependencies(NewCMParams{
		Repository: mockRepository,
		Workspace:  mockWorkspace,
		Config:     createTestConfig(),
	})

	// Override IDE manager with mock
	c := cm.(*realCM)
	c.ideManager = mockIDE

	// Mock repository detection and worktree creation
	mockRepository.EXPECT().IsGitRepository().Return(true, nil).AnyTimes()
	mockRepository.EXPECT().Validate().Return(nil)
	mockRepository.EXPECT().CreateWorktree("test-branch").Return(nil)

	err := cm.CreateWorkTree("test-branch")
	assert.NoError(t, err)
}

func TestCM_CreateWorkTreeWithIDE(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockRepository := repository.NewMockRepository(ctrl)
	mockWorkspace := workspace.NewMockWorkspace(ctrl)
	mockIDE := ide.NewMockManagerInterface(ctrl)
	mockFS := fs.NewMockFS(ctrl)
	mockGit := git.NewMockGit(ctrl)

	// Create CM with mocked dependencies
	cm := NewCMWithDependencies(NewCMParams{
		Repository: mockRepository,
		Workspace:  mockWorkspace,
		Config:     createTestConfig(),
	})

	// Override dependencies with mocks
	c := cm.(*realCM)
	c.ideManager = mockIDE
	c.FS = mockFS
	c.Git = mockGit

	// Mock repository detection and worktree creation
	mockRepository.EXPECT().IsGitRepository().Return(true, nil).AnyTimes()
	mockRepository.EXPECT().Validate().Return(nil)
	mockRepository.EXPECT().CreateWorktree("test-branch").Return(nil)

	// Mock Git and FS operations for OpenWorktree
	mockGit.EXPECT().GetRepositoryName(".").Return("github.com/octocat/Hello-World", nil)
	mockFS.EXPECT().Exists(gomock.Any()).Return(true, nil)

	// Mock IDE opening
	mockIDE.EXPECT().OpenIDE("cursor", gomock.Any(), false).Return(nil)

	err := cm.CreateWorkTree("test-branch", CreateWorkTreeOpts{IDEName: "cursor"})
	assert.NoError(t, err)
}

func TestCM_Run_VerboseMode(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockRepository := repository.NewMockRepository(ctrl)
	mockWorkspace := workspace.NewMockWorkspace(ctrl)

	// Create CM with mocked dependencies
	cm := NewCMWithDependencies(NewCMParams{
		Repository: mockRepository,
		Workspace:  mockWorkspace,
		Config:     createTestConfig(),
	})

	// Enable verbose mode
	cm.SetVerbose(true)

	// Mock repository detection and worktree creation
	mockRepository.EXPECT().IsGitRepository().Return(true, nil).AnyTimes()
	mockRepository.EXPECT().Validate().Return(nil)
	mockRepository.EXPECT().CreateWorktree("test-branch").Return(nil)

	err := cm.CreateWorkTree("test-branch")
	assert.NoError(t, err)
}
