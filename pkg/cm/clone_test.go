//go:build unit

package cm

import (
	"testing"

	fsmocks "github.com/lerenn/code-manager/pkg/fs/mocks"
	"github.com/lerenn/code-manager/pkg/git"
	gitmocks "github.com/lerenn/code-manager/pkg/git/mocks"
	"github.com/lerenn/code-manager/pkg/mode/repository"
	repositorymocks "github.com/lerenn/code-manager/pkg/mode/repository/mocks"
	"github.com/lerenn/code-manager/pkg/mode/workspace"
	workspacemocks "github.com/lerenn/code-manager/pkg/mode/workspace/mocks"
	promptmocks "github.com/lerenn/code-manager/pkg/prompt/mocks"
	"github.com/lerenn/code-manager/pkg/status"
	statusmocks "github.com/lerenn/code-manager/pkg/status/mocks"
	"github.com/stretchr/testify/assert"
	"go.uber.org/mock/gomock"
)

func TestRealCM_Clone_Success(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockFS := fsmocks.NewMockFS(ctrl)
	mockGit := gitmocks.NewMockGit(ctrl)
	mockStatus := statusmocks.NewMockManager(ctrl)
	mockRepository := repositorymocks.NewMockRepository(ctrl)
	mockWorkspace := workspacemocks.NewMockWorkspace(ctrl)

	cm, err := NewCM(NewCMParams{
		RepositoryProvider: func(params repository.NewRepositoryParams) repository.Repository {
			return mockRepository
		},
		WorkspaceProvider: func(params workspace.NewWorkspaceParams) workspace.Workspace {
			return mockWorkspace
		},
		Config: createTestConfig(),
		FS:     mockFS,
		Git:    mockGit,
		Status: mockStatus,
	})
	assert.NoError(t, err)

	repoURL := "https://github.com/octocat/Hello-World.git"
	normalizedURL := "github.com/octocat/Hello-World"
	defaultBranch := "main"
	targetPath := "/test/base/path/github.com/octocat/Hello-World/origin/main"

	// Mock repository existence check
	mockStatus.EXPECT().ListRepositories().Return(map[string]status.Repository{}, nil)

	// Mock default branch detection
	mockGit.EXPECT().GetDefaultBranch(repoURL).Return(defaultBranch, nil)

	// Mock directory creation
	mockFS.EXPECT().MkdirAll("/test/base/path/github.com/octocat/Hello-World/origin", gomock.Any()).Return(nil)

	// Mock clone operation
	mockGit.EXPECT().Clone(git.CloneParams{
		RepoURL:    repoURL,
		TargetPath: targetPath,
		Recursive:  true,
	}).Return(nil)

	// Mock repository initialization
	mockStatus.EXPECT().AddRepository(normalizedURL, status.AddRepositoryParams{
		Path: targetPath,
		Remotes: map[string]status.Remote{
			"origin": {
				DefaultBranch: defaultBranch,
			},
		},
	}).Return(nil)

	err = cm.Clone(repoURL)
	assert.NoError(t, err)
}

func TestRealCM_Clone_ShallowSuccess(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockFS := fsmocks.NewMockFS(ctrl)
	mockGit := gitmocks.NewMockGit(ctrl)
	mockStatus := statusmocks.NewMockManager(ctrl)
	mockRepository := repositorymocks.NewMockRepository(ctrl)
	mockWorkspace := workspacemocks.NewMockWorkspace(ctrl)
	mockPrompt := promptmocks.NewMockPrompter(ctrl)

	cm, err := NewCM(NewCMParams{
		RepositoryProvider: func(params repository.NewRepositoryParams) repository.Repository {
			return mockRepository
		},
		WorkspaceProvider: func(params workspace.NewWorkspaceParams) workspace.Workspace {
			return mockWorkspace
		},
		Config: createTestConfig(),
		FS:     mockFS,
		Git:    mockGit,
		Status: mockStatus,
		Prompt: mockPrompt,
	})
	assert.NoError(t, err)

	repoURL := "https://github.com/octocat/Hello-World.git"
	normalizedURL := "github.com/octocat/Hello-World"
	defaultBranch := "main"
	targetPath := "/test/base/path/github.com/octocat/Hello-World/origin/main"

	// Mock repository existence check
	mockStatus.EXPECT().ListRepositories().Return(map[string]status.Repository{}, nil)

	// Mock default branch detection
	mockGit.EXPECT().GetDefaultBranch(repoURL).Return(defaultBranch, nil)

	// Mock directory creation
	mockFS.EXPECT().MkdirAll("/test/base/path/github.com/octocat/Hello-World/origin", gomock.Any()).Return(nil)

	// Mock clone operation (shallow)
	mockGit.EXPECT().Clone(git.CloneParams{
		RepoURL:    repoURL,
		TargetPath: targetPath,
		Recursive:  false,
	}).Return(nil)

	// Mock repository initialization
	mockStatus.EXPECT().AddRepository(normalizedURL, status.AddRepositoryParams{
		Path: targetPath,
		Remotes: map[string]status.Remote{
			"origin": {
				DefaultBranch: defaultBranch,
			},
		},
	}).Return(nil)

	opts := CloneOpts{Recursive: false}
	err = cm.Clone(repoURL, opts)
	assert.NoError(t, err)
}

