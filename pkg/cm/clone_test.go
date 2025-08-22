//go:build unit

package cm

import (
	"testing"

	basepkg "github.com/lerenn/code-manager/internal/base"
	"github.com/lerenn/code-manager/pkg/fs"
	"github.com/lerenn/code-manager/pkg/git"
	"github.com/lerenn/code-manager/pkg/ide"
	"github.com/lerenn/code-manager/pkg/logger"
	"github.com/lerenn/code-manager/pkg/prompt"
	"github.com/lerenn/code-manager/pkg/status"
	"github.com/stretchr/testify/assert"
	"go.uber.org/mock/gomock"
)

func TestRealCM_Clone_Success(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockFS := fs.NewMockFS(ctrl)
	mockGit := git.NewMockGit(ctrl)
	mockStatus := status.NewMockManager(ctrl)
	mockLogger := logger.NewNoopLogger()
	mockPrompt := prompt.NewMockPrompt(ctrl)

	cm := &realCM{
		Base: basepkg.NewBase(basepkg.NewBaseParams{
			FS:            mockFS,
			Git:           mockGit,
			Config:        createTestConfig(),
			StatusManager: mockStatus,
			Logger:        mockLogger,
			Prompt:        mockPrompt,
			Verbose:       false,
		}),
		ideManager: ide.NewManager(mockFS, mockLogger),
	}

	repoURL := "https://github.com/octocat/Hello-World.git"
	normalizedURL := "github.com/octocat/Hello-World"
	defaultBranch := "main"
	targetPath := "/test/base/path/github.com/octocat/Hello-World/origin/main"

	// Mock repository existence check
	mockStatus.EXPECT().ListRepositories().Return(map[string]status.Repository{}, nil)

	// Mock default branch detection
	mockGit.EXPECT().GetDefaultBranch(repoURL).Return(defaultBranch, nil)

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

	err := cm.Clone(repoURL)
	assert.NoError(t, err)
}

func TestRealCM_Clone_ShallowSuccess(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockFS := fs.NewMockFS(ctrl)
	mockGit := git.NewMockGit(ctrl)
	mockStatus := status.NewMockManager(ctrl)
	mockLogger := logger.NewNoopLogger()
	mockPrompt := prompt.NewMockPrompt(ctrl)

	cm := &realCM{
		Base: basepkg.NewBase(basepkg.NewBaseParams{
			FS:            mockFS,
			Git:           mockGit,
			Config:        createTestConfig(),
			StatusManager: mockStatus,
			Logger:        mockLogger,
			Prompt:        mockPrompt,
			Verbose:       false,
		}),
		ideManager: ide.NewManager(mockFS, mockLogger),
	}

	repoURL := "https://github.com/octocat/Hello-World.git"
	normalizedURL := "github.com/octocat/Hello-World"
	defaultBranch := "main"
	targetPath := "/test/base/path/github.com/octocat/Hello-World/origin/main"

	// Mock repository existence check
	mockStatus.EXPECT().ListRepositories().Return(map[string]status.Repository{}, nil)

	// Mock default branch detection
	mockGit.EXPECT().GetDefaultBranch(repoURL).Return(defaultBranch, nil)

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
	err := cm.Clone(repoURL, opts)
	assert.NoError(t, err)
}

func TestRealCM_Clone_EmptyURL(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockFS := fs.NewMockFS(ctrl)
	mockGit := git.NewMockGit(ctrl)
	mockStatus := status.NewMockManager(ctrl)
	mockLogger := logger.NewNoopLogger()
	mockPrompt := prompt.NewMockPrompt(ctrl)

	cm := &realCM{
		Base: basepkg.NewBase(basepkg.NewBaseParams{
			FS:            mockFS,
			Git:           mockGit,
			Config:        createTestConfig(),
			StatusManager: mockStatus,
			Logger:        mockLogger,
			Prompt:        mockPrompt,
			Verbose:       false,
		}),
		ideManager: ide.NewManager(mockFS, mockLogger),
	}

	err := cm.Clone("")
	assert.Error(t, err)
	assert.ErrorIs(t, err, ErrRepositoryURLEmpty)
}

