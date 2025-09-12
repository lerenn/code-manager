//go:build unit

package repository

import (
	"fmt"
	"testing"

	fsmocks "github.com/lerenn/code-manager/pkg/fs/mocks"
	gitmocks "github.com/lerenn/code-manager/pkg/git/mocks"
	promptMocks "github.com/lerenn/code-manager/pkg/prompt/mocks"
	"github.com/lerenn/code-manager/pkg/status"
	statusMocks "github.com/lerenn/code-manager/pkg/status/mocks"
	"github.com/lerenn/code-manager/pkg/worktree"
	worktreeMocks "github.com/lerenn/code-manager/pkg/worktree/mocks"
	"github.com/stretchr/testify/assert"
	"go.uber.org/mock/gomock"
)

func TestRealRepository_DeleteAllWorktrees_Success(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockFS := fsmocks.NewMockFS(ctrl)
	mockGit := gitmocks.NewMockGit(ctrl)
	mockStatus := statusMocks.NewMockManager(ctrl)
	// logger will be nil and use NewNoopLogger in real usage
	mockPrompt := promptMocks.NewMockPrompter(ctrl)
	mockWorktree := worktreeMocks.NewMockWorktree(ctrl)

	// Create repository instance
	repo := NewRepository(NewRepositoryParams{
		FS:               mockFS,
		Git:              mockGit,
		Config:           createTestConfig(),
		StatusManager:    mockStatus,
		Prompt:           mockPrompt,
		WorktreeProvider: func(params worktree.NewWorktreeParams) worktree.Worktree { return mockWorktree },
	})

	// Mock validation - IsGitRepository calls
	mockFS.EXPECT().Exists(".git").Return(true, nil)
	mockFS.EXPECT().IsDir(".git").Return(true, nil)

	// Mock validation - getRepositoryURL calls (called twice: once in validation, once in ListWorktrees)
	mockGit.EXPECT().GetRepositoryName(gomock.Any()).Return("test-repo", nil).Times(2)

	// Mock worktree listing
	mockStatus.EXPECT().GetRepository("test-repo").Return(&status.Repository{
		Worktrees: map[string]status.WorktreeInfo{
			"feature/branch1": {Branch: "feature/branch1", Remote: "origin"},
			"feature/branch2": {Branch: "feature/branch2", Remote: "origin"},
		},
	}, nil)

	// Mock worktree path retrieval
	mockGit.EXPECT().GetWorktreePath(gomock.Any(), "feature/branch1").Return("/test/repos/worktrees/test-repo/feature/branch1", nil)
	mockGit.EXPECT().GetWorktreePath(gomock.Any(), "feature/branch2").Return("/test/repos/worktrees/test-repo/feature/branch2", nil)

	// Mock worktree deletion
	mockWorktree.EXPECT().Delete(gomock.Any()).Return(nil).Times(2)

	// No logging expectations needed for nil logger

	err := repo.DeleteAllWorktrees(true)
	assert.NoError(t, err)
}

func TestRealRepository_DeleteAllWorktrees_NoWorktrees(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockFS := fsmocks.NewMockFS(ctrl)
	mockGit := gitmocks.NewMockGit(ctrl)
	mockStatus := statusMocks.NewMockManager(ctrl)
	// logger will be nil and use NewNoopLogger in real usage
	mockPrompt := promptMocks.NewMockPrompter(ctrl)
	mockWorktree := worktreeMocks.NewMockWorktree(ctrl)

	// Create repository instance
	repo := NewRepository(NewRepositoryParams{
		FS:               mockFS,
		Git:              mockGit,
		Config:           createTestConfig(),
		StatusManager:    mockStatus,
		Prompt:           mockPrompt,
		WorktreeProvider: func(params worktree.NewWorktreeParams) worktree.Worktree { return mockWorktree },
	})

	// Mock validation - IsGitRepository calls
	mockFS.EXPECT().Exists(".git").Return(true, nil)
	mockFS.EXPECT().IsDir(".git").Return(true, nil)

	// Mock validation - getRepositoryURL calls (called twice: once in validation, once in ListWorktrees)
	mockGit.EXPECT().GetRepositoryName(gomock.Any()).Return("test-repo", nil).Times(2)

	// Mock empty worktree list
	mockStatus.EXPECT().GetRepository("test-repo").Return(&status.Repository{
		Worktrees: map[string]status.WorktreeInfo{},
	}, nil)

	// No logging expectations needed for nil logger

	err := repo.DeleteAllWorktrees(true)
	assert.NoError(t, err)
}

