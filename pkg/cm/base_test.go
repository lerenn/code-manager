//go:build unit

package cm

import (
	"testing"

	"github.com/lerenn/cm/pkg/config"
	"github.com/lerenn/cm/pkg/fs"
	"github.com/lerenn/cm/pkg/git"
	"github.com/lerenn/cm/pkg/logger"
	"github.com/lerenn/cm/pkg/status"
	"github.com/stretchr/testify/assert"
	"go.uber.org/mock/gomock"
)

func TestBase_verbosePrint(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockFS := fs.NewMockFS(ctrl)
	mockGit := git.NewMockGit(ctrl)
	mockStatus := status.NewMockManager(ctrl)
	mockLogger := logger.NewNoopLogger()

	tests := []struct {
		name    string
		verbose bool
		msg     string
		args    []interface{}
	}{
		{
			name:    "Verbose enabled with simple message",
			verbose: true,
			msg:     "Test message",
			args:    []interface{}{},
		},
		{
			name:    "Verbose enabled with formatted message",
			verbose: true,
			msg:     "Test %s %d",
			args:    []interface{}{"message", 42},
		},
		{
			name:    "Verbose disabled",
			verbose: false,
			msg:     "Test message",
			args:    []interface{}{},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			base := newBase(mockFS, mockGit, &config.Config{}, mockStatus, mockLogger, tt.verbose)

			// This should not panic regardless of verbose setting
			base.verbosePrint(tt.msg, tt.args...)
		})
	}
}

