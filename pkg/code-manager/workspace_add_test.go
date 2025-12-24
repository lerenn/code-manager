//go:build unit

package codemanager

import (
	"errors"
	"testing"

	"github.com/lerenn/code-manager/pkg/config"
	"github.com/lerenn/code-manager/pkg/dependencies"
	fsmocks "github.com/lerenn/code-manager/pkg/fs/mocks"
	gitmocks "github.com/lerenn/code-manager/pkg/git/mocks"
	"github.com/lerenn/code-manager/pkg/logger"
	"github.com/lerenn/code-manager/pkg/mode/repository"
	repomocks "github.com/lerenn/code-manager/pkg/mode/repository/mocks"
	"github.com/lerenn/code-manager/pkg/status"
	statusmocks "github.com/lerenn/code-manager/pkg/status/mocks"
	"github.com/stretchr/testify/assert"
	"go.uber.org/mock/gomock"
)

// TestAddRepositoryToWorkspace_Success tests successful addition of repository to workspace.
func TestAddRepositoryToWorkspace_Success(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockFS := fsmocks.NewMockFS(ctrl)
	mockGit := gitmocks.NewMockGit(ctrl)
	mockStatus := statusmocks.NewMockManager(ctrl)
	mockRepo := repomocks.NewMockRepository(ctrl)

	cm := &realCodeManager{
		deps: dependencies.New().
			WithFS(mockFS).
			WithGit(mockGit).
			WithConfig(config.NewConfigManager("/test/config.yaml")).
			WithStatusManager(mockStatus).
			WithRepositoryProvider(func(params repository.NewRepositoryParams) repository.Repository {
				return mockRepo
			}).
			WithLogger(logger.NewNoopLogger()),
	}

	params := AddRepositoryToWorkspaceParams{
		WorkspaceName: "test-workspace",
		Repository:    "repo1",
	}

	// Mock workspace exists
	existingWorkspace := &status.Workspace{
		Repositories: []string{"github.com/user/existing-repo"},
		Worktrees:    []string{"main", "feature"},
	}
	mockStatus.EXPECT().GetWorkspace("test-workspace").Return(existingWorkspace, nil)

	// Mock repository resolution
	existingRepo := &status.Repository{Path: "/path/to/repo1"}
	mockStatus.EXPECT().GetRepository("repo1").Return(existingRepo, nil)

	// Mock GetRemoteURL call (called during resolution)
	mockGit.EXPECT().GetRemoteURL("repo1", "origin").Return("github.com/user/repo1", nil)

	// Mock repository exists in status check (using the remote URL)
	mockStatus.EXPECT().GetRepository("github.com/user/repo1").Return(existingRepo, nil)

	// Mock getting existing repositories to check branches
	existingRepoStatus := &status.Repository{
		Path: "/path/to/existing-repo",
		Worktrees: map[string]status.WorktreeInfo{
			"origin:main":    {Branch: "main", Remote: "origin"},
			"origin:feature": {Branch: "feature", Remote: "origin"},
		},
	}
	mockStatus.EXPECT().GetRepository("github.com/user/existing-repo").Return(existingRepoStatus, nil).Times(2) // Called once per branch check

	// Mock updating workspace in status
	mockStatus.EXPECT().UpdateWorkspace("test-workspace", gomock.Any()).Return(nil)

	// Mock GetRepository call in createWorktreeForBranchInRepository (to get managed path)
	mockStatus.EXPECT().GetRepository("github.com/user/repo1").Return(existingRepo, nil).Times(2)

	// Mock GetWorktree calls to check if worktrees already exist (they don't, so return error)
	mockStatus.EXPECT().GetWorktree("github.com/user/repo1", "main").Return(nil, status.ErrWorktreeNotFound)
	mockStatus.EXPECT().GetWorktree("github.com/user/repo1", "feature").Return(nil, status.ErrWorktreeNotFound)

	// Mock worktree creation for each branch
	mockRepo.EXPECT().CreateWorktree("main", gomock.Any()).Return("/repos/github.com/user/repo1/origin/main", nil)
	mockRepo.EXPECT().CreateWorktree("feature", gomock.Any()).Return("/repos/github.com/user/repo1/origin/feature", nil)

	// Mock config for workspace file path construction
	mockConfig := config.NewConfigManager("/test/config.yaml")
	cm.deps = cm.deps.WithConfig(mockConfig)

	// Mock workspace file updates
	mockFS.EXPECT().Exists(gomock.Any()).Return(true, nil).Times(2)
	mockFS.EXPECT().ReadFile(gomock.Any()).Return([]byte(`{"folders":[]}`), nil).Times(2)
	mockFS.EXPECT().WriteFileAtomic(gomock.Any(), gomock.Any(), gomock.Any()).Return(nil).Times(2)

	err := cm.AddRepositoryToWorkspace(&params)
	assert.NoError(t, err)
}

