//go:build unit

package workspace

import (
	"errors"
	"path/filepath"
	"testing"

	"github.com/lerenn/code-manager/pkg/config"
	fsmocks "github.com/lerenn/code-manager/pkg/fs/mocks"
	gitmocks "github.com/lerenn/code-manager/pkg/git/mocks"
	"github.com/lerenn/code-manager/pkg/logger"
	"github.com/lerenn/code-manager/pkg/mode/repository"
	repositorymocks "github.com/lerenn/code-manager/pkg/mode/repository/mocks"
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
	mockRepository := repositorymocks.NewMockRepository(ctrl)

	workspace := &realWorkspace{
		fs:                 mockFS,
		git:                mockGit,
		statusManager:      mockStatus,
		logger:             logger.NewNoopLogger(),
		prompt:             mockPrompt,
		worktreeProvider:   func(_ worktree.NewWorktreeParams) worktree.Worktree { return worktreemocks.NewMockWorktree(ctrl) },
		repositoryProvider: func(_ repository.NewRepositoryParams) repository.Repository { return mockRepository },
		config: config.Config{
			RepositoriesDir: "/test/repos",
			WorkspacesDir:   "/test/workspaces",
		},
	}

	workspaceName := "test-workspace"
	branch := "feature-branch"
	repositories := []string{"github.com/user/repo1", "github.com/user/repo2"}

	// Mock workspace retrieval
	mockStatus.EXPECT().GetWorkspace(workspaceName).Return(&status.Workspace{
		Repositories: repositories,
	}, nil)

	// Mock repository path validation for each repo
	for _, repoURL := range repositories {
		repoPath := filepath.Join("/test/repos", repoURL)
		// Mock GetRepository to return error (so it falls back to constructed path)
		mockStatus.EXPECT().GetRepository(repoURL).Return(nil, errors.New("not found"))
		mockFS.EXPECT().Exists(repoPath).Return(true, nil)
		mockFS.EXPECT().Exists(filepath.Join(repoPath, ".git")).Return(true, nil)
	}

	// Mock repository worktree creation for each repository
	for _, repoURL := range repositories {
		worktreePath := filepath.Join("/test/repos", repoURL, "worktrees", "origin", branch)
		mockRepository.EXPECT().CreateWorktree(branch, gomock.Any()).Return(worktreePath, nil)
	}

	// Mock workspace file creation
	mockFS.EXPECT().MkdirAll("/test/workspaces", gomock.Any()).Return(nil)
	mockFS.EXPECT().CreateFileWithContent(gomock.Any(), gomock.Any(), gomock.Any()).Return(nil)

	// Mock workspace status update
	mockStatus.EXPECT().GetWorkspace(workspaceName).Return(&status.Workspace{
		Worktrees:    []string{},
		Repositories: repositories,
	}, nil)
	// Mock UpdateWorkspace call - should contain the branch name in worktrees
	mockStatus.EXPECT().UpdateWorkspace(workspaceName, gomock.Any()).DoAndReturn(func(name string, workspace status.Workspace) error {
		// Verify that the worktree array contains just the branch name
		assert.Contains(t, workspace.Worktrees, branch)
		return nil
	})

	opts := []CreateWorktreeOpts{
		{WorkspaceName: workspaceName},
	}

	result, err := workspace.CreateWorktree(branch, opts...)

	assert.NoError(t, err)
	assert.Contains(t, result, "test-workspace-feature-branch.code-workspace")
}

func TestCreateWorktree_MissingWorkspaceName(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	workspace := &realWorkspace{
		logger: logger.NewNoopLogger(),
	}

	opts := []CreateWorktreeOpts{
		{WorkspaceName: ""},
	}

	_, err := workspace.CreateWorktree("feature-branch", opts...)

	assert.Error(t, err)
	assert.Contains(t, err.Error(), "workspace name is required")
}

func TestCreateWorktree_WorkspaceNotFound(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockStatus := statusmocks.NewMockManager(ctrl)

	workspace := &realWorkspace{
		statusManager: mockStatus,
		logger:        logger.NewNoopLogger(),
	}

	workspaceName := "nonexistent-workspace"

	// Mock workspace not found
	mockStatus.EXPECT().GetWorkspace(workspaceName).Return(nil, errors.New("workspace not found"))

	opts := []CreateWorktreeOpts{
		{WorkspaceName: workspaceName},
	}

	_, err := workspace.CreateWorktree("feature-branch", opts...)

	assert.Error(t, err)
	assert.Contains(t, err.Error(), "workspace 'nonexistent-workspace' not found")
}

