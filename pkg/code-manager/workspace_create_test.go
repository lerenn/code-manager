//go:build unit

package codemanager

import (
	"errors"
	"testing"

	"github.com/lerenn/code-manager/pkg/config"
	"github.com/lerenn/code-manager/pkg/dependencies"
	fsmocks "github.com/lerenn/code-manager/pkg/fs/mocks"
	gitmocks "github.com/lerenn/code-manager/pkg/git/mocks"
	"github.com/lerenn/code-manager/pkg/logger"
	"github.com/lerenn/code-manager/pkg/status"
	statusmocks "github.com/lerenn/code-manager/pkg/status/mocks"
	"github.com/stretchr/testify/assert"
	"go.uber.org/mock/gomock"
)

// TestCreateWorkspace_Success tests successful workspace creation.
func TestCreateWorkspace_Success(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockFS := fsmocks.NewMockFS(ctrl)
	mockGit := gitmocks.NewMockGit(ctrl)
	mockStatus := statusmocks.NewMockManager(ctrl)

	cm := &realCodeManager{
		deps: dependencies.New().
			WithFS(mockFS).
			WithGit(mockGit).
			WithConfig(config.NewConfigManager("/test/config.yaml")).
			WithStatusManager(mockStatus).
			WithLogger(logger.NewNoopLogger()),
	}

	params := CreateWorkspaceParams{
		WorkspaceName: "test-workspace",
		Repositories:  []string{"repo1", "/absolute/path/repo2"},
	}

	// Mock workspace doesn't exist
	mockStatus.EXPECT().GetWorkspace("test-workspace").Return(nil, errors.New("not found"))

	// Mock repository resolution
	existingRepo := &status.Repository{Path: "/path/to/repo1"}
	mockStatus.EXPECT().GetRepository("repo1").Return(existingRepo, nil)

	// Mock absolute path validation
	mockFS.EXPECT().Exists("/absolute/path/repo2").Return(true, nil)
	mockFS.EXPECT().ValidateRepositoryPath("/absolute/path/repo2").Return(true, nil)

	// Mock repository not found in status (first call during resolution)
	mockStatus.EXPECT().GetRepository("/absolute/path/repo2").Return(nil, errors.New("not found"))

	// In addRepositoriesToStatus, it loops through repos and calls GetRemoteURL for each first
	// Then for each repo, it checks if it exists in status
	// For repo1: exists, so skips
	// For repo2: doesn't exist, so calls addRepositoryToStatus

	// First loop: GetRemoteURL for all repos
	mockGit.EXPECT().GetRemoteURL("repo1", "origin").Return("https://github.com/user/repo1.git", nil)
	mockGit.EXPECT().GetRemoteURL("/absolute/path/repo2", "origin").Return("https://github.com/user/repo2.git", nil)

	// Check repo1 in status using original remote URL (exists, so skip)
	mockStatus.EXPECT().GetRepository("https://github.com/user/repo1.git").Return(existingRepo, nil)

	// Check repo2 in status using original remote URL (doesn't exist, so proceed to addRepositoryToStatus)
	mockStatus.EXPECT().GetRepository("https://github.com/user/repo2.git").Return(nil, errors.New("not found"))

	// Now addRepositoryToStatus is called for repo2:
	// 1. GetRemoteURL (again)
	mockGit.EXPECT().GetRemoteURL("/absolute/path/repo2", "origin").Return("https://github.com/user/repo2.git", nil)

	// 2. Normalize URL (happens internally, no mock needed)

	// 3. Check if normalized URL exists in status
	mockStatus.EXPECT().GetRepository("github.com/user/repo2").Return(nil, errors.New("not found"))

	// 4. GetDefaultBranch
	mockGit.EXPECT().GetDefaultBranch("https://github.com/user/repo2.git").Return("main", nil)

	// 5. MkdirAll
	mockFS.EXPECT().MkdirAll(gomock.Any(), gomock.Any()).Return(nil)

	// 6. Clone
	mockGit.EXPECT().Clone(gomock.Any()).Return(nil)

	// 7. AddRepository
	mockStatus.EXPECT().AddRepository("github.com/user/repo2", gomock.Any()).Return(nil)

	// Mock adding workspace to status
	mockStatus.EXPECT().AddWorkspace("test-workspace", gomock.Any()).Return(nil)

	err := cm.CreateWorkspace(params)
	assert.NoError(t, err)
}