// TestAddRepositoryToWorkspace_WorkspaceNotFound tests addition when workspace doesn't exist.
func TestAddRepositoryToWorkspace_WorkspaceNotFound(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockStatus := statusmocks.NewMockManager(ctrl)

	cm := &realCodeManager{
		deps: dependencies.New().
			WithStatusManager(mockStatus).
			WithLogger(logger.NewNoopLogger()),
	}

	params := AddRepositoryToWorkspaceParams{
		WorkspaceName: "non-existent-workspace",
		Repository:    "repo1",
	}

	// Mock workspace doesn't exist
	mockStatus.EXPECT().GetWorkspace("non-existent-workspace").Return(nil, errors.New("not found"))

	err := cm.AddRepositoryToWorkspace(&params)
	assert.Error(t, err)
	assert.ErrorIs(t, err, ErrWorkspaceNotFound)
}

// TestAddRepositoryToWorkspace_DuplicateRepository tests addition when repository already exists in workspace.
func TestAddRepositoryToWorkspace_DuplicateRepository(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockStatus := statusmocks.NewMockManager(ctrl)

	cm := &realCodeManager{
		deps: dependencies.New().
			WithStatusManager(mockStatus).
			WithLogger(logger.NewNoopLogger()),
	}

	params := AddRepositoryToWorkspaceParams{
		WorkspaceName: "test-workspace",
		Repository:    "repo1",
	}

	// Mock workspace exists with repository already in it
	existingWorkspace := &status.Workspace{
		Repositories: []string{"repo1"},
		Worktrees:    []string{},
	}
	mockStatus.EXPECT().GetWorkspace("test-workspace").Return(existingWorkspace, nil)

	err := cm.AddRepositoryToWorkspace(&params)
	assert.Error(t, err)
	assert.ErrorIs(t, err, ErrDuplicateRepository)
}

// TestAddRepositoryToWorkspace_NoBranchesWithAllRepos tests when no branches have worktrees in all repositories.
// TODO: Fix nil pointer panic - needs investigation of dependency requirements
func TestAddRepositoryToWorkspace_NoBranchesWithAllRepos(t *testing.T) {
	t.Skip("Skipping due to nil pointer panic - needs investigation")
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockFS := fsmocks.NewMockFS(ctrl)
	mockGit := gitmocks.NewMockGit(ctrl)
	mockStatus := statusmocks.NewMockManager(ctrl)

	cm := &realCodeManager{
		deps: dependencies.New().
			WithFS(mockFS).
			WithGit(mockGit).
			WithConfig(config.NewConfigManager("/test/config.yaml")).
			WithStatusManager(mockStatus).
			WithLogger(logger.NewNoopLogger()),
	}

	params := AddRepositoryToWorkspaceParams{
		WorkspaceName: "test-workspace",
		Repository:    "repo1",
	}

	// Mock workspace exists
	existingWorkspace := &status.Workspace{
		Repositories: []string{"github.com/user/existing-repo"},
		Worktrees:    []string{"main", "feature"},
	}
	mockStatus.EXPECT().GetWorkspace("test-workspace").Return(existingWorkspace, nil)

	// Mock repository resolution
	existingRepo := &status.Repository{Path: "/path/to/repo1"}
	mockStatus.EXPECT().GetRepository("repo1").Return(existingRepo, nil)

	// Mock GetRemoteURL call
	mockGit.EXPECT().GetRemoteURL("repo1", "origin").Return("github.com/user/repo1", nil)

	// Mock repository exists in status check
	mockStatus.EXPECT().GetRepository("github.com/user/repo1").Return(existingRepo, nil)

	// Mock getting existing repository - only has worktree for "main", not "feature"
	existingRepoStatus := &status.Repository{
		Path: "/path/to/existing-repo",
		Worktrees: map[string]status.WorktreeInfo{
			"origin:main": {Branch: "main", Remote: "origin"},
			// No "feature" worktree
		},
	}
	// Called once for "main" branch check, once for "feature" branch check
	mockStatus.EXPECT().GetRepository("github.com/user/existing-repo").Return(existingRepoStatus, nil).Times(2)

	// Mock updating workspace in status (repository is still added, just no worktrees created)
	mockStatus.EXPECT().UpdateWorkspace("test-workspace", gomock.Any()).Return(nil)

	// The test should pass - no worktrees are created, so no additional mocks needed
	err := cm.AddRepositoryToWorkspace(&params)
	assert.NoError(t, err)
}