func TestCreateWorktree_EmptyRepositories(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockStatus := statusmocks.NewMockManager(ctrl)

	workspace := &realWorkspace{
		statusManager: mockStatus,
		logger:        logger.NewNoopLogger(),
	}

	workspaceName := "empty-workspace"

	// Mock workspace with no repositories
	mockStatus.EXPECT().GetWorkspace(workspaceName).Return(&status.Workspace{
		Repositories: []string{},
	}, nil)

	opts := []CreateWorktreeOpts{
		{WorkspaceName: workspaceName},
	}

	_, err := workspace.CreateWorktree("feature-branch", opts...)

	assert.Error(t, err)
	assert.Contains(t, err.Error(), "workspace 'empty-workspace' has no repositories defined")
}

func TestCreateWorktree_RepositoryPathNotFound(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockFS := fsmocks.NewMockFS(ctrl)
	mockStatus := statusmocks.NewMockManager(ctrl)

	workspace := &realWorkspace{
		fs:            mockFS,
		statusManager: mockStatus,
		logger:        logger.NewNoopLogger(),
		config: config.Config{
			RepositoriesDir: "/test/repos",
		},
	}

	workspaceName := "test-workspace"
	repositories := []string{"github.com/user/repo1"}

	// Mock workspace retrieval
	mockStatus.EXPECT().GetWorkspace(workspaceName).Return(&status.Workspace{
		Repositories: repositories,
	}, nil)

	// Mock repository path not found
	repoPath := filepath.Join("/test/repos", repositories[0])
	// Mock GetRepository to return error (so it falls back to constructed path)
	mockStatus.EXPECT().GetRepository(repositories[0]).Return(nil, errors.New("not found"))
	mockFS.EXPECT().Exists(repoPath).Return(false, nil)

	opts := []CreateWorktreeOpts{
		{WorkspaceName: workspaceName},
	}

	_, err := workspace.CreateWorktree("feature-branch", opts...)

	assert.Error(t, err)
	assert.Contains(t, err.Error(), "repository path does not exist")
}

func TestCreateWorktree_InvalidGitRepository(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockFS := fsmocks.NewMockFS(ctrl)
	mockStatus := statusmocks.NewMockManager(ctrl)

	workspace := &realWorkspace{
		fs:            mockFS,
		statusManager: mockStatus,
		logger:        logger.NewNoopLogger(),
		config: config.Config{
			RepositoriesDir: "/test/repos",
		},
	}

	workspaceName := "test-workspace"
	repositories := []string{"github.com/user/repo1"}

	// Mock workspace retrieval
	mockStatus.EXPECT().GetWorkspace(workspaceName).Return(&status.Workspace{
		Repositories: repositories,
	}, nil)

	// Mock repository path exists but not a Git repository
	repoPath := filepath.Join("/test/repos", repositories[0])
	// Mock GetRepository to return error (so it falls back to constructed path)
	mockStatus.EXPECT().GetRepository(repositories[0]).Return(nil, errors.New("not found"))
	mockFS.EXPECT().Exists(repoPath).Return(true, nil)
	mockFS.EXPECT().Exists(filepath.Join(repoPath, ".git")).Return(false, nil)

	opts := []CreateWorktreeOpts{
		{WorkspaceName: workspaceName},
	}

	_, err := workspace.CreateWorktree("feature-branch", opts...)

	assert.Error(t, err)
	assert.Contains(t, err.Error(), "path is not a Git repository")
}

