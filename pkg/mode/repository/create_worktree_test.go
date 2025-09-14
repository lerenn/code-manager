//go:build unit

package repository

import (
	"errors"
	"fmt"
	"testing"

	"github.com/lerenn/code-manager/pkg/config"
	fsmocks "github.com/lerenn/code-manager/pkg/fs/mocks"
	gitmocks "github.com/lerenn/code-manager/pkg/git/mocks"
	"github.com/lerenn/code-manager/pkg/issue"
	"github.com/lerenn/code-manager/pkg/logger"
	promptmocks "github.com/lerenn/code-manager/pkg/prompt/mocks"
	"github.com/lerenn/code-manager/pkg/status"
	statusmocks "github.com/lerenn/code-manager/pkg/status/mocks"
	"github.com/lerenn/code-manager/pkg/worktree"
	worktreemocks "github.com/lerenn/code-manager/pkg/worktree/mocks"
	"github.com/stretchr/testify/assert"
	"go.uber.org/mock/gomock"
)

func TestCreateWorktree_Success(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockFS := fsmocks.NewMockFS(ctrl)
	mockGit := gitmocks.NewMockGit(ctrl)
	mockStatus := statusmocks.NewMockManager(ctrl)
	mockPrompt := promptmocks.NewMockPrompter(ctrl)
	mockWorktree := worktreemocks.NewMockWorktree(ctrl)

	repository := &realRepository{
		fs:               mockFS,
		git:              mockGit,
		config:           config.Config{RepositoriesDir: "/test/repos"},
		statusManager:    mockStatus,
		logger:           logger.NewNoopLogger(),
		prompt:           mockPrompt,
		worktreeProvider: func(_ worktree.NewWorktreeParams) worktree.Worktree { return mockWorktree },
		repositoryPath:   "/test/repo",
	}

	// Mock ValidateRepository
	mockFS.EXPECT().Exists("/test/repo/.git").Return(true, nil)
	mockFS.EXPECT().IsDir("/test/repo/.git").Return(true, nil)
	mockGit.EXPECT().GetRepositoryName("/test/repo").Return("github.com/test/repo", nil)
	mockStatus.EXPECT().GetWorktree("github.com/test/repo", "test-branch").Return(nil, status.ErrWorktreeNotFound)
	mockGit.EXPECT().IsClean("/test/repo").Return(true, nil)
	// Mock origin URL validation - using placeholder for now

	// Mock worktree creation
	mockWorktree.EXPECT().BuildPath("github.com/test/repo", "origin", "test-branch").Return("/test/repos/github.com/test/repo/worktrees/origin/test-branch")
	mockWorktree.EXPECT().ValidateCreation(gomock.Any()).Return(nil)
	mockWorktree.EXPECT().Create(gomock.Any()).Return(nil)
	mockWorktree.EXPECT().CheckoutBranch("/test/repos/github.com/test/repo/worktrees/origin/test-branch", "test-branch").Return(nil)
	mockGit.EXPECT().SetUpstreamBranch("/test/repos/github.com/test/repo/worktrees/origin/test-branch", "origin", "test-branch").Return(nil)

	// Mock status management
	mockWorktree.EXPECT().AddToStatus(gomock.Any()).Return(nil)

	result, err := repository.CreateWorktree("test-branch")
	assert.NoError(t, err)
	assert.Equal(t, "/test/repos/github.com/test/repo/worktrees/origin/test-branch", result)
}

func TestCreateWorktree_ValidationError(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockFS := fsmocks.NewMockFS(ctrl)
	mockGit := gitmocks.NewMockGit(ctrl)
	mockStatus := statusmocks.NewMockManager(ctrl)
	mockPrompt := promptmocks.NewMockPrompter(ctrl)
	mockWorktree := worktreemocks.NewMockWorktree(ctrl)

	repository := &realRepository{
		fs:               mockFS,
		git:              mockGit,
		config:           config.Config{RepositoriesDir: "/test/repos"},
		statusManager:    mockStatus,
		logger:           logger.NewNoopLogger(),
		prompt:           mockPrompt,
		worktreeProvider: func(_ worktree.NewWorktreeParams) worktree.Worktree { return mockWorktree },
		repositoryPath:   "/test/repo",
	}

	// Mock ValidateRepository to return error
	mockFS.EXPECT().Exists("/test/repo/.git").Return(false, nil)

	result, err := repository.CreateWorktree("test-branch")
	assert.Error(t, err)
	assert.Empty(t, result)
	assert.Contains(t, err.Error(), "current directory is not a Git repository")
}

