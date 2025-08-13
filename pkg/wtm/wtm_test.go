//go:build unit

package wtm

import (
	"testing"

	"github.com/lerenn/wtm/pkg/config"
	"github.com/lerenn/wtm/pkg/fs"
	"github.com/lerenn/wtm/pkg/git"
	"github.com/stretchr/testify/assert"
	"go.uber.org/mock/gomock"
)

// createTestConfig creates a test configuration for unit tests.
func createTestConfig() *config.Config {
	return &config.Config{
		BasePath: "/test/base/path",
	}
}

func TestWTM_Run_SingleRepository(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockFS := fs.NewMockFS(ctrl)
	mockGit := git.NewMockGit(ctrl)

	wtm := NewWTM(createTestConfig())

	// Override adapters with mocks
	c := wtm.(*realWTM)
	c.fs = mockFS
	c.git = mockGit

	// Mock single repo detection - .git found (called 2 times: detectProjectType and validateGitDirectory)
	mockFS.EXPECT().Exists(".git").Return(true, nil).Times(2)
	mockFS.EXPECT().IsDir(".git").Return(true, nil).Times(2) // Called in detectSingleRepoMode and validateGitDirectory

	// Mock Git status for validation (called 2 times: validateGitStatus and validateGitConfiguration)
	mockGit.EXPECT().Status(".").Return("On branch main", nil).Times(2)

	err := wtm.CreateWorkTree()
	assert.NoError(t, err)
}

func TestWTM_Run_VerboseMode(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockFS := fs.NewMockFS(ctrl)
	mockGit := git.NewMockGit(ctrl)

	wtm := NewWTM(createTestConfig())
	wtm.SetVerbose(true)

	// Override adapters with mocks
	c := wtm.(*realWTM)
	c.fs = mockFS
	c.git = mockGit

	// Mock single repo detection - .git found (called 2 times: detectProjectType and validateGitDirectory)
	mockFS.EXPECT().Exists(".git").Return(true, nil).Times(2)
	mockFS.EXPECT().IsDir(".git").Return(true, nil).Times(2) // Called in detectSingleRepoMode and validateGitDirectory

	// Mock Git status for validation (called 2 times: validateGitStatus and validateGitConfiguration)
	mockGit.EXPECT().Status(".").Return("On branch main", nil).Times(2)

	err := wtm.CreateWorkTree()
	assert.NoError(t, err)
}

func TestWTM_ValidateSingleRepository_Success(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockFS := fs.NewMockFS(ctrl)
	mockGit := git.NewMockGit(ctrl)

	wtm := NewWTM(createTestConfig())
	wtm.SetVerbose(true)

	// Override adapters with mocks
	c := wtm.(*realWTM)
	c.fs = mockFS
	c.git = mockGit

	// Mock repository validation
	mockFS.EXPECT().Exists(".git").Return(true, nil)
	mockFS.EXPECT().IsDir(".git").Return(true, nil)
	mockGit.EXPECT().Status(".").Return("On branch main", nil)
	mockGit.EXPECT().Status(".").Return("On branch main", nil) // Called twice for validation

	err := wtm.(*realWTM).validateSingleRepository()
	assert.NoError(t, err)
}

func TestWTM_ValidateSingleRepository_NoGitDir(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockFS := fs.NewMockFS(ctrl)

	wtm := NewWTM(createTestConfig())
	wtm.SetVerbose(true)

	// Override adapters with mocks
	c := wtm.(*realWTM)
	c.fs = mockFS

	// Mock repository validation - .git not found
	mockFS.EXPECT().Exists(".git").Return(false, nil)

	err := wtm.(*realWTM).validateSingleRepository()
	assert.ErrorIs(t, err, ErrGitRepositoryNotFound)
}

func TestWTM_ValidateSingleRepository_GitStatusError(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockFS := fs.NewMockFS(ctrl)
	mockGit := git.NewMockGit(ctrl)

	wtm := NewWTM(createTestConfig())
	wtm.SetVerbose(true)

	// Override adapters with mocks
	c := wtm.(*realWTM)
	c.fs = mockFS
	c.git = mockGit

	// Mock repository validation
	mockFS.EXPECT().Exists(".git").Return(true, nil)
	mockFS.EXPECT().IsDir(".git").Return(true, nil)
	mockGit.EXPECT().Status(".").Return("", assert.AnError)

	err := wtm.(*realWTM).validateSingleRepository()
	assert.ErrorIs(t, err, ErrGitRepositoryInvalid)
}

func TestRealWTM_CreateReposDirectoryStructure(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockFS := fs.NewMockFS(ctrl)
	mockGit := git.NewMockGit(ctrl)

	wtm := NewWTM(createTestConfig())
	wtm.SetVerbose(true)

	// Override adapters with mocks
	c := wtm.(*realWTM)
	c.fs = mockFS
	c.git = mockGit

	// Test successful directory creation
	mockFS.EXPECT().MkdirAll("/test/base/path/repos/github.com/lerenn/wtm/feature_new-branch", gomock.Any()).Return(nil)

	path, err := wtm.(*realWTM).createReposDirectoryStructure("github.com/lerenn/wtm", "feature/new-branch")
	assert.NoError(t, err)
	assert.Equal(t, "/test/base/path/repos/github.com/lerenn/wtm/feature_new-branch", path)
}

