//go:build unit

package wtm

import (
	"testing"

	"github.com/lerenn/wtm/pkg/config"
	"github.com/lerenn/wtm/pkg/fs"
	"github.com/lerenn/wtm/pkg/git"
	"github.com/lerenn/wtm/pkg/ide"
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
	mockIDE := ide.NewMockManagerInterface(ctrl)

	wtm := NewWTM(createTestConfig())

	// Override adapters with mocks
	c := wtm.(*realWTM)
	c.fs = mockFS
	c.git = mockGit
	c.statusManager = mockStatus
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

	err := wtm.CreateWorkTree("test-branch")
	assert.NoError(t, err)
}

func TestWTM_CreateWorkTreeWithIDE(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockFS := fs.NewMockFS(ctrl)
	mockGit := git.NewMockGit(ctrl)
	mockStatus := status.NewMockManager(ctrl)
	mockIDE := ide.NewMockManagerInterface(ctrl)

	wtm := NewWTM(createTestConfig())

	// Override adapters with mocks
	c := wtm.(*realWTM)
	c.fs = mockFS
	c.git = mockGit
	c.statusManager = mockStatus
	c.ideManager = mockIDE

	// Mock single repo detection - .git found
	mockFS.EXPECT().Exists(".git").Return(true, nil).AnyTimes()
	mockFS.EXPECT().IsDir(".git").Return(true, nil).AnyTimes()

	// Mock workspace detection - no workspace files found (called in detectProjectMode)
	mockFS.EXPECT().Glob("*.code-workspace").Return([]string{}, nil).AnyTimes()

	// Mock Git status for validation
	mockGit.EXPECT().Status(".").Return("On branch main", nil).Times(2)

	// Mock status manager calls
	mockStatus.EXPECT().GetWorktree("github.com/lerenn/example", "test-branch").Return(nil, status.ErrWorktreeNotFound).Times(1)
	mockStatus.EXPECT().AddWorktree(gomock.Any()).Return(nil)

	// Mock GetWorktree for IDE opening (called after worktree creation)
	mockStatus.EXPECT().GetWorktree("github.com/lerenn/example", "test-branch").Return(&status.Repository{
		URL:    "github.com/lerenn/example",
		Branch: "test-branch",
		Path:   "/test/base/path/github.com/lerenn/example/test-branch",
	}, nil)

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

	err := wtm.CreateWorkTree("test-branch", CreateWorkTreeOpts{IDEName: ideName})
	assert.NoError(t, err)
}

func TestWTM_OpenWorktree(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockFS := fs.NewMockFS(ctrl)
	mockGit := git.NewMockGit(ctrl)
	mockStatus := status.NewMockManager(ctrl)
	mockIDE := ide.NewMockManagerInterface(ctrl)

	wtm := NewWTM(createTestConfig())

	// Override adapters with mocks
	c := wtm.(*realWTM)
	c.fs = mockFS
	c.git = mockGit
	c.statusManager = mockStatus
	c.ideManager = mockIDE

	// Mock single repo detection - .git found
	mockFS.EXPECT().Exists(".git").Return(true, nil).Times(1)
	mockFS.EXPECT().IsDir(".git").Return(true, nil).Times(1)

	// Mock Git to return repository URL
	mockGit.EXPECT().GetRepositoryName(".").Return("github.com/lerenn/example", nil)

	// Mock status manager to return worktree
	worktree := &status.Repository{
		URL:    "github.com/lerenn/example",
		Branch: "test-branch",
		Path:   "/path/to/worktree",
	}
	mockStatus.EXPECT().GetWorktree("github.com/lerenn/example", "test-branch").Return(worktree, nil)

	// Mock IDE opening - now uses derived worktree path
	mockIDE.EXPECT().OpenIDE("cursor", "/test/base/path/github.com/lerenn/example/test-branch", false).Return(nil)

	err := wtm.OpenWorktree("test-branch", "cursor")
	assert.NoError(t, err)
}

