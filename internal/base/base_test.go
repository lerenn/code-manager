//go:build unit

package base

import (
	"testing"

	"github.com/lerenn/code-manager/pkg/config"
	"github.com/lerenn/code-manager/pkg/fs"
	"github.com/lerenn/code-manager/pkg/git"
	"github.com/lerenn/code-manager/pkg/logger"
	"github.com/lerenn/code-manager/pkg/prompt"
	"github.com/lerenn/code-manager/pkg/status"
	"github.com/stretchr/testify/assert"
	"go.uber.org/mock/gomock"
)

func TestBase_VerbosePrint_Enabled(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockFS := fs.NewMockFS(ctrl)
	mockGit := git.NewMockGit(ctrl)
	mockStatus := status.NewMockManager(ctrl)
	mockLogger := logger.NewMockLogger(ctrl)
	mockPrompt := prompt.NewMockPrompt(ctrl)

	base := NewBase(NewBaseParams{
		FS:            mockFS,
		Git:           mockGit,
		Config:        &config.Config{},
		StatusManager: mockStatus,
		Logger:        mockLogger,
		Prompt:        mockPrompt,
		Verbose:       true,
	})

	// Expect verbose print to be called
	mockLogger.EXPECT().Logf("Test message with arg").Times(1)

	base.VerbosePrint("Test message with %s", "arg")
}

func TestBase_VerbosePrint_Disabled(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockFS := fs.NewMockFS(ctrl)
	mockGit := git.NewMockGit(ctrl)
	mockStatus := status.NewMockManager(ctrl)
	mockLogger := logger.NewMockLogger(ctrl)
	mockPrompt := prompt.NewMockPrompt(ctrl)

	base := NewBase(NewBaseParams{
		FS:            mockFS,
		Git:           mockGit,
		Config:        &config.Config{},
		StatusManager: mockStatus,
		Logger:        mockLogger,
		Prompt:        mockPrompt,
		Verbose:       false,
	})

	// Expect no verbose print to be called
	// No expectations set on mockLogger

	base.VerbosePrint("Test message with %s", "arg")
}

