//go:build unit

package forge

import (
	"testing"

	"github.com/lerenn/wtm/pkg/issue"
	"github.com/lerenn/wtm/pkg/logger"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestNewManager(t *testing.T) {
	loggerInstance := logger.NewNoopLogger()
	manager := NewManager(loggerInstance)

	assert.NotNil(t, manager)
	assert.NotNil(t, manager.forges)
}

func TestManager_GetForge(t *testing.T) {
	loggerInstance := logger.NewNoopLogger()
	manager := NewManager(loggerInstance)

	// Test getting GitHub forge
	githubForge, err := manager.GetForge("github")
	require.NoError(t, err)
	assert.NotNil(t, githubForge)
	assert.Equal(t, "github", githubForge.Name())

	// Test getting non-existent forge
	_, err = manager.GetForge("nonexistent")
	assert.Error(t, err)
	assert.ErrorIs(t, err, ErrUnsupportedForge)
}

func TestGitHub_Name(t *testing.T) {
	github := NewGitHub()
	assert.Equal(t, "github", github.Name())
}

func TestGitHub_GenerateBranchName(t *testing.T) {
	github := NewGitHub()

	tests := []struct {
		name     string
		issue    *issue.Info
		expected string
	}{
		{
			name: "simple title",
			issue: &issue.Info{
				Number: 123,
				Title:  "Fix Login Bug",
			},
			expected: "123-fix-login-bug",
		},
		{
			name: "title with special characters",
			issue: &issue.Info{
				Number: 456,
				Title:  "Add new feature! @#$%^&*()",
			},
			expected: "456-add-new-feature",
		},
		{
			name: "title with multiple spaces",
			issue: &issue.Info{
				Number: 789,
				Title:  "Update   documentation   files",
			},
			expected: "789-update-documentation-files",
		},
		{
			name: "title with hyphens",
			issue: &issue.Info{
				Number: 101,
				Title:  "Fix-bug-in-login-system",
			},
			expected: "101-fix-bug-in-login-system",
		},
		{
			name: "very long title",
			issue: &issue.Info{
				Number: 202,
				Title:  "This is a very long title that should be truncated to fit within the maximum length limit of 80 characters for branch names",
			},
			expected: "202-this-is-a-very-long-title-that-should-be-truncated-to-fit-within-the-maximum-len",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := github.GenerateBranchName(tt.issue)
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestGitHub_ParseIssueReference(t *testing.T) {
	github := NewGitHub()

	tests := []struct {
		name        string
		issueRef    string
		expectError bool
		expected    *issue.Reference
	}{
		{
			name:     "GitHub URL format",
			issueRef: "https://github.com/owner/repo/issues/123",
			expected: &issue.Reference{
				Owner:       "owner",
				Repository:  "repo",
				IssueNumber: 123,
				URL:         "https://github.com/owner/repo/issues/123",
			},
		},
		{
			name:     "owner/repo#issue format",
			issueRef: "owner/repo#456",
			expected: &issue.Reference{
				Owner:       "owner",
				Repository:  "repo",
				IssueNumber: 456,
				URL:         "https://github.com/owner/repo/issues/456",
			},
		},
		{
			name:        "issue number only (requires context)",
			issueRef:    "789",
			expectError: true,
		},
		{
			name:        "invalid URL format",
			issueRef:    "https://github.com/owner/repo/pulls/123",
			expectError: true,
		},
		{
			name:        "invalid owner/repo format",
			issueRef:    "owner#123",
			expectError: true,
		},
		{
			name:        "invalid issue number",
			issueRef:    "owner/repo#abc",
			expectError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := github.ParseIssueReference(tt.issueRef)
			if tt.expectError {
				assert.Error(t, err)
				return
			}
			require.NoError(t, err)
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestGitHub_sanitizeTitle(t *testing.T) {
	github := NewGitHub()

	tests := []struct {
		input    string
		expected string
	}{
		{"Fix Login Bug", "fix-login-bug"},
		{"Add new feature!", "add-new-feature"},
		{"Update   documentation", "update-documentation"},
		{"Fix-bug-in-system", "fix-bug-in-system"},
		{"Title with @#$%^&*() symbols", "title-with-symbols"},
		{"Mixed Case Title", "mixed-case-title"},
		{"Title with numbers 123", "title-with-numbers-123"},
		{"", ""},
		{"   spaces   ", "spaces"},
		{"multiple---hyphens", "multiple-hyphens"},
	}

	for _, tt := range tests {
		t.Run(tt.input, func(t *testing.T) {
			result := github.sanitizeTitle(tt.input)
			assert.Equal(t, tt.expected, result)
		})
	}
}