// TestCreateWorkspace_InvalidWorkspaceName tests workspace creation with invalid name.
func TestCreateWorkspace_InvalidWorkspaceName(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockFS := fsmocks.NewMockFS(ctrl)
	mockGit := gitmocks.NewMockGit(ctrl)
	mockStatus := statusmocks.NewMockManager(ctrl)

	cm := &realCodeManager{
		deps: dependencies.New().
			WithFS(mockFS).
			WithGit(mockGit).
			WithConfig(config.NewConfigManager("/test/config.yaml")).
			WithStatusManager(mockStatus).
			WithLogger(logger.NewNoopLogger()),
	}

	params := CreateWorkspaceParams{
		WorkspaceName: "", // Invalid empty name
		Repositories:  []string{"repo1"},
	}

	err := cm.CreateWorkspace(params)
	assert.Error(t, err)
	assert.ErrorIs(t, err, ErrInvalidWorkspaceName)
}

// TestCreateWorkspace_WorkspaceAlreadyExists tests workspace creation when workspace already exists.
func TestCreateWorkspace_WorkspaceAlreadyExists(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockFS := fsmocks.NewMockFS(ctrl)
	mockGit := gitmocks.NewMockGit(ctrl)
	mockStatus := statusmocks.NewMockManager(ctrl)

	cm := &realCodeManager{
		deps: dependencies.New().
			WithFS(mockFS).
			WithGit(mockGit).
			WithConfig(config.NewConfigManager("/test/config.yaml")).
			WithStatusManager(mockStatus).
			WithLogger(logger.NewNoopLogger()),
	}

	params := CreateWorkspaceParams{
		WorkspaceName: "existing-workspace",
		Repositories:  []string{"repo1"},
	}

	// Mock workspace already exists
	existingWorkspace := &status.Workspace{Repositories: []string{"repo1"}}
	mockStatus.EXPECT().GetWorkspace("existing-workspace").Return(existingWorkspace, nil)

	err := cm.CreateWorkspace(params)
	assert.Error(t, err)
	assert.ErrorIs(t, err, ErrWorkspaceAlreadyExists)
}

// TestCreateWorkspace_EmptyRepositories tests workspace creation with no repositories.
func TestCreateWorkspace_EmptyRepositories(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockFS := fsmocks.NewMockFS(ctrl)
	mockGit := gitmocks.NewMockGit(ctrl)
	mockStatus := statusmocks.NewMockManager(ctrl)

	cm := &realCodeManager{
		deps: dependencies.New().
			WithFS(mockFS).
			WithGit(mockGit).
			WithConfig(config.NewConfigManager("/test/config.yaml")).
			WithStatusManager(mockStatus).
			WithLogger(logger.NewNoopLogger()),
	}

	params := CreateWorkspaceParams{
		WorkspaceName: "test-workspace",
		Repositories:  []string{}, // Empty repositories
	}

	// Mock workspace doesn't exist
	mockStatus.EXPECT().GetWorkspace("test-workspace").Return(nil, errors.New("not found"))

	err := cm.CreateWorkspace(params)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "at least one repository must be specified")
}