func TestRealWTM_CreateReposDirectoryStructure_EmptyRepoName(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockFS := fs.NewMockFS(ctrl)
	mockGit := git.NewMockGit(ctrl)

	wtm := NewWTM(createTestConfig())
	wtm.SetVerbose(true)

	// Override adapters with mocks
	c := wtm.(*realWTM)
	c.fs = mockFS
	c.git = mockGit

	_, err := wtm.(*realWTM).createReposDirectoryStructure("", "feature/new-branch")
	assert.ErrorIs(t, err, ErrRepositoryURLEmpty)
}

func TestRealWTM_CreateReposDirectoryStructure_EmptyBranchName(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockFS := fs.NewMockFS(ctrl)
	mockGit := git.NewMockGit(ctrl)

	wtm := NewWTM(createTestConfig())
	wtm.SetVerbose(true)

	// Override adapters with mocks
	c := wtm.(*realWTM)
	c.fs = mockFS
	c.git = mockGit

	_, err := wtm.(*realWTM).createReposDirectoryStructure("github.com/lerenn/wtm", "")
	assert.ErrorIs(t, err, ErrBranchNameEmpty)
}

func TestRealWTM_CreateReposDirectoryStructure_NoConfig(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockFS := fs.NewMockFS(ctrl)
	mockGit := git.NewMockGit(ctrl)

	// Create WTM with nil config
	wtm := NewWTM(nil)
	wtm.SetVerbose(true)

	// Override adapters with mocks
	c := wtm.(*realWTM)
	c.fs = mockFS
	c.git = mockGit

	_, err := wtm.(*realWTM).createReposDirectoryStructure("github.com/lerenn/wtm", "feature/new-branch")
	assert.ErrorIs(t, err, ErrConfigurationNotInitialized)
}

func TestRealWTM_sanitizeRepositoryName(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockFS := fs.NewMockFS(ctrl)
	mockGit := git.NewMockGit(ctrl)

	wtm := NewWTM(createTestConfig())

	// Override adapters with mocks
	c := wtm.(*realWTM)
	c.fs = mockFS
	c.git = mockGit

	tests := []struct {
		name     string
		input    string
		expected string
		wantErr  bool
	}{
		{
			name:     "HTTPS URL with .git",
			input:    "https://github.com/lerenn/wtm.git",
			expected: "lerenn/wtm",
			wantErr:  false,
		},
		{
			name:     "HTTPS URL without .git",
			input:    "https://github.com/lerenn/wtm",
			expected: "lerenn/wtm",
			wantErr:  false,
		},
		{
			name:     "SSH URL with .git",
			input:    "git@github.com:lerenn/wtm.git",
			expected: "lerenn/wtm",
			wantErr:  false,
		},
		{
			name:     "SSH URL without .git",
			input:    "git@github.com:lerenn/wtm",
			expected: "lerenn/wtm",
			wantErr:  false,
		},
		{
			name:     "Repository with invalid characters",
			input:    "https://github.com/user/repo:name",
			expected: "user/repo_name",
			wantErr:  false,
		},
		{
			name:     "Empty repository URL",
			input:    "",
			expected: "",
			wantErr:  true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := wtm.(*realWTM).sanitizeRepositoryName(tt.input)
			if tt.wantErr {
				if tt.input == "" {
					assert.ErrorIs(t, err, ErrRepositoryURLEmpty)
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

func TestRealWTM_sanitizeBranchName(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockFS := fs.NewMockFS(ctrl)
	mockGit := git.NewMockGit(ctrl)

	wtm := NewWTM(createTestConfig())

	// Override adapters with mocks
	c := wtm.(*realWTM)
	c.fs = mockFS
	c.git = mockGit

	tests := []struct {
		name     string
		input    string
		expected string
		wantErr  bool
	}{
		{
			name:     "Simple branch name",
			input:    "feature/new-branch",
			expected: "feature_new-branch",
			wantErr:  false,
		},
		{
			name:     "Branch name with invalid characters",
			input:    "bugfix/issue#123",
			expected: "bugfix_issue_123",
			wantErr:  false,
		},
		{
			name:     "Branch name with dots",
			input:    "release/v1.0.0",
			expected: "release_v1.0.0",
			wantErr:  false,
		},
		{
			name:     "Empty branch name",
			input:    "",
			expected: "",
			wantErr:  true,
		},
		{
			name:     "Branch name with leading/trailing dots",
			input:    ".hidden-branch.",
			expected: "hidden-branch",
			wantErr:  false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := wtm.(*realWTM).sanitizeBranchName(tt.input)
			if tt.wantErr {
				if tt.input == "" {
					assert.ErrorIs(t, err, ErrBranchNameEmpty)
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

func TestRealWTM_getBasePath(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockFS := fs.NewMockFS(ctrl)
	mockGit := git.NewMockGit(ctrl)

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
			wtm := NewWTM(tt.config)

			// Override adapters with mocks
			c := wtm.(*realWTM)
			c.fs = mockFS
			c.git = mockGit

			result, err := wtm.(*realWTM).getBasePath()
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
