//go:build unit

package branch

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestSanitizeBranchName(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected string
		wantErr  bool
	}{
		{
			name:     "Simple branch name with slash",
			input:    "feature/new-branch",
			expected: "feature/new-branch",
			wantErr:  false,
		},
		{
			name:     "Branch name with invalid characters",
			input:    "bugfix/issue#123",
			expected: "bugfix/issue_123",
			wantErr:  false,
		},
		{
			name:     "Branch name with dots",
			input:    "release/v1.0.0",
			expected: "release/v1.0.0",
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
		{
			name:     "Branch name without slash",
			input:    "main",
			expected: "main",
			wantErr:  false,
		},
		{
			name:     "Branch name with leading dash",
			input:    "-invalid-branch",
			expected: "invalid-branch",
			wantErr:  false,
		},
		{
			name:     "Branch name with consecutive dots",
			input:    "feature..test",
			expected: "feature_test",
			wantErr:  false,
		},
		{
			name:     "Branch name with consecutive slashes",
			input:    "feature//test",
			expected: "feature/test",
			wantErr:  false,
		},
		{
			name:     "Single @ character (not allowed)",
			input:    "@",
			expected: "",
			wantErr:  true,
		},
		{
			name:     "Branch name with @{ sequence (not allowed)",
			input:    "feature@{test}",
			expected: "",
			wantErr:  true,
		},
		{
			name:     "Branch name with backslash (not allowed)",
			input:    "feature\\test",
			expected: "",
			wantErr:  true,
		},
		{
			name:     "Branch name with spaces and special chars",
			input:    "feature test~with^special:chars",
			expected: "feature_test_with_special_chars",
			wantErr:  false,
		},
		{
			name:     "Branch name with question marks and asterisks",
			input:    "feature?test*",
			expected: "feature_test",
			wantErr:  false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := SanitizeBranchName(tt.input)
			if tt.wantErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
				assert.Equal(t, tt.expected, result)
			}
		})
	}
}

func TestSanitizeBranchNameForFilename(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected string
	}{
		{
			name:     "Simple branch name with slash",
			input:    "feature/test-branch",
			expected: "feature-test-branch",
		},
		{
			name:     "Branch name with multiple slashes",
			input:    "feature/test/sub-branch",
			expected: "feature-test-sub-branch",
		},
		{
			name:     "Branch name with backslash",
			input:    "feature\\test",
			expected: "feature-test",
		},
		{
			name:     "Branch name with colon",
			input:    "feature:test",
			expected: "feature-test",
		},
		{
			name:     "Branch name with asterisk and question mark",
			input:    "feature*test?",
			expected: "feature-test",
		},
		{
			name:     "Branch name with quotes and brackets",
			input:    "feature\"test<>|",
			expected: "feature-test",
		},
		{
			name:     "Branch name with consecutive hyphens",
			input:    "feature---test",
			expected: "feature-test",
		},
		{
			name:     "Branch name with leading/trailing hyphens",
			input:    "-feature-test-",
			expected: "feature-test",
		},
		{
			name:     "Empty branch name",
			input:    "",
			expected: "",
		},
		{
			name:     "Branch name without special characters",
			input:    "main",
			expected: "main",
		},
		{
			name:     "Branch name with only special characters",
			input:    "/*?<>|",
			expected: "",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := SanitizeBranchNameForFilename(tt.input)
			assert.Equal(t, tt.expected, result)
		})
	}
}