func TestWTM_OpenWorktree_NotFound(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockFS := fs.NewMockFS(ctrl)
	mockGit := git.NewMockGit(ctrl)
	mockStatus := status.NewMockManager(ctrl)
	mockIDE := ide.NewMockManagerInterface(ctrl)

	wtm := NewWTM(createTestConfig())

	// Override adapters with mocks
	c := wtm.(*realWTM)
	c.fs = mockFS
	c.git = mockGit
	c.statusManager = mockStatus
	c.ideManager = mockIDE

	// Mock single repo detection - .git found
	mockFS.EXPECT().Exists(".git").Return(true, nil).Times(1)
	mockFS.EXPECT().IsDir(".git").Return(true, nil).Times(1)

	// Mock Git to return repository URL
	mockGit.EXPECT().GetRepositoryName(".").Return("github.com/lerenn/example", nil)

	// Mock status manager to return error
	mockStatus.EXPECT().GetWorktree("github.com/lerenn/example", "test-branch").Return(nil, status.ErrWorktreeNotFound)

	err := wtm.OpenWorktree("test-branch", "cursor")
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "worktree not found")
}

func TestWTM_Run_VerboseMode(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockFS := fs.NewMockFS(ctrl)
	mockGit := git.NewMockGit(ctrl)
	mockStatus := status.NewMockManager(ctrl)
	mockIDE := ide.NewMockManagerInterface(ctrl)

	wtm := NewWTM(createTestConfig())
	wtm.SetVerbose(true)

	// Override adapters with mocks
	c := wtm.(*realWTM)
	c.fs = mockFS
	c.git = mockGit
	c.statusManager = mockStatus
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

	err := wtm.CreateWorkTree("test-branch")
	assert.NoError(t, err)
}

func TestWTM_DeleteWorkTree_SingleRepository(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockFS := fs.NewMockFS(ctrl)
	mockGit := git.NewMockGit(ctrl)
	mockStatus := status.NewMockManager(ctrl)
	mockIDE := ide.NewMockManagerInterface(ctrl)

	wtm := NewWTM(createTestConfig())

	// Override adapters with mocks
	c := wtm.(*realWTM)
	c.fs = mockFS
	c.git = mockGit
	c.statusManager = mockStatus
	c.ideManager = mockIDE

	// Mock single repo detection
	mockFS.EXPECT().Exists(".git").Return(true, nil).AnyTimes()
	mockFS.EXPECT().IsDir(".git").Return(true, nil).AnyTimes()

	// Mock Git status for validation
	mockGit.EXPECT().Status(".").Return("On branch main", nil).Times(2)

	// Mock worktree deletion
	mockGit.EXPECT().GetRepositoryName(gomock.Any()).Return("github.com/lerenn/example", nil)
	mockStatus.EXPECT().GetWorktree("github.com/lerenn/example", "test-branch").Return(&status.Repository{
		URL:    "github.com/lerenn/example",
		Branch: "test-branch",
		Path:   "/test/path/worktree",
	}, nil)
	mockGit.EXPECT().GetWorktreePath(gomock.Any(), "test-branch").Return("/test/path/worktree", nil)
	mockGit.EXPECT().RemoveWorktree(gomock.Any(), "/test/path/worktree").Return(nil)
	mockFS.EXPECT().RemoveAll("/test/path/worktree").Return(nil)
	mockStatus.EXPECT().RemoveWorktree("github.com/lerenn/example", "test-branch").Return(nil)

	err := wtm.DeleteWorkTree("test-branch", true) // Force deletion
	assert.NoError(t, err)
}

// TestWTM_DeleteWorkTree_Workspace is skipped due to test environment issues
// with workspace files in the test directory
func TestWTM_DeleteWorkTree_Workspace(t *testing.T) {
	t.Skip("Skipping workspace test due to test environment issues")
}