func TestRealRepository_DeleteAllWorktrees_ValidationError(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockFS := fsmocks.NewMockFS(ctrl)
	mockGit := gitmocks.NewMockGit(ctrl)
	mockStatus := statusMocks.NewMockManager(ctrl)
	// logger will be nil and use NewNoopLogger in real usage
	mockPrompt := promptMocks.NewMockPrompter(ctrl)

	// Create repository instance
	repo := NewRepository(NewRepositoryParams{
		FS:               mockFS,
		Git:              mockGit,
		Config:           createTestConfig(),
		StatusManager:    mockStatus,
		Prompt:           mockPrompt,
		WorktreeProvider: func(params worktree.NewWorktreeParams) worktree.Worktree { return nil },
	})

	// Mock validation failure - IsGitRepository calls
	mockFS.EXPECT().Exists(".git").Return(true, nil)
	mockFS.EXPECT().IsDir(".git").Return(true, nil)

	// Mock validation failure - getRepositoryURL calls
	mockGit.EXPECT().GetRepositoryName(gomock.Any()).Return("", fmt.Errorf("not a git repository"))

	// No logging expectations needed for nil logger

	err := repo.DeleteAllWorktrees(true)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "not a git repository")
}

func TestRealRepository_DeleteAllWorktrees_ListWorktreesError(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockFS := fsmocks.NewMockFS(ctrl)
	mockGit := gitmocks.NewMockGit(ctrl)
	mockStatus := statusMocks.NewMockManager(ctrl)
	// logger will be nil and use NewNoopLogger in real usage
	mockPrompt := promptMocks.NewMockPrompter(ctrl)

	// Create repository instance
	repo := NewRepository(NewRepositoryParams{
		FS:               mockFS,
		Git:              mockGit,
		Config:           createTestConfig(),
		StatusManager:    mockStatus,
		Prompt:           mockPrompt,
		WorktreeProvider: func(params worktree.NewWorktreeParams) worktree.Worktree { return nil },
	})

	// Mock validation - IsGitRepository calls
	mockFS.EXPECT().Exists(".git").Return(true, nil)
	mockFS.EXPECT().IsDir(".git").Return(true, nil)

	// Mock validation - getRepositoryURL calls (called twice: once in validation, once in ListWorktrees)
	mockGit.EXPECT().GetRepositoryName(gomock.Any()).Return("test-repo", nil).Times(2)

	// Mock worktree listing error
	mockStatus.EXPECT().GetRepository("test-repo").Return(nil, fmt.Errorf("status file error"))

	// No logging expectations needed for nil logger

	err := repo.DeleteAllWorktrees(true)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "failed to list worktrees")
}

func TestRealRepository_DeleteAllWorktrees_PartialFailure(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockFS := fsmocks.NewMockFS(ctrl)
	mockGit := gitmocks.NewMockGit(ctrl)
	mockStatus := statusMocks.NewMockManager(ctrl)
	// logger will be nil and use NewNoopLogger in real usage
	mockPrompt := promptMocks.NewMockPrompter(ctrl)
	mockWorktree := worktreeMocks.NewMockWorktree(ctrl)

	// Create repository instance
	repo := NewRepository(NewRepositoryParams{
		FS:               mockFS,
		Git:              mockGit,
		Config:           createTestConfig(),
		StatusManager:    mockStatus,
		Prompt:           mockPrompt,
		WorktreeProvider: func(params worktree.NewWorktreeParams) worktree.Worktree { return mockWorktree },
	})

	// Mock validation - IsGitRepository calls
	mockFS.EXPECT().Exists(".git").Return(true, nil)
	mockFS.EXPECT().IsDir(".git").Return(true, nil)

	// Mock validation - getRepositoryURL calls (called twice: once in validation, once in ListWorktrees)
	mockGit.EXPECT().GetRepositoryName(gomock.Any()).Return("test-repo", nil).Times(2)

	// Mock worktree listing
	mockStatus.EXPECT().GetRepository("test-repo").Return(&status.Repository{
		Worktrees: map[string]status.WorktreeInfo{
			"feature/branch1": {Branch: "feature/branch1", Remote: "origin"},
			"feature/branch2": {Branch: "feature/branch2", Remote: "origin"},
		},
	}, nil)

	// Mock worktree path retrieval
	mockGit.EXPECT().GetWorktreePath(gomock.Any(), "feature/branch1").Return("/test/repos/worktrees/test-repo/feature/branch1", nil)
	mockGit.EXPECT().GetWorktreePath(gomock.Any(), "feature/branch2").Return("/test/repos/worktrees/test-repo/feature/branch2", nil)

	// Mock worktree deletion: first succeeds, second fails
	mockWorktree.EXPECT().Delete(gomock.Any()).Return(nil).Times(1)
	mockWorktree.EXPECT().Delete(gomock.Any()).Return(fmt.Errorf("deletion failed")).Times(1)

	// No logging expectations needed for nil logger

	err := repo.DeleteAllWorktrees(true)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "some worktrees failed to delete")
	assert.Contains(t, err.Error(), "feature/branch2")
}