func TestRealCM_Clone_EmptyURL(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockFS := fsmocks.NewMockFS(ctrl)
	mockGit := gitmocks.NewMockGit(ctrl)
	mockStatus := statusmocks.NewMockManager(ctrl)
	mockRepository := repositorymocks.NewMockRepository(ctrl)
	mockWorkspace := workspacemocks.NewMockWorkspace(ctrl)
	mockPrompt := promptmocks.NewMockPrompter(ctrl)

	cm, err := NewCM(NewCMParams{
		RepositoryProvider: func(params repository.NewRepositoryParams) repository.Repository {
			return mockRepository
		},
		WorkspaceProvider: func(params workspace.NewWorkspaceParams) workspace.Workspace {
			return mockWorkspace
		},
		Config: createTestConfig(),
		FS:     mockFS,
		Git:    mockGit,
		Status: mockStatus,

		Prompt: mockPrompt,
	})
	assert.NoError(t, err)

	err = cm.Clone("")
	assert.Error(t, err)
	assert.ErrorIs(t, err, ErrRepositoryURLEmpty)
}

func TestRealCM_Clone_RepositoryExists(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockFS := fsmocks.NewMockFS(ctrl)
	mockGit := gitmocks.NewMockGit(ctrl)
	mockStatus := statusmocks.NewMockManager(ctrl)
	mockRepository := repositorymocks.NewMockRepository(ctrl)
	mockWorkspace := workspacemocks.NewMockWorkspace(ctrl)
	mockPrompt := promptmocks.NewMockPrompter(ctrl)

	cm, err := NewCM(NewCMParams{
		RepositoryProvider: func(params repository.NewRepositoryParams) repository.Repository {
			return mockRepository
		},
		WorkspaceProvider: func(params workspace.NewWorkspaceParams) workspace.Workspace {
			return mockWorkspace
		},
		Config: createTestConfig(),
		FS:     mockFS,
		Git:    mockGit,
		Status: mockStatus,

		Prompt: mockPrompt,
	})
	assert.NoError(t, err)

	repoURL := "https://github.com/octocat/Hello-World.git"
	normalizedURL := "github.com/octocat/Hello-World"

	// Mock repository existence check - repository already exists
	existingRepos := map[string]status.Repository{
		normalizedURL: {
			Path: "/existing/path",
		},
	}
	mockStatus.EXPECT().ListRepositories().Return(existingRepos, nil)

	err = cm.Clone(repoURL)
	assert.Error(t, err)
	assert.ErrorIs(t, err, ErrRepositoryExists)
}

func TestRealCM_Clone_DefaultBranchDetectionFailure(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockFS := fsmocks.NewMockFS(ctrl)
	mockGit := gitmocks.NewMockGit(ctrl)
	mockStatus := statusmocks.NewMockManager(ctrl)
	mockRepository := repositorymocks.NewMockRepository(ctrl)
	mockWorkspace := workspacemocks.NewMockWorkspace(ctrl)
	mockPrompt := promptmocks.NewMockPrompter(ctrl)

	cm, err := NewCM(NewCMParams{
		RepositoryProvider: func(params repository.NewRepositoryParams) repository.Repository {
			return mockRepository
		},
		WorkspaceProvider: func(params workspace.NewWorkspaceParams) workspace.Workspace {
			return mockWorkspace
		},
		Config: createTestConfig(),
		FS:     mockFS,
		Git:    mockGit,
		Status: mockStatus,

		Prompt: mockPrompt,
	})
	assert.NoError(t, err)

	repoURL := "https://github.com/octocat/Hello-World.git"

	// Mock repository existence check
	mockStatus.EXPECT().ListRepositories().Return(map[string]status.Repository{}, nil)

	// Mock default branch detection failure
	mockGit.EXPECT().GetDefaultBranch(repoURL).Return("", assert.AnError)

	err = cm.Clone(repoURL)
	assert.Error(t, err)
	assert.ErrorIs(t, err, ErrFailedToDetectDefaultBranch)
}