// TestCreateWorkspace_DuplicateRepositories tests workspace creation with duplicate repositories.
func TestCreateWorkspace_DuplicateRepositories(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockFS := fsmocks.NewMockFS(ctrl)
	mockGit := gitmocks.NewMockGit(ctrl)
	mockStatus := statusmocks.NewMockManager(ctrl)

	cm := &realCodeManager{
		deps: dependencies.New().
			WithFS(mockFS).
			WithGit(mockGit).
			WithConfig(config.NewConfigManager("/test/config.yaml")).
			WithStatusManager(mockStatus).
			WithLogger(logger.NewNoopLogger()),
	}

	params := CreateWorkspaceParams{
		WorkspaceName: "test-workspace",
		Repositories:  []string{"repo1", "repo1"}, // Duplicate repositories
	}

	// Mock workspace doesn't exist
	mockStatus.EXPECT().GetWorkspace("test-workspace").Return(nil, errors.New("not found"))

	// Mock repository not found in status (first call)
	mockStatus.EXPECT().GetRepository("repo1").Return(nil, errors.New("not found"))

	// Mock path resolution for the second repo1 (treated as relative path)
	mockFS.EXPECT().ResolvePath(gomock.Any(), "repo1").Return("/current/dir/repo1", nil)

	// Mock path validation for the resolved path
	mockFS.EXPECT().Exists("/current/dir/repo1").Return(true, nil)
	mockFS.EXPECT().ValidateRepositoryPath("/current/dir/repo1").Return(true, nil)

	err := cm.CreateWorkspace(params)
	assert.Error(t, err)
	assert.ErrorIs(t, err, ErrDuplicateRepository)
}

// TestCreateWorkspace_RepositoryNotFound tests workspace creation with non-existent repository.
func TestCreateWorkspace_RepositoryNotFound(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockFS := fsmocks.NewMockFS(ctrl)
	mockGit := gitmocks.NewMockGit(ctrl)
	mockStatus := statusmocks.NewMockManager(ctrl)

	cm := &realCodeManager{
		deps: dependencies.New().
			WithFS(mockFS).
			WithGit(mockGit).
			WithConfig(config.NewConfigManager("/test/config.yaml")).
			WithStatusManager(mockStatus).
			WithLogger(logger.NewNoopLogger()),
	}

	params := CreateWorkspaceParams{
		WorkspaceName: "test-workspace",
		Repositories:  []string{"/non/existent/path"},
	}

	// Mock workspace doesn't exist
	mockStatus.EXPECT().GetWorkspace("test-workspace").Return(nil, errors.New("not found"))

	// Mock repository not found in status
	mockStatus.EXPECT().GetRepository("/non/existent/path").Return(nil, errors.New("not found"))

	// Mock path doesn't exist
	mockFS.EXPECT().Exists("/non/existent/path").Return(false, nil)

	err := cm.CreateWorkspace(params)
	assert.Error(t, err)
	assert.ErrorIs(t, err, ErrRepositoryNotFound)
}

// TestCreateWorkspace_InvalidRepository tests workspace creation with invalid repository.
func TestCreateWorkspace_InvalidRepository(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockFS := fsmocks.NewMockFS(ctrl)
	mockGit := gitmocks.NewMockGit(ctrl)
	mockStatus := statusmocks.NewMockManager(ctrl)

	cm := &realCodeManager{
		deps: dependencies.New().
			WithFS(mockFS).
			WithGit(mockGit).
			WithConfig(config.NewConfigManager("/test/config.yaml")).
			WithStatusManager(mockStatus).
			WithLogger(logger.NewNoopLogger()),
	}

	params := CreateWorkspaceParams{
		WorkspaceName: "test-workspace",
		Repositories:  []string{"/not/a/git/repo"},
	}

	// Mock workspace doesn't exist
	mockStatus.EXPECT().GetWorkspace("test-workspace").Return(nil, errors.New("not found"))

	// Mock repository not found in status
	mockStatus.EXPECT().GetRepository("/not/a/git/repo").Return(nil, errors.New("not found"))

	// Mock path exists but is not a Git repository
	mockFS.EXPECT().Exists("/not/a/git/repo").Return(true, nil)
	mockFS.EXPECT().ValidateRepositoryPath("/not/a/git/repo").Return(false, nil)

	err := cm.CreateWorkspace(params)
	assert.Error(t, err)
	assert.ErrorIs(t, err, ErrInvalidRepository)
}

