//go:build unit

package dependencies

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// TestDependencies_Validate_MissingFS tests validation failure when FS is missing
func TestDependencies_Validate_MissingFS(t *testing.T) {
	deps := New()
	deps.FS = nil // Override the default

	err := deps.Validate()
	assert.Error(t, err)
	assert.ErrorIs(t, err, ErrFSMissing)
}

// TestDependencies_Validate_MissingGit tests validation failure when Git is missing
func TestDependencies_Validate_MissingGit(t *testing.T) {
	deps := New()
	deps.Git = nil // Override the default

	err := deps.Validate()
	assert.Error(t, err)
	assert.ErrorIs(t, err, ErrGitMissing)
}

// TestDependencies_Validate_MissingConfig tests validation failure when Config is missing
func TestDependencies_Validate_MissingConfig(t *testing.T) {
	deps := New()
	deps.Config = nil // Override the default

	err := deps.Validate()
	assert.Error(t, err)
	assert.ErrorIs(t, err, ErrConfigMissing)
}

// TestDependencies_Validate_AllMissing tests validation failure when all dependencies are missing
func TestDependencies_Validate_AllMissing(t *testing.T) {
	deps := &Dependencies{} // All fields are nil

	err := deps.Validate()
	assert.Error(t, err)
	// Should return the first missing dependency (FS)
	assert.ErrorIs(t, err, ErrFSMissing)
}

// TestDependencies_New_Defaults tests that New() creates a Dependencies instance with proper defaults
func TestDependencies_New_Defaults(t *testing.T) {
	deps := New()

	// Check that defaults are set
	assert.NotNil(t, deps.FS)
	assert.NotNil(t, deps.Git)
	assert.NotNil(t, deps.Logger)
	assert.NotNil(t, deps.Prompt)
	assert.NotNil(t, deps.HookManager)

	// Check that configurable dependencies are nil by default
	assert.Nil(t, deps.Config)
	assert.Nil(t, deps.StatusManager)
	assert.Nil(t, deps.RepositoryProvider)
	assert.Nil(t, deps.WorkspaceProvider)
	assert.Nil(t, deps.WorktreeProvider)

	// Validation should fail because configurable dependencies are missing
	err := deps.Validate()
	assert.Error(t, err)
	assert.ErrorIs(t, err, ErrConfigMissing)
}

// TestDependencies_ErrorTypes tests that error types are properly defined and catchable
func TestDependencies_ErrorTypes(t *testing.T) {
	// Test that all error types are defined and can be caught
	testCases := []struct {
		name     string
		setup    func() *Dependencies
		expected error
	}{
		{
			name: "FS missing",
			setup: func() *Dependencies {
				deps := New()
				deps.FS = nil
				return deps
			},
			expected: ErrFSMissing,
		},
		{
			name: "Git missing",
			setup: func() *Dependencies {
				deps := New()
				deps.Git = nil
				return deps
			},
			expected: ErrGitMissing,
		},
		{
			name: "Config missing",
			setup: func() *Dependencies {
				deps := New()
				deps.Config = nil
				return deps
			},
			expected: ErrConfigMissing,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			deps := tc.setup()
			err := deps.Validate()
			require.Error(t, err)
			assert.ErrorIs(t, err, tc.expected)
		})
	}
}

// TestDependencies_ErrorMessages tests that error messages are descriptive
func TestDependencies_ErrorMessages(t *testing.T) {
	testCases := []struct {
		name     string
		setup    func() *Dependencies
		expected string
	}{
		{
			name: "FS missing message",
			setup: func() *Dependencies {
				deps := New()
				deps.FS = nil
				return deps
			},
			expected: "fs dependency is required but not set",
		},
		{
			name: "Config missing message",
			setup: func() *Dependencies {
				deps := New()
				deps.Config = nil
				return deps
			},
			expected: "config dependency is required but not set",
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			deps := tc.setup()
			err := deps.Validate()
			require.Error(t, err)
			assert.Equal(t, tc.expected, err.Error())
		})
	}
}

// TestDependencies_ValidationOrder tests that validation checks dependencies in the correct order
func TestDependencies_ValidationOrder(t *testing.T) {
	// Test that validation stops at the first missing dependency
	deps := &Dependencies{} // All fields are nil

	err := deps.Validate()
	assert.Error(t, err)
	// Should return the first missing dependency (FS), not any later ones
	assert.ErrorIs(t, err, ErrFSMissing)
	assert.NotErrorIs(t, err, ErrConfigMissing)
	assert.NotErrorIs(t, err, ErrStatusManagerMissing)
}

// TestDependencies_ErrorVariables tests that all error variables are properly defined
func TestDependencies_ErrorVariables(t *testing.T) {
	// Test that all error variables are defined and have the expected messages
	errorTests := []struct {
		err      error
		expected string
	}{
		{ErrFSMissing, "fs dependency is required but not set"},
		{ErrGitMissing, "git dependency is required but not set"},
		{ErrConfigMissing, "config dependency is required but not set"},
		{ErrStatusManagerMissing, "status manager dependency is required but not set"},
		{ErrLoggerMissing, "logger dependency is required but not set"},
		{ErrPromptMissing, "prompt dependency is required but not set"},
		{ErrHookManagerMissing, "hook manager dependency is required but not set"},
		{ErrRepositoryProviderMissing, "repository provider dependency is required but not set"},
		{ErrWorkspaceProviderMissing, "workspace provider dependency is required but not set"},
		{ErrWorktreeProviderMissing, "worktree provider dependency is required but not set"},
	}

	for _, test := range errorTests {
		t.Run(test.err.Error(), func(t *testing.T) {
			assert.Equal(t, test.expected, test.err.Error())
		})
	}
}