func TestRealCM_Clone_RepositoryExists(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockFS := fs.NewMockFS(ctrl)
	mockGit := git.NewMockGit(ctrl)
	mockStatus := status.NewMockManager(ctrl)
	mockLogger := logger.NewNoopLogger()
	mockPrompt := prompt.NewMockPrompt(ctrl)

	cm := &realCM{
		Base: basepkg.NewBase(basepkg.NewBaseParams{
			FS:            mockFS,
			Git:           mockGit,
			Config:        createTestConfig(),
			StatusManager: mockStatus,
			Logger:        mockLogger,
			Prompt:        mockPrompt,
			Verbose:       false,
		}),
		ideManager: ide.NewManager(mockFS, mockLogger),
	}

	repoURL := "https://github.com/octocat/Hello-World.git"
	normalizedURL := "github.com/octocat/Hello-World"

	// Mock repository existence check - repository already exists
	existingRepos := map[string]status.Repository{
		normalizedURL: {
			Path: "/existing/path",
		},
	}
	mockStatus.EXPECT().ListRepositories().Return(existingRepos, nil)

	err := cm.Clone(repoURL)
	assert.Error(t, err)
	assert.ErrorIs(t, err, ErrRepositoryExists)
}

func TestRealCM_Clone_DefaultBranchDetectionFailure(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockFS := fs.NewMockFS(ctrl)
	mockGit := git.NewMockGit(ctrl)
	mockStatus := status.NewMockManager(ctrl)
	mockLogger := logger.NewNoopLogger()
	mockPrompt := prompt.NewMockPrompt(ctrl)

	cm := &realCM{
		Base: basepkg.NewBase(basepkg.NewBaseParams{
			FS:            mockFS,
			Git:           mockGit,
			Config:        createTestConfig(),
			StatusManager: mockStatus,
			Logger:        mockLogger,
			Prompt:        mockPrompt,
			Verbose:       false,
		}),
		ideManager: ide.NewManager(mockFS, mockLogger),
	}

	repoURL := "https://github.com/octocat/Hello-World.git"

	// Mock repository existence check
	mockStatus.EXPECT().ListRepositories().Return(map[string]status.Repository{}, nil)

	// Mock default branch detection failure
	mockGit.EXPECT().GetDefaultBranch(repoURL).Return("", assert.AnError)

	err := cm.Clone(repoURL)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "failed to detect default branch")
}

func TestRealCM_Clone_CloneFailure(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockFS := fs.NewMockFS(ctrl)
	mockGit := git.NewMockGit(ctrl)
	mockStatus := status.NewMockManager(ctrl)
	mockLogger := logger.NewNoopLogger()
	mockPrompt := prompt.NewMockPrompt(ctrl)

	cm := &realCM{
		Base: basepkg.NewBase(basepkg.NewBaseParams{
			FS:            mockFS,
			Git:           mockGit,
			Config:        createTestConfig(),
			StatusManager: mockStatus,
			Logger:        mockLogger,
			Prompt:        mockPrompt,
			Verbose:       false,
		}),
		ideManager: ide.NewManager(mockFS, mockLogger),
	}

	repoURL := "https://github.com/octocat/Hello-World.git"
	defaultBranch := "main"

	// Mock repository existence check
	mockStatus.EXPECT().ListRepositories().Return(map[string]status.Repository{}, nil)

	// Mock default branch detection
	mockGit.EXPECT().GetDefaultBranch(repoURL).Return(defaultBranch, nil)

	// Mock clone operation failure
	mockGit.EXPECT().Clone(gomock.Any()).Return(assert.AnError)

	err := cm.Clone(repoURL)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "failed to clone repository")
}

func TestRealCM_Clone_InitializationFailure(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockFS := fs.NewMockFS(ctrl)
	mockGit := git.NewMockGit(ctrl)
	mockStatus := status.NewMockManager(ctrl)
	mockLogger := logger.NewNoopLogger()
	mockPrompt := prompt.NewMockPrompt(ctrl)

	cm := &realCM{
		Base: basepkg.NewBase(basepkg.NewBaseParams{
			FS:            mockFS,
			Git:           mockGit,
			Config:        createTestConfig(),
			StatusManager: mockStatus,
			Logger:        mockLogger,
			Prompt:        mockPrompt,
			Verbose:       false,
		}),
		ideManager: ide.NewManager(mockFS, mockLogger),
	}

	repoURL := "https://github.com/octocat/Hello-World.git"
	defaultBranch := "main"

	// Mock repository existence check
	mockStatus.EXPECT().ListRepositories().Return(map[string]status.Repository{}, nil)

	// Mock default branch detection
	mockGit.EXPECT().GetDefaultBranch(repoURL).Return(defaultBranch, nil)

	// Mock clone operation success
	mockGit.EXPECT().Clone(gomock.Any()).Return(nil)

	// Mock repository initialization failure
	mockStatus.EXPECT().AddRepository(gomock.Any(), gomock.Any()).Return(assert.AnError)

	err := cm.Clone(repoURL)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "failed to initialize repository in CM")
}

