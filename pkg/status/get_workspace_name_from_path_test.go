//go:build unit

package status

import (
	"testing"

	"github.com/lerenn/code-manager/pkg/config"
	"github.com/lerenn/code-manager/pkg/fs"
	"github.com/stretchr/testify/assert"
	"go.uber.org/mock/gomock"
)

func TestGetWorkspaceNameFromPath(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockFS := fs.NewMockFS(ctrl)

	cfg := &config.Config{
		BasePath:   "/home/user/.cm",
		StatusFile: "/home/user/.cmstatus.yaml",
	}

	manager := &realManager{
		fs:         mockFS,
		config:     cfg,
		workspaces: make(map[string]map[string][]WorktreeInfo),
	}

	// Test cases
	testCases := []struct {
		path     string
		expected string
	}{
		{
			path:     "/home/user/workspace.code-workspace",
			expected: "workspace",
		},
		{
			path:     "/path/to/my-project.code-workspace",
			expected: "my-project",
		},
		{
			path:     "simple.code-workspace",
			expected: "simple",
		},
		{
			path:     "/home/user/no-extension",
			expected: "no-extension",
		},
		{
			path:     "",
			expected: ".",
		},
	}

	for _, tc := range testCases {
		t.Run(tc.path, func(t *testing.T) {
			result := manager.getWorkspaceNameFromPath(tc.path)
			assert.Equal(t, tc.expected, result)
		})
	}
}