// TestCreateWorkspace_RelativePathResolution tests workspace creation with relative paths.
func TestCreateWorkspace_RelativePathResolution(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockFS := fsmocks.NewMockFS(ctrl)
	mockGit := gitmocks.NewMockGit(ctrl)
	mockStatus := statusmocks.NewMockManager(ctrl)

	cm := &realCodeManager{
		deps: dependencies.New().
			WithFS(mockFS).
			WithGit(mockGit).
			WithConfig(config.NewConfigManager("/test/config.yaml")).
			WithStatusManager(mockStatus).
			WithLogger(logger.NewNoopLogger()),
	}

	params := CreateWorkspaceParams{
		WorkspaceName: "test-workspace",
		Repositories:  []string{"./relative/repo"},
	}

	// Mock workspace doesn't exist
	mockStatus.EXPECT().GetWorkspace("test-workspace").Return(nil, errors.New("not found"))

	// Mock repository not found in status
	mockStatus.EXPECT().GetRepository("./relative/repo").Return(nil, errors.New("not found"))

	// Mock path resolution - use gomock.Any() for the first argument since os.Getwd() returns actual current dir
	mockFS.EXPECT().ResolvePath(gomock.Any(), "./relative/repo").Return("/current/dir/relative/repo", nil)

	// Mock path validation
	mockFS.EXPECT().Exists("/current/dir/relative/repo").Return(true, nil)
	mockFS.EXPECT().ValidateRepositoryPath("/current/dir/relative/repo").Return(true, nil)

	// Mock GetRemoteURL call (addRepositoriesToStatus calls it for all repos first)
	mockGit.EXPECT().GetRemoteURL("/current/dir/relative/repo", "origin").Return("https://github.com/user/relative-repo.git", nil)

	// Mock repository not found in status check (using the original remote URL)
	mockStatus.EXPECT().GetRepository("https://github.com/user/relative-repo.git").Return(nil, errors.New("not found"))

	// Mock GetRemoteURL call again (addRepositoryToStatus calls it)
	mockGit.EXPECT().GetRemoteURL("/current/dir/relative/repo", "origin").Return("https://github.com/user/relative-repo.git", nil)

	// Mock repository not found in status check (in addRepositoryToStatus, after normalization)
	mockStatus.EXPECT().GetRepository("github.com/user/relative-repo").Return(nil, errors.New("not found"))

	// Mock GetDefaultBranch for cloning
	mockGit.EXPECT().GetDefaultBranch("https://github.com/user/relative-repo.git").Return("main", nil)

	// Mock MkdirAll for creating parent directories
	mockFS.EXPECT().MkdirAll(gomock.Any(), gomock.Any()).Return(nil)

	// Mock Clone for cloning repository
	mockGit.EXPECT().Clone(gomock.Any()).Return(nil)

	// Mock adding repository to status (using normalized URL)
	mockStatus.EXPECT().AddRepository("github.com/user/relative-repo", gomock.Any()).Return(nil)

	// Mock adding workspace to status
	mockStatus.EXPECT().AddWorkspace("test-workspace", gomock.Any()).Return(nil)

	err := cm.CreateWorkspace(params)
	assert.NoError(t, err)
}

