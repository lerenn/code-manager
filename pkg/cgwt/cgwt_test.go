//go:build unit

package cgwt

import (
	"testing"

	"github.com/lerenn/cgwt/pkg/config"
	"github.com/lerenn/cgwt/pkg/fs"
	"github.com/lerenn/cgwt/pkg/git"
	"github.com/stretchr/testify/assert"
	"go.uber.org/mock/gomock"
)

// createTestConfig creates a test configuration for unit tests.
func createTestConfig() *config.Config {
	return &config.Config{
		BasePath: "/test/base/path",
	}
}

func TestCGWT_Run_SingleRepository(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockFS := fs.NewMockFS(ctrl)
	mockGit := git.NewMockGit(ctrl)

	cgwt := NewCGWT(createTestConfig())

	// Override adapters with mocks
	c := cgwt.(*realCGWT)
	c.fs = mockFS
	c.git = mockGit

	// Mock single repo detection - .git found (called 2 times: detectProjectType and validateGitDirectory)
	mockFS.EXPECT().Exists(".git").Return(true, nil).Times(2)
	mockFS.EXPECT().IsDir(".git").Return(true, nil).Times(2) // Called in detectSingleRepoMode and validateGitDirectory

	// Mock Git status for validation (called 2 times: validateGitStatus and validateGitConfiguration)
	mockGit.EXPECT().Status(".").Return("On branch main", nil).Times(2)

	err := cgwt.CreateWorkTree()
	assert.NoError(t, err)
}

func TestCGWT_Run_VerboseMode(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockFS := fs.NewMockFS(ctrl)
	mockGit := git.NewMockGit(ctrl)

	cgwt := NewCGWT(createTestConfig())
	cgwt.SetVerbose(true)

	// Override adapters with mocks
	c := cgwt.(*realCGWT)
	c.fs = mockFS
	c.git = mockGit

	// Mock single repo detection - .git found (called 2 times: detectProjectType and validateGitDirectory)
	mockFS.EXPECT().Exists(".git").Return(true, nil).Times(2)
	mockFS.EXPECT().IsDir(".git").Return(true, nil).Times(2) // Called in detectSingleRepoMode and validateGitDirectory

	// Mock Git status for validation (called 2 times: validateGitStatus and validateGitConfiguration)
	mockGit.EXPECT().Status(".").Return("On branch main", nil).Times(2)

	err := cgwt.CreateWorkTree()
	assert.NoError(t, err)
}

func TestCGWT_ValidateSingleRepository_Success(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockFS := fs.NewMockFS(ctrl)
	mockGit := git.NewMockGit(ctrl)

	cgwt := NewCGWT(createTestConfig())
	cgwt.SetVerbose(true)

	// Override adapters with mocks
	c := cgwt.(*realCGWT)
	c.fs = mockFS
	c.git = mockGit

	// Mock repository validation
	mockFS.EXPECT().Exists(".git").Return(true, nil)
	mockFS.EXPECT().IsDir(".git").Return(true, nil)
	mockGit.EXPECT().Status(".").Return("On branch main", nil)
	mockGit.EXPECT().Status(".").Return("On branch main", nil) // Called twice for validation

	err := cgwt.(*realCGWT).validateSingleRepository()
	assert.NoError(t, err)
}

func TestCGWT_ValidateSingleRepository_NoGitDir(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockFS := fs.NewMockFS(ctrl)

	cgwt := NewCGWT(createTestConfig())
	cgwt.SetVerbose(true)

	// Override adapters with mocks
	c := cgwt.(*realCGWT)
	c.fs = mockFS

	// Mock repository validation - .git not found
	mockFS.EXPECT().Exists(".git").Return(false, nil)

	err := cgwt.(*realCGWT).validateSingleRepository()
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "not a valid Git repository: .git directory not found")
}

func TestCGWT_ValidateSingleRepository_GitStatusError(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockFS := fs.NewMockFS(ctrl)
	mockGit := git.NewMockGit(ctrl)

	cgwt := NewCGWT(createTestConfig())
	cgwt.SetVerbose(true)

	// Override adapters with mocks
	c := cgwt.(*realCGWT)
	c.fs = mockFS
	c.git = mockGit

	// Mock repository validation
	mockFS.EXPECT().Exists(".git").Return(true, nil)
	mockFS.EXPECT().IsDir(".git").Return(true, nil)
	mockGit.EXPECT().Status(".").Return("", assert.AnError)

	err := cgwt.(*realCGWT).validateSingleRepository()
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "not a valid Git repository")
}