func TestCreateWorktree_WorktreeValidationFailure(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockFS := fsmocks.NewMockFS(ctrl)
	mockStatus := statusmocks.NewMockManager(ctrl)
	mockWorktree := worktreemocks.NewMockWorktree(ctrl)

	workspace := &realWorkspace{
		fs:               mockFS,
		statusManager:    mockStatus,
		logger:           logger.NewNoopLogger(),
		worktreeProvider: func(_ worktree.NewWorktreeParams) worktree.Worktree { return mockWorktree },
		repositoryProvider: func(_ repository.NewRepositoryParams) repository.Repository {
			return repositorymocks.NewMockRepository(ctrl)
		},
		config: config.Config{
			RepositoriesDir: "/test/repos",
		},
	}

	workspaceName := "test-workspace"
	repositories := []string{"github.com/user/repo1"}

	// Mock workspace retrieval
	mockStatus.EXPECT().GetWorkspace(workspaceName).Return(&status.Workspace{
		Repositories: repositories,
	}, nil)

	// Mock repository path validation
	repoPath := filepath.Join("/test/repos", repositories[0])
	// Mock GetRepository to return error (so it falls back to constructed path)
	mockStatus.EXPECT().GetRepository(repositories[0]).Return(nil, errors.New("not found"))
	mockFS.EXPECT().Exists(repoPath).Return(true, nil)
	mockFS.EXPECT().Exists(filepath.Join(repoPath, ".git")).Return(true, nil)

	// Mock repository worktree creation failure
	mockRepository := repositorymocks.NewMockRepository(ctrl)
	workspace.repositoryProvider = func(_ repository.NewRepositoryParams) repository.Repository { return mockRepository }
	mockRepository.EXPECT().CreateWorktree("feature-branch", gomock.Any()).Return("", errors.New("validation failed"))

	opts := []CreateWorktreeOpts{
		{WorkspaceName: workspaceName},
	}

	_, err := workspace.CreateWorktree("feature-branch", opts...)

	assert.Error(t, err)
	assert.Contains(t, err.Error(), "failed to create worktree using repository")
}

func TestCreateWorktree_WorktreeCreationFailure(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockFS := fsmocks.NewMockFS(ctrl)
	mockStatus := statusmocks.NewMockManager(ctrl)
	mockWorktree := worktreemocks.NewMockWorktree(ctrl)

	workspace := &realWorkspace{
		fs:               mockFS,
		statusManager:    mockStatus,
		logger:           logger.NewNoopLogger(),
		worktreeProvider: func(_ worktree.NewWorktreeParams) worktree.Worktree { return mockWorktree },
		repositoryProvider: func(_ repository.NewRepositoryParams) repository.Repository {
			return repositorymocks.NewMockRepository(ctrl)
		},
		config: config.Config{
			RepositoriesDir: "/test/repos",
		},
	}

	workspaceName := "test-workspace"
	repositories := []string{"github.com/user/repo1"}

	// Mock workspace retrieval
	mockStatus.EXPECT().GetWorkspace(workspaceName).Return(&status.Workspace{
		Repositories: repositories,
	}, nil)

	// Mock repository path validation
	repoPath := filepath.Join("/test/repos", repositories[0])
	// Mock GetRepository to return error (so it falls back to constructed path)
	mockStatus.EXPECT().GetRepository(repositories[0]).Return(nil, errors.New("not found"))
	mockFS.EXPECT().Exists(repoPath).Return(true, nil)
	mockFS.EXPECT().Exists(filepath.Join(repoPath, ".git")).Return(true, nil)

	// Mock repository worktree creation failure
	mockRepository := repositorymocks.NewMockRepository(ctrl)
	workspace.repositoryProvider = func(_ repository.NewRepositoryParams) repository.Repository { return mockRepository }
	mockRepository.EXPECT().CreateWorktree("feature-branch", gomock.Any()).Return("", errors.New("creation failed"))

	opts := []CreateWorktreeOpts{
		{WorkspaceName: workspaceName},
	}

	_, err := workspace.CreateWorktree("feature-branch", opts...)

	assert.Error(t, err)
	assert.Contains(t, err.Error(), "failed to create worktree")
}