// TestCreateWorkspace_StatusUpdateFailure tests workspace creation when status update fails.
func TestCreateWorkspace_StatusUpdateFailure(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockFS := fsmocks.NewMockFS(ctrl)
	mockGit := gitmocks.NewMockGit(ctrl)
	mockStatus := statusmocks.NewMockManager(ctrl)

	cm := &realCodeManager{
		deps: dependencies.New().
			WithFS(mockFS).
			WithGit(mockGit).
			WithConfig(config.NewConfigManager("/test/config.yaml")).
			WithStatusManager(mockStatus).
			WithLogger(logger.NewNoopLogger()),
	}

	params := CreateWorkspaceParams{
		WorkspaceName: "test-workspace",
		Repositories:  []string{"repo1"},
	}

	// Mock workspace doesn't exist
	mockStatus.EXPECT().GetWorkspace("test-workspace").Return(nil, errors.New("not found"))

	// Mock repository exists in status (first call during resolution)
	existingRepo := &status.Repository{Path: "/path/to/repo1"}
	mockStatus.EXPECT().GetRepository("repo1").Return(existingRepo, nil)

	// Mock GetRemoteURL call (addRepositoriesToStatus calls it for all repos first)
	mockGit.EXPECT().GetRemoteURL("repo1", "origin").Return("github.com/user/repo1", nil)

	// Mock repository exists in status check (using the remote URL)
	mockStatus.EXPECT().GetRepository("github.com/user/repo1").Return(existingRepo, nil)

	// Mock adding workspace to status fails
	mockStatus.EXPECT().AddWorkspace("test-workspace", gomock.Any()).Return(errors.New("status update failed"))

	err := cm.CreateWorkspace(params)
	assert.Error(t, err)
	assert.ErrorIs(t, err, ErrStatusUpdate)
}

// TestCreateWorkspace_RepositoryAdditionFailure tests workspace creation when repository addition fails.
func TestCreateWorkspace_RepositoryAdditionFailure(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockFS := fsmocks.NewMockFS(ctrl)
	mockGit := gitmocks.NewMockGit(ctrl)
	mockStatus := statusmocks.NewMockManager(ctrl)

	cm := &realCodeManager{
		deps: dependencies.New().
			WithFS(mockFS).
			WithGit(mockGit).
			WithConfig(config.NewConfigManager("/test/config.yaml")).
			WithStatusManager(mockStatus).
			WithLogger(logger.NewNoopLogger()),
	}

	params := CreateWorkspaceParams{
		WorkspaceName: "test-workspace",
		Repositories:  []string{"/new/repo"},
	}

	// Mock workspace doesn't exist
	mockStatus.EXPECT().GetWorkspace("test-workspace").Return(nil, errors.New("not found"))

	// Mock repository not found in status (first call during resolution)
	mockStatus.EXPECT().GetRepository("/new/repo").Return(nil, errors.New("not found"))

	// Mock path validation
	mockFS.EXPECT().Exists("/new/repo").Return(true, nil)
	mockFS.EXPECT().ValidateRepositoryPath("/new/repo").Return(true, nil)

	// Mock GetRemoteURL call (addRepositoriesToStatus calls it for all repos first)
	mockGit.EXPECT().GetRemoteURL("/new/repo", "origin").Return("https://github.com/user/new-repo.git", nil)

	// Mock repository not found in status check (using the original remote URL)
	mockStatus.EXPECT().GetRepository("https://github.com/user/new-repo.git").Return(nil, errors.New("not found"))

	// Mock GetRemoteURL call again (addRepositoryToStatus calls it)
	mockGit.EXPECT().GetRemoteURL("/new/repo", "origin").Return("https://github.com/user/new-repo.git", nil)

	// Mock repository not found in status check (in addRepositoryToStatus, after normalization)
	mockStatus.EXPECT().GetRepository("github.com/user/new-repo").Return(nil, errors.New("not found"))

	// Mock GetDefaultBranch for cloning
	mockGit.EXPECT().GetDefaultBranch("https://github.com/user/new-repo.git").Return("main", nil)

	// Mock MkdirAll for creating parent directories
	mockFS.EXPECT().MkdirAll(gomock.Any(), gomock.Any()).Return(nil)

	// Mock Clone for cloning repository
	mockGit.EXPECT().Clone(gomock.Any()).Return(nil)

	// Mock adding repository to status fails (using normalized URL)
	mockStatus.EXPECT().AddRepository("github.com/user/new-repo", gomock.Any()).Return(errors.New("repository addition failed"))

	err := cm.CreateWorkspace(params)
	assert.Error(t, err)
	assert.ErrorIs(t, err, ErrRepositoryAddition)
}
