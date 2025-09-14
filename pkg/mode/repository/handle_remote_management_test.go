//go:build unit

package repository

import (
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

func TestHandleRemoteManagement_Origin(t *testing.T) {
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

	// Origin remote should be handled without any git operations
	err := repository.HandleRemoteManagement("origin")
	assert.NoError(t, err)
}

func TestHandleRemoteManagement_ExistingRemote(t *testing.T) {
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

	// Mock existing remote check
	mockGit.EXPECT().RemoteExists("/test/repo", "upstream").Return(true, nil)
	mockGit.EXPECT().GetRemoteURL("/test/repo", "upstream").Return("https://github.com/upstream/example.git", nil)

	err := repository.HandleRemoteManagement("upstream")
	assert.NoError(t, err)
}

func TestHandleRemoteManagement_NewRemote(t *testing.T) {
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

	// Mock remote doesn't exist, so we need to add it
	mockGit.EXPECT().RemoteExists("/test/repo", "upstream").Return(false, nil)
	mockGit.EXPECT().GetRepositoryName("/test/repo").Return("github.com/octocat/Hello-World", nil)
	mockGit.EXPECT().GetRemoteURL("/test/repo", "origin").Return("https://github.com/octocat/Hello-World.git", nil)
	mockGit.EXPECT().AddRemote("/test/repo", "upstream", "https://github.com/upstream/Hello-World.git").Return(nil)

	err := repository.HandleRemoteManagement("upstream")
	assert.NoError(t, err)
}

func TestExtractHostFromURL_HTTPS(t *testing.T) {
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

	host := repository.ExtractHostFromURL("https://github.com/octocat/Hello-World.git")
	assert.Equal(t, "github.com", host)
}

func TestExtractHostFromURL_SSH(t *testing.T) {
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

	host := repository.ExtractHostFromURL("git@github.com:octocat/Hello-World.git")
	assert.Equal(t, "github.com", host)
}

func TestDetermineProtocol_HTTPS(t *testing.T) {
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

	protocol := repository.DetermineProtocol("https://github.com/octocat/Hello-World.git")
	assert.Equal(t, "https", protocol)
}

func TestDetermineProtocol_SSH(t *testing.T) {
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

	protocol := repository.DetermineProtocol("git@github.com:octocat/Hello-World.git")
	assert.Equal(t, "ssh", protocol)
}

func TestExtractHostFromURL_Invalid(t *testing.T) {
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

	host := repository.ExtractHostFromURL("invalid-url")
	assert.Empty(t, host)
}

func TestExtractRepoNameFromFullPath_Success(t *testing.T) {
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

	repoName := repository.ExtractRepoNameFromFullPath("github.com/octocat/Hello-World")
	assert.Equal(t, "Hello-World", repoName)
}

func TestExtractRepoNameFromFullPath_SinglePart(t *testing.T) {
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

	repoName := repository.ExtractRepoNameFromFullPath("example")
	assert.Equal(t, "example", repoName)
}

func TestConstructRemoteURL_HTTPS(t *testing.T) {
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

	remoteURL, err := repository.ConstructRemoteURL("https://github.com/octocat/Hello-World.git", "upstream", "github.com/octocat/Hello-World")
	assert.NoError(t, err)
	assert.Equal(t, "https://github.com/upstream/Hello-World.git", remoteURL)
}

func TestConstructRemoteURL_SSH(t *testing.T) {
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

	remoteURL, err := repository.ConstructRemoteURL("git@github.com:lerenn/example.git", "upstream", "github.com/octocat/Hello-World")
	assert.NoError(t, err)
	assert.Equal(t, "git@github.com:upstream/Hello-World.git", remoteURL)
}

func TestConstructRemoteURL_InvalidHost(t *testing.T) {
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

	remoteURL, err := repository.ConstructRemoteURL("invalid-url", "upstream", "github.com/octocat/Hello-World")
	assert.Error(t, err)
	assert.Empty(t, remoteURL)
}
