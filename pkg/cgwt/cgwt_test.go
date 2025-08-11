//go:build unit

package cgwt

import (
	"testing"

	"github.com/lerenn/cgwt/pkg/fs"
	"github.com/lerenn/cgwt/pkg/git"
	"github.com/lerenn/cgwt/pkg/logger"
	"github.com/stretchr/testify/assert"
	"go.uber.org/mock/gomock"
)

func TestCGWT_Run_SingleRepository(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockFS := fs.NewMockFS(ctrl)
	mockGit := git.NewMockGit(ctrl)
	mockLogger := logger.NewMockLogger(ctrl)

	cgwt := NewCGWT()
	cgwt.SetLogger(mockLogger)

	// Override adapters with mocks
	c := cgwt.(*realCGWT)
	c.fs = mockFS
	c.git = mockGit

	// Mock single repo detection - .git found (called 3 times: detectProjectMode, validateProjectStructure, validateGitDirectory)
	mockFS.EXPECT().Exists(".git").Return(true, nil).Times(3)
	mockFS.EXPECT().IsDir(".git").Return(true, nil).Times(3) // Called in detectSingleRepoMode (2x) and validateGitDirectory (1x)

	// Mock Git status for validation (called 2 times: validateGitStatus and validateGitConfiguration)
	mockGit.EXPECT().Status(".").Return("On branch main", nil).Times(2)

	err := cgwt.Run()
	assert.NoError(t, err)
}

func TestCGWT_Run_VerboseMode(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockFS := fs.NewMockFS(ctrl)
	mockGit := git.NewMockGit(ctrl)
	mockLogger := logger.NewMockLogger(ctrl)

	cgwt := NewCGWT()
	cgwt.SetVerbose(true)
	cgwt.SetLogger(mockLogger)

	// Override adapters with mocks
	c := cgwt.(*realCGWT)
	c.fs = mockFS
	c.git = mockGit

	// Mock single repo detection - .git found (called 3 times: detectProjectMode, validateProjectStructure, validateGitDirectory)
	mockFS.EXPECT().Exists(".git").Return(true, nil).Times(3)
	mockFS.EXPECT().IsDir(".git").Return(true, nil).Times(3) // Called in detectSingleRepoMode (2x) and validateGitDirectory (1x)

	// Mock Git status for validation (called 2 times: validateGitStatus and validateGitConfiguration)
	mockGit.EXPECT().Status(".").Return("On branch main", nil).Times(2)

	// Mock verbose logging
	mockLogger.EXPECT().Logf("Starting CGWT execution")
	mockLogger.EXPECT().Logf("Checking for .git directory...")
	mockLogger.EXPECT().Logf("Verifying .git is a directory...")
	mockLogger.EXPECT().Logf("Git repository detected")
	mockLogger.EXPECT().Logf("Starting project structure validation")
	mockLogger.EXPECT().Logf("Checking for .git directory...")
	mockLogger.EXPECT().Logf("Verifying .git is a directory...")
	mockLogger.EXPECT().Logf("Git repository detected")
	mockLogger.EXPECT().Logf("Validating single repository mode")
	mockLogger.EXPECT().Logf("Validating repository: %s", ".")
	mockLogger.EXPECT().Logf("Executing git status in: %s", ".")
	mockLogger.EXPECT().Logf("Validating Git configuration in: %s", ".")
	mockLogger.EXPECT().Logf("Executing git status in: %s", ".")
	mockLogger.EXPECT().Logf("CGWT execution completed successfully")

	err := cgwt.Run()
	assert.NoError(t, err)
}

func TestCGWT_ValidateSingleRepository_Success(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockFS := fs.NewMockFS(ctrl)
	mockGit := git.NewMockGit(ctrl)
	mockLogger := logger.NewMockLogger(ctrl)

	cgwt := NewCGWT()
	cgwt.SetVerbose(true)
	cgwt.SetLogger(mockLogger)

	// Override adapters with mocks
	c := cgwt.(*realCGWT)
	c.fs = mockFS
	c.git = mockGit

	// Mock repository validation
	mockFS.EXPECT().Exists(".git").Return(true, nil)
	mockFS.EXPECT().IsDir(".git").Return(true, nil)
	mockGit.EXPECT().Status(".").Return("On branch main", nil)
	mockGit.EXPECT().Status(".").Return("On branch main", nil) // Called twice for validation

	// Mock verbose logging
	mockLogger.EXPECT().Logf("Validating repository: %s", ".")
	mockLogger.EXPECT().Logf("Executing git status in: %s", ".")
	mockLogger.EXPECT().Logf("Validating Git configuration in: %s", ".")
	mockLogger.EXPECT().Logf("Executing git status in: %s", ".")

	err := cgwt.ValidateSingleRepository()
	assert.NoError(t, err)
}

func TestCGWT_ValidateSingleRepository_NoGitDir(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockFS := fs.NewMockFS(ctrl)
	mockLogger := logger.NewMockLogger(ctrl)

	cgwt := NewCGWT()
	cgwt.SetVerbose(true)
	cgwt.SetLogger(mockLogger)

	// Override adapters with mocks
	c := cgwt.(*realCGWT)
	c.fs = mockFS

	// Mock repository validation - .git not found
	mockFS.EXPECT().Exists(".git").Return(false, nil)

	// Mock verbose logging
	mockLogger.EXPECT().Logf("Validating repository: %s", ".")
	mockLogger.EXPECT().Logf("Error: .git directory not found")

	err := cgwt.ValidateSingleRepository()
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "not a valid Git repository: .git directory not found")
}

func TestCGWT_ValidateSingleRepository_GitStatusError(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockFS := fs.NewMockFS(ctrl)
	mockGit := git.NewMockGit(ctrl)
	mockLogger := logger.NewMockLogger(ctrl)

	cgwt := NewCGWT()
	cgwt.SetVerbose(true)
	cgwt.SetLogger(mockLogger)

	// Override adapters with mocks
	c := cgwt.(*realCGWT)
	c.fs = mockFS
	c.git = mockGit

	// Mock repository validation
	mockFS.EXPECT().Exists(".git").Return(true, nil)
	mockFS.EXPECT().IsDir(".git").Return(true, nil)
	mockGit.EXPECT().Status(".").Return("", assert.AnError)

	// Mock verbose logging
	mockLogger.EXPECT().Logf("Validating repository: %s", ".")
	mockLogger.EXPECT().Logf("Executing git status in: %s", ".")
	mockLogger.EXPECT().Logf("Error: %v", gomock.Any())

	err := cgwt.ValidateSingleRepository()
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "not a valid Git repository")
}
