//go:build unit

package codemanager

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/lerenn/code-manager/pkg/config"
	configmocks "github.com/lerenn/code-manager/pkg/config/mocks"
	fsmocks "github.com/lerenn/code-manager/pkg/fs/mocks"
	"github.com/lerenn/code-manager/pkg/logger"
	"github.com/lerenn/code-manager/pkg/status"
	statusmocks "github.com/lerenn/code-manager/pkg/status/mocks"
	"github.com/stretchr/testify/assert"
	"go.uber.org/mock/gomock"
)

func TestValidateRepositoryName(t *testing.T) {
	tests := []struct {
		name           string
		repositoryName string
		expectedError  string
	}{
		{
			name:           "valid repository name",
			repositoryName: "my-repo",
			expectedError:  "",
		},
		{
			name:           "empty repository name",
			repositoryName: "",
			expectedError:  "repository name cannot be empty",
		},
		{
			name:           "repository name with forward slash (URL format)",
			repositoryName: "github.com/user/repo",
			expectedError:  "",
		},
		{
			name:           "repository name with backslash",
			repositoryName: "my\\repo",
			expectedError:  "repository name cannot contain backslashes",
		},
		{
			name:           "reserved name: dot",
			repositoryName: ".",
			expectedError:  "repository name '.' is reserved",
		},
		{
			name:           "reserved name: double dot",
			repositoryName: "..",
			expectedError:  "repository name '..' is reserved",
		},
		{
			name:           "reserved name: status.yaml",
			repositoryName: "status.yaml",
			expectedError:  "repository name 'status.yaml' is reserved",
		},
		{
			name:           "reserved name: config.yaml",
			repositoryName: "config.yaml",
			expectedError:  "repository name 'config.yaml' is reserved",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Create CM instance with minimal setup
			cmInstance := &realCodeManager{
				logger: logger.NewNoopLogger(),
			}

			// Execute test
			err := cmInstance.validateRepositoryName(tt.repositoryName)

			// Assert results
			if tt.expectedError != "" {
				assert.Error(t, err)
				assert.Contains(t, err.Error(), tt.expectedError)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

func TestValidateRepositoryNotInWorkspace(t *testing.T) {
	tests := []struct {
		name           string
		repositoryName string
		workspaces     map[string]status.Workspace
		expectedError  string
	}{
		{
			name:           "repository not in any workspace",
			repositoryName: "my-repo",
			workspaces: map[string]status.Workspace{
				"workspace1": {
					Repositories: []string{"other-repo"},
				},
			},
			expectedError: "",
		},
		{
			name:           "repository in workspace",
			repositoryName: "my-repo",
			workspaces: map[string]status.Workspace{
				"workspace1": {
					Repositories: []string{"my-repo", "other-repo"},
				},
			},
			expectedError: "repository 'my-repo' is part of workspace 'workspace1'. Remove it from the workspace before deleting",
		},
		{
			name:           "repository in multiple workspaces",
			repositoryName: "my-repo",
			workspaces: map[string]status.Workspace{
				"workspace1": {
					Repositories: []string{"my-repo"},
				},
				"workspace2": {
					Repositories: []string{"my-repo", "other-repo"},
				},
			},
			expectedError: "is part of workspace",
		},
		{
			name:           "no workspaces exist",
			repositoryName: "my-repo",
			workspaces:     map[string]status.Workspace{},
			expectedError:  "",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctrl := gomock.NewController(t)
			defer ctrl.Finish()

			// Create status manager mock
			statusMock := statusmocks.NewMockManager(ctrl)
			statusMock.EXPECT().ListWorkspaces().Return(tt.workspaces, nil)

			// Create CM instance
			cmInstance := &realCodeManager{
				statusManager: statusMock,
				logger:        logger.NewNoopLogger(),
			}

			// Execute test
			err := cmInstance.validateRepositoryNotInWorkspace(tt.repositoryName)

			// Assert results
			if tt.expectedError != "" {
				assert.Error(t, err)
				assert.Contains(t, err.Error(), tt.expectedError)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

func TestCleanupEmptyParentDirectories(t *testing.T) {
	tests := []struct {
		name              string
		repositoryPath    string
		repositoriesDir   string
		directoryExists   map[string]bool
		directoryContents map[string][]os.DirEntry
		expectedRemovals  []string
		expectedError     string
	}{
		{
			name:            "cleanup all empty parent directories",
			repositoryPath:  "/base/github.com/user/repo/origin/main",
			repositoriesDir: "/base",
			directoryExists: map[string]bool{
				"/base/github.com/user/repo/origin": true,
				"/base/github.com/user/repo":        true,
				"/base/github.com/user":             true,
				"/base/github.com":                  true,
			},
			directoryContents: map[string][]os.DirEntry{
				"/base/github.com/user/repo/origin": {},
				"/base/github.com/user/repo":        {},
				"/base/github.com/user":             {},
				"/base/github.com":                  {},
			},
			expectedRemovals: []string{
				"/base/github.com/user/repo/origin",
				"/base/github.com/user/repo",
				"/base/github.com/user",
				"/base/github.com",
			},
			expectedError: "",
		},
		{
			name:            "stop at non-empty directory",
			repositoryPath:  "/base/github.com/user/repo/origin/main",
			repositoriesDir: "/base",
			directoryExists: map[string]bool{
				"/base/github.com/user/repo/origin": true,
				"/base/github.com/user/repo":        true,
				"/base/github.com/user":             true,
				"/base/github.com":                  true,
			},
			directoryContents: map[string][]os.DirEntry{
				"/base/github.com/user/repo/origin": {},
				"/base/github.com/user/repo":        {},
				"/base/github.com/user":             {&mockDirEntry{name: "other-repo"}},
				"/base/github.com":                  {},
			},
			expectedRemovals: []string{
				"/base/github.com/user/repo/origin",
				"/base/github.com/user/repo",
			},
			expectedError: "",
		},
		{
			name:            "stop at repositories directory",
			repositoryPath:  "/base/github.com/user/repo/origin/main",
			repositoriesDir: "/base",
			directoryExists: map[string]bool{
				"/base/github.com/user/repo/origin": true,
				"/base/github.com/user/repo":        true,
				"/base/github.com/user":             true,
				"/base/github.com":                  true,
			},
			directoryContents: map[string][]os.DirEntry{
				"/base/github.com/user/repo/origin": {},
				"/base/github.com/user/repo":        {},
				"/base/github.com/user":             {},
				"/base/github.com":                  {},
			},
			expectedRemovals: []string{
				"/base/github.com/user/repo/origin",
				"/base/github.com/user/repo",
				"/base/github.com/user",
				"/base/github.com",
			},
			expectedError: "",
		},
		{
			name:            "handle non-existent directories",
			repositoryPath:  "/base/github.com/user/repo/origin/main",
			repositoriesDir: "/base",
			directoryExists: map[string]bool{
				"/base/github.com/user/repo/origin": true,
				"/base/github.com/user/repo":        false, // doesn't exist
				"/base/github.com/user":             true,
				"/base/github.com":                  true,
			},
			directoryContents: map[string][]os.DirEntry{
				"/base/github.com/user/repo/origin": {},
				"/base/github.com/user":             {},
				"/base/github.com":                  {},
			},
			expectedRemovals: []string{
				"/base/github.com/user/repo/origin",
				"/base/github.com/user",
				"/base/github.com",
			},
			expectedError: "",
		},
		{
			name:            "handle read directory error",
			repositoryPath:  "/base/github.com/user/repo/origin/main",
			repositoriesDir: "/base",
			directoryExists: map[string]bool{
				"/base/github.com/user/repo/origin": true,
				"/base/github.com/user/repo":        true,
			},
			directoryContents: map[string][]os.DirEntry{
				"/base/github.com/user/repo/origin": {},
			},
			expectedRemovals: []string{},
			expectedError:    "failed to check if directory is empty",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctrl := gomock.NewController(t)
			defer ctrl.Finish()

			// Create file system mock
			fsMock := fsmocks.NewMockFS(ctrl)

			// Create config manager mock
			configMock := configmocks.NewMockManager(ctrl)

			// Setup config mock expectations
			testConfig := config.Config{
				RepositoriesDir: tt.repositoriesDir,
				WorkspacesDir:   "/test/workspaces",
				StatusFile:      "/test/status.yaml",
			}
			configMock.EXPECT().GetConfigWithFallback().Return(testConfig, nil).AnyTimes()

			// Setup mock expectations
			parentDir := filepath.Dir(tt.repositoryPath)
			removalCount := 0

			for parentDir != tt.repositoriesDir && parentDir != filepath.Dir(parentDir) {
				// Mock Exists call
				exists := tt.directoryExists[parentDir]
				fsMock.EXPECT().Exists(parentDir).Return(exists, nil)

				if exists {
					// Mock ReadDir call
					contents, hasContents := tt.directoryContents[parentDir]
					if hasContents {
						fsMock.EXPECT().ReadDir(parentDir).Return(contents, nil)
					} else {
						// This will cause the error case
						fsMock.EXPECT().ReadDir(parentDir).Return(nil, os.ErrPermission)
						break
					}

					// If directory is empty, mock Remove call
					if len(contents) == 0 {
						fsMock.EXPECT().Remove(parentDir).Return(nil)
						removalCount++
					} else {
						// Directory not empty, stop cleanup
						break
					}
				}

				// Move up to next parent
				parentDir = filepath.Dir(parentDir)
			}

			// Create CM instance
			cmInstance := &realCodeManager{
				fs:            fsMock,
				logger:        logger.NewNoopLogger(),
				configManager: configMock,
			}

			// Execute test
			err := cmInstance.cleanupEmptyParentDirectories(tt.repositoryPath)

			// Assert results
			if tt.expectedError != "" {
				assert.Error(t, err)
				assert.Contains(t, err.Error(), tt.expectedError)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

func TestIsDirectoryEmpty(t *testing.T) {
	tests := []struct {
		name          string
		dirPath       string
		readDirResult []os.DirEntry
		readDirError  error
		expectedEmpty bool
		expectedError string
	}{
		{
			name:          "empty directory",
			dirPath:       "/test/empty",
			readDirResult: []os.DirEntry{},
			readDirError:  nil,
			expectedEmpty: true,
			expectedError: "",
		},
		{
			name:    "non-empty directory",
			dirPath: "/test/non-empty",
			readDirResult: []os.DirEntry{
				&mockDirEntry{name: "file1"},
				&mockDirEntry{name: "file2"},
			},
			readDirError:  nil,
			expectedEmpty: false,
			expectedError: "",
		},
		{
			name:          "read directory error",
			dirPath:       "/test/error",
			readDirResult: nil,
			readDirError:  os.ErrPermission,
			expectedEmpty: false,
			expectedError: "failed to read directory",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctrl := gomock.NewController(t)
			defer ctrl.Finish()

			// Create file system mock
			fsMock := fsmocks.NewMockFS(ctrl)
			fsMock.EXPECT().ReadDir(tt.dirPath).Return(tt.readDirResult, tt.readDirError)

			// Create CM instance
			cmInstance := &realCodeManager{
				fs:     fsMock,
				logger: logger.NewNoopLogger(),
			}

			// Execute test
			isEmpty, err := cmInstance.isDirectoryEmpty(tt.dirPath)

			// Assert results
			if tt.expectedError != "" {
				assert.Error(t, err)
				assert.Contains(t, err.Error(), tt.expectedError)
				assert.False(t, isEmpty)
			} else {
				assert.NoError(t, err)
				assert.Equal(t, tt.expectedEmpty, isEmpty)
			}
		})
	}
}

// mockDirEntry is a simple implementation of fs.DirEntry for testing
type mockDirEntry struct {
	name string
}

func (m *mockDirEntry) Name() string               { return m.name }
func (m *mockDirEntry) IsDir() bool                { return false }
func (m *mockDirEntry) Type() os.FileMode          { return 0 }
func (m *mockDirEntry) Info() (os.FileInfo, error) { return nil, nil }
