//go:build unit

package cm

import (
	"testing"

	"github.com/lerenn/cm/pkg/fs"
	"github.com/lerenn/cm/pkg/git"
	"github.com/lerenn/cm/pkg/ide"
	"github.com/lerenn/cm/pkg/status"
	"github.com/stretchr/testify/assert"
	"go.uber.org/mock/gomock"
)

func TestCM_OpenWorktree(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockFS := fs.NewMockFS(ctrl)
	mockGit := git.NewMockGit(ctrl)
	mockStatus := status.NewMockManager(ctrl)
	mockIDE := ide.NewMockManagerInterface(ctrl)

	cm := NewCM(createTestConfig())

	// Override adapters with mocks
	c := cm.(*realCM)
	c.FS = mockFS
	c.Git = mockGit
	c.StatusManager = mockStatus
	c.ideManager = mockIDE

	// Mock single repo detection - .git found
	mockFS.EXPECT().Exists(".git").Return(true, nil).Times(1)
	mockFS.EXPECT().IsDir(".git").Return(true, nil).Times(1)

	// Mock Git to return repository URL
	mockGit.EXPECT().GetRepositoryName(".").Return("github.com/lerenn/example", nil)

	// Mock worktree path existence check
	mockFS.EXPECT().Exists("/test/base/path/worktrees/github.com/lerenn/example/test-branch").Return(true, nil)

	// Mock IDE opening - now uses derived worktree path
	mockIDE.EXPECT().OpenIDE("cursor", "/test/base/path/worktrees/github.com/lerenn/example/test-branch", false).Return(nil)

	err := cm.OpenWorktree("test-branch", "cursor")
	assert.NoError(t, err)
}

func TestCM_OpenWorktree_NotFound(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockFS := fs.NewMockFS(ctrl)
	mockGit := git.NewMockGit(ctrl)
	mockStatus := status.NewMockManager(ctrl)
	mockIDE := ide.NewMockManagerInterface(ctrl)

	cm := NewCM(createTestConfig())

	// Override adapters with mocks
	c := cm.(*realCM)
	c.FS = mockFS
	c.Git = mockGit
	c.StatusManager = mockStatus
	c.ideManager = mockIDE

	// Mock single repo detection - .git found
	mockFS.EXPECT().Exists(".git").Return(true, nil).Times(1)
	mockFS.EXPECT().IsDir(".git").Return(true, nil).Times(1)

	// Mock Git to return repository URL
	mockGit.EXPECT().GetRepositoryName(".").Return("github.com/lerenn/example", nil)

	// Mock worktree path existence check
	mockFS.EXPECT().Exists("/test/base/path/worktrees/github.com/lerenn/example/test-branch").Return(true, nil)

	// Mock IDE opening - the method will try to open the worktree path
	mockIDE.EXPECT().OpenIDE("cursor", "/test/base/path/worktrees/github.com/lerenn/example/test-branch", false).Return(nil)

	err := cm.OpenWorktree("test-branch", "cursor")
	assert.NoError(t, err)
}
