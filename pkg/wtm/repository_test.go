//go:build unit

package wtm

import (
	"testing"

	"github.com/lerenn/wtm/pkg/config"
	"github.com/lerenn/wtm/pkg/fs"
	"github.com/lerenn/wtm/pkg/git"
	"github.com/lerenn/wtm/pkg/logger"
	"github.com/lerenn/wtm/pkg/status"
	"github.com/stretchr/testify/assert"
	"go.uber.org/mock/gomock"
)

func TestRepository_Validate_Success(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockFS := fs.NewMockFS(ctrl)
	mockGit := git.NewMockGit(ctrl)
	mockStatus := status.NewMockManager(ctrl)

	repo := newRepository(mockFS, mockGit, createTestConfig(), mockStatus, logger.NewNoopLogger(), true)

	// Mock repository validation
	mockFS.EXPECT().Exists(".git").Return(true, nil)
	mockFS.EXPECT().IsDir(".git").Return(true, nil)
	mockGit.EXPECT().Status(".").Return("On branch main", nil)
	mockGit.EXPECT().Status(".").Return("On branch main", nil) // Called twice for validation

	err := repo.Validate()
	assert.NoError(t, err)
}

func TestRepository_Validate_NoGitDir(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockFS := fs.NewMockFS(ctrl)
	mockGit := git.NewMockGit(ctrl)
	mockStatus := status.NewMockManager(ctrl)

	repo := newRepository(mockFS, mockGit, createTestConfig(), mockStatus, logger.NewNoopLogger(), true)

	// Mock repository validation - .git not found
	mockFS.EXPECT().Exists(".git").Return(false, nil)

	err := repo.Validate()
	assert.ErrorIs(t, err, ErrGitRepositoryNotFound)
}

func TestRepository_Validate_GitStatusError(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockFS := fs.NewMockFS(ctrl)
	mockGit := git.NewMockGit(ctrl)
	mockStatus := status.NewMockManager(ctrl)

	repo := newRepository(mockFS, mockGit, createTestConfig(), mockStatus, logger.NewNoopLogger(), true)

	// Mock repository validation
	mockFS.EXPECT().Exists(".git").Return(true, nil)
	mockFS.EXPECT().IsDir(".git").Return(true, nil)
	mockGit.EXPECT().Status(".").Return("", assert.AnError)

	err := repo.Validate()
	assert.ErrorIs(t, err, ErrGitRepositoryInvalid)
}

func TestRepository_CreateWorktree_Success(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockFS := fs.NewMockFS(ctrl)
	mockGit := git.NewMockGit(ctrl)
	mockStatus := status.NewMockManager(ctrl)

	repo := newRepository(mockFS, mockGit, createTestConfig(), mockStatus, logger.NewNoopLogger(), true)

	// Mock repository validation
	mockFS.EXPECT().Exists(".git").Return(true, nil).AnyTimes()
	mockFS.EXPECT().IsDir(".git").Return(true, nil).AnyTimes()
	mockGit.EXPECT().Status(".").Return("On branch main", nil).AnyTimes()

	// Mock worktree creation
	mockGit.EXPECT().GetRepositoryName(gomock.Any()).Return("github.com/lerenn/example", nil)
	mockStatus.EXPECT().GetWorktree("github.com/lerenn/example", "test-branch").Return(nil, status.ErrWorktreeNotFound)
	mockGit.EXPECT().IsClean(gomock.Any()).Return(true, nil)
	mockFS.EXPECT().Exists(gomock.Any()).Return(false, nil).AnyTimes()
	mockFS.EXPECT().MkdirAll(gomock.Any(), gomock.Any()).Return(nil)
	mockStatus.EXPECT().AddWorktree("github.com/lerenn/example", "test-branch", gomock.Any(), "").Return(nil)
	mockGit.EXPECT().BranchExists(gomock.Any(), "test-branch").Return(false, nil)
	mockGit.EXPECT().CreateBranch(gomock.Any(), "test-branch").Return(nil)
	mockGit.EXPECT().CreateWorktree(gomock.Any(), gomock.Any(), "test-branch").Return(nil)

	err := repo.CreateWorktree("test-branch")
	assert.NoError(t, err)
}

func TestRepository_CheckGitDirExists_Success(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockFS := fs.NewMockFS(ctrl)
	mockGit := git.NewMockGit(ctrl)
	mockStatus := status.NewMockManager(ctrl)

	repo := newRepository(mockFS, mockGit, createTestConfig(), mockStatus, logger.NewNoopLogger(), true)

	// Mock .git directory check
	mockFS.EXPECT().Exists(".git").Return(true, nil)
	mockFS.EXPECT().IsDir(".git").Return(true, nil)

	exists, err := repo.CheckGitDirExists()
	assert.NoError(t, err)
	assert.True(t, exists)
}

func TestRepository_CheckGitDirExists_NotFound(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockFS := fs.NewMockFS(ctrl)
	mockGit := git.NewMockGit(ctrl)
	mockStatus := status.NewMockManager(ctrl)

	repo := newRepository(mockFS, mockGit, createTestConfig(), mockStatus, logger.NewNoopLogger(), true)

	// Mock .git directory not found
	mockFS.EXPECT().Exists(".git").Return(false, nil)

	exists, err := repo.CheckGitDirExists()
	assert.NoError(t, err)
	assert.False(t, exists)
}

func TestRepository_CheckGitDirExists_NotDirectory(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockFS := fs.NewMockFS(ctrl)
	mockGit := git.NewMockGit(ctrl)
	mockStatus := status.NewMockManager(ctrl)

	repo := newRepository(mockFS, mockGit, createTestConfig(), mockStatus, logger.NewNoopLogger(), true)

	// Mock .git exists but is not a directory
	mockFS.EXPECT().Exists(".git").Return(true, nil)
	mockFS.EXPECT().IsDir(".git").Return(false, nil)

	exists, err := repo.CheckGitDirExists()
	assert.NoError(t, err)
	assert.False(t, exists)
}

func TestRepository_getBasePath(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockFS := fs.NewMockFS(ctrl)
	mockGit := git.NewMockGit(ctrl)
	mockStatus := status.NewMockManager(ctrl)

	tests := []struct {
		name     string
		config   *config.Config
		expected string
		wantErr  bool
	}{
		{
			name: "Valid config",
			config: &config.Config{
				BasePath: "/custom/path",
			},
			expected: "/custom/path",
			wantErr:  false,
		},
		{
			name:     "Nil config",
			config:   nil,
			expected: "",
			wantErr:  true,
		},
		{
			name: "Empty base path",
			config: &config.Config{
				BasePath: "",
			},
			expected: "",
			wantErr:  true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			repo := newRepository(mockFS, mockGit, tt.config, mockStatus, logger.NewNoopLogger(), true)

			result, err := repo.getBasePath()
			if tt.wantErr {
				if tt.config == nil {
					assert.ErrorIs(t, err, ErrConfigurationNotInitialized)
				} else {
					assert.Error(t, err)
				}
			} else {
				assert.NoError(t, err)
				assert.Equal(t, tt.expected, result)
			}
		})
	}
}