func TestCreateWorktree_WorktreeCreationError(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockFS := fsmocks.NewMockFS(ctrl)
	mockGit := gitmocks.NewMockGit(ctrl)
	mockStatus := statusmocks.NewMockManager(ctrl)
	mockPrompt := promptmocks.NewMockPrompter(ctrl)
	mockWorktree := worktreemocks.NewMockWorktree(ctrl)

	repository := &realRepository{
		fs:               mockFS,
		git:              mockGit,
		config:           config.Config{RepositoriesDir: "/test/repos"},
		statusManager:    mockStatus,
		logger:           logger.NewNoopLogger(),
		prompt:           mockPrompt,
		worktreeProvider: func(_ worktree.NewWorktreeParams) worktree.Worktree { return mockWorktree },
		repositoryPath:   "/test/repo",
	}

	// Mock ValidateRepository
	mockFS.EXPECT().Exists("/test/repo/.git").Return(true, nil)
	mockFS.EXPECT().IsDir("/test/repo/.git").Return(true, nil)
	mockGit.EXPECT().GetRepositoryName("/test/repo").Return("github.com/test/repo", nil)
	mockStatus.EXPECT().GetWorktree("github.com/test/repo", "test-branch").Return(nil, status.ErrWorktreeNotFound)
	mockGit.EXPECT().IsClean("/test/repo").Return(true, nil)
	// Mock origin URL validation - using placeholder for now

	// Mock worktree creation to fail
	mockWorktree.EXPECT().BuildPath("github.com/test/repo", "origin", "test-branch").Return("/test/repos/github.com/test/repo/worktrees/origin/test-branch")
	mockWorktree.EXPECT().ValidateCreation(gomock.Any()).Return(errors.New("validation failed"))

	result, err := repository.CreateWorktree("test-branch")
	assert.Error(t, err)
	assert.Empty(t, result)
	assert.Contains(t, err.Error(), "validation failed")
}

func TestExtractRemote_WithOptions(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	repository := &realRepository{}

	opts := []CreateWorktreeOpts{
		{Remote: "upstream"},
	}

	result := repository.extractRemote(opts)
	assert.Equal(t, "upstream", result)
}

func TestExtractRemote_WithoutOptions(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	repository := &realRepository{}

	result := repository.extractRemote([]CreateWorktreeOpts{})
	assert.Equal(t, DefaultRemote, result)
}

func TestExtractIssueInfo_WithIssueInfo(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	repository := &realRepository{}

	issueInfo := &issue.Info{Number: 123}
	opts := []CreateWorktreeOpts{
		{IssueInfo: issueInfo},
	}

	result := repository.extractIssueInfo(opts)
	assert.Equal(t, issueInfo, result)
}

func TestExtractIssueInfo_WithoutIssueInfo(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	repository := &realRepository{}

	result := repository.extractIssueInfo([]CreateWorktreeOpts{})
	assert.Nil(t, result)
}

// Additional tests moved from repository_test.go

func TestCreateWorktree_Success_FromRepository(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockFS := fsmocks.NewMockFS(ctrl)
	mockGit := gitmocks.NewMockGit(ctrl)
	mockStatus := statusmocks.NewMockManager(ctrl)
	mockPrompt := promptmocks.NewMockPrompter(ctrl)
	mockWorktree := worktreemocks.NewMockWorktree(ctrl)

	repo := NewRepository(NewRepositoryParams{
		FS:            mockFS,
		Git:           mockGit,
		Config:        createTestConfig(),
		StatusManager: mockStatus,
		Prompt:        mockPrompt,
		WorktreeProvider: func(params worktree.NewWorktreeParams) worktree.Worktree {
			return mockWorktree
		},
	})

	// Mock repository validation
	mockFS.EXPECT().Exists(".git").Return(true, nil).AnyTimes()
	mockFS.EXPECT().IsDir(".git").Return(true, nil).AnyTimes()
	mockGit.EXPECT().Status(".").Return("On branch main", nil).AnyTimes()

	// Mock worktree creation
	mockGit.EXPECT().GetRepositoryName(gomock.Any()).Return("github.com/octocat/Hello-World", nil)
	mockStatus.EXPECT().GetWorktree("github.com/octocat/Hello-World", "test-branch").Return(nil, status.ErrWorktreeNotFound)
	mockGit.EXPECT().IsClean(gomock.Any()).Return(true, nil)
	mockWorktree.EXPECT().BuildPath("github.com/octocat/Hello-World", "origin", "test-branch").Return("/test/path/github.com/octocat/Hello-World/origin/test-branch")
	mockWorktree.EXPECT().ValidateCreation(gomock.Any()).Return(nil)
	mockWorktree.EXPECT().Create(gomock.Any()).Return(nil)
	mockWorktree.EXPECT().CheckoutBranch("/test/path/github.com/octocat/Hello-World/origin/test-branch", "test-branch").Return(nil)
	mockGit.EXPECT().SetUpstreamBranch("/test/path/github.com/octocat/Hello-World/origin/test-branch", "origin", "test-branch").Return(nil)
	mockWorktree.EXPECT().AddToStatus(gomock.Any()).Return(nil)

	worktreePath, err := repo.CreateWorktree("test-branch")
	assert.NoError(t, err)
	assert.NotEmpty(t, worktreePath)
}

