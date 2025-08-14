//go:build unit

package wtm

import (
	"testing"

	"github.com/lerenn/wtm/pkg/config"
	"github.com/lerenn/wtm/pkg/fs"
	"github.com/lerenn/wtm/pkg/git"
	"github.com/lerenn/wtm/pkg/status"
	"github.com/stretchr/testify/assert"
	"go.uber.org/mock/gomock"
)

// createTestConfig creates a test configuration for unit tests.
func createTestConfig() *config.Config {
	return &config.Config{
		BasePath:   "/test/base/path",
		StatusFile: "/test/base/path/status.yaml",
	}
}

func TestWTM_Run_SingleRepository(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockFS := fs.NewMockFS(ctrl)
	mockGit := git.NewMockGit(ctrl)
	mockStatus := status.NewMockManager(ctrl)

	wtm := NewWTM(createTestConfig())

	// Override adapters with mocks
	c := wtm.(*realWTM)
	c.fs = mockFS
	c.git = mockGit
	c.statusManager = mockStatus

	// Mock single repo detection - .git found (called multiple times: detectProjectType, validateGitDirectory, and createWorktreeForSingleRepo)
	mockFS.EXPECT().Exists(".git").Return(true, nil).AnyTimes() // detectProjectType, validateGitDirectory, createWorktreeForSingleRepo (validateRepository), prepareWorktreePath, and additional validation
	mockFS.EXPECT().IsDir(".git").Return(true, nil).AnyTimes()  // Called in detectSingleRepoMode, validateGitDirectory, createWorktreeForSingleRepo (validateRepository), prepareWorktreePath, and additional validation

	// Mock Git status for validation (called 2 times: validateGitStatus and validateGitConfiguration)
	mockGit.EXPECT().Status(".").Return("On branch main", nil).Times(2)

	// Mock status manager calls
	mockStatus.EXPECT().GetWorktree("github.com/lerenn/example", "test-branch").Return(nil, status.ErrWorktreeNotFound).AnyTimes()
	mockStatus.EXPECT().AddWorktree("github.com/lerenn/example", "test-branch", gomock.Any(), "").Return(nil)

	// Mock worktree creation calls
	mockGit.EXPECT().GetRepositoryName(gomock.Any()).Return("github.com/lerenn/example", nil)
	mockGit.EXPECT().IsClean(gomock.Any()).Return(true, nil)
	mockGit.EXPECT().BranchExists(gomock.Any(), "test-branch").Return(false, nil)
	mockGit.EXPECT().CreateBranch(gomock.Any(), "test-branch").Return(nil)
	mockFS.EXPECT().Exists(gomock.Any()).Return(false, nil).AnyTimes() // Worktree directory doesn't exist
	mockFS.EXPECT().MkdirAll(gomock.Any(), gomock.Any()).Return(nil)   // Create directory structure
	mockGit.EXPECT().CreateWorktree(gomock.Any(), gomock.Any(), "test-branch").Return(nil)

	err := wtm.CreateWorkTree("test-branch")
	assert.NoError(t, err)
}

func TestWTM_Run_VerboseMode(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockFS := fs.NewMockFS(ctrl)
	mockGit := git.NewMockGit(ctrl)
	mockStatus := status.NewMockManager(ctrl)

	wtm := NewWTM(createTestConfig())
	wtm.SetVerbose(true)

	// Override adapters with mocks
	c := wtm.(*realWTM)
	c.fs = mockFS
	c.git = mockGit
	c.statusManager = mockStatus

	// Mock single repo detection - .git found (called multiple times: detectProjectType, validateGitDirectory, and createWorktreeForSingleRepo)
	mockFS.EXPECT().Exists(".git").Return(true, nil).AnyTimes() // detectProjectType, validateGitDirectory, createWorktreeForSingleRepo (validateRepository), prepareWorktreePath, and additional validation
	mockFS.EXPECT().IsDir(".git").Return(true, nil).AnyTimes()  // Called in detectSingleRepoMode, validateGitDirectory, createWorktreeForSingleRepo (validateRepository), prepareWorktreePath, and additional validation

	// Mock Git status for validation (called 2 times: validateGitStatus and validateGitConfiguration)
	mockGit.EXPECT().Status(".").Return("On branch main", nil).Times(2)

	// Mock status manager calls
	mockStatus.EXPECT().GetWorktree("github.com/lerenn/example", "test-branch").Return(nil, status.ErrWorktreeNotFound).AnyTimes()
	mockStatus.EXPECT().AddWorktree("github.com/lerenn/example", "test-branch", gomock.Any(), "").Return(nil)

	// Mock worktree creation calls
	mockGit.EXPECT().GetRepositoryName(gomock.Any()).Return("github.com/lerenn/example", nil)
	mockGit.EXPECT().IsClean(gomock.Any()).Return(true, nil)
	mockGit.EXPECT().BranchExists(gomock.Any(), "test-branch").Return(false, nil)
	mockGit.EXPECT().CreateBranch(gomock.Any(), "test-branch").Return(nil)
	mockFS.EXPECT().Exists(gomock.Any()).Return(false, nil).AnyTimes() // Worktree directory doesn't exist
	mockFS.EXPECT().MkdirAll(gomock.Any(), gomock.Any()).Return(nil)   // Create directory structure
	mockGit.EXPECT().CreateWorktree(gomock.Any(), gomock.Any(), "test-branch").Return(nil)

	err := wtm.CreateWorkTree("test-branch")
	assert.NoError(t, err)
}