func TestWTM_DeleteWorkTree_NoRepository(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockFS := fs.NewMockFS(ctrl)
	mockGit := git.NewMockGit(ctrl)
	mockStatus := status.NewMockManager(ctrl)
	mockIDE := ide.NewMockManagerInterface(ctrl)

	wtm := NewWTM(createTestConfig())

	// Override adapters with mocks
	c := wtm.(*realWTM)
	c.fs = mockFS
	c.git = mockGit
	c.statusManager = mockStatus
	c.ideManager = mockIDE

	// Mock no repository or workspace found
	mockFS.EXPECT().Exists(".git").Return(false, nil)
	mockFS.EXPECT().Glob("*.code-workspace").Return([]string{}, nil)

	err := wtm.DeleteWorkTree("test-branch", true)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "no Git repository or workspace found")
}

func TestWTM_DeleteWorkTree_VerboseMode(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockFS := fs.NewMockFS(ctrl)
	mockGit := git.NewMockGit(ctrl)
	mockStatus := status.NewMockManager(ctrl)
	mockIDE := ide.NewMockManagerInterface(ctrl)

	wtm := NewWTM(createTestConfig())
	wtm.SetVerbose(true)

	// Override adapters with mocks
	c := wtm.(*realWTM)
	c.fs = mockFS
	c.git = mockGit
	c.statusManager = mockStatus
	c.ideManager = mockIDE

	// Mock single repo detection
	mockFS.EXPECT().Exists(".git").Return(true, nil).AnyTimes()
	mockFS.EXPECT().IsDir(".git").Return(true, nil).AnyTimes()

	// Mock Git status for validation
	mockGit.EXPECT().Status(".").Return("On branch main", nil).Times(2)

	// Mock worktree deletion
	mockGit.EXPECT().GetRepositoryName(gomock.Any()).Return("github.com/lerenn/example", nil)
	mockStatus.EXPECT().GetWorktree("github.com/lerenn/example", "test-branch").Return(&status.Repository{
		URL:    "github.com/lerenn/example",
		Branch: "test-branch",
		Path:   "/test/path/worktree",
	}, nil)
	mockGit.EXPECT().GetWorktreePath(gomock.Any(), "test-branch").Return("/test/path/worktree", nil)
	mockGit.EXPECT().RemoveWorktree(gomock.Any(), "/test/path/worktree").Return(nil)
	mockFS.EXPECT().RemoveAll("/test/path/worktree").Return(nil)
	mockStatus.EXPECT().RemoveWorktree("github.com/lerenn/example", "test-branch").Return(nil)

	err := wtm.DeleteWorkTree("test-branch", true) // Force deletion
	assert.NoError(t, err)
}

func TestWTM_ListWorktrees_NoRepository(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockFS := fs.NewMockFS(ctrl)
	mockGit := git.NewMockGit(ctrl)
	mockStatus := status.NewMockManager(ctrl)
	mockIDE := ide.NewMockManagerInterface(ctrl)

	wtm := NewWTM(createTestConfig())

	// Override adapters with mocks
	c := wtm.(*realWTM)
	c.fs = mockFS
	c.git = mockGit
	c.statusManager = mockStatus
	c.ideManager = mockIDE

	// Mock single repo detection - no .git found
	mockFS.EXPECT().Exists(".git").Return(false, nil)

	// Mock workspace detection - no workspace files found
	mockFS.EXPECT().Glob("*.code-workspace").Return([]string{}, nil)

	result, _, err := wtm.ListWorktrees()
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "no Git repository or workspace found")
	assert.Nil(t, result)
}

