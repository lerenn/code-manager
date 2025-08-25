//go:build unit

package cm

import (
	"testing"

	"github.com/lerenn/code-manager/pkg/fs"
	"github.com/lerenn/code-manager/pkg/git"
	"github.com/lerenn/code-manager/pkg/ide"
	"github.com/lerenn/code-manager/pkg/repository"
	"github.com/lerenn/code-manager/pkg/workspace"
	"github.com/stretchr/testify/assert"
	"go.uber.org/mock/gomock"
)

func TestCM_LoadWorktree_Success(t *testing.T) {
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

	// Mock repository detection and worktree loading
	mockRepository.EXPECT().IsGitRepository().Return(true, nil).AnyTimes()
	mockRepository.EXPECT().LoadWorktree("origin", "feature-branch").Return(nil)

	err := cm.LoadWorktree("origin:feature-branch")
	assert.NoError(t, err)
}

func TestCM_LoadWorktree_WithIDE(t *testing.T) {
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

	// Mock repository detection and worktree loading
	mockRepository.EXPECT().IsGitRepository().Return(true, nil).AnyTimes()
	mockRepository.EXPECT().LoadWorktree("origin", "feature-branch").Return(nil)

	// Mock Git and FS operations for OpenWorktree
	mockGit.EXPECT().GetRepositoryName(".").Return("github.com/octocat/Hello-World", nil)
	mockFS.EXPECT().Exists(gomock.Any()).Return(true, nil)

	// Mock IDE opening
	mockIDE.EXPECT().OpenIDE("cursor", gomock.Any(), false).Return(nil)

	err := cm.LoadWorktree("origin:feature-branch", LoadWorktreeOpts{IDEName: "cursor"})
	assert.NoError(t, err)
}

func TestCM_LoadWorktree_NewRemote(t *testing.T) {
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

	// Mock repository detection and worktree loading with new remote
	mockRepository.EXPECT().IsGitRepository().Return(true, nil).AnyTimes()
	mockRepository.EXPECT().LoadWorktree("otheruser", "feature-branch").Return(nil)

	err := cm.LoadWorktree("otheruser:feature-branch")
	assert.NoError(t, err)
}

func TestCM_LoadWorktree_SSHProtocol(t *testing.T) {
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

	// Mock repository detection and worktree loading with SSH protocol
	mockRepository.EXPECT().IsGitRepository().Return(true, nil).AnyTimes()
	mockRepository.EXPECT().LoadWorktree("otheruser", "feature-branch").Return(nil)

	err := cm.LoadWorktree("otheruser:feature-branch")
	assert.NoError(t, err)
}

func TestCM_LoadWorktree_OriginRemoteNotFound(t *testing.T) {
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

	// Mock repository detection and worktree loading to return an error
	mockRepository.EXPECT().IsGitRepository().Return(true, nil).AnyTimes()
	mockRepository.EXPECT().LoadWorktree("origin", "feature-branch").Return(ErrOriginRemoteNotFound)

	err := cm.LoadWorktree("origin:feature-branch")
	assert.Error(t, err)
	assert.ErrorIs(t, err, ErrOriginRemoteNotFound)
}

func TestCM_LoadWorktree_OriginRemoteInvalidURL(t *testing.T) {
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

	// Mock repository detection and worktree loading to return an error
	mockRepository.EXPECT().IsGitRepository().Return(true, nil).AnyTimes()
	mockRepository.EXPECT().LoadWorktree("origin", "feature-branch").Return(ErrOriginRemoteInvalidURL)

	err := cm.LoadWorktree("origin:feature-branch")
	assert.Error(t, err)
	assert.ErrorIs(t, err, ErrOriginRemoteInvalidURL)
}

func TestCM_LoadWorktree_FetchFailed(t *testing.T) {
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

	// Mock repository detection and worktree loading to return an error
	mockRepository.EXPECT().IsGitRepository().Return(true, nil).AnyTimes()
	mockRepository.EXPECT().LoadWorktree("origin", "feature-branch").Return(git.ErrFetchFailed)

	err := cm.LoadWorktree("origin:feature-branch")
	assert.Error(t, err)
	assert.ErrorIs(t, err, git.ErrFetchFailed)
}

func TestCM_LoadWorktree_BranchNotFound(t *testing.T) {
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

	// Mock repository detection and worktree loading to return an error
	mockRepository.EXPECT().IsGitRepository().Return(true, nil).AnyTimes()
	mockRepository.EXPECT().LoadWorktree("origin", "feature-branch").Return(git.ErrBranchNotFoundOnRemote)

	err := cm.LoadWorktree("origin:feature-branch")
	assert.Error(t, err)
	assert.ErrorIs(t, err, git.ErrBranchNotFoundOnRemote)
}

func TestCM_LoadWorktree_DefaultRemote(t *testing.T) {
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

	// Mock repository detection and worktree loading with default remote (origin)
	mockRepository.EXPECT().IsGitRepository().Return(true, nil).AnyTimes()
	mockRepository.EXPECT().LoadWorktree("", "feature-branch").Return(nil)

	err := cm.LoadWorktree("feature-branch")
	assert.NoError(t, err)
}