func TestBase_validateGitConfiguration(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockFS := fs.NewMockFS(ctrl)
	mockGit := git.NewMockGit(ctrl)
	mockStatus := status.NewMockManager(ctrl)
	mockLogger := logger.NewNoopLogger()
	mockPrompt := prompt.NewMockPrompt(ctrl)
	tests := []struct {
		name        string
		workDir     string
		gitStatus   string
		gitError    error
		verbose     bool
		expectError bool
	}{
		{
			name:        "Valid Git configuration",
			workDir:     "/test/repo",
			gitStatus:   "On branch main",
			gitError:    nil,
			verbose:     true,
			expectError: false,
		},
		{
			name:        "Git status error",
			workDir:     "/test/repo",
			gitStatus:   "",
			gitError:    assert.AnError,
			verbose:     true,
			expectError: true,
		},
		{
			name:        "Valid Git configuration without verbose",
			workDir:     "/test/repo",
			gitStatus:   "On branch main",
			gitError:    nil,
			verbose:     false,
			expectError: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			base := NewBase(NewBaseParams{
				FS:            mockFS,
				Git:           mockGit,
				Config:        &config.Config{},
				StatusManager: mockStatus,
				Logger:        mockLogger,
				Prompt:        mockPrompt,
				Verbose:       tt.verbose,
			})

			mockGit.EXPECT().Status(tt.workDir).Return(tt.gitStatus, tt.gitError).Times(1)

			err := base.ValidateGitConfiguration(tt.workDir)

			if tt.expectError {
				assert.Error(t, err)
				assert.ErrorIs(t, err, ErrGitConfiguration)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

func TestBase_cleanupWorktreeDirectory(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockFS := fs.NewMockFS(ctrl)
	mockGit := git.NewMockGit(ctrl)
	mockStatus := status.NewMockManager(ctrl)
	mockLogger := logger.NewNoopLogger()
	mockPrompt := prompt.NewMockPrompt(ctrl)
	tests := []struct {
		name          string
		worktreePath  string
		exists        bool
		existsError   error
		removeError   error
		verbose       bool
		expectError   bool
		expectedError string
	}{
		{
			name:         "Directory exists and removed successfully",
			worktreePath: "/test/worktree",
			exists:       true,
			existsError:  nil,
			removeError:  nil,
			verbose:      true,
			expectError:  false,
		},
		{
			name:         "Directory does not exist",
			worktreePath: "/test/worktree",
			exists:       false,
			existsError:  nil,
			removeError:  nil,
			verbose:      true,
			expectError:  false,
		},
		{
			name:          "Exists check fails",
			worktreePath:  "/test/worktree",
			exists:        false,
			existsError:   assert.AnError,
			removeError:   nil,
			verbose:       true,
			expectError:   true,
			expectedError: "failed to check if worktree directory exists",
		},
		{
			name:          "Remove fails",
			worktreePath:  "/test/worktree",
			exists:        true,
			existsError:   nil,
			removeError:   assert.AnError,
			verbose:       true,
			expectError:   true,
			expectedError: "failed to remove worktree directory",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			base := NewBase(NewBaseParams{
				FS:            mockFS,
				Git:           mockGit,
				Config:        &config.Config{},
				StatusManager: mockStatus,
				Logger:        mockLogger,
				Prompt:        mockPrompt,
				Verbose:       tt.verbose,
			})

			mockFS.EXPECT().Exists(tt.worktreePath).Return(tt.exists, tt.existsError).Times(1)

			if tt.exists && tt.existsError == nil {
				mockFS.EXPECT().RemoveAll(tt.worktreePath).Return(tt.removeError).Times(1)
			}

			err := base.CleanupWorktreeDirectory(tt.worktreePath)

			if tt.expectError {
				assert.Error(t, err)
				if tt.expectedError != "" {
					if tt.expectedError == "failed to check if worktree directory exists" {
						assert.ErrorIs(t, err, ErrFailedToCheckWorktreeDirectoryExists)
					} else if tt.expectedError == "failed to remove worktree directory" {
						assert.ErrorIs(t, err, ErrFailedToRemoveWorktreeDirectory)
					}
				}
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

func TestBase_buildWorktreePath(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockFS := fs.NewMockFS(ctrl)
	mockGit := git.NewMockGit(ctrl)
	mockStatus := status.NewMockManager(ctrl)
	mockLogger := logger.NewNoopLogger()
	mockPrompt := prompt.NewMockPrompt(ctrl)
	tests := []struct {
		name       string
		basePath   string
		repoURL    string
		remoteName string
		branch     string
		expected   string
	}{
		{
			name:       "Simple path construction with base path",
			basePath:   "/base/path",
			repoURL:    "github.com/lerenn/example",
			remoteName: "origin",
			branch:     "main",
			expected:   "/base/path/github.com/lerenn/example/origin/main",
		},
		{
			name:       "Path with branch containing slash",
			basePath:   "/base/path",
			repoURL:    "github.com/lerenn/example",
			remoteName: "origin",
			branch:     "feature/new-feature",
			expected:   "/base/path/github.com/lerenn/example/origin/feature/new-feature",
		},
		{
			name:       "Empty base path",
			basePath:   "",
			repoURL:    "github.com/lerenn/example",
			remoteName: "origin",
			branch:     "main",
			expected:   "github.com/lerenn/example/origin/main",
		},
		{
			name:       "Custom remote name",
			basePath:   "/base/path",
			repoURL:    "github.com/lerenn/example",
			remoteName: "upstream",
			branch:     "main",
			expected:   "/base/path/github.com/lerenn/example/upstream/main",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			base := NewBase(NewBaseParams{
				FS:  mockFS,
				Git: mockGit,
				Config: &config.Config{
					BasePath: tt.basePath,
				},
				StatusManager: mockStatus,
				Logger:        mockLogger,
				Prompt:        mockPrompt,
				Verbose:       false,
			})

			result := base.BuildWorktreePath(tt.repoURL, tt.remoteName, tt.branch)
			assert.Equal(t, tt.expected, result)
		})
	}
}