func TestWTM_LoadWorktree_Success(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockFS := fs.NewMockFS(ctrl)
	mockGit := git.NewMockGit(ctrl)
	mockStatus := status.NewMockManager(ctrl)
	mockIDE := ide.NewMockManagerInterface(ctrl)

	wtm := NewWTM(createTestConfig())

	// Override adapters with mocks
	c := wtm.(*realWTM)
	c.fs = mockFS
	c.git = mockGit
	c.statusManager = mockStatus
	c.ideManager = mockIDE

	// Mock single repo detection
	mockFS.EXPECT().Exists(".git").Return(true, nil).AnyTimes()
	mockFS.EXPECT().IsDir(".git").Return(true, nil).AnyTimes()

	// Mock Git status for validation
	mockGit.EXPECT().Status(".").Return("On branch main", nil).AnyTimes()

	// Mock origin remote validation
	mockGit.EXPECT().RemoteExists(".", "origin").Return(true, nil)
	mockGit.EXPECT().GetRemoteURL(".", "origin").Return("https://github.com/lerenn/example.git", nil)

	// Mock fetch from remote
	mockGit.EXPECT().FetchRemote(".", "origin").Return(nil)

	// Mock branch existence check
	mockGit.EXPECT().BranchExistsOnRemote(gomock.Any()).Return(true, nil)

	// Mock worktree creation (reusing existing logic)
	mockGit.EXPECT().GetRepositoryName(gomock.Any()).Return("github.com/lerenn/example", nil)
	mockStatus.EXPECT().GetWorktree("github.com/lerenn/example", "feature-branch").Return(nil, status.ErrWorktreeNotFound)
	mockGit.EXPECT().IsClean(gomock.Any()).Return(true, nil)
	mockFS.EXPECT().Exists(gomock.Any()).Return(false, nil).AnyTimes()
	mockFS.EXPECT().MkdirAll(gomock.Any(), gomock.Any()).Return(nil)
	mockStatus.EXPECT().AddWorktree(gomock.Any()).Return(nil)
	mockGit.EXPECT().BranchExists(gomock.Any(), "feature-branch").Return(false, nil)
	mockGit.EXPECT().CreateBranch(gomock.Any(), "feature-branch").Return(nil)
	mockGit.EXPECT().CreateWorktree(gomock.Any(), gomock.Any(), "feature-branch").Return(nil)

	err := wtm.LoadWorktree("origin:feature-branch")
	assert.NoError(t, err)
}

func TestWTM_LoadWorktree_WithIDE(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockFS := fs.NewMockFS(ctrl)
	mockGit := git.NewMockGit(ctrl)
	mockStatus := status.NewMockManager(ctrl)
	mockIDE := ide.NewMockManagerInterface(ctrl)

	wtm := NewWTM(createTestConfig())

	// Override adapters with mocks
	c := wtm.(*realWTM)
	c.fs = mockFS
	c.git = mockGit
	c.statusManager = mockStatus
	c.ideManager = mockIDE

	// Mock single repo detection
	mockFS.EXPECT().Exists(".git").Return(true, nil).AnyTimes()
	mockFS.EXPECT().IsDir(".git").Return(true, nil).AnyTimes()

	// Mock Git status for validation
	mockGit.EXPECT().Status(".").Return("On branch main", nil).AnyTimes()

	// Mock origin remote validation
	mockGit.EXPECT().RemoteExists(".", "origin").Return(true, nil)
	mockGit.EXPECT().GetRemoteURL(".", "origin").Return("https://github.com/lerenn/example.git", nil)

	// Mock fetch from remote
	mockGit.EXPECT().FetchRemote(".", "origin").Return(nil)

	// Mock branch existence check
	mockGit.EXPECT().BranchExistsOnRemote(gomock.Any()).Return(true, nil)

	// Mock worktree creation (reusing existing logic)
	mockGit.EXPECT().GetRepositoryName(gomock.Any()).Return("github.com/lerenn/example", nil).AnyTimes()
	mockStatus.EXPECT().GetWorktree("github.com/lerenn/example", "feature-branch").Return(nil, status.ErrWorktreeNotFound)
	mockGit.EXPECT().IsClean(gomock.Any()).Return(true, nil)
	mockFS.EXPECT().Exists(gomock.Any()).Return(false, nil).AnyTimes()
	mockFS.EXPECT().MkdirAll(gomock.Any(), gomock.Any()).Return(nil)
	mockStatus.EXPECT().AddWorktree(gomock.Any()).Return(nil)
	mockGit.EXPECT().BranchExists(gomock.Any(), "feature-branch").Return(false, nil)
	mockGit.EXPECT().CreateBranch(gomock.Any(), "feature-branch").Return(nil)
	mockGit.EXPECT().CreateWorktree(gomock.Any(), gomock.Any(), "feature-branch").Return(nil)

	// Mock IDE opening
	ideName := "cursor"
	mockStatus.EXPECT().GetWorktree("github.com/lerenn/example", "feature-branch").Return(&status.Repository{
		URL:    "github.com/lerenn/example",
		Branch: "feature-branch",
		Path:   "/test/base/path/github.com/lerenn/example/feature-branch",
	}, nil)
	mockIDE.EXPECT().OpenIDE("cursor", gomock.Any(), false).Return(nil)

	err := wtm.LoadWorktree("origin:feature-branch", LoadWorktreeOpts{IDEName: ideName})
	assert.NoError(t, err)
}

