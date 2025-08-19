//go:build unit

package cm

import (
	"testing"

	"github.com/lerenn/cm/pkg/config"
	"github.com/lerenn/cm/pkg/fs"
	"github.com/lerenn/cm/pkg/git"
	"github.com/lerenn/cm/pkg/ide"
	"github.com/lerenn/cm/pkg/status"
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

	// Mock single repo detection - .git found (called multiple times: detectProjectType, validateGitDirectory, and createWorktreeForSingleRepo)
	mockFS.EXPECT().Exists(".git").Return(true, nil).AnyTimes() // detectProjectType, validateGitDirectory, createWorktreeForSingleRepo (validateRepository), prepareWorktreePath, and additional validation
	mockFS.EXPECT().IsDir(".git").Return(true, nil).AnyTimes()  // Called in detectSingleRepoMode, validateGitDirectory, createWorktreeForSingleRepo (validateRepository), prepareWorktreePath, and additional validation

	// Mock Git status for validation (called 2 times: validateGitStatus and validateGitConfiguration)
	mockGit.EXPECT().Status(".").Return("On branch main", nil).Times(2)

	// Mock status manager calls
	mockStatus.EXPECT().GetWorktree("github.com/lerenn/example", "test-branch").Return(nil, status.ErrWorktreeNotFound).AnyTimes()
	mockStatus.EXPECT().AddWorktree(gomock.Any()).Return(nil)

	// Mock worktree creation calls
	mockGit.EXPECT().GetRepositoryName(gomock.Any()).Return("github.com/lerenn/example", nil)
	mockGit.EXPECT().IsClean(gomock.Any()).Return(true, nil)
	mockGit.EXPECT().BranchExists(gomock.Any(), "test-branch").Return(false, nil)
	mockGit.EXPECT().CreateBranch(gomock.Any(), "test-branch").Return(nil)
	mockFS.EXPECT().Exists(gomock.Any()).Return(false, nil).AnyTimes() // Worktree directory doesn't exist
	mockFS.EXPECT().MkdirAll(gomock.Any(), gomock.Any()).Return(nil)   // Create directory structure
	mockGit.EXPECT().CreateWorktree(gomock.Any(), gomock.Any(), "test-branch").Return(nil)

	err := cm.CreateWorkTree("test-branch")
	assert.NoError(t, err)
}

func TestCM_CreateWorkTreeWithIDE(t *testing.T) {
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
	mockFS.EXPECT().Exists(".git").Return(true, nil).AnyTimes()
	mockFS.EXPECT().IsDir(".git").Return(true, nil).AnyTimes()

	// Mock workspace detection - no workspace files found (called in detectProjectMode)
	mockFS.EXPECT().Glob("*.code-workspace").Return([]string{}, nil).AnyTimes()

	// Mock Git status for validation
	mockGit.EXPECT().Status(".").Return("On branch main", nil).Times(2)

	// Mock status manager calls
	mockStatus.EXPECT().GetWorktree("github.com/lerenn/example", "test-branch").Return(nil, status.ErrWorktreeNotFound)
	mockStatus.EXPECT().AddWorktree(gomock.Any()).Return(nil)

	// Mock worktree creation calls
	mockGit.EXPECT().GetRepositoryName(gomock.Any()).Return("github.com/lerenn/example", nil).AnyTimes()
	mockGit.EXPECT().IsClean(gomock.Any()).Return(true, nil)
	mockGit.EXPECT().BranchExists(gomock.Any(), "test-branch").Return(false, nil)
	mockGit.EXPECT().CreateBranch(gomock.Any(), "test-branch").Return(nil)
	mockFS.EXPECT().Exists(gomock.Any()).Return(false, nil).AnyTimes()
	mockFS.EXPECT().MkdirAll(gomock.Any(), gomock.Any()).Return(nil)
	mockGit.EXPECT().CreateWorktree(gomock.Any(), gomock.Any(), "test-branch").Return(nil)

	// Mock IDE opening
	ideName := "cursor"
	mockIDE.EXPECT().OpenIDE("cursor", gomock.Any(), false).Return(nil)

	err := cm.CreateWorkTree("test-branch", CreateWorkTreeOpts{IDEName: ideName})
	assert.NoError(t, err)
}

func TestCM_Run_VerboseMode(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockFS := fs.NewMockFS(ctrl)
	mockGit := git.NewMockGit(ctrl)
	mockStatus := status.NewMockManager(ctrl)
	mockIDE := ide.NewMockManagerInterface(ctrl)

	cm := NewCM(createTestConfig())
	cm.SetVerbose(true)

	// Override adapters with mocks
	c := cm.(*realCM)
	c.FS = mockFS
	c.Git = mockGit
	c.StatusManager = mockStatus
	c.ideManager = mockIDE

	// Mock single repo detection - .git found (called multiple times: detectProjectType, validateGitDirectory, and createWorktreeForSingleRepo)
	mockFS.EXPECT().Exists(".git").Return(true, nil).AnyTimes() // detectProjectType, validateGitDirectory, createWorktreeForSingleRepo (validateRepository), prepareWorktreePath, and additional validation
	mockFS.EXPECT().IsDir(".git").Return(true, nil).AnyTimes()  // Called in detectSingleRepoMode, validateGitDirectory, createWorktreeForSingleRepo (validateRepository), prepareWorktreePath, and additional validation

	// Mock Git status for validation (called 2 times: validateGitStatus and validateGitConfiguration)
	mockGit.EXPECT().Status(".").Return("On branch main", nil).Times(2)

	// Mock status manager calls
	mockStatus.EXPECT().GetWorktree("github.com/lerenn/example", "test-branch").Return(nil, status.ErrWorktreeNotFound).AnyTimes()
	mockStatus.EXPECT().AddWorktree(gomock.Any()).Return(nil)

	// Mock worktree creation calls
	mockGit.EXPECT().GetRepositoryName(gomock.Any()).Return("github.com/lerenn/example", nil)
	mockGit.EXPECT().IsClean(gomock.Any()).Return(true, nil)
	mockGit.EXPECT().BranchExists(gomock.Any(), "test-branch").Return(false, nil)
	mockGit.EXPECT().CreateBranch(gomock.Any(), "test-branch").Return(nil)
	mockFS.EXPECT().Exists(gomock.Any()).Return(false, nil).AnyTimes() // Worktree directory doesn't exist
	mockFS.EXPECT().MkdirAll(gomock.Any(), gomock.Any()).Return(nil)   // Create directory structure
	mockGit.EXPECT().CreateWorktree(gomock.Any(), gomock.Any(), "test-branch").Return(nil)

	err := cm.CreateWorkTree("test-branch")
	assert.NoError(t, err)
}