func TestRealCM_NormalizeRepositoryURL_HTTPS(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockFS := fs.NewMockFS(ctrl)
	mockGit := git.NewMockGit(ctrl)
	mockStatus := status.NewMockManager(ctrl)
	mockLogger := logger.NewNoopLogger()
	mockPrompt := prompt.NewMockPrompt(ctrl)

	cm := &realCM{
		Base: basepkg.NewBase(basepkg.NewBaseParams{
			FS:            mockFS,
			Git:           mockGit,
			Config:        createTestConfig(),
			StatusManager: mockStatus,
			Logger:        mockLogger,
			Prompt:        mockPrompt,
			Verbose:       false,
		}),
		ideManager: ide.NewManager(mockFS, mockLogger),
	}

	// Test HTTPS URL with .git suffix
	result, err := cm.normalizeRepositoryURL("https://github.com/octocat/Hello-World.git")
	assert.NoError(t, err)
	assert.Equal(t, "github.com/octocat/Hello-World", result)

	// Test HTTPS URL without .git suffix
	result, err = cm.normalizeRepositoryURL("https://github.com/octocat/Hello-World")
	assert.NoError(t, err)
	assert.Equal(t, "github.com/octocat/Hello-World", result)
}

func TestRealCM_NormalizeRepositoryURL_SSH(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockFS := fs.NewMockFS(ctrl)
	mockGit := git.NewMockGit(ctrl)
	mockStatus := status.NewMockManager(ctrl)
	mockLogger := logger.NewNoopLogger()
	mockPrompt := prompt.NewMockPrompt(ctrl)

	cm := &realCM{
		Base: basepkg.NewBase(basepkg.NewBaseParams{
			FS:            mockFS,
			Git:           mockGit,
			Config:        createTestConfig(),
			StatusManager: mockStatus,
			Logger:        mockLogger,
			Prompt:        mockPrompt,
			Verbose:       false,
		}),
		ideManager: ide.NewManager(mockFS, mockLogger),
	}

	// Test SSH URL with .git suffix
	result, err := cm.normalizeRepositoryURL("git@github.com:lerenn/example.git")
	assert.NoError(t, err)
	assert.Equal(t, "github.com/octocat/Hello-World", result)

	// Test SSH URL without .git suffix
	result, err = cm.normalizeRepositoryURL("git@github.com:lerenn/example")
	assert.NoError(t, err)
	assert.Equal(t, "github.com/octocat/Hello-World", result)
}

func TestRealCM_NormalizeRepositoryURL_InvalidURL(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockFS := fs.NewMockFS(ctrl)
	mockGit := git.NewMockGit(ctrl)
	mockStatus := status.NewMockManager(ctrl)
	mockLogger := logger.NewNoopLogger()
	mockPrompt := prompt.NewMockPrompt(ctrl)

	cm := &realCM{
		Base: basepkg.NewBase(basepkg.NewBaseParams{
			FS:            mockFS,
			Git:           mockGit,
			Config:        createTestConfig(),
			StatusManager: mockStatus,
			Logger:        mockLogger,
			Prompt:        mockPrompt,
			Verbose:       false,
		}),
		ideManager: ide.NewManager(mockFS, mockLogger),
	}

	// Test invalid URL
	_, err := cm.normalizeRepositoryURL("not-a-valid-url")
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "unsupported repository URL format")
}

func TestRealCM_GenerateClonePath(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockFS := fs.NewMockFS(ctrl)
	mockGit := git.NewMockGit(ctrl)
	mockStatus := status.NewMockManager(ctrl)
	mockLogger := logger.NewNoopLogger()
	mockPrompt := prompt.NewMockPrompt(ctrl)

	cm := &realCM{
		Base: basepkg.NewBase(basepkg.NewBaseParams{
			FS:            mockFS,
			Git:           mockGit,
			Config:        createTestConfig(),
			StatusManager: mockStatus,
			Logger:        mockLogger,
			Prompt:        mockPrompt,
			Verbose:       false,
		}),
		ideManager: ide.NewManager(mockFS, mockLogger),
	}

	normalizedURL := "github.com/octocat/Hello-World"
	defaultBranch := "main"

	result := cm.generateClonePath(normalizedURL, defaultBranch)
	expected := "/test/base/path/github.com/octocat/Hello-World/origin/main"
	assert.Equal(t, expected, result)
}