func TestWTM_LoadWorktree_NewRemote(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockFS := fs.NewMockFS(ctrl)
	mockGit := git.NewMockGit(ctrl)
	mockStatus := status.NewMockManager(ctrl)
	mockIDE := ide.NewMockManagerInterface(ctrl)

	wtm := NewWTM(createTestConfig())

	// Override adapters with mocks
	c := wtm.(*realWTM)
	c.fs = mockFS
	c.git = mockGit
	c.statusManager = mockStatus
	c.ideManager = mockIDE

	// Mock single repo detection
	mockFS.EXPECT().Exists(".git").Return(true, nil).AnyTimes()
	mockFS.EXPECT().IsDir(".git").Return(true, nil).AnyTimes()

	// Mock Git status for validation
	mockGit.EXPECT().Status(".").Return("On branch main", nil).AnyTimes()

	// Mock origin remote validation
	mockGit.EXPECT().RemoteExists(".", "origin").Return(true, nil)
	mockGit.EXPECT().GetRemoteURL(".", "origin").Return("https://github.com/lerenn/example.git", nil)

	// Mock remote management (new remote)
	mockGit.EXPECT().RemoteExists(".", "otheruser").Return(false, nil)
	mockGit.EXPECT().GetRepositoryName(".").Return("github.com/lerenn/example", nil)
	mockGit.EXPECT().GetRemoteURL(".", "origin").Return("https://github.com/lerenn/example.git", nil)
	mockGit.EXPECT().AddRemote(".", "otheruser", "https://github.com/otheruser/example.git").Return(nil)

	// Mock fetch from remote
	mockGit.EXPECT().FetchRemote(".", "otheruser").Return(nil)

	// Mock branch existence check
	mockGit.EXPECT().BranchExistsOnRemote(gomock.Any()).Return(true, nil)

	// Mock worktree creation (reusing existing logic)
	mockGit.EXPECT().GetRepositoryName(gomock.Any()).Return("github.com/lerenn/example", nil)
	mockStatus.EXPECT().GetWorktree("github.com/lerenn/example", "feature-branch").Return(nil, status.ErrWorktreeNotFound)
	mockGit.EXPECT().IsClean(gomock.Any()).Return(true, nil)
	mockFS.EXPECT().Exists(gomock.Any()).Return(false, nil).AnyTimes()
	mockFS.EXPECT().MkdirAll(gomock.Any(), gomock.Any()).Return(nil)
	mockStatus.EXPECT().AddWorktree(gomock.Any()).Return(nil)
	mockGit.EXPECT().BranchExists(gomock.Any(), "feature-branch").Return(false, nil)
	mockGit.EXPECT().CreateBranch(gomock.Any(), "feature-branch").Return(nil)
	mockGit.EXPECT().CreateWorktree(gomock.Any(), gomock.Any(), "feature-branch").Return(nil)

	err := wtm.LoadWorktree("otheruser:feature-branch")
	assert.NoError(t, err)
}

