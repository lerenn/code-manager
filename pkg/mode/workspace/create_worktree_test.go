//go:build unit

package workspace

import (
	"errors"
	"path/filepath"
	"testing"

	"github.com/lerenn/code-manager/pkg/config"
	configmocks "github.com/lerenn/code-manager/pkg/config/mocks"
	"github.com/lerenn/code-manager/pkg/dependencies"
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
	mockConfig := configmocks.NewMockManager(ctrl)

	workspace := &realWorkspace{
		deps: &dependencies.Dependencies{
			FS:                 mockFS,
			Git:                mockGit,
			StatusManager:      mockStatus,
			Logger:             logger.NewNoopLogger(),
			Prompt:             mockPrompt,
			WorktreeProvider:   func(params worktree.NewWorktreeParams) worktree.Worktree { return worktreemocks.NewMockWorktree(ctrl) },
			RepositoryProvider: func(params repository.NewRepositoryParams) repository.Repository { return mockRepository },
			Config:             mockConfig,
		},
	}

	workspaceName := "test-workspace"
	branch := "feature-branch"
	repositories := []string{"github.com/user/repo1", "github.com/user/repo2"}

	// Mock config manager
	testConfig := config.Config{
		RepositoriesDir: "/test/repos",
		WorkspacesDir:   "/test/workspaces",
		StatusFile:      "/test/status.yaml",
	}
	mockConfig.EXPECT().GetConfigWithFallback().Return(testConfig, nil).AnyTimes()

	// Mock workspace retrieval
	mockStatus.EXPECT().GetWorkspace(workspaceName).Return(&status.Workspace{
		Repositories: repositories,
	}, nil)

	// Mock repository path validation and worktree creation for each repo
	// Use AnyTimes() to allow flexible ordering
	for _, repoURL := range repositories {
		repoPath := filepath.Join("/test/repos", repoURL)
		// Mock GetRepository to return error (so it falls back to constructed path)
		mockStatus.EXPECT().GetRepository(repoURL).Return(nil, errors.New("not found")).AnyTimes()
		mockFS.EXPECT().Exists(repoPath).Return(true, nil).AnyTimes()
		// Mock getRepositoryURLFromPath calls - this is called by createSingleRepositoryWorktreeWithURL
		mockFS.EXPECT().Exists(filepath.Join(repoPath, ".git")).Return(true, nil).AnyTimes()
		mockGit.EXPECT().GetRemoteURL(repoPath, "origin").Return("https://github.com/user/"+filepath.Base(repoURL)+".git", nil).AnyTimes()

		// Mock repository worktree creation for this repository
		worktreePath := filepath.Join("/test/repos", repoURL, "worktrees", "origin", branch)
		mockRepository.EXPECT().CreateWorktree(branch, gomock.Any()).Return(worktreePath, nil).AnyTimes()
	}

	// Mock workspace file creation
	mockFS.EXPECT().MkdirAll("/test/workspaces", gomock.Any()).Return(nil)
	mockFS.EXPECT().CreateFileWithContent(gomock.Any(), gomock.Any(), gomock.Any()).Return(nil)

	// Mock workspace status update
	mockStatus.EXPECT().GetWorkspace(workspaceName).Return(&status.Workspace{
		Worktrees:    []string{},
		Repositories: repositories,
	}, nil).AnyTimes()
	// Mock UpdateWorkspace call - should contain the branch name in worktrees and actual repository URLs
	mockStatus.EXPECT().UpdateWorkspace(workspaceName, gomock.Any()).DoAndReturn(func(name string, workspace status.Workspace) error {
		// Verify that the worktree array contains just the branch name
		assert.Contains(t, workspace.Worktrees, branch)
		// Verify that the repositories array contains the extracted repository URLs
		expectedRepos := []string{"github.com/user/repo1", "github.com/user/repo2"}
		assert.Equal(t, expectedRepos, workspace.Repositories)
		return nil
	}).AnyTimes()

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
		deps: &dependencies.Dependencies{
			Logger: logger.NewNoopLogger(),
		},
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
		deps: &dependencies.Dependencies{
			StatusManager: mockStatus,
			Logger:        logger.NewNoopLogger(),
		},
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
		deps: &dependencies.Dependencies{
			StatusManager: mockStatus,
			Logger:        logger.NewNoopLogger(),
		},
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
	mockConfig := configmocks.NewMockManager(ctrl)

	workspace := &realWorkspace{
		deps: &dependencies.Dependencies{
			FS:            mockFS,
			StatusManager: mockStatus,
			Logger:        logger.NewNoopLogger(),
			Config:        mockConfig,
		},
	}

	workspaceName := "test-workspace"
	repositories := []string{"github.com/user/repo1"}

	// Mock config manager
	testConfig := config.Config{
		RepositoriesDir: "/test/repos",
		WorkspacesDir:   "/test/workspaces",
		StatusFile:      "/test/status.yaml",
	}
	mockConfig.EXPECT().GetConfigWithFallback().Return(testConfig, nil).AnyTimes()

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
	mockConfig := configmocks.NewMockManager(ctrl)

	workspace := &realWorkspace{
		deps: &dependencies.Dependencies{
			FS:            mockFS,
			StatusManager: mockStatus,
			Logger:        logger.NewNoopLogger(),
			Config:        mockConfig,
		},
	}

	workspaceName := "test-workspace"
	repositories := []string{"github.com/user/repo1"}

	// Mock config manager
	testConfig := config.Config{
		RepositoriesDir: "/test/repos",
		WorkspacesDir:   "/test/workspaces",
		StatusFile:      "/test/status.yaml",
	}
	mockConfig.EXPECT().GetConfigWithFallback().Return(testConfig, nil).AnyTimes()

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
	mockGit := gitmocks.NewMockGit(ctrl)
	mockConfig := configmocks.NewMockManager(ctrl)

	workspace := &realWorkspace{
		deps: &dependencies.Dependencies{
			FS:               mockFS,
			Git:              mockGit,
			StatusManager:    mockStatus,
			Logger:           logger.NewNoopLogger(),
			WorktreeProvider: func(params worktree.NewWorktreeParams) worktree.Worktree { return mockWorktree },
			RepositoryProvider: func(params repository.NewRepositoryParams) repository.Repository {
				return repositorymocks.NewMockRepository(ctrl)
			},
			Config: mockConfig,
		},
	}

	workspaceName := "test-workspace"
	repositories := []string{"github.com/user/repo1"}

	// Mock config manager
	testConfig := config.Config{
		RepositoriesDir: "/test/repos",
		WorkspacesDir:   "/test/workspaces",
		StatusFile:      "/test/status.yaml",
	}
	mockConfig.EXPECT().GetConfigWithFallback().Return(testConfig, nil).AnyTimes()

	// Mock workspace retrieval
	mockStatus.EXPECT().GetWorkspace(workspaceName).Return(&status.Workspace{
		Repositories: repositories,
	}, nil)

	// Mock repository path validation
	repoPath := filepath.Join("/test/repos", repositories[0])
	// Mock GetRepository to return error (so it falls back to constructed path)
	mockStatus.EXPECT().GetRepository(repositories[0]).Return(nil, errors.New("not found")).AnyTimes()
	mockFS.EXPECT().Exists(repoPath).Return(true, nil).AnyTimes()
	// Mock getRepositoryURLFromPath calls
	mockFS.EXPECT().Exists(filepath.Join(repoPath, ".git")).Return(true, nil).AnyTimes()
	mockGit.EXPECT().GetRemoteURL(repoPath, "origin").Return("https://github.com/user/repo1.git", nil).AnyTimes()

	// Mock repository worktree creation failure
	mockRepository := repositorymocks.NewMockRepository(ctrl)
	workspace.deps.RepositoryProvider = func(params repository.NewRepositoryParams) repository.Repository { return mockRepository }
	mockRepository.EXPECT().CreateWorktree("feature-branch", gomock.Any()).Return("", errors.New("validation failed")).AnyTimes()

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
	mockGit := gitmocks.NewMockGit(ctrl)
	mockConfig := configmocks.NewMockManager(ctrl)

	workspace := &realWorkspace{
		deps: &dependencies.Dependencies{
			FS:               mockFS,
			Git:              mockGit,
			StatusManager:    mockStatus,
			Logger:           logger.NewNoopLogger(),
			WorktreeProvider: func(params worktree.NewWorktreeParams) worktree.Worktree { return mockWorktree },
			RepositoryProvider: func(params repository.NewRepositoryParams) repository.Repository {
				return repositorymocks.NewMockRepository(ctrl)
			},
			Config: mockConfig,
		},
	}

	workspaceName := "test-workspace"
	repositories := []string{"github.com/user/repo1"}

	// Mock config manager
	testConfig := config.Config{
		RepositoriesDir: "/test/repos",
		WorkspacesDir:   "/test/workspaces",
		StatusFile:      "/test/status.yaml",
	}
	mockConfig.EXPECT().GetConfigWithFallback().Return(testConfig, nil).AnyTimes()

	// Mock workspace retrieval
	mockStatus.EXPECT().GetWorkspace(workspaceName).Return(&status.Workspace{
		Repositories: repositories,
	}, nil)

	// Mock repository path validation
	repoPath := filepath.Join("/test/repos", repositories[0])
	// Mock GetRepository to return error (so it falls back to constructed path)
	mockStatus.EXPECT().GetRepository(repositories[0]).Return(nil, errors.New("not found")).AnyTimes()
	mockFS.EXPECT().Exists(repoPath).Return(true, nil).AnyTimes()
	// Mock getRepositoryURLFromPath calls
	mockFS.EXPECT().Exists(filepath.Join(repoPath, ".git")).Return(true, nil).AnyTimes()
	mockGit.EXPECT().GetRemoteURL(repoPath, "origin").Return("https://github.com/user/repo1.git", nil).AnyTimes()

	// Mock repository worktree creation failure
	mockRepository := repositorymocks.NewMockRepository(ctrl)
	workspace.deps.RepositoryProvider = func(params repository.NewRepositoryParams) repository.Repository { return mockRepository }
	mockRepository.EXPECT().CreateWorktree("feature-branch", gomock.Any()).Return("", errors.New("creation failed")).AnyTimes()

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
	mockConfig := configmocks.NewMockManager(ctrl)

	mockRepository := repositorymocks.NewMockRepository(ctrl)
	workspace := &realWorkspace{
		deps: &dependencies.Dependencies{
			FS:                 mockFS,
			Git:                mockGit,
			StatusManager:      mockStatus,
			Logger:             logger.NewNoopLogger(),
			WorktreeProvider:   func(params worktree.NewWorktreeParams) worktree.Worktree { return mockWorktree },
			RepositoryProvider: func(params repository.NewRepositoryParams) repository.Repository { return mockRepository },
			Config:             mockConfig,
		},
	}

	workspaceName := "test-workspace"
	repositories := []string{"github.com/user/repo1"}

	// Mock config manager
	testConfig := config.Config{
		RepositoriesDir: "/test/repos",
		WorkspacesDir:   "/test/workspaces",
		StatusFile:      "/test/status.yaml",
	}
	mockConfig.EXPECT().GetConfigWithFallback().Return(testConfig, nil).AnyTimes()

	// Mock workspace retrieval
	mockStatus.EXPECT().GetWorkspace(workspaceName).Return(&status.Workspace{
		Repositories: repositories,
	}, nil)

	// Mock repository path validation
	repoPath := filepath.Join("/test/repos", repositories[0])
	// Mock GetRepository to return error (so it falls back to constructed path)
	mockStatus.EXPECT().GetRepository(repositories[0]).Return(nil, errors.New("not found")).AnyTimes()
	mockFS.EXPECT().Exists(repoPath).Return(true, nil).AnyTimes()
	// Mock getRepositoryURLFromPath calls
	mockFS.EXPECT().Exists(filepath.Join(repoPath, ".git")).Return(true, nil).AnyTimes()
	mockGit.EXPECT().GetRemoteURL(repoPath, "origin").Return("https://github.com/user/repo1.git", nil).AnyTimes()

	// Mock repository worktree creation success
	worktreePath := filepath.Join("/test/repos", repositories[0], "worktrees", "origin", "feature-branch")
	mockRepository.EXPECT().CreateWorktree("feature-branch", gomock.Any()).Return(worktreePath, nil).AnyTimes()

	// Mock workspace file creation failure
	mockFS.EXPECT().MkdirAll("/test/workspaces", gomock.Any()).Return(errors.New("mkdir failed")).AnyTimes()

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
		deps: &dependencies.Dependencies{
			StatusManager: mockStatus,
			Logger:        logger.NewNoopLogger(),
		},
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
		deps: &dependencies.Dependencies{
			StatusManager: mockStatus,
			Logger:        logger.NewNoopLogger(),
		},
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
		deps: &dependencies.Dependencies{
			StatusManager: mockStatus,
			Logger:        logger.NewNoopLogger(),
		},
	}

	workspaceName := "empty-workspace"

	mockStatus.EXPECT().GetWorkspace(workspaceName).Return(&status.Workspace{
		Repositories: []string{},
	}, nil)

	_, err := workspace.validateAndGetRepositories(workspaceName)

	assert.Error(t, err)
	assert.Contains(t, err.Error(), "workspace 'empty-workspace' has no repositories defined")
}
