//go:build unit

package repository

import (
	"errors"
	"testing"

	"github.com/lerenn/code-manager/pkg/config"
	fsmocks "github.com/lerenn/code-manager/pkg/fs/mocks"
	gitmocks "github.com/lerenn/code-manager/pkg/git/mocks"
	"github.com/lerenn/code-manager/pkg/logger"
	promptmocks "github.com/lerenn/code-manager/pkg/prompt/mocks"
	statusmocks "github.com/lerenn/code-manager/pkg/status/mocks"
	"github.com/lerenn/code-manager/pkg/worktree"
	worktreemocks "github.com/lerenn/code-manager/pkg/worktree/mocks"
	"github.com/stretchr/testify/assert"
	"go.uber.org/mock/gomock"
)

func TestIsGitRepository_GitDirectory(t *testing.T) {
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
		configManager:    config.NewManager("/test/config.yaml"),
		statusManager:    mockStatus,
		logger:           logger.NewNoopLogger(),
		prompt:           mockPrompt,
		worktreeProvider: func(_ worktree.NewWorktreeParams) worktree.Worktree { return mockWorktree },
		repositoryPath:   "/test/repo",
	}

	// Mock .git directory exists and is a directory
	mockFS.EXPECT().Exists("/test/repo/.git").Return(true, nil)
	mockFS.EXPECT().IsDir("/test/repo/.git").Return(true, nil)

	result, err := repository.IsGitRepository()
	assert.NoError(t, err)
	assert.True(t, result)
}

func TestIsGitRepository_GitWorktreeFile(t *testing.T) {
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
		configManager:    config.NewManager("/test/config.yaml"),
		statusManager:    mockStatus,
		logger:           logger.NewNoopLogger(),
		prompt:           mockPrompt,
		worktreeProvider: func(_ worktree.NewWorktreeParams) worktree.Worktree { return mockWorktree },
		repositoryPath:   "/test/repo",
	}

	// Mock .git file exists and is not a directory
	mockFS.EXPECT().Exists("/test/repo/.git").Return(true, nil)
	mockFS.EXPECT().IsDir("/test/repo/.git").Return(false, nil)
	mockFS.EXPECT().ReadFile("/test/repo/.git").Return([]byte("gitdir: /path/to/main/repo/.git/worktrees/branch"), nil)

	result, err := repository.IsGitRepository()
	assert.NoError(t, err)
	assert.True(t, result)
}

func TestIsGitRepository_NoGitDirectory(t *testing.T) {
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
		configManager:    config.NewManager("/test/config.yaml"),
		statusManager:    mockStatus,
		logger:           logger.NewNoopLogger(),
		prompt:           mockPrompt,
		worktreeProvider: func(_ worktree.NewWorktreeParams) worktree.Worktree { return mockWorktree },
		repositoryPath:   "/test/repo",
	}

	// Mock .git does not exist
	mockFS.EXPECT().Exists("/test/repo/.git").Return(false, nil)

	result, err := repository.IsGitRepository()
	assert.NoError(t, err)
	assert.False(t, result)
}

func TestIsGitRepository_InvalidWorktreeFile(t *testing.T) {
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
		configManager:    config.NewManager("/test/config.yaml"),
		statusManager:    mockStatus,
		logger:           logger.NewNoopLogger(),
		prompt:           mockPrompt,
		worktreeProvider: func(_ worktree.NewWorktreeParams) worktree.Worktree { return mockWorktree },
		repositoryPath:   "/test/repo",
	}

	// Mock .git file exists but is not a valid worktree file
	mockFS.EXPECT().Exists("/test/repo/.git").Return(true, nil)
	mockFS.EXPECT().IsDir("/test/repo/.git").Return(false, nil)
	mockFS.EXPECT().ReadFile("/test/repo/.git").Return([]byte("invalid content"), nil)

	result, err := repository.IsGitRepository()
	assert.NoError(t, err)
	assert.False(t, result)
}

func TestIsGitRepository_FileSystemError(t *testing.T) {
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
		configManager:    config.NewManager("/test/config.yaml"),
		statusManager:    mockStatus,
		logger:           logger.NewNoopLogger(),
		prompt:           mockPrompt,
		worktreeProvider: func(_ worktree.NewWorktreeParams) worktree.Worktree { return mockWorktree },
		repositoryPath:   "/test/repo",
	}

	// Mock filesystem error
	mockFS.EXPECT().Exists("/test/repo/.git").Return(false, errors.New("filesystem error"))

	result, err := repository.IsGitRepository()
	assert.Error(t, err)
	assert.False(t, result)
	assert.Contains(t, err.Error(), "failed to check .git existence")
}