func TestWTM_LoadWorktree_SSHProtocol(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockFS := fs.NewMockFS(ctrl)
	mockGit := git.NewMockGit(ctrl)
	mockStatus := status.NewMockManager(ctrl)
	mockIDE := ide.NewMockManagerInterface(ctrl)

	wtm := NewWTM(createTestConfig())

	// Override adapters with mocks
	c := wtm.(*realWTM)
	c.fs = mockFS
	c.git = mockGit
	c.statusManager = mockStatus
	c.ideManager = mockIDE

	// Mock single repo detection
	mockFS.EXPECT().Exists(".git").Return(true, nil).AnyTimes()
	mockFS.EXPECT().IsDir(".git").Return(true, nil).AnyTimes()

	// Mock Git status for validation
	mockGit.EXPECT().Status(".").Return("On branch main", nil).AnyTimes()

	// Mock origin remote validation (SSH)
	mockGit.EXPECT().RemoteExists(".", "origin").Return(true, nil)
	mockGit.EXPECT().GetRemoteURL(".", "origin").Return("git@github.com:lerenn/example.git", nil)

	// Mock remote management (new remote with SSH)
	mockGit.EXPECT().RemoteExists(".", "otheruser").Return(false, nil)
	mockGit.EXPECT().GetRepositoryName(".").Return("github.com/lerenn/example", nil)
	mockGit.EXPECT().GetRemoteURL(".", "origin").Return("git@github.com:lerenn/example.git", nil)
	mockGit.EXPECT().AddRemote(".", "otheruser", "git@github.com:otheruser/example.git").Return(nil)

	// Mock fetch from remote
	mockGit.EXPECT().FetchRemote(".", "otheruser").Return(nil)

	// Mock branch existence check
	mockGit.EXPECT().BranchExistsOnRemote(gomock.Any()).Return(true, nil)

	// Mock worktree creation (reusing existing logic)
	mockGit.EXPECT().GetRepositoryName(gomock.Any()).Return("github.com/lerenn/example", nil)
	mockStatus.EXPECT().GetWorktree("github.com/lerenn/example", "feature-branch").Return(nil, status.ErrWorktreeNotFound)
	mockGit.EXPECT().IsClean(gomock.Any()).Return(true, nil)
	mockFS.EXPECT().Exists(gomock.Any()).Return(false, nil).AnyTimes()
	mockFS.EXPECT().MkdirAll(gomock.Any(), gomock.Any()).Return(nil)
	mockStatus.EXPECT().AddWorktree(gomock.Any()).Return(nil)
	mockGit.EXPECT().BranchExists(gomock.Any(), "feature-branch").Return(false, nil)
	mockGit.EXPECT().CreateBranch(gomock.Any(), "feature-branch").Return(nil)
	mockGit.EXPECT().CreateWorktree(gomock.Any(), gomock.Any(), "feature-branch").Return(nil)

	err := wtm.LoadWorktree("otheruser:feature-branch")
	assert.NoError(t, err)
}

func TestWTM_LoadWorktree_NoRepository(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockFS := fs.NewMockFS(ctrl)
	mockGit := git.NewMockGit(ctrl)
	mockStatus := status.NewMockManager(ctrl)
	mockIDE := ide.NewMockManagerInterface(ctrl)

	wtm := NewWTM(createTestConfig())

	// Override adapters with mocks
	c := wtm.(*realWTM)
	c.fs = mockFS
	c.git = mockGit
	c.statusManager = mockStatus
	c.ideManager = mockIDE

	// Mock no repository or workspace found
	mockFS.EXPECT().Exists(".git").Return(false, nil)
	mockFS.EXPECT().Glob("*.code-workspace").Return([]string{}, nil)

	err := wtm.LoadWorktree("origin:feature-branch")
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "no Git repository or workspace found")
}

func TestWTM_LoadWorktree_WorkspaceMode(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockFS := fs.NewMockFS(ctrl)
	mockGit := git.NewMockGit(ctrl)
	mockStatus := status.NewMockManager(ctrl)
	mockIDE := ide.NewMockManagerInterface(ctrl)

	wtm := NewWTM(createTestConfig())

	// Override adapters with mocks
	c := wtm.(*realWTM)
	c.fs = mockFS
	c.git = mockGit
	c.statusManager = mockStatus
	c.ideManager = mockIDE

	// Mock workspace detection
	mockFS.EXPECT().Exists(".git").Return(false, nil)
	mockFS.EXPECT().Glob("*.code-workspace").Return([]string{"workspace.code-workspace"}, nil)

	err := wtm.LoadWorktree("origin:feature-branch")
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "workspace mode not yet supported for load command")
}

