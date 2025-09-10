//go:build unit

package repository

import (
	"fmt"
	"testing"

	"github.com/lerenn/code-manager/pkg/config"
	fsmocks "github.com/lerenn/code-manager/pkg/fs/mocks"
	gitmocks "github.com/lerenn/code-manager/pkg/git/mocks"
	promptmocks "github.com/lerenn/code-manager/pkg/prompt/mocks"
	"github.com/lerenn/code-manager/pkg/status"
	statusmocks "github.com/lerenn/code-manager/pkg/status/mocks"
	"github.com/stretchr/testify/assert"
	"go.uber.org/mock/gomock"
)

// createTestConfig creates a test configuration.
func createTestConfig() config.Config {
	return config.Config{
		BasePath:   "/test/base/path",
		StatusFile: "/test/status.yaml",
	}
}

func TestRepository_Validate_Success(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockFS := fsmocks.NewMockFS(ctrl)
	mockGit := gitmocks.NewMockGit(ctrl)
	mockStatus := statusmocks.NewMockManager(ctrl)

	mockPrompt := promptmocks.NewMockPrompter(ctrl)

	repo := NewRepository(NewRepositoryParams{
		FS:            mockFS,
		Git:           mockGit,
		Config:        createTestConfig(),
		StatusManager: mockStatus,
		Prompt:        mockPrompt,
	})

	// Mock repository validation
	mockFS.EXPECT().Exists(".git").Return(true, nil)
	mockFS.EXPECT().IsDir(".git").Return(true, nil)
	mockGit.EXPECT().Status(".").Return("On branch main", nil)
	mockGit.EXPECT().GetCurrentBranch(".").Return("main", nil)

	err := repo.Validate()
	assert.NoError(t, err)
}

func TestRepository_Validate_NoGitDir(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockFS := fsmocks.NewMockFS(ctrl)
	mockGit := gitmocks.NewMockGit(ctrl)
	mockStatus := statusmocks.NewMockManager(ctrl)

	mockPrompt := promptmocks.NewMockPrompter(ctrl)

	repo := NewRepository(NewRepositoryParams{
		FS:            mockFS,
		Git:           mockGit,
		Config:        createTestConfig(),
		StatusManager: mockStatus,

		Prompt: mockPrompt,
	})

	// Mock repository validation - .git not found
	mockFS.EXPECT().Exists(".git").Return(false, nil)

	err := repo.Validate()
	assert.ErrorIs(t, err, ErrGitRepositoryNotFound)
}

func TestRepository_Validate_GitStatusError(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockFS := fsmocks.NewMockFS(ctrl)
	mockGit := gitmocks.NewMockGit(ctrl)
	mockStatus := statusmocks.NewMockManager(ctrl)

	mockPrompt := promptmocks.NewMockPrompter(ctrl)

	repo := NewRepository(NewRepositoryParams{
		FS:            mockFS,
		Git:           mockGit,
		Config:        createTestConfig(),
		StatusManager: mockStatus,

		Prompt: mockPrompt,
	})

	// Mock repository validation - .git exists but git status fails
	mockFS.EXPECT().Exists(".git").Return(true, nil)
	mockFS.EXPECT().IsDir(".git").Return(true, nil)
	mockGit.EXPECT().Status(".").Return("", fmt.Errorf("git error"))

	err := repo.Validate()
	assert.Error(t, err)
	assert.ErrorIs(t, err, ErrGitRepositoryInvalid)
}

func TestRepository_Validate_NotClean(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockFS := fsmocks.NewMockFS(ctrl)
	mockGit := gitmocks.NewMockGit(ctrl)
	mockStatus := statusmocks.NewMockManager(ctrl)

	mockPrompt := promptmocks.NewMockPrompter(ctrl)

	repo := NewRepository(NewRepositoryParams{
		FS:            mockFS,
		Git:           mockGit,
		Config:        createTestConfig(),
		StatusManager: mockStatus,

		Prompt: mockPrompt,
	})

	// Mock repository validation - .git exists but repository is not clean
	mockFS.EXPECT().Exists(".git").Return(true, nil)
	mockFS.EXPECT().IsDir(".git").Return(true, nil)
	mockGit.EXPECT().Status(".").Return("On branch main\nChanges not staged for commit", nil)
	mockGit.EXPECT().GetCurrentBranch(".").Return("main", nil)

	err := repo.Validate()
	assert.NoError(t, err)
}

