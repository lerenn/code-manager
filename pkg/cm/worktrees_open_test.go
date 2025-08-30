//go:build unit

package cm

import (
	"testing"

	"github.com/lerenn/code-manager/pkg/fs"
	"github.com/lerenn/code-manager/pkg/git"
	"github.com/lerenn/code-manager/pkg/hooks/ide_opening"
	"github.com/lerenn/code-manager/pkg/repository"
	"github.com/lerenn/code-manager/pkg/workspace"
	"github.com/stretchr/testify/assert"
	"go.uber.org/mock/gomock"
)

func TestCM_OpenWorktree(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockRepository := repository.NewMockRepository(ctrl)
	mockWorkspace := workspace.NewMockWorkspace(ctrl)
	mockIDE := ide_opening.NewMockManagerInterface(ctrl)
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

	// Mock repository detection
	mockRepository.EXPECT().IsGitRepository().Return(true, nil).AnyTimes()

	// Mock Git to return repository URL
	mockGit.EXPECT().GetRepositoryName(".").Return("github.com/octocat/Hello-World", nil)

	// Mock worktree path existence check
	mockFS.EXPECT().Exists("/test/base/path/github.com/octocat/Hello-World/origin/test-branch").Return(true, nil)

	// Mock IDE opening
	mockIDE.EXPECT().OpenIDE("vscode", "/test/base/path/github.com/octocat/Hello-World/origin/test-branch", false).Return(nil)

	err := cm.OpenWorktree("test-branch", "vscode")
	assert.NoError(t, err)
}

func TestCM_OpenWorktree_NotFound(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockRepository := repository.NewMockRepository(ctrl)
	mockWorkspace := workspace.NewMockWorkspace(ctrl)
	mockIDE := ide_opening.NewMockManagerInterface(ctrl)
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

	// Mock repository detection
	mockRepository.EXPECT().IsGitRepository().Return(true, nil).AnyTimes()

	// Mock Git to return repository URL
	mockGit.EXPECT().GetRepositoryName(".").Return("github.com/octocat/Hello-World", nil)

	// Mock worktree path existence check - worktree not found
	mockFS.EXPECT().Exists("/test/base/path/github.com/octocat/Hello-World/origin/test-branch").Return(false, nil)

	err := cm.OpenWorktree("test-branch", "vscode")
	assert.Error(t, err)
	assert.ErrorIs(t, err, ErrWorktreeNotInStatus)
}