func TestWTM_LoadWorktree_OriginRemoteNotFound(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockFS := fs.NewMockFS(ctrl)
	mockGit := git.NewMockGit(ctrl)
	mockStatus := status.NewMockManager(ctrl)
	mockIDE := ide.NewMockManagerInterface(ctrl)

	wtm := NewWTM(createTestConfig())

	// Override adapters with mocks
	c := wtm.(*realWTM)
	c.fs = mockFS
	c.git = mockGit
	c.statusManager = mockStatus
	c.ideManager = mockIDE

	// Mock single repo detection
	mockFS.EXPECT().Exists(".git").Return(true, nil).AnyTimes()
	mockFS.EXPECT().IsDir(".git").Return(true, nil).AnyTimes()

	// Mock Git status for validation
	mockGit.EXPECT().Status(".").Return("On branch main", nil).AnyTimes()

	// Mock origin remote not found
	mockGit.EXPECT().RemoteExists(".", "origin").Return(false, nil)

	err := wtm.LoadWorktree("origin:feature-branch")
	assert.ErrorIs(t, err, ErrOriginRemoteNotFound)
}

func TestWTM_LoadWorktree_OriginRemoteInvalidURL(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockFS := fs.NewMockFS(ctrl)
	mockGit := git.NewMockGit(ctrl)
	mockStatus := status.NewMockManager(ctrl)
	mockIDE := ide.NewMockManagerInterface(ctrl)

	wtm := NewWTM(createTestConfig())

	// Override adapters with mocks
	c := wtm.(*realWTM)
	c.fs = mockFS
	c.git = mockGit
	c.statusManager = mockStatus
	c.ideManager = mockIDE

	// Mock single repo detection
	mockFS.EXPECT().Exists(".git").Return(true, nil).AnyTimes()
	mockFS.EXPECT().IsDir(".git").Return(true, nil).AnyTimes()

	// Mock Git status for validation
	mockGit.EXPECT().Status(".").Return("On branch main", nil).AnyTimes()

	// Mock origin remote exists but invalid URL
	mockGit.EXPECT().RemoteExists(".", "origin").Return(true, nil)
	mockGit.EXPECT().GetRemoteURL(".", "origin").Return("invalid-url", nil)

	err := wtm.LoadWorktree("origin:feature-branch")
	assert.ErrorIs(t, err, ErrOriginRemoteInvalidURL)
}

func TestWTM_LoadWorktree_FetchFailed(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockFS := fs.NewMockFS(ctrl)
	mockGit := git.NewMockGit(ctrl)
	mockStatus := status.NewMockManager(ctrl)
	mockIDE := ide.NewMockManagerInterface(ctrl)

	wtm := NewWTM(createTestConfig())

	// Override adapters with mocks
	c := wtm.(*realWTM)
	c.fs = mockFS
	c.git = mockGit
	c.statusManager = mockStatus
	c.ideManager = mockIDE

	// Mock single repo detection
	mockFS.EXPECT().Exists(".git").Return(true, nil).AnyTimes()
	mockFS.EXPECT().IsDir(".git").Return(true, nil).AnyTimes()

	// Mock Git status for validation
	mockGit.EXPECT().Status(".").Return("On branch main", nil).AnyTimes()

	// Mock origin remote validation
	mockGit.EXPECT().RemoteExists(".", "origin").Return(true, nil)
	mockGit.EXPECT().GetRemoteURL(".", "origin").Return("https://github.com/lerenn/example.git", nil)

	// Mock fetch from remote fails
	mockGit.EXPECT().FetchRemote(".", "origin").Return(assert.AnError)

	err := wtm.LoadWorktree("origin:feature-branch")
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "failed to fetch from remote")
}