func TestBase_validateGitConfiguration(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockFS := fs.NewMockFS(ctrl)
	mockGit := git.NewMockGit(ctrl)
	mockStatus := status.NewMockManager(ctrl)
	mockLogger := logger.NewNoopLogger()

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
			base := newBase(mockFS, mockGit, &config.Config{}, mockStatus, mockLogger, tt.verbose)

			mockGit.EXPECT().Status(tt.workDir).Return(tt.gitStatus, tt.gitError).Times(1)

			err := base.validateGitConfiguration(tt.workDir)

			if tt.expectError {
				assert.Error(t, err)
				assert.Contains(t, err.Error(), "git configuration error")
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
			base := newBase(mockFS, mockGit, &config.Config{}, mockStatus, mockLogger, tt.verbose)

			mockFS.EXPECT().Exists(tt.worktreePath).Return(tt.exists, tt.existsError).Times(1)

			if tt.exists && tt.existsError == nil {
				mockFS.EXPECT().RemoveAll(tt.worktreePath).Return(tt.removeError).Times(1)
			}

			err := base.cleanupWorktreeDirectory(tt.worktreePath)

			if tt.expectError {
				assert.Error(t, err)
				if tt.expectedError != "" {
					assert.Contains(t, err.Error(), tt.expectedError)
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

	tests := []struct {
		name         string
		basePath     string
		worktreesDir string
		repoURL      string
		branch       string
		expected     string
	}{
		{
			name:         "Simple path construction with base path",
			basePath:     "/base/path",
			worktreesDir: "",
			repoURL:      "github.com/lerenn/example",
			branch:       "main",
			expected:     "/base/path/github.com/lerenn/example/main",
		},
		{
			name:         "Path construction with worktrees directory",
			basePath:     "/base/path",
			worktreesDir: "/custom/worktrees",
			repoURL:      "github.com/lerenn/example",
			branch:       "main",
			expected:     "/custom/worktrees/github.com/lerenn/example/main",
		},
		{
			name:         "Path with branch containing slash using worktrees directory",
			basePath:     "/base/path",
			worktreesDir: "/custom/worktrees",
			repoURL:      "github.com/lerenn/example",
			branch:       "feature/new-feature",
			expected:     "/custom/worktrees/github.com/lerenn/example/feature/new-feature",
		},
		{
			name:         "Empty base path with worktrees directory",
			basePath:     "",
			worktreesDir: "/custom/worktrees",
			repoURL:      "github.com/lerenn/example",
			branch:       "main",
			expected:     "/custom/worktrees/github.com/lerenn/example/main",
		},
		{
			name:         "Empty base path and worktrees directory",
			basePath:     "",
			worktreesDir: "",
			repoURL:      "github.com/lerenn/example",
			branch:       "main",
			expected:     "github.com/lerenn/example/main",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			base := newBase(mockFS, mockGit, &config.Config{
				BasePath:     tt.basePath,
				WorktreesDir: tt.worktreesDir,
			}, mockStatus, mockLogger, false)

			result := base.buildWorktreePath(tt.repoURL, tt.branch)
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestBase_parseConfirmationInput(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockFS := fs.NewMockFS(ctrl)
	mockGit := git.NewMockGit(ctrl)
	mockStatus := status.NewMockManager(ctrl)
	mockLogger := logger.NewNoopLogger()

	tests := []struct {
		name        string
		input       string
		expected    bool
		expectError bool
		errorMsg    string
	}{
		{
			name:        "Yes - lowercase",
			input:       "y",
			expected:    true,
			expectError: false,
		},
		{
			name:        "Yes - uppercase",
			input:       "Y",
			expected:    true,
			expectError: false,
		},
		{
			name:        "Yes - full word lowercase",
			input:       "yes",
			expected:    true,
			expectError: false,
		},
		{
			name:        "Yes - full word uppercase",
			input:       "YES",
			expected:    true,
			expectError: false,
		},
		{
			name:        "No - lowercase",
			input:       "n",
			expected:    false,
			expectError: false,
		},
		{
			name:        "No - uppercase",
			input:       "N",
			expected:    false,
			expectError: false,
		},
		{
			name:        "No - full word lowercase",
			input:       "no",
			expected:    false,
			expectError: false,
		},
		{
			name:        "No - full word uppercase",
			input:       "NO",
			expected:    false,
			expectError: false,
		},
		{
			name:        "Empty input",
			input:       "",
			expected:    false,
			expectError: false,
		},
		{
			name:        "Whitespace only",
			input:       "   ",
			expected:    false,
			expectError: false,
		},
		{
			name:        "Quit command",
			input:       "q",
			expected:    false,
			expectError: true,
			errorMsg:    "user cancelled",
		},
		{
			name:        "Quit command - full word",
			input:       "quit",
			expected:    false,
			expectError: true,
			errorMsg:    "user cancelled",
		},
		{
			name:        "Exit command",
			input:       "exit",
			expected:    false,
			expectError: true,
			errorMsg:    "user cancelled",
		},
		{
			name:        "Cancel command",
			input:       "cancel",
			expected:    false,
			expectError: true,
			errorMsg:    "user cancelled",
		},
		{
			name:        "Invalid input",
			input:       "maybe",
			expected:    false,
			expectError: true,
			errorMsg:    "invalid input",
		},
		{
			name:        "Invalid input with numbers",
			input:       "123",
			expected:    false,
			expectError: true,
			errorMsg:    "invalid input",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			base := newBase(mockFS, mockGit, &config.Config{}, mockStatus, mockLogger, false)

			result, err := base.parseConfirmationInput(tt.input)

			if tt.expectError {
				assert.Error(t, err)
				if tt.errorMsg != "" {
					assert.Contains(t, err.Error(), tt.errorMsg)
				}
			} else {
				assert.NoError(t, err)
			}
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestBase_isQuitCommand(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockFS := fs.NewMockFS(ctrl)
	mockGit := git.NewMockGit(ctrl)
	mockStatus := status.NewMockManager(ctrl)
	mockLogger := logger.NewNoopLogger()

	tests := []struct {
		name     string
		input    string
		expected bool
	}{
		{
			name:     "Quit command - q",
			input:    "q",
			expected: true,
		},
		{
			name:     "Quit command - quit",
			input:    "quit",
			expected: true,
		},
		{
			name:     "Quit command - exit",
			input:    "exit",
			expected: true,
		},
		{
			name:     "Quit command - cancel",
			input:    "cancel",
			expected: true,
		},
		{
			name:     "Not a quit command - yes",
			input:    "yes",
			expected: false,
		},
		{
			name:     "Not a quit command - no",
			input:    "no",
			expected: false,
		},
		{
			name:     "Not a quit command - 1",
			input:    "1",
			expected: false,
		},
		{
			name:     "Not a quit command - empty",
			input:    "",
			expected: false,
		},
		{
			name:     "Not a quit command - partial match",
			input:    "qu",
			expected: false,
		},
		{
			name:     "Not a quit command - case sensitive",
			input:    "QUIT",
			expected: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			base := newBase(mockFS, mockGit, &config.Config{}, mockStatus, mockLogger, false)

			result := base.isQuitCommand(tt.input)
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestBase_parseNumericInput(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockFS := fs.NewMockFS(ctrl)
	mockGit := git.NewMockGit(ctrl)
	mockStatus := status.NewMockManager(ctrl)
	mockLogger := logger.NewNoopLogger()

	tests := []struct {
		name        string
		input       string
		expected    int
		expectError bool
	}{
		{
			name:        "Valid single digit",
			input:       "1",
			expected:    1,
			expectError: false,
		},
		{
			name:        "Valid multiple digits",
			input:       "42",
			expected:    42,
			expectError: false,
		},
		{
			name:        "Valid zero",
			input:       "0",
			expected:    0,
			expectError: false,
		},
		{
			name:        "Valid negative number",
			input:       "-5",
			expected:    -5,
			expectError: false,
		},
		{
			name:        "Invalid input - letters",
			input:       "abc",
			expected:    0,
			expectError: true,
		},
		{
			name:        "Mixed input with leading number",
			input:       "12abc",
			expected:    12,
			expectError: false,
		},
		{
			name:        "Invalid input - empty",
			input:       "",
			expected:    0,
			expectError: true,
		},
		{
			name:        "Invalid input - whitespace",
			input:       "   ",
			expected:    0,
			expectError: true,
		},
		{
			name:        "Decimal input",
			input:       "3.14",
			expected:    3,
			expectError: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			base := newBase(mockFS, mockGit, &config.Config{}, mockStatus, mockLogger, false)

			result, err := base.parseNumericInput(tt.input)

			if tt.expectError {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestBase_isValidChoice(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockFS := fs.NewMockFS(ctrl)
	mockGit := git.NewMockGit(ctrl)
	mockStatus := status.NewMockManager(ctrl)
	mockLogger := logger.NewNoopLogger()

	tests := []struct {
		name      string
		choice    int
		maxChoice int
		expected  bool
	}{
		{
			name:      "Valid choice at minimum",
			choice:    1,
			maxChoice: 3,
			expected:  true,
		},
		{
			name:      "Valid choice at maximum",
			choice:    3,
			maxChoice: 3,
			expected:  true,
		},
		{
			name:      "Valid choice in middle",
			choice:    2,
			maxChoice: 3,
			expected:  true,
		},
		{
			name:      "Invalid choice - too low",
			choice:    0,
			maxChoice: 3,
			expected:  false,
		},
		{
			name:      "Invalid choice - negative",
			choice:    -1,
			maxChoice: 3,
			expected:  false,
		},
		{
			name:      "Invalid choice - too high",
			choice:    4,
			maxChoice: 3,
			expected:  false,
		},
		{
			name:      "Edge case - single choice",
			choice:    1,
			maxChoice: 1,
			expected:  true,
		},
		{
			name:      "Edge case - zero max choice",
			choice:    1,
			maxChoice: 0,
			expected:  false,
		},
		{
			name:      "Edge case - negative max choice",
			choice:    1,
			maxChoice: -1,
			expected:  false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			base := newBase(mockFS, mockGit, &config.Config{}, mockStatus, mockLogger, false)

			result := base.isValidChoice(tt.choice, tt.maxChoice)
			assert.Equal(t, tt.expected, result)
		})
	}
}
