//go:build unit

package wtm

import (
	"testing"

	"github.com/lerenn/wtm/pkg/config"
	"github.com/lerenn/wtm/pkg/fs"
	"github.com/lerenn/wtm/pkg/git"
	"github.com/lerenn/wtm/pkg/status"
	"github.com/stretchr/testify/assert"
	"go.uber.org/mock/gomock"
)

// createTestConfig creates a test configuration for unit tests.
func createTestConfig() *config.Config {
	return &config.Config{
		BasePath:   "/test/base/path",
		StatusFile: "/test/base/path/status.yaml",
	}
}

func TestWTM_Run_SingleRepository(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockFS := fs.NewMockFS(ctrl)
	mockGit := git.NewMockGit(ctrl)
	mockStatus := status.NewMockManager(ctrl)

	wtm := NewWTM(createTestConfig())

	// Override adapters with mocks
	c := wtm.(*realWTM)
	c.fs = mockFS
	c.git = mockGit
	c.statusManager = mockStatus

	// Mock single repo detection - .git found (called multiple times: detectProjectType, validateGitDirectory, and createWorktreeForSingleRepo)
	mockFS.EXPECT().Exists(".git").Return(true, nil).AnyTimes() // detectProjectType, validateGitDirectory, createWorktreeForSingleRepo (validateRepository), prepareWorktreePath, and additional validation
	mockFS.EXPECT().IsDir(".git").Return(true, nil).AnyTimes()  // Called in detectSingleRepoMode, validateGitDirectory, createWorktreeForSingleRepo (validateRepository), prepareWorktreePath, and additional validation

	// Mock Git status for validation (called 2 times: validateGitStatus and validateGitConfiguration)
	mockGit.EXPECT().Status(".").Return("On branch main", nil).Times(2)

	// Mock status manager calls
	mockStatus.EXPECT().GetWorktree("github.com/lerenn/example", "heads/test-branch").Return(nil, status.ErrWorktreeNotFound).AnyTimes()
	mockStatus.EXPECT().AddWorktree("github.com/lerenn/example", "heads/test-branch", gomock.Any(), "").Return(nil)

	// Mock worktree creation calls
	mockGit.EXPECT().GetRepositoryName(gomock.Any()).Return("github.com/lerenn/example", nil)
	mockGit.EXPECT().IsClean(gomock.Any()).Return(true, nil)
	mockGit.EXPECT().BranchExists(gomock.Any(), "heads/test-branch").Return(false, nil)
	mockGit.EXPECT().CreateBranch(gomock.Any(), "heads/test-branch").Return(nil)
	mockFS.EXPECT().Exists(gomock.Any()).Return(false, nil).AnyTimes() // Worktree directory doesn't exist
	mockFS.EXPECT().MkdirAll(gomock.Any(), gomock.Any()).Return(nil)   // Create directory structure
	mockGit.EXPECT().CreateWorktree(gomock.Any(), gomock.Any(), "heads/test-branch").Return(nil)

	err := wtm.CreateWorkTree("test-branch")
	assert.NoError(t, err)
}

func TestWTM_Run_VerboseMode(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockFS := fs.NewMockFS(ctrl)
	mockGit := git.NewMockGit(ctrl)
	mockStatus := status.NewMockManager(ctrl)

	wtm := NewWTM(createTestConfig())
	wtm.SetVerbose(true)

	// Override adapters with mocks
	c := wtm.(*realWTM)
	c.fs = mockFS
	c.git = mockGit
	c.statusManager = mockStatus

	// Mock single repo detection - .git found (called multiple times: detectProjectType, validateGitDirectory, and createWorktreeForSingleRepo)
	mockFS.EXPECT().Exists(".git").Return(true, nil).AnyTimes() // detectProjectType, validateGitDirectory, createWorktreeForSingleRepo (validateRepository), prepareWorktreePath, and additional validation
	mockFS.EXPECT().IsDir(".git").Return(true, nil).AnyTimes()  // Called in detectSingleRepoMode, validateGitDirectory, createWorktreeForSingleRepo (validateRepository), prepareWorktreePath, and additional validation

	// Mock Git status for validation (called 2 times: validateGitStatus and validateGitConfiguration)
	mockGit.EXPECT().Status(".").Return("On branch main", nil).Times(2)

	// Mock status manager calls
	mockStatus.EXPECT().GetWorktree("github.com/lerenn/example", "heads/test-branch").Return(nil, status.ErrWorktreeNotFound).AnyTimes()
	mockStatus.EXPECT().AddWorktree("github.com/lerenn/example", "heads/test-branch", gomock.Any(), "").Return(nil)

	// Mock worktree creation calls
	mockGit.EXPECT().GetRepositoryName(gomock.Any()).Return("github.com/lerenn/example", nil)
	mockGit.EXPECT().IsClean(gomock.Any()).Return(true, nil)
	mockGit.EXPECT().BranchExists(gomock.Any(), "heads/test-branch").Return(false, nil)
	mockGit.EXPECT().CreateBranch(gomock.Any(), "heads/test-branch").Return(nil)
	mockFS.EXPECT().Exists(gomock.Any()).Return(false, nil).AnyTimes() // Worktree directory doesn't exist
	mockFS.EXPECT().MkdirAll(gomock.Any(), gomock.Any()).Return(nil)   // Create directory structure
	mockGit.EXPECT().CreateWorktree(gomock.Any(), gomock.Any(), "heads/test-branch").Return(nil)

	err := wtm.CreateWorkTree("test-branch")
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

	err := wtm.(*realWTM).validateCurrentDirIsGitRepository()
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

	err := wtm.(*realWTM).validateCurrentDirIsGitRepository()
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

	err := wtm.(*realWTM).validateCurrentDirIsGitRepository()
	assert.ErrorIs(t, err, ErrGitRepositoryInvalid)
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