func TestRepository_ValidateRepository_Success(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockFS := fsmocks.NewMockFS(ctrl)
	mockGit := gitmocks.NewMockGit(ctrl)
	mockStatus := statusmocks.NewMockManager(ctrl)

	mockPrompt := promptmocks.NewMockPrompter(ctrl)

	repo := NewRepository(NewRepositoryParams{
		FS:            mockFS,
		Git:           mockGit,
		Config:        createTestConfig(),
		StatusManager: mockStatus,

		Prompt: mockPrompt,
	})

	// Mock repository validation
	mockFS.EXPECT().Exists(".git").Return(true, nil)
	mockFS.EXPECT().IsDir(".git").Return(true, nil)
	mockGit.EXPECT().GetRepositoryName(gomock.Any()).Return("github.com/octocat/Hello-World", nil)
	mockStatus.EXPECT().GetWorktree("github.com/octocat/Hello-World", "test-branch").Return(nil, status.ErrWorktreeNotFound)
	mockGit.EXPECT().IsClean(gomock.Any()).Return(true, nil)

	result, err := repo.ValidateRepository(ValidationParams{Branch: "test-branch"})
	assert.NoError(t, err)
	assert.Equal(t, "github.com/octocat/Hello-World", result.RepoURL)
}

func TestRepository_ValidateRepository_WorktreeExists(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockFS := fsmocks.NewMockFS(ctrl)
	mockGit := gitmocks.NewMockGit(ctrl)
	mockStatus := statusmocks.NewMockManager(ctrl)

	mockPrompt := promptmocks.NewMockPrompter(ctrl)

	repo := NewRepository(NewRepositoryParams{
		FS:            mockFS,
		Git:           mockGit,
		Config:        createTestConfig(),
		StatusManager: mockStatus,

		Prompt: mockPrompt,
	})

	// Mock repository validation
	mockFS.EXPECT().Exists(".git").Return(true, nil)
	mockFS.EXPECT().IsDir(".git").Return(true, nil)
	mockGit.EXPECT().GetRepositoryName(gomock.Any()).Return("github.com/octocat/Hello-World", nil)
	mockStatus.EXPECT().GetWorktree("github.com/octocat/Hello-World", "test-branch").Return(&status.WorktreeInfo{}, nil)

	result, err := repo.ValidateRepository(ValidationParams{Branch: "test-branch"})
	assert.Error(t, err)
	assert.ErrorIs(t, err, ErrWorktreeExists)
	assert.Nil(t, result)
}

func TestRepository_ValidateRepository_NotClean(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockFS := fsmocks.NewMockFS(ctrl)
	mockGit := gitmocks.NewMockGit(ctrl)
	mockStatus := statusmocks.NewMockManager(ctrl)

	mockPrompt := promptmocks.NewMockPrompter(ctrl)

	repo := NewRepository(NewRepositoryParams{
		FS:            mockFS,
		Git:           mockGit,
		Config:        createTestConfig(),
		StatusManager: mockStatus,

		Prompt: mockPrompt,
	})

	// Mock repository validation
	mockFS.EXPECT().Exists(".git").Return(true, nil)
	mockFS.EXPECT().IsDir(".git").Return(true, nil)
	mockGit.EXPECT().GetRepositoryName(gomock.Any()).Return("github.com/octocat/Hello-World", nil)
	mockStatus.EXPECT().GetWorktree("github.com/octocat/Hello-World", "test-branch").Return(nil, status.ErrWorktreeNotFound)
	mockGit.EXPECT().IsClean(gomock.Any()).Return(false, nil)

	result, err := repo.ValidateRepository(ValidationParams{Branch: "test-branch"})
	assert.Error(t, err)
	assert.ErrorIs(t, err, ErrRepositoryNotClean)
	assert.Nil(t, result)
}

func TestRepository_ValidateWorktreeExists_Success(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockFS := fsmocks.NewMockFS(ctrl)
	mockGit := gitmocks.NewMockGit(ctrl)
	mockStatus := statusmocks.NewMockManager(ctrl)

	mockPrompt := promptmocks.NewMockPrompter(ctrl)

	repo := NewRepository(NewRepositoryParams{
		FS:            mockFS,
		Git:           mockGit,
		Config:        createTestConfig(),
		StatusManager: mockStatus,

		Prompt: mockPrompt,
	})

	mockStatus.EXPECT().GetWorktree("github.com/octocat/Hello-World", "test-branch").Return(&status.WorktreeInfo{}, nil)

	err := repo.ValidateWorktreeExists("github.com/octocat/Hello-World", "test-branch")
	assert.NoError(t, err)
}

