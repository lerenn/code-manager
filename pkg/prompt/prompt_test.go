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