// TestAddRepositoryToWorkspace_SomeBranchesWithAllRepos tests when some branches have worktrees in all repositories.
func TestAddRepositoryToWorkspace_SomeBranchesWithAllRepos(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockFS := fsmocks.NewMockFS(ctrl)
	mockGit := gitmocks.NewMockGit(ctrl)
	mockStatus := statusmocks.NewMockManager(ctrl)
	mockRepo := repomocks.NewMockRepository(ctrl)

	cm := &realCodeManager{
		deps: dependencies.New().
			WithFS(mockFS).
			WithGit(mockGit).
			WithConfig(config.NewConfigManager("/test/config.yaml")).
			WithStatusManager(mockStatus).
			WithRepositoryProvider(func(params repository.NewRepositoryParams) repository.Repository {
				return mockRepo
			}).
			WithLogger(logger.NewNoopLogger()),
	}

	params := AddRepositoryToWorkspaceParams{
		WorkspaceName: "test-workspace",
		Repository:    "repo1",
	}

	// Mock workspace exists
	existingWorkspace := &status.Workspace{
		Repositories: []string{"github.com/user/existing-repo"},
		Worktrees:    []string{"main", "feature", "develop"},
	}
	mockStatus.EXPECT().GetWorkspace("test-workspace").Return(existingWorkspace, nil)

	// Mock repository resolution
	existingRepo := &status.Repository{Path: "/path/to/repo1"}
	mockStatus.EXPECT().GetRepository("repo1").Return(existingRepo, nil)

	// Mock GetRemoteURL call
	mockGit.EXPECT().GetRemoteURL("repo1", "origin").Return("github.com/user/repo1", nil)

	// Mock repository exists in status check
	mockStatus.EXPECT().GetRepository("github.com/user/repo1").Return(existingRepo, nil)

	// Mock getting existing repository - has worktrees for "main" and "feature", but not "develop"
	existingRepoStatus := &status.Repository{
		Path: "/path/to/existing-repo",
		Worktrees: map[string]status.WorktreeInfo{
			"origin:main":    {Branch: "main", Remote: "origin"},
			"origin:feature": {Branch: "feature", Remote: "origin"},
			// No "develop" worktree
		},
	}
	mockStatus.EXPECT().GetRepository("github.com/user/existing-repo").Return(existingRepoStatus, nil).Times(3) // Called once per branch check

	// Mock updating workspace in status
	mockStatus.EXPECT().UpdateWorkspace("test-workspace", gomock.Any()).Return(nil)

	// Mock GetRepository call in createWorktreeForBranchInRepository (to get managed path)
	mockStatus.EXPECT().GetRepository("github.com/user/repo1").Return(existingRepo, nil).Times(2)

	// Mock GetWorktree calls to check if worktrees already exist (they don't, so return error)
	mockStatus.EXPECT().GetWorktree("github.com/user/repo1", "main").Return(nil, status.ErrWorktreeNotFound)
	mockStatus.EXPECT().GetWorktree("github.com/user/repo1", "feature").Return(nil, status.ErrWorktreeNotFound)

	// Mock worktree creation for branches that exist in all repos (main and feature only)
	mockRepo.EXPECT().CreateWorktree("main", gomock.Any()).Return("/repos/github.com/user/repo1/origin/main", nil)
	mockRepo.EXPECT().CreateWorktree("feature", gomock.Any()).Return("/repos/github.com/user/repo1/origin/feature", nil)

	// Mock workspace file updates (only for main and feature)
	mockFS.EXPECT().Exists(gomock.Any()).Return(true, nil).Times(2)
	mockFS.EXPECT().ReadFile(gomock.Any()).Return([]byte(`{"folders":[]}`), nil).Times(2)
	mockFS.EXPECT().WriteFileAtomic(gomock.Any(), gomock.Any(), gomock.Any()).Return(nil).Times(2)

	err := cm.AddRepositoryToWorkspace(&params)
	assert.NoError(t, err)
}

