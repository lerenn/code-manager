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

func TestCM_LoadWorktree_Success(t *testing.T) {
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
	mockGit.EXPECT().BranchExistsOnRemote(git.BranchExistsOnRemoteParams{
		RepoPath:   ".",
		RemoteName: "origin",
		Branch:     "feature-branch",
	}).Return(true, nil)

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

	err := cm.LoadWorktree("origin:feature-branch")
	assert.NoError(t, err)
}

func TestCM_LoadWorktree_WithIDE(t *testing.T) {
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
	mockGit.EXPECT().BranchExistsOnRemote(git.BranchExistsOnRemoteParams{
		RepoPath:   ".",
		RemoteName: "origin",
		Branch:     "feature-branch",
	}).Return(true, nil)

	// Mock worktree creation (reusing existing logic)
	mockGit.EXPECT().GetRepositoryName(gomock.Any()).Return("github.com/lerenn/example", nil).AnyTimes()
	mockStatus.EXPECT().GetWorktree("github.com/lerenn/example", "feature-branch").Return(nil, status.ErrWorktreeNotFound).AnyTimes()
	mockGit.EXPECT().IsClean(gomock.Any()).Return(true, nil)
	// Mock worktree directory doesn't exist during creation
	mockFS.EXPECT().Exists("/test/base/path/worktrees/github.com/lerenn/example/feature-branch").Return(false, nil)
	mockFS.EXPECT().MkdirAll(gomock.Any(), gomock.Any()).Return(nil)
	mockStatus.EXPECT().AddWorktree(gomock.Any()).Return(nil)
	mockGit.EXPECT().BranchExists(gomock.Any(), "feature-branch").Return(false, nil)
	mockGit.EXPECT().CreateBranch(gomock.Any(), "feature-branch").Return(nil)
	mockGit.EXPECT().CreateWorktree(gomock.Any(), gomock.Any(), "feature-branch").Return(nil)

	// Mock worktree path existence for OpenWorktree call (after creation)
	mockFS.EXPECT().Exists("/test/base/path/worktrees/github.com/lerenn/example/feature-branch").Return(true, nil)

	// Mock IDE opening
	ideName := "cursor"
	mockIDE.EXPECT().OpenIDE("cursor", "/test/base/path/worktrees/github.com/lerenn/example/feature-branch", false).Return(nil)

	err := cm.LoadWorktree("origin:feature-branch", LoadWorktreeOpts{IDEName: ideName})
	assert.NoError(t, err)
}

func TestCM_LoadWorktree_NewRemote(t *testing.T) {
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
	mockGit.EXPECT().BranchExistsOnRemote(git.BranchExistsOnRemoteParams{
		RepoPath:   ".",
		RemoteName: "otheruser",
		Branch:     "feature-branch",
	}).Return(true, nil)

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

	err := cm.LoadWorktree("otheruser:feature-branch")
	assert.NoError(t, err)
}

func TestCM_LoadWorktree_SSHProtocol(t *testing.T) {
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
	mockGit.EXPECT().BranchExistsOnRemote(git.BranchExistsOnRemoteParams{
		RepoPath:   ".",
		RemoteName: "otheruser",
		Branch:     "feature-branch",
	}).Return(true, nil)

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

	err := cm.LoadWorktree("otheruser:feature-branch")
	assert.NoError(t, err)
}

func TestCM_LoadWorktree_NoRepository(t *testing.T) {
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

	// Mock no repository or workspace found
	mockFS.EXPECT().Exists(".git").Return(false, nil)
	mockFS.EXPECT().Glob("*.code-workspace").Return([]string{}, nil)

	err := cm.LoadWorktree("origin:feature-branch")
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "no Git repository or workspace found")
}

func TestCM_LoadWorktree_WorkspaceMode(t *testing.T) {
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

	// Mock workspace detection
	mockFS.EXPECT().Exists(".git").Return(false, nil)
	mockFS.EXPECT().Glob("*.code-workspace").Return([]string{"workspace.code-workspace"}, nil)

	err := cm.LoadWorktree("origin:feature-branch")
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "workspace mode not yet supported for load command")
}

func TestCM_LoadWorktree_OriginRemoteNotFound(t *testing.T) {
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

	// Mock single repo detection
	mockFS.EXPECT().Exists(".git").Return(true, nil).AnyTimes()
	mockFS.EXPECT().IsDir(".git").Return(true, nil).AnyTimes()

	// Mock Git status for validation
	mockGit.EXPECT().Status(".").Return("On branch main", nil).AnyTimes()

	// Mock origin remote not found
	mockGit.EXPECT().RemoteExists(".", "origin").Return(false, nil)

	err := cm.LoadWorktree("origin:feature-branch")
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "origin remote not found or invalid")
}

func TestCM_LoadWorktree_OriginRemoteInvalidURL(t *testing.T) {
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

	// Mock single repo detection
	mockFS.EXPECT().Exists(".git").Return(true, nil).AnyTimes()
	mockFS.EXPECT().IsDir(".git").Return(true, nil).AnyTimes()

	// Mock Git status for validation
	mockGit.EXPECT().Status(".").Return("On branch main", nil).AnyTimes()

	// Mock origin remote exists but invalid URL
	mockGit.EXPECT().RemoteExists(".", "origin").Return(true, nil)
	mockGit.EXPECT().GetRemoteURL(".", "origin").Return("invalid-url", nil)

	err := cm.LoadWorktree("origin:feature-branch")
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "origin remote URL is not a valid Git hosting service URL")
}

func TestCM_LoadWorktree_FetchFailed(t *testing.T) {
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

	err := cm.LoadWorktree("origin:feature-branch")
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "failed to fetch from remote")
}

func TestCM_LoadWorktree_BranchNotFound(t *testing.T) {
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
	mockGit.EXPECT().BranchExistsOnRemote(git.BranchExistsOnRemoteParams{
		RepoPath:   ".",
		RemoteName: "origin",
		Branch:     "feature-branch",
	}).Return(false, nil)

	err := cm.LoadWorktree("origin:feature-branch")
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "branch not found on remote")
}

func TestCM_LoadWorktree_DefaultRemote(t *testing.T) {
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
	mockGit.EXPECT().BranchExistsOnRemote(git.BranchExistsOnRemoteParams{
		RepoPath:   ".",
		RemoteName: "origin",
		Branch:     "feature-branch",
	}).Return(true, nil)

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
	err := cm.LoadWorktree("feature-branch")
	assert.NoError(t, err)
}