func TestRealRepository_DeleteAllWorktrees_AllFailures(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockFS := fsmocks.NewMockFS(ctrl)
	mockGit := gitmocks.NewMockGit(ctrl)
	mockStatus := statusMocks.NewMockManager(ctrl)
	// logger will be nil and use NewNoopLogger in real usage
	mockPrompt := promptMocks.NewMockPrompter(ctrl)
	mockWorktree := worktreeMocks.NewMockWorktree(ctrl)

	// Create repository instance
	repo := NewRepository(NewRepositoryParams{
		FS:               mockFS,
		Git:              mockGit,
		Config:           createTestConfig(),
		StatusManager:    mockStatus,
		Prompt:           mockPrompt,
		WorktreeProvider: func(params worktree.NewWorktreeParams) worktree.Worktree { return mockWorktree },
	})

	// Mock validation - IsGitRepository calls
	mockFS.EXPECT().Exists(".git").Return(true, nil)
	mockFS.EXPECT().IsDir(".git").Return(true, nil)

	// Mock validation - getRepositoryURL calls (called twice: once in validation, once in ListWorktrees)
	mockGit.EXPECT().GetRepositoryName(gomock.Any()).Return("test-repo", nil).Times(2)

	// Mock worktree listing
	mockStatus.EXPECT().GetRepository("test-repo").Return(&status.Repository{
		Worktrees: map[string]status.WorktreeInfo{
			"feature/branch1": {Branch: "feature/branch1", Remote: "origin"},
			"feature/branch2": {Branch: "feature/branch2", Remote: "origin"},
		},
	}, nil)

	// Mock worktree path retrieval
	mockGit.EXPECT().GetWorktreePath(gomock.Any(), "feature/branch1").Return("/test/repos/worktrees/test-repo/feature/branch1", nil)
	mockGit.EXPECT().GetWorktreePath(gomock.Any(), "feature/branch2").Return("/test/repos/worktrees/test-repo/feature/branch2", nil)

	// Mock worktree deletion: both fail
	mockWorktree.EXPECT().Delete(gomock.Any()).Return(fmt.Errorf("deletion failed")).Times(2)

	// No logging expectations needed for nil logger

	err := repo.DeleteAllWorktrees(true)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "failed to delete all worktrees")
}

func TestRealRepository_DeleteAllWorktrees_GetWorktreePathError(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockFS := fsmocks.NewMockFS(ctrl)
	mockGit := gitmocks.NewMockGit(ctrl)
	mockStatus := statusMocks.NewMockManager(ctrl)
	// logger will be nil and use NewNoopLogger in real usage
	mockPrompt := promptMocks.NewMockPrompter(ctrl)
	mockWorktree := worktreeMocks.NewMockWorktree(ctrl)

	// Create repository instance
	repo := NewRepository(NewRepositoryParams{
		FS:               mockFS,
		Git:              mockGit,
		Config:           createTestConfig(),
		StatusManager:    mockStatus,
		Prompt:           mockPrompt,
		WorktreeProvider: func(params worktree.NewWorktreeParams) worktree.Worktree { return mockWorktree },
	})

	// Mock validation - IsGitRepository calls
	mockFS.EXPECT().Exists(".git").Return(true, nil)
	mockFS.EXPECT().IsDir(".git").Return(true, nil)

	// Mock validation - getRepositoryURL calls (called twice: once in validation, once in ListWorktrees)
	mockGit.EXPECT().GetRepositoryName(gomock.Any()).Return("test-repo", nil).Times(2)

	// Mock worktree listing
	mockStatus.EXPECT().GetRepository("test-repo").Return(&status.Repository{
		Worktrees: map[string]status.WorktreeInfo{
			"feature/branch1": {Branch: "feature/branch1", Remote: "origin"},
		},
	}, nil)

	// Mock worktree path retrieval error
	mockGit.EXPECT().GetWorktreePath(gomock.Any(), "feature/branch1").Return("", fmt.Errorf("path not found"))

	// Mock RemoveFromStatus call (new behavior when worktree doesn't exist in Git)
	mockWorktree.EXPECT().RemoveFromStatus("test-repo", "feature/branch1").Return(nil)

	// No logging expectations needed for nil logger

	err := repo.DeleteAllWorktrees(true)
	assert.NoError(t, err) // Should succeed now since we remove from status
}