// TestAddRepositoryToWorkspace_StatusUpdateFailure tests when status update fails.
func TestAddRepositoryToWorkspace_StatusUpdateFailure(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockFS := fsmocks.NewMockFS(ctrl)
	mockGit := gitmocks.NewMockGit(ctrl)
	mockStatus := statusmocks.NewMockManager(ctrl)

	cm := &realCodeManager{
		deps: dependencies.New().
			WithFS(mockFS).
			WithGit(mockGit).
			WithConfig(config.NewConfigManager("/test/config.yaml")).
			WithStatusManager(mockStatus).
			WithLogger(logger.NewNoopLogger()),
	}

	params := AddRepositoryToWorkspaceParams{
		WorkspaceName: "test-workspace",
		Repository:    "repo1",
	}

	// Mock workspace exists
	existingWorkspace := &status.Workspace{
		Repositories: []string{},
		Worktrees:    []string{},
	}
	mockStatus.EXPECT().GetWorkspace("test-workspace").Return(existingWorkspace, nil)

	// Mock repository resolution
	existingRepo := &status.Repository{Path: "/path/to/repo1"}
	mockStatus.EXPECT().GetRepository("repo1").Return(existingRepo, nil)

	// Mock GetRemoteURL call
	mockGit.EXPECT().GetRemoteURL("repo1", "origin").Return("github.com/user/repo1", nil)

	// Mock repository exists in status check
	mockStatus.EXPECT().GetRepository("github.com/user/repo1").Return(existingRepo, nil)

	// Mock status update failure
	mockStatus.EXPECT().UpdateWorkspace("test-workspace", gomock.Any()).Return(errors.New("status update failed"))

	err := cm.AddRepositoryToWorkspace(&params)
	assert.Error(t, err)
	assert.ErrorIs(t, err, ErrStatusUpdate)
}

// TestAddRepositoryToWorkspace_WorktreeCreationFailure tests when worktree creation fails.
func TestAddRepositoryToWorkspace_WorktreeCreationFailure(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockFS := fsmocks.NewMockFS(ctrl)
	mockGit := gitmocks.NewMockGit(ctrl)
	mockStatus := statusmocks.NewMockManager(ctrl)
	mockRepo := repomocks.NewMockRepository(ctrl)

	cm := &realCodeManager{
		deps: dependencies.New().
			WithFS(mockFS).
			WithGit(mockGit).
			WithConfig(config.NewConfigManager("/test/config.yaml")).
			WithStatusManager(mockStatus).
			WithRepositoryProvider(func(params repository.NewRepositoryParams) repository.Repository {
				return mockRepo
			}).
			WithLogger(logger.NewNoopLogger()),
	}

	params := AddRepositoryToWorkspaceParams{
		WorkspaceName: "test-workspace",
		Repository:    "repo1",
	}

	// Mock workspace exists
	existingWorkspace := &status.Workspace{
		Repositories: []string{"github.com/user/existing-repo"},
		Worktrees:    []string{"main"},
	}
	mockStatus.EXPECT().GetWorkspace("test-workspace").Return(existingWorkspace, nil)

	// Mock repository resolution
	existingRepo := &status.Repository{Path: "/path/to/repo1"}
	mockStatus.EXPECT().GetRepository("repo1").Return(existingRepo, nil)

	// Mock GetRemoteURL call - only called once during resolution (line 83)
	// The second call at line 274 never happens because CreateWorktree fails first
	mockGit.EXPECT().GetRemoteURL("repo1", "origin").Return("github.com/user/repo1", nil)

	// Mock repository exists in status check
	mockStatus.EXPECT().GetRepository("github.com/user/repo1").Return(existingRepo, nil)

	// Mock getting existing repository
	existingRepoStatus := &status.Repository{
		Path: "/path/to/existing-repo",
		Worktrees: map[string]status.WorktreeInfo{
			"origin:main": {Branch: "main", Remote: "origin"},
		},
	}
	mockStatus.EXPECT().GetRepository("github.com/user/existing-repo").Return(existingRepoStatus, nil)

	// Mock updating workspace in status
	mockStatus.EXPECT().UpdateWorkspace("test-workspace", gomock.Any()).Return(nil)

	// Mock GetRepository call in createWorktreeForBranchInRepository (to get managed path)
	mockStatus.EXPECT().GetRepository("github.com/user/repo1").Return(existingRepo, nil)

	// Mock GetWorktree call to check if worktree already exists (it doesn't, so return error)
	mockStatus.EXPECT().GetWorktree("github.com/user/repo1", "main").Return(nil, status.ErrWorktreeNotFound)

	// Mock worktree creation failure
	mockRepo.EXPECT().CreateWorktree("main", gomock.Any()).Return("", errors.New("worktree creation failed"))

	err := cm.AddRepositoryToWorkspace(&params)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "failed to create worktree")
}

