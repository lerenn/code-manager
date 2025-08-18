//go:build unit

package cm

import (
	"strings"
	"testing"

	"github.com/lerenn/cm/pkg/fs"
	"github.com/lerenn/cm/pkg/git"
	"github.com/stretchr/testify/assert"
	"go.uber.org/mock/gomock"
)

func TestRealCM_sanitizeBranchName(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockFS := fs.NewMockFS(ctrl)
	mockGit := git.NewMockGit(ctrl)

	cm := NewCM(createTestConfig())

	// Override adapters with mocks
	c := cm.(*realCM)
	c.fs = mockFS
	c.git = mockGit

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
			input:    "feature?test*with[wildcards]",
			expected: "feature_test_with_wildcards",
			wantErr:  false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := cm.(*realCM).sanitizeBranchName(tt.input)
			if tt.wantErr {
				if tt.input == "" {
					assert.ErrorIs(t, err, ErrBranchNameEmpty)
				} else {
					assert.Error(t, err)
					// Check for specific error messages for new Git rule violations
					if tt.input == "@" {
						assert.Contains(t, err.Error(), "branch name cannot be the single character @")
					} else if strings.Contains(tt.input, "@{") {
						assert.Contains(t, err.Error(), "branch name cannot contain the sequence @{")
					} else if strings.Contains(tt.input, "\\") {
						assert.Contains(t, err.Error(), "branch name cannot contain backslash")
					}
				}
			} else {
				assert.NoError(t, err)
				assert.Equal(t, tt.expected, result)
			}
		})
	}
}
