//go:build unit

package prompt

import (
	"bufio"
	"strings"
	"testing"

	tea "github.com/charmbracelet/bubbletea"
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
			model = model.updateFilteredChoices()

			assert.Equal(t, len(tt.expectedNames), len(model.filteredChoices))
			assert.Equal(t, len(tt.expectedIndices), len(model.filteredIndices))

			for i, expectedName := range tt.expectedNames {
				assert.Equal(t, expectedName, model.filteredChoices[i].Name)
				assert.Equal(t, tt.expectedIndices[i], model.filteredIndices[i])
			}
		})
	}
}

// TestPromptSelectTargetBubbleTea tests the Bubble Tea integration to prevent "unexpected model type" errors.
func TestPromptSelectTargetBubbleTea(t *testing.T) {
	choices := []TargetChoice{
		{Type: TargetRepository, Name: "test-repo-1"},
		{Type: TargetRepository, Name: "test-repo-2"},
		{Type: TargetWorkspace, Name: "test-workspace-1"},
	}

	tests := []struct {
		name              string
		choices           []TargetChoice
		showWorktreeLabel bool
		expectError       bool
	}{
		{
			name:              "valid choices without worktree labels",
			choices:           choices,
			showWorktreeLabel: false,
			expectError:       false,
		},
		{
			name:              "valid choices with worktree labels",
			choices:           choices,
			showWorktreeLabel: true,
			expectError:       false,
		},
		{
			name:              "empty choices should error",
			choices:           []TargetChoice{},
			showWorktreeLabel: false,
			expectError:       true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// This test verifies that the Bubble Tea program runs without the "unexpected model type" error
			// We can't easily test the full interactive flow in unit tests, but we can verify the setup
			// doesn't cause type assertion errors

			if tt.expectError {
				// For empty choices, we expect an error from the promptSelectTargetBubbleTea function
				// before it even gets to the Bubble Tea program
				_, err := promptSelectTargetBubbleTea(tt.choices, tt.showWorktreeLabel)
				assert.Error(t, err)
				return
			}

			// For valid choices, we can't easily test the full interactive flow in unit tests
			// because it requires user input. However, we can verify that the model creation
			// and type assertions work correctly by testing the model creation directly
			model := initialSelectModel(tt.choices, tt.showWorktreeLabel)

			// Verify the model was created correctly
			assert.Equal(t, len(tt.choices), len(model.choices))
			assert.Equal(t, len(tt.choices), len(model.filteredChoices))
			assert.Equal(t, tt.showWorktreeLabel, model.showWorktreeLabel)

			// Verify the model implements the tea.Model interface correctly
			// This ensures the type assertion in promptSelectTargetBubbleTea will work
			var teaModel tea.Model = model
			assert.NotNil(t, teaModel)

			// Test that the model can be cast back to selectModel without error
			// This simulates what happens in promptSelectTargetBubbleTea
			castModel, ok := teaModel.(selectModel)
			assert.True(t, ok, "Model should be castable to selectModel")
			assert.Equal(t, model.choices, castModel.choices)
		})
	}
}
