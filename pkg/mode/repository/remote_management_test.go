//go:build unit

package repository

import (
	"testing"

	fsmocks "github.com/lerenn/code-manager/pkg/fs/mocks"
	gitmocks "github.com/lerenn/code-manager/pkg/git/mocks"
	promptmocks "github.com/lerenn/code-manager/pkg/prompt/mocks"
	statusmocks "github.com/lerenn/code-manager/pkg/status/mocks"
	"github.com/stretchr/testify/assert"
	"go.uber.org/mock/gomock"
)

func TestRepository_HandleRemoteManagement_Origin(t *testing.T) {
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

	err := repo.HandleRemoteManagement("origin")
	assert.NoError(t, err)
}

func TestRepository_HandleRemoteManagement_ExistingRemote(t *testing.T) {
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

	mockGit.EXPECT().RemoteExists(".", "upstream").Return(true, nil)
	mockGit.EXPECT().GetRemoteURL(".", "upstream").Return("https://github.com/upstream/example.git", nil)

	err := repo.HandleRemoteManagement("upstream")
	assert.NoError(t, err)
}

func TestRepository_HandleRemoteManagement_NewRemote(t *testing.T) {
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

	mockGit.EXPECT().RemoteExists(".", "upstream").Return(false, nil)
	mockGit.EXPECT().GetRepositoryName(".").Return("github.com/octocat/Hello-World", nil)
	mockGit.EXPECT().GetRemoteURL(".", "origin").Return("https://github.com/octocat/Hello-World.git", nil)
	mockGit.EXPECT().AddRemote(".", "upstream", "https://github.com/upstream/Hello-World.git").Return(nil)

	err := repo.HandleRemoteManagement("upstream")
	assert.NoError(t, err)
}

func TestRepository_ExtractHostFromURL_HTTPS(t *testing.T) {
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

	host := repo.ExtractHostFromURL("https://github.com/octocat/Hello-World.git")
	assert.Equal(t, "github.com", host)
}

func TestRepository_ExtractHostFromURL_SSH(t *testing.T) {
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

	host := repo.ExtractHostFromURL("git@github.com:lerenn/example.git")
	assert.Equal(t, "github.com", host)
}

func TestRepository_ExtractHostFromURL_Invalid(t *testing.T) {
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

	host := repo.ExtractHostFromURL("invalid-url")
	assert.Empty(t, host)
}

func TestRepository_DetermineProtocol_HTTPS(t *testing.T) {
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

	protocol := repo.DetermineProtocol("https://github.com/octocat/Hello-World.git")
	assert.Equal(t, "https", protocol)
}

func TestRepository_DetermineProtocol_SSH(t *testing.T) {
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

	protocol := repo.DetermineProtocol("git@github.com:lerenn/example.git")
	assert.Equal(t, "ssh", protocol)
}

func TestRepository_ExtractRepoNameFromFullPath_Success(t *testing.T) {
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

	repoName := repo.ExtractRepoNameFromFullPath("github.com/octocat/Hello-World")
	assert.Equal(t, "Hello-World", repoName)
}

func TestRepository_ExtractRepoNameFromFullPath_SinglePart(t *testing.T) {
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

	repoName := repo.ExtractRepoNameFromFullPath("example")
	assert.Equal(t, "example", repoName)
}

func TestRepository_ConstructRemoteURL_HTTPS(t *testing.T) {
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

	remoteURL, err := repo.ConstructRemoteURL("https://github.com/octocat/Hello-World.git", "upstream", "github.com/octocat/Hello-World")
	assert.NoError(t, err)
	assert.Equal(t, "https://github.com/upstream/Hello-World.git", remoteURL)
}

func TestRepository_ConstructRemoteURL_SSH(t *testing.T) {
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

	remoteURL, err := repo.ConstructRemoteURL("git@github.com:lerenn/example.git", "upstream", "github.com/octocat/Hello-World")
	assert.NoError(t, err)
	assert.Equal(t, "git@github.com:upstream/Hello-World.git", remoteURL)
}

func TestRepository_ConstructRemoteURL_InvalidHost(t *testing.T) {
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

	remoteURL, err := repo.ConstructRemoteURL("invalid-url", "upstream", "github.com/octocat/Hello-World")
	assert.Error(t, err)
	assert.Empty(t, remoteURL)
}