func TestCreateWorktree_WorkspaceFileCreationFailure(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockFS := fsmocks.NewMockFS(ctrl)
	mockGit := gitmocks.NewMockGit(ctrl)
	mockStatus := statusmocks.NewMockManager(ctrl)
	mockWorktree := worktreemocks.NewMockWorktree(ctrl)

	mockRepository := repositorymocks.NewMockRepository(ctrl)
	workspace := &realWorkspace{
		fs:                 mockFS,
		git:                mockGit,
		statusManager:      mockStatus,
		logger:             logger.NewNoopLogger(),
		worktreeProvider:   func(_ worktree.NewWorktreeParams) worktree.Worktree { return mockWorktree },
		repositoryProvider: func(_ repository.NewRepositoryParams) repository.Repository { return mockRepository },
		config: config.Config{
			RepositoriesDir: "/test/repos",
			WorkspacesDir:   "/test/workspaces",
		},
	}

	workspaceName := "test-workspace"
	repositories := []string{"github.com/user/repo1"}

	// Mock workspace retrieval
	mockStatus.EXPECT().GetWorkspace(workspaceName).Return(&status.Workspace{
		Repositories: repositories,
	}, nil)

	// Mock repository path validation
	repoPath := filepath.Join("/test/repos", repositories[0])
	// Mock GetRepository to return error (so it falls back to constructed path)
	mockStatus.EXPECT().GetRepository(repositories[0]).Return(nil, errors.New("not found"))
	mockFS.EXPECT().Exists(repoPath).Return(true, nil)
	mockFS.EXPECT().Exists(filepath.Join(repoPath, ".git")).Return(true, nil)

	// Mock repository worktree creation success
	worktreePath := filepath.Join("/test/repos", repositories[0], "worktrees", "origin", "feature-branch")
	mockRepository.EXPECT().CreateWorktree("feature-branch", gomock.Any()).Return(worktreePath, nil)

	// Mock workspace file creation failure
	mockFS.EXPECT().MkdirAll("/test/workspaces", gomock.Any()).Return(errors.New("mkdir failed"))

	opts := []CreateWorktreeOpts{
		{WorkspaceName: workspaceName},
	}

	_, err := workspace.CreateWorktree("feature-branch", opts...)

	assert.Error(t, err)
	assert.Contains(t, err.Error(), "failed to create workspace file")
}

func TestExtractWorkspaceName_Success(t *testing.T) {
	workspace := &realWorkspace{}

	opts := []CreateWorktreeOpts{
		{WorkspaceName: "test-workspace"},
	}

	result, err := workspace.extractWorkspaceName(opts)

	assert.NoError(t, err)
	assert.Equal(t, "test-workspace", result)
}

func TestExtractWorkspaceName_EmptyOptions(t *testing.T) {
	workspace := &realWorkspace{}

	opts := []CreateWorktreeOpts{}

	_, err := workspace.extractWorkspaceName(opts)

	assert.Error(t, err)
	assert.Contains(t, err.Error(), "workspace name is required")
}

func TestExtractWorkspaceName_EmptyWorkspaceName(t *testing.T) {
	workspace := &realWorkspace{}

	opts := []CreateWorktreeOpts{
		{WorkspaceName: ""},
	}

	_, err := workspace.extractWorkspaceName(opts)

	assert.Error(t, err)
	assert.Contains(t, err.Error(), "workspace name is required")
}

func TestValidateAndGetRepositories_Success(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockStatus := statusmocks.NewMockManager(ctrl)

	workspace := &realWorkspace{
		statusManager: mockStatus,
		logger:        logger.NewNoopLogger(),
	}

	workspaceName := "test-workspace"
	repositories := []string{"github.com/user/repo1", "github.com/user/repo2"}

	mockStatus.EXPECT().GetWorkspace(workspaceName).Return(&status.Workspace{
		Repositories: repositories,
	}, nil)

	result, err := workspace.validateAndGetRepositories(workspaceName)

	assert.NoError(t, err)
	assert.Equal(t, repositories, result)
}

func TestValidateAndGetRepositories_WorkspaceNotFound(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockStatus := statusmocks.NewMockManager(ctrl)

	workspace := &realWorkspace{
		statusManager: mockStatus,
		logger:        logger.NewNoopLogger(),
	}

	workspaceName := "nonexistent-workspace"

	mockStatus.EXPECT().GetWorkspace(workspaceName).Return(nil, errors.New("not found"))

	_, err := workspace.validateAndGetRepositories(workspaceName)

	assert.Error(t, err)
	assert.Contains(t, err.Error(), "workspace 'nonexistent-workspace' not found")
}

func TestValidateAndGetRepositories_EmptyRepositories(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockStatus := statusmocks.NewMockManager(ctrl)

	workspace := &realWorkspace{
		statusManager: mockStatus,
		logger:        logger.NewNoopLogger(),
	}

	workspaceName := "empty-workspace"

	mockStatus.EXPECT().GetWorkspace(workspaceName).Return(&status.Workspace{
		Repositories: []string{},
	}, nil)

	_, err := workspace.validateAndGetRepositories(workspaceName)

	assert.Error(t, err)
	assert.Contains(t, err.Error(), "workspace 'empty-workspace' has no repositories defined")
}