func TestRealCGWT_CreateReposDirectoryStructure(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockFS := fs.NewMockFS(ctrl)
	mockGit := git.NewMockGit(ctrl)

	cgwt := NewCGWT(createTestConfig())
	cgwt.SetVerbose(true)

	// Override adapters with mocks
	c := cgwt.(*realCGWT)
	c.fs = mockFS
	c.git = mockGit

	// Test successful directory creation
	mockFS.EXPECT().MkdirAll("/test/base/path/repos/github.com/lerenn/cgwt/feature_new-branch", gomock.Any()).Return(nil)

	path, err := cgwt.(*realCGWT).createReposDirectoryStructure("github.com/lerenn/cgwt", "feature/new-branch")
	assert.NoError(t, err)
	assert.Equal(t, "/test/base/path/repos/github.com/lerenn/cgwt/feature_new-branch", path)
}

func TestRealCGWT_CreateReposDirectoryStructure_EmptyRepoName(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockFS := fs.NewMockFS(ctrl)
	mockGit := git.NewMockGit(ctrl)

	cgwt := NewCGWT(createTestConfig())
	cgwt.SetVerbose(true)

	// Override adapters with mocks
	c := cgwt.(*realCGWT)
	c.fs = mockFS
	c.git = mockGit

	_, err := cgwt.(*realCGWT).createReposDirectoryStructure("", "feature/new-branch")
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "repository URL cannot be empty")
}

func TestRealCGWT_CreateReposDirectoryStructure_EmptyBranchName(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockFS := fs.NewMockFS(ctrl)
	mockGit := git.NewMockGit(ctrl)

	cgwt := NewCGWT(createTestConfig())
	cgwt.SetVerbose(true)

	// Override adapters with mocks
	c := cgwt.(*realCGWT)
	c.fs = mockFS
	c.git = mockGit

	_, err := cgwt.(*realCGWT).createReposDirectoryStructure("github.com/lerenn/cgwt", "")
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "branch name cannot be empty")
}

func TestRealCGWT_CreateReposDirectoryStructure_NoConfig(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockFS := fs.NewMockFS(ctrl)
	mockGit := git.NewMockGit(ctrl)

	// Create CGWT with nil config
	cgwt := NewCGWT(nil)
	cgwt.SetVerbose(true)

	// Override adapters with mocks
	c := cgwt.(*realCGWT)
	c.fs = mockFS
	c.git = mockGit

	_, err := cgwt.(*realCGWT).createReposDirectoryStructure("github.com/lerenn/cgwt", "feature/new-branch")
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "configuration is not initialized")
}

func TestRealCGWT_sanitizeRepositoryName(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockFS := fs.NewMockFS(ctrl)
	mockGit := git.NewMockGit(ctrl)

	cgwt := NewCGWT(createTestConfig())

	// Override adapters with mocks
	c := cgwt.(*realCGWT)
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
			input:    "https://github.com/lerenn/cgwt.git",
			expected: "lerenn/cgwt",
			wantErr:  false,
		},
		{
			name:     "HTTPS URL without .git",
			input:    "https://github.com/lerenn/cgwt",
			expected: "lerenn/cgwt",
			wantErr:  false,
		},
		{
			name:     "SSH URL with .git",
			input:    "git@github.com:lerenn/cgwt.git",
			expected: "lerenn/cgwt",
			wantErr:  false,
		},
		{
			name:     "SSH URL without .git",
			input:    "git@github.com:lerenn/cgwt",
			expected: "lerenn/cgwt",
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
			result, err := cgwt.(*realCGWT).sanitizeRepositoryName(tt.input)
			if tt.wantErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
				assert.Equal(t, tt.expected, result)
			}
		})
	}
}

func TestRealCGWT_sanitizeBranchName(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockFS := fs.NewMockFS(ctrl)
	mockGit := git.NewMockGit(ctrl)

	cgwt := NewCGWT(createTestConfig())

	// Override adapters with mocks
	c := cgwt.(*realCGWT)
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
			result, err := cgwt.(*realCGWT).sanitizeBranchName(tt.input)
			if tt.wantErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
				assert.Equal(t, tt.expected, result)
			}
		})
	}
}

func TestRealCGWT_getBasePath(t *testing.T) {
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
			cgwt := NewCGWT(tt.config)

			// Override adapters with mocks
			c := cgwt.(*realCGWT)
			c.fs = mockFS
			c.git = mockGit

			result, err := cgwt.(*realCGWT).getBasePath()
			if tt.wantErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
				assert.Equal(t, tt.expected, result)
			}
		})
	}
}