// Additional tests moved from repository_test.go

func TestIsGitRepository_Directory_FromRepository(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockFS := fsmocks.NewMockFS(ctrl)
	mockGit := gitmocks.NewMockGit(ctrl)
	mockStatus := statusmocks.NewMockManager(ctrl)
	mockPrompt := promptmocks.NewMockPrompter(ctrl)

	repo := NewRepository(NewRepositoryParams{
		FS:            mockFS,
		Git:           mockGit,
		ConfigManager: config.NewManager("/test/config.yaml"),
		StatusManager: mockStatus,
		Prompt:        mockPrompt,
	})

	// Mock .git exists and is a directory (regular repository)
	mockFS.EXPECT().Exists(".git").Return(true, nil)
	mockFS.EXPECT().IsDir(".git").Return(true, nil)

	exists, err := repo.IsGitRepository()
	assert.NoError(t, err)
	assert.True(t, exists) // Should return true for regular repositories
}

func TestIsGitRepository_File_FromRepository(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockFS := fsmocks.NewMockFS(ctrl)
	mockGit := gitmocks.NewMockGit(ctrl)
	mockStatus := statusmocks.NewMockManager(ctrl)
	mockPrompt := promptmocks.NewMockPrompter(ctrl)

	repo := NewRepository(NewRepositoryParams{
		FS:            mockFS,
		Git:           mockGit,
		ConfigManager: config.NewManager("/test/config.yaml"),
		StatusManager: mockStatus,
		Prompt:        mockPrompt,
	})

	// Mock .git exists but is not a directory (worktree case)
	mockFS.EXPECT().Exists(".git").Return(true, nil)
	mockFS.EXPECT().IsDir(".git").Return(false, nil)
	// Mock .git file content with valid worktree format
	mockFS.EXPECT().ReadFile(".git").Return([]byte("gitdir: /path/to/main/repo/.git/worktrees/worktree-name"), nil)

	exists, err := repo.IsGitRepository()
	assert.NoError(t, err)
	assert.True(t, exists) // Should return true for valid worktrees
}

func TestIsGitRepository_InvalidFile_FromRepository(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockFS := fsmocks.NewMockFS(ctrl)
	mockGit := gitmocks.NewMockGit(ctrl)
	mockStatus := statusmocks.NewMockManager(ctrl)
	mockPrompt := promptmocks.NewMockPrompter(ctrl)

	repo := NewRepository(NewRepositoryParams{
		FS:            mockFS,
		Git:           mockGit,
		ConfigManager: config.NewManager("/test/config.yaml"),
		StatusManager: mockStatus,
		Prompt:        mockPrompt,
	})

	// Mock .git exists but is not a directory
	mockFS.EXPECT().Exists(".git").Return(true, nil)
	mockFS.EXPECT().IsDir(".git").Return(false, nil)
	// Mock .git file content without valid worktree format
	mockFS.EXPECT().ReadFile(".git").Return([]byte("not a git worktree file"), nil)

	exists, err := repo.IsGitRepository()
	assert.NoError(t, err)
	assert.False(t, exists) // Should return false for invalid .git files
}

func TestIsGitRepository_NotExists_FromRepository(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockFS := fsmocks.NewMockFS(ctrl)
	mockGit := gitmocks.NewMockGit(ctrl)
	mockStatus := statusmocks.NewMockManager(ctrl)
	mockPrompt := promptmocks.NewMockPrompter(ctrl)

	repo := NewRepository(NewRepositoryParams{
		FS:            mockFS,
		Git:           mockGit,
		ConfigManager: config.NewManager("/test/config.yaml"),
		StatusManager: mockStatus,
		Prompt:        mockPrompt,
	})

	// Mock .git does not exist
	mockFS.EXPECT().Exists(".git").Return(false, nil)

	exists, err := repo.IsGitRepository()
	assert.NoError(t, err)
	assert.False(t, exists) // Should return false when .git doesn't exist
}