func TestCreateWorktree_SetUpstreamBranch_Success(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockFS := fsmocks.NewMockFS(ctrl)
	mockGit := gitmocks.NewMockGit(ctrl)
	mockStatus := statusmocks.NewMockManager(ctrl)
	mockPrompt := promptmocks.NewMockPrompter(ctrl)
	mockWorktree := worktreemocks.NewMockWorktree(ctrl)

	repo := NewRepository(NewRepositoryParams{
		FS:            mockFS,
		Git:           mockGit,
		Config:        createTestConfig(),
		StatusManager: mockStatus,
		Prompt:        mockPrompt,
		WorktreeProvider: func(params worktree.NewWorktreeParams) worktree.Worktree {
			return mockWorktree
		},
	})

	// Mock repository validation
	mockFS.EXPECT().Exists(".git").Return(true, nil).AnyTimes()
	mockFS.EXPECT().IsDir(".git").Return(true, nil).AnyTimes()
	mockGit.EXPECT().Status(".").Return("On branch main", nil).AnyTimes()

	// Mock worktree creation
	mockGit.EXPECT().GetRepositoryName(gomock.Any()).Return("github.com/octocat/Hello-World", nil)
	mockStatus.EXPECT().GetWorktree("github.com/octocat/Hello-World", "test-branch").Return(nil, status.ErrWorktreeNotFound)
	mockGit.EXPECT().IsClean(gomock.Any()).Return(true, nil)
	mockWorktree.EXPECT().BuildPath("github.com/octocat/Hello-World", "origin", "test-branch").Return("/test/path/github.com/octocat/Hello-World/origin/test-branch")
	mockWorktree.EXPECT().ValidateCreation(gomock.Any()).Return(nil)
	mockWorktree.EXPECT().Create(gomock.Any()).Return(nil)
	mockWorktree.EXPECT().CheckoutBranch("/test/path/github.com/octocat/Hello-World/origin/test-branch", "test-branch").Return(nil)

	// Mock successful upstream setup
	mockGit.EXPECT().
		SetUpstreamBranch("/test/path/github.com/octocat/Hello-World/origin/test-branch", "origin", "test-branch").
		Return(nil)

	mockWorktree.EXPECT().AddToStatus(gomock.Any()).Return(nil)

	// Call the method
	worktreePath, err := repo.CreateWorktree("test-branch")

	// Verify no error
	assert.NoError(t, err)
	assert.NotEmpty(t, worktreePath)
}

func TestCreateWorktree_SetUpstreamBranch_Failure(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockFS := fsmocks.NewMockFS(ctrl)
	mockGit := gitmocks.NewMockGit(ctrl)
	mockStatus := statusmocks.NewMockManager(ctrl)
	mockPrompt := promptmocks.NewMockPrompter(ctrl)
	mockWorktree := worktreemocks.NewMockWorktree(ctrl)

	repo := NewRepository(NewRepositoryParams{
		FS:            mockFS,
		Git:           mockGit,
		Config:        createTestConfig(),
		StatusManager: mockStatus,
		Prompt:        mockPrompt,
		WorktreeProvider: func(params worktree.NewWorktreeParams) worktree.Worktree {
			return mockWorktree
		},
	})

	// Mock repository validation
	mockFS.EXPECT().Exists(".git").Return(true, nil).AnyTimes()
	mockFS.EXPECT().IsDir(".git").Return(true, nil).AnyTimes()
	mockGit.EXPECT().Status(".").Return("On branch main", nil).AnyTimes()

	// Mock worktree creation
	mockGit.EXPECT().GetRepositoryName(gomock.Any()).Return("github.com/octocat/Hello-World", nil)
	mockStatus.EXPECT().GetWorktree("github.com/octocat/Hello-World", "test-branch").Return(nil, status.ErrWorktreeNotFound)
	mockGit.EXPECT().IsClean(gomock.Any()).Return(true, nil)
	mockWorktree.EXPECT().BuildPath("github.com/octocat/Hello-World", "origin", "test-branch").Return("/test/path/github.com/octocat/Hello-World/origin/test-branch")
	mockWorktree.EXPECT().ValidateCreation(gomock.Any()).Return(nil)
	mockWorktree.EXPECT().Create(gomock.Any()).Return(nil)
	mockWorktree.EXPECT().CheckoutBranch("/test/path/github.com/octocat/Hello-World/origin/test-branch", "test-branch").Return(nil)

	// Mock failed upstream setup
	expectedError := "git branch --set-upstream-to failed: exit status 128"
	mockGit.EXPECT().
		SetUpstreamBranch("/test/path/github.com/octocat/Hello-World/origin/test-branch", "origin", "test-branch").
		Return(fmt.Errorf("%s", expectedError))

	// Mock cleanup on failure
	mockWorktree.EXPECT().CleanupDirectory("/test/path/github.com/octocat/Hello-World/origin/test-branch").Return(nil)

	// Call the method
	worktreePath, err := repo.CreateWorktree("test-branch")

	// Verify error is returned
	assert.Error(t, err)
	assert.Empty(t, worktreePath)
	assert.Contains(t, err.Error(), "failed to set upstream branch tracking")
	assert.Contains(t, err.Error(), expectedError)
}