func TestRealCM_Clone_CloneFailure(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockFS := fsmocks.NewMockFS(ctrl)
	mockGit := gitmocks.NewMockGit(ctrl)
	mockStatus := statusmocks.NewMockManager(ctrl)
	mockRepository := repositorymocks.NewMockRepository(ctrl)
	mockWorkspace := workspacemocks.NewMockWorkspace(ctrl)
	mockPrompt := promptmocks.NewMockPrompter(ctrl)

	cm, err := NewCM(NewCMParams{
		RepositoryProvider: func(params repository.NewRepositoryParams) repository.Repository {
			return mockRepository
		},
		WorkspaceProvider: func(params workspace.NewWorkspaceParams) workspace.Workspace {
			return mockWorkspace
		},
		Config: createTestConfig(),
		FS:     mockFS,
		Git:    mockGit,
		Status: mockStatus,

		Prompt: mockPrompt,
	})
	assert.NoError(t, err)

	repoURL := "https://github.com/octocat/Hello-World.git"
	defaultBranch := "main"

	// Mock repository existence check
	mockStatus.EXPECT().ListRepositories().Return(map[string]status.Repository{}, nil)

	// Mock default branch detection
	mockGit.EXPECT().GetDefaultBranch(repoURL).Return(defaultBranch, nil)

	// Mock directory creation
	mockFS.EXPECT().MkdirAll("/test/base/path/github.com/octocat/Hello-World/origin", gomock.Any()).Return(nil)

	// Mock clone operation failure
	mockGit.EXPECT().Clone(gomock.Any()).Return(assert.AnError)

	err = cm.Clone(repoURL)
	assert.Error(t, err)
	assert.ErrorIs(t, err, ErrFailedToCloneRepository)
}

func TestRealCM_Clone_InitializationFailure(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockFS := fsmocks.NewMockFS(ctrl)
	mockGit := gitmocks.NewMockGit(ctrl)
	mockStatus := statusmocks.NewMockManager(ctrl)
	mockRepository := repositorymocks.NewMockRepository(ctrl)
	mockWorkspace := workspacemocks.NewMockWorkspace(ctrl)
	mockPrompt := promptmocks.NewMockPrompter(ctrl)

	cm, err := NewCM(NewCMParams{
		RepositoryProvider: func(params repository.NewRepositoryParams) repository.Repository {
			return mockRepository
		},
		WorkspaceProvider: func(params workspace.NewWorkspaceParams) workspace.Workspace {
			return mockWorkspace
		},
		Config: createTestConfig(),
		FS:     mockFS,
		Git:    mockGit,
		Status: mockStatus,

		Prompt: mockPrompt,
	})
	assert.NoError(t, err)

	repoURL := "https://github.com/octocat/Hello-World.git"
	defaultBranch := "main"

	// Mock repository existence check
	mockStatus.EXPECT().ListRepositories().Return(map[string]status.Repository{}, nil)

	// Mock default branch detection
	mockGit.EXPECT().GetDefaultBranch(repoURL).Return(defaultBranch, nil)

	// Mock directory creation
	mockFS.EXPECT().MkdirAll("/test/base/path/github.com/octocat/Hello-World/origin", gomock.Any()).Return(nil)

	// Mock clone operation success
	mockGit.EXPECT().Clone(gomock.Any()).Return(nil)

	// Mock repository initialization failure
	mockStatus.EXPECT().AddRepository(gomock.Any(), gomock.Any()).Return(assert.AnError)

	err = cm.Clone(repoURL)
	assert.Error(t, err)
	assert.ErrorIs(t, err, ErrFailedToInitializeRepository)
}