func TestRepository_ValidateWorktreeExists_NotFound(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockFS := fsmocks.NewMockFS(ctrl)
	mockGit := gitmocks.NewMockGit(ctrl)
	mockStatus := statusmocks.NewMockManager(ctrl)

	mockPrompt := promptmocks.NewMockPrompter(ctrl)

	repo := NewRepository(NewRepositoryParams{
		FS:            mockFS,
		Git:           mockGit,
		Config:        createTestConfig(),
		StatusManager: mockStatus,

		Prompt: mockPrompt,
	})

	mockStatus.EXPECT().GetWorktree("github.com/octocat/Hello-World", "test-branch").Return(nil, status.ErrWorktreeNotFound)

	err := repo.ValidateWorktreeExists("github.com/octocat/Hello-World", "test-branch")
	assert.Error(t, err)
	assert.ErrorIs(t, err, ErrWorktreeNotInStatus)
}

func TestRepository_ValidateGitStatus_Success(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockFS := fsmocks.NewMockFS(ctrl)
	mockGit := gitmocks.NewMockGit(ctrl)
	mockStatus := statusmocks.NewMockManager(ctrl)

	mockPrompt := promptmocks.NewMockPrompter(ctrl)

	repo := NewRepository(NewRepositoryParams{
		FS:            mockFS,
		Git:           mockGit,
		Config:        createTestConfig(),
		StatusManager: mockStatus,

		Prompt: mockPrompt,
	})

	mockGit.EXPECT().Status(".").Return("On branch main", nil)

	err := repo.ValidateGitStatus()
	assert.NoError(t, err)
}

func TestRepository_ValidateGitStatus_Error(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockFS := fsmocks.NewMockFS(ctrl)
	mockGit := gitmocks.NewMockGit(ctrl)
	mockStatus := statusmocks.NewMockManager(ctrl)

	mockPrompt := promptmocks.NewMockPrompter(ctrl)

	repo := NewRepository(NewRepositoryParams{
		FS:            mockFS,
		Git:           mockGit,
		Config:        createTestConfig(),
		StatusManager: mockStatus,

		Prompt: mockPrompt,
	})

	mockGit.EXPECT().Status(".").Return("", fmt.Errorf("git error"))

	err := repo.ValidateGitStatus()
	assert.Error(t, err)
	assert.ErrorIs(t, err, ErrGitRepositoryInvalid)
}

func TestRepository_ValidateOriginRemote_Success(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockFS := fsmocks.NewMockFS(ctrl)
	mockGit := gitmocks.NewMockGit(ctrl)
	mockStatus := statusmocks.NewMockManager(ctrl)

	mockPrompt := promptmocks.NewMockPrompter(ctrl)

	repo := NewRepository(NewRepositoryParams{
		FS:            mockFS,
		Git:           mockGit,
		Config:        createTestConfig(),
		StatusManager: mockStatus,

		Prompt: mockPrompt,
	})

	mockGit.EXPECT().RemoteExists(".", "origin").Return(true, nil)
	mockGit.EXPECT().GetRemoteURL(".", "origin").Return("https://github.com/octocat/Hello-World.git", nil)

	err := repo.ValidateOriginRemote()
	assert.NoError(t, err)
}

func TestRepository_ValidateOriginRemote_NotFound(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockFS := fsmocks.NewMockFS(ctrl)
	mockGit := gitmocks.NewMockGit(ctrl)
	mockStatus := statusmocks.NewMockManager(ctrl)

	mockPrompt := promptmocks.NewMockPrompter(ctrl)

	repo := NewRepository(NewRepositoryParams{
		FS:            mockFS,
		Git:           mockGit,
		Config:        createTestConfig(),
		StatusManager: mockStatus,

		Prompt: mockPrompt,
	})

	mockGit.EXPECT().RemoteExists(".", "origin").Return(false, nil)

	err := repo.ValidateOriginRemote()
	assert.Error(t, err)
	assert.ErrorIs(t, err, ErrOriginRemoteNotFound)
}

func TestRepository_ValidateOriginRemote_InvalidURL(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockFS := fsmocks.NewMockFS(ctrl)
	mockGit := gitmocks.NewMockGit(ctrl)
	mockStatus := statusmocks.NewMockManager(ctrl)

	mockPrompt := promptmocks.NewMockPrompter(ctrl)

	repo := NewRepository(NewRepositoryParams{
		FS:            mockFS,
		Git:           mockGit,
		Config:        createTestConfig(),
		StatusManager: mockStatus,

		Prompt: mockPrompt,
	})

	mockGit.EXPECT().RemoteExists(".", "origin").Return(true, nil)
	mockGit.EXPECT().GetRemoteURL(".", "origin").Return("invalid-url", nil)

	err := repo.ValidateOriginRemote()
	assert.Error(t, err)
	assert.ErrorIs(t, err, ErrOriginRemoteInvalidURL)
}