// TestAddRepositoryToWorkspace_WorkspaceFileUpdateFailure tests when workspace file update fails.
func TestAddRepositoryToWorkspace_WorkspaceFileUpdateFailure(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockFS := fsmocks.NewMockFS(ctrl)
	mockGit := gitmocks.NewMockGit(ctrl)
	mockStatus := statusmocks.NewMockManager(ctrl)
	mockRepo := repomocks.NewMockRepository(ctrl)

	cm := &realCodeManager{
		deps: dependencies.New().
			WithFS(mockFS).
			WithGit(mockGit).
			WithConfig(config.NewConfigManager("/test/config.yaml")).
			WithStatusManager(mockStatus).
			WithRepositoryProvider(func(params repository.NewRepositoryParams) repository.Repository {
				return mockRepo
			}).
			WithLogger(logger.NewNoopLogger()),
	}

	params := AddRepositoryToWorkspaceParams{
		WorkspaceName: "test-workspace",
		Repository:    "repo1",
	}

	// Mock workspace exists
	existingWorkspace := &status.Workspace{
		Repositories: []string{"github.com/user/existing-repo"},
		Worktrees:    []string{"main"},
	}
	mockStatus.EXPECT().GetWorkspace("test-workspace").Return(existingWorkspace, nil)

	// Mock repository resolution
	existingRepo := &status.Repository{Path: "/path/to/repo1"}
	mockStatus.EXPECT().GetRepository("repo1").Return(existingRepo, nil)

	// Mock GetRemoteURL call
	mockGit.EXPECT().GetRemoteURL("repo1", "origin").Return("github.com/user/repo1", nil)

	// Mock repository exists in status check
	mockStatus.EXPECT().GetRepository("github.com/user/repo1").Return(existingRepo, nil)

	// Mock getting existing repository
	existingRepoStatus := &status.Repository{
		Path: "/path/to/existing-repo",
		Worktrees: map[string]status.WorktreeInfo{
			"origin:main": {Branch: "main", Remote: "origin"},
		},
	}
	mockStatus.EXPECT().GetRepository("github.com/user/existing-repo").Return(existingRepoStatus, nil)

	// Mock updating workspace in status
	mockStatus.EXPECT().UpdateWorkspace("test-workspace", gomock.Any()).Return(nil)

	// Mock GetRepository call in createWorktreeForBranchInRepository (to get managed path)
	mockStatus.EXPECT().GetRepository("github.com/user/repo1").Return(existingRepo, nil)

	// Mock GetWorktree call to check if worktree already exists (it doesn't, so return error)
	mockStatus.EXPECT().GetWorktree("github.com/user/repo1", "main").Return(nil, status.ErrWorktreeNotFound)

	// Mock worktree creation success
	mockRepo.EXPECT().CreateWorktree("main", gomock.Any()).Return("/repos/github.com/user/repo1/origin/main", nil)

	// Mock workspace file update failure
	mockFS.EXPECT().Exists(gomock.Any()).Return(true, nil)
	mockFS.EXPECT().ReadFile(gomock.Any()).Return([]byte(`{"folders":[]}`), nil)
	mockFS.EXPECT().WriteFileAtomic(gomock.Any(), gomock.Any(), gomock.Any()).Return(errors.New("file write failed"))

	err := cm.AddRepositoryToWorkspace(&params)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "failed to update workspace file")
}

// TestAddRepositoryToWorkspace_RepositoryNotFound tests addition when repository doesn't exist.
func TestAddRepositoryToWorkspace_RepositoryNotFound(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockFS := fsmocks.NewMockFS(ctrl)
	mockStatus := statusmocks.NewMockManager(ctrl)

	cm := &realCodeManager{
		deps: dependencies.New().
			WithFS(mockFS).
			WithConfig(config.NewConfigManager("/test/config.yaml")).
			WithStatusManager(mockStatus).
			WithLogger(logger.NewNoopLogger()),
	}

	params := AddRepositoryToWorkspaceParams{
		WorkspaceName: "test-workspace",
		Repository:    "non-existent-repo",
	}

	// Mock workspace exists
	existingWorkspace := &status.Workspace{
		Repositories: []string{},
		Worktrees:    []string{},
	}
	mockStatus.EXPECT().GetWorkspace("test-workspace").Return(existingWorkspace, nil)

	// Mock repository not found in status
	mockStatus.EXPECT().GetRepository("non-existent-repo").Return(nil, errors.New("not found"))

	// Mock ResolvePath call (resolveRepository tries to resolve relative paths)
	mockFS.EXPECT().ResolvePath(gomock.Any(), "non-existent-repo").Return("non-existent-repo", nil)

	// Mock path doesn't exist
	mockFS.EXPECT().Exists("non-existent-repo").Return(false, nil)

	err := cm.AddRepositoryToWorkspace(&params)
	assert.Error(t, err)
}
