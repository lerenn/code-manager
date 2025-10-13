//go:build unit

package prompt

import (
	"bufio"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestRealPrompt_PromptForRepositoriesDir(t *testing.T) {
	tests := []struct {
		name        string
		defaultPath string
		input       string
		expected    string
	}{
		{
			name:        "empty input uses default",
			defaultPath: "~/Code",
			input:       "\n",
			expected:    "~/Code",
		},
		{
			name:        "whitespace input uses default",
			defaultPath: "~/Code",
			input:       "   \n",
			expected:    "~/Code",
		},
		{
			name:        "custom path",
			defaultPath: "~/Code",
			input:       "~/Projects\n",
			expected:    "~/Projects",
		},
		{
			name:        "custom path with whitespace",
			defaultPath: "~/Code",
			input:       "  ~/Development  \n",
			expected:    "~/Development",
		},
		{
			name:        "empty default uses hardcoded default",
			defaultPath: "",
			input:       "\n",
			expected:    "~/Code/repos",
		},
		{
			name:        "custom default path",
			defaultPath: "~/Custom/Path",
			input:       "\n",
			expected:    "~/Custom/Path",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Create a prompt with a string reader
			p := &realPrompt{
				reader: bufio.NewReader(strings.NewReader(tt.input)),
			}

			result, err := p.PromptForRepositoriesDir(tt.defaultPath)
			assert.NoError(t, err)
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestRealPrompt_PromptForConfirmation(t *testing.T) {
	tests := []struct {
		name        string
		message     string
		defaultYes  bool
		input       string
		expected    bool
		expectError bool
	}{
		{
			name:       "yes input",
			message:    "Continue?",
			defaultYes: false,
			input:      "y\n",
			expected:   true,
		},
		{
			name:       "YES input",
			message:    "Continue?",
			defaultYes: false,
			input:      "YES\n",
			expected:   true,
		},
		{
			name:       "no input",
			message:    "Continue?",
			defaultYes: true,
			input:      "n\n",
			expected:   false,
		},
		{
			name:       "NO input",
			message:    "Continue?",
			defaultYes: true,
			input:      "NO\n",
			expected:   false,
		},
		{
			name:       "empty input with default yes",
			message:    "Continue?",
			defaultYes: true,
			input:      "\n",
			expected:   true,
		},
		{
			name:       "empty input with default no",
			message:    "Continue?",
			defaultYes: false,
			input:      "\n",
			expected:   false,
		},
		{
			name:        "invalid input",
			message:     "Continue?",
			defaultYes:  false,
			input:       "maybe\n",
			expectError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Create a prompt with a string reader
			p := &realPrompt{
				reader: bufio.NewReader(strings.NewReader(tt.input)),
			}

			result, err := p.PromptForConfirmation(tt.message, tt.defaultYes)
			if tt.expectError {
				assert.Error(t, err)
				assert.ErrorIs(t, err, ErrInvalidConfirmationInput)
			} else {
				assert.NoError(t, err)
				assert.Equal(t, tt.expected, result)
			}
		})
	}
}

func TestFormatChoice(t *testing.T) {
	tests := []struct {
		name              string
		choice            TargetChoice
		showWorktreeLabel bool
		expected          string
	}{
		{
			name: "repository without worktree label",
			choice: TargetChoice{
				Type: TargetRepository,
				Name: "my-repo",
			},
			showWorktreeLabel: false,
			expected:          "[repository] my-repo",
		},
		{
			name: "workspace without worktree label",
			choice: TargetChoice{
				Type: TargetWorkspace,
				Name: "my-workspace",
			},
			showWorktreeLabel: false,
			expected:          "[workspace] my-workspace",
		},
		{
			name: "repository with worktree label",
			choice: TargetChoice{
				Type:     TargetRepository,
				Name:     "my-repo",
				Worktree: "main",
			},
			showWorktreeLabel: true,
			expected:          "[repository] my-repo : main",
		},
		{
			name: "workspace with worktree label",
			choice: TargetChoice{
				Type:     TargetWorkspace,
				Name:     "my-workspace",
				Worktree: "feature-branch",
			},
			showWorktreeLabel: true,
			expected:          "[workspace] my-workspace : feature-branch",
		},
		{
			name: "repository with worktree label but showWorktreeLabel false",
			choice: TargetChoice{
				Type:     TargetRepository,
				Name:     "my-repo",
				Worktree: "main",
			},
			showWorktreeLabel: false,
			expected:          "[repository] my-repo",
		},
		{
			name: "workspace with empty worktree",
			choice: TargetChoice{
				Type:     TargetWorkspace,
				Name:     "my-workspace",
				Worktree: "",
			},
			showWorktreeLabel: true,
			expected:          "[workspace] my-workspace",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := formatChoice(tt.choice, tt.showWorktreeLabel, true)
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestSelectModel_UpdateFilteredChoices(t *testing.T) {
	choices := []TargetChoice{
		{Type: TargetRepository, Name: "alpha-repo"},
		{Type: TargetWorkspace, Name: "beta-workspace"},
		{Type: TargetRepository, Name: "gamma-repo"},
		{Type: TargetWorkspace, Name: "delta-workspace"},
	}

	tests := []struct {
		name            string
		filter          string
		expectedNames   []string
		expectedIndices []int
	}{
		{
			name:            "empty filter shows all",
			filter:          "",
			expectedNames:   []string{"alpha-repo", "beta-workspace", "gamma-repo", "delta-workspace"},
			expectedIndices: []int{0, 1, 2, 3},
		},
		{
			name:            "filter by 'repo'",
			filter:          "repo",
			expectedNames:   []string{"alpha-repo", "gamma-repo"},
			expectedIndices: []int{0, 2},
		},
		{
			name:            "filter by 'workspace'",
			filter:          "workspace",
			expectedNames:   []string{"beta-workspace", "delta-workspace"},
			expectedIndices: []int{1, 3},
		},
		{
			name:            "filter by 'alpha'",
			filter:          "alpha",
			expectedNames:   []string{"alpha-repo"},
			expectedIndices: []int{0},
		},
		{
			name:            "case insensitive filter",
			filter:          "ALPHA",
			expectedNames:   []string{"alpha-repo"},
			expectedIndices: []int{0},
		},
		{
			name:            "no matches",
			filter:          "nonexistent",
			expectedNames:   []string{},
			expectedIndices: []int{},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			model := initialSelectModel(choices, false)
			model.filter = tt.filter
			model.updateFilteredChoices()

			assert.Equal(t, len(tt.expectedNames), len(model.filteredChoices))
			assert.Equal(t, len(tt.expectedIndices), len(model.filteredIndices))

			for i, expectedName := range tt.expectedNames {
				assert.Equal(t, expectedName, model.filteredChoices[i].Name)
				assert.Equal(t, tt.expectedIndices[i], model.filteredIndices[i])
			}
		})
	}
}