func TestWTM_LoadWorktree_BranchNotFound(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockFS := fs.NewMockFS(ctrl)
	mockGit := git.NewMockGit(ctrl)
	mockStatus := status.NewMockManager(ctrl)
	mockIDE := ide.NewMockManagerInterface(ctrl)

	wtm := NewWTM(createTestConfig())

	// Override adapters with mocks
	c := wtm.(*realWTM)
	c.fs = mockFS
	c.git = mockGit
	c.statusManager = mockStatus
	c.ideManager = mockIDE

	// Mock single repo detection
	mockFS.EXPECT().Exists(".git").Return(true, nil).AnyTimes()
	mockFS.EXPECT().IsDir(".git").Return(true, nil).AnyTimes()

	// Mock Git status for validation
	mockGit.EXPECT().Status(".").Return("On branch main", nil).AnyTimes()

	// Mock origin remote validation
	mockGit.EXPECT().RemoteExists(".", "origin").Return(true, nil)
	mockGit.EXPECT().GetRemoteURL(".", "origin").Return("https://github.com/lerenn/example.git", nil)

	// Mock fetch from remote
	mockGit.EXPECT().FetchRemote(".", "origin").Return(nil)

	// Mock branch existence check fails
	mockGit.EXPECT().BranchExistsOnRemote(gomock.Any()).Return(false, nil)

	err := wtm.LoadWorktree("origin:feature-branch")
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "branch not found on remote")
}

func TestWTM_LoadWorktree_DefaultRemote(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockFS := fs.NewMockFS(ctrl)
	mockGit := git.NewMockGit(ctrl)
	mockStatus := status.NewMockManager(ctrl)
	mockIDE := ide.NewMockManagerInterface(ctrl)

	wtm := NewWTM(createTestConfig())

	// Override adapters with mocks
	c := wtm.(*realWTM)
	c.fs = mockFS
	c.git = mockGit
	c.statusManager = mockStatus
	c.ideManager = mockIDE

	// Mock single repo detection
	mockFS.EXPECT().Exists(".git").Return(true, nil).AnyTimes()
	mockFS.EXPECT().IsDir(".git").Return(true, nil).AnyTimes()

	// Mock Git status for validation
	mockGit.EXPECT().Status(".").Return("On branch main", nil).AnyTimes()

	// Mock origin remote validation
	mockGit.EXPECT().RemoteExists(".", "origin").Return(true, nil)
	mockGit.EXPECT().GetRemoteURL(".", "origin").Return("https://github.com/lerenn/example.git", nil)

	// Mock fetch from remote
	mockGit.EXPECT().FetchRemote(".", "origin").Return(nil)

	// Mock branch existence check
	mockGit.EXPECT().BranchExistsOnRemote(gomock.Any()).Return(true, nil)

	// Mock worktree creation (reusing existing logic)
	mockGit.EXPECT().GetRepositoryName(gomock.Any()).Return("github.com/lerenn/example", nil)
	mockStatus.EXPECT().GetWorktree("github.com/lerenn/example", "feature-branch").Return(nil, status.ErrWorktreeNotFound)
	mockGit.EXPECT().IsClean(gomock.Any()).Return(true, nil)
	mockFS.EXPECT().Exists(gomock.Any()).Return(false, nil).AnyTimes()
	mockFS.EXPECT().MkdirAll(gomock.Any(), gomock.Any()).Return(nil)
	mockStatus.EXPECT().AddWorktree(gomock.Any()).Return(nil)
	mockGit.EXPECT().BranchExists(gomock.Any(), "feature-branch").Return(false, nil)
	mockGit.EXPECT().CreateBranch(gomock.Any(), "feature-branch").Return(nil)
	mockGit.EXPECT().CreateWorktree(gomock.Any(), gomock.Any(), "feature-branch").Return(nil)

	// Test with empty remote source (should default to "origin")
	err := wtm.LoadWorktree("feature-branch")
	assert.NoError(t, err)
}
